# Task 0140.1

Agent: Verifier

## Current Repo Context

- Task 0140 implementer work is complete enough for verification: PR #140 (`Task 0140: Implement orun github logs content display`) is open against `main` from branch `impl/task-0140-github-logs-content`.
- PR #140 contains exactly three changed files by GitHub diff: `cmd/orun/command_github.go`, `cmd/orun/command_github_test.go`, and `ai/reports/task-0140-implementer.md`.
- PR #140 is not a draft, has `mergeStateStatus: CLEAN`, and its latest head commit is `a6c073d0794486960034d922046737e434fe1d61` (`docs: add task-0140 implementer report (PR #140)`).
- PR CI is currently green by GitHub checks:
  - `CI` run `26604189097`: `Orun Plan` passed; matrix job skipped because changed-only plan had 0 jobs.
  - `orun remote-state conformance` run `26604189095`: `Harness dry-run guard` passed; downstream remote-state jobs skipped.
- The implementer report is present at `ai/reports/task-0140-implementer.md` and records real PR number `#140`.

## Objective

Verify PR #140 against Task 0140, GitHub Artifacts Requirement 14, focused Requirement 20 test coverage, and the Verifier Standard in `agents/orchestrator.md`. If verification passes and PR CI remains green, merge PR #140, sync local `main`, and leave the repository clean.

## PR Boundary

Verification is limited to the Task 0140 implementation boundary:

1. `orun github logs` should print actual log file contents from downloaded job shards.
2. Printed log sections should use headers in the format `=== {shard-name} / {step-id} ===`.
3. Only manifest entries whose logical names start with `log:` should be printed as log content.
4. Individual unreadable/missing log files should warn to stderr and continue.
5. Existing `--job` filtering and no-match error behavior should remain intact.
6. Focused tests should prove the above without live GitHub API access.

No scope expansion: do not implement or request `orun github runs --details`, new workflows, TUI cockpit work, artifact schema changes, upload-helper changes, or unrelated cleanup as part of this verification.

## Read First

- `agents/orchestrator.md` — Verifier Standard and Verifier Merge Protocol, especially lines 331-372.
- `ai/tasks/task-0140.md` — original implementer contract.
- `ai/reports/task-0140-implementer.md` — implementer summary, checks, assumptions, and PR number.
- `.kiro/specs/github-artifacts/requirements.md` — Requirement 14 and Requirement 20.
- `.kiro/specs/github-artifacts/design.md` — Phase 9 log-content gap and Security Considerations.
- PR #140 diff, commits, and CI logs via `gh pr view 140`, `gh pr diff 140`, `gh pr checks 140`, and `gh run view`.
- `cmd/orun/command_github.go` and `cmd/orun/command_github_test.go` on the PR branch.

## Required Outcomes

- [ ] Confirm PR #140 maps exactly to Task 0140 and does not include unrelated feature scope.
- [ ] Review `runGithubLogs()` and `printShardLogs()` code paths for correctness, warning behavior, and path containment.
- [ ] Confirm only `log:*` manifest entries are printed and non-log files are ignored.
- [ ] Confirm per-step headers and actual log content are produced by tests.
- [ ] Confirm unreadable or missing log files warn and continue without aborting the entire command.
- [ ] Confirm path traversal is blocked defensively even though download/manifest validation already exists.
- [ ] Confirm existing `--job` no-match behavior still returns an error.
- [ ] Run required local checks and record exact results.
- [ ] Inspect GitHub Actions logs, not just status summaries, for PR #140.
- [ ] Confirm no token/env dumping or credential exposure was introduced.
- [ ] Write `ai/reports/task-0140-verifier.md` with PASS or FAIL.
- [ ] If PASS and CI is green, merge PR #140, checkout `main`, fast-forward pull from `origin/main`, and ensure `git status --short` is clean except for intentional orchestration files/report updates that are committed.
- [ ] If FAIL, leave PR #140 open and document blockers clearly.

## Non-Goals

- Do not implement `orun github runs --details` / Level 2 manifest downloads.
- Do not add GitHub Actions workflows or live E2E coverage.
- Do not alter artifact upload helper behavior, token export, naming, retention, or shard schema.
- Do not start or modify the `.kiro/specs/orun-tui-cockpit/` implementation.
- Do not broaden tests or refactor unrelated GitHub CLI paths unless required to fix a verification blocker in PR #140.

## Constraints

