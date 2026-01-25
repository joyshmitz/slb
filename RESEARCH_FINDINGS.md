# RESEARCH FINDINGS: slb (Simultaneous Launch Button) - TOON Integration Analysis

**Date**: 2026-01-25
**Bead**: bd-3ua
**Researcher**: Claude Code Agent (cc)

## 1. Project Overview

| Attribute | Value |
|-----------|-------|
| **Language** | Go 1.24.4 |
| **CLI Framework** | Cobra |
| **TUI Framework** | Bubble Tea + Lipgloss |
| **Tier** | 3 (Lower Impact - Smaller outputs) |
| **Directory** | `/dp/slb` |

### Purpose
SLB implements a two-person authorization rule for dangerous commands. Before running commands like `rm -rf`, `git push --force`, or `kubectl delete`, SLB requires approval from another authorized agent or human reviewer.

## 2. TOON Integration Status: COMPLETE

**TOON support is already fully implemented in slb.**

### CLI Flags
```
-o, --output string   output format: text, json, yaml, toon (default "text")
-j, --json            shorthand for --output=json
-t, --toon            shorthand for --output=toon
```

### Supported Formats
| Format | Flag | Description |
|--------|------|-------------|
| `text` | (default) | Human-readable output |
| `json` | `--json` or `-j` | Standard JSON |
| `yaml` | `--output yaml` | YAML format |
| `toon` | `--toon` or `-t` | TOON format (token-optimized) |

## 3. Implementation Details

### Source Files

| File | Purpose |
|------|---------|
| `internal/output/output.go` | Main Writer struct with format switching |
| `internal/output/toon.go` | TOON encoding via tru CLI wrapper |
| `internal/output/toon_test.go` | TOON round-trip tests |
| `internal/output/format.go` | Global output mode (legacy) |

### Format Enum
```go
// File: internal/output/output.go
type Format string

const (
    FormatText Format = "text"
    FormatJSON Format = "json"
    FormatYAML Format = "yaml"
    FormatTOON Format = "toon"
)
```

### TOON Encoding Flow
```go
// File: internal/output/toon.go
func EncodeTOON(data any) (string, error) {
    binPath, err := findTOONBinary()  // Finds "tru" binary
    jsonBytes, err := json.Marshal(data)
    cmd := exec.Command(binPath, "-e")
    cmd.Stdin = bytes.NewReader(jsonBytes)
    // Returns TOON-encoded string
}
```

### Binary Discovery
The `tru` binary is searched in:
1. `TOON_TRU_BIN`, `TOON_BIN`, `TRU_PATH` env vars
2. `$PATH` (with toon_rust verification)
3. `~/.local/bin/tru`, `~/bin/tru`, `/usr/local/bin/tru`
4. Development paths: `/data/projects/toon_rust/target/{release,debug}/tru`

### Graceful Degradation
```go
func (w *Writer) writeTOON(data any) error {
    toonStr, err := EncodeTOON(data)
    if err != nil {
        // Fall back to JSON if TOON encoding fails
        fmt.Fprintf(w.errOut, "warning: TOON encoding failed, falling back to JSON: %v\n", err)
        enc := json.NewEncoder(w.out)
        return enc.Encode(data)
    }
    _, err = w.out.Write([]byte(toonStr))
    return err
}
```

## 4. Output Analysis

### Commands Supporting TOON

| Command | Description |
|---------|-------------|
| `slb pending` | List pending requests |
| `slb history` | Browse request history |
| `slb config show` | Show configuration |
| `slb status` | Request status |
| `slb show` | Request details |
| `slb watch` | Watch for pending requests |

### Token Savings Measurement

| Command | JSON Size | TOON Size | Savings |
|---------|-----------|-----------|---------|
| `config show` | 3,387 bytes | 2,559 bytes | **24%** |
| `pending` (empty) | 2 bytes | 5 bytes | (negligible) |

**Note**: Config is the largest typical output. With real pending requests, expect higher savings (30-40%) due to tabular structure.

## 5. Integration Assessment

### Complexity Rating: **ALREADY COMPLETE**

**Rationale**:
- Full TOON support via `--output toon` / `--toon` flag
- Graceful fallback to JSON on encoding failure
- Proper tru binary discovery with env var overrides
- Round-trip tests in `toon_test.go`

### Key Implementation Quality Points

1. **Binary Validation**: Verifies tru is actually toon_rust (not coreutils `tr`)
2. **Multiple Discovery Paths**: Env vars, PATH, common locations
3. **Error Messages**: Helpful error with install instructions
4. **Fallback Handling**: Graceful degradation with stderr warning

## 6. What's Already Done

- [x] `--output toon` CLI flag
- [x] `--toon` shorthand flag
- [x] TOON encoding via tru wrapper
- [x] Binary discovery logic
- [x] Graceful fallback to JSON
- [x] Error output in TOON format
- [x] Round-trip tests

## 7. Remaining Work (if any)

- [ ] **Optional**: Add `--stats` flag for token comparison
- [ ] **Optional**: Add NDJSON-TOON streaming for `watch` command
- [ ] **Optional**: Environment variable precedence (TOON_DEFAULT_FORMAT)

These are enhancements, not blockers. The core TOON integration is complete.

## 8. Verification Commands

```bash
# Verify TOON works
slb pending --toon
slb config show --toon | head -20

# Compare sizes
slb config show --json | wc -c  # JSON size
slb config show --toon | wc -c  # TOON size

# Verify tru detection
slb config show --toon 2>&1  # Should not show "warning"
```

## 9. Conclusion

**bd-2ti (Integrate TOON into slb) should be marked COMPLETE.**

The slb project already has full TOON integration with:
- CLI flags (`--output toon`, `--toon`)
- Proper tru binary discovery
- Graceful fallback behavior
- Test coverage

No additional implementation work is required.

## 10. Related Beads

- **bd-2ti**: "Integrate TOON into slb" → Should be closed as complete
- **bd-3ua**: This research bead → Complete
