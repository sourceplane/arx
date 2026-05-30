# Current Roadmap Position

## Active Spec
`specs/orun-state-redesign/` (Phase 1, local-only) — trigger-first revision-first
local state model. **Phase 1 STRUCTURALLY COMPLETE** with M6 merged.
See `specs/orun-state-redesign/README.md` for the index and read order.

## Active Milestone
**M6 — E2E + property gates — CLOSED.** PR **#165** squash-merged to `main`
as `ad3656e` on 2026-05-30T19:17:23Z. Branch
`impl/task-0021-m6-e2e-and-property-gates` deleted.

All six milestones now closed on main:

| M  | PR    | Main commit |
|----|-------|-------------|
| M0 | #152  | `4ea1980`   |
| M1 | #153  | `db342dd`   |
| M2 | #156  | `cd8b3e8`   |
| M3 | #158  | `bfc2ae6`   |
| M4 | #159 / #160 | `ed48633` / `d51e828` |
| M5 | #161–#164 | `7a9c494` … `17ef788` |
| M6 | #165  | `ad3656e`   |

**Next focus: re-scope.** Phase 2 (remote backend) items, the deferred
M6 carry-forward (MirrorModeHardlink, async mirror, --persist-revision,
Option B resolver), and the paused .kiro/specs roadmaps are all
candidates — to be selected in a follow-on session.

