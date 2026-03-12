#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

BIN="${EXA_CLI_BINARY:-$TMPDIR/exa-cli}"
CONFIG_PATH="$TMPDIR/config.toml"
ORIGINAL_EXA_CLI_CONFIG="${EXA_CLI_CONFIG-}"
ORIGINAL_EXA_API_KEY="${EXA_API_KEY-}"
LIVE_MODE="${EXA_BATTLE_TEST_LIVE:-0}"
HAS_JQ=0

config_path_for_lookup() {
  if [[ -n "${ORIGINAL_EXA_CLI_CONFIG:-}" ]]; then
    printf '%s\n' "$ORIGINAL_EXA_CLI_CONFIG"
    return 0
  fi
  if [[ -n "${XDG_CONFIG_HOME:-}" ]]; then
    printf '%s/exa-cli/config.toml\n' "$XDG_CONFIG_HOME"
    return 0
  fi
  if [[ "$(uname -s)" == "Darwin" ]]; then
    printf '%s/Library/Application Support/exa-cli/config.toml\n' "$HOME"
    return 0
  fi
  printf '%s/.config/exa-cli/config.toml\n' "$HOME"
}

api_key_from_config() {
  local path
  path="$(config_path_for_lookup)"
  [[ -f "$path" ]] || return 1
  sed -nE "s/^[[:space:]]*api_key[[:space:]]*=[[:space:]]*['\"]([^'\"]*)['\"].*/\1/p" "$path" | head -n 1
}

log() {
  printf '%s\n' "$*"
}

pass() {
  printf 'PASS %s\n' "$*"
}

fail() {
  printf 'FAIL %s\n' "$*" >&2
  exit 1
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"
  if [[ "$haystack" != *"$needle"* ]]; then
    fail "$label (missing: $needle)"
  fi
  pass "$label"
}

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"
  if [[ "$haystack" == *"$needle"* ]]; then
    fail "$label (unexpected: $needle)"
  fi
  pass "$label"
}

run_capture() {
  local label="$1"
  shift
  local output
  if ! output="$("$@" 2>&1)"; then
    fail "$label"$'\n'"$output"
  fi
  printf '%s' "$output"
}

run_expect_fail() {
  local label="$1"
  local needle="$2"
  shift 2
  local output
  if output="$("$@" 2>&1)"; then
    fail "$label (expected failure)"
  fi
  if [[ "$output" != *"$needle"* ]]; then
    fail "$label (unexpected stderr: $output)"
  fi
  pass "$label"
}

log "Building exa-cli battle-test binary"
(
  cd "$ROOT"
  go build -o "$BIN" ./cmd/exa-cli
)

if command -v jq >/dev/null 2>&1; then
  HAS_JQ=1
fi

export EXA_CLI_CONFIG="$CONFIG_PATH"
export EXA_CLI_NO_BANNER=1
export NO_COLOR=1
unset EXA_API_KEY

log "Running offline battle-test checks"

home_json="$(run_capture "home JSON output" "$BIN" --format json)"
assert_contains "$home_json" '"command": "home"' "home emits JSON envelope"
if [[ "$HAS_JQ" == "1" ]]; then
  "$BIN" --format json >"$TMPDIR/home.json"
  jq -e '.meta.command == "home"' "$TMPDIR/home.json" >/dev/null || fail "home JSON survives redirect and jq parsing"
  jq -e '(.data.workflows | length) > 0' "$TMPDIR/home.json" >/dev/null || fail "home JSON carries workflows"
  pass "home JSON survives redirect and jq parsing"
fi

version_json="$(run_capture "version JSON output" "$BIN" version --format json)"
assert_contains "$version_json" '"command": "version"' "version honors explicit JSON format"
if [[ "$HAS_JQ" == "1" ]]; then
  version_value="$("$BIN" version --format json | jq -r '.data.version')"
  [[ -n "$version_value" && "$version_value" != "null" ]] || fail "version JSON survives pipe and jq parsing"
  pass "version JSON survives pipe and jq parsing"
fi

run_expect_fail "version rejects invalid format" 'invalid format "yaml"' "$BIN" version --format yaml

printf 'test-key\n' | "$BIN" auth login --stdin >/dev/null
auth_json="$(run_capture "auth status JSON output" "$BIN" auth status --format json)"
assert_contains "$auth_json" '"auth_source": "config"' "auth status reflects stored config key"

doctor_json="$(run_capture "doctor JSON output" "$BIN" doctor --format json)"
assert_contains "$doctor_json" '"api_key_configured": true' "doctor reports configured auth"
if [[ "$HAS_JQ" == "1" ]]; then
  "$BIN" doctor --format json | jq -e '.data.api_key_configured == true' >/dev/null || fail "doctor JSON survives pipe and jq parsing"
  pass "doctor JSON survives pipe and jq parsing"
fi

mcp_snippet="$(run_capture "MCP codex snippet" "$BIN" mcp print codex)"
assert_contains "$mcp_snippet" 'codex mcp add exa --url' "mcp print emits codex snippet"
assert_not_contains "$mcp_snippet" 'exaApiKey=' "mcp print stays keyless by default"

