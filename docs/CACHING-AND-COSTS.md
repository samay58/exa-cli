# Caching And Costs

## Cache Strategy

`exa-cli` keeps a light SQLite cache to reduce repeated latency and repeated spend during iterative terminal work.

What is cached:

- search responses
- answer responses
- code-context responses
- contents responses
- tracked research payloads

What is not hidden:

- cache location
- cache size
- prune behavior
- whether a response was served from cache

Use:

```bash
exa-cli cache stats
exa-cli cache prune
exa-cli cache clear
exa-cli find "query" --no-cache
```

## Cost Posture

The CLI is designed around Exa’s current API shape:

- `find` defaults to Exa `auto`
- `fast` and `instant` are explicit low-latency modes via `find --profile`
- `deep` and `reasoned` are explicit opt-in modes
- `research run` is separate because it is a different cost and latency class from search

Operational guidance:

- start with `find`
- tune retrieval depth with `exa-cli find --profile instant|fast|balanced|deep|reasoned`
- escalate to `ask` only when synthesis is the actual goal
- use `code` for implementation tasks instead of forcing a generic search flow
- use `research run` for broad, async jobs, not everyday retrieval

## Freshness

For page retrieval, use `read --max-age-hours` rather than older livecrawl semantics.

Examples:

- `0` forces fresh retrieval
- `-1` prefers cache
- omit the flag for normal Exa-managed behavior
