# Agent-Friendliness Report: slb (Simultaneous Launch Button)

**Date**: 2026-01-25
**Bead**: bd-19g
**Analyst**: Claude Code Agent (cc)

## Executive Summary

slb is a well-designed tool for multi-agent safety coordination. It has good robot mode support but needs some improvements for optimal agent usage.

**Overall Score: 7/10** (Good foundation, minor issues)

## Current State

### Strengths

| Area | Rating | Notes |
|------|--------|-------|
| **TOON Support** | ✓ Excellent | `--toon`, `--output toon` flags, graceful fallback |
| **JSON Output** | ✓ Good | `--json` on most commands |
| **Documentation** | ✓ Good | AGENTS.md, README.md, SKILL.md exist |
| **Command Structure** | ✓ Good | Clear subcommand hierarchy |
| **Error Messages** | ✓ Good | Helpful error output |

### Issues Found

| Issue | Severity | Description |
|-------|----------|-------------|
| **Flag conflict panic** | HIGH | `slb patterns list` panics due to `-t` shorthand conflict |
| **Check output format** | MEDIUM | `slb check` outputs Go map syntax, not JSON |
| **Missing --json on check** | LOW | `slb check` doesn't support `--json` flag |

## Detailed Analysis

### 1. Documentation Audit

| Document | Status | Notes |
|----------|--------|-------|
| README.md | ✓ Exists | 34 KB, comprehensive |
| AGENTS.md | ✓ Exists | 19 KB, good coverage |
| SKILL.md | ✓ Exists | 18 KB, Claude Code skill |
| RESEARCH_FINDINGS.md | ✓ Exists | TOON integration documented |

**Recommendation**: Documentation is in good shape.

### 2. Robot Mode Completeness

| Command | --json | --toon | Notes |
|---------|--------|--------|-------|
| `pending` | ✓ | ✓ | Works |
| `history` | ✓ | ✓ | Works |
| `config show` | ✓ | ✓ | Works |
| `daemon status` | ✓ | ✓ | Works |
| `session list` | ✓ | ✓ | Works |
| `patterns list` | ✗ PANIC | ✗ PANIC | **BUG: -t flag conflict** |
| `check` | ✗ | ✗ | Outputs Go map format |
| `show` | ✓ | ✓ | Works |
| `watch` | NDJSON | ✓ | Streaming |

### 3. Bug Report: Flag Conflict

**Command**: `slb patterns list --json`
**Error**: Panic due to `-t` shorthand conflict

```
panic: unable to redefine 't' shorthand in "patterns" flagset:
it's already used for "tier" flag
```

**Root Cause**: Both `--toon` (global) and `--tier` (patterns command) use `-t` shorthand.

**Fix**: Change `--tier` shorthand to `-T` or remove shorthand from global `--toon`.

### 4. CLI Ergonomics

**Good patterns**:
- Clear subcommand hierarchy
- Consistent `--json`/`--toon` flags
- Good help text

**Issues**:
- `slb check` outputs non-standard format
- Missing JSON Schema documentation
- No `--schema` flag for self-describing output

### 5. Agent Workflow Quality

**Two-person rule workflow**:
1. Agent A: `slb request 'rm -rf /tmp/old'` → Creates request
2. Agent B: `slb pending --json` → Sees pending request
3. Agent B: `slb approve <id>` → Approves
4. Agent A: `slb execute <id>` → Runs command

**Workflow is well-designed** for multi-agent coordination.

## Recommendations

### Priority 1: Fix Critical Bug

```go
// In patterns.go, change:
cmd.Flags().StringVarP(&tier, "tier", "t", "", "filter by tier")
// To:
cmd.Flags().StringVarP(&tier, "tier", "T", "", "filter by tier")
```

### Priority 2: Fix `slb check` Output

Current:
```
map[command:rm -rf / is_safe:false matched_pattern:^rm\s+(-[rf]+\s+)+/($|\s) min_approvals:2 needs_approval:true tier:critical]
```

Should be:
```json
{
  "command": "rm -rf /",
  "is_safe": false,
  "tier": "critical",
  "needs_approval": true,
  "min_approvals": 2,
  "matched_pattern": "^rm\\s+(-[rf]+\\s+)+/($|\\s)"
}
```

### Priority 3: Add JSON Schema Support

```bash
slb schema pending    # Emit JSON Schema for pending output
slb schema config     # Emit JSON Schema for config output
```

## Test Commands for Agents

```bash
# Quick health check
slb version
slb daemon status --json

# Check command classification
slb check 'rm -rf /tmp'
slb check 'git status'

# List pending requests
slb pending --toon

# Watch for requests (for reviewing agent)
slb watch --output json
```

## Acceptance Criteria Status

From bd-19g:

- [x] Documentation Audit completed
- [x] Robot Mode Completeness evaluated
- [x] CLI Ergonomics evaluated
- [x] Agent Interaction Patterns documented
- [x] Safety Model evaluated
- [x] Bug found and documented

## Related Beads

- **bd-19g**: This re-underwriting bead
- **bd-3ua**: TOON research (complete)
- **bd-2ti**: TOON integration (complete)

## Files to Modify

1. `cmd/patterns.go` - Fix `-t` flag conflict
2. `cmd/check.go` - Add proper JSON output
3. `internal/output/` - Consider adding JSON Schema support
