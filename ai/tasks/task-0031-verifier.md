# Task 0031

# Agent
Verifier

# Current Repo Context
Phase 2 Milestone C4 PR-1 (Task 0030 implementer) is open as PR #173 on
branch `task-0030-catalogstore-c4-pr1`. CI is fully green (4/4 SUCCESS,
SKIPPED matrix expansions are normal). Status:
`MERGEABLE` / `mergeStateStatus: CLEAN`.

Implementer report shipped on the PR branch as `reports/task-0030-catalogstore-c4-pr1.md`
(see "Issues to call out" below — task-0030.md mandated `/ai/reports/task-0030-implementer.md`
and the implementer chose a non-canonical location).

Expected delivered surface per Task 0030 acceptance:
- `internal/catalogstore/{paths.go, paths_test.go, writer.go, writer_test.go,
  errors.go, store.go, store_test.go, doc.go}` — no other paths touched.
- `internal/catalogstore` package coverage ≥ 90 %.
- Phase 1 floors held: `internal/statestore` ≥ 95.7, `internal/revision` ≥ 90.3,
  `internal/executionstate` ≥ 90.0.
- Phase 2 floors held: `internal/catalogmodel` ≥ 91.1, `internal/sourcectx` ≥ 91.1,
  `internal/catalogresolve` ≥ 90.9.

Repo health: green. `main` at `75082ca`. No other open PRs. No
proposals owed.

# Objective
Verify PR #173 against Task 0030's acceptance criteria. PASS = merge
the PR via the Verifier Merge Protocol, sync local main, write the
verifier report, leave the worktree clean. FAIL = leave the PR open
with clear, actionable blockers in the report.

# PR Boundary
PR #173 only. No scope expansion. Verification-only fixes (e.g.
moving the implementer report to the canonical path; minor doc
polish strictly required to PASS) MAY be committed to the PR
branch and re-pushed before merge. Anything beyond that becomes
Task 0031.x or a follow-on.

# Read First
- `ai/tasks/task-0030.md` — original implementer prompt; treat its
  Required Outcomes / Acceptance Criteria / Non-Goals as the
  contract.
- `reports/task-0030-catalogstore-c4-pr1.md` (on the PR branch) —
  implementer self-report.
- `agents/orchestrator.md` — Verifier Standard, Verifier Merge
  Protocol.
- `specs/orun-component-catalog/catalog-store.md` — §1 public
  surface, §2 path helpers, §3 write order (steps A & B only;
  C & D belong to PR-2 — must remain stubbed), §5 atomicity,
  §6 error taxonomy.
- `specs/orun-component-catalog/identity-and-keys.md` — `srcKey` /
  `catKey` / component-name sanitization rules used by paths.
- `specs/orun-component-catalog/data-model.md` — input types
  (`SourceSnapshot`, `CatalogSnapshot`, `ComponentManifest`,
  `CatalogGraphs`, `CatalogLocalIndexes`).

Reference Only:
- `internal/statestore/store.go` — confirm `errors.Is` chain to
  `statestore.ErrExists` is preserved by the new mismatch sentinels.
- `internal/catalogmodel` canonical encoder — confirm `writer.go`
  uses it (or `PrettyEncode` for body writes per the implementer's
  stated decision, with `CanonicalEncode` reserved for hashing).

# Required Outcomes
1. PR-boundary audit: only the eight files under
   `internal/catalogstore/` (plus the implementer report) touched.
   No edits to `internal/statestore/`, `internal/catalogresolve/`,
   `internal/catalogmodel/`, `internal/sourcectx/`,
   `internal/triggerctx/`, `internal/revision/`,
   `internal/executionstate/`, `cmd/orun/`. No `go.mod` / `go.sum`
   churn beyond `main`. Forbidden files (`refs.go`, `indexes.go`,
   `resolver.go`) ABSENT.
2. Write-order audit: `WriteCatalogSnapshot` issues B.1 → B.2 → B.3 → B.4
   in code AND a spy-based test asserts that exact order regardless
   of input map order. Graph write order is exactly
   `dependencies, systems, apis, resources, owners`.
3. Pre-flight inconsistency guard (`ErrInputsInconsistent`) fires
   BEFORE any write when (a) `cat.source.sourceSnapshotKey` ≠
   `src.sourceSnapshotKey`, or (b) any manifest's
   `source.sourceSnapshotKey` / `source.catalogSnapshotKey` doesn't
   match the (src, cat) tuple. Verify by reading the test, NOT by
   trusting the report.
4. Idempotence: byte-identical re-write through `CreateIfAbsent`'s
   `ErrExists` is a success path for source / catalog / manifest;
   different-body re-write returns the typed `Err*Mismatch`. Each
   typed sentinel preserves `errors.Is` against
   `statestore.ErrExists` (double-wrap pattern claimed in the
   report — confirm in code).
