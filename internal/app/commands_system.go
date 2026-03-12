package app

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/samaydhawan/exa-cli/internal/config"
	"github.com/samaydhawan/exa-cli/internal/docs"
	"github.com/samaydhawan/exa-cli/internal/exa"
	"github.com/samaydhawan/exa-cli/internal/mcp"
)

func (a *App) newMCPCmd() *cobra.Command {
	var (
		allTools bool
		embedKey bool
	)

	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP helpers for Codex, Claude Code, Cursor, and generic clients",
	}

	printCmd := &cobra.Command{
		Use:   "print <codex|claude-code|cursor|generic>",
		Short: "Print an MCP snippet wired to Exa hosted MCP",
		Example: strings.Join([]string{
			"exa-cli mcp print codex",
			"exa-cli mcp print claude-code --all-tools",
			"exa-cli mcp print cursor --embed-key",
		}, "\n"),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			tools := mcp.DefaultTools()
			if allTools {
				tools = mcp.FullTools()
			}

			apiKey := ""
			if embedKey {
				apiKey = a.cfg.APIKey
			}
			snippet, err := mcp.Render(target, a.cfg.MCPURL, apiKey, tools)
			if err != nil {
				return wrap(2, err)
			}

			fmt.Fprintln(a.out, snippet)
			return nil
		},
	}
	printCmd.Flags().BoolVar(&allTools, "all-tools", false, "Include Exa's optional hosted MCP tools")
	printCmd.Flags().BoolVar(&embedKey, "embed-key", false, "Inline the current API key into the generated snippet")

	doctorCmd := &cobra.Command{
		Use:   "doctor",
		Short: "Show hosted MCP configuration and auth guidance",
		RunE: func(cmd *cobra.Command, args []string) error {
			data := map[string]any{
				"mcp_url":            a.cfg.MCPURL,
				"default_tools":      mcp.DefaultTools(),
				"api_key_configured": a.cfg.APIKey != "",
				"embed_key_default":  false,
				"recommended":        []string{"codex", "claude-code", "cursor"},
			}
			return a.writeEnvelope(data, a.newMeta("mcp doctor", exa.RequestMeta{}), "table")
		},
	}

	mcpCmd.AddCommand(printCmd, doctorCmd)
	return mcpCmd
}

func (a *App) newAuthCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage Exa API credentials",
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show current auth source and cache path",
		RunE: func(cmd *cobra.Command, args []string) error {
			source := a.authSource()

			data := map[string]any{
				"auth_source": source,
				"config_path": a.configPath,
				"cache_path":  a.cfg.CachePath,
			}
			if source != "none" {
				data["key_preview"] = maskSecret(firstString(strings.TrimSpace(a.env["EXA_API_KEY"]), strings.TrimSpace(a.cfg.APIKey)))
			}
			if warning := a.authWarning(); warning != "" {
				data["warning"] = warning
				data["env_override"] = true
			}
			return a.writeEnvelope(data, a.newMeta("auth status", exa.RequestMeta{}), "table")
		},
	}

	var (
		loginKey  string
		fromStdin bool
	)
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Store an Exa API key in config.toml",
		Example: strings.Join([]string{
			`exa-cli auth login`,
			`exa-cli auth login --api-key "your-key"`,
			`printf '%s\n' "$EXA_API_KEY" | exa-cli auth login --stdin`,
		}, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := strings.TrimSpace(loginKey)
			if fromStdin {
				data, err := io.ReadAll(a.in)
				if err != nil {
					return wrap(1, err)
				}
				key = strings.TrimSpace(string(data))
			}
			if key == "" {
				if !a.isInteractive() {
					return wrap(2, errors.New("provide --api-key or pipe a key with --stdin when stdin is not interactive"))
				}
				fmt.Fprint(a.out, "Exa API key: ")
				if reader, ok := a.in.(*os.File); ok && term.IsTerminal(int(reader.Fd())) {
					bytes, err := term.ReadPassword(int(reader.Fd()))
					fmt.Fprintln(a.out)
					if err != nil {
						return wrap(1, err)
					}
					key = strings.TrimSpace(string(bytes))
				} else {
					reader := bufio.NewReader(a.in)
					line, err := reader.ReadString('\n')
					if err != nil && !errors.Is(err, io.EOF) {
						return wrap(1, err)
					}
					key = strings.TrimSpace(line)
				}
			}
			if key == "" {
				return wrap(2, errors.New("empty API key"))
			}
			a.cfg.APIKey = key
			if err := config.Save(a.configPath, a.cfg); err != nil {
				return wrap(1, err)
			}
			fmt.Fprintf(a.out, "Saved API key to %s\n", a.configPath)
			return nil
		},
	}
	loginCmd.Flags().StringVar(&loginKey, "api-key", "", "Exa API key to save into config.toml")
	loginCmd.Flags().BoolVar(&fromStdin, "stdin", false, "Read the API key from stdin")

	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove the stored API key from config.toml",
		RunE: func(cmd *cobra.Command, args []string) error {
			a.cfg.APIKey = ""
			if err := config.Save(a.configPath, a.cfg); err != nil {
				return wrap(1, err)
			}
			fmt.Fprintf(a.out, "Removed stored API key from %s\n", a.configPath)
			return nil
		},
	}

	authCmd.AddCommand(statusCmd, loginCmd, logoutCmd)
	return authCmd
}

