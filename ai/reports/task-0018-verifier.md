# Task 0018 — Verifier Report

## Result: PASS

## Scope

PR #162 — `impl/task-0018-m5b-orun-run-rewire` → `main`.

M5.b slice: rewire `orun run` to resolve `PlanRevision` via
`internal/revision.ResolveRevision`, synthesize a fresh `system.manual`
revision when the resolver returns a benign miss, persist an
`ExecutionRun` via `internal/executionstate.CreateExecution`, mirror
runner artifacts into the new layout via `Bridge.MirrorRunnerOutput` on
every tick, mark the execution terminal with summary counts on runner
return, add `--revision <key>` flag (cli-surface.md §2.3), and pin the
`bridge-mirror-failed` event payload schema in `data-model.md` §9.1
before any second consumer.

## Checks

| Check | Result |
|---|---|
| Scope discipline (only `cmd/orun/command_run.go`, new `cmd/orun/command_run_revision*.go`, `internal/runner/runner.go`, `specs/orun-state-redesign/data-model.md`, plus tests + ai/ tracking) | PASS |
| `RunnerHooks.AfterStateUpdate` is the ONLY runner edit; no `internal/state`, `internal/runbundle`, or `internal/executor` touched | PASS |
| `go build ./...` | PASS |
| `go vet ./...` | PASS |
| `go test -race ./...` (full module) | PASS |
| `make test-state-redesign` coverage gates: statestore ≥95% (95.7%), revision ≥90% (90.4%), executionstate ≥90% (90.0%, exact floor held) | PASS |
| E2E smoke against `examples/intent.yaml`: fresh `orun run --component web-console-pages --env preview` lays down `revisions/<key>/executions/run-001/{execution,state,metadata}.json` + `events/00000000000000000001-execution-created.json` + `refs/latest-execution.json` | PASS |
| Final `execution.json` carries `status="completed"`, `reason="direct-run"`, `summary={total:1,completed:1,failed:0,running:0,pending:0}`, populated `revisionId`/`triggerId` ULIDs, `originalKey=<legacy-execID>` | PASS |
| Bridge mirrors `state.json` + `metadata.json` from `.orun/executions/<legacyExecID>/` into `revisions/<key>/executions/run-001/` (no `bridge-mirror-failed` events emitted on same-FS smoke) | PASS |
| Resolver branches 1–7 covered upstream by `internal/revision/resolver_test.go::TestResolveRevision_Branch1..7` (re-confirmed green) | PASS |
| CLI-layer tests cover synth round-trip, happy-path setup+finalize, failed-runner path, `--revision` short-circuit, `--exec-id` plumbing, nil-rx no-op, absolute store-root invariant, defensive nil-plan / empty-execID rejection | PASS |
| `data-model.md` §9.1 `bridge-mirror-failed` schema field table matches `internal/executionstate.bridgeMirrorFailedPayload` exactly (kind, at, payload.{executionKey, revisionKey, legacyExecId, artifact, stage, mode, error}) | PASS |
| `--dry-run` and remote-state branches NOT wired through revision-first path (correct: bridge has nothing to mirror in dry-run; remote runs own their lifecycle via the backend) | PASS |
| `--persist-revision` Phase 1 reservation honoured — flag NOT added; synthesize-fallback covers the gap | PASS |
| `Reason` field always `"direct-run"` from this path (rerun/retry/migration deferred to M5.c+) | PASS |
| Required PR CI checks: `CI / Orun Plan` PASS (45s), `Harness dry-run guard` PASS (12s); 5 matrix legs SKIPPED (empty matrix at M5.b, same shape as M5.a #161) | PASS |
| `gh pr view 162 --json mergeable` → `MERGEABLE` / `CLEAN` at merge time | PASS |

## Issues

None. No verifier fixes required.

## Risk Notes (carried forward, updated)

- ~~`bridge-mirror-failed` payload schema un-pinned in data-model.md §9~~ — **CLOSED in M5.b** (§9.1 added).
- `MirrorRunnerOutput` now has its first production caller (M5.b CLI). Resolver legacy-fallback is still the only convergence path for legacy on-disk state.
- `internal/executionstate` coverage at exact 90.0% floor — must not regress in M5.c/d.
- `MirrorModeHardlink` debug-fold decision deferred to M6.
- `emitFailure` best-effort dropped events still un-instrumented — deferred to M5.c stderr/metric integration.
- Event-sequence retry budget of 32 acceptable for single-writer Phase 1; re-evaluate when remote drivers come online.
- `RunnerHooks.AfterStateUpdate` is fired synchronously inside `updateState` (between `SaveState` and the next runner step). Bridge mirror runs on the runner goroutine; on slow filesystems this could measurably extend per-tick wall time. M5.c may want to move the mirror to a buffered channel + dedicated goroutine if real workloads regress.

## Spec Proposals

None required.

## Recommended Next Move

M5.b closed. Next per `specs/orun-state-redesign/implementation-plan.md`
§M5: **Task 0019 = M5.c (`orun status / logs / describe / get` rewire)
implementer** from `main` @ post-merge head `59d06f3`.

## PR Number

**#162** — https://github.com/sourceplane/orun/pull/162
Squash-merged 2026-05-30T13:42:02Z as `59d06f3`.
