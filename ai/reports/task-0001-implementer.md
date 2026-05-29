# Task 0001 — Implementer Report

**Agent:** Implementer
**Task:** 0001
**Milestone:** M0 — Foundation
**Spec:** `specs/orun-state-redesign/`
**PR:** [#152](https://github.com/sourceplane/orun/pull/152)
**Branch:** `impl/task-0001-m0-foundation`
**Base:** `main`

## Summary
- Landed the roadmap pivot from `.kiro/specs/orun-tui-cockpit/` to `specs/orun-state-redesign/` on a single feature branch — the spec pack, the source design doc, the orchestrator-doc edits, and the rebuilt `ai/` tree (plus deletions of the TUI-era task/report files) all ride together so `main` will reflect the pivot coherently when merged.
- Pinned `github.com/oklog/ulid/v2 v2.1.1` as a direct dependency. Kept live via a `tools`-tagged blank import in `internal/testfx/statefs/tools.go` until M1 introduces the first real production import — at which point that file can be deleted.
- Scaffolded `internal/testfx/statefs` with `NewWorkspace`, `AssertJSONFile`, and `ReadJSON[T any]`. Helpers accept `testing.TB` so their failure paths can be unit-tested via a fakeT wrapper; happy + failure paths covered for all three. Package imports stdlib + `testing` only — no `internal/*` imports.
- Added `make test-state-redesign` Makefile target running `./internal/testfx/statefs/...` with a marker comment for the M1+ packages.
- All local checks green: `go build ./...`, `go vet ./...`, `go test ./...`, `go test -count=1 ./internal/testfx/statefs/...`, `make test-state-redesign`, and the orun validate / plan --changed / run --dry-run sanity loop on `examples/intent.yaml`.

## Files Changed
**Spec pack (new):**
- `specs/orun-state-redesign/README.md`
- `specs/orun-state-redesign/design.md`
- `specs/orun-state-redesign/data-model.md`
- `specs/orun-state-redesign/state-store.md`
- `specs/orun-state-redesign/cli-surface.md`
- `specs/orun-state-redesign/test-plan.md`
- `specs/orun-state-redesign/implementation-plan.md`
- `specs/orun-state-redesign/compatibility-and-migration.md`
- `specs/orun-state-redesign/risks-and-open-questions.md`
- `orun-state-redesign.md` (source design doc at repo root)

**Orchestrator doc:**
- `agents/orchestrator.md` (modified — points at the new spec)

**`ai/` tree (rebuilt):**
- `ai/state.json` (modified — new lineage)
- `ai/context/current.md` (modified)
- `ai/context/decisions.md` (new)
- `ai/context/open-risks.md` (new)
- `ai/context/task-ledger.md` (modified)
- `ai/waiting_for_input.md` (modified)
- `ai/tasks/task-0001.md` (new)
- `ai/reports/task-0001-implementer.md` (this file)
- Deleted: `ai/reports/task-014{0..7}-{implementer,verifier}.md`, `ai/tasks/task-014{1..7}{,-verifier}.md` (TUI-era).

**Dependency pin:**
- `go.mod`, `go.sum` — `github.com/oklog/ulid/v2 v2.1.1` as direct require.

**`internal/testfx/statefs` (new package):**
- `internal/testfx/statefs/statefs.go` — `NewWorkspace`, `AssertJSONFile`, `ReadJSON[T]`.
- `internal/testfx/statefs/statefs_test.go` — happy + failure paths for all three helpers (parallel-safe, internal package test).
- `internal/testfx/statefs/tools.go` — `//go:build tools` blank import that keeps `github.com/oklog/ulid/v2` a direct require until M1.

**Makefile:**
- `Makefile` — `test-state-redesign` target + `.PHONY` listing.

## Checks Run
| Command | Exit |
|---|---|
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `go test ./...` | 0 |
| `go test -count=1 ./internal/testfx/statefs/...` | 0 |
| `make test-state-redesign` | 0 |
| `go list -m github.com/oklog/ulid/v2` → `v2.1.1` | 0 |
| `grep -r '"internal/' internal/testfx/statefs` → empty | 1 (grep no-match, expected) |
| `go run ./cmd/orun validate --intent examples/intent.yaml` | 0 |
| `go run ./cmd/orun plan --changed --intent examples/intent.yaml --output /tmp/orun-t1-plan.json` | 0 (no-op: 0 components × 5 envs → 0 jobs) |
| `go run ./cmd/orun run --plan /tmp/orun-t1-plan.json --dry-run --runner github-actions` | 0 (no jobs to run) |

## Assumptions
- The `flyingmutant/rapid` import path called out in `specs/orun-state-redesign/test-plan.md §3` is interpreted as a stale name for the already-pinned `pgregory.net/rapid v1.1.0`. Used the current import path; left `pgregory.net/rapid` in `go.mod` untouched (per the task's explicit non-goal). A small spec-clarification proposal is recommended but was not filed in this PR — see "Spec Proposals" below.
- Typed the test helpers on `testing.TB` rather than `*testing.T`. The task prompt spells the signatures with `*testing.T`; `testing.TB` is a deliberate superset (`*testing.T` and `*testing.B` both satisfy it) and lets the helpers' failure paths be unit-tested from a fakeT wrapper without aborting the outer test. Caller ergonomics are unchanged.
- The ULID dependency is gated behind a `tools`-tagged blank import (`internal/testfx/statefs/tools.go`) so `go mod tidy` keeps it in the direct-require block while no production code uses it yet. The file is documented to be deleted by M1.
- Kept the spec pivot and M0 deliverables in one PR (the latitude clause permitted a split, but a single coherent PR was easier to review and there's no CI risk in bundling them).

## Spec Proposals
None filed in this PR. A small clarification under `ai/proposals/task-0001-spec-update.md` is encouraged to rename `flyingmutant/rapid` → `pgregory.net/rapid` in `specs/orun-state-redesign/test-plan.md §3`; deferring to the next implementer or orchestrator cycle.

## Remaining Gaps
None for M0. The `tools.go` blank import is a known transitional artefact — M1 should remove it as soon as `internal/triggerctx` introduces a real production import of `github.com/oklog/ulid/v2`.

## Next Task Dependencies
M0 done unblocks:
- **M1 — `internal/triggerctx`**: can now import `github.com/oklog/ulid/v2` for the `trg_`-prefixed monotonic ULID generator, and consume `internal/testfx/statefs` (`NewWorkspace`, `AssertJSONFile`, `ReadJSON[T]`) for its table-driven coverage of `ResolveTriggerContext` branches.
- **M2 — `internal/statestore`** (local driver): same dependency surface — workspace fixtures for atomic temp+rename tests, JSON helpers for ref/index round-trips.

## PR Number
**#152** — https://github.com/sourceplane/orun/pull/152
