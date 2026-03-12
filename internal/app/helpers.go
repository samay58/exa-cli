package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/samaydhawan/exa-cli/internal/config"
	"github.com/samaydhawan/exa-cli/internal/exa"
)

func envMap(items []string) map[string]string {
	result := make(map[string]string, len(items))
	for _, item := range items {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			continue
		}
		result[parts[0]] = parts[1]
	}
	return result
}

func truthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func profileToSearchType(profile string) string {
	switch strings.ToLower(strings.TrimSpace(profile)) {
	case "instant":
		return "instant"
	case "fast":
		return "fast"
	case "deep":
		return "deep"
	case "reasoned":
		return "deep-reasoning"
	default:
		return "auto"
	}
}

func validateFormat(value string) error {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "table", "markdown", "json", "jsonl", "llm":
		return nil
	default:
		return fmt.Errorf("invalid format %q; use table, markdown, json, jsonl, or llm", value)
	}
}

func validateProfile(value string) error {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "instant", "fast", "balanced", "deep", "reasoned":
		return nil
	default:
		return fmt.Errorf("invalid profile %q; use instant, fast, balanced, deep, or reasoned", value)
	}
}

func validateTokens(value string) error {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "dynamic") {
		return nil
	}
	number, err := parseInt(value)
	if err != nil {
		return fmt.Errorf("invalid token budget %q; use dynamic or a positive integer", value)
	}
	if number <= 0 {
		return fmt.Errorf("invalid token budget %q; use dynamic or a positive integer", value)
	}
	return nil
}

func normalizeTokensNum(value string) any {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "dynamic") {
		return "dynamic"
	}
	if number, err := parseInt(value); err == nil {
		return number
	}
	return value
}

func parseInt(value string) (int, error) {
	var result int
	_, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &result)
	if err != nil {
		return 0, fmt.Errorf("invalid integer %q", value)
	}
	return result, nil
}

func firstString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func asString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}

func readFloat(payload map[string]any, key string) float64 {
	if payload == nil {
		return 0
	}
	value, ok := payload[key]
	if !ok {
		return 0
	}
	return coerceFloat(value)
}

func wrap(code int, err error) error {
	if err == nil {
		return nil
	}
	return &cliError{Code: code, Err: err}
}

func normalizeAPIError(err error) error {
	var apiErr *exa.APIError
	if errors.As(err, &apiErr) {
		switch {
		case apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden:
			return wrap(3, err)
		case apiErr.StatusCode == http.StatusTooManyRequests:
			return wrap(5, err)
		default:
			return wrap(4, err)
		}
	}
	return wrap(4, err)
}

func validateCategoryFilters(category string, includeDomains, excludeDomains []string) error {
	category = strings.ToLower(strings.TrimSpace(category))
	switch category {
	case "company":
		if len(excludeDomains) > 0 {
			return errors.New("company search does not support --exclude-domain in Exa's current API")
		}
	case "people":
		if len(excludeDomains) > 0 {
			return errors.New("people search does not support --exclude-domain in Exa's current API")
		}
		for _, domain := range includeDomains {
			if !strings.Contains(strings.ToLower(domain), "linkedin.") {
				return errors.New("people search only accepts LinkedIn domains via --include-domain")
			}
		}
	}
	return nil
}

func readInputPayload(path string, fallback any, stdin io.Reader) (any, error) {
	if strings.TrimSpace(path) == "" {
		return fallback, nil
	}

	var data []byte
	var err error
	if path == "-" {
		data, err = io.ReadAll(stdin)
	} else {
		data, err = os.ReadFile(filepath.Clean(path))
	}
	if err != nil {
		return nil, err
	}
	var payload any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return payload, nil
}

func maskSecret(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 8 {
		return "********"
	}
	return value[:4] + strings.Repeat("*", len(value)-8) + value[len(value)-4:]
}

func configToMap(cfg config.Config) map[string]any {
	view := map[string]any{
		"api_key":           "",
		"base_url":          cfg.BaseURL,
		"format":            cfg.Format,
		"profile":           cfg.Profile,
		"no_banner":         cfg.NoBanner,
		"cache_ttl_minutes": cfg.CacheTTLMinutes,
		"cache_path":        cfg.CachePath,
		"mcp_url":           cfg.MCPURL,
	}
	if cfg.APIKey != "" {
		view["api_key"] = maskSecret(cfg.APIKey)
	}
	return view
}

func jsonlRecords(data any) []any {
	if data == nil {
		return nil
	}
	if slice, ok := data.([]any); ok {
		return slice
	}
	if payload, ok := data.(map[string]any); ok {
		if results, ok := payload["results"].([]any); ok {
			return results
		}
		if citations, ok := payload["citations"].([]any); ok {
			return citations
		}
		if statuses, ok := payload["statuses"].([]any); ok {
			return statuses
		}
		if records, ok := payload["data"].([]any); ok {
			return records
		}
		if tasks, ok := payload["tasks"].([]any); ok {
			return tasks
		}
	}
	return []any{data}
}

func coerceFloat(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case json.Number:
		v, _ := typed.Float64()
		return v
	case string:
		value := strings.TrimSpace(typed)
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			return v
		}
		if strings.HasPrefix(value, "{") {
			var payload map[string]any
			if err := json.Unmarshal([]byte(value), &payload); err == nil {
				return coerceFloat(payload)
			}
		}
		return 0
	case map[string]any:
		if total, ok := typed["total"]; ok {
			return coerceFloat(total)
		}
		if value, ok := typed["value"]; ok {
			return coerceFloat(value)
		}
	}
	return 0
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
