# Current Orchestration Context

Last updated: 2026-05-29 (Task 0143 scoped — repair/narrow PR #142)

## Repo Reality

- Local branch: `main` at `813e682` (`docs: Task 0142 verifier report — PR #142 FAIL, left OPEN`), synced with `origin/main`.
- One PR open: **#142** (`happy-patch-113`, title `chore: update happy-patch-113`) — verified FAIL, not merged.
- PR #142 CI remains unhealthy: `Orun Plan` is QUEUED/unknown and `mergeStateStatus = UNSTABLE`; `Harness dry-run guard` succeeded, downstream remote-state conformance jobs are skipped.
- Repo health: **yellow** because PR #142 contains valid CLI work but is blocked by CI and unrelated scope.

## Last Completed Task (0141.1)

Task 0141 verified PASS and merged via PR #141 at `1ebcb46`. `orun github runs --details` now downloads manifest-only data for each Orun shard and prints Level 2 detail: role, exec-id, status, job, component, environment. Default `orun github runs` remains Level 1 (no downloads). GitHub Artifacts Requirement 11 satisfied.

## Last Verification (0142 — FAIL)

Task 0142 verifier inspected PR #142 and wrote `ai/reports/task-0142-verifier.md`.

**Result: FAIL.** The CLI code change is correct and locally tested, but the PR must not merge because:

1. `Orun Plan` CI is queued/unknown and `mergeStateStatus = UNSTABLE`.
2. `examples/apps/api-edge/component.yaml` contains dummy CI trigger label `trigger: pr-142-dummy-change`.
3. PR scope is massively unrelated: TUI cockpit spec pack, root `orun-tui-cockpit.md`, `agents/orchestrator.md`, historical task prompts, and stale `ai/waiting_for_input.md` are bundled with a small GitHub CLI UX fix.
4. PR title/body are placeholder (`chore: update happy-patch-113`, `Created by rh-ghflow`) and no implementer report exists.

PR #142 remains OPEN. Do not treat Task 0142 as completed/merged.

## Current Task (0143 — Implementer)

Prompt: `ai/tasks/task-0143.md`

Objective: repair PR #142 into a coherent GitHub CLI UX fix PR by preserving only the valid `--orun-dir` normalization, `orun github status` resolver flags, matching CLI docs, and direct UX-review context; remove dummy/unrelated files; retitle/rewrite the PR; commit an implementer report; and re-trigger CI.

### PR Boundary

In scope:

- `cmd/orun/command_github.go` changes for `--orun-dir` normalization and `github status` flags.
- `cmd/orun/command_github_test.go` only if focused tests are added or adjusted.
- `website/docs/cli/orun-github.md` docs matching the code behavior.
- `docs/github-log-pull-ux-review.md` only if retained as direct context for the UX fix.
- `ai/reports/task-0143-implementer.md` with real PR number and test/CI evidence.
- PR metadata cleanup for PR #142, or successor PR if #142 cannot be safely repaired.

Out of scope / must be removed from this PR:

- `.kiro/specs/orun-tui-cockpit/**`
- `orun-tui-cockpit.md`
- `agents/orchestrator.md`
- historical `ai/tasks/task-0139-verifier.md`, `task-0140.md`, `task-0140-verifier.md`, `task-0141-verifier.md`
- stale `ai/waiting_for_input.md`
- dummy `trigger: pr-142-dummy-change` in `examples/apps/api-edge/component.yaml`

### Acceptance Summary

- PR #142 diff is narrow and no longer includes dummy/unrelated files.
- Local Go tests/build pass:
  - `go test ./cmd/orun/ -run 'TestGithub(Status|Pull|Logs|Runs)|TestGithubCommand' -v`
  - `go test ./internal/artifactstore/github/... -v`
  - `go test ./internal/runbundle/... -v`
  - `go test ./cmd/orun/... -v`
  - `go build ./cmd/orun/`
- PR title/body tell the actual story.
- `ai/reports/task-0143-implementer.md` is committed with real PR number.
- Required GitHub checks are re-run and no required check is queued, failing, cancelled, or unknown when implementer reports complete.

## Current Roadmap Position

GitHub Artifacts stabilization remains active. Immediate priority is repairing the failed open PR before generating fresh feature work.

After Task 0143 and its verifier complete, remaining GitHub Artifacts gaps are:

1. Partial hydration display verification and CLI integration tests (Requirements 10 and 20).
2. Workflow template/root workflow decision and E2E workflow coverage (Requirements 17 and 21).
3. ArtifactStore memory/local test implementation gap (Requirement 5, low priority).
4. TUI cockpit Phase 1 from `.kiro/specs/orun-tui-cockpit/tasks.md`, only after the spec pack lands in a dedicated PR or the user approves using it as-is.

## Next Task After 0143

Generate a **Verifier** task for Task 0143 / PR #142 (or the successor PR) once the implementer reports that the branch has been narrowed, PR metadata fixed, report committed, and CI re-run. The verifier must inspect diff scope, code behavior, tests, PR CI logs, secret safety, and then merge only on PASS plus green CI.

If Task 0143 cannot repair PR #142 because CI remains stuck after re-run or branch ownership prevents rewriting, the implementer should report the blocker and the orchestrator should decide whether to close/supersede PR #142 with a clean successor PR.

## Deferred Dedicated PRs

- `chore(spec): add orun TUI cockpit spec pack` for `.kiro/specs/orun-tui-cockpit/**` and `orun-tui-cockpit.md`.
- `docs(agents): add orchestrator operating protocol` for `agents/orchestrator.md`.
- `chore(ai): archive historical task prompts and refresh waiting input state` for old `ai/tasks/*` prompts and `ai/waiting_for_input.md`.
