# Task 0146.1

Agent: Verifier

## Current Repo Context

- Implementer Task 0146 completed on branch `impl/task-0146-plan-studio` and opened PR #145: `feat(tui): plan studio with real GeneratePlan (task-0146)`.
- PR #145 is open, not draft, mergeStateStatus is `CLEAN`, and CI checks are currently green: CI / Orun Plan succeeded and remote-state conformance / Harness dry-run guard succeeded; matrix jobs are skipped as expected for this diff shape.
- Task 0146 scoped the first coherent TUI Cockpit Phase 2 Plan Studio slice: real `LiveOrunService.GeneratePlan`, Plan Studio generate/review/error state, root-model routing, and focused service/view tests.
- Current branch includes the implementer report at `ai/reports/task-0146-implementer.md`. The verifier must inspect the PR diff and CI logs directly rather than relying only on the report.

## Objective

Verify PR #145 against Task 0146, the implementer report, `.kiro/specs/orun-tui-cockpit` Requirements 4/13/14, and the Verifier Standard in `agents/orchestrator.md`. If and only if verification and CI both PASS, merge PR #145, sync local `main`, and leave the repo clean.

## PR Boundary

The verifier must keep the review scoped to the exact Task 0146 boundary:

1. `LiveOrunService.GeneratePlan` calls Orun planning internals directly and returns `services.PlanResult` with plan/checksum/job/component/warning metadata.
2. `PlanStudioModel` supports Plan Studio form-ish request state, generate-in-flight state, review rendering, error recovery, save request dispatch, and focused navigation/tests.
3. Root `internal/tui/model.go` routes Plan Studio entry and `services.PlanGeneratedMsg` enough for `p` -> Plan Studio -> `g` -> generate/review/error.
4. Focused service/view/root tests cover cancellation, invalid inputs, request/config precedence, state transitions, review rendering, save dispatch, and property-style invariants.

No scope expansion: do not require Phase 3 execution, real `RunPlan`, `Describe`, command-palette completion, full graphical DAG rendering, remote-state execution behavior, or GitHub CLI UX follow-ups.

## Read First

- `agents/orchestrator.md` — Verifier Standard and Verifier Merge Protocol, especially lines 331-372.
- `ai/tasks/task-0146.md` — original implementer contract and acceptance criteria.
- `ai/reports/task-0146-implementer.md` — implementer summary, checks, and declared gaps.
- `.kiro/specs/orun-tui-cockpit/requirements.md` — Requirement 4, Requirement 5 non-goals/safety, Requirement 13, Requirement 14.
- `.kiro/specs/orun-tui-cockpit/design.md` — Architecture §3-6 and Components §3-4 for `OrunService` and `PlanStudioModel`.
- `.kiro/specs/orun-tui-cockpit/tasks.md` — Phase 2 tasks 14-16 and the rapid import-path drift note.
- PR #145 diff, commits, and GitHub Actions logs.

## Required Outcomes

- [ ] Inspect PR #145 diff, commits, and implementer report.
- [ ] Confirm the PR maps to exactly Task 0146 and does not include unrelated scope.
- [ ] Run local validation commands listed below.
- [ ] Inspect CI logs with `gh run view`, including successful jobs, and confirm expected commands actually ran.
- [ ] Inspect code paths for no `exec.Command`, no `os/exec`, and no literal shell-out path under `internal/tui/` for plan generation.
- [ ] Verify `GeneratePlan` service behavior is production-grade enough for this slice: request/config precedence, safe context checks, no stdout/logged secrets, nil store handling for `NamedPlan`, and clear warning for changed-only safe subset.
- [ ] Verify Plan Studio state-machine behavior, error recovery, save dispatch, and review rendering against Requirement 4 and Task 0146 acceptance criteria.
- [ ] Write `ai/reports/task-0146-verifier.md` with Result PASS or FAIL and evidence.
- [ ] If PASS and CI is green, merge PR #145, checkout `main`, fast-forward pull from `origin/main`, and leave `git status --short` clean.
- [ ] If FAIL, leave PR #145 open and document precise blockers in the verifier report.

## Non-Goals

- Do not implement fixes unless they are tiny verifier-only report/task/state corrections. Any code issue that affects acceptance should make verification FAIL and remain on the PR for implementer repair.
- Do not require `RunPlan`, dry-run execution from the TUI, live run dashboard, real `Describe`, follow-mode `TailLogs`, remote-state execution, or command-palette completion in this PR.
- Do not pick up `docs/github-log-pull-ux-review.md` section 3 follow-ups.
- Do not broaden the task into a TUI visual redesign or spec rewrite.

