# Task 0013 — Verifier (M4 PR-A)

Agent: Verifier

## Current Repo Context

- Active spec: `specs/orun-state-redesign/` (Phase 1, local-only). Active
  milestone: **M4 — `internal/executionstate` + runner bridge.** PR-A
  (model + writer + resolver) is the first half; PR-B (`bridge.go` /
  `MirrorRunnerOutput` / EXDEV fallback) is intentionally deferred to Task
  0014.
- M3 closed at PR #158 (squash-merge `bfc2ae6`, 2026-05-30, verified PASS via
  Task 0011). M4 begins with **Task 0012 — M4 PR-A implementer** (PR #159).
- Implementer Task 0012 shipped on branch `impl/task-0012-m4-executionstate-pra`.
  Implementer report: `ai/reports/task-0012-implementer.md`. Reported PR
  number: **#159**.
- PR **#159** state at orchestrator emission time:
  - `state=OPEN`, `mergeable=MERGEABLE`, `mergeStateStatus=CLEAN`
  - head SHA `8a0c409ef885f5f354c645eea981723b86d0b0e2` (commit `8a0c409`
    "docs: add task-0012 implementer report"; parent `2c239d7`
    "Task 0012: M4 PR-A — internal/executionstate model + writer + resolver")
  - Required CI both SUCCESS at log level:
    - `CI / Orun Plan` — run `26675724704`, job `78626951236`, completed
      2026-05-30T05:30:17Z
    - `orun remote-state conformance / Harness dry-run guard` — run
      `26675724720`, job `78626951221`, completed 2026-05-30T05:29:47Z
  - Five matrix legs SKIPPED legitimately (empty matrix at M4 PR-A — same
    shape as #152/#155/#156/#157/#158).
  - Diff stat (12 files, +2593 / -2):
    - new: `internal/executionstate/{model,writer,resolver,internal}.go`
      and tests `internal/executionstate/{model,writer,resolver,property,coverage_extra}_test.go`
    - edits: `internal/statestore/paths.go` (additive helpers only —
      `ExecutionsDir`, `ExecutionDocPath`, `ExecutionIndex*`,
      `LegacyExecution*`, `EventPath`, `SnapshotPath`); `Makefile`
      (extends `test-state-redesign` with `./internal/executionstate/...` and
      ≥90 % coverage gate)
    - artifacts: `ai/reports/task-0012-implementer.md`
- Implementer report `ai/reports/task-0012-implementer.md` claims:
  - **Coverage:** `internal/executionstate` **90.0 %** (gate ≥ 90 %, exact
    floor — this is the closest M-series PR has come to the gate; verify
    locally and confirm the `make test-state-redesign` gate enforces it),
    `internal/revision` **90.4 %** (M3 floor preserved), `internal/statestore`
    **95.4 %** (M2 floor preserved).
  - **Property test:** `NextExecutionKey` monotonicity under N=100
    concurrent `CreateExecution` calls (per `implementation-plan.md` §M4
    "Done when").
  - **Resolver:** seven-branch ladder + legacy `.orun/executions/` fallback
    (`compatibility-and-migration.md` §3 / §4).
  - **Writer:** wires the FIRST real in-tree caller of
    `revision.UpdateLatestExecutionSummary` (shipped under PR #158 with zero
    callers) — this is the first load-bearing test of the CAS +
    idempotent-short-circuit dance.
  - **Assumption recorded:** "M3 callers writing a revision but not a
    manifest is treated as an error (loud surfacing) rather than silently
    no-op'ing the latest-execution-summary update — matches the
    conservative-on-unknowns posture of `revision/writer.go`." Verifier owns
    adjudication: is this consistent with `data-model.md` §4
    (`RevisionManifest`) + `state-store.md` §3 / §6, or does it warrant a
    spec proposal pinning the contract?
  - **Assumption recorded:** "Legacy executions with no `revisionKey` carry
    `triggerKey = \"system.migrated\"` and `Reason = \"migration\"` per
    compat §4; status defaults to `\"completed\"` when absent." Verifier
    owns adjudication: confirm `compatibility-and-migration.md` §4 actually
    prescribes these literal values, or treat as a documented fallback
    convention worth proposal-ing.
  - **No spec proposals filed.** No new error sentinels (reuses
    `statestore` four-error taxonomy via `fmt.Errorf("%w: …", …)`).
  - `kiox` was unavailable in the implementer sandbox; `validate` / `plan` /
    `run --dry-run` were skipped (`kiox-skip`). Verifier MUST run these
    locally per the Verifier Merge Protocol if `kiox` is on disk
    (`/Users/irinelinson/.local/bin/kiox`).
- After PR #159 merges and is verified, M4 PR-A closes. Task 0014 (M4 PR-B
  implementer — `bridge.go` + `MirrorRunnerOutput` + EXDEV fallback test)
  follows; Task 0015 (M4 PR-B verifier) closes M4 entirely.

## Objective

Validate Task 0012 / PR #159 against the Verifier Standard in
`agents/orchestrator.md` and the **M4 "Done when"** criteria in
`specs/orun-state-redesign/implementation-plan.md`. Specifically confirm:

1. The `ExecutionRun` / `RunnerProfile` / `ExecSummary` typed model matches
   `data-model.md` §5 byte-for-byte (field names lowerCamelCase,
   `apiVersion: "orun.io/v1alpha1"`, `kind: "ExecutionRun"`, RFC 3339 times,
   `Status` enum literals).
2. `NextExecutionKey`, `SanitizeExecID`, `CreateExecution`,
   `UpdateSnapshot`, `MarkTerminal` honour the `state-store.md` four-error
   taxonomy (no new sentinels) and the §3 / §6 atomicity guarantees.
3. The first real caller of `revision.UpdateLatestExecutionSummary` exercises
   CAS retry + idempotent-short-circuit correctly (no clobbers under
   concurrent terminal transitions on the same `(revisionKey)`).
4. The seven-branch `ResolveExecution` resolver covers the spec ladder and
   the legacy `.orun/executions/` fallback (compat §3 / §4) — write a
   synthesized old layout under a temp dir and confirm the resolver finds
   it.
5. **No overreach** into `cmd/orun`, `internal/state`, `internal/runner`,
   `internal/runbundle`. Run
   `go list -deps ./internal/executionstate | grep -E "cmd/orun|^orun/internal/state$|internal/runner|internal/runbundle"`
   and confirm the result is empty.
6. **No string path literals outside `internal/statestore/paths.go`** —
   `rg -n '\.orun/' internal/executionstate/` should only return literals
   that come from helpers (or test fixtures), not hardcoded execution
   paths.
7. `bridge.go` / `MirrorRunnerOutput` is genuinely absent from the PR (PR-B
   scope), and the public surface listed in the implementer report's
   "Next-Task Dependencies" section is the actual exported surface — no
   accidental over-export.

On PASS, merge PR #159 per the Verifier Merge Protocol, fast-forward `main`,
and leave the working tree clean.

## PR Boundary

- Verification only. Verifier may commit to the PR branch:
  - `ai/reports/task-0013-verifier.md` (this report).
  - Optionally `ai/proposals/task-0012-spec-update.md` if spec amendment is
    required (e.g. the "manifest required to update latest-execution-summary"
    contract needs to be normative in `data-model.md` §4 or
    `state-store.md` §6; or the legacy-execution `triggerKey =
    "system.migrated"` / `Reason = "migration"` defaults need to be
    normative in `compatibility-and-migration.md` §4).
  - Any tiny verification-only fix that is strictly necessary for
    mergeability (typo, stray TODO removal). Anything beyond that → FAIL
    with explicit blockers; do not edit production code.
- Out of scope: M4 PR-B (`bridge.go` / `MirrorRunnerOutput` / EXDEV
  fallback), M5 work (CLI rewire), production-caller wiring beyond what PR-A
  already lit up.

## Read First

1. `agents/orchestrator.md` — Verifier Standard + Verifier Merge Protocol.
2. `specs/orun-state-redesign/README.md` — index + read order.
3. `specs/orun-state-redesign/implementation-plan.md` — Milestone **M4**
   goal, suggested PR scope, "Done when" checklist (≥ 90 % coverage,
   `NextExecutionKey` monotonicity property under N=100 concurrent
   `CreateExecution`, resolver legacy-fallback test against synthesized
   `.orun/executions/`).
4. `specs/orun-state-redesign/data-model.md` — §5 (`ExecutionRun`),
   §5.1 (`snapshot.latest.json`), §6 (Refs — confirm
   `refs/latest-execution.json` is left to PR-B / M5 if PR-A doesn't write
   it), §7 (Indexes — `indexes/executions/<execKey>.json`), §9 (Event
   stream — confirm `execution-created` is wired or a documented PR-B
   gap).
5. `specs/orun-state-redesign/state-store.md` — §1 (frozen interface), §3
   (`CreateIfAbsent` / `CompareAndSwap` / `Write` semantics + four-error
   taxonomy), §6 (CAS retry / crash-recovery invariants).
6. `specs/orun-state-redesign/design.md` — §5.1 (Package boundaries — leaf
   constraint), §9 (Correctness properties).
7. `specs/orun-state-redesign/compatibility-and-migration.md` — §3
   (resolution chain), §4 (Reader fallback for legacy
   `.orun/executions/`).
8. `ai/reports/task-0012-implementer.md` — implementer claims to validate.
9. `ai/context/open-risks.md` — R-002 (ExecID shape preservation:
   `gh-{run_id}-{attempt}-{sha}` from `internal/runbundle` must remain
   compatible with `SanitizeExecID` output).

## Required Outcomes

- PR #159 either merged on PASS (squash, branch deleted, local `main`
  fast-forwarded, `git status --short` clean) or left OPEN with a clear
  FAIL and explicit blockers.
