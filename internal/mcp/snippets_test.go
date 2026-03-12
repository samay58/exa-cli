package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRenderCodexSnippetIsKeylessByDefault(t *testing.T) {
	output, err := Render("codex", "", "", DefaultTools())
	if err != nil {
		t.Fatalf("render codex snippet: %v", err)
	}

	if !strings.Contains(output, "codex mcp add exa --url https://mcp.exa.ai/mcp?tools=") {
		t.Fatalf("expected hosted MCP URL, got:\n%s", output)
	}
	if strings.Contains(output, "exaApiKey=") {
		t.Fatalf("did not expect embedded API key, got:\n%s", output)
	}
}

func TestRenderCursorSnippetEmbedsKeyWhenRequested(t *testing.T) {
	output, err := Render("cursor", "https://mcp.example.com/mcp", "secret-key", []string{"web_search_exa"})
	if err != nil {
		t.Fatalf("render cursor snippet: %v", err)
	}

	var payload struct {
		MCPServers map[string]struct {
			URL string `json:"url"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("decode cursor json: %v\n%s", err, output)
	}

	url := payload.MCPServers["exa"].URL
	if !strings.Contains(url, "https://mcp.example.com/mcp?") {
		t.Fatalf("expected custom MCP URL, got %q", url)
	}
	if !strings.Contains(url, "exaApiKey=secret-key") {
		t.Fatalf("expected embedded key in URL, got %q", url)
	}
}

func TestRenderRejectsUnsupportedTarget(t *testing.T) {
	_, err := Render("not-a-client", "", "", nil)
	if err == nil {
		t.Fatal("expected unsupported client error")
	}
	if !strings.Contains(err.Error(), "unsupported client") {
		t.Fatalf("unexpected error: %v", err)
	}
}
