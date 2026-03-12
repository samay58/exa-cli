package app

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/samaydhawan/exa-cli/internal/config"
	"github.com/samaydhawan/exa-cli/internal/exa"
	"github.com/samaydhawan/exa-cli/internal/portal"
)

func (a *App) newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "exa-cli",
		Short:         "Agent-native Exa CLI for search, code context, contents, research, and MCP workflows",
		Long:          "exa-cli is a clean-slate, agent-native command line interface for Exa. It is optimized for fast human use, reliable shell composition, and high-signal workflows inside coding agents.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.renderHome(cmd)
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return a.initRuntime(cmd)
		},
	}

	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.PersistentFlags().StringVar(&a.parsedFlags.Format, "format", config.DefaultFormat, "Output format: table, markdown, json, jsonl, llm")
	cmd.PersistentFlags().BoolVar(&a.parsedFlags.Verbose, "verbose", false, "Print additional request metadata")
	cmd.PersistentFlags().BoolVar(&a.parsedFlags.NoColor, "no-color", false, "Disable ANSI color output")
	cmd.PersistentFlags().BoolVar(&a.parsedFlags.NoBanner, "no-banner", false, "Disable the interactive portal banner")
	cmd.PersistentFlags().BoolVar(&a.parsedFlags.NoCache, "no-cache", false, "Bypass the local SQLite cache")

	cmd.AddCommand(
		a.newFindCmd(),
		a.newAskCmd(),
		a.newCodeCmd(),
		a.newReadCmd(),
		a.newResearchCmd(),
		a.newMCPCmd(),
		a.newAuthCmd(),
		a.newCacheCmd(),
		a.newConfigCmd(),
		a.newDoctorCmd(),
		a.newVersionCmd(),
		a.newRawCmd(),
		a.newGenCmd(),
	)

	return cmd
}

func (a *App) renderHome(cmd *cobra.Command) error {
	a.hydratePersistentFlagValues(cmd)
	if err := validateFormat(a.options.Format); err != nil {
		return wrap(2, err)
	}

	workflows := []string{
		`exa-cli find "best practices for Go CLIs"`,
		`exa-cli ask "How does Exa answer differ from search?"`,
		`exa-cli code "How should a Cobra app split smart and raw commands?"`,
		`exa-cli read https://exa.ai/docs/reference/search --summary`,
		`exa-cli research run "design a search CLI for coding agents"`,
		`exa-cli mcp print codex`,
	}

	lines := make([]string, 0, 20)
	if !a.options.NoBanner && a.isInteractive() {
		lines = append(lines, portal.Render(a.shouldUseColor()))
		lines = append(lines, "")
	}

	authSource := a.authSource()

	nextSteps := []string{}
	if authSource == "none" {
		nextSteps = []string{
			`exa-cli auth login`,
			`exa-cli doctor`,
			`exa-cli find "your first query"`,
		}
	}

	renderFormat := a.effectiveFormat("table")
	configuredDefaultFormat := a.configuredDefaultFormat()
	configuredDefaultProfile := a.configuredDefaultProfile()
	if renderFormat != "table" {
		data := map[string]any{
			"workflows":       workflows,
			"auth_source":     authSource,
			"default_format":  configuredDefaultFormat,
			"default_profile": configuredDefaultProfile,
			"help_hint":       "Use `exa-cli --help` for the full command tree or `exa-cli doctor` to inspect local state.",
		}
		if len(nextSteps) > 0 {
			data["next_steps"] = nextSteps
		}
		if warning := a.authWarning(); warning != "" {
			data["warning"] = warning
		}
		return a.writeEnvelope(data, a.newMeta("home", exa.RequestMeta{}), "table")
	}

	lines = append(lines,
		"Quickstart:",
	)
	for _, workflow := range workflows {
		lines = append(lines, "  "+workflow)
	}
	lines = append(lines,
		"",
		"Status:",
		fmt.Sprintf("  auth: %s", authSource),
		fmt.Sprintf("  default format: %s", configuredDefaultFormat),
		fmt.Sprintf("  default profile: %s", configuredDefaultProfile),
	)
	if warning := a.authWarning(); warning != "" {
		lines = append(lines, "  warning: "+warning)
	}

	if len(nextSteps) > 0 {
		lines = append(lines,
			"",
			"Next steps:",
		)
		for _, step := range nextSteps {
			lines = append(lines, "  "+step)
		}
	}

	lines = append(lines,
		"",
		"Use `exa-cli --help` for the full command tree or `exa-cli doctor` to inspect local state.",
	)

	_, err := fmt.Fprintln(a.out, strings.Join(lines, "\n"))
	return err
}
