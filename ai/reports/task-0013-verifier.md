# Task 0013 — Verifier Report (M4 PR-A)

## Result: PASS

## Scope

Verify Task 0012 / PR #159 (`impl/task-0012-m4-executionstate-pra`) against
the Verifier Standard in `agents/orchestrator.md` and the **M4 "Done when"**
criteria in `specs/orun-state-redesign/implementation-plan.md`. PR-A scope:
`internal/executionstate` model + writer + resolver. PR-B (`bridge.go` /
`MirrorRunnerOutput` / EXDEV fallback) is intentionally deferred to Task 0014.

PR head verified: `8a0c409` (matches the orchestrator-emitted SHA byte-for-byte).
Parent: `2c239d7` "Task 0012: M4 PR-A — internal/executionstate model + writer
+ resolver".

## Checks

| # | Check | Result |
|---|-------|--------|
| 1 | `gh pr checkout 159` | OK; branch `impl/task-0012-m4-executionstate-pra` checked out, head matches `8a0c409`. |
| 2 | `git log --oneline -5` | Confirmed head `8a0c409 docs: add task-0012 implementer report` over `2c239d7 Task 0012: M4 PR-A …`. |
| 3 | `go build ./...` | Exit 0; clean. |
| 4 | `go vet ./...` | Exit 0; clean. |
| 5 | `go test -race -count=1 ./...` | All packages PASS, including `internal/executionstate` 45.9 s with `-race`. No DATA RACE warnings. |
| 6 | `make test-state-redesign` | PASS. Coverage gates measured: `internal/statestore` 95.4 % (≥ 95 %), `internal/revision` 90.4 % (≥ 90 %), `internal/executionstate` **90.0 %** (≥ 90 %, exact floor — gate enforced; verified `make` exits 0). |
| 7 | `go test -race -count=1 -run TestNextExecutionKey_MonotonicityUnderConcurrency -v ./internal/executionstate/...` | PASS in 30.47 s. Property test exists at `property_test.go:22`. |
| 8 | `go test -race -count=1 -run "TestResolveExecution_Branch5_LegacyFallback\|TestResolveExecution_LegacyLatestFallback" -v ./internal/executionstate/...` | Both PASS. Resolver legacy-fallback covered against synthesized `.orun/executions/`. |
| 9 | `go list -deps ./internal/executionstate \| grep -E "cmd/orun\|sourceplane/orun/internal/state$\|internal/runner\|internal/runbundle"` | **Empty.** Leaf-clean constraint satisfied (design.md §5.1). |
| 10 | `grep -n '"\.orun/' $(production .go in internal/executionstate/)` | **Empty.** All `.orun/` literals in production code are inside comments only (`model.go`, `resolver.go` doc comments). Path strings come from `internal/statestore/paths.go` helpers (`ExecutionDir`, `ExecutionDocPath`, `ExecutionIndexEntry`, `EventPath`, `SnapshotPath`). Test fixtures in `resolver_test.go` synthesize legacy layouts under `t.TempDir()` — acceptable per the task's verification step #10. |
| 11 | `bridge.go` / `MirrorRunnerOutput` absence | Confirmed. No `bridge.go` file in package; only one passing reference in `model.go` package-doc comment ("Out of scope until M4 PR-B"). No `// TODO(bridge)` markers leaked. PR-B scope cleanly deferred. |
| 12 | Public surface audit | Exported surface from `model.go` + `writer.go` + `resolver.go`: `APIVersion`, `KindName`, `Reason*`, `Status*`, `ExecutionRun`, `RunnerProfile`, `ExecSummary`, `IsTerminal`, `Config`, `SanitizeExecID`, `NextExecutionKey`, `CreateExecutionInput`, `CreateExecution`, `UpdateSnapshot`, `MarkTerminal`, `ResolveSource`, `ResolveSource*`, `ExecutionRef`, `LegacyRoot`, `ResolveOptions`, `ResolveExecution`. No accidental over-export. |
| 13 | R-002 ExecID shape compat (one-shot test) | Wrote `TestSanitizeExecID_GitHubShape_R002` exercising `SanitizeExecID("gh-12345678901-2-abc123def456abc123def456abc123def456abc1")` and `gh-1-1-deadbeef`; both round-trip identity (sanitized output == input — alphabet `[a-z0-9-]`, no projection needed). Test deleted after the verification spot-check; `OriginalKey` round-trip confirmed compatible with the runbundle GHA shape. |
| 14 | `gh pr view 159 --json statusCheckRollup` | Both required CI checks SUCCESS: `CI / Orun Plan` run `26675724704` job `78626951236` completed 2026-05-30T05:30:17Z; `orun remote-state conformance / Harness dry-run guard` run `26675724720` job `78626951221` completed 2026-05-30T05:29:47Z. Five matrix legs SKIPPED legitimately (empty matrix at M4 PR-A — same shape as #152/#155–#158). `state=OPEN`, `mergeable=MERGEABLE`, `mergeStateStatus=CLEAN`. |
| 15 | `kiox` validate / plan / dry-run | Skipped — repository root has no top-level `intent.yaml` (only `examples/intent.yaml`); the task's verification steps #11–#13 only require running them when `intent.yaml` exists at root. CI's `Orun Plan` and `Harness dry-run guard` jobs cover this surface authoritatively. Documented per the verifier skill's "no top-level intent.yaml" allowance. |

## Code-path adjudication

### data-model.md §5 byte-stable model — match
`ExecutionRun`, `RunnerProfile`, `ExecSummary` field tag order in `model.go`
matches the spec block in `data-model.md` §5 byte-for-byte (lowerCamelCase,
`apiVersion: "orun.io/v1alpha1"`, `kind: "ExecutionRun"`, RFC 3339
`time.Time`, `StartedAt`/`FinishedAt` pointer-typed for omitempty correctness,
status enum literals match `pending|running|completed|failed|cancelled`).

### state-store.md §3 / §6 — four-error taxonomy preserved
Production code in `internal/executionstate/` introduces zero new sentinels.
Every error path wraps an existing `statestore` sentinel (`ErrInvalid`,
`ErrNotFound`, `ErrExists`, `ErrConflict`) via `fmt.Errorf("%w: …", …)`
(grep confirmed: 11 wrap sites in `writer.go`, 0 `errors.New(`, 0 `var Err…`).
Writer step semantics:
- Step 2 (`CreateIfAbsent` claim) on derived key: `ErrExists` retries;
  on caller-supplied key: surfaces verbatim (spec-correct: re-deriving would
  change a user-visible key).
- Step 3 (execution-index) `CreateIfAbsent`: idempotent re-claim returns
  success on `ErrExists`.
- Step 4 (`refs/latest-execution.json`) via `Write` last-write-wins per
  data-model.md §6.2 — matches §6 Atomicity table (multi-object writes are not
  transactional; ref-scan is the consistency fallback).
- Step 5 (`execution-created` event) `CreateIfAbsent` with `ErrExists`-as-OK.
- Step 6 (manifest CAS) delegates to `revision.UpdateLatestExecutionSummary`
  which already implements bounded retry + idempotent-short-circuit.

### CAS + idempotent-short-circuit — first real caller correctness
`revision.UpdateLatestExecutionSummary` (manifest.go:136) is now exercised by
`CreateExecution.finalizeExecution` and by `MarkTerminal` (writer.go:431,
writer.go:531-540). The CAS dance is symmetric: read manifest → compute
`next` with mutated `LatestExecutionKey`/`Status` → bytes-equal short-circuit
→ `CompareAndSwap`; on `ErrConflict` retry within `casRetryBudget`; otherwise
wrap. The race property test
(`TestNextExecutionKey_MonotonicityUnderConcurrency`) and the
`TestMarkTerminal_HappyAndIdempotent` test together exercise both the conflict
retry and idempotent short-circuit branches. No clobber risk under
concurrent terminal transitions on the same `(revisionKey)`.

### Resolver branch ladder + legacy fallback (compat §3 / §4) — match
`ResolveExecution` (`resolver.go:105`) implements:
1. empty arg → `refs/latest-execution.json`,
2. `<revKey>/<execKey>` form → direct doc read,
3. `<execKey>` (with revHint) → revision-scoped doc read,
4. ULID/`exec_…` form → execution-index lookup,
5. legacy `.orun/executions/<arg>/execution.json` direct read,
6. legacy `.orun/executions/` newest-mtime scan,
7. typed `ErrNotFound` on no match.

`TestResolveExecution_Branch5_LegacyFallback` synthesizes a legacy layout
under `t.TempDir()` and confirms branch 5 finds it.
`TestResolveExecution_LegacyLatestFallback` exercises branch 6.
`synthesizeFromLegacy` projects onto `system.migrated` provenance per
compat §4 (`triggerKey = string(triggerctx.SystemMigrated)`,
`Reason = ReasonMigration`, `Status` defaults to `StatusCompleted` when the
legacy doc lacks one).

### No overreach — confirmed
PR diff (12 files, +2593 / −2): adds `internal/executionstate/{model,writer,
resolver,internal}.go` and tests; additive helpers in
`internal/statestore/paths.go` only (`ExecutionsDir`, `ExecutionDocPath`,
`ExecutionIndex*`, `LegacyExecution*`, `EventPath`, `SnapshotPath`); `Makefile`
extended with the executionstate coverage gate. Zero edits to `cmd/orun`,
`internal/state`, `internal/runner`, `internal/runbundle`. `go list -deps`
confirms the leaf-clean invariant.

### R-002 ExecID shape compat
Spot-check via temp test: `SanitizeExecID("gh-{run_id}-{attempt}-{sha}")`
round-trips identity for representative GHA shapes — the input is already
inside the `[a-z0-9-]` alphabet so no projection occurs and `OriginalKey ==
ExecutionKey`. Compatible with `internal/runbundle`'s emitted ExecIDs without
further migration work.

## Issues

None. No verifier fixes were required.

## Risk Notes

- **Adjudication of implementer assumption #1 (manifest required to update
  latest-execution-summary).** Implementer chose to surface `ErrNotFound`
  loudly when `revision.UpdateLatestExecutionSummary` is called against a
  revision that never wrote `manifest.json`. Cross-reference:
  `data-model.md` §4 says `summary.latestExecutionKey` /
  `summary.latestExecutionStatus` "are updated by `executionstate.writer`
  whenever an execution under this revision changes terminal state" —
  this implicitly assumes the manifest exists. `state-store.md` §6 says
  multi-object compound writes are not transactional and "callers order
  writes so that the body lands before the ref". The conservative-on-
  unknowns posture matches the surrounding `revision/writer.go` style
  (also surfaces `ErrNotFound` rather than synthesizing). **Accepted with
  Risk Note; no spec proposal filed.** Carry-forward: M5 CLI rewire must
  ensure `WriteRevision` + `WriteManifest` precede `CreateExecution` for
  the same `revKey`. If M5 needs the option to skip the manifest step,
  the contract should then be normative-pinned in `data-model.md` §4
  via a proposal.

