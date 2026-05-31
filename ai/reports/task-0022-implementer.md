# Task 0022 Implementer Report — Phase 2 Rollover

## Summary

Rolled the Orun repository from Phase 1 (`orun-state-redesign`, M0–M6,
closed via PR #165 / release-notes #166) to Phase 2
(`orun-component-catalog`, milestones C0–C9, local-only). This PR is
docs-and-bookkeeping only — no Go code. C0's code half ships
separately as Task 0023.

## Files Changed

**Commit 1 — spec pack (12 files, +2 945):**
- `specs/orun-component-catalog/README.md`
- `specs/orun-component-catalog/design.md`
- `specs/orun-component-catalog/data-model.md`
- `specs/orun-component-catalog/identity-and-keys.md`
- `specs/orun-component-catalog/resolution-pipeline.md`
- `specs/orun-component-catalog/catalog-store.md`
- `specs/orun-component-catalog/compatibility-and-migration.md`
- `specs/orun-component-catalog/cli-surface.md`
- `specs/orun-component-catalog/sync-model.md`
- `specs/orun-component-catalog/implementation-plan.md`
- `specs/orun-component-catalog/test-plan.md`
- `specs/orun-component-catalog/risks-and-open-questions.md`

**Commit 2 — orchestrator + state rotation (2 files, +202 / -93):**
- `agents/orchestrator.md` — rewritten to cite the new spec; adds
  "Deferred Decision Protocol" section (~36 lines) with parking-lot
  pattern, terminal-state rule, and status-briefing requirement. The
  original orchestrator.md never had this wording — Constraint 4
  preserved trivially; addition is proactive.
- `ai/state.json` — rotated to `current_task: "0022"`,
  `next_focus: "orun-component-catalog Milestone C0"`,
  `task_agent: "ai/tasks/task-0022.md"`. Added
  `phase_history.phase_1_orun_state_redesign` block recording M0–M6
  COMPLETE, closure date 2026-05-30, final PR #165, coverage floors
  (`statestore` 95.7 %, `revision` 90.3 %, `executionstate` 90.0 %).

**Commit 3 — archive + context refresh (6 files, +2 519 / -194):**
- `specs/orun-component-catalog/_archive/full-design-monolith.md` —
  the root-level `orun-catalog-full-design.md` moved here under
  `_archive/` (preserves design provenance per user preference).
- `specs/orun-component-catalog/_archive/README.md` — declares files
  under `_archive/` non-authoritative; points at the canonical 12-doc
  set.
- `ai/context/current.md` — refreshed to narrate Phase 1 closure +
  Phase 2 active state. Cites `orun-component-catalog/README.md` and
  `implementation-plan.md` as canonical; milestone cursor at C0;
  Task 0022 in progress; Task 0023 named as next.
- `ai/context/task-ledger.md` — Task 0022 entry status updated to
  "implementer in progress" on the rollover branch.
- `ai/tasks/task-0022.md` — committed for provenance.
- `ai/waiting_for_input.md` — already correctly stated "no input
  requested" pre-task; left unchanged in this PR.

## Checks Run

- `go build ./...` — PASS
- `go test ./...` — PASS (all packages including the Phase 1
  coverage-floored ones: `internal/statestore`, `internal/revision`,
  `internal/executionstate`, `internal/triggerctx`).
- `make verify-generated` — not applicable; spec/docs-only PR.
- `kiox -- orun validate / plan / run --dry-run` — not run; this PR
  touches no Go and no intent files. CI is authoritative.
- Required CI on PR head — not yet observed at report write time;
  branch pushed and PR #167 just opened.

## Assumptions

1. Phase 2 spec content (the 12 docs under `specs/orun-component-catalog/`)
   was provided pre-authored by the orchestrator and is shipped as-is in
   this PR. Substantive edits to spec content go through `/ai/proposals/`
   in a separate cycle.
2. The Deferred Decision Protocol section added to `agents/orchestrator.md`
   is a proactive improvement (not a regression repair) — the prior
   orchestrator.md only had "Human Input Pause Protocol".
3. Example task IDs (`task-0021`) embedded in orchestrator.md as
   illustrative placeholders were left in place; they are not active
   references.
4. `task_agent` path in `ai/state.json` uses repo-relative form
   (`ai/tasks/task-0022.md`) without leading slash, consistent with
   actual file layout.
5. The root-level monolith `orun-catalog-full-design.md` is archived,
   not deleted, per user preference for preserving design provenance.

## Spec Proposals

None. No spec edits are required by this PR. If implementation of C0
(Task 0023) surfaces drift between the spec docs and the code, those
edits will be raised as `/ai/proposals/` artifacts at that time.

## Remaining Gaps

1. CI status on PR #167 not yet confirmed PASS — verifier should
   re-check before merging.
2. `ai/state.json` `task_agent` field currently points at the
   implementer prompt (`ai/tasks/task-0022.md`); per protocol, after
   this report is committed, `task_agent` flips to point at this
   report. That commit ships under this same PR (follow-up commit on
   this branch before merge) or as a tiny follow-up PR — orchestrator's
   call.
3. `ai/deferred.md` does not yet exist. The Deferred Decision Protocol
   section in `agents/orchestrator.md` references it but creation is
   intentionally deferred until the first parked candidate emerges.
4. No verification of kiox/orun runtime checks in this cycle — the
   PR is spec-only and CI is authoritative for runtime gates.

## Next Task Dependencies

**Task 0023 — C0 code half** depends on this PR merging to `main` so
the canonical 12-doc spec pack is the discoverable source for the
implementer. Reads
`specs/orun-component-catalog/{data-model.md, identity-and-keys.md, implementation-plan.md§C0, test-plan.md§1}`
and ships:

- `internal/catalogmodel` — pure data types, canonical-JSON encoder,
  JSON-Schema generator (under `internal/catalogmodel/schema/` via
  `go generate ./...`), golden roundtrip fixtures under `testdata/`.
  Leaf-clean: imports no other `internal/` package. ≥ 90 % coverage;
  100 % on `Sanitize*`.
- `internal/sourcectx` — skeleton types only (Git resolver lands in C1).
- `make test-state-redesign` extension to gate the new packages.

Phase 1 invariants (Constraint 4) carry forward unchanged: do not
rename Phase 1 fields, do not lower Phase 1 coverage floors, do not
remove preserved Phase 1 CLI surface.

## PR Number

**#167** — https://github.com/sourceplane/orun/pull/167

Branch: `impl/task-0022-phase2-rollover`
Commits on branch (3):
- `7503297` [1/3] land orun-component-catalog spec pack
- `c3743a9` [2/3] rotate orchestrator + state.json
- `455d635` [3/3] archive monolith, refresh context, land task prompt
