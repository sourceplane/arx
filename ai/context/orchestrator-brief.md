# Orchestrator Brief

## Cache Fingerprint
- generated_at: 2026-05-31T10:30:00Z
- cycle_seq: 5
- head_sha: 7669cc0faa07a7a7da6d683384e3fa53184e7264
- state_json_sha256: 5293eb90dfad10a7f5cef8c4b27ebd282085fa60c2e7c0b3df574ab97112be6a
- merged_pr_count: 166
- open_pr_count: 1
- last_task_agent: ai/tasks/task-0031-verifier.md
- last_worker_result: implementer-pass
- cycle_5_action: emit Task 0031 verifier on PR #173 (Task 0030 implementer-pass).
  PR is MERGEABLE/CLEAN, CI 4/4 SUCCESS. State / current / ledger updated;
  this brief overwritten. Bookkeeping commit pending on `main` after this
  brief lands.

## Cache Validity Rule
The next cycle MAY skip the cold read (loop steps 1–7) iff ALL of:
- `git rev-parse HEAD` (on main) ==
  `7669cc0faa07a7a7da6d683384e3fa53184e7264` (pre-bookkeeping-commit
  expected; once the bookkeeping commit lands the SHA will advance —
  fingerprint mismatch is INTENTIONAL there and forces a one-shot
  cold-read at the start of cycle 6 to pick up the merge of PR #173
  if it has happened by then).
- `shasum -a 256 ai/state.json` first field ==
  `5293eb90dfad10a7f5cef8c4b27ebd282085fa60c2e7c0b3df574ab97112be6a`
- `gh pr list --state merged --limit 1000 | wc -l` == 166 (PR #173
  not yet merged) OR == 167 (verifier-pass + merge happened — that
  is the predicted fingerprint mismatch and forces cold read).
