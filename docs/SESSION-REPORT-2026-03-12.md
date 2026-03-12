# Session Report: Operator Dogfooding and Tightening

Date: 2026-03-12
Binary: `./bin/exa-cli` (dev build)
Auth: config-stored key via `exa-cli auth login`, verified with `env -u EXA_API_KEY`

## Baseline

All automated checks passed before changes:

| Check | Result |
|-------|--------|
| `make test` | PASS (all packages) |
| `make docs` | PASS (snapshot matches) |
| `go test -cover ./...` | PASS (app: 61.6%) |
| `make battle-test` | PASS (15/15 offline) |
| `make test-live` | PASS |
| `make battle-test-live` | PASS (12/12 live + 15/15 offline) |

## Friction Ledger

### Finding 1: LLM format drops per-result content (HIGH, output contract)

**Commands**: `read --summary --format llm`, `find --format llm`

**Before**: LLM format rendered sources as bare title + URL only. An agent calling `read --summary --format llm` got:
```
Sources:
1. Search - Exa
   https://exa.ai/docs/reference/search
```

The summary (the entire point of `--summary`) was discarded.

**After**: LLM format now includes per-result summaries and highlights:
```
Sources:
1. Search - Exa
   https://exa.ai/docs/reference/search
   API documentation for Exa search endpoint covering query parameters...
```

**Why this matters**: The `llm` format exists specifically for agent consumption. Dropping the most valuable content from the format designed for agents is a direct product failure. Both `find` (highlights) and `read` (summaries) now include per-result content.

**Files changed**: `internal/app/render.go` (lines 32-36)
**Tests added**: `TestRenderLLMIncludesResultSnippets`, `TestRenderLLMIncludesHighlightsWhenNoSnippet`

### Finding 2: --exclude-domain blocked for regular searches (HIGH, bug)

**Command**: `find "test" --exclude-domain example.com`

**Before**: Failed with exit 2: `company search does not support --exclude-domain in Exa's current API`

The `validateCategoryFilters` function grouped `case "", "company":` in a switch, so any search without an explicit `--category` was treated as a company search for validation purposes.

**After**: Only `--category company` blocks `--exclude-domain`. Regular searches pass through to the API.

**Files changed**: `internal/app/helpers.go` (line 162)
**Tests added**: `TestValidateCategoryFiltersAllowsExcludeDomainWithoutCategory`, `TestValidateCategoryFiltersBlocksExcludeDomainForCompany`

### Finding 3: No `make install` target (MEDIUM, UX)

**Symptom**: Must use `./bin/exa-cli` everywhere instead of `exa-cli`.

**Fix**: Added `make install` target that copies the built binary to `/usr/local/bin/exa-cli`.

**File changed**: `Makefile`

## Commands Exercised During Dogfooding

All commands run against the compiled binary with `env -u EXA_API_KEY` to force config-path auth:

- `./bin/exa-cli` (home screen, table)
- `./bin/exa-cli --format json` (home screen, JSON envelope)
- `./bin/exa-cli version --format json` (version, JSON)
- `./bin/exa-cli version --format yaml` (clean failure, exit 2)
- `env -u EXA_API_KEY ./bin/exa-cli doctor` (doctor, table)
- `env -u EXA_API_KEY ./bin/exa-cli auth status --format json` (auth, JSON)
- `./bin/exa-cli config show --format markdown` (config, markdown)
- `./bin/exa-cli config show --format json` (config, JSON)
- `env -u EXA_API_KEY ./bin/exa-cli find "Exa search API docs" --num-results 3 --format json --no-cache` (live search)
- `env -u EXA_API_KEY ./bin/exa-cli ask "How does Exa answer differ from search?" --format markdown --no-cache` (live answer)
- `env -u EXA_API_KEY ./bin/exa-cli code "How should a Go Cobra app split smart and raw commands?" --format llm --no-cache` (live context)
- `env -u EXA_API_KEY ./bin/exa-cli read https://exa.ai/docs/reference/search --summary --format markdown --no-cache` (live contents)
- `env -u EXA_API_KEY ./bin/exa-cli read https://exa.ai/docs/reference/search --summary --text --format markdown --no-cache` (live contents with text)
- `env -u EXA_API_KEY ./bin/exa-cli read https://exa.ai/docs/reference/search --summary --format llm --no-cache` (verified fix)
- `env -u EXA_API_KEY ./bin/exa-cli research list --limit 3 --format json --no-cache` (live research list)
- `./bin/exa-cli mcp print codex` (MCP snippet)
- `./bin/exa-cli mcp doctor --format json` (MCP doctor)
- `NO_COLOR=1 EXA_CLI_NO_BANNER=1 env -u EXA_API_KEY ./bin/exa-cli` (env var probes)
- `env -u EXA_API_KEY ./bin/exa-cli find "test" --exclude-domain example.com` (verified fix)
- `env -u EXA_API_KEY ./bin/exa-cli find "test" --exclude-domain example.com --category company` (correct rejection)
- Malformed JSON to `raw request --input -` (clean error)
- JSON output piped through `jq` (clean parsing)
- Recursive self-referential usage: used `find` and `read` to look up Exa API docs for exclude-domain semantics

## Self-Referential Usage Assessment

exa-cli was genuinely useful during this session:
- `find` located the correct Exa API docs pages
- `read --summary` gave concise page overviews without leaving the terminal
- `code` returned relevant Cobra CLI patterns from real articles
- `ask` synthesized a coherent answer about Exa search vs answer

The LLM format fix was the most impactful change: before the fix, `read --summary --format llm` was effectively broken for agent workflows (returned only URLs). After the fix, an agent gets the actual summary content it requested.

## Final Quality Gate

| Check | Result |
|-------|--------|
| `make test` | PASS |
| `make docs` | PASS (snapshot matches) |
| `go test -cover ./...` | PASS (app: 62.8%, up from 61.6%) |
| `make battle-test` | PASS (15/15) |
| `make test-live` | PASS |
| `make battle-test-live` | PASS (27/27) |

## Changes Summary

| File | Change |
|------|--------|
| `internal/app/render.go` | LLM format now includes per-result summaries and highlights |
| `internal/app/helpers.go` | Fixed validateCategoryFilters to allow --exclude-domain for uncategorized search |
| `internal/app/output_test.go` | Added 4 tests covering both fixes |
| `Makefile` | Added `make install` target |

## Open Issues (pre-existing, not changed)

- `exa-cli-rzi`: Add dedicated company and people commands (P2)
- `exa-cli-5mb`: Automate signed release artifacts and Homebrew publishing (P2)
- `exa-cli-rkr`: Add terminal-safe device-flow auth (P2)
- `exa-cli-jun`: Add live smoke coverage for research create/cancel (P2)
