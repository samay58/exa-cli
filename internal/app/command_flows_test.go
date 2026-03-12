package app

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResearchRunPollsUntilCompleted(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var requests []string
	var pollCount int

	code := executeTest(
		t,
		[]string{"research", "run", "design a search cli", "--poll-interval", "1ms", "--format", "json", "--no-cache"},
		testEnv(t, map[string]string{"EXA_API_KEY": "test-key"}),
		roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requests = append(requests, req.Method+" "+req.URL.RequestURI())
			switch {
			case req.Method == http.MethodPost && req.URL.Path == "/research/v1":
				return jsonResponse(`{"requestId":"req_create","id":"research_123","status":"queued"}`), nil
			case req.Method == http.MethodGet && req.URL.RequestURI() == "/research/v1/research_123?events=true":
				pollCount++
				if pollCount == 1 {
					return jsonResponse(`{"requestId":"req_poll_1","id":"research_123","status":"running"}`), nil
				}
				return jsonResponse(`{"requestId":"req_poll_2","id":"research_123","status":"completed","report":"Final report","results":[{"title":"Design memo","url":"https://example.com/design"}]}`), nil
			default:
				t.Fatalf("unexpected request: %s %s", req.Method, req.URL.RequestURI())
				return nil, nil
			}
		}),
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("unexpected exit code: %d stderr=%s", code, stderr.String())
	}
	if len(requests) != 3 {
		t.Fatalf("expected create + two polls, got %v", requests)
	}

	var payload struct {
		Meta struct {
			Command   string `json:"command"`
			RequestID string `json:"requestId"`
		} `json:"meta"`
		Data struct {
			Status string `json:"status"`
			Report string `json:"report"`
		} `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode json: %v\n%s", err, stdout.String())
	}
	if payload.Meta.Command != "research run" || payload.Meta.RequestID != "req_poll_2" {
		t.Fatalf("unexpected meta: %+v", payload.Meta)
	}
	if payload.Data.Status != "completed" || payload.Data.Report != "Final report" {
		t.Fatalf("unexpected payload: %+v", payload.Data)
	}
}

func TestResearchRunDetachSkipsPolling(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var requestCount int

	code := executeTest(
		t,
		[]string{"research", "run", "design a search cli", "--detach", "--format", "json", "--no-cache"},
		testEnv(t, map[string]string{"EXA_API_KEY": "test-key"}),
		roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requestCount++
			if req.Method != http.MethodPost || req.URL.Path != "/research/v1" {
				t.Fatalf("unexpected request: %s %s", req.Method, req.URL.RequestURI())
			}
			return jsonResponse(`{"requestId":"req_create","id":"research_123","status":"queued"}`), nil
		}),
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("unexpected exit code: %d stderr=%s", code, stderr.String())
	}
	if requestCount != 1 {
		t.Fatalf("expected detach mode to skip polling, got %d requests", requestCount)
	}
	if !strings.Contains(stdout.String(), `"id": "research_123"`) {
		t.Fatalf("expected detached payload, got:\n%s", stdout.String())
	}
}

func TestResearchListUsesLimit(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeTest(
		t,
		[]string{"research", "list", "--limit", "3", "--format", "json", "--no-cache"},
		testEnv(t, map[string]string{"EXA_API_KEY": "test-key"}),
		roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodGet || req.URL.RequestURI() != "/research/v1?limit=3" {
				t.Fatalf("unexpected request: %s %s", req.Method, req.URL.RequestURI())
			}
			return jsonResponse(`{"requestId":"req_list","tasks":[{"id":"research_1","status":"completed"}]}`), nil
		}),
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("unexpected exit code: %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"requestId": "req_list"`) {
		t.Fatalf("expected request id in JSON output, got:\n%s", stdout.String())
	}
}

