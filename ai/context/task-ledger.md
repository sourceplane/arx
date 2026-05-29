# Task Ledger — orun-state-redesign lineage

Task numbering restarted at 0001 with the 2026-05-30 pivot from
`.kiro/specs/orun-tui-cockpit/` (paused at PR #146 merged) to
`specs/orun-state-redesign/` (Phase 1 local-only trigger-first revision-first
local state model).

Prior TUI cockpit tasks (0139–0147 with sub-tasks) are preserved in git
history; reports/prompts are no longer present in the working tree.

---

## Task 0001

|- Agent: Implementer
|- Prompt: `ai/tasks/task-0001.md`
|- Status: scoped and ready to begin (2026-05-30)
|- Milestone: M0 — Foundation
|- Objective: pin `github.com/oklog/ulid/v2`, scaffold `internal/testfx/statefs`
   helper package with `NewWorkspace`, `AssertJSONFile`, and `ReadJSON[T]`,
   and add `make test-state-redesign` (initially a no-op).
|- Scope boundary: `go.mod`/`go.sum`, `internal/testfx/statefs/`, `Makefile`.
   No production-code or CLI-surface changes. Implementer may split into two
   PRs (dependency pin vs. harness) if cleaner; both must land before M1.
|- Acceptance: `go build ./...` and `go test ./...` green; unit tests for all
   three helpers; `internal/testfx/statefs` imports no other `internal/` package;
   `make test-state-redesign` target exists and runs.
|- Expected outcome: M1 (`internal/triggerctx`) and M2 (`internal/statestore`)
   are unblocked.

## Historical Notes

- 2026-05-30: roadmap pivoted from TUI cockpit (Phase 3) to orun-state-redesign
  Phase 1. `agents/orchestrator.md` and `specs/orun-state-redesign/` set the
  new authoritative spec pack. `ai/` directory rebuilt by orchestrator under
  the new lineage.
