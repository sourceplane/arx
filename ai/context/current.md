# Current Roadmap Position

## Active Spec
`specs/orun-component-catalog/` (Phase 2, local-only) — content-addressed
SourceSnapshot/CatalogSnapshot model wrapping the Phase 1 trigger /
revision / execution lineage. **Local-only** for the entire phase: no
HTTP, no SaaS, no DB schema. `internal/catalogsync` ships only `Syncer`
interface + `NoopSyncer` (C9).

## Active Milestone
**C4 — `internal/catalogstore` Writer + Resolver atomic writes.** C3
closed on 2026-05-31 via Task 0028 / PR #172 (squash `75082ca`). C4 is
the next milestone; Task 0030 will scope it.

C2 (`internal/catalogresolve`) closed on 2026-05-31 via Tasks 0025 (PR #170,
`723be32`) and 0026 (PR #171, `74b88e0`). C3 closed via Task 0028 / PR #172.

## Just Completed — Task 0028 + 0029 (C3 — `CatalogSnapshot` + graph builder + `catalogHash`)
- **Status:** ✅ Verified PASS (Task 0029) and merged via PR #172 (squash
  commit `75082ca`) on 2026-05-31. Reports:
  - Implementer: `ai/reports/task-0028-implementer.md`
  - Verifier: `ai/reports/task-0029-verifier.md`
- **Outcome on `main`:** `internal/catalogresolve` gained the post-
  resolution layer wiring `resolution-pipeline.md` §1 stages 11
  (graph build) → 12 (`catalogHash`) → 13 (snapshot assemble + key
  derive + Source-block back-fill). New exported entry point:
  `BuildCatalog(ctx, opts, ResolverInputs) (*CatalogView,
  []ValidationIssue, error)`. `CatalogView` carries the existing
  `*ResolvedCatalog` plus `*CatalogSnapshot` and `[]*CatalogGraph`.
  Five `CatalogGraph` siblings (`dependencies`, `systems`, `apis`,
  `resources`, `owners`) emitted in the §9 fixed order with sorted
  nodes-by-`key` and edges-by-`(from, to, type, optional)`.
  `ResolverInputs` is fully caller-supplied; missing fields produce a
  typed `ErrResolverInputsIncomplete` (extractable via
  `IsResolverInputsIncomplete(err)`). Three new files
  (`graph.go`, `catalog_hash.go`, `catalog_snapshot.go`) plus their
  test siblings; convention from C2 PR-1 honoured (no edits to
  existing source files).
- **Properties proven:**
  1. T-IDK-1 — 1000 random orderings of the manifest input bundle ⇒
     identical `catalogHash` (rapid).
  2. `metadata.owner` edit changes both `manifestHash` AND `catalogHash`
     (deterministic backstop).
  3. `resolution.inheritedFrom` (provenance-only) edit does NOT change
     `manifestHash` AND does NOT change `catalogHash` (T-IDK-2 propagated
     through to the C3 layer).
  4. Two consecutive `BuildCatalog` calls produce byte-identical
     canonical-encoded `(*CatalogSnapshot, []*CatalogGraph)` after
     clearing the per-call ULID `CatalogSnapshotID`.
  5. `summary.*` counts equal sorted-distinct enumeration (components /
     systems / apis / resources / owners / domains).
  6. `catalogSnapshotKey` matches `^cat-[a-f0-9]{6,16}$` (width 8 default;
     collision policy `-x<n>` left to C4 writer).
  7. `manifestHash` invariant held: Source block fully excluded from the
     hashed payload (`hash.go`), so the post-stage-13 back-fill of
     `Source.{SourceSnapshotKey, CatalogSnapshotKey, HeadRevision,
     TreeHash, WorkingTree}` is safe by construction.
- **Coverage floors on main:** `internal/catalogresolve` **90.9 %**
  (90.2 → 90.9, +0.7 pp); Phase 2 floors held byte-for-byte
  (catalogmodel 91.1 %, sourcectx 91.1 %, Sanitize* 100 %); Phase 1
  floors held (statestore 95.7 %, revision 90.3 %, executionstate
  90.0 %).
- **Local gates on main:** `go build`, `go vet`, `go test ./... -race`,
  `make test-state-redesign`, `make verify-generated` all green.
- **Risk note:** the C3 layer trusts the caller to compute
  `Authoritative` / `Preview` correctly (no zero-value sentinel for
  booleans). The C4 writer is the next guardrail (`authoritative=true`
  must imply `workingTree=clean` per data-model §2).

## Current Task — Task 0031 (C4 PR-1 verifier — verify PR #173)
- **Status:** Scoped 2026-05-31. Prompt at `ai/tasks/task-0031-verifier.md`.
- **Target:** PR #173 (branch `task-0030-catalogstore-c4-pr1`,
  `MERGEABLE` / `CLEAN`, CI 4/4 SUCCESS at scope time). Implementer
  report shipped at `reports/task-0030-catalogstore-c4-pr1.md` on the
  PR branch; canonical path `ai/reports/task-0030-implementer.md`
  was not used — verifier's call whether to relocate or accept.
