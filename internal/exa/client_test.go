package exa_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/samaydhawan/exa-cli/internal/exa"
)

func TestClientSearchSetsAuthAndExtractsRequestID(t *testing.T) {
	client := exa.New("https://api.exa.ai", "test-key")
	client.HTTP.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/search" {
			t.Fatalf("unexpected path: %s", req.URL.Path)
		}
		if got := req.Header.Get("x-api-key"); got != "test-key" {
			t.Fatalf("missing x-api-key header: %q", got)
		}
		if got := req.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("missing Authorization header: %q", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"X-Request-Id": []string{"req_from_header"}},
			Body:       io.NopCloser(strings.NewReader(`{"results":[{"title":"ok"}]}`)),
		}, nil
	})

	payload, meta, err := client.Search(t.Context(), exa.SearchRequest{Query: "golang"})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if meta.RequestID != "req_from_header" {
		t.Fatalf("unexpected request id: %+v", meta)
	}
	if len(payload["results"].([]any)) != 1 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestClientReturnsAPIError(t *testing.T) {
	client := exa.New("https://api.exa.ai", "test-key")
	client.HTTP.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     http.Header{},
			Body:       io.NopCloser(strings.NewReader(`{"message":"rate limited"}`)),
		}, nil
	})

	_, _, err := client.Search(t.Context(), exa.SearchRequest{Query: "golang"})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *exa.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("unexpected status: %+v", apiErr)
	}
	if !strings.Contains(apiErr.Error(), "rate limited") {
		t.Fatalf("unexpected error: %v", apiErr)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
