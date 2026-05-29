# Current Roadmap Position

## Active Spec
`specs/orun-state-redesign/` (Phase 1, local-only) — trigger-first revision-first
local state model. See `specs/orun-state-redesign/README.md` for the index and
read order.

## Active Milestone
**M0 — Foundation.** Unblock every later milestone without touching production code.

## Current Task (0001)
- Agent: Implementer
- Prompt: `ai/tasks/task-0001.md`
- Objective: pin `github.com/oklog/ulid/v2`, scaffold `internal/testfx/statefs`
  test helpers (`NewWorkspace`, `AssertJSONFile`, `ReadJSON[T]`), and add a
  `make test-state-redesign` target (initially a no-op until M1+ packages land).
- PR Boundary: one PR for `go.mod` / `go.sum` change, `internal/testfx/statefs/`
  with helpers + unit tests, and the Makefile target. No production-code or
  CLI-surface changes. Implementer has latitude to split into two PRs (dependency
  pin vs. harness) if it improves reviewability; both must land before M1.

## Next Task After 0001
**Task 0002 (Implementer) — Milestone M1: `internal/triggerctx`.**
Model `TriggerOccurrence`, `TriggerSource`, `PlanScope`; ULID-prefixed ID
generator (`trg_`); system trigger constructors; `FromDeclaredTrigger` +
`ResolveProviderEvent` wrapping `internal/trigger`; top-level
`ResolveTriggerContext` dispatcher. ≥90 % coverage, property test on
`TriggerKey` stability + format. Depends on M0.

## Repo Checkpoint

| Attribute | Value |
|---|---|
| Branch | main (synced with origin/main) |
| Last commit on main | d2ab48e — docs: rearchitect website around planner–cockpit positioning (#151) |
| Open PRs | none |
| Repo health | 🟢 Green |
| Last verified | 2026-05-29 (TUI cockpit Task 0147.1, PR #146) |
| Active milestone | M0 (Foundation) |

## Roadmap (M0 → M6)
1. **M0 Foundation** ← current
2. M1 `internal/triggerctx`
3. M2 `internal/statestore` (local driver) — contract frozen here
4. M3 `internal/revision`
5. M4 `internal/executionstate` + runner bridge
6. M5 CLI rewire (`orun plan/run/status/logs/describe/get plans` + hidden `state migrate`)
7. M6 End-to-end + property gates

## Known Spec Drift / Open Questions
- `specs/orun-state-redesign/test-plan.md` §1 references
  `github.com/flyingmutant/rapid`; `go.mod` already pins `pgregory.net/rapid`
  v1.1.0 (the same module under its current import path). Implementer should
  use the already-pinned `pgregory.net/rapid` and may file a small
  spec-clarification proposal under `/ai/proposals/`. Not a blocker.

## Secondary Specs (not driving new tasks this phase)
- `.kiro/specs/orun-tui-cockpit/` — paused. Resumes after M5 lands.
- `.kiro/specs/github-artifacts/` — cross-check only; new revision/execution
  keys must remain compatible with the existing
  `gh-{run_id}-{attempt}-{sha}` ExecID shape produced by `internal/runbundle`.
