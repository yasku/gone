# Gone — Task Tracking

**Last Updated:** 2026-04-27
**Total Tasks:** 29
**Completed:** 15
**In Progress:** 1
**Pending:** 13

---

## Legend

| Status | Meaning |
|--------|---------|
| `blocked` | Waiting on dependency |
| `pending` | Not started |
| `in_progress` | Currently working |
| `review` | Waiting for code review |
| `done` | Completed |

| Priority | Meaning |
|----------|---------|
| `critical` | Must complete before next phase |
| `high` | Should complete in planned timeframe |
| `medium` | Nice to have |
| `low` | Future consideration |

---

## Phase 1: Research & Architecture

| ID | Description | Status | Priority | Notes |
|----|-------------|--------|----------|-------|
| T-001 | Create `internal/cli/` package structure | done | high | |
| T-002 | Implement `Runner` with timeout and env support | done | high | |
| T-003 | Add JSON parsing helpers | done | high | |
| T-004 | Create tool discovery utilities | done | medium | |
| T-005 | Write ADR documents | pending | low | Architecture Decision Records |
| T-006 | Update DESIGN.md with new architecture | done | high | |

## Phase 2: CLI Integration Layer

| ID | Description | Status | Priority | Notes |
|----|-------------|--------|----------|-------|
| T-007 | Implement `fd` subprocess wrapper | done | high | internal/cli/fd.go |
| T-008 | Create `osquery` query executor | done | high | internal/cli/osquery.go |
| T-009 | Add network metrics to sysinfo | done | high | Using gopsutil IOCounters |
| T-010 | Implement tool availability detection | done | medium | internal/cli/tool.go |
| T-011 | Create fallback internal implementations | done | medium | Graceful degradation |

## Phase 3: New Tabs Implementation

| ID | Description | Status | Priority | Notes |
|----|-------------|--------|----------|-------|
| T-012 | Create `internal/tui/network.go` | done | high | |
| T-013 | Create `internal/tui/logs.go` | done | high | |
| T-014 | Create `internal/tui/audit.go` | done | high | |
| T-015 | Update `app.go` tab management (add 2 new tabs) | done | high | Total: 5 tabs |
| T-016 | Update styles for new tabs | pending | medium | |
| T-017 | Add keyboard shortcuts for new tabs | done | medium | |
| T-018 | Update help overlay | done | low | |

## Phase 4: Testing & QA

| ID | Description | Status | Priority | Notes |
|----|-------------|--------|----------|-------|
| T-019 | Write unit tests for `internal/cli` | done | high | 39 tests passing |
| T-020 | Write unit tests for new TUI models | pending | high | Target 70% coverage |
| T-021 | Create integration test suite | done | medium | Mock subprocess via scripts |
| T-022 | Run `go test -race` on all packages | done | high | All packages pass |
| T-023 | Manual QA on macOS 14+ | pending | high | |
| T-024 | Performance profiling | pending | low | |

## Automation Scripts

| ID | Description | Status | Priority |
|----|-------------|--------|----------|
| S-001 | `scripts/test-all.sh` - Full test suite | done | high |
| S-002 | `scripts/coverage.sh` - Coverage reports | done | medium |
| S-003 | `scripts/task-report.sh` - Task status report | done | medium |
| S-004 | `scripts/release.sh` - Release preparation | done | medium |
| S-005 | `scripts/check-tools.sh` - Tool verification | done | medium |

## Phase 5: Release

| ID | Description | Status | Priority | Notes |
|----|-------------|--------|----------|-------|
| T-025 | Finalize CHANGELOG.md | done | high | |
| T-026 | Update version constants | pending | high | v2.0.0 |
| T-027 | Build and test binary | pending | high | |
| T-028 | Create GitHub release | pending | medium | |
| T-029 | Update README with new features | done | medium | |

---

## Completed Tasks

| ID | Date | Description |
|----|------|-------------|
| T-001 | 2026-04-27 | Create `internal/cli/` package structure |
| T-002 | 2026-04-27 | Implement `Runner` with timeout and env support |
| T-003 | 2026-04-27 | Add JSON parsing helpers |
| T-004 | 2026-04-27 | Create tool discovery utilities |
| T-006 | 2026-04-27 | Update DESIGN.md with new architecture |
| T-007 | 2026-04-27 | Implement `fd` subprocess wrapper |
| T-008 | 2026-04-27 | Create `osquery` query executor |
| T-009 | 2026-04-27 | Add network metrics to sysinfo |
| T-010 | 2026-04-27 | Implement tool availability detection |
| T-011 | 2026-04-27 | Create fallback internal implementations |
| T-012 | 2026-04-27 | Create `internal/tui/network.go` |
| T-013 | 2026-04-27 | Create `internal/tui/logs.go` |
| T-014 | 2026-04-27 | Create `internal/tui/audit.go` |
| T-015 | 2026-04-27 | Update `app.go` tab management (5 tabs) |
| T-017 | 2026-04-27 | Add keyboard shortcuts for new tabs |
| T-018 | 2026-04-27 | Update help overlay |
| T-019 | 2026-04-27 | Write unit tests for `internal/cli` (39 tests) |
| T-021 | 2026-04-27 | Create integration test suite |
| T-022 | 2026-04-27 | Run `go test -race` on all packages |
| T-025 | 2026-04-27 | Finalize CHANGELOG.md |
| T-029 | 2026-04-27 | Update README with new features |

---

## Blocked Tasks

_(none yet)_

---

## Task Dependencies

```
T-001 ─┬─ T-002 ─ T-003 ─ T-004 ✓
       │
       └─ T-005 ─ T-006 ✓

T-007 ✓┬─ T-008 ✓─ T-010 ✓─ T-011 ✓
       │
       └─ T-009 ✓─ T-010 ✓

T-012 ✓┬─ T-013 ✓─┬─ T-015 ✓─┬─ T-016 ─ T-017 ✓─ T-018 ✓
       │          │
       └─ T-014 ✓─┘

T-019 ✓┬─ T-020 ─ T-021 ✓─ T-022 ✓─ T-023 ─ T-024
       │
       └─ T-025 ✓─ T-026 ─ T-027 ─ T-028 ─ T-029 ✓
```

---

## Weekly Goals

### Week 1 (2026-04-28 to 2026-05-02) — COMPLETED
- [x] T-001, T-002, T-003 (Phase 1 core) — DONE
- [x] T-006 (Update architecture docs) — DONE
- [x] T-007, T-008 (CLI wrappers) — DONE
- [x] T-010 (Tool discovery) — DONE
- [x] T-009, T-011 (Network metrics + fallback) — DONE
- [x] T-012, T-013, T-014 (New tabs) — DONE
- [x] T-015, T-017, T-018 (Tab integration) — DONE
- [x] T-019, T-021, T-022 (Tests) — DONE
- [x] T-025, T-029 (Documentation) — DONE

### Week 2 (2026-05-05 to 2026-05-09)
- [ ] T-016 (Update styles for new tabs)
- [ ] T-020 (Unit tests for new TUI models)
- [ ] T-023 (Manual QA on macOS 14+)

### Week 3-4 (2026-05-12 to 2026-05-23)
- [ ] T-024 (Performance profiling)
- [ ] T-026 (Version constants v2.0.0)
- [ ] T-027 (Build and test binary)
- [ ] T-028 (Create GitHub release)

---

## Report Generation

```bash
# Generate task report
grep -E "^| T-" TASKS.md | head -50

# Count by status
grep "pending" TASKS.md | wc -l
grep "done" TASKS.md | wc -l
```
