package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/samaydhawan/exa-cli/internal/version"
)

func TestHomeScreen(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeTest(t, nil, testEnv(t, nil), nil, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("unexpected exit code: %d\nstderr=%s", code, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Quickstart:") {
		t.Fatalf("expected home screen, got:\n%s", output)
	}
	if !strings.Contains(output, "Next steps:") {
		t.Fatalf("expected onboarding checklist, got:\n%s", output)
	}
	if strings.Contains(output, "built for shells, prompts, and long threads.") {
		t.Fatalf("did not expect banner on non-interactive output:\n%s", output)
	}
}

func TestHomeUsesConfiguredAuthSource(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte("api_key = 'stored-key'\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	code := executeTest(t, nil, []string{
		"EXA_CLI_CONFIG=" + configPath,
		"EXA_CLI_NO_BANNER=1",
		"EXA_CLI_NO_CACHE=1",
	}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("unexpected exit code: %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "auth: config") {
		t.Fatalf("expected home screen to reflect stored auth, got:\n%s", stdout.String())
	}
}

func TestFindJSONEnvelope(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeTest(
		t,
		[]string{"find", "golang", "--format", "json", "--no-cache"},
		testEnv(t, map[string]string{
			"EXA_API_KEY": "test-key",
		}),
		roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/search" {
				t.Fatalf("unexpected path: %s", req.URL.Path)
			}
			if got := req.Header.Get("x-api-key"); got != "test-key" {
				t.Fatalf("missing api key, got %q", got)
			}
			return jsonResponse(`{
				"requestId":"req_search",
				"type":"auto",
				"searchTime":0.42,
				"results":[{"title":"Example result","url":"https://example.com","summary":"Short summary"}]
			}`), nil
		}),
		&stdout,
		&stderr,
	)

	if code != 0 {
		t.Fatalf("unexpected exit code: %d\nstderr=%s", code, stderr.String())
	}

	var payload struct {
		Meta struct {
			Command   string `json:"command"`
			Cache     string `json:"cache"`
			RequestID string `json:"requestId"`
		} `json:"meta"`
		Data struct {
			Results []map[string]any `json:"results"`
		} `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode json: %v\n%s", err, stdout.String())
	}

	if payload.Meta.Command != "find" {
		t.Fatalf("unexpected command: %+v", payload.Meta)
	}
	if payload.Meta.Cache != "disabled" {
		t.Fatalf("expected cache disabled, got %+v", payload.Meta)
	}
	if payload.Meta.RequestID != "req_search" {
		t.Fatalf("unexpected request id: %+v", payload.Meta)
	}
	if len(payload.Data.Results) != 1 {
		t.Fatalf("unexpected results: %+v", payload.Data.Results)
	}
}

func TestCodeDefaultsToLLMFormat(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeTest(
		t,
		[]string{"code", "streaming", "--no-cache"},
		testEnv(t, map[string]string{
			"EXA_API_KEY": "test-key",
		}),
		roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/context" {
				t.Fatalf("unexpected path: %s", req.URL.Path)
			}
			return jsonResponse(`{
				"requestId":"req_context",
				"response":"Use the streaming helper and pass the request context through every API call.",
				"costDollars":0.01,
				"searchTime":1.5,
				"results":[{"title":"SDK docs","url":"https://docs.example.com"}]
			}`), nil
		}),
		&stdout,
		&stderr,
	)

	if code != 0 {
		t.Fatalf("unexpected exit code: %d\nstderr=%s", code, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Use the streaming helper") {
		t.Fatalf("expected llm output, got:\n%s", output)
	}
	if !strings.Contains(output, "SDK docs") {
		t.Fatalf("expected source list, got:\n%s", output)
	}
}

func TestGeneratedDocs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeTest(t, []string{"gen", "docs"}, testEnv(t, nil), nil, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("unexpected exit code: %d\nstderr=%s", code, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "## exa-cli find") {
		t.Fatalf("expected generated reference, got:\n%s", output)
	}
	if strings.Contains(output, "## exa-cli completion") {
		t.Fatalf("did not expect completion in generated docs, got:\n%s", output)
	}
	if strings.Contains(output, "## exa-cli help") {
		t.Fatalf("did not expect help in generated docs, got:\n%s", output)
	}
}

func TestMCPPrintDoesNotEmbedKeyByDefault(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeTest(t, []string{"mcp", "print", "codex"}, testEnv(t, map[string]string{
		"EXA_API_KEY": "secret-key",
	}), nil, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("unexpected exit code: %d\nstderr=%s", code, stderr.String())
	}

	output := stdout.String()
	if strings.Contains(output, "secret-key") {
		t.Fatalf("expected keyless snippet, got:\n%s", output)
	}
	if !strings.Contains(output, "https://mcp.exa.ai/mcp?tools=") {
		t.Fatalf("expected MCP URL, got:\n%s", output)
	}
}

func TestMCPPrintEmbedKeyRequiresExplicitOptIn(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeTest(t, []string{"mcp", "print", "codex", "--embed-key"}, testEnv(t, map[string]string{
		"EXA_API_KEY": "secret-key",
	}), nil, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("unexpected exit code: %d\nstderr=%s", code, stderr.String())
	}

	if !strings.Contains(stdout.String(), "secret-key") {
		t.Fatalf("expected embedded key in snippet, got:\n%s", stdout.String())
	}
}

func TestAuthLoginFromStdin(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	configPath := filepath.Join(t.TempDir(), "config.toml")
	env := []string{
		"EXA_CLI_CONFIG=" + configPath,
		"EXA_CLI_NO_BANNER=1",
		"EXA_CLI_NO_CACHE=1",
	}

	code := executeTestWithInput(t, []string{"auth", "login", "--stdin"}, env, nil, strings.NewReader("stdin-key\n"), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("unexpected exit code: %d\nstderr=%s", code, stderr.String())
	}

	data, err := io.ReadAll(mustOpen(t, configPath))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(data), "api_key = 'stdin-key'") {
		t.Fatalf("expected stored key, got:\n%s", string(data))
	}
}

func TestVersionSkipsHeavyRuntime(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeTest(t, []string{"version"}, []string{
		"EXA_CLI_CONFIG=/dev/null/not-a-dir/config.toml",
		"EXA_CLI_NO_BANNER=1",
	}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected version to skip heavy runtime, got code=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "version: test") {
		t.Fatalf("unexpected version output:\n%s", stdout.String())
	}
}

func TestVersionHonorsFormatFlagOnLightweightRuntime(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeTest(t, []string{"version", "--format", "json"}, []string{
		"EXA_CLI_CONFIG=/dev/null/not-a-dir/config.toml",
		"EXA_CLI_NO_BANNER=1",
	}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected JSON version output, got code=%d stderr=%s", code, stderr.String())
	}

	var payload struct {
		Meta struct {
			Command string `json:"command"`
			Format  string `json:"format"`
		} `json:"meta"`
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode json: %v\n%s", err, stdout.String())
	}
	if payload.Meta.Command != "version" || payload.Meta.Format != "json" {
		t.Fatalf("unexpected meta: %+v", payload.Meta)
	}
	if payload.Data["version"] != "test" {
		t.Fatalf("unexpected payload: %+v", payload.Data)
	}
}

func TestVersionRejectsInvalidFormatOnLightweightRuntime(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeTest(t, []string{"version", "--format", "yaml"}, []string{
		"EXA_CLI_CONFIG=/dev/null/not-a-dir/config.toml",
		"EXA_CLI_NO_BANNER=1",
	}, nil, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "invalid format") {
		t.Fatalf("expected validation error, got: %s", stderr.String())
	}
}

func TestHomeHonorsExplicitJSONFormat(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeTest(t, []string{"--format", "json"}, testEnv(t, nil), nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("unexpected exit code: %d\nstderr=%s", code, stderr.String())
	}

	var payload struct {
		Meta struct {
			Command string `json:"command"`
			Format  string `json:"format"`
		} `json:"meta"`
		Data struct {
			AuthSource     string   `json:"auth_source"`
			DefaultFormat  string   `json:"default_format"`
			DefaultProfile string   `json:"default_profile"`
			Workflows      []string `json:"workflows"`
			NextSteps      []string `json:"next_steps"`
		} `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode json: %v\n%s", err, stdout.String())
	}
	if payload.Meta.Command != "home" || payload.Meta.Format != "json" {
		t.Fatalf("unexpected meta: %+v", payload.Meta)
	}
	if payload.Data.AuthSource != "none" || payload.Data.DefaultFormat != "table" || payload.Data.DefaultProfile != "balanced" {
		t.Fatalf("unexpected home payload: %+v", payload.Data)
	}
	if len(payload.Data.Workflows) == 0 || len(payload.Data.NextSteps) == 0 {
		t.Fatalf("expected workflows and next steps, got %+v", payload.Data)
	}
}

