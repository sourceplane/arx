# Task 0028 — Implementer Report

## Summary
- Landed Milestone **C3** (snapshot + graph assembly) for the Orun
  Component Catalog as **PR #172** on branch
  `task-0028-catalog-c3-snapshot-graph`. Three new files in
  `internal/catalogresolve/` (graph.go, catalog_hash.go,
  catalog_snapshot.go) implement resolution-pipeline.md §1 stages
  11–13 on top of the existing `Resolve` (stages 1–10). Existing
  `Resolve` callers are untouched.
- Exported new entrypoint `BuildCatalog(ctx, opts, ResolverInputs)
  (*CatalogView, []ValidationIssue, error)` returning
  `CatalogView{*ResolvedCatalog; Snapshot *CatalogSnapshot;
  Graphs []*CatalogGraph}`. `ResolverInputs` is fully caller-supplied
  — the resolver never invents `authoritative`, `preview`,
  `sourceSnapshotKey`, `catalogInputHash`, `headRevision`, `treeHash`,
  or `workingTree`. Missing fields ⇒ typed
  `ErrResolverInputsIncomplete`.
- Tests: T-IDK-1 property test with `pgregory.net/rapid` (1000
  random orderings ⇒ identical `catalogHash`); owner-edit acceptance
  signal; provenance-only edits stable for both `manifestHash` and
  `catalogHash`; `BuildCatalog` byte-stable across consecutive
  calls; `summary.*` counts derived from sorted-distinct enumeration;
  graph node/edge ordering verified per data-model.md §4.
- Coverage: `internal/catalogresolve` rose from **90.2 % → 90.9 %**
  (+0.7 pp, floor held). All Phase 1 + Phase 2 floors held byte-for-
  byte. `make verify-generated` clean.
- No filesystem writes, no `internal/catalogstore`, no CLI surface,
  no edits to Phase 1 packages. Additive only — no renames or
  reshapes in `catalogresolve` or `catalogmodel`.

