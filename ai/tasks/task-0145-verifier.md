# Task 0145.1

Agent: Verifier

## Current Repo Context

- Task 0145 implementer work is complete on PR #144: `fix(github): normalize --orun-dir and add status selectors (supersedes #142)`.
- PR #144 head branch is `impl/task-0145-github-cli-pr142-supersede`; head commit at scoping time is `7209c86` (`docs: finalize task-0145 implementer report (PR #144, #142 closed)`).
- PR #144 is open, non-draft, and `mergeStateStatus: CLEAN` at scoping time. Required checks reported SUCCESS for `CI / Orun Plan` and `orun remote-state conformance / Harness dry-run guard`; downstream matrix jobs were SKIPPED as expected for a CLI/docs-only diff.
- Dirty PR #142 (`happy-patch-113`) is CLOSED as superseded by PR #144. Task 0142 had previously failed PR #142 for queued CI, dummy trigger scope, and unrelated TUI/process/history files.
- Repo health remains yellow until PR #144 is independently verified and merged. Do not advance to TUI Phase 2 until this verifier task reaches PASS and main is synced.

## Objective

Verify PR #144 against Task 0145, the implementer report, the Task 0142 failure findings, and the Verifier Standard in `agents/orchestrator.md`. If and only if verification and PR CI both PASS, merge PR #144, sync local `main`, leave the repo clean, and file the verifier report.

## PR Boundary

Verify exactly one PR: PR #144.

The allowed implementation scope is the clean successor to PR #142 containing only:

1. `cmd/orun/command_github.go` — `--orun-dir` normalization and `github status` selector flag registration.
2. `cmd/orun/command_github_test.go` — focused tests for parent-dir and already-`.orun` normalization plus status selector flag parsing/registration.
3. `website/docs/cli/orun-github.md` — docs matching the final CLI behavior.
4. `docs/github-log-pull-ux-review.md` — optional, direct UX-review note for this exact CLI/docs fix only.
5. `ai/reports/task-0145-implementer.md` — implementer report with real PR number and PR #142 disposition.

No TUI Phase 2, TUI spec/process/history cleanup, dummy component trigger, artifact schema/workflow change, or broad GitHub Artifacts feature work may be included.

## Read First

- `agents/orchestrator.md` — Verifier Standard and Verifier Merge Protocol, especially lines 331-372.
- `ai/tasks/task-0145.md` — original implementer contract and acceptance criteria.
- `ai/reports/task-0145-implementer.md` — implementer self-report, checks, PR #142 disposition.
- `ai/reports/task-0142-verifier.md` — previous FAIL report for PR #142; use it to confirm the old blockers are absent from PR #144.
- PR #144 metadata, diff, commits, and CI logs:
  - `gh pr view 144 --json number,title,body,headRefName,baseRefName,state,isDraft,mergeStateStatus,statusCheckRollup,commits,files,url`
  - `gh pr diff 144 --name-only`
  - `gh pr diff 144`
- PR #142 disposition:
  - `gh pr view 142 --json number,title,state,closed,closedAt,url,headRefName,baseRefName`
- Relevant code/docs:
  - `cmd/orun/command_github.go`
  - `cmd/orun/command_github_test.go`
  - `website/docs/cli/orun-github.md`
  - `docs/github-log-pull-ux-review.md`

## Required Outcomes

- [ ] Confirm PR #144 corresponds exactly to Task 0145 and contains no unrelated scope.
- [ ] Confirm PR #142 is closed as superseded, or FAIL with a clear blocker if that is not true.
- [ ] Run the local focused tests/build and record results.
- [ ] Inspect PR #144 CI logs, not just summary status; confirm expected checks actually ran and no required check is failing/queued/unknown.
- [ ] Review code behavior for `normalizeOrunDir()` and `github status` selector flags.
- [ ] Confirm docs accurately describe `--orun-dir`, status selectors, resolution order, full-SHA caveat, and job-filter caveat.
- [ ] Confirm no secrets, signed artifact URLs, dummy triggers, unrelated TUI specs/process docs/historical prompts, or stale orchestration state are included.
- [ ] Write `ai/reports/task-0145-verifier.md` with Result, Checks, Issues, Risk Notes, Spec Proposals, Recommended Next Move, PR/CI evidence, and merge outcome.
- [ ] On PASS only: merge PR #144, checkout `main`, fast-forward pull from `origin/main`, ensure `git status --short` is clean, then commit the verifier report/state-file closure to `main` if repo workflow requires post-merge orchestration artifacts.
- [ ] On FAIL: leave PR #144 open, do not merge, and document exact blockers in the verifier report.

## Non-Goals

- Do not implement fixes unless they are tiny verification-only corrections required to complete Task 0145; if you change PR files, commit them to the PR branch, push, and wait for CI again before merging.
- Do not start TUI Phase 2 or scope Task 0146 inside this verifier task.
- Do not revive or repair PR #142; it should remain closed as superseded unless you discover evidence that PR #144 did not fully replace it.
- Do not broaden GitHub Artifacts functionality beyond the Task 0145 CLI/docs/tests surface.
- Do not merge with failing, queued, cancelled, or unknown required CI.

## Constraints

