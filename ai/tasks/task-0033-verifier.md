# Task 0033

# Agent
Verifier

# Current Repo Context
Phase 2 Milestone C4 PR-2 (Task 0032 implementer) is open as PR #174 on
branch `task-0032-catalogstore-c4-pr2`. CI is fully green (3/3 SUCCESS,
SKIPPED matrix expansions are normal). Status:
`MERGEABLE` / `mergeStateStatus: CLEAN`.

PR replaces the three deferred `ErrNotImplemented` Writer stubs landed
by PR #173 (C4 PR-1) with real implementations of `WriteRefs`,
`WriteGlobalIndexes`, and `AppendComponentEvent`, plus a new
`ErrRefStale` sentinel for retry-budget exhaustion. Diff is +1602 / -52
across 14 files (all under `internal/catalogstore/` per implementer
report).

Implementer report shipped at canonical path:
`ai/reports/task-0032-implementer.md`.

**FLAGGED BY ORCHESTRATOR — explicit verification target:**

The implementer report and PR body BOTH self-report
`internal/catalogstore` coverage as **85.3 %** with the line "target
≥ 85 %". Task 0032's acceptance criteria required **≥ 91 %** with
floor 90 % and explicitly stated "PR-2 must add enough net coverage
to keep 1 % buffer for future PRs" (see ai/tasks/task-0032.md lines
217–222 and 283–289). PR-1 shipped at 90.7 %.

If the live coverage is genuinely < 90 %, this is a **hard FAIL** —
PR-2 has lowered the package floor below the value PR-1 established
and PR-3 (resolver fallback chain) will inherit zero buffer. The
verifier must:

1. Re-run `go test ./internal/catalogstore/... -race -count=1 -cover`
   from the PR branch and capture the exact percentage.
2. Determine whether the drop is mechanical (deletion of 3 stub-pin
   test entries shifted the denominator) or substantive (new code
   genuinely under-tested).
3. Decide which path applies (see "Coverage Adjudication" below).

Repo health: green. `main` at `fdf72f5` (post-Task 0031 merge bookkeeping).
Sole open PR: #174. No proposals owed.

# Objective
Verify PR #174 against Task 0032's acceptance criteria. PASS = merge
the PR via the Verifier Merge Protocol, sync local main, write the
verifier report, leave the worktree clean. FAIL = leave the PR open
with clear, actionable blockers in the report (with a coverage
remediation path if that is the failing axis).

# PR Boundary
PR #174 only. No scope expansion. Verification-only fixes (e.g.
adding focused tests to lift coverage back above the 90 % floor;
minor doc polish strictly required to PASS) MAY be committed to the
PR branch and re-pushed before merge. Anything beyond that becomes
Task 0033.x or a follow-on (e.g. a coverage-recovery PR).

# Read First
- `ai/tasks/task-0032.md` — original implementer prompt; treat its
  Required Outcomes / Acceptance Criteria / Non-Goals as the
  contract. Pay special attention to the coverage clause (≥ 91 %).
- `ai/reports/task-0032-implementer.md` — implementer self-report
  (committed on the PR branch).
- `agents/orchestrator.md` — Verifier Standard, Verifier Merge
  Protocol (lines 644–657).
- `specs/orun-component-catalog/catalog-store.md` — §3.C
  (`WriteGlobalIndexes`), §3.D (`WriteRefs`), §5 (atomicity / retry),
  §6 (error taxonomy). Note: spec §5 advises retry budget = 8;
  implementer standardised on 16. Verify the inline justification.
- `specs/orun-component-catalog/identity-and-keys.md` — `srcKey` /
  `catKey` / `ComponentKey` / branch sanitisation rules used by
  `refs.go`.
- `specs/orun-component-catalog/data-model.md` — `RefUpdate`,
  `GlobalIndexUpdate`, `ComponentCatalogEvent`, `ComponentLatestRef`,
  `ComponentMainRef`, `PreviewRef` shapes.
- `ai/reports/task-0031-verifier.md` — PR-1 verifier baseline (path
  layer, error taxonomy, write-order convention to be preserved).

Reference Only:
- `internal/catalogstore/writer.go` (PR-1 baseline) — confirms the
  `createOrReconcile` pattern PR-2 should reuse (or document divergence).