- **Adjudication of implementer assumption #2 (legacy executions default to
  `triggerKey="system.migrated"`, `Reason="migration"`,
  `Status="completed"`).** `compatibility-and-migration.md` §4 prescribes
  the trigger flavor literally as `triggerType: "system"`,
  `triggerName: "system.migrated"`, `source.workingTree: "unknown"`. It
  does NOT explicitly normate the M4-projected fields
  `triggerKey = string(triggerctx.SystemMigrated)` (= `"system.migrated"`),
  `Reason = "migration"` (`ReasonMigration`), or `Status = "completed"`
  default. The implementer's projection is a defensible read of the
  M4 model (compat §4 says "unmarshalled into the new model in memory"
  without specifying every field literal), but not strictly normative.
  **Accepted as a documented fallback convention; no spec proposal filed
  in this PR.** Carry-forward for M5 / Phase 2: when migration command
  (compat §5) lands, those literals should be pinned normatively in compat
  §4 alongside the existing `triggerType`/`triggerName`/`workingTree`
  prescriptions.

- **R-002 ExecID shape compat — no risk.** `gh-{run_id}-{attempt}-{sha}` is
  already alphabet-clean for `SanitizeExecID`; no migration required when
  PR-B wires the bridge.

- **Coverage at the floor.** `internal/executionstate` 90.0 % is the exact
  gate. Future small refactors that delete a covered branch could trip the
  gate; the writer file has the largest uncovered surface (~10 % residual
  in branch-edge paths). M4 PR-B and M5 wiring will add callers and tests;
  the floor should rise organically.

