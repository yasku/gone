# Gone — Professional Development Plan

**Version:** 2.0.0-dev
**Last Updated:** 2026-04-27
**Status:** Active Development

---

## Table of Contents

1. [Overview](#overview)
2. [Phases Overview](#phases-overview)
3. [Phase 1: Research & Architecture](#phase-1-research--architecture)
4. [Phase 2: CLI Integration Layer](#phase-2-cli-integration-layer)
5. [Phase 3: New Tabs Implementation](#phase-3-new-tabs-implementation)
6. [Phase 4: Testing & QA](#phase-4-testing--qa)
7. [Phase 5: Release](#phase-5-release)
8. [Task Tracking](#task-tracking)
9. [Changelog](#changelog)

---

## Overview

This plan outlines the integration of CLI diagnostic tools into `gone`, transforming it from a macOS uninstaller + system monitor into a comprehensive system administration TUI.

### Goals

1. **Enhance Monitor Tab** — Add network stats, disk analysis, process detailed view
2. **Add Network Tab** — Real-time network monitoring using `bmon`, `vnstat`, `mtr`
3. **Add Logs Tab** — Log file analysis using `angle-grinder`, `lnav` patterns
4. **Add Audit Tab** — Security/system auditing using `osquery` subprocess queries
5. **Improve Scanner** — Use `fd`-inspired algorithms for faster file discovery

### Non-Goals

- No direct library dependencies on external CLI tools (subprocess + JSON only)
- No macOS App Store integration
- No network-based remote management

---

## Phases Overview

| Phase | Name | Duration | Status |
|-------|------|----------|--------|
| 1 | Research & Architecture | 1 week | `in_progress` |
| 2 | CLI Integration Layer | 1 week | `pending` |
| 3 | New Tabs Implementation | 2 weeks | `pending` |
| 4 | Testing & QA | 1 week | `pending` |
| 5 | Release | 1 week | `pending` |

---

## Phase 1: Research & Architecture

**Duration:** 1 week
**Goal:** Define technical architecture for CLI tool integration

### 1.1 Subprocess Management Layer

Create `internal/cli/` package for managing external tool execution:

```go
// internal/cli/runner.go
type Runner struct {
    timeout time.Duration
    env     []string
}

func (r *Runner) ExecJSON(cmd string, args []string, output interface{}) error
func (r *Runner) ExecStream(cmd string, args []string, handler func(line []byte)) error
```

### 1.2 Tool Discovery & Validation

- Detect if tools are installed via `which`
- Fallback to internal Go implementations when tools unavailable
- Version checking for compatibility

### 1.3 Architecture Decision Records (ADRs)

| ADR-001 | Use subprocess for external tools, not library bindings |
|---------|------------------------------------------------------|
| ADR-002 | JSON as primary interchange format between tools |
| ADR-003 | Graceful degradation when tools unavailable |
| ADR-004 | All external calls timeout after 30s |

### Tasks

- [ ] **T-001**: Create `internal/cli/` package structure
- [ ] **T-002**: Implement `Runner` with timeout and env support
- [ ] **T-003**: Add JSON parsing helpers
- [ ] **T-004**: Create tool discovery utilities
- [ ] **T-005**: Write ADR documents
- [ ] **T-006**: Update DESIGN.md with new architecture

---

## Phase 2: CLI Integration Layer

**Duration:** 1 week
**Goal:** Working integration with `fd`, `osquery`, and `glances`

### 2.1 File Scanner Enhancement (`fd` inspired)

| Feature | Description | Priority |
|---------|-------------|----------|
| Parallel traversal | Use worker pool for dirs | high |
| Pattern matching | Glob and regex support | high |
| JSON output | Parse `fd --json` for structured results | high |
| Skip patterns | `.gitignore`-style ignore files | medium |

### 2.2 OS Query Integration (`osquery`)

| Query | Purpose |
|-------|---------|
| `SELECT name, path FROM apps` | App inventory |
| `SELECT name, cpu_percent FROM processes` | Process list |
| `SELECT * FROM launchd WHERE run_at_load=1` | Startup items |
| `SELECT path, sha256 FROM file WHERE path LIKE '/Applications/%'` | App hashes |

### 2.3 System Metrics Enhancement (`glances` inspired)

- Add network interface stats
- Add per-core CPU breakdown
- Add temperature sensors (if available)
- Add battery health (laptops)

### Tasks

- [ ] **T-007**: Implement `fd` subprocess wrapper
- [ ] **T-008**: Create `osquery` query executor
- [ ] **T-009**: Add network metrics to sysinfo
- [ ] **T-010**: Implement tool availability detection
- [ ] **T-011**: Create fallback internal implementations

---

## Phase 3: New Tabs Implementation

**Duration:** 2 weeks
**Goal:** Three new tabs following existing patterns

### 3.1 Network Tab

**File:** `internal/tui/network.go`

```
┌─────────────────────────────────────────────────────────┐
│  ◉ Uninstall    ○ Monitor    ◉ Network    ○ Audit      │
├─────────────────────────────────────────────────────────┤
│  [bmon-style gauges]                                   │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐       │
│  │  eth0   │ │   wlan0 │ │   utun  │ │   apfs  │       │
│  │ RX: ███ │ │ RX: █░░ │ │ RX: ██░ │ │         │       │
│  │ TX: █░░ │ │ TX: █░░ │ │ TX: ███ │ │         │       │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘       │
│                                                         │
│  [/] Filter interfaces  [m] Detailed view  [r] Reset   │
└─────────────────────────────────────────────────────────┘
```

### 3.2 Logs Tab

**File:** `internal/tui/logs.go`

```
┌─────────────────────────────────────────────────────────┐
│  ○ Uninstall    ○ Monitor    ○ Network    ◉ Logs       │
├─────────────────────────────────────────────────────────┤
│  Filter: [fd --json | agrind '* | parse']             │
│  ┌─────────────────────────────────────────────────┐    │
│  │ 2026-04-27 10:23:11  TRASH   /var/folders/...  │    │
│  │ 2026-04-27 10:22:45  SCAN    Found 154 items  │    │
│  │ 2026-04-27 10:22:30  START   gone v1.0.0      │    │
│  └─────────────────────────────────────────────────┘    │
│                                                         │
│  [Enter] View details  [/] Filter  [c] Clear          │
└─────────────────────────────────────────────────────────┘
```

### 3.3 Audit Tab

**File:** `internal/tui/audit.go`

```
┌─────────────────────────────────────────────────────────┐
│  ○ Uninstall    ○ Monitor    ○ Network    ○ Audit     │
├─────────────────────────────────────────────────────────┤
│  Security Audit via osquery                            │
│  ┌─────────────────────────────────────────────────┐    │
│  │ ✓ Startup Items (12)        [view]            │    │
│  │ ✓ Browser Plugins (3)        [view]            │    │
│  │ ✓ Network Connections (45)   [view]            │    │
│  │ ⚠ Suspicious Processes (2)   [view]            │    │
│  │ ⚠ Open Ports (8)            [view]            │    │
│  └─────────────────────────────────────────────────┘    │
│                                                         │
│  [1-5] Run specific check  [r] Refresh  [q] Quit      │
└─────────────────────────────────────────────────────────┘
```

### Tasks

- [ ] **T-012**: Create `internal/tui/network.go`
- [ ] **T-013**: Create `internal/tui/logs.go`
- [ ] **T-014**: Create `internal/tui/audit.go`
- [ ] **T-015**: Update `app.go` tab management (add 2 new tabs)
- [ ] **T-016**: Update styles for new tabs
- [ ] **T-017**: Add keyboard shortcuts for new tabs
- [ ] **T-018**: Update help overlay

---

## Phase 4: Testing & QA

**Duration:** 1 week
**Goal:** All tests pass, manual QA complete

### 4.1 Unit Tests

| Package | Coverage Target |
|---------|---------------|
| `internal/cli` | 80% |
| `internal/tui/network` | 70% |
| `internal/tui/logs` | 70% |
| `internal/tui/audit` | 70% |

### 4.2 Integration Tests

- Mock subprocess responses for offline testing
- Test JSON parsing edge cases
- Test timeout and error handling

### 4.3 Manual QA Checklist

- [ ] Splash screen renders correctly
- [ ] All tabs switch without lag
- [ ] Search in Uninstall tab works
- [ ] Process kill in Monitor tab works
- [ ] Network tab shows real interfaces
- [ ] Logs tab displays operations.log
- [ ] Audit tab queries osquery (if installed)
- [ ] Graceful fallback when tools unavailable

### Tasks

- [ ] **T-019**: Write unit tests for `internal/cli`
- [ ] **T-020**: Write unit tests for new TUI models
- [ ] **T-021**: Create integration test suite
- [ ] **T-022**: Run `go test -race` on all packages
- [ ] **T-023**: Manual QA on macOS 14+
- [ ] **T-024**: Performance profiling

---

## Phase 5: Release

**Duration:** 1 week
**Goal:** v2.0.0 release ready

### 5.1 Version Bump

```
v1.0.0 → v2.0.0
```

### 5.2 Release Checklist

- [ ] Update CHANGELOG.md
- [ ] Update version in code
- [ ] Create GitHub release
- [ ] Update README.md
- [ ] Announce to community

### Tasks

- [ ] **T-025**: Finalize CHANGELOG.md
- [ ] **T-026**: Update version constants
- [ ] **T-027**: Build and test binary
- [ ] **T-028**: Create GitHub release
- [ ] **T-029**: Update README with new features

---

## Task Tracking

All tasks are tracked in `TASKS.md`. Format:

```markdown
| ID | Description | Phase | Status | Priority | Assignee |
|----|-------------|-------|--------|----------|----------|
| T-001 | Create internal/cli/ package | 1 | done | high | mad-max |
```

### Priority Levels

- `critical` — Must complete before next phase
- `high` — Should complete in planned timeframe
- `medium` — Nice to have
- `low` — Future consideration

### Status Values

- `blocked` — Waiting on dependency
- `pending` — Not started
- `in_progress` — Currently working
- `review` — Waiting for code review
- `done` — Completed

---

## Changelog

See `CHANGELOG.md` for version history. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## Appendix A: CLI Tools Reference

| Tool | Purpose | Integration Method |
|------|---------|-------------------|
| `fd` | Fast file finder | Subprocess JSON |
| `osquery` | SQL OS queries | Subprocess CLI |
| `glances` | System monitor | Subprocess JSON/REST |
| `bmon` | Network monitor | Inspiration only |
| `angle-grinder` | Log parser | Subprocess streaming |
| `mtr` | Traceroute | Potential future |
| `nmap` | Port scanner | Potential future |

## Appendix B: Architecture Diagram

```
gone
├── cmd/gone/main.go
└── internal/
    ├── cli/              # NEW: External tool runner
    │   ├── runner.go
    │   ├── fd.go
    │   ├── osquery.go
    │   └── glances.go
    ├── scanner/          # Existing: File discovery
    ├── sysinfo/           # Existing: System metrics
    ├── remover/           # Existing: Trash operations
    └── tui/               # Existing: UI
        ├── app.go         # Updated: 4 tabs
        ├── uninstall.go   # Existing
        ├── monitor.go     # Enhanced
        ├── network.go     # NEW
        ├── logs.go        # NEW
        └── audit.go       # NEW
```

## Appendix C: Dependencies

```go
// New dependencies for v2.0.0
// (subject to review)
```

---

**Document Version:** 1.0
**Next Review:** 2026-05-04
