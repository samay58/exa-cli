package app

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/samaydhawan/exa-cli/internal/config"
	"github.com/samaydhawan/exa-cli/internal/exa"
)

func (a *App) newFindCmd() *cobra.Command {
	var (
		numResults       int
		category         string
		includeDomains   []string
		excludeDomains   []string
		profile          string
		systemPrompt     string
		outputSchemaPath string
	)

	cmd := &cobra.Command{
		Use:   "find <query>",
		Short: "Search Exa with agent-friendly defaults",
		Example: strings.Join([]string{
			`exa-cli find "Go CLI design patterns for machine-readable output"`,
			`exa-cli find "AI infra companies building eval tooling" --category company`,
			`exa-cli find "Rust sqlite WAL mode" --profile fast --format json`,
			`exa-cli find "SEC filings for NVIDIA Q4 2025" --profile deep --system-prompt "Focus on revenue breakdown and forward guidance"`,
		}, "\n"),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.requireAPIKey(); err != nil {
				return err
			}

			if err := validateCategoryFilters(category, includeDomains, excludeDomains); err != nil {
				return wrap(2, err)
			}
			effectiveProfile := a.options.Profile
			if cmd.Flags().Changed("profile") {
				effectiveProfile = profile
			}
			if err := validateProfile(effectiveProfile); err != nil {
				return wrap(2, err)
			}

			var outputSchema map[string]any
			if outputSchemaPath != "" {
				raw, err := readInputPayload(outputSchemaPath, nil, a.in)
				if err != nil {
					return wrap(2, fmt.Errorf("reading output schema: %w", err))
				}
				if m, ok := raw.(map[string]any); ok {
					outputSchema = m
				} else {
					return wrap(2, fmt.Errorf("output schema must be a JSON object"))
				}
			}

			query := strings.Join(args, " ")
			req := exa.SearchRequest{
				Query:          query,
				Type:           profileToSearchType(effectiveProfile),
				Category:       category,
				NumResults:     numResults,
				IncludeDomains: includeDomains,
				ExcludeDomains: excludeDomains,
				SystemPrompt:   systemPrompt,
				OutputSchema:   outputSchema,
				Contents: map[string]any{
					"highlights": map[string]any{"maxCharacters": 280},
				},
			}

			payload, meta, err := a.cachedJSON(cmd.Context(), "search", req, func() (map[string]any, exa.RequestMeta, error) {
				return a.client.Search(cmd.Context(), req)
			})
			if err != nil {
				return err
			}

			meta.Command = "find"
			meta.Profile = effectiveProfile
			meta.SearchType = firstString(asString(payload["type"]), profileToSearchType(effectiveProfile))
			return a.writeEnvelope(payload, meta, "table")
		},
	}

	cmd.Flags().IntVar(&numResults, "num-results", 5, "Number of results to request")
	cmd.Flags().StringVar(&category, "category", "", "Optional search category, for example company or people")
	cmd.Flags().StringSliceVar(&includeDomains, "include-domain", nil, "Limit search to the provided domains")
	cmd.Flags().StringSliceVar(&excludeDomains, "exclude-domain", nil, "Exclude the provided domains")
	cmd.Flags().StringVar(&profile, "profile", config.DefaultProfile, "Latency/cost profile: instant, fast, balanced (default), deep ($12/1k, 4-12s), reasoned ($15/1k, 12-50s)")
	cmd.Flags().StringVar(&systemPrompt, "system-prompt", "", "Instructions to guide deep search synthesis (deep and reasoned profiles)")
	cmd.Flags().StringVar(&outputSchemaPath, "output-schema", "", "Path to JSON schema file for structured deep search output (use - for stdin)")
	return cmd
}

func (a *App) newAskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ask <question>",
		Short: "Get a cited answer from Exa Answer",
		Example: strings.Join([]string{
			`exa-cli ask "How does Exa answer differ from search?"`,
			`exa-cli ask "What are the tradeoffs between fast and deep search?" --format json`,
		}, "\n"),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.requireAPIKey(); err != nil {
				return err
			}

			req := exa.AnswerRequest{Query: strings.Join(args, " ")}
			payload, meta, err := a.cachedJSON(cmd.Context(), "answer", req, func() (map[string]any, exa.RequestMeta, error) {
				return a.client.Answer(cmd.Context(), req)
			})
			if err != nil {
				return err
			}

			meta.Command = "ask"
			return a.writeEnvelope(payload, meta, "markdown")
		},
	}
	return cmd
}

