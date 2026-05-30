# Task 0018 — Implementer Report

Agent: Implementer (M5.b)

## PR

**#162** — https://github.com/sourceplane/orun/pull/162
Branch: `impl/task-0018-m5b-orun-run-rewire` (squash-merged 2026-05-30T13:42:02Z, deleted)
Merge commit: `59d06f3` "M5.b: rewire `orun run` onto the revision-first execution path (#162)"
Head SHA at merge: `e5dd5808c0f982c44da8bca0d3b34dac8866cee3`

## Scope

`specs/orun-state-redesign/implementation-plan.md` §M5.b — rewire `orun
run` onto the revision-first execution path:

1. Resolve `PlanRevision` via `internal/revision.ResolveRevision` (the
   seven-branch resolver from `compatibility-and-migration.md` §3) when
   the user invokes `orun run`.
2. Synthesize a fresh `system.manual` revision when the resolver returns
   a benign miss, so a fresh `orun run` (no preceding `orun plan`) still
   lands a real on-disk triplet under `revisions/<key>/`.
3. Persist an `ExecutionRun` under that revision via
   `internal/executionstate.CreateExecution`, mirror the runner's
   `state.json` / `metadata.json` into the new layout via
   `Bridge.MirrorRunnerOutput` on every runner tick (mode `Auto`:
   hardlink → copy fallback on EXDEV per `design.md` §11), and flip the
   execution to a terminal status with summary counts when the runner
   returns.
4. Add `--revision <key>` flag (cli-surface.md §2.3) that
   short-circuits the resolution chain by feeding the raw value
   straight to `ResolveRevision` (binds on branch 3 / 4 / 5).
5. Pin the `bridge-mirror-failed` event payload schema in
   `data-model.md` §9 before any second consumer (carried forward as
   risk note from Task 0016).

Scope EXCLUDES `orun status / logs / describe / get` (M5.c) and the
hidden `orun state migrate` command (M5.d).

## Diff Summary

5 files changed, +757 / -0:

- `cmd/orun/command_run.go` — register `--revision` flag, call
  `setupRevisionExecution` + `installRevisionHooks` before
  `r.Run(plan)`, `finalizeRevisionExecution` + `printRevisionRunSummary`
  on return. Skipped for `--dry-run` and remote-state runs (their
  lifecycle is owned upstream).
- `cmd/orun/command_run_revision.go` *(new, 365 LOC)* — opens the local
  statestore, calls `revision.ResolveRevision` (latest / file /
  revision-key / named-ref / legacy-hash / component-name / ambiguous),
  synthesizes-and-re-resolves on a benign miss so `CreateExecution`
  gets a populated `RevisionID` / `TriggerID`, builds a `Bridge` with
  `MirrorModeAuto`, and chains the mirror onto
  `RunnerHooks.AfterStateUpdate`. `finalizeRevisionExecution` projects
  `state.SummarizeExecutionState` → `executionstate.ExecSummary` and
  calls `MarkTerminal`.
- `cmd/orun/command_run_revision_test.go` *(new, 298 LOC)* — covers
  synth round-trip, happy-path setup + finalize (status=completed,
  summary tally), failed-runner path, `--revision` short-circuit,
  `--exec-id` plumbing into `OriginalKey`, nil-rx no-op, absolute
  store-root invariant, nil-plan / empty-execID rejection.
- `internal/runner/runner.go` — add `RunnerHooks.AfterStateUpdate`,
  fired from `updateState` after `Store.SaveState`. Decouples
  `internal/runner` from `internal/executionstate` while still letting
  the CLI layer drive the bridge.
- `specs/orun-state-redesign/data-model.md` — pin §9.1
  `bridge-mirror-failed` event payload schema. Field table matches
  `internal/executionstate.bridgeMirrorFailedPayload` exactly.

## Tests

Resolver branches 1–7 themselves are exhaustively tested upstream in
`internal/revision/resolver_test.go` (`TestResolveRevision_Branch1..7`).
The new CLI-layer tests cover the glue:

- `TestSynthesizeRevisionForRun_PersistsRevisionTriplet` — synthesized
  rev is round-trippable through `ResolveRevision`.
- `TestSetupAndFinalizeRevisionExecution_HappyPath` — end-to-end
  setup + finalize lifecycle, `state.json` mirror, summary tally.
