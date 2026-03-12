package cache_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/samaydhawan/exa-cli/internal/cache"
)

func TestStorePutGetAndPrune(t *testing.T) {
	store, err := cache.Open(filepath.Join(t.TempDir(), "cache.db"))
	if err != nil {
		t.Fatalf("open cache: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	type payload struct {
		Message string `json:"message"`
	}

	if err := store.Put(t.Context(), "search", "abc", payload{Message: "hello"}, time.Hour); err != nil {
		t.Fatalf("put: %v", err)
	}

	var got payload
	hit, err := store.Get(t.Context(), "search", "abc", &got)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !hit || got.Message != "hello" {
		t.Fatalf("unexpected cache value: hit=%t got=%+v", hit, got)
	}

	if err := store.Put(t.Context(), "search", "expired", payload{Message: "bye"}, -time.Minute); err != nil {
		t.Fatalf("put expired: %v", err)
	}

	rows, err := store.Prune(t.Context())
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if rows == 0 {
		t.Fatalf("expected prune to delete at least one row")
	}

	stats, err := store.Stats(t.Context())
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if stats.Entries == 0 {
		t.Fatalf("expected remaining entries after prune")
	}
}
