package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	toml "github.com/pelletier/go-toml/v2"
)

const (
	DefaultBaseURL         = "https://api.exa.ai"
	DefaultMCPURL          = "https://mcp.exa.ai/mcp"
	DefaultFormat          = "table"
	DefaultProfile         = "balanced"
	DefaultCacheTTLMinutes = 720
)

type Config struct {
	APIKey          string `toml:"api_key"`
	BaseURL         string `toml:"base_url"`
	Format          string `toml:"format"`
	Profile         string `toml:"profile"`
	NoBanner        bool   `toml:"no_banner"`
	CacheTTLMinutes int    `toml:"cache_ttl_minutes"`
	CachePath       string `toml:"cache_path"`
	MCPURL          string `toml:"mcp_url"`
}

func Default() Config {
	return Config{
		BaseURL:         DefaultBaseURL,
		Format:          DefaultFormat,
		Profile:         DefaultProfile,
		CacheTTLMinutes: DefaultCacheTTLMinutes,
		MCPURL:          DefaultMCPURL,
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	if cfg.Format == "" {
		cfg.Format = DefaultFormat
	}
	if cfg.Profile == "" {
		cfg.Profile = DefaultProfile
	}
	if cfg.CacheTTLMinutes <= 0 {
		cfg.CacheTTLMinutes = DefaultCacheTTLMinutes
	}
	if cfg.MCPURL == "" {
		cfg.MCPURL = DefaultMCPURL
	}

	return cfg, nil
}

func Save(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0o600)
}

func DefaultConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "exa-cli", "config.toml"), nil
}

func DefaultCachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "exa-cli", "cache.db"), nil
}

func CacheTTL(cfg Config) time.Duration {
	minutes := cfg.CacheTTLMinutes
	if minutes <= 0 {
		minutes = DefaultCacheTTLMinutes
	}
	return time.Duration(minutes) * time.Minute
}