## Constraints

1. PR #145 may merge only when both local verification and PR CI/log inspection are acceptable.
2. The TUI service layer must call internal packages directly; shelling out to `orun` from `internal/tui/` is a blocker.
3. Plan Studio must not expose a key path that starts real execution in this slice; `RunPlan` is intentionally out of scope.
4. Secret safety: verifier reports may mention file paths, job names, and checksums but must not include credentials, tokens, signed URLs, or raw secret values.
5. Spec drift handling: if stale `github.com/flyingmutant/rapid` references remain only in `.kiro/specs/orun-tui-cockpit/tasks.md`, treat that as a documented non-blocking follow-up unless the PR reintroduces the stale import path in Go code.
6. If verification needs to add the verifier report or small state-file corrections to the PR branch, push them and wait for CI again before merging.

## Acceptance Criteria

✅ PR #145 corresponds exactly to Task 0146 and the implementer report.

✅ Local checks pass from repo root:

```bash
go test ./internal/tui/... -count=1
go test ./cmd/orun/... ./internal/planner/... ./internal/render/... -count=1
go build ./cmd/orun/...
```

✅ Orun validation/dry-run checks pass or verifier records an exact accepted no-op/blocker:

```bash
/Users/irinelinson/.local/bin/kiox -- orun validate --intent intent.yaml
/Users/irinelinson/.local/bin/kiox -- orun plan --changed --intent intent.yaml --output plan.json
/Users/irinelinson/.local/bin/kiox -- orun run --plan plan.json --dry-run --runner github-actions
```

✅ Code inspection confirms no `exec.Command`, `os/exec`, or CLI shell-out path was introduced under `internal/tui/` for plan generation.

✅ `LiveOrunService.GeneratePlan` returns a non-nil plan/checksum/job metadata on a valid fixture and returns errors for missing/invalid intent without panicking.

✅ Plan Studio tests prove generate -> review, generate error -> error/form recovery, clear/esc-style safe reset, save dispatch, deterministic rendering/property invariants, and no real run execution path.

✅ GitHub Actions logs for PR #145 are inspected and show required CI success; mergeStateStatus remains CLEAN before merge.

✅ If PASS, PR #145 is merged and local `main` is synced and clean. If FAIL, PR #145 remains open with blockers.

## Verification Steps

1. Confirm repo and PR state:

```bash
git status --short
gh pr view 145 --json number,title,state,headRefName,baseRefName,mergeStateStatus,isDraft,statusCheckRollup,commits,files,url
```

2. Inspect scope and shell-out safety:

```bash
git diff --stat origin/main...HEAD
git diff --name-status origin/main...HEAD
git diff origin/main...HEAD -- internal/tui ai/tasks/task-0146.md ai/reports/task-0146-implementer.md go.mod go.sum
grep -RInE 'exec\.Command|os/exec|"orun"' internal/tui || true
```

3. Run local checks:

```bash
go test ./internal/tui/... -count=1
go test ./cmd/orun/... ./internal/planner/... ./internal/render/... -count=1
go build ./cmd/orun/...
/Users/irinelinson/.local/bin/kiox -- orun validate --intent intent.yaml
/Users/irinelinson/.local/bin/kiox -- orun plan --changed --intent intent.yaml --output plan.json
/Users/irinelinson/.local/bin/kiox -- orun run --plan plan.json --dry-run --runner github-actions
```

4. Inspect CI logs, not just status summaries:

```bash
gh run view 26610789285 --json name,conclusion,status,jobs
gh run view 26610789289 --json name,conclusion,status,jobs
gh run view 26610789285 --log --job 78416011673
gh run view 26610789289 --log --job 78416011670
```

5. Write `ai/reports/task-0146-verifier.md` with the mandatory sections below.
6. If PASS, merge per `agents/orchestrator.md` Verifier Merge Protocol. If FAIL, leave PR open.

## When Done Report

Write `/ai/reports/task-0146-verifier.md` with these sections:

- Result: PASS or FAIL
- Checks
- Issues
- CI Log Review
- Scope / Overreach Review
- Secret Handling Review
- Spec Proposals
- Risk Notes
- Recommended Next Move
- PR Number
