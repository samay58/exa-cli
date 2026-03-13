package portal

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestRenderPlainPortalHasNoANSI(t *testing.T) {
	output := Render(false)
	if strings.Contains(output, "\x1b[") {
		t.Fatalf("did not expect ANSI escapes in plain portal output:\n%s", output)
	}
	if !strings.Contains(output, "\u2584\u2584\u2584\u2584\u2584") || !strings.Contains(output, "built for shells, prompts, and long threads.") {
		t.Fatalf("expected portal art and tagline, got:\n%s", output)
	}
}

func TestRenderColorPortalIncludesANSI(t *testing.T) {
	output := Render(true)
	if !strings.Contains(output, "\x1b[1;38;2;18;58;188m") {
		t.Fatalf("expected Exa blue ANSI sequence, got:\n%s", output)
	}
	if !strings.Contains(output, "search. answer. code context. research.") {
		t.Fatalf("expected tagline, got:\n%s", output)
	}
}

func TestRenderPortalStaysCompact(t *testing.T) {
	output := Render(false)
	for _, line := range strings.Split(output, "\n") {
		if utf8.RuneCountInString(line) > 52 {
			t.Fatalf("expected compact portal, found %d-column line:\n%s", utf8.RuneCountInString(line), line)
		}
	}
}
