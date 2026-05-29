# Current Orchestration Context

Last updated: 2026-05-29 (Task 0145.1 verified PASS, PR #144 merged)

## Repo Reality

- Local `main` synced with `origin/main` at `300a436` (`fix(github): normalize --orun-dir and add status selectors (#144)`). Working tree clean after orchestration cleanup commit.
- Completed checkpoint: Task 0145.1 verified PASS and merged PR #144. The GitHub CLI UX fix (`--orun-dir` normalization, `github status` selector flags, matching docs/tests, UX-review note) is now durable on `main`.
- Open PRs: none. PR #142 is CLOSED as superseded (closedAt 2026-05-29T00:01:57Z).
- Repo health: **green**. TUI Cockpit Phase 2 is now unblocked.

## Last Completed Task (0145.1)

Task 0145.1 verified PASS and merged PR #144 at merge commit `300a436` (2026-05-29T00:14:11Z). Verifier report: `ai/reports/task-0145-verifier.md`. Implementer report: `ai/reports/task-0145-implementer.md`.

Durable outcomes now on `main`:

- `cmd/orun/command_github.go`: `normalizeOrunDir()` helper centralizes `--orun-dir` resolution. Empty input → `./.orun`; parent input → `<parent>/.orun`; already-`.orun` input unchanged. `runGithubPull` calls the helper once instead of bifurcated logic.
- `orun github status` registers the same six selector flags as `pull`/`logs`: `--run-id`, `--exec-id`, `--sha`, `--branch`, `--latest`, `--failed`. Status, logs, and pull now share a uniform vocabulary.
- `cmd/orun/command_github_test.go`: five new focused tests cover the three normalization branches plus selector flag registration and parse-time acceptance.
- `website/docs/cli/orun-github.md`: public docs match the final CLI behavior, including the full-SHA caveat and `--job` substring-match caveat.
- `docs/github-log-pull-ux-review.md`: short UX-review note with reproducer commands, code-level root cause for the two fixed bugs, and a prioritized list of three open friction items (short-SHA support, `--job` logical-id matching, `--latest` branch disclosure).
- PR #142 (`happy-patch-113`) is CLOSED as superseded. No remaining open-dirty PRs.

## Prior Checkpoint (Task 0144.1)

Task 0144.1 verified PASS and merged PR #143 at merge commit `17d3b58`; orchestration cleanup landed at `9ac35d3`. Verifier report: `ai/reports/task-0144-verifier.md`.

Durable outcomes now on `main`:

- `orun tui` Cobra subcommand registered (`cmd/orun/command_tui.go`, `cmd/orun/commands_root.go`).
- `--remote-state` fails closed (`✕ --remote-state requires --backend-url or ORUN_BACKEND_URL`) before `tea.NewProgram(...).Run()`, with focused command tests.
- `internal/tui` Phase 1 foundation: root Bubble Tea model, Mode/Panel enums, async workspace load, loading/error states, three-panel/status/key-hint rendering, quit/reload/focus/help bindings.
- `internal/tui/services` boundary calls Orun internals directly. No `exec.Command` and no `"orun"` literal under `internal/tui/`.
- Phase 2/3 surfaces (`GeneratePlan`, `RunPlan`, `Describe`, `TailLogs(Follow=true)`, `ListRuns(RemoteState=true)`) remain explicit stubs/errors.
- Charm deps pinned: `bubbletea v1.3.5`, `bubbles v0.21.0`, `lipgloss v1.1.0`. Actual property-test import path is `pgregory.net/rapid v1.1.0`, not the stale GitHub mirror named in the spec.

## Current Task

None scoped. Repo is green; orchestrator is ready to scope the next cycle.

## Current Roadmap Position

1. ✅ GitHub Artifacts Level 1/2 CLI gaps through Task 0141.1.
2. ✅ Orun Cockpit TUI Phase 1 foundation through Task 0144.1.
3. ✅ Task 0145 / 0145.1: dirty PR #142 resolved, GitHub CLI UX cleanup merged via PR #144.
4. ⏭️ Task 0146: TUI Cockpit Phase 2 / Plan Studio wiring through `internal/planner.GeneratePlan` — repo health is green so this is now safe to scope.
5. ⏭️ Optional: pick up `docs/github-log-pull-ux-review.md` section 3 follow-ups (short-SHA support, `--job` logical-id matching, `--latest` branch disclosure) as a small focused CLI UX task either before or in parallel with 0146.

## Next Task

Task 0146 (Implementer): scope TUI Cockpit Phase 2 — implement `GeneratePlan` through `internal/planner`, Plan Studio form/review state, focused property/unit tests, and the one-line `pgregory.net/rapid` import-path spec housekeeping if still needed.

## Open Risks

- None blocking. Open follow-ups in `docs/github-log-pull-ux-review.md` section 3 are non-blocking UX polish.