func (a *App) newCodeCmd() *cobra.Command {
	var tokens string

	cmd := &cobra.Command{
		Use:   "code <query>",
		Short: "Fetch code-focused context for coding agents",
		Example: strings.Join([]string{
			`exa-cli code "OpenAI Python streaming client patterns"`,
			`exa-cli code "SQLite WAL busy timeout in Go" --tokens 5000`,
		}, "\n"),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.requireAPIKey(); err != nil {
				return err
			}
			if err := validateTokens(tokens); err != nil {
				return wrap(2, err)
			}

			req := exa.ContextRequest{
				Query:     strings.Join(args, " "),
				TokensNum: normalizeTokensNum(tokens),
			}

			payload, meta, err := a.cachedJSON(cmd.Context(), "context", req, func() (map[string]any, exa.RequestMeta, error) {
				return a.client.Context(cmd.Context(), req)
			})
			if err != nil {
				return err
			}

			meta.Command = "code"
			meta.CostDollars = readFloat(payload, "costDollars")
			meta.SearchTime = readFloat(payload, "searchTime")
			return a.writeEnvelope(payload, meta, "llm")
		},
	}

	cmd.Flags().StringVar(&tokens, "tokens", "dynamic", `Token budget for code context. Use "dynamic" or a positive integer like 5000`)
	return cmd
}

func (a *App) newReadCmd() *cobra.Command {
	var (
		maxAgeHours      int
		withText         bool
		withSummary      bool
		withHighlights   bool
		livecrawlTimeout int
	)

	cmd := &cobra.Command{
		Use:   "read <url-or-id...>",
		Short: "Retrieve page contents via Exa Contents",
		Example: strings.Join([]string{
			`exa-cli read https://exa.ai/docs/reference/search --summary`,
			`exa-cli read https://exa.ai/docs/reference/search --summary --text`,
			`exa-cli read https://exa.ai/docs/reference/get-contents --max-age-hours 0`,
		}, "\n"),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.requireAPIKey(); err != nil {
				return err
			}

			req := exa.ContentsRequest{}
			for _, arg := range args {
				if strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") {
					req.URLs = append(req.URLs, arg)
				} else {
					req.IDs = append(req.IDs, arg)
				}
			}
			if withText {
				req.Text = true
			}
			if withSummary {
				req.Summary = map[string]any{"query": "Summarize the page for fast terminal reading."}
			}
			if withHighlights {
				req.Highlights = map[string]any{"maxCharacters": 800}
			}
			if cmd.Flags().Changed("max-age-hours") {
				req.MaxAgeHours = &maxAgeHours
			}
			if livecrawlTimeout > 0 {
				req.LivecrawlTimeout = livecrawlTimeout
			}

			payload, meta, err := a.cachedJSON(cmd.Context(), "contents", req, func() (map[string]any, exa.RequestMeta, error) {
				return a.client.Contents(cmd.Context(), req)
			})
			if err != nil {
				return err
			}

			meta.Command = "read"
			return a.writeEnvelope(payload, meta, "markdown")
		},
	}

	cmd.Flags().BoolVar(&withText, "text", false, "Include full markdown text when available")
	cmd.Flags().BoolVar(&withSummary, "summary", false, "Include Exa-generated summaries")
	cmd.Flags().BoolVar(&withHighlights, "highlights", true, "Include highlights when available")
	cmd.Flags().IntVar(&livecrawlTimeout, "livecrawl-timeout", 0, "Optional livecrawl timeout in milliseconds")
	cmd.Flags().IntVar(&maxAgeHours, "max-age-hours", 0, "Freshness control for contents; 0 forces live crawl, -1 forces cache")
	return cmd
}