- **Verification scope:**
  - PR-boundary audit (eight `internal/catalogstore/` files only;
    no edits outside that dir; no `refs.go`/`indexes.go`/`resolver.go`).
  - Write-order audit (B.1→B.2→B.3→B.4 in code AND spy test;
    fixed graph order `dependencies, systems, apis, resources, owners`).
  - Pre-flight `ErrInputsInconsistent` exercised for cat↔src,
    manifest↔src, manifest↔cat mismatch shapes BEFORE any write.
  - Idempotence + double-wrap mismatch sentinels (`errors.Is`
    chains to both typed sentinel and `statestore.ErrExists`).
  - Stub policy: `WriteRefs`/`WriteGlobalIndexes`/`AppendComponentEvent`
    + every `Resolver` method return `ErrNotImplemented`, pinned by
    test.
  - Step B.4 uses plain `Write` (NOT CAS) — local indexes are
    rebuildable per spec.
  - No raw FS imports (`os`/`io/ioutil`/`path/filepath`) under
    `internal/catalogstore/`.
  - Coverage: `internal/catalogstore` ≥ 90 %; Phase 1 floors held
    byte-for-byte (statestore 95.7, revision 90.3, executionstate 90.0);
    Phase 2 floors held (catalogmodel 91.1, sourcectx 91.1,
    catalogresolve 90.9).
- **Verifier-only fixes allowed:** moving the implementer report to
  the canonical path (cosmetic), tiny doc polish required to PASS.
  Anything beyond that becomes Task 0031.x.
- **On PASS:** merge PR #173 via Verifier Merge Protocol, sync `main`,
  scope **Task 0032** (C4 PR-2 implementer — `refs.go` + `indexes.go`
  covering write-order steps C/D + `AppendComponentEvent` with
  `seq.lock` retry-up-to-16 contract).
- **On FAIL:** leave PR #173 open with blockers; remediation stays
  in the same PR.

## Just Completed — Task 0030 (C4 PR-1 implementer — emitted PR #173)
- **Status:** Implementer pushed PR #173 on 2026-05-31; CI green;
  awaiting Task 0031 verifier. Branch
  `task-0030-catalogstore-c4-pr1`, head `7fec059`. Implementer
  self-report on the branch at
  `reports/task-0030-catalogstore-c4-pr1.md`.
- **Surface delivered (per implementer report):**
  - `internal/catalogstore/{paths.go, paths_test.go, writer.go,
    writer_test.go, errors.go, store.go, store_test.go, doc.go}`.
  - Path helpers from `catalog-store.md` §2 (sources, catalogs,
    components, refs, local indexes, global indexes, history events)
    plus `Validate*` siblings; helpers return `(string, error)`,
    no panics.
  - Mismatch sentinels `ErrSourceMismatch` / `ErrCatalogMismatch` /
    `ErrManifestMismatch` double-wrap `statestore.ErrExists` so
    `errors.Is` succeeds against both.
  - `ErrInputsInconsistent` pre-flight rejects mismatched (src, cat,
    manifests) tuples BEFORE any write.
  - `WriteSourceSnapshot` step A and `WriteCatalogSnapshot` steps
    B.1→B.2→B.3→B.4 implemented; graph write order fixed via
    `CatalogGraphKinds()`; B.4 local indexes via plain `Write`
    (overwrite-OK).
  - `Writer` / `Resolver` / `Store` interfaces frozen with
    compile-time assertions on `*store`. `WriteRefs`,
    `WriteGlobalIndexes`, `AppendComponentEvent`, all `Resolver`
    methods return `ErrNotImplemented` (pinned by test).
- **Coverage (claimed):** `internal/catalogstore` 90.7 %; Phase 1 +
  Phase 2 floors held. Verifier will re-measure and capture exact %.
- **Implementer assumption to verify:** chose `PrettyEncode` (not
  `CanonicalEncode`) for body writes. `CanonicalEncode` reserved for
  hash inputs (used by upstream layers). Confirm consistency in code.


## Repo Checkpoint