func TestDoctorReportsConfiguredDefaultsInsteadOfInvocationFormat(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeTest(t, []string{"doctor", "--format", "json"}, testEnv(t, nil), nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("unexpected exit code: %d\nstderr=%s", code, stderr.String())
	}

	var payload struct {
		Meta struct {
			Format string `json:"format"`
		} `json:"meta"`
		Data struct {
			DefaultFormat  string `json:"default_format"`
			DefaultProfile string `json:"default_profile"`
			CacheEnabled   bool   `json:"cache_enabled"`
		} `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode json: %v\n%s", err, stdout.String())
	}
	if payload.Meta.Format != "json" {
		t.Fatalf("unexpected meta: %+v", payload.Meta)
	}
	if payload.Data.DefaultFormat != "table" || payload.Data.DefaultProfile != "balanced" {
		t.Fatalf("expected configured defaults, got %+v", payload.Data)
	}
	if payload.Data.CacheEnabled {
		t.Fatalf("expected cache to be disabled by test env, got %+v", payload.Data)
	}
}

func TestDoctorSkipsCacheOpenForBrokenCachePath(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte("cache_path = \"/dev/null/not-a-real-dir/cache.db\"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	code := executeTest(t, []string{"doctor", "--format", "json"}, []string{
		"EXA_CLI_CONFIG=" + configPath,
		"EXA_CLI_NO_BANNER=1",
	}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected doctor to skip opening cache, got code=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"/dev/null/not-a-real-dir/cache.db"`) {
		t.Fatalf("expected broken cache path to be reported without opening it, got:\n%s", stdout.String())
	}
}