## Last Completed Implementer/Verifier (0021 — M6)
- Single-pass closure (implementer + verifier in one cycle).
- PR **#165** on `impl/task-0021-m6-e2e-and-property-gates`. Squash-merged
  to `main` as `ad3656e` "Task 0021 / M6: end-to-end + property gates
  for state redesign (#165)" on 2026-05-30T19:17:23Z.
- Required CI both PASS on PR head: `Orun Plan` SUCCESS (53 s);
  `Harness dry-run guard` SUCCESS (15 s). Matrix legs SKIPPED
  legitimately (empty matrix — test-only change).
- Diff stat (test-only / coverage-gate work):
  - new: `cmd/orun/state_e2e_test.go` (~340 LOC, TestStateE2E — 14
    sub-tests walking all of test-plan.md §4 from plan synthesis
    through state-migrate idempotence).
  - new: `internal/revision/keys_property_test.go` (~190 LOC,
    rapid-driven RevisionKey determinism + distinctness +
    ResolveCollision suffix contiguity).
  - new: `internal/revision/m6_coverage_test.go` (~160 LOC, lifts
    `internal/revision` from 84.9 % → 90.3 % via ScanLegacyPlanHashes
    happy/filter/nil-store paths, WriteLegacyNamedPlan error paths,
    RevisionKey input validation).
  - modified: `Makefile` (`test-state-redesign` propagates `-race`
    to every step, adds TestStateE2E invocation).
  - artifacts: `ai/tasks/task-0021.md`, `ai/tasks/task-0021-report.md`,
    `ai/reports/task-0021-verifier.md`.
- Post-merge gates on main: `go test ./... -race -count=1 -timeout 600s`
  all-green; `make test-state-redesign` all four gates green
  (statestore 95.7 %, revision 90.3 %, executionstate 90.0 %).
- The `internal/revision` floor had been silently breached at 84.9 %
  on main tip pre-M6; M6 restores the documented ≥ 90 % floor
  without lowering any threshold.
- Verifier report: `ai/reports/task-0021-verifier.md`. Single-pass
  closure recorded per `orun-saas-verifier` skill convention.
- Phase 1 reservations honoured (NOT wired): MirrorModeHardlink
  debug-fold decision deferred, RunnerHooks.AfterStateUpdate
  re-architecture deferred, `--persist-revision` flag wiring deferred
  to Phase 2, Option B trigger-name resolver branch deferred,
  `--prune-legacy` deferred to Phase 2 §6.

## Past Completed (0020 — M5.d)
- Implementer + verifier both Task 0020 (single-pass closure) → PR **#164** on
  `impl/task-0020-m5d-orun-state-migrate`.
- Squash-merged to `main` as `17ef788` "Task 0020: M5.d — hidden orun state
  migrate command (#164)" on 2026-05-30T16:03:59Z. Final head SHA on PR:
  `c3cddb7` (verifier-report commit on top of feature commit `06b64d9`).
- Required CI both PASS at log level on final head SHA: `Orun Plan` SUCCESS
  (52 s); `Harness dry-run guard` SUCCESS (13 s). 6 matrix legs SKIPPED
  (empty matrix, expected for state-redesign tasks).
- Diff stat (14 files changed, +1709 / -81):
  - new: `cmd/orun/command_state_migrate.go` (~430 LOC, hidden `orun state
    migrate` command + `--dry-run` flag, two-phase walk: plans then executions).
  - new tests: `cmd/orun/command_state_migrate_test.go` (~260 LOC, 5 tests
    covering happy path, dry-run, idempotent rerun, orphan execution,
    Option A literal-`latest` normalization).
  - new: `internal/revision/legacy_scan.go` (~95 LOC, `ScanLegacyPlanHashes`
    helper — sorted, hex-validated, latest.json-filtered legacy plan
    inventory).
  - modified: `cmd/orun/commands_root.go` (`registerStateCommand(rootCmd)`
    wired in init()), `cmd/orun/command_describe.go` (Option A:
    describeRevision + describeTrigger normalize literal `"latest"` to `""`
    at the entrance — pure CLI-layer fix per
    `ai/proposals/task-0019-spec-update.md`).
  - artifacts: `ai/tasks/task-0020.md` (task prompt, filed proactively
    because Task 0019's verifier merged #163 without emitting one),
    `ai/reports/task-0020-implementer.md`, `ai/reports/task-0020-verifier.md`,
    plus state-file updates and the late-add `ai/tasks/task-0019*.md`
    artifacts that hadn't been committed in the previous cycle.
- Tests: `go test ./... -count=1 -timeout 180s` all-green locally
  (cmd/orun 9.94 s, revision 5.33 s, executionstate 29.43 s, statestore
  20.16 s, every other package green). `go vet ./...` clean.
- Coverage: not regressed. M5.d touched `internal/revision` only via
  the new `ScanLegacyPlanHashes` helper (covered by the resolver +
  migrate test paths) and `internal/executionstate` only via API
  consumption (CreateExecution + Bridge.MirrorRunnerOutput).
- Verifier report: `ai/reports/task-0020-verifier.md`. Single-pass
  closure recorded per `orun-saas-verifier` skill convention.
- Phase 1 reservations honoured (NOT wired): `--persist-revision` flag
  remains reserved (migrate persists unconditionally because
  CreateExecution requires the manifest on disk; no need for the flag
  here), `--prune-legacy` deferred to spec §6 Phase 2 (explicitly out
  of M5).

## Past Completed (0019 — M5.c)
PR #163 → `73108ee` on 2026-05-30. `orun status / logs / describe / get`
revision-first rewire + new `--revision` / `--exec-id` / `--all` flags.
Verified PASS. `ai/proposals/task-0019-spec-update.md` Option A folded
into Task 0020. Option B (trigger-name resolver branch) intentionally
deferred — retain proposal file as queued source of truth.

## Past Completed (0018 — M5.b)
PR #162 → `59d06f3`. `orun run` rewire + `RunnerHooks.AfterStateUpdate`
+ §9.1 `bridge-mirror-failed` event payload pinned. Verified PASS
(single-pass closure).

## Past Completed (0016 — M5.a)
PR #161 → `7a9c494`. `orun plan` rewire onto canonical revision-first layout.
Verified PASS (single-pass closure).

## Past Completed (0014 / 0015 — M4 PR-B)
PR #160 → `d51e828`. Bridge + EXDEV fallback. Verified PASS by Task 0015.

## Past Completed (0012 / 0013 — M4 PR-A)
PR #159 → `ed48633`. Verified PASS by Task 0013.

## Current Task
**Task 0021 — M6 implementer (E2E + property gates).** Scoped 2026-05-30.
Branch base: `main` at `32d026f` (post-#164 housekeeping commit).
Branch: `impl/task-0021-m6-e2e-and-property-gates`.
Prompt: `ai/tasks/task-0021.md`.
PR boundary: (1) `cmd/orun/state_e2e_test.go` — 15-step revision-first walk
per `test-plan.md` §4 with one `t.Run` per step; (2)
`internal/revision/keys_property_test.go` — rapid-driven revision-key
uniqueness + collision-suffix correctness per §3.2; (3) statestore decode-
strict concurrent property (or cite the existing equivalent in
`local_prb_test.go`); (4) `Makefile` `test-state-redesign` extension: add
`-race` and `TestStateE2E` invocation, keep coverage gates unchanged.

## Repo Checkpoint

| Attribute | Value |
|---|---|
| Branch (local checkout) | `main` (clean post-#164 merge) |
| `main` tip | `17ef788` — Task 0020: M5.d — hidden orun state migrate command (#164) |
| Open PRs (state-redesign lineage) | none — #164 merged on 2026-05-30 |
| Repo health | 🟢 Green — Task 0020 PASS+merged; M5 fully closed; awaiting Task 0021 (M6) emission |
| Last verified | 2026-05-30 (Task 0020, PR #164, merge `17ef788`) |
| Active milestone | M6 (E2E + property gates) — pending |
| Tasks completed | 0001, 0002, 0003, 0004, 0005, 0007, 0008, 0009, 0010, 0011, 0012, 0013, 0014, 0015, 0016, 0018, 0019, 0020 (18 total) |
| Current task | **0021 (M6 implementer)** — scoped, ready to begin |

## Roadmap (M0 → M6)
1. ✅ **M0 Foundation** — landed on main at `4ea1980` (PR #152).
2. ✅ **M1 `internal/triggerctx`** — landed on main at `db342dd` (PR #153).
3. ✅ **M2 `internal/statestore`** — closed at PR #156 (`cd8b3e8`, 2026-05-30).
4. ✅ **M3 `internal/revision`** — closed at PR #158 (`bfc2ae6`, 2026-05-30).
5. ✅ **M4 `internal/executionstate` + runner bridge** — closed.
   - ✅ PR-A — model + writer + resolver (PR #159 → `ed48633`).
   - ✅ PR-B — bridge + EXDEV fallback (PR #160 → `d51e828`).
6. ✅ **M5 CLI rewire** — closed. Sub-tasks: ✅ M5.a (Task 0016, PR #161),
   ✅ M5.b (Task 0018, PR #162), ✅ M5.c (Task 0019, PR #163),
   ✅ M5.d (Task 0020, PR #164 → `17ef788`).
7. **M6 End-to-end + property gates** ← next. Scope per
   `specs/orun-state-redesign/implementation-plan.md`: end-to-end Go
   tests exercising the full revision-first path through real fixtures,
   plus property-based tests for the resolver branch table and the
   bridge atomicity contract.

## Next Task After 0020 (proposed)
**Task 0021 = M6 implementer** — E2E + property gates. Branch base:
`main` at `17ef788`. Optional sub-scope: roll Option B (trigger-name
resolver branch reading `refs/triggers/<name>/latest.json`) from
`ai/proposals/task-0019-spec-update.md` into M6 if E2E coverage of the
trigger-name path is in scope; otherwise leave as standalone polish.

## Known Spec Drift / Open Questions
- ~~**`bridge-mirror-failed` payload schema not pinned in `data-model.md` §9**~~
  CLOSED in M5.b.
- **`MirrorMode` trinary surface** (Task 0015 adjudicated, accepted with Risk
  Note). Reconsider when M5/M6 remote-driver Phase 2 wiring picks the right name.
- ~~**`MirrorRunnerOutput` has no production callers until M5.b.**~~ CLOSED.
- **`MirrorModeHardlink` is currently a test/drift-detection mode.** If no
  production caller emerges by **M6**, fold into a debug flag (carry-forward
  decision point arriving now).
- ~~**`emitFailure` is best-effort**~~ ADDRESSED in M5.c. Future work
  (M6 candidate): parse `data-model.md` §9.1 fields for richer diagnostics.
- **Event-sequence retry budget of 32** is acceptable for single-writer
  Phase 1; re-evaluate when remote drivers come online.
- **Manifest required for `UpdateLatestExecutionSummary`** (Task 0013
  carry-forward). Pin normatively in `data-model.md` §4 via proposal if
  any M6 path needs the option to skip the manifest step.
- ~~**Legacy-execution literal defaults**~~ Migrate command in M5.d
  read state.json's existing fields; no new normative literals
  introduced. Carry-forward considered closed.
- **`internal/executionstate` coverage at 90.0 % exact floor.** Carry-
  forward risk: small refactors deleting covered branches could trip the
  gate. M6 E2E tests should bump headroom.
- **`RunnerHooks.AfterStateUpdate` fires bridge mirror synchronously on
  the runner goroutine** (Task 0018 carry-forward). M6 E2E may want to
  measure real-workload impact and decide if buffered channel + dedicated
  goroutine is needed.
- **NEW (Task 0020 carry-forward): unknown-hash placeholder body.**
  Migrate writes a sentinel JSON body for orphan executions — by
  construction migrate is the only writer of unknown-hash revisions
  (CreateIfAbsent + ErrExists). Low risk; flag if a future path wants
  to write real plan bytes to the same revision dir.
- **NEW (Task 0020 carry-forward): `hashToRev` dual-keying** (canonical
  `sha256:<hex>` AND bare-hex stem) depends on `state.PlanChecksumShort`
  continuing to emit bare-hex. Covered by tests but worth flagging if
  that helper is ever changed.
- **NEW (Task 0019 carry-forward, partially-resolved): describe alias
  `latest`-literal RESOLVED in M5.d via Option A (CLI normalization);
  trigger-name resolver branch (Option B) still open** — fold into M6
  scope if E2E exercises it, else standalone polish.
- **Persistent local environment quirk (NOT a regression):**
  `kiox -- orun plan --changed --intent examples/intent.yaml` fails on
  composition-cache resolution. Reproduced on every state-redesign
  verifier pass since Task 0014. CI is authoritative.
- **Half-shipped delivery anti-pattern.** Task 0007 first observed; the
  explicit `gh pr list --head` check has shipped every prompt since
  Task 0010 — clean record on Tasks 0010/0012/0014/0016/0018/0019/0020.
