package app

import (
	"encoding/json"
	"fmt"
	"strings"
)

func renderLLM(envelope Envelope, verbose bool) string {
	var lines []string

	if payload, ok := envelope.Data.(map[string]any); ok {
		if text := bestNarrativeText(payload); text != "" {
			lines = append(lines, text)
		}
		if results, ok := payload["results"].([]any); ok && len(results) > 0 {
			if len(lines) > 0 {
				lines = append(lines, "")
			}
			lines = append(lines, "Sources:")
			for idx, item := range results {
				result, ok := item.(map[string]any)
				if !ok {
					continue
				}
				title := firstString(asString(result["title"]), asString(result["url"]))
				url := asString(result["url"])
				lines = append(lines, fmt.Sprintf("%d. %s", idx+1, title))
				if url != "" {
					lines = append(lines, "   "+url)
				}
				if snippet := resultSnippet(envelope.Meta.Command, result, 1200); snippet != "" {
					lines = append(lines, "   "+snippet)
				} else if highlights, ok := result["highlights"].([]any); ok && len(highlights) > 0 {
					lines = append(lines, "   "+condense(strings.Join(stringSlice(highlights), " | "), 1200))
				}
			}
		}
		if grounding := renderGrounding(payload); len(grounding) > 0 {
			lines = append(lines, "")
			lines = append(lines, grounding...)
		}
		if verbose {
			lines = append(lines, "")
			lines = append(lines, fmt.Sprintf("request_id: %s", envelope.Meta.RequestID))
			lines = append(lines, fmt.Sprintf("cache: %s", envelope.Meta.Cache))
		}
		if len(lines) == 0 {
			lines = append(lines, condense(mustJSON(payload), 2000))
		}
		return strings.Join(lines, "\n")
	}

	return renderMarkdown(envelope, verbose)
}

func renderMarkdown(envelope Envelope, verbose bool) string {
	if payload, ok := envelope.Data.(map[string]any); ok {
		var lines []string
		lines = append(lines, fmt.Sprintf("## %s", commandTitle(envelope.Meta.Command)))

		if text := bestNarrativeText(payload); text != "" {
			lines = append(lines, "", text)
		}

		if results, ok := payload["results"].([]any); ok && len(results) > 0 {
			lines = append(lines, "", "### Results")
			for _, item := range results {
				result, ok := item.(map[string]any)
				if !ok {
					continue
				}
				title := firstString(asString(result["title"]), asString(result["url"]))
				url := asString(result["url"])
				lines = append(lines, fmt.Sprintf("- %s", title))
				if url != "" {
					lines = append(lines, fmt.Sprintf("  %s", url))
				}
				limit := 2200
				if envelope.Meta.Command == "read" {
					limit = 900
				}
				if snippet := resultSnippet(envelope.Meta.Command, result, limit); snippet != "" {
					lines = append(lines, "  "+snippet)
				}
				if highlights, ok := result["highlights"].([]any); ok && len(highlights) > 0 {
					lines = append(lines, "  Highlights: "+strings.Join(stringSlice(highlights), " | "))
				}
			}
		}

		if grounding := renderGrounding(payload); len(grounding) > 0 {
			lines = append(lines, "", "### Citations")
			for _, g := range grounding[1:] {
				lines = append(lines, "- "+strings.TrimSpace(g))
			}
		}

		if outputs, ok := payload["outputs"].([]any); ok && len(outputs) > 0 {
			lines = append(lines, "", "### Outputs")
			for _, item := range outputs {
				lines = append(lines, "- "+condense(mustJSON(item), 240))
			}
		}

		if len(lines) == 1 {
			for _, key := range sortedKeys(payload) {
				lines = append(lines, fmt.Sprintf("- `%s`: %s", key, condense(displayValue(payload[key]), 600)))
			}
		}

		if verbose && envelope.Meta.RequestID != "" {
			lines = append(lines, "", fmt.Sprintf("_request_id: %s_", envelope.Meta.RequestID))
		}

		return strings.Join(lines, "\n")
	}
	return condense(mustJSON(envelope.Data), 1200)
}

