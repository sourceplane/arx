# Open Risks

## Spec Drift

- **R-001: `flyingmutant/rapid` vs `pgregory.net/rapid`.** `specs/orun-state-redesign/test-plan.md` §1 references `github.com/flyingmutant/rapid` as the property-based test library. `go.mod` already pins `pgregory.net/rapid v1.1.0` (the same library under its current import path). Implementer agents must use `pgregory.net/rapid`; a small spec-clarification proposal should be filed under `/ai/proposals/` when convenient. Not a blocker for any milestone.

## Compatibility

- **R-002: ExecID shape preservation for github-artifacts cross-spec compatibility.** The new revision/execution keys defined in `data-model.md` must remain compatible with the existing `gh-{run_id}-{attempt}-{sha}` ExecID format produced by `internal/runbundle`. Cross-check during M4 (executionstate + runner bridge) and M5 (CLI rewire).

## Local-only Phase 1 Boundary

- **R-003: Out-of-scope creep.** Phase 1 explicitly excludes R2/S3/Cloud StateStore driver, Supabase/DO coordination, distributed locking, TUI surface changes, and deletion of legacy `.orun/executions/` paths. Implementer prompts must call this out per task to prevent drift.

## Operational

- **R-004: `ai/` directory rebuilt from scratch.** The 2026-05-30 pivot left the old `ai/` tree as unstaged deletions in `git status`. The orchestrator has rebuilt the four compact context files plus `state.json` and `task-0001.md`. The deletions of old TUI-era tasks/reports must be committed (or reverted) when the first M0 PR opens — currently they sit dirty in the working tree. Implementer for Task 0001 will fold these `ai/` deletions into their PR so main has a coherent tree.
