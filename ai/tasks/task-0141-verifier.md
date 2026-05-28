# Task 0141.1

Agent: Verifier

## Current Repo Context

- Task 0141 Implementer completed and opened PR #141: https://github.com/sourceplane/orun/pull/141
- PR #141 branch: `impl/task-0141-runs-details` against `main`; current head is `986fac6` and base `origin/main` is `8753b95`.
- GitHub reports `mergeStateStatus: CLEAN`. Required visible checks are green: CI / Orun Plan run `26605237056` SUCCESS and remote-state conformance / Harness dry-run guard run `26605237077` SUCCESS; matrix execution jobs are skipped as expected for a no-op plan.
- Implementer report is present at `ai/reports/task-0141-implementer.md` and records PR Number `#141`.
- Task 0141 objective was Requirement 11 Level 2 remote inspection: implement `orun github runs --details` manifest-level status detail without full hydration, logs display, or Level 1 download cost.

## Objective

Verify PR #141 against Task 0141, GitHub Artifacts Requirement 11, Requirement 20 test expectations, the Phase 9 three-level artifact detail design, and the Verifier Standard in `agents/orchestrator.md`. If PASS and CI/log inspection are acceptable, merge PR #141, sync local `main`, and leave the repo clean. If FAIL, leave PR #141 open and document precise blockers.

## PR Boundary

Verification scope is exactly Task 0141 / PR #141:

1. Confirm `orun github runs --details` downloads manifest-level data for Orun artifact shards and prints exact manifest-derived role/status/exec-id/job/component/environment detail.
2. Confirm default `orun github runs` remains Level 1 and does not download shard ZIPs or manifests.
3. Confirm manifest download/read failures degrade with a warning for that shard and do not hide the workflow run or abort unrelated shards.
4. Confirm focused tests and narrow specs/docs status updates match Requirement 11 only.

Explicit non-expansion: do not implement or request new code for partial hydration display (Requirement 10), E2E workflow/root workflow decisions (Requirements 17/21), `orun github pull`, `orun github logs`, TUI cockpit, or broad artifact backend refactors. Verification-only report commits are allowed only if the PR otherwise passes and must be pushed to the PR branch before merge.

## Read First

- `agents/orchestrator.md` — Verifier Standard and Verifier Merge Protocol, especially lines/sections covering PASS/FAIL, CI log inspection, merge, sync main, and clean worktree.
- `ai/tasks/task-0141.md` — original Implementer prompt and acceptance criteria.
- `ai/reports/task-0141-implementer.md` — claimed implementation, checks, assumptions, and PR number.
- `.kiro/specs/github-artifacts/requirements.md` — Requirement 11 and Requirement 20 expectations. Note: PR #141 adds/updates this file; verify the added content is appropriate and not an accidental broad spec import.
- `.kiro/specs/github-artifacts/design.md` — Three-Level Artifact Detail Model and Level 2 status rows.
- PR #141 diff and commits via `gh pr diff 141`, `gh pr view 141 --json ...`, and local `git diff origin/main...HEAD` if on the PR branch.
- Changed code: `cmd/orun/command_github.go`, `cmd/orun/command_github_test.go`, `internal/artifactstore/github/manifest.go`, `internal/artifactstore/github/manifest_test.go`.

## Required Outcomes

- [ ] Verifier report written to `ai/reports/task-0141-verifier.md` with Result: PASS or FAIL.
- [ ] PR #141 is inspected against the implementer prompt, not only the report.
- [ ] Local checks listed below are run and recorded with exact PASS/FAIL results.
- [ ] GitHub Actions logs for the successful checks are inspected, not only status summaries.
- [ ] Secret-handling and path traversal implications of the new manifest helper are reviewed.
- [ ] If PASS: PR #141 is merged, local `main` is checked out, fast-forwarded from `origin/main`, and the local repo is left clean except for known untracked orchestration/spec files that pre-existed this verifier task.
- [ ] If FAIL: PR #141 remains open and the verifier report clearly lists blockers and recommended fixes.

## Non-Goals

- No new feature implementation beyond verification-only fixes if absolutely necessary.
- No changes to `orun github pull`, `orun github logs`, upload helper, RunBundle schema, or artifact naming semantics unless required to fix a Task 0141 acceptance blocker.
- No live GitHub artifact mutation or external deployment.
- No TUI cockpit task scoping in this verifier pass.
- No acceptance of failing CI or unresolved verification blockers.

## Constraints

