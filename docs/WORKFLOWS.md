# Workflows

## 1. Fast Search

Use `find` for almost everything before reaching for `ask` or `research`.

```bash
exa-cli find "Go CLI docs generation best practices"
exa-cli find "React Compiler caveats" --profile fast
exa-cli find "AI infra companies building eval tooling" --category company
```

Use `--format json` when another tool or agent needs a stable envelope.

## 2. Direct Answering

Use `ask` when you want Exa to synthesize an answer instead of returning a ranked list.

```bash
exa-cli ask "How does Exa answer differ from search?"
```

Use `find` when source discovery matters more than synthesis.

## 3. Coding-Agent Context

Use `code` when the task is implementation-oriented and you want compact context for a model prompt.

```bash
exa-cli code "How should a Go CLI handle raw and smart commands cleanly?"
exa-cli code "OpenAI streaming in Python SDK" --tokens 5000
```

The default output is `llm`, which is intentionally optimized for direct use in agent prompts.

## 4. Page Retrieval

Use `read` when you already know the source and want grounded page retrieval without rerunning search.

```bash
exa-cli read https://exa.ai/docs/reference/search --summary
exa-cli read https://exa.ai/docs/reference/get-contents --summary --text
exa-cli read https://exa.ai/docs/reference/get-contents --max-age-hours 0
```

`read --summary` stays concise by default. Add `--text` only when you want the full page body.
`--max-age-hours 0` forces a live crawl.
`--max-age-hours -1` prefers cache.

## 5. Long Research

Use `research run` when the task is too broad or too deep for a single search or answer request.

```bash
exa-cli research run "design a search CLI for coding agents"
exa-cli research run "AI code search market map" --detach
exa-cli research get exa_research_123
exa-cli research list --limit 20
```

The default mode waits and prints progress for human operators. Use `--detach` when an external orchestrator will manage task state.

## 6. MCP Setup

Use `mcp print` to generate the right snippet for your client instead of hand-writing config.

```bash
exa-cli mcp print codex
exa-cli mcp print claude-code --all-tools
exa-cli mcp doctor
```

`mcp print` is keyless by default. Use `--embed-key` only when you explicitly want a secret-bearing snippet.

## 7. Raw Control

Use `raw` when you want exact endpoint control or to test a new Exa body shape before a smart command supports it.

```bash
exa-cli raw request --method POST --path /search --input request.json
exa-cli raw search --input request.json
exa-cli raw contents https://exa.ai/docs/reference/search
exa-cli raw context "golang sqlite busy timeout"
exa-cli raw research get research_id
```