| Attribute | Value |
|---|---|
| Branch (local checkout) | `main` |
| `main` tip | `75082ca` — Task 0028 / C3 (PR #172) on 2026-05-31 |
| Open PRs | **#173** — Task 0030 (C4 PR-1, branch `task-0030-catalogstore-c4-pr1`, MERGEABLE/CLEAN, CI green) |
| Repo health | 🟢 Green — main untouched since C3; PR #173 awaiting verifier |
| Last verified | 2026-05-31 (Task 0029 verifier PASS) |
| Active phase | Phase 2 (orun-component-catalog) |
| Active milestone | C4 (`internal/catalogstore` writer) — PR-1 in flight |
| Tasks completed | 0001–0005, 0007–0016, 0018–0029 (27 total) |
| Current task | **0031 (C4 PR-1 verifier on PR #173)** |
| Next task | TBD — Task 0032 (C4 PR-2 implementer: refs.go + indexes.go + AppendComponentEvent) once 0031 is PASS-and-merged |

## Milestone Sequence (C0 → C9)
| C  | Status | Goal |
|----|--------|------|
| C0 | ✅ done (PR #168 / `7f3f2bf`) | Spec foundation + pure data models |
| C1 | ✅ done (PR #169 / `b50d799`) | `internal/sourcectx` resolver |
| C2 | ✅ done (Tasks 0025 + 0026 / PRs #170 + #171 / `723be32` + `74b88e0`) | `internal/catalogresolve` — discovery + manifest resolver |
| C3 | ✅ done (Task 0028 / PR #172 / `75082ca`) | `CatalogSnapshot` + graph builder + `catalogHash` |
| C4 | ▶ active | `internal/catalogstore` Writer + Resolver atomic writes |
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

## Just Completed — Task 0026 (C2 PR-2 infer + deps + validate + manifestHash)
- **Status:** ✅ Verified PASS (Task 0027) and merged via PR #171 (squash
  commit `74b88e0`) on 2026-05-31T08:36:04Z. Reports:
  - Implementer: `ai/reports/task-0026-implementer.md`
  - Verifier: `ai/reports/task-0027-verifier.md`
- **Outcome on `main`:** top-level
  `Resolve(ctx, opts) (*ResolvedCatalog, []ValidationIssue, error)` covering
  resolution-pipeline stages 4 (infer), 5/6 (validate), 7 (assemble),
  8 (deps), 9 (validate post-deps), 10 (`manifestHash`). New files in
  `internal/catalogresolve/`: `assemble.go`, `clock.go`, `dependencies.go`,
  `errors.go`, `hash.go`, `infer.go`, `resolve_full.go`, `validate.go`,
  `resolve_full_test.go`, `testdata/resolve_e2e/`, `testdata/resolve_cycle/`.
  Additive edits to `intent.go` (intentInference pointer-mirror) and
  `types.go` (+`ResolvedCatalog`; +`Options.{Strict,Repo,Namespace,Clock}`).
  No edits outside `internal/catalogresolve/`.
- **Coverage floors on main:** `internal/catalogresolve` **90.2%**
  (gate ≥ 90%, +0.2pp headroom over PR-1's exact 90.0%); Phase 2 floors
  held byte-for-byte (catalogmodel 91.1%, sourcectx 91.1%, Sanitize* 100%);
  Phase 1 floors held (statestore 95.7%, revision 90.3%, executionstate 90.0%).
- **Properties proven:**
  1. T-RES-1 — `Resolve` × 2 on a fixture produces byte-identical
     canonical encodings AND identical per-component `manifestHash` values.
  2. T-RES-2 — every `inheritedFrom` / `inferredFrom` pointer references a
     real authored / inferred origin in the fixture.
  3. `manifestHash` is provenance-invariant: flipping
     `resolution.inheritedFrom` does NOT change the hash; spec edits DO.
     Computed via `catalogmodel.CanonicalEncode` over
     `{identity, metadata, spec, runtime}`.
  4. `ErrDependencyMissing` carries both endpoints (`From`, `To`).
  5. `deploy-after` cycles error always; `calls` cycles warn (default) or
     error (strict).
  6. Inference is `recover()`-safe — failures emit warn-severity
     `ErrInferenceFailed` and skip rather than panic.
- **Determinism stress on main:** `go test -count=10 -race
  ./internal/catalogresolve/...` zero failures.
- **Local gates on main:** `go build`, `go vet`, `go test ./... -race`,
  `make test-state-redesign`, `make verify-generated`, `kiox -- orun
  validate / plan --changed / run --dry-run` all green.

## Current Task (none — between cycles)
- C2 closed; C3 awaiting orchestrator scope as Task 0028.

## Next Task After 0027 — Task 0028 (C3 implementer)
- **Milestone:** C3 — `CatalogSnapshot` and graph builder (single PR
  per `implementation-plan.md` §C3).
- **Adds:** `internal/catalogresolve/graph.go` building `dependencies`,
  `systems`, `apis`, `resources`, `owners` graphs;
  `internal/catalogresolve/resolver.go` (or extension of `resolve_full.go`)
  surfacing `ResolvedCatalog` with `CatalogGraph`, `summary.*` counts
  from sorted collections, and `catalogHash` per `identity-and-keys.md`
  §9 (inputs: `catalogInputHash` + sorted `(componentKey, manifestHash)`
  pairs + canonical `CatalogGraph` + `resolver.resolverVersion`).
- **"Done when":** T-IDK-1 (same source + inputs ⇒ same `catalogHash`);
  `metadata.owner` edit changes `catalogHash`; `resolution.inheritedFrom`
  edit does NOT change `manifestHash` (already proven by Task 0026 —
  verifier confirms this still holds); graph files byte-stable across
  runs.
- **Spec sources:** `implementation-plan.md` §C3, `resolution-pipeline.md`
  §1 + §7, `identity-and-keys.md` §9 + §10, `data-model.md` §3 + §6 + §7.

## Repo Checkpoint (historical — superseded by Phase 2 C3 close above)

| Attribute | Value |
|---|---|
| `main` tip after C2 close | `74b88e0` — Task 0026 / C2 PR-2 (PR #171) on 2026-05-31T08:36:04Z |
| Tasks completed at C2 close | 0001–0005, 0007–0016, 0018–0027 (25 total) |

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
