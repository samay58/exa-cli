package cache

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Stats struct {
	Entries int64 `json:"entries"`
	Bytes   int64 `json:"bytes"`
}

type Store struct {
	db   *sql.DB
	path string
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	store := &Store{db: db, path: path}
	if err := store.init(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *Store) Path() string {
	return s.path
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) init() error {
	const pragmas = `
PRAGMA journal_mode = WAL;
PRAGMA busy_timeout = 5000;`

	if _, err := s.db.Exec(pragmas); err != nil {
		return err
	}

	const schema = `
CREATE TABLE IF NOT EXISTS entries (
  kind TEXT NOT NULL,
  cache_key TEXT NOT NULL,
  payload BLOB NOT NULL,
  created_at TEXT NOT NULL,
  expires_at TEXT NOT NULL,
  PRIMARY KEY (kind, cache_key)
);

CREATE TABLE IF NOT EXISTS runs (
  run_id TEXT PRIMARY KEY,
  kind TEXT NOT NULL,
  payload BLOB NOT NULL,
  updated_at TEXT NOT NULL
);`

	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) Get(ctx context.Context, kind, key string, target any) (bool, error) {
	const query = `
SELECT payload
FROM entries
WHERE kind = ? AND cache_key = ? AND expires_at > ?`

	var payload []byte
	err := s.db.QueryRowContext(ctx, query, kind, key, time.Now().UTC().Format(time.RFC3339Nano)).Scan(&payload)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	if err := json.Unmarshal(payload, target); err != nil {
		return false, err
	}

	return true, nil
}

func (s *Store) Put(ctx context.Context, kind, key string, value any, ttl time.Duration) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(ttl)

	const query = `
INSERT INTO entries (kind, cache_key, payload, created_at, expires_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(kind, cache_key) DO UPDATE SET
  payload = excluded.payload,
  created_at = excluded.created_at,
  expires_at = excluded.expires_at`

	_, err = s.db.ExecContext(
		ctx,
		query,
		kind,
		key,
		payload,
		now.Format(time.RFC3339Nano),
		expiresAt.Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) PutRun(ctx context.Context, runID, kind string, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	const query = `
INSERT INTO runs (run_id, kind, payload, updated_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(run_id) DO UPDATE SET
  kind = excluded.kind,
  payload = excluded.payload,
  updated_at = excluded.updated_at`

	_, err = s.db.ExecContext(ctx, query, runID, kind, payload, time.Now().UTC().Format(time.RFC3339Nano))
	return err
}

func (s *Store) GetRun(ctx context.Context, runID string, target any) (bool, error) {
	const query = `SELECT payload FROM runs WHERE run_id = ?`
	var payload []byte
	err := s.db.QueryRowContext(ctx, query, runID).Scan(&payload)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	if err := json.Unmarshal(payload, target); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) Clear(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM entries; DELETE FROM runs;`)
	return err
}

func (s *Store) Prune(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(
		ctx,
		`DELETE FROM entries WHERE expires_at <= ?`,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *Store) Stats(ctx context.Context) (Stats, error) {
	row := s.db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(SUM(LENGTH(payload)), 0) FROM entries`)
	var stats Stats
	if err := row.Scan(&stats.Entries, &stats.Bytes); err != nil {
		return Stats{}, err
	}
	return stats, nil
}

func (s *Store) String() string {
	return fmt.Sprintf("sqlite:%s", s.path)
}
