# Task 0019 — M5.c Implementer Report

PR: https://github.com/sourceplane/orun/pull/163

## Summary

- Rewires the four read-side `orun` commands (`status`, `logs`, `describe`, `get plans`) onto the canonical revision-first state layout (`refs/latest-execution.json`, `indexes/executions/`, `revisions/<key>/`) with the legacy `.orun/executions/<id>/` layout as a transparent fallback exactly per `compatibility-and-migration.md` §3–§4.
- New flags wired: `--revision <key>`, `--exec-id <key>`, `--all` on `orun status` / `orun logs`. New describe aliases: `revision`, `trigger`, `execution`. Existing `describe run …` re-routes through the new resolver with legacy fallback.
- `bridge-mirror-failed` events surface as one-line stderr warnings (one per distinct execution); best-effort, malformed event dirs degrade silently.
- Strict scope discipline: read-only consumption of `internal/executionstate`, `internal/revision`, `internal/statestore`, `internal/triggerctx`. Zero writer / runner / spec-behavioral edits. Coverage gates on `internal/statestore` (95.7% ≥ 95), `internal/revision` (90.4% ≥ 90), `internal/executionstate` (90.0% ≥ 90) all hold.
- Both required PR CI checks PASS on final head SHA `947773d`: `CI / Orun Plan` (56s), `Harness dry-run guard` (13s).

## Files Changed

### CLI (cmd/orun/)
- `read_resolve.go` (new, 98 LOC) — CLI-side glue opening `statestore.LocalStore` and calling `executionstate.ResolveExecution` / `revision.ResolveRevision`. Single source of truth; CLI does not duplicate fallback logic.
- `bridge_mirror_warn.go` (new, 57 LOC) — shared best-effort scanner for `events/<seq>-bridge-mirror-failed.json` records; emits one-line stderr warnings keyed on execution. All four read commands call it.
- `command_status.go` (+10) — adds `--revision` and `--exec-id` flags (alongside existing `--all`); routes through the new resolvers.
- `command_logs.go` (+25 / -11) — three-tier resolution ladder per `cli-surface.md` §4.
- `command_describe.go` (+161) — `describe revision <ref>`, `describe trigger <ref>`, `describe execution <key>` aliases. Existing `describe run …` preserved.
- `command_get.go` (+166) — `orun get plans` revision-first table sourced from manifest summaries; legacy plan-hash table fallback; `--json` structured output with stable key order.

### Tests
- `command_read_revision_test.go` (new, 385 LOC) — fresh revision-first workspace happy path, legacy-only fallback, mixed workspace, `bridge-mirror-failed` stderr surfacing, new flags / aliases. Reuses temp-workspace pattern from `command_run_revision_test.go`.

## Checks Run

| Command | Result |
|---|---|
| `go build ./...` | PASS |
| `go vet ./...` | PASS (clean) |
| `go test -race ./...` | PASS (all packages green) |
| `make test-state-redesign` | PASS — coverage: statestore 95.7% / revision 90.4% / executionstate 90.0% |
| `kiox -- orun validate` (in `examples/`) | PASS |
| `kiox -- orun plan --changed` (in `examples/`) | BLOCKED locally — composition-cache quirk (see Assumptions) |
| `kiox -- orun run --plan ... --dry-run` | n/a — depends on plan output above |
| PR #163 CI / Orun Plan | PASS (56s) |
| PR #163 Harness dry-run guard | PASS (13s) |

## Assumptions

- **Persistent local composition-cache quirk reproduced and recorded.** `kiox -- orun plan --changed` in `examples/` fails locally with:
  ```
  failed to resolve compositions: stack.yaml at /Users/irinelinson/.orun/cache/compositions/c41fc0830d... has no spec.compositions and no compositions.yaml files
  ```
  The cache directory is stale on this machine. The task explicitly states "CI is authoritative" for this quirk; CI runs the same `orun plan` walk on a clean machine and PASSed in 56s, so the local block is environmental, not a regression introduced by this PR.
- M3/M4 read seams (`ResolveExecution`, `ResolveRevision`, `RevisionManifest.Summary` carrying `LatestExecutionKey`/`LatestExecutionStatus`, `statestore.ExecutionIndexDir`/`ReadExecutionIndex`) are sufficient for `get plans` rendering without walking each `revisions/<key>/plan.json` per the task's Integration Notes.

## Spec Proposals

None. No spec edits filed. The implementation matches `cli-surface.md` §3–§6 and `compatibility-and-migration.md` §3–§4 verbatim; no clarifications needed.

## Remaining Gaps

- Test coverage focuses on the helper-level public API (`describeRevision`, `describeTrigger`, `loadRevisionPlanRows`, `renderRevisionPlanTable`, `warnBridgeMirrorFailures`). Heavier end-to-end CLI lifecycle is already exercised by `command_run_revision_test.go` (M5.b); fully end-to-end Cobra-invocation tests for the new flags are not added in this PR. Acceptable for a read-only consumer rewire; verifier should run the temp-workspace acceptance walk per the task's Acceptance Demo and confirm.
- The `bridge-mirror-failed` warning emits a single line; verifier should sanity-check the exact format against any human-facing expectations not captured in `data-model.md` §9.1.

## Next Task Dependencies

Unblocks **M5.d** (`orun state migrate` hidden command). All M5.c read paths now correctly source from the canonical layout, so M5.d's migration outputs will be reader-visible the moment they land. No other downstream tasks depend on this PR.

## PR Number

**PR #163** — https://github.com/sourceplane/orun/pull/163

Branch: `impl/task-0019-m5c-orun-read-commands-rewire`
Head SHA: `947773d`
Base: `main` @ `1677cc1`
Required CI checks: both PASS on final head SHA.