- `internal/catalogstore/errors.go` (PR-1) — confirm `ErrRefStale` is
  added consistently with the existing double-wrap pattern.
- `internal/statestore/store.go` — confirm `errors.Is` chain to
  `statestore.ErrConflict` (CAS) and `statestore.ErrExists`
  (CreateIfAbsent) is preserved by PR-2 sentinels.
- `internal/catalogmodel/sanitize.go` — confirm `SanitizeBranch` is the
  helper called by `refs.go` step D.4.

# Required Outcomes

1. **PR-boundary audit.** Only files under `internal/catalogstore/`
   (plus `ai/reports/task-0032-implementer.md`) are touched. No edits
   to `internal/statestore/`, `internal/catalogresolve/`,
   `internal/catalogmodel/`, `internal/sourcectx/`,
   `internal/triggerctx/`, `internal/revision/`,
   `internal/executionstate/`, `cmd/orun/`. No `go.mod` / `go.sum`
   churn. Forbidden file `resolver.go` (PR-3 surface) ABSENT or
   untouched.

2. **Step D write-order audit (`refs.go`).** `WriteRefs` issues
   D.1 → D.2 → D.3 → D.4 → D.5 → D.6 in code AND a spy/recorder test
   asserts the exact sequence. The Source==nil and Catalog==nil
   skip-silently behaviour is exercised. Branch and PR sub-steps fire
   only when the corresponding fields are non-empty.
   Spec §3.D is the contract.

3. **Branch sanitisation.** `refs.go` D.4 passes branch names through
   `catalogmodel.SanitizeBranch` (or equivalent). An empty-after-sanitise
   result surfaces a typed error (`ErrInvalidPathInput` per the
   implementer report). A test exercises at least one rejecting input
   and one accepting input that gets non-trivially sanitised.

4. **Step C write-order audit (`indexes.go`).** `WriteGlobalIndexes`
   issues C.1 (source global index, plain `Write`, skip if Source==nil) →
   C.2 (catalog global index, plain `Write`, skip if Catalog==nil) →
   C.3 (per-component CAS w/ merge-on-conflict, deterministic ascending
   `ComponentKey` order). Read the code to confirm C.1 / C.2 use plain
   `Write` (overwrite-OK per spec — global indexes are rebuildable),
   NOT `CreateIfAbsent` / CAS. C.3 deterministic ordering is asserted
   by a byte-identical-trace test.

5. **`mergeComponentGlobalIndex` policy.** Confirm the policy described
   in the implementer report matches the code:
   - Identity fields (APIVersion/Kind/ComponentKey/Name/Repo): caller
     wins when non-empty (preserve otherwise).
   - `Latest` / `Main`: caller wins when `SourceSnapshotKey != ""`
     (freshness signal); preserve otherwise.
   - `Previews`: union by `SourceSnapshotKey`, caller wins on tie,
     emitted sorted ascending in the merged body.
   A unit test exercises each clause; conflict-injection (CAS losing
   once, retry winning) covers the merge path.