1. Trust code reality and PR diffs over reports. The implementer report is evidence, not proof.
2. Merge only when both local verification and GitHub CI/log inspection are acceptable.
3. The old PR #142 blockers must be absent from PR #144:
   - `.kiro/specs/orun-tui-cockpit/**`
   - `orun-tui-cockpit.md`
   - `agents/orchestrator.md`
   - historical task prompt additions from PR #142
   - stale `ai/waiting_for_input.md`
   - `examples/apps/api-edge/component.yaml` dummy trigger
4. The string `pr-142-dummy-change` must not appear in product/config files or the PR diff. Historical reports/prompts that describe the old failure are allowed.
5. Do not expose GitHub tokens, bearer headers, signed artifact URLs, or credentials in the report.
6. Use `/Users/irinelinson/.local/bin/kiox` for Orun validation if a root `intent.yaml` exists; otherwise record Orun validation as N/A with the reason.

## Verification Commands

Run from the repo root after checking out the PR branch or using `gh pr checkout 144`:

```bash
git status --short --branch
gh pr view 144 --json number,title,state,isDraft,mergeStateStatus,statusCheckRollup,commits,files,url
gh pr diff 144 --name-only
gh pr view 142 --json number,title,state,closed,closedAt,url
```

Diff-scope blocker guard:

```bash
gh pr diff 144 --name-only | grep -E '^(\.kiro/specs/orun-tui-cockpit/|orun-tui-cockpit\.md|agents/orchestrator\.md|ai/tasks/task-0139-verifier\.md|ai/tasks/task-0140\.md|ai/tasks/task-0140-verifier\.md|ai/tasks/task-0141-verifier\.md|ai/waiting_for_input\.md|examples/apps/api-edge/component\.yaml)$' && exit 1 || true
```

Dummy-trigger guard:

```bash
git grep -n 'pr-142-dummy-change' -- . ':!ai/reports/task-0142-verifier.md' ':!ai/tasks/task-0143.md' ':!ai/tasks/task-0145.md' ':!ai/tasks/task-0145-verifier.md' ':!ai/context/task-ledger.md' ':!ai/state.json'
```

Expected: no matches outside historical reports/prompts/state entries that mention the old blocker.

Focused local checks:

```bash
go test ./cmd/orun/ -run 'TestGithub(Status|Pull|Logs|Runs)|TestGithubCommand|TestNormalizeOrunDir' -v -count=1
go test ./internal/artifactstore/github/... -count=1
go test ./internal/runbundle/... -count=1
go test ./cmd/orun/... -count=1
go build ./cmd/orun/
```

Orun validation fallback:

```bash
if test -f intent.yaml; then
  /Users/irinelinson/.local/bin/kiox -- orun validate --intent intent.yaml
  /Users/irinelinson/.local/bin/kiox -- orun plan --changed --intent intent.yaml --output plan.json
  /Users/irinelinson/.local/bin/kiox -- orun run --plan plan.json --dry-run --runner github-actions
else
  echo "No root intent.yaml; Orun validation not applicable for this repo checkout."
fi
```

CI log inspection examples:

```bash
gh run view 26609629520 --log --job 78412444147 | sed -n '1,220p'
gh run view 26609629547 --log --job 78412444132 | sed -n '1,220p'
```

If run/job IDs have changed, derive current IDs from `gh pr view 144 --json statusCheckRollup` and inspect the current successful jobs instead.

## Acceptance Criteria

✅ PR #144 is open, non-draft, targets `main`, and is `mergeStateStatus: CLEAN` or otherwise mergeable after up-to-date checks.

✅ PR #144 diff is limited to the Task 0145 allowed files and contains no TUI specs/process docs/history prompts/stale waiting input/dummy trigger files.

✅ PR #142 is CLOSED as superseded and no longer an open repo-health risk.

✅ `normalizeOrunDir()` preserves the intended semantics:
- empty input defaults to `.orun`;
- parent directory input resolves to `<parent>/.orun`;
- already-`.orun` input remains unchanged.

✅ `orun github status` registers and parse-accepts `--run-id`, `--exec-id`, `--sha`, `--branch`, `--latest`, and `--failed`.

✅ Local focused Go tests and build pass, or any failure is clearly unrelated and justified. Task-scoped failures are blockers.

✅ PR CI logs show the relevant checks ran and succeeded; no required check is failing, queued, cancelled, or unknown.

✅ No secrets, signed URLs, credentials, or token-bearing logs are committed or copied into the verifier report.

✅ If PASS: PR #144 is merged, local `main` is synced to `origin/main`, and the worktree is clean before reporting complete.

## Verification Decision

- PASS: only if all acceptance criteria above are satisfied. Merge PR #144, sync `main`, file the verifier report, and update orchestration state to mark 0145/0145.1 complete and unlock Task 0146.
- FAIL: if PR scope is dirty, PR #142 is not closed, required tests fail due to task changes, CI is not green, docs contradict code, or secret/dummy-trigger risk is found. Leave PR #144 open and document blockers.

## When Done Report

Write `/ai/reports/task-0145-verifier.md` with these sections:

- Result: PASS or FAIL
- Summary
- PR / Branch / Merge Evidence
- Checks Run
- CI Log Review
- Scope / Overreach Review
- Code Behavior Review
- Docs Review
- PR #142 Disposition
- Secret Handling Review
- Issues
- Risk Notes
- Spec Proposals
- Recommended Next Move
- Final Repo State
