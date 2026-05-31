# Current Roadmap Position

## Active Spec
`specs/orun-component-catalog/` (Phase 2, local-only) — content-addressed
SourceSnapshot/CatalogSnapshot model wrapping the Phase 1 trigger /
revision / execution lineage. **Local-only** for the entire phase: no
HTTP, no SaaS, no DB schema. `internal/catalogsync` ships only `Syncer`
interface + `NoopSyncer` (C9).

## Active Milestone
**C2 — `internal/catalogresolve` (discovery + manifest resolver).** Per
`specs/orun-component-catalog/implementation-plan.md` §C2 and
`resolution-pipeline.md` stages 2 / 3 / 5 / 6 / 8 / 9 / 10.

Spec suggests 2 PRs:
- **C2 PR-1 (Task 0025, ▶ active):** discover + load + inherit. No
  inference, no deps resolution, no validation matrix, no `manifestHash`.
  Output: in-memory `[]AuthoredManifest` from `DiscoverAndLoad`.
- **C2 PR-2 (Task 0026, queued):** infer + deps + validate +
  `manifestHash`. Output: full `Resolve(ctx, opts)` returning the
  resolved (but pre-graph) manifests with byte-stable `manifestHash`.

C2 "done when" (across both PRs):
- `internal/catalogresolve` ≥ 90 % coverage.
- Component fixture in `testdata/repo/` produces a byte-identical
  `[]ComponentManifest` across two consecutive `Resolve` calls (T-RES-1).
- Inheritance provenance populated for every inherited / inferred field
  (T-RES-2).
- Broken dependency reports `ErrDependencyMissing` with both endpoints.
- Cycle in `deploy-after` aborts; cycle in `calls` warns by default.

## Milestone Sequence (C0 → C9)
| C  | Status | Goal |
|----|--------|------|
| C0 | ✅ done (PR #168 / `7f3f2bf`) | Spec foundation + pure data models |
| C1 | ✅ done (PR #169 / `b50d799`) | `internal/sourcectx` resolver |
| C2 | ▶ active | `internal/catalogresolve` — discovery + manifest resolver (Tasks 0025 + 0026) |
| C3 | next  | `CatalogSnapshot` + graph builder + `catalogHash` |
| C4 |       | `internal/catalogstore` Writer + Resolver atomic writes |
| C5 |       | Catalog CLI surface |
| C6 |       | `orun plan` integration |
| C7 |       | `orun run` integration + history events |
| C8 |       | `internal/catalogdiff` + validate + rebuild |
| C9 |       | `internal/catalogsync` seam (`Syncer` + `NoopSyncer` ONLY — no HTTP, no auth) |

Phase 1 invariants preserved: do not rename Phase 1 fields, do not
lower coverage floors (`internal/statestore` 95.7 %, `internal/revision`
90.3 %, `internal/executionstate` 90.0 %), do not remove preserved
Phase 1 CLI workflows. Phase 2 floors held: `internal/catalogmodel`
91.1 %, `internal/sourcectx` 91.1 %, Sanitize* 100 %.

## Just Completed — Task 0024 (C1 resolver)
- **Status:** ✅ Verified PASS and merged via PR #169 (squash commit
  `b50d799`) on 2026-05-31. Reports:
  - Implementer: `ai/reports/task-0024-implementer.md`
  - Verifier: `ai/reports/task-0024-verifier.md`
- **Outcome on `main`:** `internal/sourcectx` resolver shipped with
  `ResolveSourceSnapshot(ctx, opts)`, Git/Clock/Filesystem adapters,
  `WorkspaceState` populated with `headRevision`, `treeHash`,
  `dirtyHash`, `catalogInputHash`. T-IDK-3 (key stability across
  random orderings) and T-IDK-4 (non-catalog files don't change
  `dirtyHash`) ship as property tests.
- **Verifier-attached fix:** added
  `internal/catalogmodel/coverage_test.go::TestCanonicalEncodeStringEscapeBranches`
  to deterministically pin the C0 catalogmodel coverage floor — rapid
  generators were probabilistically missing `\b` / `\f` escape branches
  in `writeQuotedString`, dropping coverage from 90.2 % to 87.9 % on
  some seeds. Post-fix: catalogmodel 91.1 % × 19 / 90.6 % × 1 across
  20 runs.
- **Local gates on main:** `go build`, `go vet`, `go test ./... -race`,
  `make test-state-redesign`, `make verify-generated`,
  `kiox -- orun validate --intent intent.yaml` all green.

## Just Completed — Task 0025 (C2 PR-1 discover/load/inherit)
- **Status:** ✅ Verified PASS and merged via PR #170 (squash commit
  `723be32`) on 2026-05-31T07:06:29Z. Reports:
  - Implementer: `ai/reports/task-0025-implementer.md`
  - Verifier: `ai/reports/task-0025-verifier.md`
  - Spec proposal: `ai/proposals/task-0025-spec-update.md`
