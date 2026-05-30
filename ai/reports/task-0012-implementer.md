# Task 0012 Implementer Report — M4 PR-A: internal/executionstate

**PR:** https://github.com/sourceplane/orun/pull/159 (#159)
**Branch:** `impl/task-0012-m4-executionstate-pra`

## Summary

Introduced the `internal/executionstate` package implementing the M4 execution-row model, writer, and resolver per the state-redesign design docs. The package is a leaf — zero imports of `cmd/orun`, `internal/state`, `internal/runner`, or `internal/runbundle`. Bridge/runner-mirror is intentionally deferred to PR-B per implementation-plan.md M4 split.

## Files Changed

- `internal/executionstate/model.go` — `ExecutionRun` typed model (data-model.md §5).
- `internal/executionstate/writer.go` — `NextExecutionKey`, `SanitizeExecID`, `CreateExecution`, `UpdateSnapshot`, `MarkTerminal`, `finalizeExecution`, `updateRevisionSummary`.
- `internal/executionstate/resolver.go` — `ResolveExecution` 7-branch ladder + legacy `.orun/executions` fallback (compat §3-§4).
- `internal/executionstate/internal.go` — canonical-JSON helpers shared with sibling state packages.
- `internal/executionstate/{model,writer,resolver,property,coverage_extra}_test.go` — unit + rapid property tests.
- `internal/statestore/paths.go` — additive helpers (`ExecutionsDir`, `ExecutionDocPath`, `ExecutionIndex*`, `LegacyExecution*`, `EventPath`, `SnapshotPath`).
- `Makefile` — `test-state-redesign` extended with `./internal/executionstate/...` ≥90% coverage gate.

## Checks Run

| Command | Result |
|---|---|
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `go test -race -count=1 ./...` | exit 0 (full repo green) |
| `make test-state-redesign` | exit 0 |
| `go test -count=1 -coverprofile=/tmp/cov.out ./internal/executionstate/...` | exit 0 |
| `go list -deps ./internal/executionstate \| grep -E "cmd/orun\|internal/state$\|internal/runner\|internal/runbundle"` | empty (leaf-clean ✓) |

### Coverage numbers

- `internal/statestore`: **95.4%** (gate ≥95%)
- `internal/revision`: **90.4%** (gate ≥90%)
- `internal/executionstate`: **90.0%** (gate ≥90%)  ← this PR

### kiox CLI gates

`kiox` is unavailable in this sandbox; `validate` / `plan` / `run --dry-run` were skipped (`kiox-skip`) per task spec allowance.

## Assumptions

- M3 callers writing a revision but not a manifest is treated as an error (loud surfacing) rather than silently no-op'ing the latest-execution-summary update — matches the conservative-on-unknowns posture of `revision/writer.go`.
- Legacy executions with no `revisionKey` carry `triggerKey = "system.migrated"` and `Reason = "migration"` per compat §4; status defaults to `"completed"` when absent.

## Spec Proposals

None.

## Remaining Gaps

- **PR-B** (`internal/executionstate/bridge.go`): hardlink-with-copy-fallback runner mirror under `.orun/executions/<runKey>/` so legacy runner code keeps working through the migration window. Tracked in implementation-plan.md M4 split.

## Next-Task Dependencies

- Task 0014 (and any consumer of execution-row data) inherits the public surface defined here: `ExecutionRun`, `Config`, `CreateExecutionInput`, `CreateExecution`, `UpdateSnapshot`, `MarkTerminal`, `ResolveExecution`, `ResolveOptions`, `ExecutionRef`, `ResolveSource*` constants, `Status*` constants, `IsTerminal`, `SanitizeExecID`, `NextExecutionKey`, `LegacyRoot`. No symbols are slated for breaking changes in PR-B.

## PR Number

**#159** — https://github.com/sourceplane/orun/pull/159
