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
**C1 — `internal/sourcectx` resolver.** Per
`specs/orun-component-catalog/implementation-plan.md` §C1 and
`identity-and-keys.md` §2 + §7 + §8 + `design.md` §6.

C0 shipped pure data + the `internal/sourcectx` *type skeleton*
(`WorkspaceState`, `BuildSourceSnapshotKey`, `DirtyHash`,
`CatalogInputHash`). C1 turns that skeleton into a real **resolver**:
given a workspace path + injected `Git`/`Clock`/`Filesystem`, produce a
populated `WorkspaceState` whose `SourceSnapshotKey` is byte-stable
across orderings and whose hashes track only catalog-relevant inputs.

C1 "done when":
- `ResolveSourceSnapshot(ctx, opts) (WorkspaceState, error)` lands
  with adapters for Git (HEAD revision, treeHash, branch, ref, tag,
  PR-number detection), `Clock`, and `Filesystem` (workspace walk +
  catalog-relevant dirty-file enumeration per identity-and-keys.md §7).
- `WorkspaceState.HeadRevision` / `TreeHash` / `Dirty` / `DirtyHash`
  populated from the injected adapters; `CatalogInputHash` exposed
  via `(WorkspaceState).CatalogInputHash(intent CatalogInputHashInputs)`
  or a free function on the package.
- `internal/sourcectx` ≥ 90 % coverage (existing floor).
- T-IDK-3 (key stability across 1 000 random orderings of dirty
  inputs) and **T-IDK-4 (adding a non-catalog-relevant file does NOT
  change `dirtyHash`)** ship as `internal/sourcectx` property tests.
- Clean / dirty / no-git / PR-injected / branch / tag fixtures all
  produce the documented `SourceSnapshotKey` shapes from
  `identity-and-keys.md` §2.
- `--from-ci`-style provider injection produces the documented
  error envelope on no-match (mirrors Phase 1 §11).

## Milestone Sequence (C0 → C9)
| C  | Status | Goal |
|----|--------|------|
| C0 | ✅ done | Spec foundation + pure data models (catalogmodel, sourcectx skeleton) |
| C1 | ▶ active | `internal/sourcectx` resolver (Git HEAD, treeHash, dirtyHash, catalogInputHash) |
| C2 | next | `internal/catalogresolve` — discovery, manifest load, inheritance, inference, deps, validation, manifestHash |
| C3 |       | `internal/catalogstore` — Writer/Resolver, atomic writes under `.orun/sources/` and `.orun/catalogs/` |
| C4 |       | Wire `orun plan` / `orun run` onto SourceSnapshot/CatalogSnapshot |
| C5 |       | TUI cockpit consumes `CatalogSnapshot` (unblocks `.kiro/specs/orun-tui-cockpit`) |
| C6 |       | Compatibility shims — `stateCompatibilityWrites` flag, reader fallback |
| C7 |       | `orun catalog *` CLI surface + global indexes |
| C8 |       | `internal/catalogdiff` (catalog vs catalog comparator) |
| C9 |       | `internal/catalogsync` seam (`Syncer` interface + `NoopSyncer` ONLY — no HTTP, no auth) |

Phase 1 invariants preserved: do not rename Phase 1 fields, do not
lower coverage floors (`internal/statestore` 95.7 %, `internal/revision`
90.3 %, `internal/executionstate` 90.0 %), do not remove preserved
Phase 1 CLI workflows. Phase 2 floors held: `internal/catalogmodel`
90.2 %, `internal/sourcectx` 91.3 %, Sanitize* 100 %.

## Just Completed — Task 0023 (C0 code half)
- **Status:** ✅ Verified PASS and merged via PR #168 (squash commit
  `7f3f2bf`) on 2026-05-31. Verifier report:
  `ai/reports/task-0023-verifier.md`. Implementer report:
  `ai/reports/task-0023-implementer.md`.
