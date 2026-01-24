#!/bin/bash
set -euo pipefail

log() { echo "[$(date +%H:%M:%S)] $*" >&2; }

log "=========================================="
log "SLB TOON INTEGRATION E2E TEST"
log "=========================================="

# Phase 1: Prerequisites
log ""
log "Phase 1: Prerequisites"
command -v slb || { log "FAIL: slb not found"; exit 1; }
log "  PASS: slb found at $(command -v slb)"

# Check for tru/tr binary (TOON encoder)
# Note: Cargo.toml specifies "tru" but some builds produce "tr"
if command -v tru >/dev/null 2>&1; then
    log "  PASS: tru found at $(command -v tru)"
elif command -v tr >/dev/null 2>&1 && tr --help 2>&1 | grep -qi "toon"; then
    log "  PASS: tr (TOON) found at $(command -v tr)"
elif [[ -x /data/projects/toon_rust/target/release/tru ]]; then
    log "  PASS: tru found at /data/projects/toon_rust/target/release/tru"
elif [[ -x /data/projects/toon_rust/target/release/tr ]]; then
    log "  PASS: tr found at /data/projects/toon_rust/target/release/tr"
else
    log "  WARN: tru/tr not in PATH, will rely on fallback locations"
fi

# Phase 2: Basic functionality
log ""
log "Phase 2: Basic Functionality"

# Test --output flag
json_result=$(slb version --output json 2>/dev/null)
if [[ -n "$json_result" && "$json_result" == "{"* ]]; then
    log "  PASS: JSON output works"
else
    log "  FAIL: JSON output failed"
    exit 1
fi

# Test TOON output
toon_result=$(slb version --output toon 2>/dev/null || true)
if [[ -n "$toon_result" && "$toon_result" != "{"* ]]; then
    log "  PASS: TOON output works (not JSON)"
else
    log "  FAIL: TOON output returned JSON or empty"
    exit 1
fi

# Test -t shorthand
toon_short=$(slb version -t 2>/dev/null || true)
if [[ "$toon_short" == "$toon_result" ]]; then
    log "  PASS: -t shorthand works"
else
    log "  WARN: -t shorthand gives different result"
fi

# Phase 3: Round-trip verification
log ""
log "Phase 3: Round-trip Verification"

json_output=$(slb version --output json 2>/dev/null)
toon_output=$(slb version --output toon 2>/dev/null)

# Decode TOON back to JSON using tru/tr
# Note: Binary may be named "tru" or "tr" depending on build
TRU_CMD=""
if command -v tru >/dev/null 2>&1; then
    TRU_CMD="tru"
elif command -v tr >/dev/null 2>&1 && tr --help 2>&1 | grep -qi "toon"; then
    TRU_CMD="tr"
elif [[ -x /data/projects/toon_rust/target/release/tru ]]; then
    TRU_CMD="/data/projects/toon_rust/target/release/tru"
elif [[ -x /data/projects/toon_rust/target/release/tr ]]; then
    TRU_CMD="/data/projects/toon_rust/target/release/tr"
else
    log "  SKIP: Cannot verify round-trip without tru/tr"
fi

if [[ -n "$TRU_CMD" && -n "$toon_output" ]]; then
    decoded=$(echo "$toon_output" | "$TRU_CMD" -d 2>/dev/null || echo "")
    if [[ -n "$decoded" ]]; then
        orig_sorted=$(echo "$json_output" | jq -S . 2>/dev/null || echo "$json_output")
        decoded_sorted=$(echo "$decoded" | jq -S . 2>/dev/null || echo "$decoded")

        if [[ "$orig_sorted" == "$decoded_sorted" ]]; then
            log "  PASS: Round-trip preserves data"
        else
            log "  FAIL: Round-trip mismatch"
            log "  Original: $orig_sorted"
            log "  Decoded:  $decoded_sorted"
            exit 1
        fi
    else
        log "  WARN: Could not decode TOON output"
    fi
fi

# Phase 4: Token savings
log ""
log "Phase 4: Token Savings Verification"

if [[ -n "$json_output" && -n "$toon_output" ]]; then
    json_chars=$(echo -n "$json_output" | wc -c)
    toon_chars=$(echo -n "$toon_output" | wc -c)
    savings=$(( 100 - (toon_chars * 100 / json_chars) ))
    log "  JSON: $json_chars chars, TOON: $toon_chars chars, Savings: ${savings}%"

    if [[ $savings -ge 20 ]]; then
        log "  PASS: Token savings >= 20%"
    else
        log "  WARN: Token savings < 20% (got ${savings}%)"
    fi
fi

# Phase 5: Help text verification
log ""
log "Phase 5: Help Text Verification"

help_output=$(slb --help 2>&1)
if echo "$help_output" | grep -q "toon"; then
    log "  PASS: --help mentions toon format"
else
    log "  FAIL: --help does not mention toon format"
    exit 1
fi

if echo "$help_output" | grep -q "\-t,"; then
    log "  PASS: --help mentions -t flag"
else
    log "  WARN: --help does not explicitly show -t flag"
fi

log ""
log "=========================================="
log "SLB TOON E2E TESTS COMPLETE"
log "=========================================="