6. **CAS retry semantics.** All three retry sites (`refs.go` per ref,
   `indexes.go` C.3 per component, `events.go` seq.lock allocator)
   share retry budget = 16 and on exhaustion return `ErrRefStale`
   wrapping the last `statestore.ErrConflict`. `errors.Is(err, ErrRefStale)`
   AND `errors.Is(err, statestore.ErrConflict)` BOTH succeed
   (double-wrap pattern, consistent with PR-1's mismatch sentinels).
   A test triggers exhaustion at each site.

   **Spec drift note:** `catalog-store.md` §5 advises retry = 8;
   PR-2 standardises on 16 with inline justification. Decide whether
   this needs a `/ai/proposals/task-0033-spec-update.md` entry. If
   the implementer's inline rationale is sound and the spec wording
   says "advisory", a single-sentence note in the verifier report
   suffices and the proposal is OPTIONAL. If the spec is prescriptive,
   write the proposal and FAIL the PR with that as the blocker.

7. **`AppendComponentEvent` allocator (`events.go`).**
   - `<eventsDir>/seq.lock` holds `{"next":<uint64>}`.
   - Initial attempt: `CreateIfAbsent` with `next=2` (after allocating
     event index 1).
   - On `ErrExists`: read → CAS-with-`next+1` retry loop, budget = 16.
   - Body write: `CreateIfAbsent` at `ComponentHistoryEventPath(...)`.
     Events are immutable; a re-write of the same index with a
     different body must surface a typed error rather than silently
     succeeding.
   - Concurrency: a test exercises two goroutines racing to allocate
     and confirms strictly increasing, gap-free indices (or, at
     minimum, allocator returns distinct indices and never overwrites
     an existing event body). Read the test, don't trust the report.

8. **`seq.lock` path location.** PR-1 / orchestrator hypothesis was
   `<srcKey>/<catKey>/history/components/<name>/events/seq.lock`. Confirm
   the implementer's chosen path matches that convention OR is
   documented in the report with rationale. If the path diverges
   without rationale, FAIL.

9. **`ErrRefStale` taxonomy.** `errors.go` exposes `ErrRefStale` at
   the package level (sentinel value, not a constructor). Code
   inspection confirms every CAS-exhaustion site wraps the LAST
   `statestore.ErrConflict` (not the first; not a fresh error) so
   diagnostics can be chained. A test exercises both
   `errors.Is(err, ErrRefStale)` and the underlying conflict.

10. **Stub-pin test update (`store_test.go`).** The PR-1 pin test
    (`TestStubsReturnErrNotImplemented`) was narrowed to the five
    Resolver methods; a new `TestPR2WritersImplemented` (or
    equivalent) asserts the three PR-2 writer surfaces explicitly do
    NOT return `ErrNotImplemented`. Verify both tests exist and run.

11. **Coverage Adjudication (FLAGGED).** Run from the PR branch:

        go test ./internal/catalogstore/... -race -count=1 -cover

    Capture the exact percentage. Three branches:

    a) **≥ 91 %:** acceptance met as written. PASS this clause.

    b) **90.0 % ≤ x < 91 %:** floor held but acceptance target missed.
       This is a verifier judgement call. Acceptable to PASS with an
       explicit note in the verifier report and a Risk Note that PR-3
       has < 1 % headroom. NOT acceptable to silently pass — the
       deviation must be recorded.

    c) **< 90 % (matches the report's 85.3 % claim):** **HARD FAIL.**
       PR-1 established the floor at 90 %. Lowering it below floor is
       a regression that PR-3 cannot recover from in a single cycle.
       Two remediation paths, verifier picks one:
       - **Verifier-attached fix (preferred when scoped):** add
         focused unit tests on the PR branch (commit + push) lifting
         catalogstore coverage to ≥ 90 % (target ≥ 91 %), wait for CI
         re-green, then PASS and merge. Examples: dedicated tests for
         `mergeComponentGlobalIndex` policy clauses, retry-exhaustion
         paths, seq.lock concurrency, branch sanitisation rejects.
       - **FAIL and remediate via Task 0033.1 (implementer-fix):**
         leave PR open, document specific uncovered branches in the
         verifier report, request implementer add tests. Reserve this
         path when the gap is wider than ~3 percentage points or the
         missing coverage indicates structural test gaps the verifier
         shouldn't paper over.

    Pick the path explicitly. Do not silently merge < 90 %.

12. **Phase 1 + Phase 2 coverage floors held byte-for-byte.** Capture
    exact percentages:
    - `internal/statestore` ≥ 95.7 %
    - `internal/revision` ≥ 90.3 %
    - `internal/executionstate` ≥ 90.0 %
    - `internal/catalogmodel` ≥ 91.1 %
    - `internal/sourcectx` ≥ 91.1 %
    - `internal/catalogresolve` ≥ 90.9 %

    Any drop = HARD FAIL (PR-2 must not regress sibling packages).

13. **Static guards.** `go vet ./...`, `go build ./...`,
    `go test ./... -race -count=1`, `make verify-generated` all clean.
    `grep` confirms NO `os` / `io/ioutil` / `path/filepath` imports
    introduced anywhere under `internal/catalogstore/`.

14. **CI evidence.** `gh pr view 174 --json statusCheckRollup` shows
    all required checks SUCCESS (re-confirm at merge time;
    verifier-attached fix commits MUST re-green CI before merge).

15. **kiox / orun guards.** Run when applicable; record no-op result
    when the plan has no jobs (the catalogstore package is internal
    Go and the typical Orun graph for catalog work is empty).

# Non-Goals
- No code changes inside `internal/catalogstore/` EXCEPT
  verifier-attached coverage fixes per Outcome 11 path (a) or
  trivial doc/polish strictly required to PASS.
- No PR-3 work (`resolver.go`, fallback chain `current → latest → main`,
  the five Resolver methods, `RebuildIndexes`).
- No spec edits unless a behavioural drift is identified — in which
  case write `/ai/proposals/task-0033-spec-update.md` and either FAIL
  the verification or PASS-with-proposal depending on severity.
- No coverage rebalancing across other packages.
- No changes to retry budgets agreed at 16 unless Outcome 6's spec
  proposal requires it.

# Constraints
- No edits outside `internal/catalogstore/` (and the verifier report).
- Verifier-only commits MUST be on the PR branch and pushed; CI
  must re-pass before merge.
- Use the canonical Verifier Merge Protocol from
  `agents/orchestrator.md` lines 644–657.
- Never merge while CI is red, MergeStateStatus ≠ CLEAN, any
  Required Outcome is unmet, or `internal/catalogstore` coverage
  is below 90 %.
- Preserve PR-1 conventions: no raw FS imports; double-wrap error
  pattern; deterministic-trace test idiom.

# Integration Notes
- The PR-2 contract freezes `RefUpdate` / `GlobalIndexUpdate` /
  `ComponentCatalogEvent` shapes (PR-1 already declared them; PR-2
  is the first writer). PR-3 (`resolver.go` + fallback chain +
  `RebuildIndexes`) MUST consume these without modification.
- `mergeComponentGlobalIndex` is the only domain-policy function in
  PR-2; everything else is mechanical write/CAS plumbing. Read it
  carefully — a wrong "caller-wins" rule here corrupts the global
  index across the whole catalog.
- Retry budget = 16 vs spec advisory = 8: the seq.lock allocator is
  the most contention-prone site (single counter, every event append
  hits it). Implementer's choice is defensible; flag it in the report
  but do not block the PR if the inline justification is solid.
- `ErrRefStale` joins the existing `Err*Mismatch` family as a
  retry-exhaustion signal. Future Resolver code (PR-3) and CLI
  surface (C5+) will pattern-match on it.

# Acceptance Criteria
✅ All 15 Required Outcomes met.
✅ `go vet ./...` clean.
✅ `go build ./...` clean.
✅ `go test ./internal/catalogstore/... -race -count=1 -cover` ≥ 90 %
   (≥ 91 % preferred; < 90 % is HARD FAIL absent verifier-attached fix).
✅ `go test ./... -race -count=1` green.
✅ `make verify-generated` clean.
✅ Coverage floors held byte-for-byte (record exact percentages for
   all six floor-gated packages).
✅ Spy / recorder test asserts D.1→D.6 ordering and conditional steps
   (Source/Catalog nil-skip, Branch/PR conditional emission).
✅ `mergeComponentGlobalIndex` per-clause tests present.
✅ CAS retry-exhaustion test at each of three sites surfaces
   `ErrRefStale` chained to `statestore.ErrConflict`.
✅ `seq.lock` allocator concurrency test present.
✅ Stub-pin tests updated (Resolver-only) AND PR-2-implemented test
   present (writers no longer return `ErrNotImplemented`).
✅ No raw FS imports introduced.
✅ PR #174 `mergeStateStatus: CLEAN`, `statusCheckRollup` all SUCCESS.
✅ Verifier report at `ai/reports/task-0033-verifier.md`.
✅ Local kiox guards (when applicable) recorded — record no-op if no
   plan jobs.

# Verification

Run locally and capture exact output:

    cd /Users/irinelinson/sourceplane/orun
    git fetch origin
    git checkout task-0032-catalogstore-c4-pr2
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
    grep -rn '"os"\|"io/ioutil"\|"path/filepath"' internal/catalogstore/ \
        || echo "no raw FS imports — OK"

    # PR / CI evidence
    gh pr view 174 --json title,state,mergeable,mergeStateStatus,statusCheckRollup,headRefName,url
    gh pr diff 174 --name-only

Then run kiox guards:

    /Users/irinelinson/.local/bin/kiox -- orun validate --intent intent.yaml
    /Users/irinelinson/.local/bin/kiox -- orun plan --changed --intent intent.yaml --output plan.json
    /Users/irinelinson/.local/bin/kiox -- orun run --plan plan.json --dry-run --runner github-actions

Record no-op when the plan has no jobs.

Code-path inspections (read the files, not just the tests):

- `internal/catalogstore/refs.go::WriteRefs` — confirm D.1→D.6
  ordering literal in the function body; confirm Source==nil and
  Catalog==nil short-circuits per side; confirm
  `catalogmodel.SanitizeBranch` is invoked for D.4.
- `internal/catalogstore/indexes.go::WriteGlobalIndexes` — confirm
  C.1 / C.2 use `state.Write` (overwrite OK), C.3 uses
  `state.CompareAndSwap` and iterates `ComponentKey`-ascending.
- `internal/catalogstore/indexes.go::mergeComponentGlobalIndex` —
  trace each policy clause and confirm the test for that clause
  asserts the documented winner.
- `internal/catalogstore/events.go::AppendComponentEvent` — trace
  the seq.lock CreateIfAbsent → CAS retry loop; confirm budget = 16;
  confirm event body uses `CreateIfAbsent` (immutability).
- `internal/catalogstore/errors.go` — confirm `ErrRefStale` is a
  package-level sentinel and that wrapping uses the multi-target
  `errors.Is` pattern (e.g. `fmt.Errorf("%w: %w", ErrRefStale,
  conflictErr)` or a custom `Unwrap() []error`).
- `internal/catalogstore/store_test.go` — confirm the stub-pin test
  is narrowed to Resolver methods and a positive-implementation
  pin test exists for the three writers.

# PR Creation Requirement
Implementer already created PR #174. Verifier does not create a new
PR. If a verification-only fix is required (most likely a coverage
top-up per Outcome 11), commit it to the PR branch, push, wait for
CI to re-go-green, then merge.

# Verifier Merge Protocol (from agents/orchestrator.md lines 644–657)
- If PASS:
  1. `gh pr merge 174 --squash --delete-branch` (admin if required).
  2. `git checkout main && git pull --ff-only origin main`.
  3. `git status --short` → must be clean. Resolve any verifier-side
     leftover changes before declaring done.
  4. Commit the verifier report + state-file updates to `main`
     directly per prior cycles' convention.
- If FAIL: leave PR #174 open with all blockers spelled out.
  Do NOT merge under any circumstance with red CI, dirty
  `mergeStateStatus`, any unresolved Required Outcome, or
  `internal/catalogstore` coverage below 90 %.

# When Done Report
Write `/ai/reports/task-0033-verifier.md` with:

- `Result: PASS|FAIL`
- `Checks` — every command run, exact output (or summarised when
  long; never claim a check passed without showing evidence)
- `CI Log Review` — `gh pr view 174` JSON snippet, run IDs cited
- `Issues` — empty if PASS; specific blockers if FAIL
- `Code Path Inspection` — short notes on the six reads above
- `Coverage Evidence` — exact % for `internal/catalogstore` (with
  before/after if a verifier-attached fix was applied) and every
  floor-gated package
- `Coverage Adjudication` — explicit branch (a / b / c) chosen and
  rationale; if path (a) fix was applied, list the test names added
  and the coverage delta
- `Retry-Budget Spec Drift` — explicit decision: proposal owed
  (link path) or accepted-with-rationale (one-sentence note)
- `Risk Notes` — residual risk after merge (esp. PR-3 headroom)
- `Spec Proposals` — links only; expected `_none_` unless retry-budget
  drift escalated
- `Recommended Next Move` — on PASS, point to Task 0034 = C4 PR-3
  implementer (`resolver.go`, fallback chain
  `current → latest → main`, all 5 Resolver methods,
  `RebuildIndexes`). On FAIL, name the smallest remediation that
  closes the gap (e.g. Task 0033.1 implementer-fix for coverage).
- `PR #174 Merge Status` — merged SHA + branch deletion confirmation,
  or "left open, blockers above".
