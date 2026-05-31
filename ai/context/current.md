# Current Roadmap Position

## Active Spec
`specs/orun-component-catalog/` (Phase 2, local-only) — content-addressed
SourceSnapshot/CatalogSnapshot model wrapping the Phase 1 trigger /
revision / execution lineage. See
`specs/orun-component-catalog/README.md` for the doc index and read
order. **Local-only** for the entire phase: no HTTP, no SaaS, no DB
schema. `internal/catalogsync` ships only `Syncer` interface +
`NoopSyncer` (C9).

## Active Milestone
**C0 — Spec foundation and pure data models.** Per
`specs/orun-component-catalog/implementation-plan.md`:

- C0 is split across two PRs:
  - **Task 0022 (this cycle)** — *spec-landing / rollover half*:
    docs and bookkeeping only. Lands the 12-doc spec pack, the
    rewritten `agents/orchestrator.md`, the rotated `ai/state.json`,
    refreshes the context files, and disposes of the root-level
    `orun-catalog-full-design.md` monolith (archived under
    `specs/orun-component-catalog/_archive/full-design-monolith.md`).
    No Go code.
  - **Task 0023 (next cycle)** — *C0 code half*: ships
    `internal/catalogmodel` (data types + canonical-JSON encoder +
    JSON-Schema generator + golden roundtrip fixtures), the
    `internal/sourcectx` skeleton (types only, no resolver yet), and
    extends `make test-state-redesign` to gate the new packages.
- C0 "done when": `go build ./...` and `go test ./...` green;
  catalogmodel decode/encode roundtrip per data-model.md schema;
  `Sanitize*` 100 % coverage; `internal/catalogmodel` imports no
  other `internal/` package.

## Milestone Sequence (C0 → C9)
| C  | Goal |
|----|------|
| C0 | Spec foundation + pure data models (catalogmodel, sourcectx skeleton) |
| C1 | `internal/sourcectx` resolver (Git HEAD, treeHash, dirtyHash, catalogInputHash) |
| C2 | `internal/catalogresolve` — discovery, manifest load, inheritance, inference, deps, validation, manifestHash |
| C3 | `internal/catalogstore` — Writer/Resolver, atomic writes under `.orun/sources/` and `.orun/catalogs/` |
| C4 | Wire `orun plan` / `orun run` onto SourceSnapshot/CatalogSnapshot |
| C5 | TUI cockpit consumes `CatalogSnapshot` (unblocks `.kiro/specs/orun-tui-cockpit`) |
| C6 | Compatibility shims — `stateCompatibilityWrites` flag, reader fallback |
| C7 | `orun catalog *` CLI surface + global indexes |
| C8 | `internal/catalogdiff` (catalog vs catalog comparator) |
| C9 | `internal/catalogsync` seam (`Syncer` interface + `NoopSyncer` ONLY — no HTTP, no auth) |

Phase 1 invariants preserved: do not rename Phase 1 fields, do not
lower coverage floors (`internal/statestore` 95.7 %, `internal/revision`
90.3 %, `internal/executionstate` 90.0 %), do not remove preserved
Phase 1 CLI workflows.

## Current Task (0022)
- **Agent:** Implementer
- **Prompt:** `ai/tasks/task-0022.md`
- **Branch:** `impl/task-0022-phase2-rollover`
- **Objective:** docs-and-bookkeeping rollover from Phase 1 → Phase 2.
- **PR Boundary:** spec pack + orchestrator.md + state.json + context-file
  refresh + monolith disposition. **No Go code.**
- **Constraints:** spec content shipped as-is (mechanical fixes only;
  substantive changes go through `/ai/proposals/`); Phase 1 specs
  untouched; Deferred Decision Protocol wording added to
  `agents/orchestrator.md` (was missing); PR ≤ ~3 200 LOC, clean commit
  history.
- **Acceptance:** 12 spec docs present; orchestrator.md cites new spec;
  state.json rotated; monolith archived under `_archive/`;
  `go build ./...` + `go test ./...` green; required CI PASS on PR head.

## Next Task After 0022
**Task 0023 — C0 code half.** Reads
`specs/orun-component-catalog/{data-model.md, identity-and-keys.md, implementation-plan.md§C0, test-plan.md§1}`.
Ships `internal/catalogmodel` (10 files per implementation-plan.md C0 list)
with canonical-JSON encoder, golden roundtrip fixtures under
`testdata/`, JSON-Schema generated under `internal/catalogmodel/schema/`
via `go generate ./...`. Adds `internal/sourcectx` skeleton (types
only). Extends `make test-state-redesign` to gate the new packages.
Target: ≥ 90 % coverage on `internal/catalogmodel`; 100 % on
`Sanitize*`; leaf-clean imports.

## Repo Checkpoint