func (a *App) newCacheCmd() *cobra.Command {
	cacheCmd := &cobra.Command{
		Use:   "cache",
		Short: "Inspect or maintain the local SQLite cache",
	}

	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "Show cache entry count and payload size",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.ensureCache(); err != nil {
				return wrap(1, err)
			}
			if a.cache == nil {
				fmt.Fprintln(a.out, "Cache disabled")
				return nil
			}
			stats, err := a.cache.Stats(cmd.Context())
			if err != nil {
				return wrap(1, err)
			}
			data := map[string]any{
				"entries": stats.Entries,
				"bytes":   stats.Bytes,
				"path":    a.cache.Path(),
			}
			return a.writeEnvelope(data, a.newMeta("cache stats", exa.RequestMeta{}), "table")
		},
	}

	pruneCmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove expired cache entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.ensureCache(); err != nil {
				return wrap(1, err)
			}
			if a.cache == nil {
				fmt.Fprintln(a.out, "Cache disabled")
				return nil
			}
			rows, err := a.cache.Prune(cmd.Context())
			if err != nil {
				return wrap(1, err)
			}
			fmt.Fprintf(a.out, "Pruned %d expired entries\n", rows)
			return nil
		},
	}

	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Delete all cache entries and tracked runs",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.ensureCache(); err != nil {
				return wrap(1, err)
			}
			if a.cache == nil {
				fmt.Fprintln(a.out, "Cache disabled")
				return nil
			}
			if err := a.cache.Clear(cmd.Context()); err != nil {
				return wrap(1, err)
			}
			fmt.Fprintln(a.out, "Cache cleared")
			return nil
		},
	}

	cacheCmd.AddCommand(statsCmd, pruneCmd, clearCmd)
	return cacheCmd
}

func (a *App) newConfigCmd() *cobra.Command {
	var force bool

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect and initialize exa-cli config",
	}

	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Print the resolved config file values",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.writeEnvelope(configToMap(a.cfg), a.newMeta("config show", exa.RequestMeta{}), "json")
		},
	}

	pathCmd := &cobra.Command{
		Use:   "path",
		Short: "Print the config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(a.out, a.configPath)
			return nil
		},
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Write a starter config.toml",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				if _, err := os.Stat(a.configPath); err == nil {
					return wrap(2, fmt.Errorf("%s already exists; use --force to overwrite", a.configPath))
				}
			}
			cfg := config.Default()
			cfg.CachePath = a.cfg.CachePath
			if err := config.Save(a.configPath, cfg); err != nil {
				return wrap(1, err)
			}
			fmt.Fprintf(a.out, "Wrote starter config to %s\n", a.configPath)
			return nil
		},
	}
	initCmd.Flags().BoolVar(&force, "force", false, "Overwrite an existing config.toml")

	configCmd.AddCommand(showCmd, pathCmd, initCmd)
	return configCmd
}

func (a *App) newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check local configuration, cache, and output behavior",
		RunE: func(cmd *cobra.Command, args []string) error {
			data := map[string]any{
				"version":            a.version.Version,
				"base_url":           a.cfg.BaseURL,
				"mcp_url":            a.cfg.MCPURL,
				"config_path":        a.configPath,
				"cache_enabled":      !a.options.NoCache,
				"cache_path":         a.cfg.CachePath,
				"interactive_stdout": a.isInteractive(),
				"auth_source":        a.authSource(),
				"api_key_configured": a.cfg.APIKey != "",
				"default_format":     a.configuredDefaultFormat(),
				"default_profile":    a.configuredDefaultProfile(),
			}
			if warning := a.authWarning(); warning != "" {
				data["warning"] = warning
			}
			return a.writeEnvelope(data, a.newMeta("doctor", exa.RequestMeta{}), "table")
		},
	}
}

func (a *App) newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			a.hydratePersistentFlagValues(cmd)
			if err := validateFormat(a.options.Format); err != nil {
				return wrap(2, err)
			}
			data := map[string]any{
				"version": a.version.Version,
				"commit":  a.version.Commit,
				"date":    a.version.Date,
			}
			return a.writeEnvelope(data, a.newMeta("version", exa.RequestMeta{}), "table")
		},
	}
}