- **Outcome on `main`:** `internal/catalogresolve` online with
  `DiscoverAndLoad(ctx, Options) (DiscoveryResult, error)`. Walks the
  workspace (default excludes: `.git .orun build dist node_modules
  vendor`, intent excludes appended), loads + Draft-7 schema-validates
  authored `component.yaml` / `component.yml` manifests (mixed-extension
  in same dir is a typed error), applies intent `catalog.defaults`
  inheritance (scalar-zero / per-key-map / wholesale-list rules), and
  emits a deterministic sorted `[]AuthoredManifest` with RFC 6901
  provenance pointers. Mini-T-RES-1 asserted in `resolve_test.go`.
- **Coverage floors on main:** `internal/catalogresolve` **90.0 %** (exact
  gate, deterministic across 3 local + CI runs); `internal/catalogmodel`
  91.1 %, `internal/sourcectx` 91.1 %, Sanitize* 100 %; Phase 1 floors
  byte-for-byte (statestore 95.7 %, revision 90.3 %, executionstate 90.0 %).
- **Verifier-accepted call-outs:**
  1. Additive `internal/catalogmodel/schema_embed.go` (18 lines,
     `//go:embed`-only) ACCEPTED — `//go:embed` cannot escape its
     package, vendoring is forbidden by spec. **Convention adopted
     (load-bearing for Phase 2):** *"One additive file per cross-package
     contract surface in `internal/catalogmodel/`. No edits to existing
     source files. Each additive file is `//go:embed`-only or a small
     read-only typed view — no logic."*
  2. `catalogresolve` 90.0 % no-headroom ACCEPTED WITH RISK NOTE —
     deterministic (no rapid-driven variance), CI matches local
     byte-for-byte; Task 0026 PR-2 will naturally raise the floor.
- **Spec proposal:** `ai/proposals/task-0025-spec-update.md` tightens
  the C2 PR-Boundary wording to *"No edits to **existing source files
  in** `internal/catalogmodel/` or `internal/sourcectx/`. Additive
  sibling files (embed-only exports, small read-only typed views)
  needed by dependent packages are permitted; one additive file per
  cross-package contract surface, no logic."* Fold into Task 0026
  prompt at scope time.
- **Local gates on main:** `go build`, `go vet`, `go test ./... -race`,
  `make test-state-redesign` ×3, `make verify-generated`, `kiox -- orun
  validate --intent intent.yaml`, `go test -count=10 -race
  ./internal/catalogresolve/...` all green.

## Current Task (0026 — C2 PR-2, queued)
- **Agent:** Orchestrator (next cycle scopes the implementer prompt).
- **Goal:** infer + deps + validate + `manifestHash` on top of merged
  catalogresolve. Adds resolution-pipeline stages 6 / 8 / 9 / 10 plus
  `Resolve(ctx, opts)` top-level entry. T-RES-1 (resolver determinism),
  T-RES-2 (provenance completeness), and `ErrDependencyMissing` land
  here. After Task 0026 closes, C2 is done and C3 (CatalogSnapshot +
  graph builder + `catalogHash`) opens.
- **PR-Boundary wording for the prompt:** use the tightened version
  from `ai/proposals/task-0025-spec-update.md` so additive sibling
  files in `catalogmodel/` (e.g. another small typed view if needed)
  remain explicitly permitted without re-litigating.

## Repo Checkpoint

| Attribute | Value |
|---|---|
| Branch (local checkout) | `main` (clean) |
| `main` tip | `723be32` — Task 0025 / C2 PR-1 catalogresolve discover+load+inherit (PR #170) on 2026-05-31 |
| Open PRs | none |
| Repo health | 🟢 Green — C2 PR-1 done; ready for Task 0026 (C2 PR-2) |
| Last verified | 2026-05-31 (Task 0025 verifier PASS, merged) |
| Active phase | Phase 2 (orun-component-catalog) |
| Active milestone | C2 (`internal/catalogresolve` — Tasks 0025 ✅ + 0026 queued) |
| Tasks completed | 0001–0005, 0007–0016, 0018–0025 (23 total) |
| Current task | **0026 (C2 PR-2: infer + deps + validate + manifestHash)** — orchestrator to scope |

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
- **Task 0020 carry-forward: unknown-hash placeholder body** + `hashToRev`
  dual-keying depends on `state.PlanChecksumShort` continuing to emit
  bare-hex.
- **Task 0019 carry-forward: Option B trigger-name resolver branch**
  still open; fold into a Phase 2 milestone if/when E2E exercises it.
- **Persistent local environment quirk (NOT a regression):**
  `kiox -- orun plan --changed --intent examples/intent.yaml` fails on
  composition-cache resolution on this developer machine. Reproduced on
  every state-redesign verifier pass since Task 0014. CI is authoritative.
