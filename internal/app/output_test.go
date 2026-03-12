package app

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRenderLLMVerbose(t *testing.T) {
	envelope := Envelope{
		Meta: Meta{
			Command:   "code",
			RequestID: "req_123",
			Cache:     "hit",
		},
		Data: map[string]any{
			"context": "Use the request context for every call.",
			"results": []any{
				map[string]any{
					"title": "Streaming docs",
					"url":   "https://example.com/streaming",
				},
			},
		},
	}

	output := renderLLM(envelope, true)
	if !strings.Contains(output, "Use the request context for every call.") {
		t.Fatalf("expected narrative text, got:\n%s", output)
	}
	if !strings.Contains(output, "Sources:") || !strings.Contains(output, "Streaming docs") {
		t.Fatalf("expected source list, got:\n%s", output)
	}
	if !strings.Contains(output, "request_id: req_123") || !strings.Contains(output, "cache: hit") {
		t.Fatalf("expected verbose metadata, got:\n%s", output)
	}
}

func TestRenderLLMUsesResponseField(t *testing.T) {
	envelope := Envelope{
		Meta: Meta{Command: "code"},
		Data: map[string]any{
			"response": "Grounded implementation notes from Exa Context.",
		},
	}

	output := renderLLM(envelope, false)
	if !strings.Contains(output, "Grounded implementation notes from Exa Context.") {
		t.Fatalf("expected response field to render, got:\n%s", output)
	}
}

func TestRenderMarkdownFallbackUsesReadableKeys(t *testing.T) {
	envelope := Envelope{
		Meta: Meta{Command: "mcp doctor"},
		Data: map[string]any{
			"mcp_url":            "https://mcp.exa.ai/mcp",
			"api_key_configured": false,
		},
	}

	output := renderMarkdown(envelope, false)
	if !strings.Contains(output, "## MCP Doctor") {
		t.Fatalf("expected title-cased command heading, got:\n%s", output)
	}
	if !strings.Contains(output, "- `mcp_url`: https://mcp.exa.ai/mcp") {
		t.Fatalf("expected readable map fallback, got:\n%s", output)
	}
	if !strings.Contains(output, "- `api_key_configured`: false") {
		t.Fatalf("expected boolean fallback, got:\n%s", output)
	}
}

func TestRenderTablePrintsReadableStrings(t *testing.T) {
	envelope := Envelope{
		Meta: Meta{Command: "config show"},
		Data: map[string]any{
			"cache_path": "/tmp/exa-cli/cache.db",
			"format":     "json",
		},
	}

	output := renderTable(envelope)
	if !strings.Contains(output, "cache_path: /tmp/exa-cli/cache.db") {
		t.Fatalf("expected plain string output, got:\n%s", output)
	}
	if strings.Contains(output, "\"/tmp/exa-cli/cache.db\"") {
		t.Fatalf("did not expect JSON-quoted string output, got:\n%s", output)
	}
}

func TestRenderMarkdownReadPrefersSummaryOverText(t *testing.T) {
	envelope := Envelope{
		Meta: Meta{Command: "read"},
		Data: map[string]any{
			"results": []any{
				map[string]any{
					"title":   "Search - Exa",
					"url":     "https://exa.ai/docs/reference/search",
					"summary": "Short summary first.",
					"text":    strings.Repeat("Long page text ", 80),
				},
			},
		},
	}

	output := renderMarkdown(envelope, false)
	if !strings.Contains(output, "Short summary first.") {
		t.Fatalf("expected summary to render, got:\n%s", output)
	}
	if strings.Contains(output, "Long page text Long page text Long page text Long page text") {
		t.Fatalf("did not expect full text to take precedence, got:\n%s", output)
	}
}

func TestWriteEnvelopeJSONLContract(t *testing.T) {
	var stdout bytes.Buffer
	app := &App{
		out: &stdout,
		options: GlobalOptions{
			Format:       "jsonl",
			FormatSource: "flag",
		},
	}

	err := app.writeEnvelope(map[string]any{
		"results": []any{
			map[string]any{"title": "Result One"},
			map[string]any{"title": "Result Two"},
		},
	}, Meta{Command: "find"}, "table")
	if err != nil {
		t.Fatalf("write envelope: %v", err)
	}

	scanner := bufio.NewScanner(&stdout)
	var lines []map[string]any
	for scanner.Scan() {
		var record map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			t.Fatalf("decode jsonl line: %v", err)
		}
		lines = append(lines, record)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan jsonl: %v", err)
	}
	if len(lines) != 3 {
		t.Fatalf("expected 3 JSONL records, got %d", len(lines))
	}
	if lines[0]["type"] != "meta" {
		t.Fatalf("expected first JSONL line to be meta, got %+v", lines[0])
	}
	if lines[1]["type"] != "record" || lines[2]["type"] != "record" {
		t.Fatalf("expected record lines after meta, got %+v", lines)
	}
}

func TestRenderLLMIncludesResultSnippets(t *testing.T) {
	envelope := Envelope{
		Meta: Meta{Command: "read"},
		Data: map[string]any{
			"results": []any{
				map[string]any{
					"title":   "Search - Exa",
					"url":     "https://exa.ai/docs/reference/search",
					"summary": "API documentation for Exa search endpoint.",
				},
			},
		},
	}

	output := renderLLM(envelope, false)
	if !strings.Contains(output, "API documentation for Exa search endpoint.") {
		t.Fatalf("expected LLM format to include per-result summary, got:\n%s", output)
	}
}

func TestRenderLLMIncludesHighlightsWhenNoSnippet(t *testing.T) {
	envelope := Envelope{
		Meta: Meta{Command: "find"},
		Data: map[string]any{
			"results": []any{
				map[string]any{
					"title":      "Result Page",
					"url":        "https://example.com",
					"highlights": []any{"important context from the page"},
				},
			},
		},
	}

	output := renderLLM(envelope, false)
	if !strings.Contains(output, "important context from the page") {
		t.Fatalf("expected LLM format to include highlights, got:\n%s", output)
	}
}

func TestValidateCategoryFiltersAllowsExcludeDomainWithoutCategory(t *testing.T) {
	err := validateCategoryFilters("", nil, []string{"example.com"})
	if err != nil {
		t.Fatalf("expected exclude-domain to be allowed without category, got: %v", err)
	}
}

func TestValidateCategoryFiltersBlocksExcludeDomainForCompany(t *testing.T) {
	err := validateCategoryFilters("company", nil, []string{"example.com"})
	if err == nil {
		t.Fatal("expected company category to reject exclude-domain")
	}
	if !strings.Contains(err.Error(), "company search") {
		t.Fatalf("expected company-specific error message, got: %v", err)
	}
}