5. Stub policy: `WriteRefs`, `WriteGlobalIndexes`,
   `AppendComponentEvent`, and every `Resolver` method return a
   typed `ErrNotImplemented`. A test pins this so a future
   accidental nil-return is a build/test break.
6. Step B.4 (local indexes) uses plain `Write` (overwrite-OK), NOT
   `CreateIfAbsent` / `CompareAndSwap`. Spec is explicit that local
   indexes are rebuildable. If the implementer "hardened" them with
   CAS, that is FAIL.
7. Path layer: every helper from `catalog-store.md` §2 is present,
   no raw `path.Join` of caller-supplied keys without a `Validate*`
   call first, no `os` / `io/ioutil` / `path/filepath` imports
   anywhere in `internal/catalogstore/`. Every helper is either
   panic-free (returns `(string, error)`) or panics only on
   programmer errors that `Validate*` already screened. Pick what
   the code actually does and confirm consistency.
8. `ValidateSegment` rejects: empty, `.`, `..`, `/`, `\`, space,
   tab, uppercase, colon, oversize. `ValidateRefName` allows only
   `latest` / `current` / `main`. `ValidateGraphKind` matches
   `CatalogGraphKinds()` exactly.
9. Coverage: `go test ./internal/catalogstore/... -race -count=1 -cover`
   reports ≥ 90 % statement coverage. Phase 1 floors held byte-for-byte
   (`internal/statestore` ≥ 95.7, `internal/revision` ≥ 90.3,
   `internal/executionstate` ≥ 90.0). Phase 2 floors held
   (`internal/catalogmodel` ≥ 91.1, `internal/sourcectx` ≥ 91.1,
   `internal/catalogresolve` ≥ 90.9). Capture exact percentages.
10. `go vet ./...`, `go build ./...`, `go test ./... -race -count=1`,
    `make verify-generated` all clean.
11. Implementer report location: task-0030.md required
    `/ai/reports/task-0030-implementer.md`; PR landed
    `reports/task-0030-catalogstore-c4-pr1.md` instead. EITHER
    (a) commit a copy/move at `ai/reports/task-0030-implementer.md`
    on the PR branch before merge (preferred — keeps the canonical
    path discoverable), OR (b) document the deviation in the verifier
    report and let it ride. Verifier's choice; do NOT fail the PR over
    location alone.
12. CI: `gh pr view 173 --json statusCheckRollup` shows all required
    checks SUCCESS (already green at scope time; re-confirm at merge
    time).

# Non-Goals
- No code changes inside `internal/catalogstore/` (the implementer's
  scope is locked).
- No PR-2 work (`refs.go`, `indexes.go`, `AppendComponentEvent`,
  global indexes).
- No PR-3 work (`resolver.go`, fallback chain, `RebuildIndexes`).
- No spec edits unless a behavioral drift is identified — in which
  case write `/ai/proposals/task-0031-spec-update.md` and FAIL the
  verification with that proposal as the blocker.
- No coverage rebalancing across other packages.

# Constraints
- No edits outside `internal/catalogstore/` and (optionally) the
  implementer report path.
- Verifier-only commits MUST be on the PR branch and pushed; CI
  must re-pass before merge.
- Use the canonical Verifier Merge Protocol from
  `agents/orchestrator.md` lines 644–657.
- Never merge while CI is red, MergeStateStatus ≠ CLEAN, or any
  Required Outcome is unmet.

# Integration Notes
- The double-wrap pattern (typed sentinel wrapping
  `statestore.ErrExists`) the implementer adopted is load-bearing
  for downstream consumers — confirm both `errors.Is(err, ErrSourceMismatch)`
  AND `errors.Is(err, statestore.ErrExists)` succeed via the actual
  test, not just the report's claim.
- The `Writer` / `Resolver` / `Store` interfaces frozen here are
  the contract PR-2 / PR-3 must fill without widening. Read
  `store.go` and confirm the compile-time interface assertions
  named in the report exist.
- `PrettyEncode` vs `CanonicalEncode` for body writes is an
  implementer call documented in the report. Confirm the chosen
  encoder is consistent across all body writes and that hashing
  is NOT happening in this PR (PR-1 is write-only).

# Acceptance Criteria
✅ All 12 Required Outcomes met.
✅ `go vet ./...` clean.
✅ `go build ./...` clean.
✅ `go test ./internal/catalogstore/... -race -count=1 -cover` ≥ 90 %.
✅ `go test ./... -race -count=1` green.
✅ `make verify-generated` clean.
✅ Coverage floors held byte-for-byte (record exact percentages).
✅ Spy-based call-order assertion present and exercising B.1→B.2→B.3→B.4.
✅ `ErrInputsInconsistent` covered by test for at least the three
   mismatch shapes (cat↔src, manifest↔src, manifest↔cat).
✅ Stub-pin test for every `ErrNotImplemented` path.
✅ No raw FS imports in `internal/catalogstore/`.
✅ PR #173 `mergeStateStatus: CLEAN`, `statusCheckRollup` all SUCCESS.
✅ Implementer report location decision recorded (move or accept).
✅ Verifier report at `ai/reports/task-0031-verifier.md`.
✅ Local kiox guards (when applicable) recorded — record no-op if no
   plan jobs.

# Verification
Run locally and capture exact output:

    cd /Users/irinelinson/sourceplane/orun
    git fetch origin
    git checkout task-0030-catalogstore-c4-pr1
    git pull --ff-only

    go mod tidy
    go vet ./...
    go build ./...
    go test ./internal/catalogstore/... -race -count=1 -cover
    go test ./... -race -count=1
    make verify-generated

    # Coverage floors — capture exact %
    go test ./internal/statestore/... -count=1 -cover
    go test ./internal/revision/... -count=1 -cover
    go test ./internal/executionstate/... -count=1 -cover
    go test ./internal/catalogmodel/... -count=1 -cover
    go test ./internal/sourcectx/... -count=1 -cover
    go test ./internal/catalogresolve/... -count=1 -cover

    # Static guards
    grep -n '"os"\|"io/ioutil"\|"path/filepath"' internal/catalogstore/ -r || echo "no raw FS imports — OK"

    # PR / CI evidence
    gh pr view 173 --json title,state,mergeable,mergeStateStatus,statusCheckRollup,headRefName,url
    gh pr diff 173 --name-only

Then run kiox guards:

    /Users/irinelinson/.local/bin/kiox -- orun validate --intent intent.yaml
    /Users/irinelinson/.local/bin/kiox -- orun plan --changed --intent intent.yaml --output plan.json
    /Users/irinelinson/.local/bin/kiox -- orun run --plan plan.json --dry-run --runner github-actions

Record no-op when the plan has no jobs.

Code-path inspections (read the files, not just the tests):
- `internal/catalogstore/writer.go::WriteCatalogSnapshot` — confirm
  pre-flight runs before any `state.CreateIfAbsent` / `state.Write`,
  and that the loop iterating graphs uses `CatalogGraphKinds()` (or
  an equivalent fixed slice) rather than a map range.
- `internal/catalogstore/store.go` — confirm `Resolver` methods and
  the three deferred `Writer` methods are wired to `ErrNotImplemented`
  and that compile-time `var _ Writer = (*store)(nil)` (or equivalent)
  interface assertions are present.
- `internal/catalogstore/errors.go` — confirm mismatch sentinels
  expose both their own typed identity and `statestore.ErrExists` via
  `errors.Is` (e.g. `Unwrap()` chain or multi-target `Is`).

# PR Creation Requirement
Implementer already created PR #173. Verifier does not create a new
PR. If a verification-only fix is required, commit it to the PR
branch, push, wait for CI to re-go-green, then merge.

# Verifier Merge Protocol (from agents/orchestrator.md)
- If PASS:
  1. `gh pr merge 173 --squash --delete-branch` (admin if required).
  2. `git checkout main && git pull --ff-only origin main`.
  3. `git status --short` → must be clean. Resolve any verifier-side
     leftover changes before declaring done.
  4. Commit the verifier report + state-file updates to `main`
     directly (or via a tiny follow-up PR if the repo requires it —
     historically the loop commits these to main per prior cycles).
- If FAIL: leave PR #173 open with all blockers spelled out.
  Do NOT merge under any circumstance with red CI, dirty
  `mergeStateStatus`, or any unresolved Required Outcome.

# When Done Report
Write `/ai/reports/task-0031-verifier.md` with:

- `Result: PASS|FAIL`
- `Checks` — every command run, exact output (or summarised when
  long; never claim a check passed without showing evidence)
- `CI Log Review` — `gh pr view 173` JSON snippet, run IDs cited
- `Issues` — empty if PASS; specific blockers if FAIL
- `Code Path Inspection` — short notes on the three reads above
- `Coverage Evidence` — exact % for `internal/catalogstore` and
  every floor-gated package
- `Risk Notes` — residual risk after merge (if any)
- `Spec Proposals` — links only; expected `_none_`
- `Recommended Next Move` — on PASS, point to Task 0032 = C4 PR-2
  implementer (`refs.go` + `indexes.go` + `AppendComponentEvent`).
  On FAIL, name the smallest remediation that closes the gap.
- `PR #173 Merge Status` — merged SHA + branch deletion confirmation,
  or "left open, blockers above".