func (a *App) newRawCmd() *cobra.Command {
	rawCmd := &cobra.Command{
		Use:   "raw",
		Short: "Low-level commands that map directly to Exa endpoints",
	}

	rawCmd.AddCommand(
		a.newRawRequestCmd(),
		a.newRawEndpointCmd("search", http.MethodPost, "/search", func(args []string) any {
			if len(args) == 0 {
				return map[string]any{}
			}
			return map[string]any{"query": strings.Join(args, " ")}
		}),
		a.newRawEndpointCmd("answer", http.MethodPost, "/answer", func(args []string) any {
			if len(args) == 0 {
				return map[string]any{}
			}
			return map[string]any{"query": strings.Join(args, " ")}
		}),
		a.newRawEndpointCmd("contents", http.MethodPost, "/contents", func(args []string) any {
			if len(args) == 0 {
				return map[string]any{}
			}
			return map[string]any{"urls": args}
		}),
		a.newRawEndpointCmd("context", http.MethodPost, "/context", func(args []string) any {
			if len(args) == 0 {
				return map[string]any{}
			}
			return map[string]any{"query": strings.Join(args, " ")}
		}),
		a.newRawResearchCmd(),
	)
	return rawCmd
}

func (a *App) newRawRequestCmd() *cobra.Command {
	var (
		method    string
		path      string
		inputPath string
	)

	cmd := &cobra.Command{
		Use:   "request",
		Short: "Make an exact raw Exa API request",
		Example: strings.Join([]string{
			`exa-cli raw request --method POST --path /search --input request.json`,
			`printf '%s\n' '{"query":"golang sqlite wal"}' | exa-cli raw request --method POST --path /search --input -`,
			`exa-cli raw request --method GET --path /research/v1/research_id`,
		}, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.requireAPIKey(); err != nil {
				return err
			}
			if strings.TrimSpace(path) == "" {
				return wrap(2, errors.New("provide --path"))
			}
			payload, err := readInputPayload(inputPath, nil, a.in)
			if err != nil {
				return wrap(2, err)
			}
			response, requestMeta, err := a.client.Raw(cmd.Context(), method, path, payload)
			if err != nil {
				return normalizeAPIError(err)
			}
			return a.writeEnvelope(response, a.newMeta("raw request", requestMeta), "json")
		},
	}
	cmd.Flags().StringVar(&method, "method", http.MethodPost, "HTTP method to use")
	cmd.Flags().StringVar(&path, "path", "", "Exa API path, for example /search")
	cmd.Flags().StringVar(&inputPath, "input", "", "Path to a JSON request body or - for stdin")
	return cmd
}

func (a *App) newRawResearchCmd() *cobra.Command {
	researchCmd := &cobra.Command{
		Use:   "research",
		Short: "Raw passthrough commands for /research/v1",
	}

	researchCmd.AddCommand(
		a.newRawEndpointCmd("start", http.MethodPost, "/research/v1", func(args []string) any {
			if len(args) == 0 {
				return map[string]any{}
			}
			return map[string]any{"instructions": strings.Join(args, " ")}
		}),
		a.newRawEndpointCmd("get", http.MethodGet, "/research/v1", nil),
		a.newRawEndpointCmd("cancel", http.MethodDelete, "/research/v1", nil),
	)
	return researchCmd
}

func (a *App) newRawEndpointCmd(name, method, path string, fallback func(args []string) any) *cobra.Command {
	var inputPath string

	cmd := &cobra.Command{
		Use:   name + " [query-or-args]",
		Short: "Raw passthrough for " + path,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.requireAPIKey(); err != nil {
				return err
			}

			requestPath := path
			if (method == http.MethodGet || method == http.MethodDelete) && len(args) > 0 {
				requestPath = strings.TrimRight(path, "/") + "/" + args[0]
			}

			var fallbackPayload any
			if fallback != nil {
				fallbackPayload = fallback(args)
			}
			payload, err := readInputPayload(inputPath, fallbackPayload, a.in)
			if err != nil {
				return wrap(2, err)
			}
			response, requestMeta, err := a.client.Raw(cmd.Context(), method, requestPath, payload)
			if err != nil {
				return normalizeAPIError(err)
			}
			meta := a.newMeta("raw "+name, requestMeta)
			return a.writeEnvelope(response, meta, "json")
		},
	}
	cmd.Flags().StringVar(&inputPath, "input", "", "Path to a JSON request body or - for stdin")
	return cmd
}

func (a *App) newGenCmd() *cobra.Command {
	genCmd := &cobra.Command{
		Use:    "gen",
		Short:  "Internal helpers",
		Hidden: true,
	}

	genCmd.AddCommand(&cobra.Command{
		Use:   "docs",
		Short: "Generate CLI reference markdown",
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := docs.Generate(a.root)
			if err != nil {
				return wrap(1, err)
			}
			fmt.Fprintln(a.out, doc)
			return nil
		},
	})
	return genCmd
}
