# Simultaneous Launch Button (slb)

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Build Status](https://github.com/Dicklesworthstone/slb/workflows/CI/badge.svg)](https://github.com/Dicklesworthstone/slb/actions)

A cross-platform CLI that implements a **two-person rule** for running potentially destructive commands from AI coding agents.

When an agent wants to run something risky (e.g., `rm -rf`, `git push --force`, `kubectl delete`, `DROP TABLE`), `slb` requires peer review and explicit approval before execution.

## Why This Exists

Coding agents can get tunnel vision, hallucinate, or misunderstand context. A second reviewer (ideally with a different model/tooling) catches mistakes before they become irreversible.

`slb` is built for **multi-agent workflows** where many agent terminals run in parallel and a single bad command could destroy work, data, or infrastructure.

## Key Features

- **Risk-Based Classification**: Commands are automatically classified by risk level
- **Client-Side Execution**: Commands run in YOUR shell environment (inheriting AWS credentials, kubeconfig, virtualenvs, etc.)
- **Command Hash Binding**: Approvals bind to the exact command via SHA-256 hash
- **SQLite Source of Truth**: Project state lives in `.slb/state.db`
- **Agent Mail Integration**: Notify reviewers and track audit trails via MCP Agent Mail
- **TUI Dashboard**: Beautiful terminal UI for human reviewers

## Risk Tiers

| Tier | Approvals | Auto-approve | Examples |
|------|-----------|--------------|----------|
| **CRITICAL** | 2+ | Never | `rm -rf /`, `DROP DATABASE`, `terraform destroy`, `git push --force` |
| **DANGEROUS** | 1 | Never | `rm -rf ./build`, `git reset --hard`, `kubectl delete`, `DROP TABLE` |
| **CAUTION** | 0 | After 30s | `rm file.txt`, `git branch -d`, `npm uninstall` |
| **SAFE** | 0 | Immediately | `rm *.log`, `git stash`, `kubectl delete pod` |

## Quick Start

### Installation

```bash
# One-liner install
curl -fsSL https://raw.githubusercontent.com/Dicklesworthstone/slb/main/scripts/install.sh | bash

# Or with go install
go install github.com/Dicklesworthstone/slb/cmd/slb@latest

# Or build from source
git clone https://github.com/Dicklesworthstone/slb.git
cd slb && make build
```

### Initialize a Project

```bash
cd /path/to/your/project
slb init
```

This creates a `.slb/` directory with:
- `state.db` - SQLite database for requests, reviews, and sessions
- `config.toml` - Project-specific configuration
- `pending/` - JSON files for pending requests (for watching/interop)

### Basic Workflow

```bash
# 1. Start a session (as an AI agent)
slb session start --agent "GreenLake" --program "claude-code" --model "opus"
# Returns: session_id and session_key

# 2. Run a dangerous command (blocks until approved)
slb run "rm -rf ./build" --reason "Clean build artifacts before fresh compile" --session-id <id>

# 3. Another agent reviews and approves
slb pending                    # See what's waiting for review
slb review <request-id>        # View full details
slb approve <request-id> --session-id <reviewer-id> --comment "Looks safe"

# 4. Original command executes automatically after approval
```

## Commands Reference

### Session Management

```bash
slb session start --agent <name> --program <prog> --model <model>
slb session end --session-id <id>
slb session resume --agent <name>              # Resume after crash
slb session list                               # Show active sessions
slb session heartbeat --session-id <id>        # Keep session alive
```

### Request & Run

```bash
# Primary command (atomic: check, request, wait, execute)
slb run "<command>" --reason "..." [--session-id <id>]

# Plumbing commands
slb request "<command>" --reason "..."         # Create request only
slb status <request-id> [--wait]               # Check status
slb pending [--all-projects]                   # List pending requests
slb cancel <request-id>                        # Cancel own request
```

### Review & Approve

```bash
slb review <request-id>                        # Show full details
slb approve <request-id> --session-id <id>     # Approve request
slb reject <request-id> --session-id <id> --reason "..."
```

### Execution

```bash
slb execute <request-id>                       # Execute approved request
slb emergency-execute "<cmd>" --reason "..."   # Human override (logged)
slb rollback <request-id>                      # Rollback if captured
```

### Pattern Management

```bash
slb patterns list [--tier critical|dangerous|caution|safe]
slb patterns test "<command>"                  # Check what tier a command would be
slb patterns add --tier dangerous "<pattern>"  # Agents can add patterns
```

### Daemon & TUI

```bash
slb daemon start [--foreground]                # Start background daemon
slb daemon stop                                # Stop daemon
slb daemon status                              # Check daemon status
slb tui                                        # Launch interactive TUI
slb watch --session-id <id> --json             # Stream events for agents
```

## Configuration

Configuration is hierarchical (lowest to highest priority):
1. Built-in defaults
2. User config (`~/.slb/config.toml`)
3. Project config (`.slb/config.toml`)
4. Environment variables (`SLB_*`)
5. Command-line flags

### Example Configuration

```toml
[general]
min_approvals = 2
request_timeout = 1800              # 30 minutes
approval_ttl_minutes = 30
timeout_action = "escalate"         # or "auto_reject", "auto_approve_warn"

[rate_limits]
max_pending_per_session = 5
max_requests_per_minute = 10

[notifications]
desktop_enabled = true
desktop_delay_seconds = 60

[daemon]
tcp_addr = ""                       # For Docker/remote agents
tcp_require_auth = true
```

## Default Patterns

### CRITICAL (2+ approvals)

| Pattern | Description |
|---------|-------------|
| `rm -rf /...` | Recursive delete on system paths |
| `DROP DATABASE/SCHEMA` | SQL database destruction |
| `TRUNCATE TABLE` | SQL data destruction |
| `terraform destroy` | Infrastructure destruction |
| `kubectl delete node/namespace/pv/pvc` | Kubernetes critical resources |
| `git push --force` | Force push (not with-lease) |
| `aws terminate-instances` | Cloud resource destruction |
| `dd ... of=/dev/` | Direct disk writes |

### DANGEROUS (1 approval)

| Pattern | Description |
|---------|-------------|
| `rm -rf` | Recursive force delete |
| `git reset --hard` | Discard all changes |
| `git clean -fd` | Remove untracked files |
| `kubectl delete` | Delete Kubernetes resources |
| `terraform destroy -target` | Targeted destroy |
| `DROP TABLE` | SQL table destruction |
| `chmod -R`, `chown -R` | Recursive permission changes |

### CAUTION (auto-approved after 30s)

| Pattern | Description |
|---------|-------------|
| `rm <file>` | Single file deletion |
| `git stash drop` | Discard stashed changes |
| `git branch -d` | Delete local branch |
| `npm/pip uninstall` | Package removal |

### SAFE (skip review)

| Pattern | Description |
|---------|-------------|
| `rm *.log`, `rm *.tmp`, `rm *.bak` | Temporary file cleanup |
| `git stash` | Stash changes (not drop) |
| `kubectl delete pod` | Pod deletion (pods are ephemeral) |
| `npm cache clean` | Cache cleanup |

## IDE Integration

### Claude Code Hooks

Add to your `AGENTS.md`:

```markdown
## SLB Integration

Before running any destructive command, use slb:

\`\`\`bash
# Instead of running directly:
rm -rf ./build

# Use slb:
slb run "rm -rf ./build" --reason "Clean build before fresh compile"
\`\`\`

All DANGEROUS and CRITICAL commands must go through slb review.
```

Generate Claude Code hooks:

```bash
slb integrations claude-hooks > ~/.claude/hooks.json
```

### Cursor Rules

Generate Cursor rules:

```bash
slb integrations cursor-rules > .cursorrules
```

## Shell Completions

```bash
# zsh (~/.zshrc)
eval "$(slb completion zsh)"

# bash (~/.bashrc)
eval "$(slb completion bash)"

# fish (~/.config/fish/config.fish)
slb completion fish | source
```

## Architecture

```
.slb/
├── state.db          # SQLite database (source of truth)
├── config.toml       # Project configuration
├── pending/          # JSON snapshots for watching
│   └── req-<uuid>.json
├── sessions/         # Session files
└── logs/             # Execution logs
    └── req-<uuid>.log
```

**Key Design Decision**: Client-side execution. The daemon is a NOTARY (verifies approvals) not an executor. Commands execute in the calling process's shell environment to inherit:
- AWS_PROFILE, AWS_ACCESS_KEY_ID
- KUBECONFIG
- Activated virtualenvs
- SSH_AUTH_SOCK
- Database connection strings

## Troubleshooting

### "Daemon not running" warning

This is expected - slb works without the daemon (file-based polling). Start the daemon for real-time updates:

```bash
slb daemon start
```

### "Active session already exists"

Resume your existing session instead of starting a new one:

```bash
slb session resume --agent "YourAgent" --create-if-missing
```

### Approval expired

Approvals have a TTL (30min default, 10min for CRITICAL). Re-request if expired:

```bash
slb run "<command>" --reason "..."  # Creates new request
```

### Command hash mismatch

The command was modified after approval. This is a security feature - re-request approval for the modified command.

## Safety Note

`slb` adds friction and peer review for dangerous actions. It does NOT replace:
- Least-privilege credentials
- Environment safeguards
- Proper access controls
- Backup strategies

Use slb as **defense in depth**, not your only protection.

## Claude Code Hook Integration

For seamless integration with Claude Code, `slb` provides a PreToolUse hook that intercepts Bash commands before execution.

### Quick Setup

```bash
# Install hook (generates script and updates Claude Code settings)
slb hook install

# Check installation status
slb hook status

# Test classification without executing
slb hook test "rm -rf ./build"
```

### How It Works

1. **Hook Script**: A Python script at `~/.slb/hooks/slb_guard.py` intercepts Bash tool calls
2. **Pattern Matching**: Commands are classified using embedded patterns (same as the daemon)
3. **Daemon Communication**: For approval checks, the hook connects to the SLB daemon via Unix socket
4. **Fail-Closed**: If SLB is unavailable, dangerous commands are blocked by default

### Hook Commands

```bash
slb hook generate                 # Generate hook script only
slb hook install [--global]       # Install to Claude Code settings
slb hook uninstall                # Remove hook from settings
slb hook status                   # Show installation status
slb hook test "<command>"         # Test command classification
```

The hook returns one of three actions to Claude Code:
- `allow` - Command proceeds without intervention
- `ask` - User is prompted (CAUTION tier)
- `block` - Command is blocked with message to use `slb request`

## Pattern Matching Engine

The pattern matching engine is the core of `slb`'s command classification system.

### Classification Algorithm

1. **Normalization**: Commands are parsed using shell-aware tokenization
   - Strips wrapper prefixes: `sudo`, `doas`, `env`, `time`, `nohup`, etc.
   - Extracts inner commands from `bash -c 'command'` patterns
   - Resolves paths: `./foo` → `/absolute/path/foo`

2. **Compound Command Handling**: Commands with `;`, `&&`, `||`, `|` are split and each segment is classified independently. The **highest risk segment determines the overall tier**.
   ```
   echo "done" && rm -rf /etc    →  CRITICAL (rm -rf /etc wins)
   ls && git status              →  SAFE (no dangerous patterns)
   ```

3. **Shell-Aware Splitting**: Separators inside quotes are preserved:
   ```
   psql -c "DELETE FROM users; DROP TABLE x;"  →  Single segment (SQL)
   echo "foo" && rm -rf /tmp                   →  Two segments
   ```

4. **Pattern Precedence**: Patterns are checked in order: SAFE → CRITICAL → DANGEROUS → CAUTION
   - First match wins within each tier
   - SAFE patterns skip review entirely

5. **Fail-Safe Parse Handling**: If command parsing fails (unbalanced quotes, complex escapes), the tier is **upgraded by one level**:
   - SAFE → CAUTION
   - CAUTION → DANGEROUS
   - DANGEROUS → CRITICAL

### Fallback Detection

For commands that wrap SQL (e.g., `psql -c "..."`, `mysql -e "..."`), pattern matching may not catch embedded statements. The engine includes fallback detection:

```
DELETE FROM ... (no WHERE clause)  →  CRITICAL
DELETE FROM ... WHERE ...          →  DANGEROUS
```

### Runtime Pattern Management

Agents can add patterns at runtime:

```bash
slb patterns add --tier dangerous "docker system prune"
slb patterns list --tier critical
slb patterns test "kubectl delete deployment nginx"
```

Pattern changes are persisted to SQLite and take effect immediately.

## Request Lifecycle

Requests follow a well-defined state machine with strict transition rules.

### State Diagram

```
                    ┌─────────────┐
                    │   PENDING   │
                    └──────┬──────┘
           ┌───────────────┼───────────────┐───────────────┐
           ▼               ▼               ▼               ▼
     ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
     │ APPROVED │    │ REJECTED │    │ CANCELLED│    │ TIMEOUT  │
     └────┬─────┘    └──────────┘    └──────────┘    └────┬─────┘
          │              (terminal)      (terminal)       │
          ▼                                               ▼
     ┌──────────┐                                   ┌──────────┐
     │EXECUTING │                                   │ESCALATED │
     └────┬─────┘                                   └────┬─────┘
          │                                              │
   ┌──────┴──────┬──────────┐                 ┌─────────┴─────────┐
   ▼             ▼          ▼                 ▼                   ▼
┌────────┐  ┌─────────┐  ┌────────┐      ┌──────────┐       ┌──────────┐
│EXECUTED│  │EXEC_FAIL│  │TIMED_OUT│     │ APPROVED │       │ REJECTED │
└────────┘  └─────────┘  └────────┘      └──────────┘       └──────────┘
(terminal)   (terminal)   (terminal)
```

### Terminal States

Once a request reaches a terminal state, no further transitions are allowed:
- **EXECUTED**: Command completed successfully
- **EXECUTION_FAILED**: Command returned non-zero exit code
- **TIMED_OUT**: Command exceeded execution timeout
- **CANCELLED**: Request was cancelled by the requester
- **REJECTED**: Request was rejected by a reviewer

### Approval TTL

Approvals have a time-to-live to prevent stale approvals:
- **Standard requests**: 30 minutes (configurable)
- **CRITICAL requests**: 10 minutes (stricter by default)

If an approval expires before execution, the request must be re-approved.

## Execution Verification

Before any command executes, five security gates must pass:

### Gate 1: Status Check
Request must be in APPROVED state.

### Gate 2: Approval Expiry
Approval TTL must not have elapsed.

### Gate 3: Command Hash
SHA-256 hash of the command must match. This ensures the exact approved command is executed—no modifications after approval.

### Gate 4: Tier Consistency
Risk tier must still match (patterns may have changed since approval).

### Gate 5: First-Executor-Wins
Only one executor can claim the request. Atomic database transition prevents race conditions when multiple agents try to execute.

## Dry Run & Rollback

### Dry Run Pre-flight

For supported commands, `slb` can run a dry-run variant before the real execution:

```bash
# Supported dry-run variants:
terraform plan          # instead of terraform apply
kubectl diff            # instead of kubectl apply
git diff                # show what would change
```

Enable in config:
```toml
[general]
enable_dry_run = true
```

### Rollback State Capture

Before executing, `slb` can capture state for potential rollback:

```toml
[general]
enable_rollback_capture = true
max_rollback_size_mb = 100
```

Captured state includes:
- **Filesystem**: Tar archive of affected paths
- **Git**: HEAD commit, branch, dirty state, untracked files
- **Kubernetes**: YAML manifests of affected resources

Rollback:
```bash
slb rollback <request-id>           # Restore captured state
slb rollback <request-id> --force   # Force overwrite
```

## Daemon Architecture

The daemon provides real-time notifications and execution verification.

### IPC Communication

Primary communication uses Unix domain sockets:
```
/tmp/slb-<hash>.sock
```

The socket path includes a hash derived from the project path, allowing multiple project daemons to coexist.

### JSON-RPC Protocol

All daemon communication uses JSON-RPC 2.0:

```json
{"jsonrpc": "2.0", "method": "hook_query", "params": {"command": "rm -rf /"}, "id": 1}
```

Available methods:
- `hook_query` - Classify command and check approvals
- `hook_health` - Health check with pattern hash
- `verify_execution` - Check execution gates
- `subscribe` - Subscribe to request events

### TCP Mode (Docker/Remote)

For agents in containers or remote machines:

```toml
[daemon]
tcp_addr = "0.0.0.0:9876"
tcp_require_auth = true
tcp_allowed_ips = ["192.168.1.0/24"]
```

### Timeout Handling

When a request's approval window expires:

| Action | Behavior |
|--------|----------|
| `escalate` | Transition to ESCALATED, notify humans (default) |
| `auto_reject` | Automatically reject the request |
| `auto_approve_warn` | Auto-approve CAUTION tier with warning notification |

```toml
[general]
timeout_action = "escalate"
```

### Desktop Notifications

Native notifications on macOS (AppleScript), Linux (notify-send), and Windows (PowerShell):

```toml
[notifications]
desktop_enabled = true
desktop_delay_seconds = 60    # Wait before first notification
```

## Advanced Configuration

### Cross-Project Reviews

Allow reviewers from other projects:

```toml
[general]
cross_project_reviews = true
review_pool = ["agent-a", "agent-b", "human-reviewer"]
```

### Trusted Self-Approval

Designated agents can self-approve after a delay:

```toml
[agents]
trusted_self_approve = ["senior-agent", "lead-developer"]
trusted_self_approve_delay_seconds = 300    # 5 minute delay
```

### Conflict Resolution

When approvals and rejections conflict:

```toml
[general]
conflict_resolution = "any_rejection_blocks"  # Default
# Options: any_rejection_blocks | first_wins | human_breaks_tie
```

### Different Model Requirement

Require reviewers to use a different AI model:

```toml
[general]
require_different_model = true
different_model_timeout = 300    # Escalate to human after 5 min
```

### Rate Limiting

Prevent request floods:

```toml
[rate_limits]
max_pending_per_session = 5      # Max concurrent pending requests
max_requests_per_minute = 10     # Rate limit per session
rate_limit_action = "reject"     # reject | queue | warn
```

### Dynamic Quorum

Scale approval requirements based on active reviewers:

```toml
[patterns.critical]
dynamic_quorum = true
dynamic_quorum_floor = 2    # Minimum approvals even with few reviewers
```

### Webhook Notifications

Send events to external systems:

```toml
[notifications]
webhook_url = "https://slack.com/webhook/..."
```

Payload includes request details, classification, and event type.

## Security Design Principles

### Defense in Depth

`slb` implements multiple security layers:

1. **Pattern-based classification** - First line of defense
2. **Peer review requirement** - Human/agent oversight
3. **Command hash binding** - Tamper detection
4. **Approval TTL** - Prevent stale approvals
5. **Execution verification gates** - Pre-execution checks
6. **Audit logging** - Full traceability

### Cryptographic Guarantees

- **Command binding**: SHA-256 hash computed at request time, verified at execution
- **Review signatures**: HMAC signatures using session keys prevent review forgery
- **Session keys**: Generated per-session, never stored in plaintext

### Fail-Closed Behavior

When components fail:
- Daemon unreachable → Block dangerous commands (hook)
- Parse error → Upgrade tier by one level
- Approval expired → Require new approval
- Hash mismatch → Reject execution

### Audit Trail

Every action is logged to SQLite with:
- Timestamp
- Actor (session ID, agent name)
- Action type
- Request/review details
- Outcome

Query history:
```bash
slb history [--days 7] [--session <id>] [--status executed]
```

## Environment Variables

All config options can be set via environment:

| Variable | Description |
|----------|-------------|
| `SLB_MIN_APPROVALS` | Minimum approval count |
| `SLB_REQUEST_TIMEOUT` | Request timeout in seconds |
| `SLB_TIMEOUT_ACTION` | What to do on timeout |
| `SLB_DESKTOP_NOTIFICATIONS` | Enable desktop notifications |
| `SLB_WEBHOOK_URL` | Webhook notification URL |
| `SLB_DAEMON_TCP_ADDR` | TCP listen address |
| `SLB_TRUSTED_SELF_APPROVE` | Comma-separated trusted agents |

## Planning & Development

- Design doc: `PLAN_TO_MAKE_SLB.md`
- Agent rules: `AGENTS.md`
- Task tracking: `bd ready` (beads)
- Prioritization: `bv --robot-priority`

## Contributions

> *About Contributions:* Please don't take this the wrong way, but I do not accept outside contributions for any of my projects. I simply don't have the mental bandwidth to review anything, and it's my name on the thing, so I'm responsible for any problems it causes; thus, the risk-reward is highly asymmetric from my perspective. I'd also have to worry about other "stakeholders," which seems unwise for tools I mostly make for myself for free. Feel free to submit issues, and even PRs if you want to illustrate a proposed fix, but know I won't merge them directly. Instead, I'll have Claude or Codex review submissions via `gh` and independently decide whether and how to address them. Bug reports in particular are welcome. Sorry if this offends, but I want to avoid wasted time and hurt feelings. I understand this isn't in sync with the prevailing open-source ethos that seeks community contributions, but it's the only way I can move at this velocity and keep my sanity.

## License

MIT License - See [LICENSE](LICENSE) for details.
