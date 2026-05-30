# Current Roadmap Position

## Active Spec
`specs/orun-state-redesign/` (Phase 1, local-only) — trigger-first revision-first
local state model. See `specs/orun-state-redesign/README.md` for the index and
read order.

## Active Milestone
**M2 — `internal/statestore`** (frozen interface + local driver). PR A landed
(`9b0a39c`, #154); PR B landed (`0fa2111`, #155, verified PASS); **PR C — typed
refs/indexes marshallers — is the current implementer task** and closes M2.

## Current Task
**Task 0005 — M2 PR-C (Implementer)** — scoped 2026-05-30. Prompt:
`ai/tasks/task-0005.md`. Suggested branch: `impl/task-0005-m2-statestore-prc`.

Adds:

- `internal/statestore/refs.go` — typed reader/writer + CAS helpers for all four
  ref shapes in `data-model.md` §6 (`LatestRevisionRef`, `LatestExecutionRef`,
  `TriggerRef`, `NamedRef`).
- `internal/statestore/indexes.go` — typed index writers for both shapes in
  `data-model.md` §7 (`RevisionIndexEntry`, `ExecutionIndexEntry`) plus the
  `RebuildIndexes()` stub returning `%w: rebuild deferred to M3+` wrapping
  `ErrInvalid`.
- Round-trip / `ErrNotFound` / `ErrExists` / CAS-conflict / JSON-byte-stability
  tests via `internal/testfx/statefs.AssertJSONFile`.

Strict constraints: zero string concatenation for paths (everything via
`paths.go`); deterministic JSON marshalling with trailing `\n` and no HTML
escaping; CAS helpers take `*ObjectMeta` from prior read (no re-read inside
the helper); index writers use `CreateIfAbsent` (re-write → `ErrExists`);
package stays leaf-clean (zero `internal/*` deps); NO production-caller
wiring (`cmd/orun`, `internal/state`, `internal/runner`, `internal/runbundle`
untouched); no new error sentinels; no new path helpers in `paths.go` (if
missing → write proposal, do not add silently).

Coverage gate stays ≥ 95 % on `internal/statestore`; PR-C should leave it
≥ 96 % (M2 stretch target).

## Last Completed Task (0004 — Verifier PASS, merged)
- Implementer prompt: `ai/tasks/task-0004.md`
- Verifier prompt:    `ai/tasks/task-0004-verifier.md`
- Implementer report: `ai/reports/task-0004-implementer.md`
- Verifier report:    `ai/reports/task-0004-verifier.md` (Result: PASS, no blocking issues, no spec proposals)
- PR **#155** (`impl/task-0004-m2-statestore-prb`) verified PASS and squash-merged
  on 2026-05-30 → main commit `0fa2111`. Branch deleted.
- Durable outcome on main:
  - Real `*LocalStore.CompareAndSwap` per `state-store.md` §3.3:
    Read → revision compare → Write; `ErrNotFound` (via Read), `ErrConflict`
    on revision mismatch; sentinels wrapped with `fmt.Errorf("%w: ...", ErrX, ...)`.
    Per-path `sync.Mutex` narrows the in-process race (additive, §6 permits
    "best-effort on local").
  - Real `*LocalStore.List` per `state-store.md` §3.4: `WalkDir` over translated
    prefix; symlinks skipped; `.orun-tmp-*` filtered; logical paths via
    `filepath.ToSlash`; non-existent prefix → empty slice (no error);
    `ErrInvalid` on alphabet/escape via `paths.go`.
  - 17 new tests covering: 100-goroutine `Write` atomicity, 100-goroutine
    `CreateIfAbsent` exclusivity, CAS exactly-one-wins, `pgregory.net/rapid`
    path-alphabet round-trip with stable lowercase-hex sha256 `Revision`,
    plus `List` edges (symlinks, FIFOs unix-only, tempfile filter, escape
    rejection, ctx cancel).
  - Coverage measured 95.4 % on `internal/statestore` (gate ≥ 95 %).
  - No production-caller wiring; package stays leaf-clean.
- CI evidence: `CI / Orun Plan` run **26670829548** observed real
  `orun plan --from-ci github …` invocation with `0 components × 3 envs → 0
  jobs` (legitimate empty matrix at M2 PR-B). `orun remote-state conformance /
  Harness dry-run guard` run **26670829550** logged the full `[guard] PASS:`
  battery (bash syntax, command-count thresholds, duplicate-claim helper PASS
  + FAIL, status helper PASS + FAIL, exported env asserts).
- Verifier non-blocking findings (carried forward, none blocking):
  per-path `sync.Mutex` is in-process only (irrelevant to local Phase 1
  single-process semantics; remote driver Phase 2 supersedes); empty-directory
  `Delete` returns `ErrInvalid` (state-store.md §3.4 only mandates non-empty —
  carried from Task 0003); coverage 95.4 % satisfies gate but sits below the
  96 % stretch target — PR-C will add `refs.go` / `indexes.go` and should lift it.

## Repo Checkpoint

| Attribute | Value |
|---|---|
| Branch | main (synced with origin/main) |
| Last commit on main | `0fa2111` — Task 0004: M2 PR-B — statestore `CompareAndSwap` + `List` (#155) |
| Open PRs (state-redesign lineage) | none |
| Repo health | 🟢 Green |
| Last verified | 2026-05-30 (Task 0004, PR #155) |
| Active milestone | M2 (`internal/statestore`) — PR A + PR B landed; PR C scoped |
| Tasks completed | 0001, 0002, 0003, 0004 (4 total) |
| Current task | **0005** (scoped, awaiting implementer) |

## Roadmap (M0 → M6)
1. ✅ **M0 Foundation** — landed on main at `4ea1980` (PR #152).
2. ✅ **M1 `internal/triggerctx`** — landed on main at `db342dd` (PR #153).
3. **M2 `internal/statestore`** ← current
   - ✅ PR A — frozen interface + local driver non-CAS ops (PR #154 → `9b0a39c`)
   - ✅ PR B — `CompareAndSwap`, `List`, atomicity/exclusivity/CAS/`rapid` property suite (PR #155 → `0fa2111`)
   - 🟡 PR C — typed refs/indexes marshallers + `RebuildIndexes()` stub (Task 0005 scoped)
4. M3 `internal/revision`
5. M4 `internal/executionstate` + runner bridge
6. M5 CLI rewire (`orun plan/run/status/logs/describe/get plans` + hidden `state migrate`)
7. M6 End-to-end + property gates

## Next Task After 0005 (proposed)
**Task 0006 — M2 PR-C Verifier** then immediately **Task 0007 — M3 PR-A
(Implementer)** opening the `internal/revision` package: `model.go`
(`PlanRevision`, `RevSummary`), `keys.go` (`RevisionKey(trig, planHash)`,
regex validator, collision suffix logic), and `writer.go` skeleton with the
ordered writes from `design.md` §5.1 / writer-order list (compatibility
writes gated on `stateCompatibilityWrites`, default true). Manifest +
`ResolveRevision` may split into PR-B per `implementation-plan.md` Milestone
M3 suggested PR scope (1–2 PRs). Coverage gate ≥ 90 % per "Done when".

If Task 0005 implementer hits an unrecoverable spec ambiguity (e.g. missing
path helper, ref shape disagreement with `data-model.md` §6), the implementer
files a proposal under `ai/proposals/task-0005-spec-update.md` and pauses;
orchestrator scopes a spec-update task next instead of M3.

## Known Spec Drift / Open Questions
- Persistent local-only `kiox -- orun plan --changed --intent
  examples/intent.yaml` failure on composition-cache resolution
  (`stack.yaml at ~/.orun/cache/compositions/c41fc08… has no
  spec.compositions`). Reproduces sporadically from Task 0001 onward on main
  and on PR branches; Task 0004 verifier observed it DID NOT reproduce on
  that run. Flaky local cache, not a code regression. CI is authoritative.
- Optional clarification: `state-store.md` §3.4 could explicitly state
  whether empty-directory `Delete` succeeds or returns `ErrInvalid`. The
  PR-A implementation chose the conservative refuse-all-directories path
  and PR-B did not change it. File a proposal under `ai/proposals/` if a
  future caller needs the loosened behavior.

## Secondary Specs (not driving new tasks this phase)
- `.kiro/specs/orun-tui-cockpit/` — paused. Resumes after M5 lands.
- `.kiro/specs/github-artifacts/` — cross-check only; new revision/execution
  keys must remain compatible with the existing
  `gh-{run_id}-{attempt}-{sha}` ExecID shape produced by `internal/runbundle`.
