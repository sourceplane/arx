# Task 0014 — Implementer Report

**Status:** PR #160 OPEN, awaiting CI + Verifier (Task 0015).
**Branch:** `impl/task-0014-m4-executionstate-prb` (from `main` @ `ed48633`).
**PR Number:** [#160](https://github.com/sourceplane/orun/pull/160).
**PR title:** `Task 0014: M4 PR-B — internal/executionstate bridge + EXDEV fallback`.

---

## Summary

- Shipped `internal/executionstate/bridge.go` — frozen `Bridge{Store, LegacyRoot,
  MirrorMode, Now}` surface plus `MirrorRunnerOutput(ctx, execKey, revKey,
  legacyExecID)` with hardlink-with-copy-fallback, portable EXDEV detection
  (`errors.Is(err, syscall.EXDEV)` + `*os.LinkError` type-assertion), and
  structured `bridge-mirror-failed` event emission. Failures emit + return nil
  per implementation-plan.md §M4; precondition violations return `%w`-wrapped
  `statestore.ErrInvalid`.
- Test seam is a package-level `linkFn` function variable (default `os.Link`).
  Tests swap it for an `*os.LinkError{Err: syscall.EXDEV}`-returning stub so
  the cross-device fallback path is reachable on macOS/Linux CI without any
  privileged FS mounts. The seam is restored via `t.Cleanup`.
- `bridge-mirror-failed` payload shape is fixed in code where data-model.md §9
  left it open: `{executionKey, revisionKey, legacyExecId, artifact, stage,
  mode, error}` plus envelope `{kind, at, payload}`. `at` is UTC. Event
  sequence allocation reuses the existing trigger-first events layout via
  `Store.List` and `CreateIfAbsent`, bumping past any garbage filenames.
- Coverage delta on `internal/executionstate`: PR-A landed at 90.0% (exact
  floor); PR-B holds 90.0% with bridge.go fully exercised (MirrorMode.String
  100%, MirrorRunnerOutput 100%, mirrorOne 89.5%, isCrossDevice 100%, now
  100%). Statestore lifted to 95.7% (was 95.4%) via the two additive helpers'
  unit tests; revision held at 90.4%.
- M4 closure: bridge ships, resolver-with-legacy-fallback (PR-A) is unchanged,
  CLI/runner wiring is correctly deferred to M5 per the M4 hard non-goals.

## Files Changed

- `internal/executionstate/bridge.go` (new, 425 lines)
- `internal/executionstate/bridge_test.go` (new, 770 lines, 19 test functions)
- `internal/statestore/paths.go` (additive: `LegacyExecutionFilePath`,
  `ExecutionFilePath`)
- `internal/statestore/paths_test.go` (tests for the two additive helpers)

No edits in `cmd/orun`, `internal/state`, `internal/runner`,
`internal/runbundle`, `internal/triggerctx`, or PR-A production code. No
`Makefile` change — the existing `≥ 90%` gate is unchanged.

## Checks Run

- `go build ./...` → clean.
- `go vet ./...` → clean.
- `go test -race -count=1 ./...` → all packages pass under `-race`. Full suite
  ran in ~3.5 minutes (executionstate 50.9s under -race).
- `make test-state-redesign` →
  - `internal/testfx/statefs` ok
  - `internal/triggerctx` ok
  - `internal/statestore` 95.7% (gate ≥ 95%)
  - `internal/revision` 90.4% (gate ≥ 90%)
  - `internal/executionstate` 90.0% (gate ≥ 90%)
- `go list -deps ./internal/executionstate` → only sourceplane internal
  packages: `internal/{statestore, model, trigger, triggerctx, revision}`. No
  reach into `cmd/orun`, `internal/state`, `internal/runner`,
  `internal/runbundle`. Leaf-clean.
- `kiox -- orun validate --intent intent.yaml` (from `examples/`) → "✓ All
  validation passed."
- `kiox -- orun plan --changed --intent intent.yaml --output /tmp/plan.json`
  (from `examples/`) → fails with the documented composition-cache env quirk
  ("`stack.yaml at /Users/.../cache/compositions/c41fc0…` has no
  spec.compositions and no compositions.yaml"). This is the local-only
  flakiness called out in task-0014.md §Acceptance Criteria; CI is
  authoritative. Skipped `orun run --plan plan.json --dry-run` because plan
  generation failed; CI's `Harness dry-run guard` exercises the same path.
- CI runs (after push to PR #160): not yet polled at log level by this report
  — verifier (Task 0015) re-runs the leaf-clean audit and inspects both
  required CI checks (`CI / Orun Plan`, `Harness dry-run guard`) per the M4
  verifier protocol.

## Assumptions

- **`MirrorMode` enumeration:** task-0014 §PR Boundary authorized `Auto` "if
  the spec wording supports it." Implementation-plan.md §M4 calls the
  semantics "hardlink first; copy on EXDEV," which is a single mode; PR-B
  exposes three so callers (M5+ remote drivers) have an explicit copy
  selector. `MirrorModeAuto` is the zero value so a freshly constructed
  `Bridge{}` behaves correctly. `MirrorModeHardlink` exists for callers (and
  tests) that want to learn about cross-device drift; `MirrorModeCopy` skips
  os.Link entirely. If the verifier disagrees with the trinary surface, the
  enum can be narrowed without an on-disk format change — the event payload
  encodes the mode by string name.
- **`bridge-mirror-failed` payload field set:** data-model.md §9 lists the
  event but does not pin the payload schema. PR-B fixes the field set to
  `{executionKey, revisionKey, legacyExecId, artifact, stage, mode, error}`.
  Stage values: `read-source`, `read-dest`, `translate-dest`, `mkdir-dest`,
  `remove-dest`, `link`, `copy`. The shape is additive-friendly: appending
  fields is non-breaking, renaming is.
- **Idempotent short-circuit on identical bytes:** the bridge reads the
  destination first; if bytes match the source it skips both link and copy.
  This converts naive runner re-emits into free no-ops (test
  `TestMirrorRunnerOutput_Idempotent` asserts 3 consecutive calls produce zero
  events).
- **Sequence allocation:** `nextEventSeq` starts at 2 when the events
  directory is empty (leaving seq 1 for `execution-created` per writer.go).
  Garbage filenames (no 20-digit prefix, non-numeric prefix) are skipped.
  Concurrent writers retry up to `mirrorEventRetryBudget = 32` slots.
- **Default `Now`:** when `Bridge.Now` is nil, `time.Now().UTC()` is used,
  matching the writer.go / revision.Config shim shape exactly. No new
  abstraction.

## Spec Proposals

None filed. The two ambiguities (MirrorMode enumeration and
bridge-mirror-failed payload) are documented inline; the verifier (Task 0015)
can adjudicate via Risk Note or post-hoc proposal. No spec edits in this PR.

## Remaining Gaps (M5 inherits)

- **Production wiring of `MirrorRunnerOutput`:** the runner (`orun run`) must
  call this once per terminal-state transition. PR-B exercises the bridge
  through tests only; M5.a/M5.b owns the CLI rewire that actually invokes it.
- **Migration command (`orun state migrate`):** M5.d.
- **Streaming copy path:** the in-memory `os.ReadFile` + `Store.Write` is fine
  for Phase 1 manifests (state.json / metadata.json are tiny). When M5 wires
  larger artifacts, swap to a streaming copy without churning the bridge
  surface.
- **Listing legacy executions for "drain everything":** the M5.d migrate
  command will need a helper to enumerate `<LegacyRoot>/*` execution IDs.
  Removed from PR-B scope to keep coverage focused.

## Next Task Dependencies (Task 0016: M5.a `orun plan` rewire implementer)

- The bridge's contract is **failures return nil**. M5 callers must NOT treat
  a `MirrorRunnerOutput` return as load-bearing — the resolver (PR-A) already
  prefers revision-first with legacy fallback, so a missed mirror only
  delays the new layout's eventual convergence.
- The bridge accepts the GHA-shape `gh-{run_id}-{attempt}-{sha}` legacyExecID
  without further sanitization (R-002 closed in Task 0013). M5 wiring should
  pass the runner's raw exec id through unchanged.
- `MirrorMode = MirrorModeAuto` (zero value) is the right default for the
  local driver. M5 remote drivers (S3/GCS) should set `MirrorModeCopy`
  explicitly so they never reach the local-FS-only `os.Link` seam.

## PR Number

**#160** — https://github.com/sourceplane/orun/pull/160 — OPEN as of report
write time. Verifier (Task 0015) is responsible for log-level CI inspection
and the merge.
