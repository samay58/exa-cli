# exa-cli

<p align="center">
  <img src="cover-art/cover.png" alt="exa-cli cover art" width="100%" />
</p>

The problem with most search CLIs is that they start simple. One binary, one command. Then someone adds a subcommand for a new API endpoint. Then one for a different output format. Then a wrapper for an async workflow nobody wanted to handle manually. Six months later you have three install methods, a plugin system, and agents spend more time navigating command families than doing actual searches.

exa-cli is built against that. Everything is an extension of `find`, not a parallel track alongside it.

## Quick start

Build it:

```bash
make build
```

If you haven't put `exa-cli` on your PATH yet, use `./bin/exa-cli` for everything below.

Store your API key:

```bash
./bin/exa-cli auth login
```

Or pipe it in without touching shell history:

```bash
printf '%s\n' "your-key" | ./bin/exa-cli auth login --stdin
```

For CI or one-off shells where you don't want local config at all:

```bash
EXA_API_KEY="your-key" ./bin/exa-cli find "query"
```

Then run something:

```bash
./bin/exa-cli find "best practices for Go CLI UX"
./bin/exa-cli ask "How does Exa answer differ from search?"
./bin/exa-cli code "How should I structure a Cobra app with raw and smart commands?"
./bin/exa-cli read https://exa.ai/docs/reference/search --summary
./bin/exa-cli research run "design a search CLI for coding agents"
./bin/exa-cli mcp print codex
```

Add `--text` to `read` when you want the full page body, not just a summary:

```bash
./bin/exa-cli read https://exa.ai/docs/reference/search --summary --text
```

## Why it's designed this way

The design goal is predictability under composition. That matters more when an agent is running the tool than when a human is, because a human can read an error message and adapt. An agent just fails.

`ask` and `code` are `find` with intent applied, not separate tools. `research run` hides the async start/poll loop so you ask a question and get an answer without writing a polling loop yourself. `raw` commands drop the smart defaults entirely when you need exact API control. And `mcp print` generates copy-pasteable client config, keyless by default, because secrets don't belong in shell history or screenshots.

## Commands

**Smart commands** (opinionated defaults, human-readable output):

```
exa-cli find <query>
exa-cli ask <question>
exa-cli code <query>
exa-cli read <url-or-id...>
exa-cli research run <query>
exa-cli research get <id>
exa-cli research list
exa-cli research cancel <id>
exa-cli mcp print <codex|claude-code|cursor|generic>
exa-cli mcp doctor
exa-cli auth status|login|logout
exa-cli doctor
exa-cli cache stats|prune|clear
exa-cli config show|path|init
```

**Raw commands** (exact API access, no defaults applied):

```
exa-cli raw request
exa-cli raw search
exa-cli raw answer
exa-cli raw contents
exa-cli raw context
exa-cli raw research start|get|cancel
```

## Output formats

Every high-signal command supports `--format`:

| Flag | Best for |
|------|----------|
| `table` | Fast human scanning |
| `markdown` | Terminal output and prompt handoff |
| `json` | Stable envelopes for piping |
| `jsonl` | Record streams |
| `llm` | Direct agent consumption |

Every JSON response includes `meta` (command, profile, cache state, request ID, timing and cost hints) and `data` (endpoint payload).

## Config and cache

Default paths:

| | macOS | Linux |
|---|---|---|
| Config | `~/Library/Application Support/exa-cli/config.toml` | `~/.config/exa-cli/config.toml` |
| Cache | `~/Library/Caches/exa-cli/cache.db` | `~/.cache/exa-cli/cache.db` |

Precedence: flags > environment variables > config file.

Key environment variables:

```
EXA_API_KEY
EXA_BASE_URL
EXA_CLI_CONFIG
EXA_CLI_FORMAT
EXA_CLI_PROFILE
EXA_CLI_NO_BANNER
EXA_CLI_NO_CACHE
EXA_MCP_URL
NO_COLOR
```

## MCP setup

`mcp print` is keyless by default. That keeps secrets out of shell history, copied config, and screenshots.

```bash
exa-cli mcp print codex
exa-cli mcp print claude-code --all-tools
```

Use `--embed-key` only when you explicitly want a key baked into the snippet:

```bash
exa-cli mcp print cursor --embed-key
```

## Raw API access

When the smart commands aren't the right fit, drop down to `raw request`:

```bash
exa-cli raw request --method POST --path /search --input request.json
printf '%s\n' '{"query":"golang sqlite wal"}' | exa-cli raw request --method POST --path /search --input -
```

## Testing

Three layers:

```bash
make test              # fast, deterministic
go test -cover ./...   # coverage + snapshot checks
make battle-test       # compiled-binary smoke matrix
```

For a deeper agentic pass, use the Claude Code workflow prompt in [docs/CLAUDE-CODE-HAMMER-TEST.md](docs/CLAUDE-CODE-HAMMER-TEST.md).

Live API verification (uses key from `exa-cli auth login` if `EXA_API_KEY` isn't set):

```bash
make test-live
make battle-test-live
```

Full testing docs: [docs/TESTING.md](docs/TESTING.md).

## Docs

Regenerate the CLI reference from the live command tree:

```bash
make docs
```

- [Architecture](docs/ARCHITECTURE.md)
- [Workflows](docs/WORKFLOWS.md)
- [Caching and costs](docs/CACHING-AND-COSTS.md)
- [Testing](docs/TESTING.md)
- [Claude Code hammer test](docs/CLAUDE-CODE-HAMMER-TEST.md)
- [Troubleshooting](docs/TROUBLESHOOTING.md)
- [Roadmap](docs/ROADMAP.md)
- [CLI reference](docs/REFERENCE.md)
