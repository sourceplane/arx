# Task 0019 — Verifier Report

**Result: PASS (with one Spec Proposal filed for M5.d follow-up)**

PR: https://github.com/sourceplane/orun/pull/163
Head SHA at verification: `fb364f1` (housekeeping commit on top of feature SHA `947773d`).
Branch: `impl/task-0019-m5c-orun-read-commands-rewire`
Base: `main` @ `1677cc1`. mergeStateStatus: CLEAN. mergeable: MERGEABLE.

## Checks

| Step | Result | Notes |
|---|---|---|
| 1. Scope discipline (PR boundary) | PASS | Files changed: `cmd/orun/{bridge_mirror_warn,command_describe,command_get,command_logs,command_read_revision_test,command_status,read_resolve}.go` + `ai/reports/task-0019-implementer.md`. Zero edits to `internal/runner/`, `internal/runbundle/`, `internal/executor/`, `internal/executionstate/` writer paths, `internal/revision/` writer paths, `internal/statestore/`, or `internal/state/`. Read-only consumers only. |
| 2. `go build ./...` | PASS | Clean. |
| 3. `go vet ./...` | PASS | Clean. |
| 4. `go test -race ./...` | PASS | All packages green. |
| 5. `make test-state-redesign` | PASS | Coverage gates: `internal/statestore` 95.7% (≥95), `internal/revision` 90.4% (≥90), `internal/executionstate` 90.0% (≥90). |
| 6. New M5.c tests (14 tests in `command_read_revision_test.go`) | PASS | All `TestGetPlans_*`, `TestDescribeRevision_*`, `TestDescribeTrigger_*`, `TestResolveExecutionForRead_*`, `TestWarnBridgeMirrorFailures_*`, `TestStatusFlagsRegistered`, `TestLogsFlagsRegistered`, `TestDescribeAliasesRegistered` green. |
| 7. End-to-end fresh-revision walk in `examples/` | PASS (with drift on `latest` literal — see Issues / Proposal) | `validate` → `run --runner local` materialized `rev-manual-no-git-p3b75457b` + `run-001`. `status`, `get plans`, `get plans -o json`, `describe revision <revKey>`, `describe revision` (empty), `describe execution latest`, `describe execution run-001`, `status --revision`, `status --exec-id`, `status --all`, `logs --revision`, `logs --exec-id` all rendered correctly with the revision-first table shape from `cli-surface.md` §3/§4/§6. |
| 8. Legacy-fallback walk (synthesized `.orun/executions/legacy-001/`) | PASS | All four read commands transparently fall back: `status` shows legacy plan/exec, `logs` reaches legacy log dir, `describe run latest` resolves via legacy `state.Store`, `get plans` renders the legacy plan-hash table when no `revisions/` dir exists. |
| 9. `bridge-mirror-failed` surfacing walk | PASS | Single-line stderr warning emitted by all four read commands when an `events/<seq>-bridge-mirror-failed.json` file exists. Multi-event case yields exactly one warning per execution (dedup verified — two events → one warning). Filename-based detection is body-shape-agnostic, so malformed JSON does not break the helper. Non-bridge events stay silent. Commands never block or change exit code. |
| 10. New flags + aliases | PASS | `--revision <key>`, `--exec-id <key>`, `--all` exposed on `status` and `logs` (verified via `--help` and `TestStatusFlagsRegistered`/`TestLogsFlagsRegistered`). `describe revision`, `describe trigger`, `describe execution` aliases registered (verified via `TestDescribeAliasesRegistered` + live walk). Triplet (`Revision Key`, `Execution Key`, `Legacy Exec ID`) printed in `describe execution` and `describe run` output. |
| 11. `orun get plans -o json` | PASS | Stable key order (`revisionKey`, `trigger`, `plan`, `jobs`, `latestExec`, `status`, `environments`); legacy fallback returns `[]` on empty workspace. (Note: the flag is `-o json` per `command_get.go`, not `--json`; the spec doesn't disagree.) |
| 12. `kiox -- orun validate` in `examples/` | PASS | Local. |
| 13. `kiox -- orun plan --changed` in `examples/` | LOCAL-ONLY BLOCK (carry-forward) | Persistent composition-cache quirk reproduced (`stack.yaml at /Users/.../cache/compositions/c41fc0830d... has no spec.compositions and no compositions.yaml files`). CI is authoritative — `CI / Orun Plan` ran the same walk on a clean machine and PASSed in 56s. |
| 14. PR CI log review on final head SHA | PASS | See "CI Log Review" below. |

## Issues

**One non-blocking spec drift detected.** Filed as
`ai/proposals/task-0019-spec-update.md`. PASS-with-followup decision per
the verifier prompt's "Spec drift" step.

- Severity: low.
- Drift: `cli-surface.md` §5.1 lists `orun describe revision latest`,
  `orun describe trigger latest`, and `orun describe trigger <triggerName>`
  as canonical syntax. The first two error with
  `revision: ambiguous or unknown run target: "latest"` because
  `revision.ResolveRevision` lacks a `latest`-literal branch (empty-arg
  works — that's what the new tests use, masking the gap). The third is a
  deeper resolver gap: `describe trigger system.manual` is not understood
  because the resolver is keyed on revision identifiers, not trigger names.
- Workarounds available today: `describe revision` (no arg), `describe
  revision <revKey>`, `describe trigger` (no arg), `describe trigger
  <revKey>`. All four read commands still work for the documented
  primary flows.
- Why not a verifier-side fix: the cleaner remediation crosses the
  `internal/revision` resolver boundary, which is out of scope for M5.c
  (read-only consumers). Recommendation: fold a small Option-A normalization
  into M5.d's scope. Full analysis + Option A/B remediation in the
  proposal.

No other issues. No production-safety concerns.

## CI Log Review

Reviewed both required checks via `gh run view <id> --log` (not just
`--json conclusion`). Both ran on housekeeping head SHA `fb364f1` —
identical workflow shape to the previous run on `947773d`, both PASS.

### `CI / Orun Plan` — run `26686932774`

- Duration: 56s. Conclusion: SUCCESS.
- Step `Plan with Orun` executed real binary:
  ```
  orun plan --from-ci github --event-file "$GITHUB_EVENT_PATH"
            --intent examples/intent.yaml --artifact github --github-output
  ```
- Output:
  ```
  ✓ Plan revision created
    Revision: rev-github-pull-request-fb364f1-pbfd83933
    Trigger:  github-pull-request / changed / fb364f1
    Jobs:     0
    Path:     /home/runner/work/orun/orun/examples/.orun/revisions/rev-github-pull-request-fb364f1-pbfd83933/plan.json
    │ 0 components × 3 envs → 0 jobs
    │ mode: changed-only
    │ plan: 96dd1b417511
  ```
- Plan artifact uploaded (`orun.v1.gh-26686932774-1-96dd1b417511.plan.sha256-96dd1.created`, 2382 bytes).
- 5 matrix legs SKIPPED (empty matrix at M5.c — same shape as M5.a/M5.b, expected).
- No "skip-by-condition" silent passes — the `Plan with Orun` step shows the actual command and its output.

### `Harness dry-run guard` — run `26686932783`

- Duration: 13s. Conclusion: SUCCESS.
- Step `Run dry-run guard` produced 30+ `[guard] PASS:` log-level
  assertions covering: bash syntax, foundation/api smoke job counts,
  command-marker presence, duplicate-claim helper (PASS+FAIL cases),
  status helper (PASS+FAIL cases), `ORUN_EXEC_ID` / `ORUN_REMOTE_STATE`
  exports, harness invariants (`assert_exactly_one_duplicate_claimant`,
  `assert_jobs_all_succeeded`), background PID tracking, signal-safe
  cleanup trap (INT, TERM), jq + orun-binary preflight checks,
  repo-linkage preflight. Final line: `[guard]   DRY-RUN GUARD PASSED`.
- 4 matrix legs SKIPPED (empty matrix; same shape as M5.a/M5.b).

## Coverage Evidence

From `make test-state-redesign` on the verify branch:

```
🧪 Coverage gate: ./internal/statestore/... (>= 95%)
   measured: 95.7%
🧪 Coverage gate: ./internal/revision/... (>= 90%)
   measured: 90.4%
🧪 Coverage gate: ./internal/executionstate/... (>= 90%)
   measured: 90.0%
```

All three gates held. PR introduces zero changes under those packages —
the gates are preserved by construction.

## Live Resource Evidence (temp-workspace walk)

- Fresh revision-first walk in `examples/`: `run --runner local`
  materialized `.orun/revisions/rev-manual-no-git-p3b75457b/` + `run-001`
  (3 jobs failed locally because `examples/` ships GHA-syntax steps and
  the local runner refuses `actions/setup-node@v4` — unrelated to this PR;
  the runner-engine error existed before M5.c).
- Read commands produced expected revision-first output (status, get
  plans, describe revision <revKey>, describe execution latest, status
  --revision/--exec-id/--all, logs --revision/--exec-id).
- Bridge-mirror walk: dropped `0001-bridge-mirror-failed.json` →
  single-line warning on all four reads. Two events → one warning
  (dedup confirmed). Malformed body still warns (filename-based detect).
  No event → silent. Exit code unchanged in every case.
- Legacy-only synthesized workspace (`.orun/executions/legacy-001/`):
  all four read commands fell back to legacy `state.Store` paths
  transparently.
- `kiox -- orun plan --changed` in `examples/` reproduced the local
  composition-cache quirk noted in `ai/state.json`. CI is authoritative
  and PASSed in 56s on a clean machine.

## Secret Handling Review

- No secrets / tokens / user emails in any changed log path. The new
  one-line bridge-mirror warning emits only `<execKey>` and `<revKey>`,
  both of which are content-addressed identifiers (revision-key suffix
  is a plan-hash prefix) and contain no PII.
- New describe output prints `RevisionID` (ULID), `PlanHash`,
  `TriggerKey`, `TriggerName`, `TriggerType`, `Mode`, `Provider`, `Event`,
  `Action`, `PlanScope.Mode` — all schema-defined fields with no secret
  surface.
- CI log review confirmed `GITHUB_TOKEN`, `ACTIONS_RUNTIME_TOKEN` masked
  as `***`. No bare credentials.
- No `time.Now()` direct calls introduced (`grep -n "time.Now" cmd/orun/{read_resolve,bridge_mirror_warn,command_status,command_logs,command_describe,command_get}.go` returns nothing in the new code paths; existing call sites untouched). Clock seam preserved.
- Errors flow through `errors.Is`/`errors.As` — `resolveExecutionForRead`
  returns wrapped `executionstate.ResolveExecution` errors; `describeRevision` /
  `loadRevisionPlanRows` wrap `revision.ResolveRevision` errors via `fmt.Errorf("…: %w", err)`.

## Risk Notes

Carry-forward from M5.b plus this PR's residuals:

1. **Carry-forward (M5.b):** Persistent local composition-cache quirk
   (`stack.yaml at .orun/cache/compositions/<hash> has no spec.compositions`)
   continues to repro on this workstation. CI is authoritative.
   Unchanged by M5.c.
2. **Carry-forward (M5.b):** `examples/` ships GHA-syntax steps that the
   local runner can't execute; reproducible end-to-end run requires
   `--runner github-actions` (dry-run) or `--gha`. Not introduced here.
3. **New (M5.c):** `describe revision latest` / `describe trigger latest`
   error on the literal `latest` arg (use empty-arg or explicit revision
   key). `describe trigger <triggerName>` not yet supported. See Spec
   Proposal. Non-blocking.
4. **New (M5.c):** End-to-end Cobra-invocation tests for the new flags
   are not yet added — coverage is at the helper level
   (`describeRevision`, `describeTrigger`, `loadRevisionPlanRows`,
   `renderRevisionPlanTable`, `warnBridgeMirrorFailures`). The verifier
   walk filled this gap manually. Acceptable for a read-only consumer
   rewire.
5. **New (M5.c):** `bridge_mirror_warn.warnBridgeMirrorFailures` is
   filename-based (matches `*-bridge-mirror-failed.json`) and does not
   parse the payload. `data-model.md` §9.1 documents a richer schema
   (reason, sourcePath, targetPath); future iterations could surface
   those fields when present. Non-blocking; the warning is a
   notification, not a diagnostic.

## Spec Proposals

- `ai/proposals/task-0019-spec-update.md` — describe alias drift on
  `latest` literal + `<triggerName>` arg. PASS-with-followup; fold into
  M5.d scope.

## Recommended Next Move

PASS the PR. Merge `#163` per the Verifier Merge Protocol. Update state
files. Emit the M5.d implementer task (`orun state migrate` hidden
command) as the next orchestrator cycle, and roll the
`describe revision/trigger latest` + trigger-name lookup remediation
into that task's scope per the proposal.

## PR Number

**PR #163** — https://github.com/sourceplane/orun/pull/163

Branch: `impl/task-0019-m5c-orun-read-commands-rewire`
Head SHA at verification: `fb364f1`
Base: `main` @ `1677cc1`
Required CI checks: both PASS on final head SHA per `gh run view --log`.
