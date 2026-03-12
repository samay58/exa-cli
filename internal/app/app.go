package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/samaydhawan/exa-cli/internal/cache"
	"github.com/samaydhawan/exa-cli/internal/config"
	"github.com/samaydhawan/exa-cli/internal/exa"
	"github.com/samaydhawan/exa-cli/internal/version"
)

type App struct {
	in      io.Reader
	out     io.Writer
	errOut  io.Writer
	env     map[string]string
	version version.Info

	configPath   string
	cfg          config.Config
	storedAPIKey string
	client       *exa.Client
	cache        *cache.Store

	parsedFlags  GlobalOptions
	options      GlobalOptions
	root         *cobra.Command
	runtimeReady bool
}

type GlobalOptions struct {
	Format       string
	FormatSource string
	Profile      string
	Verbose      bool
	NoColor      bool
	NoBanner     bool
	NoCache      bool
}

type Envelope struct {
	Meta Meta `json:"meta"`
	Data any  `json:"data"`
}

type Meta struct {
	Command     string   `json:"command"`
	Profile     string   `json:"profile,omitempty"`
	Format      string   `json:"format,omitempty"`
	Cache       string   `json:"cache,omitempty"`
	RequestID   string   `json:"requestId,omitempty"`
	SearchType  string   `json:"searchType,omitempty"`
	CostDollars float64  `json:"costDollars,omitempty"`
	SearchTime  float64  `json:"searchTime,omitempty"`
	Warnings    []string `json:"warnings,omitempty"`
	GeneratedAt string   `json:"generatedAt"`
}

type cachedPayload struct {
	Response map[string]any  `json:"response"`
	Request  exa.RequestMeta `json:"request"`
	SavedAt  string          `json:"savedAt"`
}

type cliError struct {
	Code int
	Err  error
}

func (e *cliError) Error() string {
	return e.Err.Error()
}

func (e *cliError) Unwrap() error {
	return e.Err
}

func Run(ctx context.Context, args []string, in io.Reader, out io.Writer, errOut io.Writer, env []string, info version.Info) int {
	app := &App{
		in:      in,
		out:     out,
		errOut:  errOut,
		env:     envMap(env),
		version: info,
	}

	root := app.newRootCmd()
	app.root = root
	root.SetArgs(args)
	root.SetOut(out)
	root.SetErr(errOut)

	err := root.ExecuteContext(ctx)
	if err == nil {
		_ = app.close()
		return 0
	}

	_ = app.close()

	var codeErr *cliError
	if errors.As(err, &codeErr) {
		fmt.Fprintln(errOut, codeErr.Err.Error())
		return codeErr.Code
	}

	fmt.Fprintln(errOut, err.Error())
	return 1
}

func (a *App) close() error {
	if a.cache != nil {
		return a.cache.Close()
	}
	return nil
}

func (a *App) initRuntime(cmd *cobra.Command) error {
	if a.runtimeReady {
		return nil
	}

	a.hydratePersistentFlagValues(cmd)

	if a.skipHeavyRuntime(cmd) {
		if err := a.seedMinimalRuntime(); err != nil {
			return wrap(2, err)
		}
		a.runtimeReady = true
		return nil
	}

	configPath := strings.TrimSpace(a.env["EXA_CLI_CONFIG"])
	if configPath == "" {
		defaultPath, err := config.DefaultConfigPath()
		if err != nil {
			return wrap(1, err)
		}
		configPath = defaultPath
	}
	a.configPath = configPath

	cfg, err := config.Load(configPath)
	if err != nil {
		return wrap(1, fmt.Errorf("load config: %w", err))
	}
	a.storedAPIKey = strings.TrimSpace(cfg.APIKey)

	if v := strings.TrimSpace(a.env["EXA_API_KEY"]); v != "" {
		cfg.APIKey = v
	}
	if v := strings.TrimSpace(a.env["EXA_BASE_URL"]); v != "" {
		cfg.BaseURL = v
	}
	if v := strings.TrimSpace(a.env["EXA_MCP_URL"]); v != "" {
		cfg.MCPURL = v
	}

	if !cmd.Flags().Changed("format") {
		if v := strings.TrimSpace(a.env["EXA_CLI_FORMAT"]); v != "" {
			a.options.Format = v
			a.options.FormatSource = "env"
		} else if cfg.Format != "" && cfg.Format != config.DefaultFormat {
			a.options.Format = cfg.Format
			a.options.FormatSource = "config"
		} else {
			a.options.Format = config.DefaultFormat
			a.options.FormatSource = "default"
		}
	} else {
		a.options.FormatSource = "flag"
	}

	if !cmd.Flags().Changed("profile") {
		if v := strings.TrimSpace(a.env["EXA_CLI_PROFILE"]); v != "" {
			a.options.Profile = v
		} else if cfg.Profile != "" {
			a.options.Profile = cfg.Profile
		}
	}

	if !cmd.Flags().Changed("no-banner") {
		if truthy(a.env["EXA_CLI_NO_BANNER"]) {
			a.options.NoBanner = true
		} else {
			a.options.NoBanner = cfg.NoBanner
		}
	}
	if !cmd.Flags().Changed("no-cache") && truthy(a.env["EXA_CLI_NO_CACHE"]) {
		a.options.NoCache = true
	}
	if !cmd.Flags().Changed("no-color") && a.env["NO_COLOR"] != "" {
		a.options.NoColor = true
	}

	if err := validateFormat(a.options.Format); err != nil {
		return wrap(2, err)
	}
	if err := validateProfile(a.options.Profile); err != nil {
		return wrap(2, err)
	}

	if cfg.CachePath == "" {
		cachePath, err := config.DefaultCachePath()
		if err != nil {
			return wrap(1, err)
		}
		cfg.CachePath = cachePath
	}

	a.cfg = cfg
	a.client = exa.New(cfg.BaseURL, cfg.APIKey)
	a.client.UserAgent = fmt.Sprintf("exa-cli/%s", a.version.Version)

	a.runtimeReady = true
	return nil
}

