# Task 0021 — Verifier report (single-pass closure)

## Verdict: PASS

PR **#165** squash-merged to `main` as `ad3656e` on 2026-05-30T19:17:23Z.
Branch `impl/task-0021-m6-e2e-and-property-gates` deleted.

## What was verified

Closes M6 (end-to-end + property gates) for orun-state-redesign. The
PR is test-only / coverage-gate work; no production source files
under `internal/` or `cmd/` were modified outside of new `_test.go`
files. Confirmed by `git diff --stat ad3656e^..ad3656e`:

```
 Makefile                                            |  12 +-
 ai/state.json                                       |  (pre-task scope)
 ai/context/current.md                               |  (pre-task scope)
 ai/context/task-ledger.md                           |  (pre-task scope)
 ai/tasks/task-0021.md                               |  (pre-task scope)
 ai/tasks/task-0021-report.md                        |  NEW
 cmd/orun/state_e2e_test.go                          |  NEW (~340 LOC)
 internal/revision/keys_property_test.go             |  NEW (~190 LOC)
 internal/revision/m6_coverage_test.go               |  NEW (~160 LOC)
```

## CI on PR head

* `Orun Plan` — **pass** (53 s).
* `Harness dry-run guard` — **pass** (15 s).
* Matrix legs (`Compile plan`, `Env fanout`, `Run`, `Verify remote
  status and logs`, `${{ matrix.component }}/${{ matrix.env }}`) —
  **skipped legitimately** (empty matrix; test-only change touches no
  components or envs).

## Post-merge gates on main (ad3656e)

```
go test ./... -race -count=1 -timeout 600s    # all green
make test-state-redesign                       # all 4 gates pass under -race
```

Coverage measured under `-race` post-merge on main:

| Package                  | Floor  | Measured |
|--------------------------|--------|----------|
| `internal/statestore`    | ≥ 95.0 | 95.7 %   |
| `internal/revision`      | ≥ 90.0 | 90.3 %   |
| `internal/executionstate`| ≥ 90.0 | 90.0 %   |

The `internal/revision` floor had been silently breached at 84.9 % on
main tip prior to M6 (the gate ran without `-race` and depended on a
prior round of unit-test additions that never materialised). M6
restores the documented ≥ 90 % floor by adding the M6-coverage unit
tests **without lowering any threshold** — guardrail honoured.

## Done-when criteria (implementation-plan.md §M6)

- [x] End-to-end walk of the revision-first pipeline covering all 14
      sub-steps of test-plan.md §4 (plan synthesis → on-disk revision
      documents → latest refs → legacy plans/ mirror → execution
      setup + finalize → execution.json terminal status → indexes →
      read-side resolvers → describe revision latest → get plans →
      state migrate dry-run + sha256 byte-equal idempotence).
- [x] Property gates for revision-key derivation per test-plan.md §3.2
      (determinism, distinctness, ResolveCollision suffix
      contiguity).
- [x] Coverage gates wired into `make test-state-redesign` under
      `-race`; all four gates green.
- [x] TestStateE2E stable across `-count=3 -race` (verified during
      implementer pass; CI runs `-count=1 -race` and is green).
- [x] No production code changed; no spec edits; no coverage thresholds
      lowered.

## Carry-forward (deferred from M6 per task-0021 §non-goals)

- MirrorModeHardlink debug-fold decision — defer to a Phase 1
  post-M6 housekeeping task; the M6 E2E provides evidence the mirror
  works under hardlink+fallback but no behavioural change observed
  that would justify removing the knob.
- RunnerHooks.AfterStateUpdate async-mirror evaluation — defer; the
  synchronous chain is stable under -race + property tests.
- `--persist-revision` flag wiring — defer to Phase 2; `state migrate`
  persists unconditionally and direct-run always persists.
- Option B trigger-name resolver branch — defer; covered by the
  proposal at `ai/proposals/task-0019-spec-update.md`.

## Closure

Phase 1 (Local) of orun-state-redesign is **structurally complete**
with M6 merged. All six milestones (M0–M6) closed on main:

| M  | PR    | Main commit |
|----|-------|-------------|
| M0 | #152  | `4ea1980`   |
| M1 | #153  | `db342dd`   |
| M2 | #156  | `cd8b3e8`   |
| M3 | #158  | `bfc2ae6`   |
| M4 | #159 / #160 | `ed48633` / `d51e828` |
| M5 | #161–#164 | `7a9c494` … `17ef788` |
| M6 | #165  | `ad3656e`   |

Spec lifecycle for `specs/orun-state-redesign` Phase 1 is closed.
Next active focus to be re-scoped in a follow-on session.