- `TestSetupRevisionExecution_FailedRunMarksFailed` — non-nil `runErr`
  forces `status=failed` even when ExecutionCounts is empty.
- `TestSetupRevisionExecution_RevisionFlagShortCircuit` — `--revision
  <key>` binds via `ResolveSourceRevisionKey` (branch 3).
- `TestSetupAndFinalizeRevisionExecution_HappyPath` — `--exec-id`
  plumbed through `CreateExecutionInput.OriginalKey`.
- `TestSetupRevisionExecution_StoreRootIsAbsolute` — bridge LegacyRoot
  invariant.
- `TestSetupRevisionExecution_NilPlanRejected`,
  `_EmptyExecIDRejected`, `TestFinalizeRevisionExecution_NilRxIsNoOp` —
  defensive guards.

## Coverage

`make test-state-redesign` final on PR head:

- `internal/statestore` **95.7 %** (≥ 95 % gate)
- `internal/revision` **90.4 %** (≥ 90 % gate)
- `internal/executionstate` **90.0 %** (≥ 90 % gate, exact floor held)

`go build ./...`, `go vet ./...`, `go test -race ./...` (full module)
all green.

## CI

PR #162 final state at merge: `state=OPEN`, `mergeable=MERGEABLE`,
`mergeStateStatus=CLEAN`. Required CI checks PASS at log level on head
SHA `e5dd580`:

- `CI / Orun Plan` — PASS (45 s).
- `orun remote-state conformance / Harness dry-run guard` — PASS (12 s).
- 5 matrix legs SKIPPED (empty matrix at M5.b — same shape as M3 / M4 /
  M5.a PRs).

## E2E Smoke

`rm -rf .orun && orun plan && orun run --component web-console-pages
--env preview` against `examples/intent.yaml`:

```
.orun/revisions/rev-manual-no-git-pd8ece5af/executions/run-001/
├── execution.json          (status=completed, summary={total:1,completed:1,…})
├── state.json              (mirrored via Bridge AfterStateUpdate)
├── metadata.json           (mirrored via Bridge AfterStateUpdate)
└── events/00000000000000000001-execution-created.json
.orun/refs/latest-execution.json   (status=completed, executionKey=run-001)
```

Final summary block prints Revision / Execution / Path lines per
cli-surface.md §1.1.

## Decisions

1. **AfterStateUpdate hook on RunnerHooks** — chose to add a new hook
   field rather than couple `internal/runner` to
   `internal/executionstate`. Keeps the package boundary clean: the
   runner emits ticks, the CLI binds the bridge.
2. **Synthesize-and-re-resolve on benign miss** — a first attempt
   passed the in-memory plan through to `CreateExecution` directly,
   which failed `statestore: missing input field RevisionID`. The fix
   is to synthesize the revision against the on-disk store, then
   re-call `ResolveRevision` so the returned `PlanRevision` carries
   real `RevisionID`/`TriggerID` ULIDs.
3. **`MirrorModeAuto`** — design.md §11 is explicit: try hardlink, fall
   back to copy on EXDEV. No reason to vary in M5.b.
4. **Revision/Trigger/Execution summary block via
   `printRevisionRunSummary`** — mirrors the M5.a `orun plan` shape so
   tooling that scans for the legacy `components × envs → jobs` line
   keeps working while gaining the new triplet.

## Phase 1 Reservations (NOT wired)

- `--persist-revision` flag — synthesized branches stay
  synthesized-and-persisted via the synthesize-fallback path; no flag
  needed in Phase 1.
- `Reason="rerun"/"retry"/"migration"` — only `"direct-run"` is emitted
  from this path. Re-run / retry / migration paths are M5.c+ surface.

## Risk Notes (carried forward)

- `MirrorModeHardlink` debug-fold decision still deferred to M6.
- `emitFailure` best-effort dropped events — not addressed in M5.b
  (deferred to M5.c stderr/metric integration).
- `internal/executionstate` coverage at exact 90.0 % floor — must not
  regress in M5.c/d.
- Event-sequence retry budget of 32 acceptable for single-writer
  Phase 1; re-evaluate when remote drivers come online.

## Recommended Next Move

M5.b closed. Next per `specs/orun-state-redesign/implementation-plan.md`
§M5: **Task 0019 = M5.c (`orun status / logs / describe / get` rewire)
implementer** from `main` @ post-merge head `59d06f3`.
