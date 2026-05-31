# Orchestrator Brief — Cycle 7

## Cache Fingerprint
- generated_at: 2026-05-31
- cycle_seq: 7
- head_sha: fdf72f50f77896b58af7fc8f7f806b222fe9358f
- state_json_sha256: 31838ce94096cc2ff8568af9a2e4e326b5b143a92dc21446fe56dad17cbaa932
- merged_pr_count: 167 (gh pr list --state merged --limit 1000 | wc -l)
- open_pr_count: 1 (PR #174)
- last_task_agent: ai/tasks/task-0033-verifier.md
- last_worker_result: implementer-pass (Task 0032 → PR #174)
- worktree_dirty: ai/ only (state.json + current.md + task-ledger.md +
  ai/tasks/task-0033-verifier.md + this brief; bookkeeping commit
  pending after this write)

## Cache Validity Rule
The next cycle MAY skip the cold read iff ALL of:
- head_sha matches `git rev-parse HEAD` (will advance once cycle-7
  bookkeeping commit lands; expected new HEAD = next-tip after
  `ai: cycle 7 — Task 0032 implementer-pass (PR #174); scope Task 0033 verifier`).
- state_json_sha256 matches recomputed hash.
- merged_pr_count = 167; open_pr_count expected 1 (PR #174 still open)
  OR 0 (verifier merged PR #174 — that DOES invalidate this brief
  because cycle 8 must scope Task 0034, not re-read this one).
- cycle_seq within 3 of next cycle's seq.
Otherwise: discard this brief and do a full cold read.

## Mental Model
Phase 2 / Milestone C4 / mid-flight. PR-1 (paths + writer Steps A & B)
landed cleanly at 90.7 % coverage. PR-2 (Steps C & D — refs, global
indexes, component events + `ErrRefStale`) is now open as PR #174,
CI green and MERGEABLE. The implementer shipped a substantial,
correctly-shaped PR (D.1–D.6 ordering, mergeComponentGlobalIndex
policy, seq.lock allocator with 16-retry budget) but **self-reports
catalogstore coverage at 85.3 %** — 5.4 pp below the floor PR-1
established and 5.7 pp below Task 0032's stated target of ≥91 %.

This is the cycle's only real tension. Mechanical cause is plausible
(stub-pin denominator shifted when 3 stubs became real implementations,
diluting the percentage even with new tests added) but the floor
regression is real either way. The verifier prompt (Task 0033) makes
this an explicit Required Outcome (11) with a three-branch decision
tree: ≥91 PASS / 90–91 PASS-with-note / <90 HARD FAIL → verifier-
attached fix preferred, else Task 0033.1 implementer-remediation.
The verifier's first action is to re-run the cover and pick a branch.

Secondary tension: spec §5 advises retry budget = 8; PR-2 standardises
on 16. Inline rationale is sound (taxonomy unification with new
`ErrRefStale` sentinel). Likely a one-line verifier note rather than
a spec proposal.

## Active Spec Pointer
- spec: specs/orun-component-catalog
- milestone: C4 (`internal/catalogstore` Writer + Resolver)
- milestone_done_when_remaining:
  - C4 PR-2 (refs/indexes/events) merged with floors held — PENDING
    (Task 0033 verifier slot).
  - C4 PR-3 (`resolver.go` + fallback chain `current → latest → main`
    + 5 Resolver methods + `RebuildIndexes`) — PENDING.
- next_milestone_after: C5 — Catalog CLI surface (`orun catalog *`
  subcommands per `cli-surface.md`). Unlocked once C4 PR-3 merges.

## Open PRs
- #174 `feat(catalogstore): C4 PR-2 — refs, global indexes, component events`
  — implementer (task-0032) — green CI / MERGEABLE / CLEAN —
  awaiting Task 0033 verifier; coverage adjudication is the gating
  question.

## Deferred Backlog
_none_ — `/ai/deferred.md` carries no entries this cycle. All
roadmap candidates remain human-independent.

## Active Proposals
_none_ — no `/ai/proposals/**` entries. Task 0033 may emit
`ai/proposals/task-0033-spec-update.md` if the verifier reads spec
§5 retry-budget wording as prescriptive rather than advisory; that
proposal would auto-create on the verifier branch and drive the
PASS/FAIL decision.

## Last Decision Rationale
Why scope a verifier (Task 0033) and not a remediation prompt:
- PR #174 is OPEN with green CI, MERGEABLE/CLEAN — there is real
  work for the verifier to do (15 outcomes), and the coverage gap
  may be mechanical (stub-pin denominator shift) rather than a
  test-quality regression.
- Surfacing the coverage flag in the verifier prompt (Outcome 11)
  rather than auto-failing the PR preserves the verifier's ability
  to apply a verifier-attached fix in path (a), which is the cheaper
  resolution and matches PR-1's verifier-attached-fix precedent
  (PR-1 verifier moved the implementer report to canonical path).
- Generating Task 0033.1 prematurely would waste a cycle if the
  verifier can land a 6–10-test top-up to recover ≥ 90 %.
- Three-branch decision tree avoids "silent merge below floor" while
  keeping the verifier's hands on the wheel.
- Retry-budget spec drift folded into the same verifier task because
  it shares review surface and the implementer's inline rationale is
  documented; no need to fork a separate proposal-review cycle yet.

## Next Cycle Hypothesis
- if **verifier-pass on PR #174 (Outcome 11 branch a, ≥91 %)**:
  cycle 8 scopes Task 0034 = C4 PR-3 implementer. Surface:
  `internal/catalogstore/{resolver.go, resolver_test.go}` plus
  `RebuildIndexes` in `writer.go` (or sibling). All 5 Resolver
  methods + fallback chain `current → latest → main`. Closes C4.
- if **verifier-pass via attached-fix (branch c → a)**:
  same as above, but cycle 7 ledger reflects verifier-added test
  files; coverage delta recorded in verifier report.
- if **verifier-pass with coverage 90 ≤ x < 91 (branch b)**:
  cycle 8 scopes Task 0034 with explicit "PR-3 must restore ≥ 91 %
  buffer" added to acceptance.
- if **verifier-fail (branch c, no attached fix)**:
  cycle 8 scopes Task 0033.1 = implementer-fix on the same PR
  branch (add focused tests for uncovered branches in
  mergeComponentGlobalIndex / retry-exhaust paths / seq.lock
  concurrency). PR #174 stays open until coverage clears 90 %.
- if **spec-proposal route on retry-budget**:
  cycle 8 reviews `ai/proposals/task-0033-spec-update.md`,
  decides accept/revise/defer. Most likely outcome: accept (codify
  16) and fold into a cli-surface.md / catalog-store.md update
  during C5 docs polish.

## Stale Signals
- New PR opened on the repo (other than #174) — invalidates
  `open_pr_count`.
- `main` advances past `fdf72f5` by anything other than the cycle-7
  bookkeeping commit before next cycle starts.
- New file under `ai/proposals/` (e.g. retry-budget proposal lands).
- `git log` shows PR #174 squash-merge commit — verifier passed,
  cycle 8 must cold-read to scope Task 0034.
- A user redirect away from C4 PR-3 (e.g. "skip resolver, go straight
  to CLI") — would force cold read and update of Active Spec Pointer.
- Any drop in a sibling coverage floor — invalidates the "floors
  held byte-for-byte" precondition this brief assumes.