- `ai/reports/task-0013-verifier.md` written, committed to PR #159 branch,
  pushed; CI green again before merge.
- Spec adjudication for the two implementer assumptions explicitly
  recorded (accept inline with Risk Notes OR file
  `ai/proposals/task-0012-spec-update.md`).
- ExecID shape compat (R-002) cross-checked: `SanitizeExecID` must accept
  `gh-{run_id}-{attempt}-{sha}` as a valid `OriginalKey` round-trip.

## Non-Goals

- M4 PR-B implementation (Task 0014 will own that).
- M5 CLI rewire.
- Re-verifying M0/M1/M2/M3 surface.
- Refactoring `internal/statestore` or `internal/revision` beyond
  preservation of their coverage floors.

## Constraints

- Reuse the `statestore` four-error taxonomy. Reject any new sentinel in
  `internal/executionstate/` not already justified by the spec.
- All execution-path strings must come from
  `internal/statestore/paths.go`; reject hardcoded `.orun/executions/...`
  literals in production code.
- Coverage floors enforced by `make test-state-redesign`:
  - `internal/executionstate` ≥ 90 %
  - `internal/revision` ≥ 90 %
  - `internal/statestore` ≥ 95 %
- No edits to `cmd/orun`, `internal/state`, `internal/runner`,
  `internal/runbundle`.
