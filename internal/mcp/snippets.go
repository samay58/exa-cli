package mcp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

func Render(target, baseURL, apiKey string, tools []string) (string, error) {
	if baseURL == "" {
		baseURL = "https://mcp.exa.ai/mcp"
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	query := u.Query()
	if len(tools) > 0 {
		query.Set("tools", strings.Join(tools, ","))
	}
	if apiKey != "" {
		query.Set("exaApiKey", apiKey)
	}
	u.RawQuery = query.Encode()
	remoteURL := u.String()

	switch target {
	case "codex":
		return fmt.Sprintf("codex mcp add exa --url %s", remoteURL), nil
	case "claude-code":
		return fmt.Sprintf("claude mcp add --transport http exa %s", remoteURL), nil
	case "cursor":
		return marshal(map[string]any{
			"mcpServers": map[string]any{
				"exa": map[string]any{
					"url": remoteURL,
				},
			},
		})
	case "generic":
		return marshal(map[string]any{
			"mcpServers": map[string]any{
				"exa": map[string]any{
					"url": remoteURL,
				},
			},
		})
	default:
		return "", fmt.Errorf("unsupported client %q", target)
	}
}

func DefaultTools() []string {
	return []string{
		"web_search_exa",
		"get_code_context_exa",
		"company_research_exa",
	}
}

func FullTools() []string {
	return []string{
		"web_search_exa",
		"web_search_advanced_exa",
		"get_code_context_exa",
		"crawling_exa",
		"company_research_exa",
		"people_search_exa",
		"deep_researcher_start",
		"deep_researcher_check",
		"deep_search_exa",
	}
}

func marshal(v any) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
