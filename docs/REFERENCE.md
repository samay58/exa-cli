# CLI Reference

Generated from the live command tree. Rebuild after changing command metadata.

## exa-cli

Agent-native Exa CLI for search, code context, contents, research, and MCP workflows

```bash
exa-cli [flags]
```

Flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

### exa-cli ask

Get a cited answer from Exa Answer

```bash
exa-cli ask <question>
```

Examples:

```bash
exa-cli ask "How does Exa answer differ from search?"
exa-cli ask "What are the tradeoffs between fast and deep search?" --format json
```

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

### exa-cli auth

Manage Exa API credentials

```bash
exa-cli auth
```

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli auth login

Store an Exa API key in config.toml

```bash
exa-cli auth login [flags]
```

Examples:

```bash
exa-cli auth login
exa-cli auth login --api-key "your-key"
printf '%s\n' "$EXA_API_KEY" | exa-cli auth login --stdin
```

Flags:

- `--api-key`: Exa API key to save into config.toml
- `--stdin`: Read the API key from stdin

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli auth logout

Remove the stored API key from config.toml

```bash
exa-cli auth logout
```

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli auth status

Show current auth source and cache path

```bash
exa-cli auth status
```

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

### exa-cli cache

Inspect or maintain the local SQLite cache

```bash
exa-cli cache
```

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli cache clear

Delete all cache entries and tracked runs

```bash
exa-cli cache clear
```

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli cache prune

Remove expired cache entries

```bash
exa-cli cache prune
```

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli cache stats

Show cache entry count and payload size

```bash
exa-cli cache stats
```

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

### exa-cli code

Fetch code-focused context for coding agents

```bash
exa-cli code <query> [flags]
```

Examples:

```bash
exa-cli code "OpenAI Python streaming client patterns"
exa-cli code "SQLite WAL busy timeout in Go" --tokens 5000
```

Flags:

- `--tokens`: Token budget for code context. Use "dynamic" or a positive integer like 5000

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

### exa-cli config

Inspect and initialize exa-cli config

```bash
exa-cli config
```

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli config init

Write a starter config.toml

```bash
exa-cli config init [flags]
```

Flags:

- `--force`: Overwrite an existing config.toml

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli config path

Print the config file path

```bash
exa-cli config path
```

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli config show

Print the resolved config file values

```bash
exa-cli config show
```

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

### exa-cli doctor

Check local configuration, cache, and output behavior

```bash
exa-cli doctor
```

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

### exa-cli find

Search Exa with agent-friendly defaults

```bash
exa-cli find <query> [flags]
```

Examples:

```bash
exa-cli find "Go CLI design patterns for machine-readable output"
exa-cli find "AI infra companies building eval tooling" --category company
exa-cli find "Rust sqlite WAL mode" --profile fast --format json
```

Flags:

- `--category`: Optional search category, for example company or people
- `--exclude-domain`: Exclude the provided domains
- `--include-domain`: Limit search to the provided domains
- `--num-results`: Number of results to request
- `--profile`: Latency/cost profile: instant, fast, balanced, deep, reasoned

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

### exa-cli mcp

MCP helpers for Codex, Claude Code, Cursor, and generic clients

```bash
exa-cli mcp
```

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli mcp doctor

Show hosted MCP configuration and auth guidance

```bash
exa-cli mcp doctor
```

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli mcp print

Print an MCP snippet wired to Exa hosted MCP

```bash
exa-cli mcp print <codex|claude-code|cursor|generic> [flags]
```

Examples:

```bash
exa-cli mcp print codex
exa-cli mcp print claude-code --all-tools
exa-cli mcp print cursor --embed-key
```

Flags:

- `--all-tools`: Include Exa's optional hosted MCP tools
- `--embed-key`: Inline the current API key into the generated snippet

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

### exa-cli raw

Low-level commands that map directly to Exa endpoints

```bash
exa-cli raw
```

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli raw answer

