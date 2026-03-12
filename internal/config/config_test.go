package config_test

import (
	"path/filepath"
	"testing"

	"github.com/samaydhawan/exa-cli/internal/config"
)

func TestLoadMissingReturnsDefaults(t *testing.T) {
	cfg, err := config.Load(filepath.Join(t.TempDir(), "missing.toml"))
	if err != nil {
		t.Fatalf("load missing: %v", err)
	}
	if cfg.BaseURL != config.DefaultBaseURL {
		t.Fatalf("unexpected base url: %+v", cfg)
	}
	if cfg.Format != config.DefaultFormat {
		t.Fatalf("unexpected format: %+v", cfg)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	want := config.Config{
		APIKey:          "test-key",
		BaseURL:         "https://example.com",
		Format:          "json",
		Profile:         "fast",
		NoBanner:        true,
		CacheTTLMinutes: 5,
		CachePath:       "/tmp/cache.db",
		MCPURL:          "https://mcp.example.com/mcp",
	}
	if err := config.Save(path, want); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := config.Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got != want {
		t.Fatalf("round trip mismatch:\nwant=%+v\ngot=%+v", want, got)
	}
}