func renderTable(envelope Envelope) string {
	var lines []string
	header := strings.TrimSpace(envelope.Meta.Command)
	if header == "" {
		header = "exa-cli"
	}
	lines = append(lines, header)
	lines = append(lines, strings.Repeat("=", len(header)))

	if envelope.Meta.Profile != "" {
		lines = append(lines, "profile: "+envelope.Meta.Profile)
	}
	if envelope.Meta.Cache != "" {
		lines = append(lines, "cache: "+envelope.Meta.Cache)
	}
	if envelope.Meta.RequestID != "" {
		lines = append(lines, "request: "+envelope.Meta.RequestID)
	}
	if envelope.Meta.CostDollars > 0 {
		lines = append(lines, fmt.Sprintf("cost: $%.6f", envelope.Meta.CostDollars))
	}
	if envelope.Meta.SearchTime > 0 {
		lines = append(lines, fmt.Sprintf("searchTime: %.2fs", envelope.Meta.SearchTime))
	}

	payload, ok := envelope.Data.(map[string]any)
	if !ok {
		lines = append(lines, "", mustJSON(envelope.Data))
		return strings.Join(lines, "\n")
	}

	if text := bestNarrativeText(payload); text != "" {
		lines = append(lines, "", condense(text, 800))
	}

	if grounding := renderGrounding(payload); len(grounding) > 0 {
		lines = append(lines, "")
		lines = append(lines, grounding...)
	}

	if results, ok := payload["results"].([]any); ok && len(results) > 0 {
		lines = append(lines, "", "results:")
		for idx, item := range results {
			result, ok := item.(map[string]any)
			if !ok {
				continue
			}
			title := firstString(asString(result["title"]), asString(result["url"]), fmt.Sprintf("result %d", idx+1))
			url := asString(result["url"])
			lines = append(lines, fmt.Sprintf("%d. %s", idx+1, title))
			if url != "" {
				lines = append(lines, "   "+url)
			}
			if published := asString(result["publishedDate"]); published != "" {
				lines = append(lines, "   published: "+published)
			}
			if summary := resultSnippet(envelope.Meta.Command, result, 600); summary != "" {
				lines = append(lines, "   "+summary)
			}
			if highlights, ok := result["highlights"].([]any); ok && len(highlights) > 0 {
				lines = append(lines, "   highlights: "+strings.Join(stringSlice(highlights), " | "))
			}
		}
		return strings.Join(lines, "\n")
	}

	for _, key := range sortedKeys(payload) {
		lines = append(lines, fmt.Sprintf("%s: %s", key, condense(displayValue(payload[key]), 220)))
	}
	return strings.Join(lines, "\n")
}

func bestNarrativeText(payload map[string]any) string {
	candidates := []string{
		asString(payload["answer"]),
		asString(payload["context"]),
		asString(payload["response"]),
		asString(payload["summary"]),
		asString(payload["output"]),
	}
	if outputMap, ok := payload["output"].(map[string]any); ok {
		candidates = append(candidates, asString(outputMap["content"]))
		if contentMap, ok := outputMap["content"].(map[string]any); ok {
			if data, err := json.MarshalIndent(contentMap, "", "  "); err == nil {
				candidates = append(candidates, string(data))
			}
		}
	}
	if report, ok := payload["report"].(string); ok {
		candidates = append(candidates, report)
	}
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) != "" {
			return candidate
		}
	}
	return ""
}

func stringSlice(values []any) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if text := strings.TrimSpace(asString(value)); text != "" {
			result = append(result, text)
		}
	}
	return result
}

func condense(value string, limit int) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\n", " "))
	if limit <= 0 || len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return value[:limit-3] + "..."
}

func mustJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(data)
}

func displayValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return mustJSON(value)
	}
}

func resultSnippet(command string, result map[string]any, limit int) string {
	if command == "read" {
		if summary := asString(result["summary"]); summary != "" {
			return condense(summary, limit)
		}
		if text := asString(result["text"]); text != "" {
			return condense(text, limit)
		}
	}
	if summary := asString(result["summary"]); summary != "" {
		return condense(summary, limit)
	}
	if text := asString(result["text"]); text != "" {
		return condense(text, limit)
	}
	return ""
}

func commandTitle(command string) string {
	command = strings.TrimSpace(command)
	if command == "" {
		return "exa-cli"
	}

	words := strings.Fields(command)
	for idx, word := range words {
		upper := strings.ToUpper(word)
		if isInitialism(upper) {
			words[idx] = upper
			continue
		}
		words[idx] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, " ")
}

func isInitialism(value string) bool {
	switch value {
	case "API", "CLI", "JSON", "JSONL", "LLM", "MCP", "TTL":
		return true
	default:
		return false
	}
}

func renderGrounding(payload map[string]any) []string {
	outputMap, ok := payload["output"].(map[string]any)
	if !ok {
		return nil
	}
	items, ok := outputMap["grounding"].([]any)
	if !ok || len(items) == 0 {
		return nil
	}
	var lines []string
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		field := asString(entry["fieldPath"])
		sources, ok := entry["sources"].([]any)
		if !ok || len(sources) == 0 {
			continue
		}
		var citations []string
		for _, src := range sources {
			s, ok := src.(map[string]any)
			if !ok {
				continue
			}
			url := asString(s["url"])
			confidence := asString(s["confidence"])
			if url == "" {
				continue
			}
			if confidence != "" {
				citations = append(citations, url+" ("+confidence+")")
			} else {
				citations = append(citations, url)
			}
		}
		if len(citations) > 0 && field != "" {
			lines = append(lines, fmt.Sprintf("  %s: %s", field, strings.Join(citations, ", ")))
		}
	}
	if len(lines) > 0 {
		return append([]string{"citations:"}, lines...)
	}
	return nil
}