mcp_doctor_json="$(run_capture "MCP doctor JSON output" "$BIN" mcp doctor --format json)"
assert_contains "$mcp_doctor_json" '"embed_key_default": false' "mcp doctor reports keyless default"

config_markdown="$(run_capture "config show markdown output" "$BIN" config show --format markdown)"
assert_contains "$config_markdown" '## Config Show' "config show renders markdown heading"
assert_contains "$config_markdown" '`cache_path`' "config show exposes readable config keys"

printf '{"query":' >"$TMPDIR/bad.json"
run_expect_fail "raw request rejects malformed JSON input" "decode" "$BIN" raw request --method POST --path /search --input "$TMPDIR/bad.json"

"$BIN" gen docs >"$TMPDIR/reference.md"
if ! cmp -s "$TMPDIR/reference.md" "$ROOT/docs/REFERENCE.md"; then
  fail "generated reference docs do not match docs/REFERENCE.md"
fi
pass "generated reference docs match checked-in snapshot"

if [[ "$LIVE_MODE" != "1" ]]; then
  log "Skipping live Exa API battle tests. Set EXA_BATTLE_TEST_LIVE=1 to enable them."
  exit 0
fi

ACTIVE_EXA_API_KEY="${ORIGINAL_EXA_API_KEY:-$(api_key_from_config || true)}"

if [[ -z "$ACTIVE_EXA_API_KEY" ]]; then
  fail "EXA_BATTLE_TEST_LIVE=1 requires EXA_API_KEY or a stored key in exa-cli config.toml"
fi

export EXA_API_KEY="$ACTIVE_EXA_API_KEY"

log "Running live Exa API battle-test checks"

find_json="$(run_capture "live find JSON output" "$BIN" find "Exa search API docs" --num-results 1 --format json --no-cache)"
assert_contains "$find_json" '"command": "find"' "live find returns JSON envelope"
assert_contains "$find_json" '"results"' "live find returns results"
if [[ "$HAS_JQ" == "1" ]]; then
  "$BIN" find "Exa search API docs" --num-results 1 --format json --no-cache >"$TMPDIR/live-find.json"
  jq -e '.meta.command == "find"' "$TMPDIR/live-find.json" >/dev/null || fail "live find JSON parses with jq"
  jq -e '(.data.results | length) > 0' "$TMPDIR/live-find.json" >/dev/null || fail "live find returns at least one result"
  pass "live find JSON parses with jq"
fi

ask_json="$(run_capture "live ask JSON output" "$BIN" ask "How does Exa answer differ from search?" --format json --no-cache)"
assert_contains "$ask_json" '"command": "ask"' "live ask returns JSON envelope"
if [[ "$HAS_JQ" == "1" ]]; then
  "$BIN" ask "How does Exa answer differ from search?" --format json --no-cache >"$TMPDIR/live-ask.json"
  jq -e '.meta.command == "ask"' "$TMPDIR/live-ask.json" >/dev/null || fail "live ask JSON parses with jq"
  jq -e '((.data.answer // "") | length) > 0' "$TMPDIR/live-ask.json" >/dev/null || fail "live ask returns answer text"
  pass "live ask JSON parses with jq"
fi

code_json="$(run_capture "live code JSON output" "$BIN" code "Go Cobra command tree patterns" --format json --no-cache)"
assert_contains "$code_json" '"command": "code"' "live code returns JSON envelope"
if [[ "$HAS_JQ" == "1" ]]; then
  "$BIN" code "Go Cobra command tree patterns" --format json --no-cache >"$TMPDIR/live-code.json"
  jq -e '.meta.command == "code"' "$TMPDIR/live-code.json" >/dev/null || fail "live code JSON parses with jq"
  jq -e '(((.data.context // .data.response // "") | tostring) | length) > 0' "$TMPDIR/live-code.json" >/dev/null || fail "live code returns context text"
  pass "live code JSON parses with jq"
fi

read_json="$(run_capture "live read JSON output" "$BIN" read https://exa.ai/docs/reference/search --summary --format json --no-cache)"
assert_contains "$read_json" '"command": "read"' "live read returns JSON envelope"
if [[ "$HAS_JQ" == "1" ]]; then
  "$BIN" read https://exa.ai/docs/reference/search --summary --format json --no-cache >"$TMPDIR/live-read.json"
  jq -e '.meta.command == "read"' "$TMPDIR/live-read.json" >/dev/null || fail "live read JSON parses with jq"
  jq -e '(.data.results | length) > 0' "$TMPDIR/live-read.json" >/dev/null || fail "live read returns at least one contents result"
  pass "live read JSON parses with jq"
fi

research_json="$(run_capture "live research list JSON output" "$BIN" research list --limit 1 --format json --no-cache)"
assert_contains "$research_json" '"command": "research list"' "live research list returns JSON envelope"
if [[ "$HAS_JQ" == "1" ]]; then
  "$BIN" research list --limit 1 --format json --no-cache >"$TMPDIR/live-research.json"
  jq -e '.meta.command == "research list"' "$TMPDIR/live-research.json" >/dev/null || fail "live research list JSON parses with jq"
  pass "live research list JSON parses with jq"
fi

pass "live Exa API battle-test matrix"