func (a *App) newResearchCmd() *cobra.Command {
	var (
		detach       bool
		model        string
		pollInterval time.Duration
		limit        int
		quiet        bool
		savePath     string
	)

	researchCmd := &cobra.Command{
		Use:   "research",
		Short: "Run or inspect async Exa research tasks",
	}

	runCmd := &cobra.Command{
		Use:   "run <query>",
		Short: "Start a research task; waits by default",
		Example: strings.Join([]string{
			`exa-cli research run "design a search CLI for coding agents"`,
			`exa-cli research run "map the AI code search landscape" --detach`,
			`exa-cli research run "NVIDIA Q4 2025 earnings" --save`,
			`exa-cli research run "AI infrastructure landscape" --save ai-infra`,
		}, "\n"),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.requireAPIKey(); err != nil {
				return err
			}

			req := exa.ResearchCreateRequest{
				Instructions: strings.Join(args, " "),
				Model:        model,
			}
			payload, requestMeta, err := a.client.ResearchCreate(cmd.Context(), req)
			if err != nil {
				return normalizeAPIError(err)
			}

			researchID := firstString(asString(payload["id"]), asString(payload["researchId"]))
			if err := a.ensureCache(); err == nil && researchID != "" && a.cache != nil {
				_ = a.cache.PutRun(cmd.Context(), researchID, "research", payload)
			}

			meta := a.newMeta("research run", requestMeta)
			if detach || researchID == "" {
				return a.writeEnvelope(payload, meta, "markdown")
			}

			statusPayload, statusMeta, err := a.waitForResearch(cmd.Context(), researchID, pollInterval, quiet)
			if err != nil {
				return err
			}
			statusMeta.Command = "research run"
			if err := a.writeEnvelope(statusPayload, statusMeta, "markdown"); err != nil {
				return err
			}
			if cmd.Flags().Changed("save") {
				return a.saveResearchReport(statusPayload, statusMeta, savePath, req.Instructions)
			}
			return nil
		},
	}
	runCmd.Flags().BoolVar(&detach, "detach", false, "Return the task ID immediately instead of waiting for completion")
	runCmd.Flags().StringVar(&model, "model", "exa-research", "Research model: exa-research or exa-research-pro")
	runCmd.Flags().DurationVar(&pollInterval, "poll-interval", 5*time.Second, "Polling interval for attached research runs")
	runCmd.Flags().BoolVar(&quiet, "quiet", false, "Suppress progress updates while waiting for completion")
	runCmd.Flags().StringVar(&savePath, "save", "", "Save report to a markdown file (auto-names from query when used without a value)")
	runCmd.Flags().Lookup("save").NoOptDefVal = "auto"

	getCmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Fetch a research task by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.requireAPIKey(); err != nil {
				return err
			}
			payload, requestMeta, err := a.client.ResearchGet(cmd.Context(), args[0], true)
			if err != nil {
				return normalizeAPIError(err)
			}
			if err := a.ensureCache(); err == nil && a.cache != nil {
				_ = a.cache.PutRun(cmd.Context(), args[0], "research", payload)
			}
			meta := a.newMeta("research get", requestMeta)
			return a.writeEnvelope(payload, meta, "markdown")
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List recent research tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.requireAPIKey(); err != nil {
				return err
			}
			payload, requestMeta, err := a.client.ResearchList(cmd.Context(), limit)
			if err != nil {
				return normalizeAPIError(err)
			}
			meta := a.newMeta("research list", requestMeta)
			return a.writeEnvelope(payload, meta, "table")
		},
	}
	listCmd.Flags().IntVar(&limit, "limit", 10, "Number of recent research tasks to fetch")

	cancelCmd := &cobra.Command{
		Use:   "cancel <id>",
		Short: "Attempt to cancel a research task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.requireAPIKey(); err != nil {
				return err
			}
			payload, requestMeta, err := a.client.ResearchCancel(cmd.Context(), args[0])
			if err != nil {
				return normalizeAPIError(err)
			}
			meta := a.newMeta("research cancel", requestMeta)
			return a.writeEnvelope(payload, meta, "markdown")
		},
	}

	researchCmd.AddCommand(runCmd, getCmd, listCmd, cancelCmd)
	return researchCmd
}

func (a *App) saveResearchReport(payload map[string]any, meta Meta, savePath, query string) error {
	filename := savePath
	if filename == "auto" {
		filename = slugify(query, 60) + ".md"
	} else if !strings.HasSuffix(filename, ".md") {
		filename += ".md"
	}

	envelope := Envelope{Meta: meta, Data: payload}
	md := renderMarkdown(envelope, false)

	if err := os.WriteFile(filename, []byte(md+"\n"), 0644); err != nil {
		return wrap(1, fmt.Errorf("saving report to %s: %w", filename, err))
	}
	fmt.Fprintf(a.errOut, "saved: %s\n", filename)
	return nil
}