- **Test runtime cost.** The race-mode property test
  (`TestNextExecutionKey_MonotonicityUnderConcurrency`) takes ~30 s under
  `-race`; total `internal/executionstate` race test wall time is ~46 s.
  Acceptable for now; if it grows much more, consider a `-short` skip path.

- **kiox local skip.** No top-level `intent.yaml` at repo root, so
  `kiox … orun validate/plan/run` were not run locally. CI's `Orun Plan`
  and `Harness dry-run guard` exercise the full surface authoritatively
  (both SUCCESS at log level for run IDs `26675724704` and `26675724720`).

## Spec Proposals

None required for PR-A merge. Two carry-forward candidates documented under
Risk Notes (manifest-required-for-summary contract pinning;
legacy-execution literal-defaults normativity in compat §4) — file when
M5 / migration command lands and either contract is operationalized by a
real caller.

## Recommended Next Move

**Emit Task 0014 — M4 PR-B implementer.** Scope: `internal/executionstate/
bridge.go` (`Bridge{Store, LegacyRoot, MirrorMode}` + `MirrorRunnerOutput`
hardlink-with-copy fallback) and the EXDEV fallback test against
synthesized cross-device boundaries. Coverage floor for
`internal/executionstate` rises from 90 % to whatever PR-B adds; the
≥ 90 % gate stays. Task 0015 (M4 PR-B verifier) closes M4 entirely; Task
0016+ enters M5 (CLI rewire).

## PR Number

**#159** — `impl/task-0012-m4-executionstate-pra` — squash-merge target.
