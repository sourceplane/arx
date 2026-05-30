# Current Roadmap Position

## Active Spec
`specs/orun-state-redesign/` (Phase 1, local-only) — trigger-first revision-first
local state model. See `specs/orun-state-redesign/README.md` for the index and
read order.

## Active Milestone
**M4 — `internal/executionstate` + runner bridge.** PR-A landed at PR #159
(`ed48633`, verified PASS via Task 0013 on 2026-05-30). **Next: Task 0014 =
M4 PR-B implementer** (`bridge.go`, `MirrorRunnerOutput`, EXDEV fallback). M4
closes when Task 0015 (PR-B verifier) PASSes.

## Last Completed Implementer/Verifier (0012 / 0013 — M4 PR-A)
- Implementer Task 0012 → PR **#159** on `impl/task-0012-m4-executionstate-pra`
  @ `8a0c409` (parent `2c239d7` "Task 0012: M4 PR-A …", child `8a0c409`
  "docs: add task-0012 implementer report").
- Verifier Task 0013 → PR #159 squash-merged to `main` as `ed48633`
  "feat(executionstate): land M4 PR-A — internal/executionstate model + writer
  + resolver".
- Required CI both SUCCESS at log level on PR-A:
  - `CI / Orun Plan` — run `26675724704`, completed 2026-05-30T05:30:17Z.
  - `orun remote-state conformance / Harness dry-run guard` — run
    `26675724720`, completed 2026-05-30T05:29:47Z.
- Diff stat (12 files, +2593 / −2):
  - new: `internal/executionstate/{model,writer,resolver,internal}.go` plus
    `internal/executionstate/{model,writer,resolver,property,coverage_extra}_test.go`
  - additive: `internal/statestore/paths.go` (`ExecutionsDir`,
    `ExecutionDocPath`, `ExecutionIndex*`, `LegacyExecution*`, `EventPath`,
    `SnapshotPath`)
  - Makefile coverage gate extended for `internal/executionstate`
  - artifacts: `ai/tasks/task-0012.md`, `ai/reports/task-0012-implementer.md`,
    `ai/tasks/task-0013-verifier.md`, `ai/reports/task-0013-verifier.md`
- Coverage: `internal/executionstate` **90.0 %** (exact floor, gate ≥ 90 %);
  `internal/revision` 90.4 %; `internal/statestore` 95.4 %.
- Implementer report: `ai/reports/task-0012-implementer.md`. Verifier report:
  `ai/reports/task-0013-verifier.md`.

## Current Task (0014 — M4 PR-B Implementer, scoped)
- Prompt: `ai/tasks/task-0014.md` (just emitted).
- Branch (to be created from `main` @ `ed48633`):
  `impl/task-0014-m4-executionstate-prb`.
- Objective: ship `internal/executionstate/bridge.go` —
  `Bridge{Store, LegacyRoot, MirrorMode}` + `MirrorRunnerOutput(ctx, execKey,
  revKey, legacyExecID)` — with hardlink-with-copy-fallback, EXDEV-injection
  test seam, `bridge-mirror-failed` event emission, idempotent re-mirror, and
  the leaf-clean invariant preserved.
- Scope boundary: `internal/executionstate/bridge.go` (+ tests), `Makefile`
  if needed, additive helpers in `internal/statestore/paths.go` only. NO
  `cmd/orun`, NO `internal/state`, NO `internal/runner`, NO
  `internal/runbundle`, NO production-runner wiring (M5 owns CLI rewire).
- Acceptance: `internal/executionstate` ≥ 90 % preserved (PR-B should lift
  the floor); hardlink success path tested; forced-EXDEV copy fallback tested
  via injected `linker` seam; `bridge-mirror-failed` event emission tested;
  idempotent re-mirror tested; both required CI checks at minimum queued.

## Repo Checkpoint

