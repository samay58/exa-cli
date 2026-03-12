# Claude Code Hammer Test

Use this when you want Claude Code to behave like a real `exa-cli` operator, not just a unit-test runner. The loop is: ground in the repo, run the actual binary, log friction, fix the highest-leverage issues, rerun the matrix, and leave the repo cleaner than it started.

This prompt assumes the session starts in `~/exa-cli`.

## When To Use It

- after meaningful CLI or UX changes
- before a dogfooding push
- when the binary feels slower, noisier, or less helpful in agent workflows than the tests suggest
- when you want the agent to pressure-test `exa-cli` by using it recursively to improve itself

## Copy-Paste Prompt

```text
You are starting in ~/exa-cli.

Read ./AGENTS.md first and follow it exactly. This repo uses bd for issue tracking, and session completion requires the AGENTS landing-plane workflow. Do not ignore that workflow. If push is impossible because no remote is configured, state that explicitly at the end instead of pretending the push happened.

Your goal in this session is not just to run tests. Your goal is to behave like a real operator and like an agentic CLI user inside your own workflow, then tighten exa-cli until the loop is materially better for both humans and coding agents.

Context you should internalize before changing anything:
- exa-cli is a clean-slate, agent-native Exa CLI
- the design bar is very high; avoid generic copy, lazy defaults, fake confidence, and ornamental complexity
- the CLI should optimize for maximum juice per squeeze inside real coding-agent sessions
- the most important product truth is not what the code claims; it is what the compiled binary feels like when you actually use it
- the launch and home experience matters, but only if it improves the real workflow instead of becoming decorative sludge

Ground yourself first:
1. Read:
   - ./README.md
   - ./docs/TESTING.md
   - ./docs/ARCHITECTURE.md
   - ./docs/WORKFLOWS.md
2. Inspect:
   - ./internal/app/root.go
   - ./internal/portal/portal.go
   - ./scripts/battle-test.sh
3. Run:
   - bd onboard
   - bd ready
   - git status --short --branch
   - git remote -v

Operating rules for this session:
- use the real binary, not just unit tests
- prefer ./bin/exa-cli unless you intentionally need go run
- assume local auth may already exist in exa-cli config; do not move secrets into ~/.zshrc or shell RC files
- if EXA_API_KEY is unset, still test the CLI through its stored config path where possible
- every time you discover friction, classify it as one of:
  - bug
  - bad default
  - output contract issue
  - docs mismatch
  - testing gap
  - UX/design regression
- be ruthless about signal; cut noise before adding anything

Phase 1: Baseline truth
1. Build and run the current automated baseline:
   - make build
   - make test
   - make docs
   - go test -cover ./...
   - make battle-test
   - make test-live
   - make battle-test-live
2. If a live test cannot run because auth or network is missing, state the exact blocker and continue with the rest of the session.
3. Do not trust green tests as proof of product quality. They are only the baseline.

Phase 2: Real operator and agent loops
Use exa-cli as if it were one of your own core tools. Drive it manually and recursively. At minimum, exercise:
- home screen:
  - ./bin/exa-cli
  - ./bin/exa-cli --format json
- doctor and auth:
  - env -u EXA_API_KEY ./bin/exa-cli doctor
  - env -u EXA_API_KEY ./bin/exa-cli auth status --format json
- search, answer, code, and read:
  - env -u EXA_API_KEY ./bin/exa-cli find "Exa search API docs" --num-results 3 --format json --no-cache
  - env -u EXA_API_KEY ./bin/exa-cli find "test" --profile instant --format json --no-cache (verify --profile works with local flag)
  - env -u EXA_API_KEY ./bin/exa-cli ask "How does Exa answer differ from search?" --format markdown --no-cache
  - env -u EXA_API_KEY ./bin/exa-cli code "How should a Go Cobra app split smart and raw commands?" --format llm --no-cache
  - env -u EXA_API_KEY ./bin/exa-cli read https://exa.ai/docs/reference/search --summary --format markdown --no-cache
  - env -u EXA_API_KEY ./bin/exa-cli read https://exa.ai/docs/reference/search --summary --text --format markdown --no-cache
- research, MCP, and config:
  - env -u EXA_API_KEY ./bin/exa-cli research list --limit 3 --format json --no-cache
  - ./bin/exa-cli mcp print codex
  - ./bin/exa-cli mcp doctor --format json
  - ./bin/exa-cli config show --format markdown
- failure and machine-contract probes:
  - ./bin/exa-cli version --format json
  - ./bin/exa-cli version --format yaml and confirm clean failure
  - env -u EXA_API_KEY ./bin/exa-cli find "test" --exclude-domain example.com (should succeed, not reject)
  - malformed raw request input
  - JSON piping through jq when available
  - NO_COLOR=1
  - EXA_CLI_NO_BANNER=1

Do not stop at canned commands. Use exa-cli to help you think about improving exa-cli:
- use find to locate Exa docs relevant to observed UX or API issues
- use read to inspect those docs
- use code to gather implementation-oriented context for changes you are considering
- use ask when you want Exa to synthesize a distinction or tradeoff quickly
Then judge whether exa-cli actually helped your own session or got in your way.

Phase 3: Friction ledger and prioritization
Create a tight ledger of everything you observed.
For each item capture:
- command used
- actual behavior
- why it hurts a human and/or agent loop
- severity
- category from the classification list above
Then choose only the highest-leverage fixes. Favor:
- tighter operator feedback loops
- better prompt-handoff quality from code and llm
- better shell composability and machine contracts
- clearer docs when real behavior was surprising
- truthful doctor, auth, and config reporting
- home and portal improvements only if they materially improve first-run UX and do not devolve into decorative sludge

Phase 4: Fix, verify, repeat
Implement the small set of highest-leverage changes.
After each meaningful change:
- run the narrowest relevant tests
- re-run the affected real CLI commands
- confirm the behavior is actually better in practice
Then rerun the full quality bar:
- make test
- make docs
- go test -cover ./...
- make battle-test
- make test-live
- make battle-test-live
If behavior changed, update docs so they match the product exactly. Keep docs human, crisp, and free of filler.

Issue tracking:
- use bd for anything that remains open
- if a problem matches an existing issue, update that issue rather than creating a duplicate
- create new issues only for genuinely new gaps discovered in this session

Quality bar:
- no lazy "looks fine" calls
- no giant change sprawl
- no speculative fixes without real CLI validation
- no docs that describe an idealized product instead of the current binary
- no fake completion if remote and push requirements cannot actually be satisfied
- no banner-font or boxed-portal regressions

Final output format:
1. Findings first, ordered by severity, with file references when relevant
2. Changes made and why they were the highest leverage
3. Exact tests and real CLI commands run
4. Remaining open issues and newly filed bd issues
5. Git state:
   - whether changes were committed
   - whether git pull --rebase succeeded
   - whether bd sync ran
   - whether git push succeeded
   - if push did not succeed, the exact blocker

Be direct. Be skeptical. Use the tool like you mean it.
```

## Expected Outcome

The session should produce more than a green test run. A good pass should:

- expose where `exa-cli` slows down or muddies an agent loop
- tighten the binary, docs, or defaults in response
- leave behind a cleaner testing trail for the next session
- record any remaining gaps in `bd` instead of hand-waving them away

The loop is reusable for future agent-native CLIs: keep the friction-ledger, fix, and rerun discipline, then swap in the product-specific commands and docs for the next tool.
