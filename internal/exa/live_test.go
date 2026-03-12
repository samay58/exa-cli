package exa_test

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/samaydhawan/exa-cli/internal/config"
	"github.com/samaydhawan/exa-cli/internal/exa"
)

func TestLiveCoreEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live Exa API checks in short mode")
	}
	if os.Getenv("EXA_LIVE_TESTS") != "1" {
		t.Skip("set EXA_LIVE_TESTS=1 to run live Exa API checks")
	}

	apiKey := strings.TrimSpace(os.Getenv("EXA_API_KEY"))
	if apiKey == "" {
		configPath, err := config.DefaultConfigPath()
		if err == nil {
			cfg, loadErr := config.Load(configPath)
			if loadErr == nil {
				apiKey = strings.TrimSpace(cfg.APIKey)
			}
		}
	}
	if apiKey == "" {
		t.Skip("set EXA_API_KEY or store a key with `exa-cli auth login` to run live Exa API checks")
	}

	client := exa.New(strings.TrimSpace(os.Getenv("EXA_BASE_URL")), apiKey)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	searchPayload, searchMeta, err := client.Search(ctx, exa.SearchRequest{
		Query:      "Exa search API docs",
		Type:       "fast",
		NumResults: 1,
	})
	if err != nil {
		t.Fatalf("live search: %v", err)
	}
	if searchMeta.StatusCode != http.StatusOK || resultCount(searchPayload) == 0 {
		t.Fatalf("unexpected live search payload: meta=%+v payload=%+v", searchMeta, searchPayload)
	}

	answerPayload, answerMeta, err := client.Answer(ctx, exa.AnswerRequest{
		Query: "How does Exa search differ from answer?",
	})
	if err != nil {
		t.Fatalf("live answer: %v", err)
	}
	if answerMeta.StatusCode != http.StatusOK || strings.TrimSpace(firstNonEmptyString(answerPayload, "answer")) == "" {
		t.Fatalf("unexpected live answer payload: meta=%+v payload=%+v", answerMeta, answerPayload)
	}

	contextPayload, contextMeta, err := client.Context(ctx, exa.ContextRequest{
		Query:     "Go Cobra command tree patterns",
		TokensNum: 1500,
	})
	if err != nil {
		t.Fatalf("live context: %v", err)
	}
	if contextMeta.StatusCode != http.StatusOK || strings.TrimSpace(firstNonEmptyString(contextPayload, "context", "response")) == "" {
		t.Fatalf("unexpected live context payload: meta=%+v payload=%+v", contextMeta, contextPayload)
	}

	contentsPayload, contentsMeta, err := client.Contents(ctx, exa.ContentsRequest{
		URLs:    []string{"https://exa.ai/docs/reference/search"},
		Summary: map[string]any{"query": "Summarize the page for quick verification."},
	})
	if err != nil {
		t.Fatalf("live contents: %v", err)
	}
	if contentsMeta.StatusCode != http.StatusOK || len(contentsPayload) == 0 {
		t.Fatalf("unexpected live contents payload: meta=%+v payload=%+v", contentsMeta, contentsPayload)
	}

	researchPayload, researchMeta, err := client.ResearchList(ctx, 1)
	if err != nil {
		t.Fatalf("live research list: %v", err)
	}
	if researchMeta.StatusCode != http.StatusOK || len(researchPayload) == 0 {
		t.Fatalf("unexpected live research list payload: meta=%+v payload=%+v", researchMeta, researchPayload)
	}
}

func resultCount(payload map[string]any) int {
	if payload == nil {
		return 0
	}
	if results, ok := payload["results"].([]any); ok {
		return len(results)
	}
	if tasks, ok := payload["tasks"].([]any); ok {
		return len(tasks)
	}
	if data, ok := payload["data"].([]any); ok {
		return len(data)
	}
	return len(payload)
}

func firstNonEmptyString(payload map[string]any, keys ...string) string {
	if payload == nil {
		return ""
	}
	for _, key := range keys {
		if value, ok := payload[key].(string); ok && strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
