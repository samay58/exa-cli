package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mattn/go-isatty"

	"github.com/samaydhawan/exa-cli/internal/config"
	"github.com/samaydhawan/exa-cli/internal/exa"
)

func (a *App) requireAPIKey() error {
	if strings.TrimSpace(a.cfg.APIKey) == "" {
		return wrap(3, errors.New("no Exa API key configured; set EXA_API_KEY or run `exa-cli auth login`"))
	}
	return nil
}

func (a *App) authSource() string {
	if strings.TrimSpace(a.env["EXA_API_KEY"]) != "" {
		return "env"
	}
	if strings.TrimSpace(a.cfg.APIKey) != "" {
		return "config"
	}
	return "none"
}

func (a *App) authWarning() string {
	envKey := strings.TrimSpace(a.env["EXA_API_KEY"])
	storedKey := strings.TrimSpace(a.storedAPIKey)
	if envKey != "" && storedKey != "" && envKey != storedKey {
		return "EXA_API_KEY overrides the stored config key for this process."
	}
	return ""
}

func (a *App) configuredDefaultFormat() string {
	if v := strings.TrimSpace(a.env["EXA_CLI_FORMAT"]); v != "" {
		return v
	}
	if v := strings.TrimSpace(a.cfg.Format); v != "" {
		return v
	}
	return config.DefaultFormat
}

func (a *App) configuredDefaultProfile() string {
	if v := strings.TrimSpace(a.env["EXA_CLI_PROFILE"]); v != "" {
		return v
	}
	if v := strings.TrimSpace(a.cfg.Profile); v != "" {
		return v
	}
	return config.DefaultProfile
}

func (a *App) cachedJSON(ctx context.Context, kind string, request any, fetch func() (map[string]any, exa.RequestMeta, error)) (map[string]any, Meta, error) {
	meta := a.newMeta("", exa.RequestMeta{})

	if err := a.ensureCache(); err != nil {
		return nil, meta, wrap(1, err)
	}

	if a.cache != nil {
		key, err := cacheKey(request)
		if err != nil {
			return nil, meta, wrap(1, err)
		}

		var cached cachedPayload
		hit, err := a.cache.Get(ctx, kind, key, &cached)
		if err != nil {
			return nil, meta, wrap(1, err)
		}
		if hit {
			meta.Cache = "hit"
			meta.RequestID = cached.Request.RequestID
			meta.SearchTime = readFloat(cached.Response, "searchTime")
			meta.CostDollars = readFloat(cached.Response, "costDollars")
			return cached.Response, meta, nil
		}

		payload, requestMeta, err := fetch()
		if err != nil {
			return nil, meta, normalizeAPIError(err)
		}
		meta = a.newMeta("", requestMeta)
		meta.Cache = "miss"
		meta.SearchTime = readFloat(payload, "searchTime")
		meta.CostDollars = readFloat(payload, "costDollars")
		_ = a.cache.Put(ctx, kind, key, cachedPayload{
			Response: payload,
			Request:  requestMeta,
			SavedAt:  time.Now().UTC().Format(time.RFC3339Nano),
		}, config.CacheTTL(a.cfg))
		return payload, meta, nil
	}

	payload, requestMeta, err := fetch()
	if err != nil {
		return nil, meta, normalizeAPIError(err)
	}
	meta = a.newMeta("", requestMeta)
	meta.Cache = "disabled"
	meta.SearchTime = readFloat(payload, "searchTime")
	meta.CostDollars = readFloat(payload, "costDollars")
	return payload, meta, nil
}

func (a *App) waitForResearch(ctx context.Context, researchID string, interval time.Duration, quiet bool) (map[string]any, Meta, error) {
	if interval <= 0 {
		interval = 5 * time.Second
	}

	var lastStatus string
	for {
		payload, requestMeta, err := a.client.ResearchGet(ctx, researchID, true)
		if err != nil {
			return nil, Meta{}, normalizeAPIError(err)
		}

		status := strings.ToLower(firstString(asString(payload["status"]), asString(payload["state"])))
		if status != "" && status != lastStatus && a.isInteractive() && !quiet {
			fmt.Fprintf(a.out, "research %s: %s\n", researchID, status)
			lastStatus = status
		}

		if status == "completed" || status == "failed" || status == "error" || status == "cancelled" {
			if err := a.ensureCache(); err == nil && a.cache != nil {
				_ = a.cache.PutRun(ctx, researchID, "research", payload)
			}
			meta := a.newMeta("", requestMeta)
			meta.SearchTime = readFloat(payload, "searchTime")
			meta.CostDollars = readFloat(payload, "costDollars")
			return payload, meta, nil
		}

		select {
		case <-ctx.Done():
			return nil, Meta{}, wrap(1, ctx.Err())
		case <-time.After(interval):
		}
	}
}

func (a *App) shouldUseColor() bool {
	return !a.options.NoColor && a.isInteractive()
}

func (a *App) isInteractive() bool {
	if a.out == nil {
		return false
	}
	type fder interface {
		Fd() uintptr
	}
	fdWriter, ok := a.out.(fder)
	if !ok {
		return false
	}
	return isatty.IsTerminal(fdWriter.Fd()) || isatty.IsCygwinTerminal(fdWriter.Fd())
}

func (a *App) newMeta(command string, requestMeta exa.RequestMeta) Meta {
	return Meta{
		Command:   command,
		RequestID: requestMeta.RequestID,
	}
}

func cacheKey(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
