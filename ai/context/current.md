# Current Roadmap Position

## Active Spec
`specs/orun-state-redesign/` (Phase 1, local-only) — trigger-first revision-first
local state model. See `specs/orun-state-redesign/README.md` for the index and
read order.

## Active Milestone
**M5 — CLI rewire.** M4 fully closed; **M5.a closed** (PR #161 → `7a9c494`).
**Next: Task 0018 = M5.b implementer** (`orun run` rewire + bridge wiring +
`--revision` flag + pin `bridge-mirror-failed` payload schema in §9).

## Last Completed Implementer/Verifier (0016 — M5.a)
- Implementer + verifier both Task 0016 (single-pass) → PR **#161** on
  `impl/task-0016-m5a-orun-plan-rewire`.
- Squash-merged to `main` as `7a9c494` "Task 0016: M5.a — orun plan rewire to
  revision-first layout (#161)" on 2026-05-30T12:31:56Z.
- Required CI both PASS at log level on final head SHA: `CI / Orun Plan` run
  `26683860043` (44s); `Harness dry-run guard` run `26683860052` (15s).
- Diff stat (10 files changed, +505 / -65):
  - modified: `cmd/orun/main.go` (+193), `internal/model/plan.go` (+18),
    `internal/revision/legacy.go` (+30).
  - new tests: `cmd/orun/command_plan_revision_test.go` (4 tests),
    `internal/revision/legacy_test.go` (4 tests).
  - artifacts: `ai/tasks/task-0016.md`, `ai/reports/task-0016-verifier.md`,
    plus ai/state.json + ai/context/* + ai/waiting_for_input.md updates.
- Coverage: `internal/statestore` **95.7 %** (≥95 %); `internal/revision`
  **90.4 %** (≥90 %); `internal/executionstate` **90.0 %** (exact floor held —
  M5.a did not touch the package).
- Verifier report: `ai/reports/task-0016-verifier.md`.

## Past Completed (0014 / 0015 — M4 PR-B)
PR #160 → `d51e828`. Bridge + EXDEV fallback. Verified PASS by Task 0015.

## Past Completed (0012 / 0013 — M4 PR-A)
PR #159 → `ed48633`. Verified PASS by Task 0013.

## Current Task (0018 — M5.b `orun run` rewire Implementer, to be emitted)
- Prompt: TBD (`ai/tasks/task-0018.md`).
- Branch (to be created from `main` @ `7a9c494`):
  `impl/task-0018-m5b-orun-run-rewire`.
- Objective: rewire `orun run` to resolve PlanRevision via
  `internal/revision.ResolveRevision`, materialize a system.manual revision
  in-memory when no revision exists, create executions via
  `internal/executionstate.CreateExecution`, hook the runner snapshot stream
  into `Bridge.MirrorRunnerOutput`, add `--revision <key>` flag (skips
  resolution chain). Pin the `bridge-mirror-failed` payload schema in
  `data-model.md` §9 before any second consumer.
- Scope boundary: `cmd/orun/command_run.go` + minimal runner-snapshot bridge
  glue + `data-model.md` §9 pin. EXCLUDES `orun status/logs/describe/get
  plans` (M5.c) and hidden `orun state migrate` (M5.d).
- Acceptance: `orun run` end-to-end against `examples/intent.yaml` produces
  canonical execution layout under `revisions/<key>/executions/<execKey>/`,
  `refs/latest-execution.json`, `indexes/executions/<execKey>.json`, while
  legacy `.orun/executions/<legacyExecID>/` continues to be mirrored via
  Bridge. `--revision <key>` resolves directly without the legacy chain.
  Coverage gates preserved.

## Repo Checkpoint

| Attribute | Value |
|---|---|
| Branch (local checkout) | `main` (clean post-Task-0016 merge) |
| `main` tip | `7a9c494` — Task 0016: M5.a — orun plan rewire to revision-first layout (#161) |
| Open PRs (state-redesign lineage) | none (PR #161 merged) |
| Repo health | 🟢 Green — M5.a closed; awaiting Task 0018 emission |
| Last verified | 2026-05-30 (Task 0016, PR #161) |
| Active milestone | M5 (CLI rewire) — awaiting Task 0018 (M5.b `orun run`) implementer |
| Tasks completed | 0001, 0002, 0003, 0004, 0005, 0007, 0008, 0009, 0010, 0011, 0012, 0013, 0014, 0015, 0016 (15 total) |
| Current task | **0018** (M5.b implementer — to be emitted) |

## Roadmap (M0 → M6)
1. ✅ **M0 Foundation** — landed on main at `4ea1980` (PR #152).
2. ✅ **M1 `internal/triggerctx`** — landed on main at `db342dd` (PR #153).
3. ✅ **M2 `internal/statestore`** — closed at PR #156 (`cd8b3e8`, 2026-05-30).
4. ✅ **M3 `internal/revision`** — closed at PR #158 (`bfc2ae6`, 2026-05-30).
5. ✅ **M4 `internal/executionstate` + runner bridge** — closed.
   - ✅ PR-A — model + writer + resolver (PR #159 → `ed48633`).
   - ✅ PR-B — bridge + EXDEV fallback (PR #160 → `d51e828`).
6. **M5 CLI rewire** ← current. Sub-tasks: ✅ M5.a `orun plan` (Task 0016, PR #161 → `7a9c494`), M5.b `orun run` + bridge wiring (Task 0018), M5.c `orun status/logs/describe/get plans`, M5.d hidden `orun state migrate`.
7. M6 End-to-end + property gates

## Next Task After 0018 (proposed)
**M5.c implementer** — `orun status` / `logs` / `describe` / `get plans`
rewire onto `refs/latest-execution.json` + `indexes/executions/`. After
M5.c PASS+merge, M5.d (hidden `orun state migrate`) closes M5.

## Known Spec Drift / Open Questions
- **`bridge-mirror-failed` payload schema not pinned in `data-model.md` §9**
  (Task 0015 carry-forward). PR-B fixed the schema in code:
  `{executionKey, revisionKey, legacyExecId, artifact, stage, mode, error}`
  with `stage ∈ {read-source, read-dest, translate-dest, mkdir-dest,
  remove-dest, link, copy}`. Schema is well-formed and additive-friendly.
  Pin in §9 during M5.b runner wiring before any second consumer (metrics,
  `orun status`) lands.
- **`MirrorMode` trinary surface** (Task 0015 adjudicated, accepted with Risk
  Note). `MirrorModeAuto` / `MirrorModeHardlink` / `MirrorModeCopy`. Auto is
  zero value matching §M4 verbatim; Hardlink supports drift detection;
  Copy pre-positions remote drivers. Renaming is non-breaking source-level.
  Reconsider when M5/M6 remote-driver Phase 2 wiring picks the right name.
- **`MirrorRunnerOutput` has no production callers until M5.b.** Resolver
  legacy-fallback (PR-A) carries convergence burden in the meantime.
- **`MirrorModeHardlink` is currently a test/drift-detection mode.** If no
  production caller emerges by M6, fold into a debug flag.
- **`emitFailure` is best-effort** — events-dir-unwritable failures are
  silently dropped. M5+ should add stderr/metric fallback.
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
- **Half-shipped delivery anti-pattern.** Task 0007 first observed; the
  explicit `gh pr list --head` check has shipped every prompt since
  Task 0010 — clean record on Tasks 0010/0012/0014.
