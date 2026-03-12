# Architecture

## Design Principles

- Smart commands should feel sharp, not magical.
- Raw commands should stay close to the Exa API.
- Output contracts matter as much as endpoint coverage.
- Local state should help with speed and auditability, not become hidden product behavior.
- Documentation should be generated from command metadata where possible.

## Layers

### Command Layer

The root command tree is built with Cobra. It defines:

- global flags
- public command taxonomy
- command-local validation
- command help metadata used by the reference generator

### Runtime Layer

Runtime initialization resolves:

- config path
- config file
- env overrides
- API key
- cache initialization
- Exa client construction

This happens once per invocation through the root persistent pre-run hook.

### Smart Commands

Smart commands apply deterministic defaults:

- `find` maps user intent to Exa search profiles
- `code` defaults to an LLM-friendly render format
- `read` exposes freshness controls through `maxAgeHours`
- `research run` waits by default to remove manual polling overhead

These commands do not invent hidden multi-step agent logic. They only add:

- better defaults
- validation
- caching
- output shaping

### Raw Commands

The `raw` subtree is the escape hatch.

- `raw request` is the exact-control surface: method, path, and JSON body are supplied explicitly.
- Thin aliases like `raw search` and `raw contents` exist for speed, not to redefine endpoint semantics.

### Cache Layer

The cache is a local SQLite database with:

- request fingerprint keys
- JSON payload storage
- TTL-based pruning
- separate tracking for research runs

The cache is inspectable and can be disabled with `--no-cache`.

### Runtime Initialization

The runtime is intentionally split:

- lightweight startup for pure commands like `version`, root home, and docs generation
- full config/cache/client initialization only for commands that actually need it

This avoids surprising failures from config-path or cache initialization on commands that should be pure.

## Documentation Flow

- command metadata feeds `exa-cli gen docs`
- generated reference lives in `docs/REFERENCE.md`
- narrative docs cover workflows, tradeoffs, and operator guidance
- examples should stay short and shell-realistic

## Verification

The verification stack is intentionally layered:

- unit and command tests lock down deterministic behavior
- snapshot tests keep `docs/REFERENCE.md` aligned with the live command tree
- opt-in live API tests verify the Exa client against real endpoints
- battle tests exercise the compiled binary the way humans and coding agents actually invoke it

See `docs/TESTING.md` for the concrete commands and release-check sequence.