- Verifier does NOT modify production code; FAIL if production fixes are
  required.

## Integration Notes

- M4 PR-A wires the first real caller of
  `revision.UpdateLatestExecutionSummary`. The existing `internal/revision`
  CAS dance must hold under the new caller. Re-run the property test from
  M3 PR-B (if any) to confirm no regression.
- ExecID shape preservation (R-002): `SanitizeExecID` must round-trip the
  GitHub Actions `gh-{run_id}-{attempt}-{sha}` shape produced by
  `internal/runbundle`. Spot-check by feeding a representative literal
  through `SanitizeExecID` and `ResolveExecution` in a one-off test or
  inline check during verification.
- The deferred `bridge.go` (PR-B) does not exist yet; confirm no lingering
  `// TODO(bridge)` references in PR-A that should have been excised, and
  confirm the resolver's legacy fallback does not pre-empt bridge logic.

## Acceptance Criteria

- PR #159 builds, vets, tests, and races clean: `go build ./...`,
  `go vet ./...`, `go test -race -count=1 ./...`, `make test-state-redesign`
  all exit 0.
- `go list -deps ./internal/executionstate | grep -E "cmd/orun|internal/state$|internal/runner|internal/runbundle"`
  is empty.
- Coverage gates above met locally and in CI.
- `NextExecutionKey` property test exists, runs, and passes under
  `-race -count=1`.