func TestResearchCancelUsesDelete(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeTest(
		t,
		[]string{"research", "cancel", "research_1", "--format", "json", "--no-cache"},
		testEnv(t, map[string]string{"EXA_API_KEY": "test-key"}),
		roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodDelete || req.URL.Path != "/research/v1/research_1" {
				t.Fatalf("unexpected request: %s %s", req.Method, req.URL.RequestURI())
			}
			return jsonResponse(`{"requestId":"req_cancel","status":"cancelled"}`), nil
		}),
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("unexpected exit code: %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"status": "cancelled"`) {
		t.Fatalf("expected cancelled payload, got:\n%s", stdout.String())
	}
}

func TestFindMapsAPIStatusCodes(t *testing.T) {
	cases := []struct {
		name       string
		statusCode int
		wantCode   int
	}{
		{name: "unauthorized", statusCode: http.StatusUnauthorized, wantCode: 3},
		{name: "rate_limit", statusCode: http.StatusTooManyRequests, wantCode: 5},
		{name: "server_error", statusCode: http.StatusInternalServerError, wantCode: 4},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			code := executeTest(
				t,
				[]string{"find", "golang", "--format", "json", "--no-cache"},
				testEnv(t, map[string]string{"EXA_API_KEY": "test-key"}),
				roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return jsonResponseStatus(tc.statusCode, `{"message":"api failure"}`), nil
				}),
				&stdout,
				&stderr,
			)
			if code != tc.wantCode {
				t.Fatalf("expected exit code %d, got %d stderr=%s", tc.wantCode, code, stderr.String())
			}
			if !strings.Contains(stderr.String(), "api failure") {
				t.Fatalf("expected API error in stderr, got: %s", stderr.String())
			}
		})
	}
}

func TestCodeRejectsInvalidTokens(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeTest(
		t,
		[]string{"code", "cobra output", "--tokens", "not-a-number"},
		testEnv(t, map[string]string{"EXA_API_KEY": "test-key"}),
		nil,
		&stdout,
		&stderr,
	)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "invalid token budget") {
		t.Fatalf("expected token validation error, got: %s", stderr.String())
	}
}

func TestFindRejectsInvalidPeopleDomain(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeTest(
		t,
		[]string{"find", "founder profile", "--category", "people", "--include-domain", "github.com"},
		testEnv(t, map[string]string{"EXA_API_KEY": "test-key"}),
		nil,
		&stdout,
		&stderr,
	)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "LinkedIn domains") {
		t.Fatalf("expected people domain validation error, got: %s", stderr.String())
	}
}

func TestRawRequestRejectsInvalidJSONFile(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte(`{"query":`), 0o644); err != nil {
		t.Fatalf("write invalid json: %v", err)
	}

	code := executeTest(
		t,
		[]string{"raw", "request", "--method", "POST", "--path", "/search", "--input", path},
		testEnv(t, map[string]string{"EXA_API_KEY": "test-key"}),
		nil,
		&stdout,
		&stderr,
	)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "decode") {
		t.Fatalf("expected decode error, got: %s", stderr.String())
	}
}

func TestReadSummaryDoesNotRequestFullTextByDefault(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var requestBody string

	code := executeTest(
		t,
		[]string{"read", "https://exa.ai/docs/reference/search", "--summary", "--format", "json", "--no-cache"},
		testEnv(t, map[string]string{"EXA_API_KEY": "test-key"}),
		roundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			requestBody = string(body)
			return jsonResponse(`{"requestId":"req_contents","results":[{"url":"https://exa.ai/docs/reference/search","summary":"Short summary"}]}`), nil
		}),
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("unexpected exit code: %d stderr=%s", code, stderr.String())
	}
	if strings.Contains(requestBody, `"text":true`) {
		t.Fatalf("did not expect read to request full text by default, got: %s", requestBody)
	}
	if !strings.Contains(requestBody, `"summary"`) {
		t.Fatalf("expected read summary request, got: %s", requestBody)
	}
}

func jsonResponseStatus(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(body)),
	}
}
