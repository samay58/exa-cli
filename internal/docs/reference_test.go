package docs

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestGenerateIncludesExamplesAndInheritedFlags(t *testing.T) {
	root := &cobra.Command{
		Use:   "exa-cli",
		Short: "Root command",
	}
	root.PersistentFlags().String("format", "table", "Output format")

	findCmd := &cobra.Command{
		Use:     "find <query>",
		Short:   "Search Exa",
		Example: `exa-cli find "golang cobra docs"`,
		Run:     func(cmd *cobra.Command, args []string) {},
	}
	hiddenCmd := &cobra.Command{
		Use:    "internal",
		Short:  "Internal command",
		Hidden: true,
	}
	completionCmd := &cobra.Command{
		Use:   "completion",
		Short: "Shell completion",
	}

	root.AddCommand(findCmd, hiddenCmd, completionCmd)

	output, err := Generate(root)
	if err != nil {
		t.Fatalf("generate docs: %v", err)
	}

	if !strings.Contains(output, "## exa-cli find") {
		t.Fatalf("expected child command in docs, got:\n%s", output)
	}
	if !strings.Contains(output, "Examples:") || !strings.Contains(output, `exa-cli find "golang cobra docs"`) {
		t.Fatalf("expected examples in docs, got:\n%s", output)
	}
	if !strings.Contains(output, "Inherited flags:") || !strings.Contains(output, "`--format`") {
		t.Fatalf("expected inherited flags in docs, got:\n%s", output)
	}
	if strings.Contains(output, "## exa-cli internal") {
		t.Fatalf("did not expect hidden command in docs, got:\n%s", output)
	}
	if strings.Contains(output, "## exa-cli completion") {
		t.Fatalf("did not expect completion command in docs, got:\n%s", output)
	}
}
