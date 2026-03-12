# Testing

This project has three testing layers:

- unit and command tests for deterministic behavior
- live API tests for Exa endpoint verification
- battle tests for real operator and agent workflows through the compiled binary

Use all three before calling the CLI ready for wider dogfooding.

## Fast Local Loop

Run the standard local checks:

```bash
make test
make docs
go test -cover ./...
```

What this covers:

- command routing and validation
- render formats and JSONL contracts
- cache and config behavior
- MCP snippet generation
- portal rendering
- generated docs staying in sync with the command tree

## Live API Tests

Live API tests are opt-in because they spend real Exa API calls.

Run them like this:

```bash
make test-live
```

The live suite uses the key stored by `exa-cli auth login` if `EXA_API_KEY` is not already set.

Optional environment variables:

- `EXA_BASE_URL` if you need to point the client at a non-default API host

Use `-v` if you want more test output while debugging failures.

What the live suite currently checks:

- `/search`
- `/answer`
- `/context`
- `/contents`
- `/research/v1` list

It intentionally does not create or cancel research jobs by default. That is better exercised in battle tests and manual verification because those flows can be slower and more spend-sensitive.

## Battle-Test Smoke Matrix

The battle-test harness compiles the CLI and drives a focused smoke matrix through the binary. It is narrower than exploratory dogfooding, but it catches the operator and agent workflows most likely to regress.

Offline battle test:

```bash
make battle-test
```

Live battle test:

```bash
make battle-test-live
```

What the offline smoke matrix checks:

- home screen and pure commands honor explicit output formats
- invalid formats fail cleanly
- `auth login --stdin` and `auth status`
- `doctor`, `config show`, `mcp print`, and `mcp doctor`
- malformed JSON input handling for `raw request`
- JSON output surviving redirects and `jq` parsing when `jq` is available
- generated reference docs match `docs/REFERENCE.md`

What the live smoke matrix checks:

- `find`
- `ask`
- `code`
- `read`
- `research list`

The live smoke matrix is intentionally conservative. It exercises real network paths without automatically creating async research jobs.

## Manual Operator Smoke

When you want to test the real CLI the way a human or coding agent will use it, run the compiled binary directly:

```bash
make build
./bin/exa-cli doctor
./bin/exa-cli auth status
./bin/exa-cli find "Exa search API docs"
./bin/exa-cli ask "How does Exa answer differ from search?"
./bin/exa-cli code "Go Cobra command tree patterns"
./bin/exa-cli read https://exa.ai/docs/reference/search --summary
./bin/exa-cli research list --limit 3
```

Add `--format json` when you want to inspect the machine contract. Add `--text` to `read` only when you need the full page body.

## Recursive Agent Hammer Test

When you want a coding agent to pressure-test the product the way a real end user would, use the maintained Claude Code prompt in [CLAUDE-CODE-HAMMER-TEST.md](CLAUDE-CODE-HAMMER-TEST.md).

That workflow is intentionally stronger than the smoke matrix:

- it starts from the compiled binary, not just `go test`
- it forces the agent to use `exa-cli` recursively inside its own session
- it requires a friction ledger before fixes are made
- it reruns the full matrix after targeted changes
- it expects `bd` issue hygiene instead of leaving observations stranded in chat

Use it when the binary feels worse than the tests, when agent workflows need tightening, or before a serious dogfooding push.

## Recommended Pre-Release Sequence

Use this sequence before shipping a meaningful change set:

1. `make test`
2. If the docs snapshot fails, run `make docs` and rerun `make test`
3. `go test -cover ./...`
4. `make battle-test`
5. `make test-live`
6. `make battle-test-live`

## Failure Handling

When something fails:

- if `make docs` or the docs snapshot test fails, regenerate docs and review the diff
- if a unit test fails, fix the behavior before trusting the smoke-matrix results
- if a live test fails, separate API drift from local regressions before changing the CLI
- if the battle-test harness fails, treat that as a workflow regression even if `go test` is green

The goal is not inflated coverage numbers. The goal is confidence that the CLI behaves predictably in scripts, terminals, and coding-agent loops.
