package portal

import (
	"fmt"
	"strings"
)

const blue = "\x1b[1;38;2;18;58;188m"
const cobalt = "\x1b[1;38;2;37;90;230m"
const cyan = "\x1b[38;2;92;167;255m"
const mist = "\x1b[38;2;137;188;255m"
const reset = "\x1b[0m"

func Render(color bool) string {
	art := []string{
		`            .-----------------------------.`,
		`         .'                                 '.`,
		`       .'     ▄▄▄▄▄  ▄   ▄  ▄▄▄▄▄            '.`,
		`      /       █      ▀▄ ▄▀  █   █               \`,
		`     |        █▄▄▄    ▀█▀   █▄▄▄█                |`,
		`      \       █      ▄▀ ▀▄  █   █               /`,
		`       '.     ▀▀▀▀▀  ▀   ▀  ▀   ▀            .'`,
		`         '.                                 .'`,
		`            '-----------------------------'`,
	}

	lines := make([]string, 0, len(art)+4)
	for idx, line := range art {
		if color {
			palette := []string{mist, cyan, cobalt, blue, blue, blue, cobalt, cyan, mist}
			lines = append(lines, palette[idx%len(palette)]+line+reset)
			continue
		}
		lines = append(lines, line)
	}

	taglines := []string{
		"search. answer. code context. research.",
		"built for shells, prompts, and long threads.",
	}
	lines = append(lines, "")
	if color {
		lines = append(lines, fmt.Sprintf("%s%s%s", cyan, taglines[0], reset))
		lines = append(lines, fmt.Sprintf("%s%s%s", mist, taglines[1], reset))
		return strings.Join(lines, "\n")
	}
	lines = append(lines, taglines...)
	return strings.Join(lines, "\n")
}
