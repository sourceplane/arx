# Task 0024 (Verifier pass)

# Agent
Verifier

# Current Repo Context
- Phase 2 / Milestone **C1** of `specs/orun-component-catalog/` (the
  `internal/sourcectx` resolver) was implemented by Task 0024 against
  branch `impl/task-0024-c1-sourcectx-resolver` and opened as **PR #169**
  (head `59a855e`). Implementer report: `ai/reports/task-0024-implementer.md`.
- Predecessor C0 code half (Task 0023, PR #168 → main commit `7f3f2bf`) is
  merged. C0 spec landing (Task 0022, PR #167 → main commit `d435d8f`)
  is merged. Phase 1 floors held: `internal/statestore` 95.7 %,
  `internal/revision` 90.3 %, `internal/executionstate` 90.0 %.
- **PR #169 status:** OPEN, MERGEABLE, mergeStateStatus **UNSTABLE** —
  required `state-redesign-tests / test` check is **FAILING**. CI run
  `26704817160` failed at the `internal/catalogmodel` coverage gate
  (`measured: 87.9% below 90% threshold`).
  - The same gate measured **90.2 %** locally for the implementer.
  - On verifier-side reproduction (this prompt was scoped from the
    orchestrator session), three back-to-back runs of
    `go test -count=1 -cover ./internal/catalogmodel/` produced
    **90.6 %, 87.9 %, 90.6 %**. The measurement is flaky across the
    90 % gate boundary on the C1 branch's catalogmodel test set.
  - C1 itself does not modify `internal/catalogmodel` (diff vs `main`
    on that path is empty), so the flake is inherited from a randomized
    test (likely `pgregory.net/rapid`-driven) whose covered branches
    vary by seed/iteration.
- Other PR #169 checks: `CI / Orun Plan` SUCCESS (run `26704817149`),
  `orun remote-state conformance / Harness dry-run guard` SUCCESS
  (run `26704817163`). Matrix legs SKIPPED legitimately (empty matrix).

# Objective
Verify PR #169 against the Verifier Standard in `agents/orchestrator.md`
and the C1 "done when" criteria in
`specs/orun-component-catalog/implementation-plan.md`. Adjudicate the
catalogmodel coverage flake and decide whether it is (a) a C1 blocker,
(b) a C0 carry-forward to fix in this PR, or (c) acceptable with a
re-run + clear justification. Your default disposition is FAIL until the
required CI check is green on the PR head.

# PR Boundary
Verification only. Permitted edits to the PR branch (commit + push):

- `ai/reports/task-0024-verifier.md` (this verifier's report).
- A minimal, surgical fix to make the catalogmodel coverage gate
  deterministic — choose ONE of:
  1. Add 1–3 deterministic table-driven tests in
     `internal/catalogmodel/*_test.go` that cover the branches the
     rapid-driven tests sometimes miss (preferred).
  2. Pin the rapid `Check`/`Custom` seed and/or iteration count in
     the property tests so the same set of branches is exercised
     every run.
  3. Replace the coverage-gate measurement in `Makefile` with a
     run that fixes randomness (e.g. add `-rapid.seed=…`) — only if
     options 1/2 cannot be made deterministic.
- Backfill the implementer report's `PR Number:` line if it still says
  `_opened by this report — see commit/push below_` instead of `#169`.

Out of scope for this verifier pass:
- Any work on C2 (`internal/catalogresolve`).
- Any change to `internal/sourcectx` source files unless required to
  make a real test failure go green (it is not currently failing).
- Any change to `agents/orchestrator.md`, `specs/`, or
  `ai/context/*` apart from the orchestrator-owned post-merge update.

# Read First
- `ai/tasks/task-0024.md` — original implementer prompt (scope contract).
- `ai/reports/task-0024-implementer.md` — implementer claims.
- `agents/orchestrator.md` — Verifier Standard, Verifier Merge Protocol.
- `specs/orun-component-catalog/README.md` — entry + read order.
- `specs/orun-component-catalog/implementation-plan.md` — **Milestone C1**
  goal, dependencies, "done when" list.
- `specs/orun-component-catalog/data-model.md` — `WorkspaceState`,
  `SourceSnapshot`, `dirtyHash`, `catalogInputHash` shapes the resolver
  must emit.
- `specs/orun-component-catalog/identity-and-keys.md` §1–§6 — frozen
  ID + key contract `internal/sourcectx` writes against.
- `specs/orun-component-catalog/resolution-pipeline.md` — Stage 1
  (workspace/source resolution) is what C1 implements.
- `specs/orun-component-catalog/test-plan.md` §1 (coverage targets),
  §3 (T-IDK-3, T-IDK-4 property tests), §7 (per-call budget).
- `specs/orun-state-redesign/state-store.md` (reference only — C1 must
  not touch any Phase 1 package or weaken any Phase 1 floor).
- PR #169 diff and CI logs:
  - `gh pr view 169 --json title,state,mergeable,mergeStateStatus,statusCheckRollup`
  - `gh pr diff 169`
  - `gh run view 26704817160 --log-failed` (the failed `test` check)

# Required Outcomes
- [ ] Verifier report at `ai/reports/task-0024-verifier.md` with the
      mandatory section set: Result, Checks, Issues, Risk Notes,
      Spec Proposals, Recommended Next Move. Include explicit
      `CI Log Review` and `Coverage Adjudication` subsections.
- [ ] PR #169 either (a) MERGED to `main` with a fast-forward local
      pull and clean worktree, or (b) left OPEN with the verifier
      report committed to the PR branch, the failing CI documented,
      and the FAIL result clearly stated.
- [ ] If MERGED: `git log --oneline -1` shows the squash commit on
      `main`, the local branch `impl/task-0024-c1-sourcectx-resolver`
      is gone (`git branch --list` empty for it), `git status --short`
      is clean.
- [ ] State files reconciled by the orchestrator on next cycle (you
      may flag missing entries; do not edit `ai/context/*` or
      `ai/state.json` yourself).

# Non-Goals
- No new C2 scaffolding (`internal/catalogresolve`).
- No tightening of Phase 1 coverage floors.
- No spec edits beyond a `/ai/proposals/` proposal if a real spec
  drift surfaces.
- No `kiox -- orun plan --changed` debugging on the developer machine
  beyond noting the long-standing composition-cache quirk if it
  reproduces.

# Constraints
1. Required CI on PR #169 head MUST be green before merge — the
   `state-redesign-tests / test` check has to be SUCCESS, not just
   "the catalogmodel package compiles". Re-run after pushing your
   fix and wait for the new conclusion before deciding.
2. Phase 1 invariants stay byte-for-byte (`internal/statestore` 95.7 %,
   `internal/revision` 90.3 %, `internal/executionstate` 90.0 %,
   `internal/triggerctx` passes). The `make test-state-redesign`
   header must continue to print these gates green.
3. C0 invariants stay: `internal/catalogmodel` ≥ 90 % AND `Sanitize*`
   == 100 % every run, deterministically. If you raise the gate
   stretch, do not lower the floor.
4. `internal/sourcectx` stays leaf-clean: imports `internal/catalogmodel`
   and stdlib only.
5. No `encoding/json` defaults for hashed payloads — keep the canonical
   encoder path.
6. No new module dependencies introduced by this verifier pass.
7. The verifier-side fix must be a SINGLE additional commit on the PR
   branch (or a single force-push amend of the verifier report
   commit), not an interleaved series.

# Integration Notes
- The catalogmodel coverage flake is the headline blocker. Treat it
  as a real defect on the PR head — the PR cannot merge while
  required CI is red, even though the new C1 code itself looks clean.
- Inspect `internal/catalogmodel/*_test.go` for `rapid.Check`,
  `rapid.Custom`, or any `t.Run` that randomizes inputs. Find the
  branch in production code that is missed when the rapid seed lands
  badly. The cheapest fix is a deterministic table-driven test
  covering that branch directly.
- Consider whether the C0 implementation-plan "≥ 90 %" was meant
  as a budget against deterministic tests; if so, file a one-line
  spec-clarification note (not a behavioral proposal) into the
  verifier report's "Spec Proposals" section. Do NOT raise a
  proposal file unless behavior changes.

# Acceptance Criteria
- ✅ PR #169 corresponds exactly to Task 0024 / Milestone C1 scope —
  no overreach into `internal/catalogresolve`, `internal/catalogstore`,
  or CLI surface.
- ✅ `ResolveSourceSnapshot` produces a `WorkspaceState` whose
  `SourceScope`, `Revision`, `DirtyHash`, and source-snapshot key
  match `data-model.md` and `identity-and-keys.md` byte-for-byte
  (spot-check via the resolver tests).
- ✅ T-IDK-3 (ordering stability) and T-IDK-4 (non-catalog-relevant
  insulation) tests exist in `internal/sourcectx/resolver_test.go`
  and pass.
- ✅ `internal/sourcectx` is leaf-clean (`go list -deps
  ./internal/sourcectx/... | grep github.com/sourceplane/orun/internal/`
  returns at most `internal/catalogmodel`).
- ✅ `make test-state-redesign` is green deterministically (run
  it ≥ 3 times back-to-back in the verifier checks). All gates
  including catalogmodel ≥ 90 % and Sanitize* == 100 % pass each run.
- ✅ `go build ./...`, `go vet ./...`,
  `go test -race -count=1 ./...` green.
- ✅ `make verify-generated` green (committed schema matches
  generator output).
- ✅ `kiox -- orun validate --intent intent.yaml` exits 0 (or
  document the persistent composition-cache quirk if `orun plan
  --changed` repros locally — CI is authoritative).
- ✅ PR #169 required `state-redesign-tests / test` check is **SUCCESS**
  on the PR head before merge.
- ✅ No secrets in the diff, the implementer report, the verifier
  report, or any CI log.
- ✅ MergeStateStatus is CLEAN at merge time.
- ✅ After merge: `main` is fast-forward-pulled locally,
  `git status --short` is empty, and the C1 branch is deleted.

# Verification

## 1. Inspect the PR boundary
```bash
gh pr view 169 --json title,state,mergeable,mergeStateStatus,headRefOid,statusCheckRollup
gh pr diff 169 --name-only
git diff --stat main...origin/impl/task-0024-c1-sourcectx-resolver
git diff main...origin/impl/task-0024-c1-sourcectx-resolver -- \
  internal/catalogmodel/ internal/statestore/ internal/revision/ \
  internal/executionstate/ internal/triggerctx/ specs/ agents/
# Phase 1 packages, specs/, agents/ MUST all show empty diff.
# internal/catalogmodel/ should also show empty diff (C0 carry-forward).
```

## 2. Inspect the failing CI run
```bash
gh run view 26704817160 --log-failed | head -200
gh run list --branch impl/task-0024-c1-sourcectx-resolver --limit 5
```
Confirm the failure is the `internal/catalogmodel` coverage gate
hitting 87.9 %, not a real `internal/sourcectx` regression.

## 3. Reproduce the flake
```bash
git fetch origin impl/task-0024-c1-sourcectx-resolver
git checkout origin/impl/task-0024-c1-sourcectx-resolver
for i in 1 2 3 4 5; do
  go test -count=1 -cover -coverprofile=/tmp/orun-catalogmodel.cov \
    ./internal/catalogmodel/ >/dev/null
  go tool cover -func=/tmp/orun-catalogmodel.cov | tail -n 1
done
```
Document the spread (e.g. "5 runs: 90.6 %, 87.9 %, 90.6 %, 90.2 %,
87.9 %"). If the flake reproduces, you have your blocker.

## 4. Identify the missed branches
```bash
go test -count=1 -cover -coverprofile=/tmp/lo.cov ./internal/catalogmodel/
go tool cover -func=/tmp/lo.cov | sort -k3 -n | head -20
# Re-run several times and diff the per-function coverage to find
# the function(s) whose coverage drops between runs. Those are the
# rapid-driven branches that need a deterministic test.
```

## 5. Apply the fix (smallest possible)
- Prefer adding 1–3 deterministic table-driven tests covering the
  identified branches.
- If the flake is in `Sanitize*`, you MUST keep `Sanitize*` at 100 %
  every run (C0 floor) — pinning a rapid seed in
  `Sanitize*` property tests is acceptable.
- Commit on the PR branch:
  ```
  git checkout impl/task-0024-c1-sourcectx-resolver
  git add internal/catalogmodel/<files>
  git commit -m "test(catalogmodel): deterministic coverage for C0 \
    floor (verifier-only, task-0024)"
  git push origin impl/task-0024-c1-sourcectx-resolver
  ```

## 6. Re-run local + CI gates
```bash
go build ./...
go vet ./...
go test -race -count=1 ./...
for i in 1 2 3; do make test-state-redesign; done   # all green, all runs
make verify-generated
kiox -- orun validate --intent intent.yaml
gh pr checks 169 --watch                            # block until green
```

## 7. Backfill PR number in the implementer report (if needed)
```bash
grep -n "PR:" ai/reports/task-0024-implementer.md
# If the PR line is still a placeholder, update to "PR: #169" and
# include in the same verifier commit.
```

## 8. Merge protocol (PASS only)
```bash
gh pr merge 169 --squash --delete-branch
git checkout main
git pull --ff-only origin main
git status --short                                  # must be empty
git log --oneline -1                                # squash commit visible
```

## 9. FAIL handling
If the catalogmodel flake cannot be eliminated within this verifier
pass — for example, the missed branch is genuinely owned by a
randomized rapid generator that needs a property-test redesign — do
NOT merge. Leave the PR open with:
- the verifier report committed to the PR branch
- the failing CI run linked
- a clear `Recommended Next Move` proposing a sub-task
  (`Task 0024.1: deterministic catalogmodel coverage`) and/or a
  `/ai/proposals/task-0024-catalogmodel-flake.md` file describing
  the C0 test-plan adjustment if a behavioral change is required.

# PR Creation Requirement
PR #169 already exists and is the verification target. Do not open a
new PR. The verifier may push at most one additional commit to the
existing branch (the verifier report + the minimal coverage fix).

# When Done Report
Save to `ai/reports/task-0024-verifier.md` with the standard
section set:

- **Result:** `PASS` or `FAIL`
- **Checks** — every command run in §1–§7 above with exit status.
- **CI Log Review** — link to the green PR-head run that replaced
  `26704817160`, plus the per-job conclusion list.
- **Coverage Adjudication** — the local 5-run spread before and
  after your fix, plus the `make test-state-redesign` 3-run
  determinism evidence.
- **Live Resource Evidence** — N/A (Phase 2 is local-only).
- **Issues** — anything you found that did not block merge but the
  orchestrator should track.
- **Spec Proposals** — link only, with one-line reason. Use
  `/ai/proposals/task-0024-*.md` if a real behavioral or contract
  change is needed; otherwise leave empty.
- **Risk Notes** — residual flake risk on other rapid-driven
  packages (`internal/sourcectx`, future C2/C3 packages) and any
  C0-floor-headroom recommendation.
- **Recommended Next Move** — name the next milestone (C2,
  `internal/catalogresolve`) and the orchestrator-side bookkeeping
  it expects.
- **PR Number:** `#169` plus merge SHA on `main` (PASS) or
  `OPEN — blocked` with a one-line reason (FAIL).

Keep the report under the orchestrator's preferred budget
(Summary 3–5 bullets, no full diffs). Do not paste full CI logs;
link them.
