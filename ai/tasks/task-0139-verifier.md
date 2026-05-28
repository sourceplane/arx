# Task 0139.1

Agent: Verifier

## Current Repo Context

- PR #139 (`fix/plan-upload-token-gate`) is open against `main` and has `mergeStateStatus: CLEAN`.
- PR #139 fixes GitHub Actions artifact upload for Orun by exporting the GitHub Actions runtime artifact variables from a JavaScript action context and by removing a brittle Go-side `ACTIONS_RUNTIME_TOKEN` preflight gate.
- CI is currently green for the PR: `CI / Orun Plan` passed and `orun remote-state conformance / Harness dry-run guard` passed. Matrix execute jobs were skipped because the changed-only plan produced 0 jobs.
- The repository did not have an existing `ai/` orchestration state directory in this checkout, so this verifier task initializes the AI state files for the active PR cycle.

## Objective

Verify PR #139 against the GitHub Artifacts requirements and production-grade standards, then decide PASS or FAIL. If PASS and CI/log inspection are acceptable, merge PR #139, sync local `main`, and leave the repo clean. If FAIL, leave the PR open and document exact blockers.

## PR Boundary

Verify exactly this PR scope:

1. GitHub Actions workflow exports `ACTIONS_RUNTIME_TOKEN` and `ACTIONS_RESULTS_URL` through `actions/github-script@v7` before Orun upload steps need them.
2. `internal/artifactstore/github.Upload` no longer blocks purely because `ACTIONS_RUNTIME_TOKEN` is absent before invoking the Node helper; helper stdout/stderr are separated so stdout remains parseable JSON.
3. `UploadShard` attempts helper-based upload inside GitHub Actions and returns clear non-GitHub-Actions guidance outside GitHub Actions.
4. `orun github pull` reads full plan/job shards via `ReadPlanShard` / `ReadJobShard` before synthesis/hydration, preventing nil plan/job state during hydration.

Explicit non-goals:

- Do not expand `orun github runs --details` Level 2 implementation.
- Do not implement `orun github logs` content streaming gaps.
- Do not broaden GitHub Artifacts docs/specs beyond what is needed to validate PR #139.
- Do not start the new `orun tui` cockpit implementation.

## Read First

- `agents/orchestrator.md` — Verifier Standard and Verifier Merge Protocol, especially lines 331-372.
- `.kiro/specs/github-artifacts/requirements.md` — Requirement 7 (GitHub Artifact Upload via Embedded Helper), Requirement 10 (Synthesis and Hydration), Requirement 12 (`orun github pull`).
- `.kiro/specs/github-artifacts/design.md` — GitHub artifact upload architecture and error handling sections.
- PR #139 body, commits, diff, checks, and logs:
  - `gh pr view 139 --json number,title,body,headRefName,baseRefName,state,mergeStateStatus,statusCheckRollup,files,commits,url`
  - `gh pr diff 139`
  - `gh pr checks 139`

## Required Outcomes

- [ ] Inspect PR #139 diff and confirm it maps exactly to the PR boundary above.
- [ ] Run local targeted tests:
  - `go test ./internal/artifactstore/github/... ./cmd/orun/... ./internal/runbundle/...`
- [ ] Run formatting/whitespace checks:
  - `git diff --check main...HEAD`
  - If this fails only because of unrelated untracked/local files outside PR #139, record it separately; do not block PR #139 for files not in the PR diff.
- [ ] Inspect successful CI logs, not only status summaries:
  - Confirm the `Export Actions runtime token` step ran before `orun plan` in run `26564009651`.
  - Confirm the `orun plan --artifact github --github-output` step saw redacted `ACTIONS_RUNTIME_TOKEN` and `ACTIONS_RESULTS_URL` in its environment.
  - Confirm the plan artifact upload succeeded with a line like `✓ uploaded plan artifact: orun.v1....plan...created`.
- [ ] Confirm `orun github pull` hydration path now reads full plan and job shard content with `ReadPlanShard` / `ReadJobShard` rather than constructing partial structs with nil `Plan`/`JobState` fields.
- [ ] Review secret handling: logs and code must not print runtime token values; redaction in CI logs is acceptable.
- [ ] Write verifier report to `ai/reports/task-0139-verifier.md`.
- [ ] If PASS: merge PR #139, checkout `main`, fast-forward pull from `origin/main`, and leave local repo clean or with only unrelated pre-existing local work clearly documented.
- [ ] If FAIL: leave PR #139 open and clearly document blockers in the verifier report.