func (a *App) seedMinimalRuntime() error {
	a.cfg = config.Default()
	a.storedAPIKey = ""
	if v := strings.TrimSpace(a.env["EXA_API_KEY"]); v != "" {
		a.cfg.APIKey = v
	}
	if v := strings.TrimSpace(a.env["EXA_BASE_URL"]); v != "" {
		a.cfg.BaseURL = v
	}
	if v := strings.TrimSpace(a.env["EXA_MCP_URL"]); v != "" {
		a.cfg.MCPURL = v
	}
	if a.options.FormatSource != "flag" {
		if v := strings.TrimSpace(a.env["EXA_CLI_FORMAT"]); v != "" {
			a.options.Format = v
			a.options.FormatSource = "env"
		} else {
			a.options.Format = config.DefaultFormat
			a.options.FormatSource = "default"
		}
	}
	if v := strings.TrimSpace(a.env["EXA_CLI_PROFILE"]); v != "" {
		a.options.Profile = v
	} else {
		a.options.Profile = config.DefaultProfile
	}
	if truthy(a.env["EXA_CLI_NO_BANNER"]) {
		a.options.NoBanner = true
	}
	if truthy(a.env["EXA_CLI_NO_CACHE"]) {
		a.options.NoCache = true
	}
	if a.env["NO_COLOR"] != "" {
		a.options.NoColor = true
	}
	if err := validateFormat(a.options.Format); err != nil {
		return err
	}
	if err := validateProfile(a.options.Profile); err != nil {
		return err
	}
	a.client = exa.New(a.cfg.BaseURL, a.cfg.APIKey)
	a.client.UserAgent = fmt.Sprintf("exa-cli/%s", a.version.Version)
	return nil
}

func (a *App) ensureCache() error {
	if a.options.NoCache || a.cache != nil {
		return nil
	}
	if strings.TrimSpace(a.cfg.CachePath) == "" {
		cachePath, err := config.DefaultCachePath()
		if err != nil {
			return err
		}
		a.cfg.CachePath = cachePath
	}
	store, err := cache.Open(a.cfg.CachePath)
	if err != nil {
		return fmt.Errorf("open cache: %w", err)
	}
	a.cache = store
	return nil
}

func (a *App) skipHeavyRuntime(cmd *cobra.Command) bool {
	switch cmd.CommandPath() {
	case "exa-cli version", "exa-cli gen", "exa-cli gen docs", "exa-cli help":
		return true
	default:
		return false
	}
}

func (a *App) hydratePersistentFlagValues(cmd *cobra.Command) {
	if flag := cmd.Flag("format"); flag != nil {
		a.options.Format = flag.Value.String()
		if flag.Changed {
			a.options.FormatSource = "flag"
		}
	}
	if flag := cmd.Flag("verbose"); flag != nil {
		a.options.Verbose = truthy(flag.Value.String())
	}
	if flag := cmd.Flag("no-color"); flag != nil {
		a.options.NoColor = truthy(flag.Value.String())
	}
	if flag := cmd.Flag("no-banner"); flag != nil {
		a.options.NoBanner = truthy(flag.Value.String())
	}
	if flag := cmd.Flag("no-cache"); flag != nil {
		a.options.NoCache = truthy(flag.Value.String())
	}
}