- **Outcome on `main`:** `internal/catalogmodel` (15 Go files +
  schema generator + 9 golden fixtures + roundtrip / property /
  sanitize / coverage tests) and `internal/sourcectx` skeleton
  (model / keys / hash + tests) shipped. Canonical encoder is
  byte-deterministic, JSON Schema for `component.yaml` lives at
  `internal/catalogmodel/schema/component-yaml.schema.json` with a
  `make verify-generated` drift gate. Golden roundtrip fixtures cover
  every `data-model.md` schema; T-IDK-1 / T-IDK-3 / T-IDK-5 property
  tests pass.
- **Verifier-attached fix on PR #168:** implementer shipped at
  81.7 % catalogmodel coverage (under the spec-mandated ≥ 90 %
  floor) — the Makefile only gated `Sanitize*` at 100 %. Verifier
  added `internal/catalogmodel/coverage_test.go` (8 tests covering
  `CanonicalEncodeString` / `CanonicalEqual` / `CatalogInputHash` +
  edge paths) and hardened the Makefile with package-level ≥ 90 %
  gates on both `internal/catalogmodel` and `internal/sourcectx`.
- **Local gates on main:** `go build`, `go vet`,
  `go test ./... -race`, `make test-state-redesign`,
  `make verify-generated` all green. Final coverage: catalogmodel
  90.2 %, sourcectx 91.3 %, Sanitize* 100 %, statestore 95.7 %,
  revision 90.3 %, executionstate 90.0 %.

## Current Task (0024)
- **Agent:** Implementer
- **Prompt:** `ai/tasks/task-0024.md`
- **Branch (planned):** `impl/task-0024-c1-sourcectx-resolver`
- **Objective:** ship the C1 resolver — turn a workspace into a
  populated `WorkspaceState` deterministically. Implement
  `ResolveSourceSnapshot(ctx, opts)` with injected `Git` /
  `Clock` / `Filesystem` adapters. Populate `headRevision`,
  `treeHash`, `dirtyHash`, `catalogInputHash`, scope detection
  (branch / PR / tag / local-dirty / local-nogit / ci-event), and
  surface a default Git adapter (in-process go-git **or**
  shell-out — implementer's choice; document the trade-off in the
  PR body).
- **Reads:** `specs/orun-component-catalog/{design.md §6,
  identity-and-keys.md §2 + §7 + §8 + §11, implementation-plan.md
  §C1, test-plan.md (T-IDK-3, T-IDK-4)}`.
- **PR Boundary:** new code under `internal/sourcectx/`
  (resolver + adapters + tests + fixtures), Go module deps for
  the chosen Git adapter (one of `go-git` or `os/exec` shell-out
  to `git`). **No CLI changes, no storage writes, no resolver
  consumption from other internal packages yet.** Zero edits to
  `internal/catalogmodel/` or any Phase 1 package.
- **Acceptance:** `go build ./...` + `go vet ./...` +
  `go test ./... -race` green; `make test-state-redesign` and
  `make verify-generated` green; `internal/sourcectx` ≥ 90 %
  coverage (existing gate); fixture-based clean / dirty / no-git
  / PR / branch / tag tests producing the spec-documented
  `SourceSnapshotKey`; T-IDK-3 + T-IDK-4 property tests; Phase 1
  + catalogmodel coverage floors held; leaf-clean (no
  `internal/sourcectx` import outside its own tree).

## Repo Checkpoint

| Attribute | Value |
|---|---|
| Branch (local checkout) | `main` (clean) |
| `main` tip | `7f3f2bf` — Task 0023 / C0 code half (PR #168) on 2026-05-31 |
| Open PRs | none |
| Repo health | 🟢 Green — C0 done; ready for C1 |
| Last verified | 2026-05-31 (Task 0023 verifier PASS, merged) |
| Active phase | Phase 2 (orun-component-catalog) |
| Active milestone | C1 (`internal/sourcectx` resolver = Task 0024) |
| Tasks completed | 0001–0005, 0007–0016, 0018–0023 (21 total) |
| Current task | **0024 (C1 resolver)** — implementer prompt emitted at `ai/tasks/task-0024.md` |

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