| Attribute | Value |
|---|---|
| Branch (local checkout) | `impl/task-0022-phase2-rollover` (PR pending) |
| `main` tip | `b4178dd` — release-notes PR #166 (orun-state-redesign Phase 1 / v2.10.0) on 2026-05-31 |
| Open PRs | Task 0022 (this PR) — Phase 2 rollover, docs-and-bookkeeping only |
| Repo health | 🟢 Green — Phase 1 closed; Phase 2 rolling over |
| Last verified | 2026-05-31 (Task 0022 rollover branch) |
| Active phase | Phase 2 (orun-component-catalog) |
| Active milestone | C0 (spec-landing half = Task 0022; code half = Task 0023) |
| Tasks completed | 0001, 0002, 0003, 0004, 0005, 0007, 0008, 0009, 0010, 0011, 0012, 0013, 0014, 0015, 0016, 0018, 0019, 0020, 0021 (19 total) |
| Current task | **0022 (Phase 2 rollover)** — implementer in progress |

---

# Past Phase — orun-state-redesign (Phase 1, COMPLETE)

Phase 1 (`specs/orun-state-redesign/`, M0–M6) closed via PR #165
(`ad3656e`) on 2026-05-30 with release-notes PR #166 (`b4178dd`)
on 2026-05-31. Coverage floors on `main` at phase close:
`internal/statestore` 95.7 %, `internal/revision` 90.3 %,
`internal/executionstate` 90.0 %.

| M  | PR    | Main commit |
|----|-------|-------------|
| M0 | #152  | `4ea1980`   |
| M1 | #153  | `db342dd`   |
| M2 | #156  | `cd8b3e8`   |
| M3 | #158  | `bfc2ae6`   |
| M4 | #159 / #160 | `ed48633` / `d51e828` |
| M5 | #161–#164 | `7a9c494` … `17ef788` |
| M6 | #165  | `ad3656e`   |

Phase 1 carry-forward (candidates for follow-on within Phase 2 scope,
NOT yet wired): MirrorModeHardlink debug-fold decision,
RunnerHooks.AfterStateUpdate async-mirror evaluation, `--persist-revision`
flag wiring, Option B trigger-name resolver branch
(`ai/proposals/task-0019-spec-update.md`), `--prune-legacy`. None of
these block Phase 2.

## Phase 1 closure summary (Tasks 0021 / M6)
- Single-pass closure (implementer + verifier in one cycle).
- PR **#165** on `impl/task-0021-m6-e2e-and-property-gates`. Squash-merged
  to `main` as `ad3656e` "Task 0021 / M6: end-to-end + property gates
  for state redesign (#165)" on 2026-05-30T19:17:23Z.
- Required CI both PASS on PR head: `Orun Plan` SUCCESS (53 s);
  `Harness dry-run guard` SUCCESS (15 s). Matrix legs SKIPPED
  legitimately (empty matrix — test-only change).
- The `internal/revision` floor had been silently breached at 84.9 %
  on main tip pre-M6; M6 restored the documented ≥ 90 % floor without
  lowering any threshold.

## Known Spec Drift / Open Questions (Phase 1 carry-forward)
- **`MirrorMode` trinary surface** (Task 0015 adjudicated, accepted with Risk
  Note). Reconsider when Phase 2 remote-driver wiring picks the right name.
- **`MirrorModeHardlink` is currently a test/drift-detection mode.** If no
  production caller emerges in Phase 2, fold into a debug flag.
- **Event-sequence retry budget of 32** is acceptable for single-writer
  Phase 1; re-evaluate when remote drivers come online (Phase 3).
- **Manifest required for `UpdateLatestExecutionSummary`** (Task 0013
  carry-forward). Pin normatively if any Phase 2 path needs to skip the
  manifest step.
- **`internal/executionstate` coverage at 90.0 % exact floor.** Carry-
  forward risk: small refactors deleting covered branches could trip the
  gate. Phase 2 work should bump headroom.
- **`RunnerHooks.AfterStateUpdate` fires bridge mirror synchronously on
  the runner goroutine** (Task 0018 carry-forward). Phase 2 may want to
  decide if buffered channel + dedicated goroutine is needed.
- **Task 0020 carry-forward: unknown-hash placeholder body.** Migrate
  writes a sentinel JSON body for orphan executions. Low risk; flag if
  a future path wants to write real plan bytes to the same revision dir.
- **Task 0020 carry-forward: `hashToRev` dual-keying** (canonical
  `sha256:<hex>` AND bare-hex stem) depends on `state.PlanChecksumShort`
  continuing to emit bare-hex.
- **Task 0019 carry-forward: Option B trigger-name resolver branch**
  still open (`ai/proposals/task-0019-spec-update.md`); fold into a
  Phase 2 milestone if/when E2E exercises the trigger-name path.
- **Persistent local environment quirk (NOT a regression):**
  `kiox -- orun plan --changed --intent examples/intent.yaml` fails on
  composition-cache resolution on this developer machine. Reproduced on
  every state-redesign verifier pass since Task 0014. CI is authoritative.