## Non-Goals

- No implementation of remaining GitHub Artifacts roadmap gaps.
- No modification to unrelated untracked `orun-tui-cockpit` specs or orchestrator files except AI state/report artifacts required for this verification cycle.
- No force-push or history rewrite.
- No cloud/resource mutation beyond the normal PR merge action if verification passes.

## Constraints

1. Do not merge unless both local verification and PR CI/log inspection are acceptable.
2. Never merge with unresolved verification blockers.
3. Do not expose `ACTIONS_RUNTIME_TOKEN`, `ACTIONS_RESULTS_URL`, `GITHUB_TOKEN`, or any other token in reports or logs.
4. Treat skipped matrix jobs as acceptable only if the plan legitimately had 0 changed jobs; verify this from the `Orun Plan` log.
5. The verifier may commit `ai/reports/task-0139-verifier.md` and orchestration state/report updates if needed, but verification-only commits must go to the PR branch first and CI must be rechecked before merge.
6. The current checkout has pre-existing untracked orchestration/spec files; distinguish PR #139 tracked changes from local untracked work before deciding cleanliness.

## Acceptance Criteria

✅ PR #139 has `mergeStateStatus: CLEAN` and all required checks are passing or intentionally skipped for no-job matrix fanout.

✅ Local targeted tests pass:

```bash
go test ./internal/artifactstore/github/... ./cmd/orun/... ./internal/runbundle/...
```

✅ CI log inspection confirms the runtime token export step ran and the subsequent `orun plan` step had redacted `ACTIONS_RUNTIME_TOKEN` / `ACTIONS_RESULTS_URL` environment variables.

✅ CI log inspection confirms plan artifact upload actually succeeded, not merely that the workflow passed.

✅ Code inspection confirms helper stdout is isolated from `@actions/artifact` logging and JSON parsing uses only stdout.

✅ Code inspection confirms upload no longer uses the broken `ACTIONS_ID_TOKEN_REQUEST_TOKEN` fallback as a substitute for `ACTIONS_RUNTIME_TOKEN`.

✅ Code inspection confirms `orun github pull` uses `runbundle.ReadPlanShard` and `runbundle.ReadJobShard` before synthesis/hydration.

✅ No token values or full credentials are committed or included in verifier report output.

✅ If PASS, verifier merges the PR and syncs local `main`; if FAIL, verifier leaves PR open and documents blockers.

## Verification

Suggested command sequence:

```bash
# Inspect PR metadata and status
gh pr view 139 --json number,title,headRefName,baseRefName,state,mergeStateStatus,statusCheckRollup,files,commits,url
gh pr checks 139

# Inspect diff
gh pr diff 139

git diff --stat main...HEAD
git diff --name-status main...HEAD

# Run targeted tests
go test ./internal/artifactstore/github/... ./cmd/orun/... ./internal/runbundle/...

# Formatting/whitespace check for PR diff
git diff --check main...HEAD

# CI log evidence
gh run view 26564009651 --job 78253885149 --log
gh run view 26564009654 --job 78253885257 --log
```

If PASS:

```bash
gh pr merge 139 --squash --delete-branch
git checkout main
git pull --ff-only origin main
git status --short
```

If `gh pr merge` reports branch protection or merge queue requirements, document the exact blocker in the verifier report and do not claim PASS+merged until the PR is actually merged and local main is synced.

## PR Creation Requirement

The Implementer has already created PR #139. Your job is to verify it. Do not create a new implementation PR unless verifier-only report/state changes are required by repository process.

## When Done Report

Write `/ai/reports/task-0139-verifier.md` with these sections:

- `Result: PASS` or `Result: FAIL`
- `PR`: PR number, URL, branch, merge state
- `Summary`
- `Checks`
- `CI Log Review`
- `Code Path Review`
- `Secret Handling Review`
- `Issues`
- `Risk Notes`
- `Spec Proposals`
- `Recommended Next Move`

If PASS and merged, include the merge result and synced `main` commit. If FAIL, list blockers and leave PR #139 open.