- `gh pr list --state open | wc -l` == 1 (PR #173 still open) OR
  == 0 (PR #173 merged — predicted verifier-pass path; mismatch
  forces cold read for next-task emission).
- next cycle_seq ≤ 8 (this brief generated at seq 5; valid for 6–7).
Otherwise: discard this brief and do a full cold read.

## Mental Model (the synthesis)
Cycle 5 is a clean implementer-pass closure. PR #173 came in exactly on
the predicted next_cycle_hypothesis path — `internal/catalogstore`
greenfielded with paths + body writer + frozen public surface, CI 4/4
green, MERGEABLE/CLEAN, no spec proposals owed. The implementer report
shipped at a non-canonical path (`reports/task-0030-catalogstore-c4-pr1.md`
on the PR branch instead of `ai/reports/task-0030-implementer.md`); I
deliberately did NOT escalate that — it's a verifier judgement call, and
the report's content is dense and accurate. The non-obvious things in
the delivered code are: (1) mismatch sentinels DOUBLE-WRAP
`statestore.ErrExists`, so callers that key off the statestore sentinel
keep working while gaining the typed-discriminator path; (2) graph write
order is locked via `CatalogGraphKinds()` in code, not derived from the
input map's iteration order — verifier should confirm this with a
spy-based call-order test that feeds graphs in random order; (3) body
writes use `PrettyEncode`, not `CanonicalEncode` — the implementer's
stated rationale is that `CanonicalEncode` is reserved for hashing
upstream and PR-1 has no hashing responsibility. That's the right call,
but verifier should confirm consistency. The leverage point for cycle 6
is binary: verifier-pass advances to C4 PR-2 (`refs.go` + `indexes.go` +
`AppendComponentEvent`), verifier-fail keeps the loop on the same PR.
Either way the next emission is fully predictable.

## Active Spec Pointer
- spec: specs/orun-component-catalog
- milestone: C4 (PR-1 awaiting verification, PR-2/PR-3 still queued)
- milestone_done_when_remaining:
  - PR-1 verified-and-merged with `internal/catalogstore` ≥ 90 % cov
    (claimed 90.7 %; verifier to re-measure).
  - All of `catalog-store.md` §3 steps C–D shipped (PR-2): refs
    write-with-CAS, global indexes write-with-CAS,
    `AppendComponentEvent` seq.lock retry-up-to-16.
  - §4 reader fallback chain (`current → latest → main`) +
    `RebuildIndexes` byte-identical idempotence (T-STORE-3) shipped
    (PR-3 or follow-up under C4 scope).
  - All writes go through `internal/statestore`; an `import-restriction`
    lint enforces it in CI (PR-2 or PR-3 will add this).
  - Refs `current` / `main` / `branches/<x>` / `prs/<n>` round-trip
    (PR-2).
  - Reader fallback exercised by a test that scrubs the global index
    and asserts walk-based recovery (PR-3).
- next_milestone_after: C5 — Catalog CLI (`orun catalog refresh / list /
  describe / refs / tree / history / validate`, `diff` stubbed for C8).
  Cannot start until C4 closes because every CLI command consumes the
  Resolver shipped in C4 PR-3.

## Open PRs (one line each)
- #173 `feat(catalogstore): C4 PR-1 — paths, errors, Writer (sources +
  catalog snapshots)` — adampullely — green (CI 4/4 SUCCESS,
  MERGEABLE/CLEAN) — Task 0030 implementer-pass; awaiting Task 0031
  verifier; orchestrator-relevance: this is the PR that unlocks PR-2
  scope on PASS.

## Deferred Backlog (parking lot summary)
_none_ — no `/ai/deferred.md` file exists. Phase 1 carry-forward
candidates (MirrorModeHardlink debug-fold, RunnerHooks.AfterStateUpdate
async-mirror, `--persist-revision` flag, Option B trigger-name resolver,
`--prune-legacy`) remain tracked in `state.json.notes` as post-Phase-2
candidates, not currently scheduled.

## Active Proposals
- `ai/proposals/task-0002-spec-update.md` — closed (rapid import-path
  clarification, folded into Phase 1 docs at M-time). Stance: archived.
- `ai/proposals/task-0019-spec-update.md` — closed (Phase 1 trigger-name
  resolver Option B; deferred as Phase 1 carry-forward in
  `state.json.notes`). Stance: deferred (Phase 2 carry-forward).
- `ai/proposals/task-0025-spec-update.md` — closed (folded into Task
  0026 prompt; convention adopted as load-bearing Phase 2 rule).
  Stance: closed-and-folded.

No new proposals owed by the orchestrator this cycle. Task 0030
implementer report explicitly recorded "_none_" for spec proposals;
the spec was followed exactly. Brief confirms: no spec-update task
warranted.

## Last Decision Rationale
Why Task 0031 = verifier on PR #173 (rather than scope a parallel PR-2
implementer slot) was the highest-leverage emission this cycle:
- The PR-Sized Task Standard explicitly forbids overlapping a verifier
  pass and the next implementer pass when both touch the same package.
  PR-2 (refs + indexes) shares `internal/catalogstore/store.go` with
  PR-1 (interface decls live there), so spawning PR-2 before PR-1
  merges would create a near-certain rebase conflict on the deferred
  method bodies and force re-review on the PR-1 surface.
- The C4 spec's 2–3 PR seam is sequentially-coupled by design:
  PR-1 freezes the public surface; PR-2 fills bodies; PR-3 adds
  reader. Parallelising PR-1 verifier and PR-2 implementer would
  also defeat the "freeze contract first" leverage that motivated
  the seam.
- The orchestrator already wrote the next_cycle_hypothesis on this
  exact path in cycle 4's brief; honouring the prediction without
  re-litigation is the warm-boot's whole purpose.
- A user-redirect alternative ("skip verification, ship PR-2 now")
  would violate the Verifier Standard and the implementer's
  PR-Creation-Requirement contract — not on the table without
  explicit instruction.
- The verifier prompt asks for inspection (read the writer code +
  the spy test, not just the report) because the implementer's
  self-report is unusually detailed AND unusually accurate; the
  cheapest way to convert that into PASS confidence is read-and-confirm,
  not re-implement.

## Next Cycle Hypothesis
- **if verifier-pass on Task 0031:** PR #173 merges, `main` advances,
  emit **Task 0032** = C4 PR-2 implementer (`refs.go` + `indexes.go`
  covering write-order steps C and D, plus `AppendComponentEvent`
  with the seq.lock retry-up-to-16 contract from `catalog-store.md`
  §3.D). Read first: `catalog-store.md` §3 steps C+D, §4 (refs only —
  ResolveCatalog stays in PR-3), §5 atomicity, §6 error taxonomy
  (add `ErrRefStale` here). Forbidden in PR-2: `resolver.go`,
  `RebuildIndexes`, anything outside `internal/catalogstore/`.
- **if verifier-fail with bounded fix:** likely surfaces — (a)
  graph-order assertion test missing or weak (only one fixed-order
  case, no random-order spy); (b) `errors.Is` chain to
  `statestore.ErrExists` not actually reachable through the
  double-wrap (would need a test that explicitly checks both
  targets); (c) coverage on `internal/catalogstore` falling under
  90 % under `-race -count=1` (the implementer's measurement was
  90.7 % — close enough to the floor that one missing branch could
  drop it). Remediation stays inside Task 0030's PR; emit no new
  task slot. Verifier commits the patch on the PR branch and
  re-merges.
- **if verifier-fail with scope expansion:** unlikely. If it
  happens, most plausible cause is that the pre-flight
  `ErrInputsInconsistent` validator should be a separately
  exposed validator method (so C5 CLI / C6 plan can call it
  without committing). Emit Task 0030.1 narrowly extracting the
  validator if that's the call.
- **if a new spec proposal arrives at `ai/proposals/task-0031-*.md`:**
  orchestrator must adjudicate before merging. Most likely topic:
  the local-index file body schema (not pinned in `data-model.md`)
  or the stub policy (typed `ErrNotImplemented` vs. panic).
  Decide accept-leaning unless it widens the public surface
  declared in `catalog-store.md` §1.
- **if user-redirect arrives:** force cold read; re-scope per
  `references/user-directed-roadmap-override.md`.

## Stale Signals (what would invalidate this brief early)
- A new spec proposal arrives at `ai/proposals/task-0031-*.md` —
  force cold read to adjudicate before continuing.
- A user redirect away from C4 PR-2 (e.g. "skip the verifier,
  ship PR-2 now" or "land all of remaining C4 in one PR") — force
  cold read and re-scope per `references/user-directed-roadmap-override.md`.
- CI starts failing on `main` (a `7669cc0` post-merge regression) —
  force cold read; stabilise before continuing C4.
- PR #173 head SHA changes (verifier-attached fix lands on the
  branch) — force a re-read of the diff before merge.
- A Phase 1 floor regression (statestore < 95.7, revision < 90.3,
  executionstate < 90.0) — force cold read; treat as a Priority-1
  invariant break and scope a stabilisation task before continuing.
- The verifier discovers `internal/catalogmodel.PrettyEncode` is
  not a stable public symbol or doesn't exist — that's a wiring
  problem the implementer skipped and would force a small fix on
  the PR branch (or in a tiny additive sibling per the C2 PR-1
  convention).
