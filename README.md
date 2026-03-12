# exa-cli

`exa-cli` is a clean-slate, agent-native command line interface for Exa. It is designed for fast human use, reliable shell composition, and high-signal workflows inside coding agents like Codex and Claude Code.

## Quick Start

Build locally:

```bash
make build
```

If you have not installed `exa-cli` onto `PATH`, use `./bin/exa-cli` in the commands below.

Authenticate locally with an API key:

```bash
./bin/exa-cli auth login
```

Or pipe it without echoing it into shell history:

```bash
printf '%s\n' "your-key" | ./bin/exa-cli auth login --stdin
```

Use `EXA_API_KEY` for CI, one-off shells, or when you do not want to write local config:

```bash
EXA_API_KEY="your-key" ./bin/exa-cli find "query"
```

Run the core workflows:

```bash
./bin/exa-cli find "best practices for Go CLI UX"
./bin/exa-cli ask "How does Exa answer differ from search?"
./bin/exa-cli code "How should I structure a Cobra app with raw and smart commands?"
./bin/exa-cli read https://exa.ai/docs/reference/search --summary
./bin/exa-cli research run "design a search CLI for coding agents"
./bin/exa-cli mcp print codex
```

Add `--text` to `read` only when you want the full page body:

```bash
./bin/exa-cli read https://exa.ai/docs/reference/search --summary --text
```

## Design Goals

The goal is not to mirror every surface area of larger research CLIs. The goal is to be sharper:

- one binary name
- one primary search verb
- deterministic JSON and JSONL output
- explicit latency and cost profiles
- first-class code-context workflows
- lightweight local caching
- documentation that stays aligned with the binary

## Why This Exists

Many research CLIs accumulate too many install surfaces, too many command families, and too much orchestration burden for agents.

`exa-cli` takes the opposite position:

- `find` is the center of gravity
- `research run` hides the start/poll dance unless you explicitly detach
- `raw` commands preserve exact API control without forcing every workflow to feel low-level
- MCP helpers emit copy-pasteable client snippets instead of assuming one editor or agent
- config is small, local, and inspectable

## Current V1 Surface

Smart commands:

- `exa-cli find <query>`
- `exa-cli ask <question>`
- `exa-cli code <query>`
- `exa-cli read <url-or-id...>`
- `exa-cli research run <query>`
- `exa-cli research get <id>`
- `exa-cli research list`
- `exa-cli research cancel <id>`
- `exa-cli mcp print <codex|claude-code|cursor|generic>`
- `exa-cli mcp doctor`
- `exa-cli auth status|login|logout`
- `exa-cli doctor`
- `exa-cli cache stats|prune|clear`
- `exa-cli config show|path|init`

Raw commands:

- `exa-cli raw request`
- `exa-cli raw search`
- `exa-cli raw answer`
- `exa-cli raw contents`
- `exa-cli raw context`
- `exa-cli raw research start|get|cancel`

## Output Modes

Every high-signal command supports:

- `--format table` for fast human scanning
- `--format markdown` for readable terminal output and prompt handoff
- `--format json` for stable envelopes
- `--format jsonl` for record streams
- `--format llm` for direct agent consumption

The JSON envelope always has:

- `meta`: command, profile, cache state, request ID, timing/cost hints
- `data`: endpoint-specific payload

## Config And Cache

Defaults:

- config path: `os.UserConfigDir()/exa-cli/config.toml`
- cache path: `os.UserCacheDir()/exa-cli/cache.db`

Typical paths:

- macOS config: `~/Library/Application Support/exa-cli/config.toml`
- macOS cache: `~/Library/Caches/exa-cli/cache.db`
- Linux config: `~/.config/exa-cli/config.toml`
- Linux cache: `~/.cache/exa-cli/cache.db`

Precedence:

- flags
- environment variables
- config file

Important environment variables:

- `EXA_API_KEY`
- `EXA_BASE_URL`
- `EXA_CLI_CONFIG`
- `EXA_CLI_FORMAT`
- `EXA_CLI_PROFILE`
- `EXA_CLI_NO_BANNER`
- `EXA_CLI_NO_CACHE`
- `EXA_MCP_URL`
- `NO_COLOR`

## MCP Setup

`exa-cli mcp print` is keyless by default. That keeps secrets out of shell history, copied config, and screenshots.

```bash
exa-cli mcp print codex
exa-cli mcp print claude-code --all-tools
```

Only use `--embed-key` if you explicitly want a secret-bearing snippet:

```bash
exa-cli mcp print cursor --embed-key
```

## Raw Control

Use the `raw request` escape hatch when you need exact Exa API control:

```bash
exa-cli raw request --method POST --path /search --input request.json
printf '%s\n' '{"query":"golang sqlite wal"}' | exa-cli raw request --method POST --path /search --input -
```

## Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Workflows](docs/WORKFLOWS.md)
- [Caching And Costs](docs/CACHING-AND-COSTS.md)
- [Testing](docs/TESTING.md)
- [Claude Code Hammer Test](docs/CLAUDE-CODE-HAMMER-TEST.md)
- [Troubleshooting](docs/TROUBLESHOOTING.md)
- [Roadmap](docs/ROADMAP.md)
- [CLI Reference](docs/REFERENCE.md)

Regenerate the reference docs from the live command tree:

```bash
make docs
```

## Testing

The project ships with three layers of verification:

- `make test` for fast deterministic coverage
- `go test -cover ./...` for coverage and snapshot checks
- `make battle-test` for the compiled-binary smoke matrix

For a deeper agentic dogfooding pass, use the maintained Claude Code workflow prompt in [docs/CLAUDE-CODE-HAMMER-TEST.md](docs/CLAUDE-CODE-HAMMER-TEST.md).

Opt-in live API verification:

```bash
make test-live
make battle-test-live
```

Both live commands use the key stored by `exa-cli auth login` if `EXA_API_KEY` is not already set.

The full testing workflow is documented in [docs/TESTING.md](docs/TESTING.md).
