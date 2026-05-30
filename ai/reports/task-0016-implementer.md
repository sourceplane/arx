# Task 0016 — Implementer Report

**Status:** PR #161 MERGED to `main` as `7a9c494` on 2026-05-30T12:31:56Z. Verified PASS (Task 0016 verifier, single-pass closure).
**Branch:** `impl/task-0016-m5a-orun-plan-rewire` (from `main` @ `d51e828`).
**PR Number:** [#161](https://github.com/sourceplane/orun/pull/161).
**PR title:** `Task 0016: M5.a — orun plan rewire to revision-first layout (#161)`.

---

## Summary

- Rewired `orun plan` to the canonical revision-first flow per
  `cli-surface.md` §1 and `compatibility-and-migration.md` §1–§2.
  TriggerOccurrence is now resolved via `internal/triggerctx` on every
  invocation (not gated on `--trigger`/`--from-ci`), the plan is persisted
  via `internal/revision.WriteRevision` (canonical
  `.orun/revisions/<key>/{plan,trigger,revision,manifest}.json` +
  `.orun/refs/latest-revision.json` +
  `.orun/refs/triggers/<type>/<name>/{latest,<scope>}.json` +
  `.orun/indexes/revisions/<key>.json`), and byte-identical compat aliases
  (`.orun/plans/<sha256>.json`, `.orun/plans/latest.json`, optional
  `.orun/plans/<name>.json`) are layered on top so every preserved workflow
  in `compatibility-and-migration.md` §1 keeps working untouched.
- Plans now embed `metadata.trigger` (Type/Name + scope/bindings) and
  `metadata.revision` (Key/PlanHash) so `plan.json` is self-describing per
  data-model.md §3.1. The plan hash is the canonical sha256 over the plan
  bytes with `metadata.checksum` and `metadata.revision` cleared — invariant
  under both checksum and revision mutation. Material plan changes do change
  the hash.
- `-o/--output` semantics preserved: when set, the user-specified path is
  *also* written (not in place of the canonical layout) and printed alongside
  the canonical revision path.
- Emitted the new on-success summary block from `cli-surface.md` §1.1 (`✓ Plan
  revision created` / `Revision` / `Trigger` / `Jobs` / `Path` / optional
  `Output`) **before** the legacy `components × envs → jobs` detail line so
  any tooling that scans for the legacy line keeps working.
- Coverage gates green: `internal/statestore` 95.7 % (≥95 %),
  `internal/revision` 90.4 % (≥90 %), `internal/executionstate` 90.0 %
  (exact floor held — package not touched in M5.a).
- Scope discipline held: zero edits to `internal/runner` /
  `internal/runbundle` / `internal/state` / `internal/executionstate`.
  M5.b (`orun run` rewire), M5.c (`orun status`/`logs`/`describe`/`get plans`),
  and M5.d (hidden `orun state migrate`) are all explicitly out of scope and
  remain unchanged.

## Files Changed

- `cmd/orun/main.go` (+193, −59) — replaced the legacy `state.SavePlan` branch
  with the full revision pipeline (`statestore.NewLocalStore` →
  `revision.Config{}.WithCompatibilityWrites(true)` → `WriteRevision` →
  `WriteManifest` → `WriteLegacyNamedPlan`). Added two helpers:
  `computePlanHashForRevision` (canonical sha256 over plan bytes with
  `checksum` + `revision` cleared, per data-model.md §3.1) and
  `canonicalPlanJSON` (deterministic indented JSON for byte-identical compat
  aliases). Emit cli-surface.md §1.1 summary block.
- `internal/model/plan.go` (+18) — added `PlanRevisionMeta` (Key/PlanHash) and
  `PlanTrigger.Type`/`Name` fields. Strictly additive; no signature changes.
- `internal/revision/legacy.go` (+30) — added `WriteLegacyNamedPlan` helper
  backing the preserved `orun plan --name <n>` workflow. Validates the path
  component through `statestore.ValidateComponent` and refuses the reserved
  `latest` name.
- `cmd/orun/command_plan_revision_test.go` (new, 4 tests) — covers plan-hash
  invariance under `metadata.checksum` and `metadata.revision` mutation,
  material-change detection, canonical JSON round-trip, nil-plan paths.
- `internal/revision/legacy_test.go` (new, 4 tests) — covers byte-identical
  alias write, reserved-name (`latest`) refusal, bad component name rejection
  via `statestore.ValidateComponent`, nil-store error path.
- `ai/tasks/task-0016.md`, `ai/reports/task-0016-verifier.md`,
  `ai/state.json`, `ai/context/current.md`, `ai/context/task-ledger.md`,
  `ai/waiting_for_input.md` — orchestrator state tracking.

Total diff: 10 files changed, +505 / −65.

## Checks Run

- `go build ./...` → clean.
- `go vet ./...` → clean.
- `go test -race -count=1 ./...` → all packages pass under `-race`.
- `make test-state-redesign` → all gates green:
  - `internal/testfx/statefs` ok
  - `internal/triggerctx` ok
  - `internal/statestore` 95.7 % (gate ≥ 95 %)
  - `internal/revision` 90.4 % (gate ≥ 90 %)
  - `internal/executionstate` 90.0 % (gate ≥ 90 %, exact floor held; package
    not touched)
- E2E smoke against `examples/intent.yaml` — confirmed canonical layout
  populated:
  ```
  .orun/revisions/rev-manual-no-git-p67a8f469/{plan,trigger,revision,manifest}.json
  .orun/refs/latest-revision.json
  .orun/refs/triggers/system.manual/{latest,manual}.json
  .orun/indexes/revisions/rev-manual-no-git-p67a8f469.json
  .orun/plans/<sha256>.json
  .orun/plans/latest.json
  .orun/plans/<name>.json     (when --name passed)
  .orun/component-tree.yaml   (legacy)
  .orun/version.json          (legacy)
  ```
  Summary block renders before legacy detail line; `-o` extra-copy honoured.
- PR CI on final head SHA after verifier-side commit `01e75bd`:
  - `CI / Orun Plan` — run `26683860043` (44s, **PASS**).
  - `orun remote-state conformance / Harness dry-run guard` — run
    `26683860052` (15s, **PASS**).
  Both required checks PASS at log level.

## Assumptions

- **Canonical plan hash uses cleared self-referential metadata.**
  `data-model.md` §3.1 specifies plan hash is computed over the plan bytes
  with `metadata.checksum` and `metadata.revision` cleared. The implementer
  hash helper (`computePlanHashForRevision`) marshals the plan with both
  fields cleared before hashing, then re-embeds `metadata.revision` for
  persistence. This makes the hash invariant under (a) re-rendering with a
  new checksum and (b) re-persisting under a different revision key —
  asserted by `TestComputePlanHashForRevision_StableAcrossChecksumAndRevision`.
- **Compatibility-writes mode for the writer.**
  `revision.Config{}.WithCompatibilityWrites(true)` is set so
  `WriteRevision` populates the legacy `.orun/plans/` aliases byte-identical
  with `revisions/<key>/plan.json`. This is the documented opt-in for §1.1
  preserved workflows in `compatibility-and-migration.md`.
- **`-o/--output` is purely additive.**
  When set, the user-specified path is written via `os.WriteFile` *after* the
  canonical layout and refs are committed; failure to write `-o` is reported
  but does not roll back the canonical revision (which has already been
  durably published). Matches `cli-surface.md` §1.2's "extra copy" semantics.
- **Reserved alias name `latest`.**
  `WriteLegacyNamedPlan` rejects the literal name `latest` because
  `WriteRevision` already manages `.orun/plans/latest.json` as the
  authoritative pointer to the just-written revision; allowing user override
  would let a stray `--name latest` desync that pointer.
- **Trigger always resolved.**
  `internal/triggerctx.ResolveTriggerContext` is called unconditionally, with
  `system.manual` as the floor when no provider event/CI env is available.
  Removes the prior "gated on `--trigger`/`--from-ci`" branch; aligns with
  cli-surface.md §1.3 "internal flow" which mandates trigger context on
  every plan.

## Spec Proposals

None filed. Implementation matches existing spec language; no ambiguities
required adjudication.

## Risk Notes (carried forward unchanged from Task 0015)

- `bridge-mirror-failed` payload schema still un-pinned in `data-model.md`
  §9. M5.a does not touch this; pin during M5.b runner wiring before any
  second consumer (metrics, `orun status`) lands.
- `MirrorRunnerOutput` still has no production callers until M5.b. Resolver
  legacy-fallback (PR-A) carries convergence burden in the meantime.
- `MirrorModeHardlink` debug-fold decision deferred to M6.
- `emitFailure` is best-effort — events-dir-unwritable failures are silently
  dropped. M5.b should add stderr/metric fallback.
- Event-sequence retry budget of 32 acceptable for single-writer Phase 1;
  re-evaluate when remote drivers come online.
- `internal/executionstate` coverage at exact 90.0 % floor — must not regress
  in M5.b/c/d.

## Remaining Gaps (M5.b inherits)

- **`orun run` rewire.** Resolve `PlanRevision` via
  `internal/revision.ResolveRevision`, materialize an in-memory
  `system.manual` revision when none exists, create executions via
  `internal/executionstate.CreateExecution`, hook the runner snapshot stream
  into `Bridge.MirrorRunnerOutput` once per terminal-state transition, add
  `--revision <key>` flag (skips resolution chain), pin
  `bridge-mirror-failed` payload schema in `data-model.md` §9.
- **`orun status` / `logs` / `describe` / `get plans` rewire** — M5.c.
  These continue to read via the legacy resolver and pick up the new layout
  transparently through the resolver's legacy fallback; full rewire onto
  `refs/latest-execution.json` + `indexes/executions/` is M5.c.
- **Hidden `orun state migrate`** — M5.d.

## Next Task Dependencies (Task 0018: M5.b `orun run` rewire implementer)

- Plan revisions are now durably persisted with `metadata.revision.{Key,
  PlanHash}` embedded. M5.b's `orun run` can resolve the latest plan via
  `revision.ResolveRevision` (or `--revision <key>` directly) and read
  `Key`/`PlanHash` straight off `plan.json` — no extra round-trip needed.
- Compat aliases at `.orun/plans/<sha256>.json` + `.orun/plans/latest.json`
  remain byte-identical to canonical `plan.json`, so M5.b code paths that
  still read the legacy alias get the same bytes (including embedded
  trigger/revision metadata) as canonical.
- The `system.manual` trigger is the floor when no provider/CI context
  exists — M5.b should mirror this contract when it materializes an
  in-memory revision for a `--plan <hash>` invocation that targets a
  pre-revision-era plan.
- `cli-surface.md` §1.1 summary block format is now established; M5.b's
  `orun run` summary block (cli-surface.md §2.1) should follow the same
  prefix-then-detail-line pattern so existing tooling that scans the legacy
  detail line keeps working.

## PR Number

**#161** — https://github.com/sourceplane/orun/pull/161 — MERGED to `main`
as commit `7a9c494` on 2026-05-30T12:31:56Z. Verifier (Task 0016 single-pass
closure) inspected required CI logs, ran the full local quality gate suite,
and adjudicated PASS with no spec proposals and no carried-forward issues
beyond the Task 0015 risk notes.
