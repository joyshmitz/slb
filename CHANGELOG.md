# Changelog

All notable changes to [SLB (Simultaneous Launch Button)](https://github.com/Dicklesworthstone/slb) are documented in this file.

SLB is a cross-platform CLI that implements a two-person rule for running potentially destructive commands from AI coding agents. It provides risk-based classification, peer review enforcement, and full audit logging for multi-agent workflows.

> Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/). This project uses [Semantic Versioning](https://semver.org/).

---

## [Unreleased] &mdash; changes on `main` since v0.2.0

Compare: [`v0.2.0...main`](https://github.com/Dicklesworthstone/slb/compare/v0.2.0...main)

### Added

- **TOON format output**: Token-efficient encoding support via `tru`/`tr` binary for structured CLI output ([`d2154d9`](https://github.com/Dicklesworthstone/slb/commit/d2154d93694ddf7ac9688d59c2585b5827e3a7fe))
- **`--stats` flag**: New CLI flag and `SLB_OUTPUT_FORMAT` env var support for output format control ([`91241f6`](https://github.com/Dicklesworthstone/slb/commit/91241f688913f99a3e67e7012e76021868dd007d))
- **Fallback SQL DELETE detection**: Pattern engine now detects dangerous `DELETE FROM` statements inside compound commands ([`c40b097`](https://github.com/Dicklesworthstone/slb/commit/c40b09758b21edd800a3b4ed0082cacf3d08bff9))
- MIT License file with OpenAI/Anthropic Rider ([`b7becfe`](https://github.com/Dicklesworthstone/slb/commit/b7becfe577b20eb89ab60c1c426cdd5238f2b3c9), [`badc986`](https://github.com/Dicklesworthstone/slb/commit/badc9863901f79ddf9a4ee7a03d923646e329173))
- GitHub social preview image ([`28c2bf7`](https://github.com/Dicklesworthstone/slb/commit/28c2bf77316658962fd72e097f11d710977a14e1))
- ACFS checksum dispatch and notification workflows ([`f0fe162`](https://github.com/Dicklesworthstone/slb/commit/f0fe16231c568220c8cc74f17cb1da7301914c8f), [`94a35eb`](https://github.com/Dicklesworthstone/slb/commit/94a35eb8927ffe2b92dfc9d4466d316d388799f6))

### Fixed

- `slb check` now outputs human-readable text instead of raw Go map syntax ([`99cda5b`](https://github.com/Dicklesworthstone/slb/commit/99cda5b660a3e44b3bc1d3ba7fde244d7e76da28))
- Status update failures after command execution are now logged instead of silently swallowed ([`2a5110e`](https://github.com/Dicklesworthstone/slb/commit/2a5110e074b96ab4c043b6d524f8efcb8130cc46))
- Resolved errcheck lint errors for unchecked error returns ([`0981efd`](https://github.com/Dicklesworthstone/slb/commit/0981efdbfbc6fee6d41afbd35be98b81db0363a5))
- TOON output: support both `tru` and `tr` binary names, additional bug fixes ([`62803d9`](https://github.com/Dicklesworthstone/slb/commit/62803d908ade7de88b87e8a22b2aad33fa99c919), [`de41dbb`](https://github.com/Dicklesworthstone/slb/commit/de41dbb6e1be4b5e037900d64fe397cd1cd3bd62))
- Tier flag updated from `-t` to `-T` in pattern tests ([`f56eeb0`](https://github.com/Dicklesworthstone/slb/commit/f56eeb0885492100f355513cbcc33344e1baeba7))
- CI: golangci-lint v2 compatibility, goreleaser v2 config updates, improved test reliability ([`15bc242`](https://github.com/Dicklesworthstone/slb/commit/15bc24259150db03bef8989b1cd54ddcb1906eb6), [`4eec3b6`](https://github.com/Dicklesworthstone/slb/commit/4eec3b61300ca001d69f434b9eeea9621922fd18), [`8aee4ec`](https://github.com/Dicklesworthstone/slb/commit/8aee4ecdb1b96b2a14e97e5571a650d49ed8b0c0), [`a1737b4`](https://github.com/Dicklesworthstone/slb/commit/a1737b4b4a6be377ebc326aad286c47fbc68e4bb))

### Changed

- README: prioritize Homebrew/Scoop installation methods, update license references ([`0b0307c`](https://github.com/Dicklesworthstone/slb/commit/0b0307c980ee4fe13868e5298b902d5da933df67), [`a82e130`](https://github.com/Dicklesworthstone/slb/commit/a82e130131ae01a78e7d9a03c286a38a8dd2128b))
- License changed to MIT with OpenAI/Anthropic Rider ([`badc986`](https://github.com/Dicklesworthstone/slb/commit/badc9863901f79ddf9a4ee7a03d923646e329173))
- TOON output handler simplified and refactored ([`9c741cc`](https://github.com/Dicklesworthstone/slb/commit/9c741cc567b0e56a90ff38bd05f7c91fc9042546))
- golangci-lint configuration refactored for better code quality checks ([`3782619`](https://github.com/Dicklesworthstone/slb/commit/3782619c5531ce41c7ceac572ae1b2d9e99abe3b))
- CI workflows improved with security and reliability enhancements ([`4acadd0`](https://github.com/Dicklesworthstone/slb/commit/4acadd084a7a4df054d3a4635ba2a228a222d239))
- Go module dependencies updated to latest stable versions ([`d66801f`](https://github.com/Dicklesworthstone/slb/commit/d66801fb728df138ea702232018635f8a5ce597a))

---

## [v0.2.0] &mdash; 2026-01-13

**GitHub Release**: [`v0.2.0`](https://github.com/Dicklesworthstone/slb/releases/tag/v0.2.0)
Compare: [`v0.1.0...v0.2.0`](https://github.com/Dicklesworthstone/slb/compare/v0.1.0...v0.2.0)
Tag type: annotated | Published: 2026-01-14

This release introduces Claude Code hook integration, Homebrew/Scoop auto-publishing via GoReleaser, and a substantial batch of state machine and pattern matching fixes discovered during post-v0.1.0 stabilization.

### Added

- **Claude Code hook integration infrastructure**: Full `slb hook` subcommand (`generate`, `install`, `uninstall`, `status`, `test`) with Python guard script at `~/.slb/hooks/slb_guard.py` that intercepts Bash tool calls, classifies risk, and communicates with the SLB daemon via Unix socket ([`39c2f87`](https://github.com/Dicklesworthstone/slb/commit/39c2f87ff26843c2bc72c528cb3614a26efddb3c))
- **Claude Code SKILL.md**: Automatic capability discovery for Claude Code agents ([`2926cda`](https://github.com/Dicklesworthstone/slb/commit/2926cda2a0eee5b8bcceb0efddbfed070f5cae9a))
- **GoReleaser auto-publishing to Homebrew and Scoop**: `brew install dicklesworthstone/tap/slb` and `scoop install dicklesworthstone/slb` now auto-update on release ([`995dd17`](https://github.com/Dicklesworthstone/slb/commit/995dd17c3fbbe84363306b020cbc9288b869e1ef))
- Comprehensive README documentation for all features (request lifecycle, execution verification, pattern engine, TUI dashboard, agent mail, outcome tracking, etc.) ([`f7d32a7`](https://github.com/Dicklesworthstone/slb/commit/f7d32a756fdd37235480441abf5d654bcab7a278), [`786899a`](https://github.com/Dicklesworthstone/slb/commit/786899a6bc492189938375e38bcd763d026576eb))

### Security

- **Compound command quote bypass vulnerability**: Shell-aware splitting now correctly handles quoted separators, preventing commands like `echo ";" && rm -rf /` from being misclassified ([`dffc948`](https://github.com/Dicklesworthstone/slb/commit/dffc94866e3f0f2e04027fedea3de2167599638f))

### Fixed

- **State machine**: Transitions now prevent panics on short IDs; cancel command and `StatusEscalated` transitions corrected; escalated requests can now be reviewed ([`ba4510d`](https://github.com/Dicklesworthstone/slb/commit/ba4510d9d089d5c5c74c72c4e8d02c94a911a842), [`1476f11`](https://github.com/Dicklesworthstone/slb/commit/1476f11d8fb5fb6d146acd3285e5fad09f316356), [`00b213c`](https://github.com/Dicklesworthstone/slb/commit/00b213cfd1fec5d4d017b559d3bc0272f71515d6))
- **Pattern matching**: Compound command tier precedence fixed; `IsSafe` initialization corrected for compound classification; `rm -fr` pattern bug resolved ([`7fd44d9`](https://github.com/Dicklesworthstone/slb/commit/7fd44d9c27bb341cba94f680ad4f51749fcdaaa4), [`036c75a`](https://github.com/Dicklesworthstone/slb/commit/036c75a6ca168dde2d74d2c8d05182c8af4177d6), [`bb9f88a`](https://github.com/Dicklesworthstone/slb/commit/bb9f88a0e2729e59e88ba980c28760c9cc0d79a0))
- **Execution**: `exit_code`/`duration` only set when command result is actually available; edge cases in command truncation and event mapping handled ([`1c733e8`](https://github.com/Dicklesworthstone/slb/commit/1c733e899a5691aaa6754bdc37f5c1596874c48f), [`8d48a9e`](https://github.com/Dicklesworthstone/slb/commit/8d48a9e607526240a830bfc36b6f2c8738804349))
- **Database**: Missing `Execution`/`Rollback` parsing in `scanRequests` ([`bc94544`](https://github.com/Dicklesworthstone/slb/commit/bc9454456117d2f3eb4f4ba81c9f5b64ffb8751c))
- **Hook**: Corrected bounds check in `slb_guard.py` substring matching; fixed Python hook daemon communication; `hookTestCmd` validation and Python fallback caution handling ([`24157ef`](https://github.com/Dicklesworthstone/slb/commit/24157ef14ea75d3b741845c77703d89f5f8acdbc), [`7268863`](https://github.com/Dicklesworthstone/slb/commit/726886309201f60a41aeba225f7834a0c1830aa1), [`571971e`](https://github.com/Dicklesworthstone/slb/commit/571971e324df1a99e4239a7f6d7ef7eb8265ceae))
- macOS test compatibility issues ([`bb9f88a`](https://github.com/Dicklesworthstone/slb/commit/bb9f88a0e2729e59e88ba980c28760c9cc0d79a0))

### Changed

- Removed AI writing patterns from README ([`05b3036`](https://github.com/Dicklesworthstone/slb/commit/05b30366482b4456105c3734b6a6d7a0805ebf88))
- GoReleaser config updated for v2 (`folder` -> `directory`, Homebrew skip toggle) ([`87db783`](https://github.com/Dicklesworthstone/slb/commit/87db7832ac0b1f578de7a59efdb60d8d0b3850e0), [`8764edd`](https://github.com/Dicklesworthstone/slb/commit/8764eddb4b87f239441c5769ccb54589404a89c2))

---

## [v0.1.0] &mdash; 2025-12-24

**GitHub Release**: [`v0.1.0`](https://github.com/Dicklesworthstone/slb/releases/tag/v0.1.0)
Tag type: lightweight | Published: 2025-12-25

The initial public release of SLB. Built from scratch in ~11 days (2025-12-13 to 2025-12-24) with 264 commits, comprehensive test coverage (80%+ CI threshold), and cross-platform binaries via GoReleaser.

### Core System

- **Risk-based command classification** with four tiers: CRITICAL (2+ approvals), DANGEROUS (1 approval), CAUTION (auto-approve after 30s), SAFE (immediate) ([`c4561db`](https://github.com/Dicklesworthstone/slb/commit/c4561db89abebab1f24e7717b2276308e6e67aca))
- **Shell-aware pattern matching engine**: Normalizes wrapper prefixes (`sudo`, `doas`, `env`, `time`, `nohup`), extracts inner commands from `bash -c`, resolves paths, and splits compound commands (`&&`, `||`, `;`, `|`) while preserving quoted separators ([`c4561db`](https://github.com/Dicklesworthstone/slb/commit/c4561db89abebab1f24e7717b2276308e6e67aca))
- **Fail-safe parse handling**: Unparable commands get their tier upgraded one level ([`c4561db`](https://github.com/Dicklesworthstone/slb/commit/c4561db89abebab1f24e7717b2276308e6e67aca))
- **Client-side execution**: Commands run in the calling process's shell environment to inherit AWS credentials, kubeconfig, virtualenvs, SSH agents, etc. ([`71d5808`](https://github.com/Dicklesworthstone/slb/commit/71d58088b945f3175560cb2d61f594cefd98415b))
- **Command hash binding**: SHA-256 hash computed at request time, verified at execution &mdash; any modification after approval is rejected ([`c4561db`](https://github.com/Dicklesworthstone/slb/commit/c4561db89abebab1f24e7717b2276308e6e67aca))

### Storage & Database

- **SQLite database layer** with WAL mode and FTS5 full-text search for request history ([`f5c8e40`](https://github.com/Dicklesworthstone/slb/commit/f5c8e4008128d7fb3f2b93406193e6b790b5fbe9))
- `.slb/` project directory with `state.db`, `config.toml`, `pending/` JSON snapshots, session files, and execution logs ([`feb8fab`](https://github.com/Dicklesworthstone/slb/commit/feb8fabf4d73ef34fa9f95de14c0af9ff0fc476a))
- `ListAllRequests`, runtime pattern changes, and enhanced test coverage for database layer ([`0c7e07d`](https://github.com/Dicklesworthstone/slb/commit/0c7e07d03503dfed779b92fa6b85b73d67828417))

### CLI

- **`slb init`**: Project initialization command ([`feb8fab`](https://github.com/Dicklesworthstone/slb/commit/feb8fabf4d73ef34fa9f95de14c0af9ff0fc476a))
- **`slb run`**: Atomic check-request-wait-execute pipeline ([`a515b09`](https://github.com/Dicklesworthstone/slb/commit/a515b09bf9a0212263550ce0ac7ab5ec8f1cc45e))
- **`slb request`** / **`slb status`** / **`slb pending`** / **`slb cancel`**: Request lifecycle plumbing ([`a515b09`](https://github.com/Dicklesworthstone/slb/commit/a515b09bf9a0212263550ce0ac7ab5ec8f1cc45e))
- **`slb approve`** / **`slb reject`** / **`slb review`**: Peer review commands ([`d4894a7`](https://github.com/Dicklesworthstone/slb/commit/d4894a776526697dc5be3b45193a24829e4930b3))
- **`slb execute`** / **`slb emergency-execute`** / **`slb rollback`**: Execution and rollback commands ([`4f1acc0`](https://github.com/Dicklesworthstone/slb/commit/4f1acc024d6ce9a55b3fc55ec7d56d55746facac))
- **`slb patterns list|test|add`**: Runtime pattern management ([`a515b09`](https://github.com/Dicklesworthstone/slb/commit/a515b09bf9a0212263550ce0ac7ab5ec8f1cc45e))
- **`slb session start|end|resume|list|heartbeat|gc|reset-limits`**: Session lifecycle management ([`a515b09`](https://github.com/Dicklesworthstone/slb/commit/a515b09bf9a0212263550ce0ac7ab5ec8f1cc45e))
- **`slb history`**: Full-text search with tier/status/agent/date filtering ([`a515b09`](https://github.com/Dicklesworthstone/slb/commit/a515b09bf9a0212263550ce0ac7ab5ec8f1cc45e))
- **`slb show`**: Detailed request view with `--with-reviews`, `--with-execution`, `--with-attachments` ([`a515b09`](https://github.com/Dicklesworthstone/slb/commit/a515b09bf9a0212263550ce0ac7ab5ec8f1cc45e))
- **`slb outcome record|list|stats`**: Execution outcome tracking and analytics ([`6c3b272`](https://github.com/Dicklesworthstone/slb/commit/6c3b272ddfbbac468476b62c57f5979bb786766e))
- **`slb watch`**: Real-time NDJSON event streaming for agent workflows, with polling fallback and `--auto-approve-caution` ([`b958a38`](https://github.com/Dicklesworthstone/slb/commit/b958a3855379c055e0805bf2d5bd0fe0d3be078b))
- **`slb daemon start|stop|status`**: Background daemon management ([`cf17daa`](https://github.com/Dicklesworthstone/slb/commit/cf17daa54521184e129a0bf92354fc18c0b5c794))
- **`slb completion bash|zsh|fish`**: Shell completion generation ([`a515b09`](https://github.com/Dicklesworthstone/slb/commit/a515b09bf9a0212263550ce0ac7ab5ec8f1cc45e))
- **`slb integrations claude-hooks|cursor-rules`**: IDE integration generators ([`7fd6a7d`](https://github.com/Dicklesworthstone/slb/commit/7fd6a7d60bd00837273bd7dec5b0da1aad18d0d6))
- **`--target-project` flag** on `approve`/`reject` for cross-project reviews ([`701df4c`](https://github.com/Dicklesworthstone/slb/commit/701df4cd349e803fc119b552bfdc6289bd8737f3))
- Request attachments: `--attach` for files/images, `--attach-cmd` for command output ([`a515b09`](https://github.com/Dicklesworthstone/slb/commit/a515b09bf9a0212263550ce0ac7ab5ec8f1cc45e))
- JSON and YAML output formats (`--output json`, `--output yaml`, `--json`) ([`de1fe83`](https://github.com/Dicklesworthstone/slb/commit/de1fe83325a6ec6f41e7140ab5b55b6c01fa4977))
- Structured exit codes (0-6) for programmatic use ([`a515b09`](https://github.com/Dicklesworthstone/slb/commit/a515b09bf9a0212263550ce0ac7ab5ec8f1cc45e))

### Daemon & IPC

- **Unix socket IPC server** with JSON-RPC 2.0 protocol (`hook_query`, `hook_health`, `verify_execution`, `subscribe`) ([`9fe267f`](https://github.com/Dicklesworthstone/slb/commit/9fe267f2b04e3d00507b57d140276323be492e16))
- **TCP transport mode** for Docker/remote agents with auth and IP whitelisting ([`007a1fd`](https://github.com/Dicklesworthstone/slb/commit/007a1fd7c647b6cf33be52cd455c03b98d870cf2))
- **Webhook notification system** for external alerting integrations (Slack, etc.) ([`4a635db`](https://github.com/Dicklesworthstone/slb/commit/4a635dbcc02dc5b33be95bb21ccdd27ad2111a16))
- **Desktop notifications** via AppleScript (macOS), notify-send (Linux), PowerShell (Windows) ([`007a1fd`](https://github.com/Dicklesworthstone/slb/commit/007a1fd7c647b6cf33be52cd455c03b98d870cf2))
- **File watcher** for `pending/` directory changes ([`007a1fd`](https://github.com/Dicklesworthstone/slb/commit/007a1fd7c647b6cf33be52cd455c03b98d870cf2))
- Timeout handling with configurable actions: `escalate`, `auto_reject`, `auto_approve_warn` ([`007a1fd`](https://github.com/Dicklesworthstone/slb/commit/007a1fd7c647b6cf33be52cd455c03b98d870cf2))

### TUI Dashboard

- **Interactive terminal UI** with three panels: Agents, Pending Requests, Activity feed ([`6eea51f`](https://github.com/Dicklesworthstone/slb/commit/6eea51f9965cd234decd44cc2b9709dcd0015e7c))
- Request detail views with approve/reject actions ([`ac3f30e`](https://github.com/Dicklesworthstone/slb/commit/ac3f30ef9a395432f7006f2ff2a85b6c8e8e6ee0), [`88c25c8`](https://github.com/Dicklesworthstone/slb/commit/88c25c8c5ecacbc3131ac4a77886d43b5f7f824c))
- Multi-view navigation: pattern management, history browser with FTS search ([`fcc6b68`](https://github.com/Dicklesworthstone/slb/commit/fcc6b68e669ce462af64c2faa71507a1b4c1a146), [`8f5248d`](https://github.com/Dicklesworthstone/slb/commit/8f5248d05e99483b60cfb6de59b7b2d19636c7c4))
- Pattern removal review view for human-in-the-loop management ([`6bba671`](https://github.com/Dicklesworthstone/slb/commit/6bba67182031a8e10599ed8be1f4dae0005d3662))
- Component library: StatusBadge, AgentCard, Timeline, icons ([`f2dca58`](https://github.com/Dicklesworthstone/slb/commit/f2dca58525da128a24e44eae1694b0c0e4cfded9))

### Configuration

- **Hierarchical TOML configuration**: built-in defaults < user config (`~/.slb/config.toml`) < project config (`.slb/config.toml`) < env vars (`SLB_*`) < CLI flags ([`58084cb`](https://github.com/Dicklesworthstone/slb/commit/58084cb17b90ad6e7f6a92661eae363be2de0a20))
- Viper config support ([`9d1b783`](https://github.com/Dicklesworthstone/slb/commit/9d1b783d15ca95379ece5d46733793c512ec4687))
- Cross-project reviews, trusted self-approval with delay, conflict resolution policies, different-model requirement ([`58084cb`](https://github.com/Dicklesworthstone/slb/commit/58084cb17b90ad6e7f6a92661eae363be2de0a20))
- Rate limiting with configurable actions (`reject`, `queue`, `warn`) ([`58084cb`](https://github.com/Dicklesworthstone/slb/commit/58084cb17b90ad6e7f6a92661eae363be2de0a20))
- Dynamic quorum scaling based on active reviewers ([`58084cb`](https://github.com/Dicklesworthstone/slb/commit/58084cb17b90ad6e7f6a92661eae363be2de0a20))

### Execution Safety

- **Five security gates** before execution: status check, approval expiry, command hash, tier consistency, first-executor-wins ([`71d5808`](https://github.com/Dicklesworthstone/slb/commit/71d58088b945f3175560cb2d61f594cefd98415b))
- **Dry run pre-flight** for terraform, kubectl, git ([`d1e8bde`](https://github.com/Dicklesworthstone/slb/commit/d1e8bde94d8d77bdafb219954149ef0ff114aad1))
- **Rollback state capture**: filesystem tar archives, git state, Kubernetes manifests ([`d1e8bde`](https://github.com/Dicklesworthstone/slb/commit/d1e8bde94d8d77bdafb219954149ef0ff114aad1))
- **Emergency override** with mandatory reason, hash acknowledgment, and permanent audit record ([`a515b09`](https://github.com/Dicklesworthstone/slb/commit/a515b09bf9a0212263550ce0ac7ab5ec8f1cc45e))

### Security

- **Session key verification** on review submission &mdash; prevents review forgery ([`59e183d`](https://github.com/Dicklesworthstone/slb/commit/59e183daaca07676000a676874e8f63097e0db7e))
- **Strict file permissions**: `.slb/` directory (0700), `state.db` and config files (0600) ([`0e69925`](https://github.com/Dicklesworthstone/slb/commit/0e699254b0dfd6f35e4a37af3539c7930b62c7c3))
- **TOCTOU, injection, and regex vulnerability patches** ([`edb3e30`](https://github.com/Dicklesworthstone/slb/commit/edb3e3083858000c105c30f41d49007d0e333991))
- **Path traversal fix** in command normalization ([`01b2719`](https://github.com/Dicklesworthstone/slb/commit/01b2719bf469cc6775e7d47cac89144a45fd91c1))
- **Optimistic locking** on `UpdateRequestStatus` to fix race conditions ([`c6cda51`](https://github.com/Dicklesworthstone/slb/commit/c6cda514e440dfb1a13f7ab74a192ea09f4f8429))
- **Transactional review submission** for concurrency safety ([`20039ad`](https://github.com/Dicklesworthstone/slb/commit/20039ad37c91a9b8403fa08e42615a7f4adb8ac9))
- `shouldAutoApproveCaution` extracted as pure function (P0 security-critical) ([`621a9ce`](https://github.com/Dicklesworthstone/slb/commit/621a9ce6dbef333cf257445654ca7216121161e7))
- Critical patterns added for disk destruction (`dd of=/dev/`) and system file changes ([`b3dd913`](https://github.com/Dicklesworthstone/slb/commit/b3dd913dc748a2afddb9fdc56a1b3a95f5277202))
- `IPC server` refuses to delete non-socket files ([`662ffee`](https://github.com/Dicklesworthstone/slb/commit/662ffeefcd58a6bec4a532045114f81e3d2f0fe4))

### Other

- **Git history audit trail** and IDE integration packages ([`9921510`](https://github.com/Dicklesworthstone/slb/commit/9921510edf4b24d13a020f779ca5a725e4919e4e))
- **Different-model escalation** with timeout support ([`9d1b783`](https://github.com/Dicklesworthstone/slb/commit/9d1b783d15ca95379ece5d46733793c512ec4687), [`d43682d`](https://github.com/Dicklesworthstone/slb/commit/d43682d58bf99dc23fdc4286ad96d54fe6ea4a86))
- State machine hardening and git hook installation refactor ([`fa3918c`](https://github.com/Dicklesworthstone/slb/commit/fa3918ce8136a7fb895d9d26ef71787cd033e9de))
- Duplicate command hashing logic removed ([`699c318`](https://github.com/Dicklesworthstone/slb/commit/699c318af3b2c391facd8d5cb6c2dc63e5aad0d7))
- Context propagation in CLI run commands ([`b612262`](https://github.com/Dicklesworthstone/slb/commit/b612262120d3bbbc9510775592ca2f6d79f10585))
- Contribution policy added to README ([`50917db`](https://github.com/Dicklesworthstone/slb/commit/50917db47a5a873e5a72d14e8db78b1593ce0c59))

### Build & CI

- **Go module foundation** with Makefile build system ([`c2376d6`](https://github.com/Dicklesworthstone/slb/commit/c2376d60cdbee4e5061d9a7d426ed6862b70e5f3))
- **CI/CD pipeline** with security scanning (gosec, staticcheck) ([`1a82a81`](https://github.com/Dicklesworthstone/slb/commit/1a82a815204d920f832805a91c43355df30c65eb))
- **GoReleaser** producing cross-platform binaries: Linux amd64/arm64 (tar.gz, deb, rpm, apk), macOS amd64/arm64 (tar.gz), Windows amd64 (zip), with SBOM generation and cosign signatures ([`cc17518`](https://github.com/Dicklesworthstone/slb/commit/cc17518fe7d699363f4bcb48670ed4a3bbc71127))
- **Codecov** integration with 80% coverage threshold ([`101ef24`](https://github.com/Dicklesworthstone/slb/commit/101ef245812fff528a32b51b63ba88473c21476e))

### Testing

- Test infrastructure foundation (`internal/testutil`) ([`966e90d`](https://github.com/Dicklesworthstone/slb/commit/966e90ddfb25484e4dcc7c0084ea09c5bb184303))
- Core package: 90.2% coverage ([`012584a`](https://github.com/Dicklesworthstone/slb/commit/012584ab6207e9a1bac6b23c254c3295368e5578))
- CLI package: 87% coverage ([`4cd7c6a`](https://github.com/Dicklesworthstone/slb/commit/4cd7c6a22bba4df91c4b209be79eb04478326e7e))
- Daemon package: 85% coverage ([`b7c65cb`](https://github.com/Dicklesworthstone/slb/commit/b7c65cbf4610314eabe1f08a429f1d1ac0564926))
- TUI package: 97.6% coverage ([`6c45bae`](https://github.com/Dicklesworthstone/slb/commit/6c45bae2cd44231dfc1622cd2da61df0528c621e))
- Watch command: 90% coverage ([`23264c4`](https://github.com/Dicklesworthstone/slb/commit/23264c45faec129ec0acc8091add269a08865d3c))
- E2E tests: multi-agent approval workflow, risk tier classification, session/timeout management, git/filesystem rollback, `slb init` ([`561133f`](https://github.com/Dicklesworthstone/slb/commit/561133f6801b9be4aadd0b9547f2f3fa56972c25), [`f0b9eec`](https://github.com/Dicklesworthstone/slb/commit/f0b9eec18e22cbd609f455281645474081ad9f30), [`1b5d91c`](https://github.com/Dicklesworthstone/slb/commit/1b5d91cf334b1d033e8f1bdefef6f1d0cba4741e), [`87834d1`](https://github.com/Dicklesworthstone/slb/commit/87834d102b812cc029c392b5ff34d1aaf88f00fb), [`feb8fab`](https://github.com/Dicklesworthstone/slb/commit/feb8fabf4d73ef34fa9f95de14c0af9ff0fc476a))
- Fixed flaky test ID generation, IPC race conditions, zombie processes in daemon tests ([`e9473e4`](https://github.com/Dicklesworthstone/slb/commit/e9473e43b80bf8de8fdbfcb3e8c08860a0c42f79), [`cb1b7fa`](https://github.com/Dicklesworthstone/slb/commit/cb1b7fafa5fe6ae6e9d2482d754ba5e480cbb598), [`ba8cae3`](https://github.com/Dicklesworthstone/slb/commit/ba8cae3035db5dfae237ab337e868f5672c61c32))
- `isUniqueConstraintError` fixed to not incorrectly match FOREIGN KEY errors ([`5561d52`](https://github.com/Dicklesworthstone/slb/commit/5561d523c64dce5e511d49f6952811559fe4f505))

---

## Pre-release Development &mdash; 2025-12-13

The project was conceived, designed, and substantially built on 2025-12-13 with rapid iteration through 2025-12-24. The initial planning document (`PLAN_TO_MAKE_SLB.md`) was revised to v2.0.0 during this period, incorporating atomic `slb run`, client-side execution, command hash binding, dynamic quorum, and improved SQL patterns.

Key design documents:
- [`eca7c4a`](https://github.com/Dicklesworthstone/slb/commit/eca7c4a8edf322217f502eb808d3c2bb215a8b8f) &mdash; Initial AGENTS.md and system documentation
- [`3cf711b`](https://github.com/Dicklesworthstone/slb/commit/3cf711b462bd40e545295161f0dfffb908cecfec) &mdash; Initial planning transcript
- [`f205128`](https://github.com/Dicklesworthstone/slb/commit/f2051282cba440d572e57293607183e804cd94bd) &mdash; PLAN_TO_MAKE_SLB.md v2.0.0

---

<!-- link definitions -->
[Unreleased]: https://github.com/Dicklesworthstone/slb/compare/v0.2.0...main
[v0.2.0]: https://github.com/Dicklesworthstone/slb/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/Dicklesworthstone/slb/releases/tag/v0.1.0