1. Follow the Verifier Merge Protocol in `agents/orchestrator.md`: merge only after local checks and PR CI logs are both acceptable.
2. Never merge with unresolved verification blockers or failing required CI checks.
3. Verification fixes, if any, must stay on the PR branch and must be pushed; then re-check CI before merging.
4. Treat log output as potentially sensitive user CI output. The implementation must not add token/env dumping or log credentials beyond already-stored log contents.
5. Path containment should be validated as defense-in-depth using absolute paths under `DownloadedShard.Dir`.
6. The verifier report must be committed with the verification/state updates according to repo rules after merge, or pushed to the PR branch if verification-only changes are required before merge.
7. Keep this verifier task focused on PR #140. Remaining roadmap gaps become future orchestrator tasks after PASS+merge.

## Acceptance Criteria

✅ PR metadata is acceptable:

```bash
gh pr view 140 --json number,title,state,isDraft,mergeStateStatus,headRefName,baseRefName,commits,files,statusCheckRollup,url
```

Expected: PR is open, not draft, targets `main`, `mergeStateStatus` is `CLEAN`, and changed files stay within Task 0140 scope.

✅ Local checks pass:

```bash
go test ./cmd/orun/... -run 'TestGithubLogs' -count=1
go test ./internal/runbundle/... ./internal/artifactstore/github/... ./cmd/orun/... -count=1
go test ./... 
git diff --check
```

✅ Orun validation applicability is recorded:

```bash
/Users/irinelinson/.local/bin/kiox -- orun validate --intent intent.yaml
/Users/irinelinson/.local/bin/kiox -- orun plan --changed --intent intent.yaml --output plan.json
/Users/irinelinson/.local/bin/kiox -- orun run --plan plan.json --dry-run --runner github-actions
```

If root `intent.yaml` is absent or local setup makes changed-plan validation inapplicable, record the exact reason. If validation succeeds despite no root file, record that observed behavior precisely and do not invent a failure.

✅ Code inspection proves:

- `runGithubLogs()` still resolves and filters target shards as before.
- `printShardLogs()` prints only logical names with `log:` prefix.
- Headers match `=== {shard-name} / {step-id} ===`.
- The implementation reads actual file contents via resolved manifest-relative paths under the extracted shard dir.
- Escape attempts are skipped with warnings.
- Missing/unreadable log files warn and continue.
- Non-log manifest entries are not printed.

✅ CI log inspection proves:

```bash
gh pr checks 140 --watch=false
gh run view 26604189097 --job 78395115737 --log
gh run view 26604189095 --job 78395115682 --log
```

Expected: `Orun Plan` and `Harness dry-run guard` passed on PR head `a6c073d0794486960034d922046737e434fe1d61`; logs show expected build/plan/harness commands and no introduced secret exposure.

✅ Final decision:

- PASS: Merge PR #140, sync local `main`, clean up branch if appropriate, write/commit verifier report and state updates, and leave repo clean.
- FAIL: Leave PR #140 open, write verifier report with blockers, and do not mark Task 0140 complete.

## Verification

Suggested verifier workflow:

1. Ensure you are on the PR branch or fetch/check out PR #140 safely:
   ```bash
   gh pr checkout 140
   git status --short --branch
   ```
2. Read `ai/tasks/task-0140.md` and `ai/reports/task-0140-implementer.md`.
3. Inspect PR metadata and diff:
   ```bash
   gh pr view 140 --json number,title,state,isDraft,mergeStateStatus,headRefName,baseRefName,commits,files,statusCheckRollup,url
   gh pr diff 140 --stat
   gh pr diff 140 --patch
   ```
4. Run the local checks from Acceptance Criteria.
5. Inspect CI logs using `gh run view`, including successful jobs.
6. Audit for accidental credential exposure in changed files and logs. At minimum inspect for new token/env dumping in the changed code and review log output around env blocks for redaction.
7. Decide PASS or FAIL.
8. Write `ai/reports/task-0140-verifier.md`.
9. If PASS and CI is still green, merge PR #140 and sync `main`:
   ```bash
   gh pr merge 140 --merge --delete-branch
   git checkout main
   git pull --ff-only origin main
   git status --short
   ```
   Use the repository's established merge strategy if `--merge` is rejected by branch protection.
10. Update orchestration state after merge: mark Task 0140/0140.1 complete, note merge commit and CI evidence, and scope the next task or return to the orchestrator for the next implementer prompt.

## PR Creation Requirement

The Implementer has already created PR #140. Your job is to verify it. Do not open a new implementation PR unless a verification-only fix is required and cannot be pushed to the existing PR branch.

## When Done Report

Write `ai/reports/task-0140-verifier.md` with these sections:

- Result: PASS or FAIL
- Summary
- Checks
- CI Log Review
- Code Path Review
- Secret Handling Review
- Issues
- Risk Notes
- Spec Proposals
- Recommended Next Move

A PASS report must include the PR merge result and final local repo status. A FAIL report must list exact blockers and leave PR #140 open.
