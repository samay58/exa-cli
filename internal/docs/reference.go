package docs

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Generate(root *cobra.Command) (string, error) {
	var buf bytes.Buffer
	buf.WriteString("# CLI Reference\n\n")
	buf.WriteString("Generated from the live command tree. Rebuild after changing command metadata.\n\n")

	var walk func(cmd *cobra.Command, level int)
	walk = func(cmd *cobra.Command, level int) {
		if cmd.Hidden || cmd.Name() == "help" || cmd.Name() == "completion" {
			return
		}

		title := strings.Repeat("#", level) + " " + cmd.CommandPath()
		buf.WriteString(title + "\n\n")

		short := strings.TrimSpace(cmd.Short)
		if short != "" {
			buf.WriteString(short + "\n\n")
		}

		if usage := strings.TrimSpace(cmd.UseLine()); usage != "" {
			buf.WriteString("```bash\n" + usage + "\n```\n\n")
		}

		if example := strings.TrimSpace(cmd.Example); example != "" {
			buf.WriteString("Examples:\n\n```bash\n" + example + "\n```\n\n")
		}

		flags := cmd.NonInheritedFlags()
		if flags.HasAvailableFlags() {
			buf.WriteString("Flags:\n\n")
			flags.VisitAll(func(flag *pflag.Flag) {
				if flag.Hidden {
					return
				}
				fmt.Fprintf(&buf, "- `--%s`: %s\n", flag.Name, strings.TrimSpace(flag.Usage))
			})
			buf.WriteString("\n")
		}

		inherited := cmd.InheritedFlags()
		if inherited.HasAvailableFlags() {
			buf.WriteString("Inherited flags:\n\n")
			inherited.VisitAll(func(flag *pflag.Flag) {
				if flag.Hidden {
					return
				}
				fmt.Fprintf(&buf, "- `--%s`: %s\n", flag.Name, strings.TrimSpace(flag.Usage))
			})
			buf.WriteString("\n")
		}

		for _, child := range cmd.Commands() {
			if child.IsAdditionalHelpTopicCommand() {
				continue
			}
			walk(child, level+1)
		}
	}

	walk(root, 2)
	return buf.String(), nil
}
