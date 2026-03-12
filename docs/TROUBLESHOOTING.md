# Troubleshooting

## Auth Errors

If a command returns an auth error:

- run `exa-cli auth status`
- if you want the key stored locally, run `exa-cli auth login`
- check `EXA_API_KEY` only if you expect env-based auth

Environment variables override config. If the wrong key is in your shell, fixing `config.toml` will not help until the env var is removed.

```bash
unset EXA_API_KEY
```

## Banner Or Color In Automation

`exa-cli` suppresses the portal banner automatically for non-TTY output, JSON, JSONL, and CI-style contexts. If you still want to force plain output:

```bash
exa-cli --no-banner --no-color find "query"
```

You can make that sticky with:

```bash
export EXA_CLI_NO_BANNER=1
export NO_COLOR=1
```

## Cache Confusion

If output looks stale:

```bash
exa-cli cache stats
exa-cli cache prune
exa-cli find "query" --no-cache
exa-cli read https://exa.ai/docs/reference/search --max-age-hours 0
```

Use `read --max-age-hours 0` when you need a fresh retrieval rather than Exa-managed caching behavior.

## JSON Contract Checks

If you are wiring `exa-cli` into another tool and want to confirm the output shape:

```bash
exa-cli find "query" --format json
exa-cli find "query" --format jsonl
exa-cli doctor --format json
```

The JSON envelope is always `meta` plus `data`. JSONL always emits a `meta` record first.

## Testing Failures

If verification starts disagreeing with the checked-in docs or expected CLI surface:

```bash
make test
make docs
go test -cover ./...
make battle-test
```

If the live checks are relevant for the change:

```bash
make test-live
make battle-test-live
```

Those commands use the locally stored `exa-cli` key if `EXA_API_KEY` is unset.

Use [TESTING.md](TESTING.md) for the full testing matrix and release-check sequence.

## Raw Endpoint Escape Hatch

If a smart command does not yet expose the body shape you need, use `raw request` instead of trying to smuggle JSON into flags:

```bash
exa-cli raw request --method POST --path /search --input request.json
printf '%s\n' '{"query":"golang sqlite wal"}' | exa-cli raw request --method POST --path /search --input -
```

That keeps the CLI predictable while still letting you reach the full Exa API surface.