Raw passthrough for /answer

```bash
exa-cli raw answer [query-or-args] [flags]
```

Flags:

- `--input`: Path to a JSON request body or - for stdin

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli raw contents

Raw passthrough for /contents

```bash
exa-cli raw contents [query-or-args] [flags]
```

Flags:

- `--input`: Path to a JSON request body or - for stdin

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli raw context

Raw passthrough for /context

```bash
exa-cli raw context [query-or-args] [flags]
```

Flags:

- `--input`: Path to a JSON request body or - for stdin

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli raw request

Make an exact raw Exa API request

```bash
exa-cli raw request [flags]
```

Examples:

```bash
exa-cli raw request --method POST --path /search --input request.json
printf '%s\n' '{"query":"golang sqlite wal"}' | exa-cli raw request --method POST --path /search --input -
exa-cli raw request --method GET --path /research/v1/research_id
```

Flags:

- `--input`: Path to a JSON request body or - for stdin
- `--method`: HTTP method to use
- `--path`: Exa API path, for example /search

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli raw research

Raw passthrough commands for /research/v1

```bash
exa-cli raw research
```

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

##### exa-cli raw research cancel

Raw passthrough for /research/v1

```bash
exa-cli raw research cancel [query-or-args] [flags]
```

Flags:

- `--input`: Path to a JSON request body or - for stdin

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

##### exa-cli raw research get

Raw passthrough for /research/v1

```bash
exa-cli raw research get [query-or-args] [flags]
```

Flags:

- `--input`: Path to a JSON request body or - for stdin

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

##### exa-cli raw research start

Raw passthrough for /research/v1

```bash
exa-cli raw research start [query-or-args] [flags]
```

Flags:

- `--input`: Path to a JSON request body or - for stdin

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli raw search

Raw passthrough for /search

```bash
exa-cli raw search [query-or-args] [flags]
```

Flags:

- `--input`: Path to a JSON request body or - for stdin

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

### exa-cli read

Retrieve page contents via Exa Contents

```bash
exa-cli read <url-or-id...> [flags]
```

Examples:

```bash
exa-cli read https://exa.ai/docs/reference/search --summary
exa-cli read https://exa.ai/docs/reference/search --summary --text
exa-cli read https://exa.ai/docs/reference/get-contents --max-age-hours 0
```

Flags:

- `--highlights`: Include highlights when available
- `--livecrawl-timeout`: Optional livecrawl timeout in milliseconds
- `--max-age-hours`: Freshness control for contents; 0 forces live crawl, -1 forces cache
- `--summary`: Include Exa-generated summaries
- `--text`: Include full markdown text when available

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

### exa-cli research

Run or inspect async Exa research tasks

```bash
exa-cli research
```

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli research cancel

Attempt to cancel a research task

```bash
exa-cli research cancel <id>
```

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli research get

Fetch a research task by ID

```bash
exa-cli research get <id>
```

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli research list

List recent research tasks

```bash
exa-cli research list [flags]
```

Flags:

- `--limit`: Number of recent research tasks to fetch

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

#### exa-cli research run

Start a research task; waits by default

```bash
exa-cli research run <query> [flags]
```

Examples:

```bash
exa-cli research run "design a search CLI for coding agents"
exa-cli research run "map the AI code search landscape" --detach
```

Flags:

- `--detach`: Return the task ID immediately instead of waiting for completion
- `--model`: Research model: exa-research or exa-research-pro
- `--poll-interval`: Polling interval for attached research runs
- `--quiet`: Suppress progress updates while waiting for completion

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata

### exa-cli version

Print version metadata

```bash
exa-cli version
```

Inherited flags:

- `--format`: Output format: table, markdown, json, jsonl, llm
- `--no-banner`: Disable the interactive portal banner
- `--no-cache`: Bypass the local SQLite cache
- `--no-color`: Disable ANSI color output
- `--verbose`: Print additional request metadata


