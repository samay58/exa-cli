package exa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	BaseURL   string
	APIKey    string
	UserAgent string
	HTTP      *http.Client
}

type SearchRequest struct {
	Query           string         `json:"query,omitempty"`
	Type            string         `json:"type,omitempty"`
	Category        string         `json:"category,omitempty"`
	NumResults      int            `json:"numResults,omitempty"`
	IncludeDomains  []string       `json:"includeDomains,omitempty"`
	ExcludeDomains  []string       `json:"excludeDomains,omitempty"`
	AdditionalQuery []string       `json:"additionalQueries,omitempty"`
	SystemPrompt    string         `json:"systemPrompt,omitempty"`
	OutputSchema    map[string]any `json:"outputSchema,omitempty"`
	Contents        map[string]any `json:"contents,omitempty"`
}

type AnswerRequest struct {
	Query        string         `json:"query,omitempty"`
	Stream       bool           `json:"stream,omitempty"`
	OutputSchema map[string]any `json:"outputSchema,omitempty"`
}

type ContentsRequest struct {
	URLs             []string       `json:"urls,omitempty"`
	IDs              []string       `json:"ids,omitempty"`
	Text             any            `json:"text,omitempty"`
	Highlights       map[string]any `json:"highlights,omitempty"`
	Summary          map[string]any `json:"summary,omitempty"`
	MaxAgeHours      *int           `json:"maxAgeHours,omitempty"`
	LivecrawlTimeout int            `json:"livecrawlTimeout,omitempty"`
}

type ContextRequest struct {
	Query     string `json:"query,omitempty"`
	TokensNum any    `json:"tokensNum,omitempty"`
}

type ResearchCreateRequest struct {
	Instructions string         `json:"instructions,omitempty"`
	Model        string         `json:"model,omitempty"`
	OutputSchema map[string]any `json:"outputSchema,omitempty"`
}

type RequestMeta struct {
	StatusCode int    `json:"statusCode"`
	RequestID  string `json:"requestId,omitempty"`
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("exa api error (%d): %s", e.StatusCode, e.Message)
}

func New(baseURL, apiKey string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.exa.ai"
	}
	return &Client{
		BaseURL:   baseURL,
		APIKey:    apiKey,
		UserAgent: "exa-cli/dev",
		HTTP: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

func (c *Client) Search(ctx context.Context, req SearchRequest) (map[string]any, RequestMeta, error) {
	return c.post(ctx, "/search", req)
}

func (c *Client) Answer(ctx context.Context, req AnswerRequest) (map[string]any, RequestMeta, error) {
	return c.post(ctx, "/answer", req)
}

func (c *Client) Contents(ctx context.Context, req ContentsRequest) (map[string]any, RequestMeta, error) {
	return c.post(ctx, "/contents", req)
}

func (c *Client) Context(ctx context.Context, req ContextRequest) (map[string]any, RequestMeta, error) {
	return c.post(ctx, "/context", req)
}

func (c *Client) ResearchCreate(ctx context.Context, req ResearchCreateRequest) (map[string]any, RequestMeta, error) {
	return c.post(ctx, "/research/v1", req)
}

func (c *Client) ResearchGet(ctx context.Context, researchID string, events bool) (map[string]any, RequestMeta, error) {
	path := "/research/v1/" + url.PathEscape(researchID)
	if events {
		path += "?events=true"
	}
	return c.get(ctx, path)
}

func (c *Client) ResearchList(ctx context.Context, limit int) (map[string]any, RequestMeta, error) {
	if limit <= 0 {
		limit = 10
	}
	return c.get(ctx, fmt.Sprintf("/research/v1?limit=%d", limit))
}

func (c *Client) ResearchCancel(ctx context.Context, researchID string) (map[string]any, RequestMeta, error) {
	path := "/research/v1/" + url.PathEscape(researchID)
	return c.delete(ctx, path)
}

func (c *Client) Raw(ctx context.Context, method, path string, payload any) (map[string]any, RequestMeta, error) {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = http.MethodPost
	}

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, RequestMeta{}, err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return nil, RequestMeta{}, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	c.setAuth(req)
	return c.do(req)
}

func (c *Client) post(ctx context.Context, path string, payload any) (map[string]any, RequestMeta, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, RequestMeta{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, RequestMeta{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuth(req)

	return c.do(req)
}

func (c *Client) get(ctx context.Context, path string) (map[string]any, RequestMeta, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return nil, RequestMeta{}, err
	}
	c.setAuth(req)
	return c.do(req)
}

func (c *Client) delete(ctx context.Context, path string) (map[string]any, RequestMeta, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.BaseURL+path, nil)
	if err != nil {
		return nil, RequestMeta{}, err
	}
	c.setAuth(req)
	return c.do(req)
}

func (c *Client) do(req *http.Request) (map[string]any, RequestMeta, error) {
	meta := RequestMeta{}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, meta, err
	}
	defer resp.Body.Close()

	meta.StatusCode = resp.StatusCode
	meta.RequestID = resp.Header.Get("x-request-id")
	if meta.RequestID == "" {
		meta.RequestID = resp.Header.Get("request-id")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, meta, err
	}

	var payload map[string]any
	if len(data) > 0 {
		if err := json.Unmarshal(data, &payload); err != nil {
			return nil, meta, fmt.Errorf("decode response: %w", err)
		}
	}

	if payload != nil && meta.RequestID == "" {
		if requestID, ok := payload["requestId"].(string); ok {
			meta.RequestID = requestID
		}
	}

	if resp.StatusCode >= 400 {
		message := string(bytes.TrimSpace(data))
		if payload != nil {
			if msg, ok := payload["error"].(string); ok {
				message = msg
			}
			if msg, ok := payload["message"].(string); ok && msg != "" {
				message = msg
			}
		}
		return nil, meta, &APIError{StatusCode: resp.StatusCode, Message: message}
	}

	return payload, meta, nil
}

func (c *Client) setAuth(req *http.Request) {
	req.Header.Set("User-Agent", c.UserAgent)
	if c.APIKey == "" {
		return
	}
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
}