- Resolver legacy-fallback test against a synthesized `.orun/executions/`
  exists and passes.
- Required CI checks `CI / Orun Plan` and
  `orun remote-state conformance / Harness dry-run guard` SUCCESS at log
  level (use
  `gh run view <run_id> --log` to confirm real plan output, not just
  status summaries).
- ExecID shape (R-002) compat confirmed.
- Two implementer assumptions adjudicated (accept-and-document OR
  proposal-filed).
- Verifier report committed and pushed; PR CI green; squash-merged into
  `main`; branch deleted; local `main` fast-forwarded; `git status --short`
  clean.

## Verification

Run, in order, from the PR branch checked out locally:

1. `gh pr checkout 159`
2. `git log --oneline -5` (confirm head matches `8a0c409`)
3. `go build ./...`
4. `go vet ./...`
5. `go test -race -count=1 ./...`
6. `make test-state-redesign` (must pass coverage gates)
7. `go test -race -count=1 -run NextExecutionKey ./internal/executionstate/...`
   (confirm property test name; adjust to match the actual test name)
8. `go test -race -count=1 -run Resolver.*Legacy ./internal/executionstate/...`
9. `go list -deps ./internal/executionstate | grep -E "cmd/orun|internal/state$|internal/runner|internal/runbundle"`
   → must be empty
10. `rg -n '"\.orun/' internal/executionstate/` → confirm zero hardcoded
    path literals in production `.go` files (test fixtures may use
    `t.TempDir()`-relative literals)
11. `/Users/irinelinson/.local/bin/kiox -- orun validate --intent intent.yaml`
    when `intent.yaml` exists
12. `/Users/irinelinson/.local/bin/kiox -- orun plan --changed --intent intent.yaml --output plan.json`
13. `/Users/irinelinson/.local/bin/kiox -- orun run --plan plan.json --dry-run --runner github-actions`
    (record no-op result if plan has no jobs)
14. `gh pr view 159 --json statusCheckRollup` — confirm both required
    checks SUCCESS at log level (use `gh run view --log`).

## PR Creation Requirement

Verifier commits `ai/reports/task-0013-verifier.md` (and optionally a
proposal file) to the existing PR #159 branch
(`impl/task-0012-m4-executionstate-pra`), pushes, waits for CI to re-green,
then squash-merges per Verifier Merge Protocol. After merge: `git checkout
main`, `git pull --ff-only origin main`, confirm `git status --short` is
clean.

## When Done Report

Write `ai/reports/task-0013-verifier.md` with the Verifier Standard
sections:

- Result: PASS|FAIL
- Checks (commands + results, including CI run IDs)
- Issues (each blocker enumerated; empty list on PASS)
- Risk Notes (especially the two adjudicated implementer assumptions and
  R-002 ExecID shape compat outcome)
- Spec Proposals (link, one-line reason; or "None")
- Recommended Next Move (on PASS: emit Task 0014 = M4 PR-B implementer;
  on FAIL: enumerate fixes the implementer must land in PR #159)
