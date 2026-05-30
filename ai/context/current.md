# Current Roadmap Position

## Active Spec
`specs/orun-state-redesign/` (Phase 1, local-only) — trigger-first revision-first
local state model. See `specs/orun-state-redesign/README.md` for the index and
read order.

## Active Milestone
**M5 — CLI rewire.** M4 fully closed; M5.a closed (PR #161 → `7a9c494`); **M5.b
closed** (PR #162 → `59d06f3`). **Next: Task 0019 = M5.c implementer**
(`orun status / logs / describe / get` rewire onto
`revisions/<key>/executions/<execKey>/`).

## Last Completed Implementer/Verifier (0018 — M5.b)
- Implementer + verifier both Task 0018 (single-pass) → PR **#162** on
  `impl/task-0018-m5b-orun-run-rewire`.
- Squash-merged to `main` as `59d06f3` "M5.b: rewire `orun run` onto the
  revision-first execution path (#162)" on 2026-05-30T13:42:02Z.
  Head SHA at merge: `e5dd580`.
- Required CI both PASS at log level on final head SHA: `CI / Orun Plan` (45 s);
  `Harness dry-run guard` (12 s). 5 matrix legs SKIPPED (empty matrix at M5.b —
  same shape as M5.a #161).
- Diff stat (5 files changed, +757 / -0):
  - new: `cmd/orun/command_run_revision.go` (365 LOC, houses
    `setupRevisionExecution` / `installRevisionHooks` /
    `finalizeRevisionExecution` / `synthesizeRevisionForRun` /
    `printRevisionRunSummary`).
  - new tests: `cmd/orun/command_run_revision_test.go` (298 LOC, 8 tests).
  - modified: `cmd/orun/command_run.go` (register `--revision` flag,
    wire setup/finalize around `r.Run(plan)`),
    `internal/runner/runner.go` (add `RunnerHooks.AfterStateUpdate` fired
    from `updateState` after `SaveState`),
    `specs/orun-state-redesign/data-model.md` (pin §9.1
    `bridge-mirror-failed` event payload schema).
  - artifacts: `ai/reports/task-0018-implementer.md`,
    `ai/reports/task-0018-verifier.md`, plus ai/state.json + ai/context/*
    updates.
- Coverage: `internal/statestore` **95.7 %** (≥95 %); `internal/revision`
  **90.4 %** (≥90 %); `internal/executionstate` **90.0 %** (exact floor held —
  M5.b touched the package only via API consumption).
- Verifier report: `ai/reports/task-0018-verifier.md`.
- Phase 1 reservations honoured (NOT wired): `--persist-revision` flag
  (synthesize-fallback covers the gap), `Reason="rerun"/"retry"/"migration"`
  (only `"direct-run"` emitted from this path).

## Past Completed (0016 — M5.a)
PR #161 → `7a9c494`. `orun plan` rewire onto canonical revision-first layout.
Verified PASS (single-pass closure).

## Past Completed (0014 / 0015 — M4 PR-B)
PR #160 → `d51e828`. Bridge + EXDEV fallback. Verified PASS by Task 0015.

## Past Completed (0012 / 0013 — M4 PR-A)
PR #159 → `ed48633`. Verified PASS by Task 0013.

## Current Task (0019 — M5.c `orun status / logs / describe / get` rewire Implementer, to be emitted)
- Prompt: TBD (`ai/tasks/task-0019.md`).
- Branch (to be created from `main` @ `59d06f3`):
  `impl/task-0019-m5c-orun-read-commands-rewire`.
- Objective: rewire `orun status` / `orun logs` / `orun describe` /
  `orun get` to read the execution lifecycle from
  `revisions/<key>/executions/<execKey>/` (execution.json + state.json +
  metadata.json + events/), using `refs/latest-execution.json` and the
  resolver legacy-fallback path for compatibility with `.orun/executions/<id>/`.
  Surface `bridge-mirror-failed` events on stderr / metrics. Expose new
  triplet shape in describe output (revisionKey + executionKey +
  legacyExecID).
- Scope boundary: read-only commands only — `cmd/orun/command_status.go`,
  `cmd/orun/command_logs.go`, `cmd/orun/command_describe.go`,
  `cmd/orun/command_get.go` (the `executions` / `runs` subcommands at
  least). EXCLUDES hidden `orun state migrate` (M5.d), runner / writer
  edits, executionstate API changes.
- Acceptance: `orun status` / `logs` / `describe` against a fresh
  `orun plan && orun run` workspace report from the new layout; against
  a legacy `.orun/executions/<id>/` workspace they fall through via the
  resolver legacy-fallback path; `bridge-mirror-failed` events surface
  on stderr (never block). Coverage gates preserved.

## Repo Checkpoint

| Attribute | Value |
|---|---|
| Branch (local checkout) | `main` (clean post-Task-0018 merge) |
| `main` tip | `59d06f3` — M5.b: rewire `orun run` onto the revision-first execution path (#162) |
| Open PRs (state-redesign lineage) | none (PR #162 merged) |
| Repo health | 🟢 Green — M5.b closed; awaiting Task 0019 emission |
| Last verified | 2026-05-30 (Task 0018, PR #162) |
| Active milestone | M5 (CLI rewire) — awaiting Task 0019 (M5.c read-command rewire) implementer |
| Tasks completed | 0001, 0002, 0003, 0004, 0005, 0007, 0008, 0009, 0010, 0011, 0012, 0013, 0014, 0015, 0016, 0018 (16 total) |
| Current task | **0019** (M5.c implementer — to be emitted) |

## Roadmap (M0 → M6)
1. ✅ **M0 Foundation** — landed on main at `4ea1980` (PR #152).
2. ✅ **M1 `internal/triggerctx`** — landed on main at `db342dd` (PR #153).
3. ✅ **M2 `internal/statestore`** — closed at PR #156 (`cd8b3e8`, 2026-05-30).
4. ✅ **M3 `internal/revision`** — closed at PR #158 (`bfc2ae6`, 2026-05-30).
5. ✅ **M4 `internal/executionstate` + runner bridge** — closed.
   - ✅ PR-A — model + writer + resolver (PR #159 → `ed48633`).
   - ✅ PR-B — bridge + EXDEV fallback (PR #160 → `d51e828`).
6. **M5 CLI rewire** ← current. Sub-tasks: ✅ M5.a `orun plan` (Task 0016, PR #161 → `7a9c494`), ✅ M5.b `orun run` + bridge wiring (Task 0018, PR #162 → `59d06f3`), M5.c `orun status / logs / describe / get` (Task 0019), M5.d hidden `orun state migrate`.
7. M6 End-to-end + property gates

## Next Task After 0019 (proposed)
**M5.d implementer** — hidden `orun state migrate` command for legacy
`.orun/executions/<id>/` → `revisions/<key>/executions/<execKey>/`
backfill. After M5.d PASS+merge, M5 closes and M6 (E2E + property gates)
opens.

## Known Spec Drift / Open Questions
- ~~**`bridge-mirror-failed` payload schema not pinned in `data-model.md` §9**~~
  CLOSED in M5.b (§9.1 added; field table matches
  `internal/executionstate.bridgeMirrorFailedPayload` exactly).
- **`MirrorMode` trinary surface** (Task 0015 adjudicated, accepted with Risk
  Note). `MirrorModeAuto` / `MirrorModeHardlink` / `MirrorModeCopy`. Auto is
  zero value matching §M4 verbatim; Hardlink supports drift detection;
  Copy pre-positions remote drivers. Renaming is non-breaking source-level.
  Reconsider when M5/M6 remote-driver Phase 2 wiring picks the right name.
- ~~**`MirrorRunnerOutput` has no production callers until M5.b.**~~ CLOSED in
  M5.b — `cmd/orun/command_run_revision.go::installRevisionHooks` is now the
  first production caller via `RunnerHooks.AfterStateUpdate`. Resolver
  legacy-fallback (PR-A) remains the convergence path for legacy on-disk state.
- **`MirrorModeHardlink` is currently a test/drift-detection mode.** If no
  production caller emerges by M6, fold into a debug flag.
- **`emitFailure` is best-effort** — events-dir-unwritable failures are
  silently dropped. M5.c should add stderr/metric fallback when surfacing
  bridge-mirror-failed events to the read-command audience.
- **Event-sequence retry budget of 32** is acceptable for single-writer
  Phase 1; re-evaluate when remote drivers come online.
- **Manifest required for `UpdateLatestExecutionSummary`** (Task 0013
  carry-forward). Pin normatively in `data-model.md` §4 via proposal when
  M5 needs the option to skip the manifest step.
- **Legacy-execution literal defaults** (Task 0013 carry-forward).
  Pin literals in compat §4 when migration command (compat §5) lands.
- **`internal/executionstate` coverage at 90.0 % exact floor.** Carry-
  forward risk: small refactors deleting covered branches could trip the
  gate.
- **NEW (Task 0018 carry-forward): `RunnerHooks.AfterStateUpdate` fires
  bridge mirror synchronously on the runner goroutine.** On slow filesystems
  this could measurably extend per-tick wall time. M5.c may want to move the
  mirror to a buffered channel + dedicated goroutine if real workloads
  regress.
- **Half-shipped delivery anti-pattern.** Task 0007 first observed; the
  explicit `gh pr list --head` check has shipped every prompt since
  Task 0010 — clean record on Tasks 0010/0012/0014/0016/0018.