| Attribute | Value |
|---|---|
| Branch (local checkout) | `main` (clean post-Task-0013 merge) |
| `main` tip | `ed48633` — feat(executionstate): land M4 PR-A — internal/executionstate model + writer + resolver (#159) |
| Open PRs (state-redesign lineage) | none (PR #159 merged) |
| Repo health | 🟢 Green — M4 PR-B awaiting implementer |
| Last verified | 2026-05-30 (Task 0013, PR #159) |
| Active milestone | M4 (`internal/executionstate`) — PR-A merged; PR-B awaiting implementer |
| Tasks completed | 0001, 0002, 0003, 0004, 0005, 0007, 0008, 0009, 0010, 0011, 0012, 0013 (12 total) |
| Current task | **0014** (M4 PR-B implementer — emitted) |

## Roadmap (M0 → M6)
1. ✅ **M0 Foundation** — landed on main at `4ea1980` (PR #152).
2. ✅ **M1 `internal/triggerctx`** — landed on main at `db342dd` (PR #153).
3. ✅ **M2 `internal/statestore`** — closed at PR #156 (`cd8b3e8`, 2026-05-30).
4. ✅ **M3 `internal/revision`** — closed at PR #158 (`bfc2ae6`, 2026-05-30).
5. **M4 `internal/executionstate` + runner bridge** ← current
   - ✅ PR-A — model + writer + resolver (PR #159 → `ed48633`, verified PASS via Task 0013 on 2026-05-30)
   - **PR-B — bridge + EXDEV fallback (Task 0014 implementer scoped)**
6. M5 CLI rewire (`orun plan/run/status/logs/describe/get plans` + hidden `state migrate`)
7. M6 End-to-end + property gates

## Next Task After 0014 (proposed)
**Task 0015 — M4 PR-B verifier.** Verifies PR-B against the M4 "Done when"
checklist (NextExecutionKey monotonicity property still green, hardlink +
EXDEV-copy paths both covered, resolver legacy-fallback still green from PR-A,
coverage floor preserved or lifted). On PASS + merge, M4 closes and Task
0016 = M5.a (`orun plan` rewire) implementer becomes the next emission. If
the verifier identifies normative gaps (e.g. `MirrorMode` enumeration,
`bridge-mirror-failed` event payload), a small `ai/proposals/task-0014-spec-update.md`
gets adjudicated before Task 0016 is generated.

## Known Spec Drift / Open Questions
- **Manifest required for `UpdateLatestExecutionSummary`** (Task 0013 carry-
  forward). Implementer chose loud `ErrNotFound` surfacing when called
  against a revision missing `manifest.json`. Conservative-on-unknowns
  matches `revision/writer.go` style; pin normatively in `data-model.md` §4
  via proposal when M5 needs the option to skip the manifest step.
- **Legacy-execution literal defaults** (Task 0013 carry-forward).
  `triggerKey=system.migrated`, `Reason=migration`, `Status=completed`
  default chosen by Task 0012 — defensible read of compat §4 but not
  strictly normative. Pin literals in compat §4 when migration command
  (compat §5) lands.
- **`MirrorMode` enumeration** (open for Task 0014). Spec says
  "hardlink with copy fallback" but does not enumerate `MirrorModeHardlink
  / MirrorModeCopy / MirrorModeAuto` values normatively. Implementer is
  authorized to pick a defensible enumeration and document the choice;
  verifier adjudicates.
- **`bridge-mirror-failed` event payload shape** (open for Task 0014).
  `data-model.md` §9 names the event but does not enumerate every payload
  field. Implementer documents; verifier adjudicates.
- **`internal/executionstate` coverage at 90.0 % exact floor.** Carry-
  forward risk: small refactors deleting covered branches could trip the
  gate. PR-B is expected to lift the floor.
- **`stateStoreVersionPath()` helper location** (carried from M3 PR-A).
  RESOLVED — defer to M5 if migration tooling needs a statestore-side
  helper.
- **Half-shipped delivery anti-pattern.** Task 0007 was the first observed
  case. Tasks 0010 / 0012 prompts embedded the explicit `gh pr list --head`
  check; both shipped on the first delivery cycle. Task 0014 prompt carries
  the same guard.