1. Merge only if both local verification and PR CI/log inspection are acceptable.
2. Never merge PR #141 if `mergeStateStatus` is not clean/up-to-date enough for the repository merge policy or if required checks fail.
3. Level 1 must stay cheap: verifier must explicitly confirm the non-`--details` path does not call `Download`, `DownloadByName`, or `DownloadManifestOnly`.
4. Level 2 must not print or intentionally read `log:*` file contents for display; logs remain owned by `orun github logs` / Level 3 paths.
5. Manifest-only helper must retain existing ZIP path traversal defenses and must clean up temp files.
6. No secrets may be printed in reports, CI log excerpts, or command output. Do not paste tokens, signed URLs, or request headers.
7. Treat the `.kiro/specs/github-artifacts/requirements.md` addition/update as part of the PR diff: verify it is an intended, bounded Requirement 11 status update and not unrelated scope creep.

## Acceptance Criteria

✅ PR #141 maps exactly to Task 0141 and the Implementer report; no unrelated roadmap scope is included.

✅ Local diff/code inspection confirms:
- `runGithubRuns()` calls manifest detail logic only when `githubRunsDetails` is true.
- `printManifestDetails()` sorts shards deterministically, uses `DownloadManifestOnly()`, prints manifest-derived fields, and warns/continues on per-shard failure.
- `DownloadManifestOnly()` uses existing GitHub download/extract behavior with path traversal defense, reads a parsed manifest, and removes its temp directory.
- The code does not intentionally read or print log file contents in `runs --details`.

✅ Focused and broad local tests pass:
```bash
go test ./cmd/orun/ -run 'TestGithubRuns|TestGithubCommandRunsHelp' -v
go test ./internal/artifactstore/github/... -v
go test ./internal/runbundle/... -v
go test ./cmd/orun/... -v
go build ./cmd/orun/
git diff --check origin/main...HEAD
```

✅ If no root `intent.yaml` exists, Orun validation is recorded as not applicable. If `intent.yaml` exists, run:
```bash
/Users/irinelinson/.local/bin/kiox -- orun validate --intent intent.yaml
/Users/irinelinson/.local/bin/kiox -- orun plan --changed --intent intent.yaml --output plan.json
/Users/irinelinson/.local/bin/kiox -- orun run --plan plan.json --dry-run --runner github-actions
```

✅ GitHub Actions evidence is inspected:
```bash
gh pr view 141 --json number,title,state,mergeStateStatus,statusCheckRollup,commits,files
gh run view 26605237056 --log
gh run view 26605237077 --log
```
Confirm logs show expected Orun plan/remote-state commands where applicable and no hidden failures.

✅ Spec/docs updates are limited to Requirement 11 / Level 2 current-status rows unless the verifier finds a justified, documented reason for broader changes.

✅ If PASS, verifier follows the merge protocol:
```bash
gh pr merge 141 --merge --delete-branch
git checkout main
git pull --ff-only origin main
git status --short
```
Use the repository's actual merge method if `gh` reports a different required strategy; do not force merge.

## Verification

Verifier procedure:

1. Confirm worktree state and branch:
```bash
git status --short --branch
git rev-parse --short HEAD
git rev-parse --short origin/main
```
2. Inspect PR metadata, files, commits, and CI status:
```bash
gh pr view 141 --json number,title,state,url,headRefName,baseRefName,mergeStateStatus,isDraft,commits,files,statusCheckRollup
```
3. Inspect implementation code and tests directly, not only summaries.
4. Run all local checks in Acceptance Criteria.
5. Inspect CI logs for run `26605237056` and run `26605237077`.
6. Review secret safety with targeted searches for token/header/signed URL leakage in changed files and reports.
7. Decide PASS or FAIL. On PASS, merge and sync main. On FAIL, leave PR open.
8. Write `ai/reports/task-0141-verifier.md` and update orchestration state files to reflect the result.

## PR Creation Requirement

The Implementer has already created PR #141. The Verifier must not create a new implementation PR. If verification-only artifacts need to be committed before merge, commit them to `impl/task-0141-runs-details`, push, wait for CI, then continue verification.

## When Done Report

Write `/ai/reports/task-0141-verifier.md` with these sections:

- `Result: PASS` or `Result: FAIL`
- `Checks` — exact commands run and outcomes
- `PR/CI Evidence` — PR #141 metadata, CI run IDs, log inspection summary
- `Code Review Notes` — Level 1/Level 2 behavior, manifest helper, warning/degradation behavior, deterministic output
- `Secret and Safety Review` — tokens/logs/path traversal/temp cleanup
- `Spec/Docs Review` — Requirement 11/design changes, any spec drift
- `Issues` — blockers or non-blocking concerns
- `Risk Notes` — residual risks
- `Recommended Next Move` — merge completed or fixes required; if PASS, likely next orchestrator task after merge