func TestAuthStatusWarnsWhenEnvOverridesStoredKey(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte("api_key = 'stored-key'\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	code := executeTest(t, []string{"auth", "status", "--format", "json"}, []string{
		"EXA_CLI_CONFIG=" + configPath,
		"EXA_CLI_NO_BANNER=1",
		"EXA_CLI_NO_CACHE=1",
		"EXA_API_KEY=env-key",
	}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("unexpected exit code: %d stderr=%s", code, stderr.String())
	}

	var payload struct {
		Data struct {
			AuthSource  string `json:"auth_source"`
			EnvOverride bool   `json:"env_override"`
			Warning     string `json:"warning"`
		} `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode json: %v\n%s", err, stdout.String())
	}
	if payload.Data.AuthSource != "env" || !payload.Data.EnvOverride || payload.Data.Warning == "" {
		t.Fatalf("expected env override warning, got %+v", payload.Data)
	}
}

func TestInvalidFormatIsRejected(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeTest(t, []string{"doctor", "--format", "yaml"}, testEnv(t, nil), nil, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected exit code 2 for invalid format, got %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "invalid format") {
		t.Fatalf("expected validation error, got: %s", stderr.String())
	}
}

func TestRawRequestReadsFromInjectedStdin(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var requestBody string

	code := executeTestWithInput(
		t,
		[]string{"raw", "request", "--method", "POST", "--path", "/search", "--input", "-"},
		testEnv(t, map[string]string{"EXA_API_KEY": "test-key"}),
		roundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("read request body: %v", err)
			}
			requestBody = string(body)
			return jsonResponse(`{"requestId":"raw_req","ok":true}`), nil
		}),
		strings.NewReader(`{"query":"injected stdin works"}`),
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("unexpected exit code: %d\nstderr=%s", code, stderr.String())
	}
	if !strings.Contains(requestBody, "injected stdin works") {
		t.Fatalf("expected request body to come from injected stdin, got: %s", requestBody)
	}
}

func TestReferenceDocSnapshot(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeTest(t, []string{"gen", "docs"}, testEnv(t, nil), nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("unexpected exit code: %d\nstderr=%s", code, stderr.String())
	}

	want, err := os.ReadFile(filepath.Join("..", "..", "docs", "REFERENCE.md"))
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	if strings.TrimSpace(stdout.String()) != strings.TrimSpace(string(want)) {
		t.Fatalf("generated reference docs differ from docs/REFERENCE.md; run make docs")
	}
}

func testEnv(t *testing.T, extra map[string]string) []string {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config.toml")
	env := []string{
		"EXA_CLI_CONFIG=" + configPath,
		"EXA_CLI_NO_BANNER=1",
		"EXA_CLI_NO_CACHE=1",
	}
	for key, value := range extra {
		env = append(env, key+"="+value)
	}
	return env
}

func executeTest(t *testing.T, args []string, env []string, transport roundTripFunc, stdout, stderr *bytes.Buffer) int {
	return executeTestWithInput(t, args, env, transport, bytes.NewBuffer(nil), stdout, stderr)
}

func executeTestWithInput(t *testing.T, args []string, env []string, transport roundTripFunc, input io.Reader, stdout, stderr *bytes.Buffer) int {
	t.Helper()

	instance := &App{
		in:      input,
		out:     stdout,
		errOut:  stderr,
		env:     envMap(env),
		version: version.Info{Version: "test"},
	}

	root := instance.newRootCmd()
	instance.root = root
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)

	if transport != nil {
		root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			if err := instance.initRuntime(cmd); err != nil {
				return err
			}
			instance.client.HTTP.Transport = transport
			return nil
		}
	}

	err := root.ExecuteContext(t.Context())
	if err == nil {
		_ = instance.close()
		return 0
	}
	_ = instance.close()

	var codeErr *cliError
	if errors.As(err, &codeErr) {
		stderr.WriteString(codeErr.Err.Error())
		return codeErr.Code
	}
	stderr.WriteString(err.Error())
	return 1
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(body)),
	}
}

func mustOpen(t *testing.T, path string) io.ReadCloser {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	return file
}