## Files Changed
**Resolver (new files only — additive per task constraint #3):**
- `internal/catalogresolve/graph.go` — `buildGraphs()` for the five
  `CatalogGraph` siblings with deterministic node/edge ordering.
- `internal/catalogresolve/catalog_hash.go` — `catalogHash` per
  identity-and-keys.md §9 input ordering.
- `internal/catalogresolve/catalog_snapshot.go` — `BuildCatalog`
  entrypoint, `ResolverInputs`, `CatalogView`,
  `ErrResolverInputsIncomplete`, `assembleSnapshot`,
  `computeSummary`, `validateResolverInputs`.

**Tests (new files only):**
- `internal/catalogresolve/graph_test.go` — graph builder shape +
  ordering + edge vocabulary + empty-input.
- `internal/catalogresolve/catalog_hash_test.go` — T-IDK-1 (rapid
  1000 orderings), owner-edit changes hash, provenance-only stable,
  resolverVersion bump sensitivity, catalogInputHash sensitivity.
- `internal/catalogresolve/build_catalog_test.go` — E2E happy path,
  determinism across two consecutive calls, summary counts,
  ErrResolverInputsIncomplete.

No existing files modified inside `internal/catalogresolve/`.
No edits to `internal/catalogmodel/`, Phase 1 packages, or
`internal/catalogsync`.

## Checks Run
- `go build ./...` ✅ (exit 0)
- `go vet ./...` ✅ (exit 0)
- `go test ./... -race -count=1` ✅ (all packages green)
- `go test -coverprofile=cover.out ./internal/catalogresolve/...
  && go tool cover -func=cover.out | tail -1` → catalogresolve =
  **90.9 %** (≥ 90.2 floor).
- `make verify-generated` ✅ (`generated artifacts up-to-date`)
- Per-package coverage (cached):
  - `internal/catalogresolve` 90.9 %
  - `internal/catalogmodel`  91.1 %
  - `internal/sourcectx`     91.1 %
  - `internal/statestore`    95.7 %
  - `internal/revision`      90.3 %
  - `internal/executionstate` 90.0 %

## Coverage Numbers
| package | before (main) | after | delta vs floor |
|---|---|---|---|
| `internal/catalogresolve` | 90.2 % | **90.9 %** | +0.7 pp |
| `internal/catalogmodel`   | 91.1 % | 91.1 %     | 0.0 (floor held) |
| `internal/sourcectx`      | 91.1 % | 91.1 %     | 0.0 (floor held) |
| `internal/statestore`     | 95.7 % | 95.7 %     | 0.0 (floor held) |
| `internal/revision`       | 90.3 % | 90.3 %     | 0.0 (floor held) |
| `internal/executionstate` | 90.0 % | 90.0 %     | 0.0 (floor held) |

All Phase 1 and Phase 2 floors held byte-for-byte; no probabilistic
gaps in new code paths (rapid property test backstopped by
deterministic sub-tests for owner edit, provenance edit, resolver-
version bump, and catalogInputHash sensitivity).

## Assumptions (durable)
- **`computeSummary.systems`** is computed from
  `spec.system` (the resolved manifest field) — `metadata.system`
  does not exist on `ComponentManifest` per
  `internal/catalogmodel/component_manifest.go:65`. The task brief
  said "metadata.system" but the only field carrying system
  membership is `spec.system`; mirroring the actual field per
  Integration Notes guidance.
- **`summary.domains`** maps to `spec.domain` (same reasoning;
  `metadata.domain` does not exist — the field is `spec.domain` at
  `component_manifest.go:66`).
- **`objects.components[i].path`** is set to
  `"components/<name>/manifest.json"` — this matches the writer's
  expected on-disk layout from data-model.md §2 even though the
  resolver itself does not write to disk. The writer (C4) will
  validate that this path matches its actual on-disk location; if
  the layout convention changes the resolver path string moves with
  it.
- **`catalogSnapshotKey`** uses width 8 per identity-and-keys.md §3
  ("start at 8"). Collision policy (`-x<n>` suffix or width
  expansion) is the C4 writer's job; the resolver only emits the
  canonical 8-hex form.
- **Source key back-fill**: `BuildCatalog` post-fills
  `ComponentManifest.Source.{SourceSnapshotKey, CatalogSnapshotKey,
  HeadRevision, TreeHash, WorkingTree}` after `catalogSnapshotKey`
  is derived. `manifestHash` was already computed at C2 stage 10
  before any Source fields were stamped (`hash.go` excludes Source
  from the hashed payload), so this back-fill cannot perturb
  `manifestHash`. T-IDK-2 (provenance-only stability) re-asserts
  this at the C3 layer.

## Spec Proposals
None. C3 is implementable as written from the existing specs.

## Remaining Gaps
- The collision-policy expansion path (width 10 → 12 → 14 → 16 →
  `-x<n>` suffix) is only exercised at the writer layer (C4). The
  resolver's `FormatCatalogSnapshotKey(hash, 8)` always emits the
  un-suffixed 8-hex form; if C4 expands width, the writer will
  re-derive the key independently from the same `catalogHash`.
- Stage error wrapping uses `ErrResolverInternal{Stage: 12, ...}`
  for both `catalogHash` failure and `ValidateCatalogSnapshotKey`
  failure (the latter is theoretically unreachable since the input
  is `FormatCatalogSnapshotKey(sha256-hex, 8)`). This is correct
  per the §1 stage numbering but the `ValidateCatalogSnapshotKey`
  branch is uncovered by tests because manufacturing a sha256 hex
  that fails the regex is impossible. Acceptable.

## Next Task Dependencies
**C4 — `internal/catalogstore` Writer + Resolver atomic writes.**
Consumes `*CatalogView` from `BuildCatalog` and writes:
- `sources/<sourceSnapshotKey>/catalogs/<catalogSnapshotKey>/catalog.json`
- `sources/<sourceSnapshotKey>/catalogs/<catalogSnapshotKey>/components/<name>/manifest.json`
- `sources/<sourceSnapshotKey>/catalogs/<catalogSnapshotKey>/graph/{dependencies,systems,apis,resources,owners}.json`

C4 must enforce data-model.md §2 validation rules including
`objects.components[*].manifestHash == manifest.source.manifestHash`
(both come from the same `manifestHash()` compute, so this is a
trip-wire for accidental writer-side rewriting). C4 also owns
collision policy for `catalogSnapshotKey` (the `-x<n>` suffix).

## PR Number
**#172** — https://github.com/sourceplane/orun/pull/172
