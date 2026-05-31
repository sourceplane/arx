# Task 0038 — Implementer Report

**Task:** C5 PR-2 — `orun catalog` read surface (`list`/`describe`/`tree`/`history`/`validate` + `diff` stub)
**Branch:** `task-0038-catalog-cli-c5-pr2-read-surface`
**PR:** #<PR_NUMBER>

## Summary

- Shipped the five C5 read commands plus the `diff` stub under `orun catalog`, all
  wired through the PR-1 shared selector parser (`parseCatalogSelector`), the §11
  envelope writer (`writeCatalogEnvelope`), and the C4 `catalogstore.Resolver` —
  no new write paths, no new on-disk contract, no event APPEND.
- Added two pure read seams in `internal/catalogstore/reads.go`
  (`ReadCatalogGraph`, `ReadComponentExecutionIndex`) plus a manifest reader and a
  sorted `EnumerateComponentManifests`, keeping graph-walking and history
  enumeration behind tested helpers rather than in cobra `RunE`. No raw
  `os`/`io/ioutil`/`path/filepath` imports were added to `internal/catalogstore/`.
- `list`/`describe` derive `type`/`owner`/`system`/`domain` directly from the
  resolved `ComponentManifest`s (the recommended manifest-walk data source); the
  empty `CatalogLocalIndexes` non-component axes were NOT populated (out of scope,
  per Integration Notes). `lastExecution*`/`STATUS` come from the catalog-local
  `ComponentExecutionIndex`.
- `describe` exits 4 with the candidate list on an ambiguous bare name and 6 on a
  missing component; `validate` exits 1 on error / 0-with-warnings / 1-on-warning
  under `--strict`; the `diff` stub exits 5 (documented below).
- Help-text fixtures for all nine catalog subcommands committed under
  `cmd/orun/testdata/help/catalog/*.txt`, gated by a golden fixture test with an
  `-update` flag. E2E test refreshes a seeded git workspace then exercises every
  read command in both text and `--json`.

## Files Changed

**CLI (`cmd/orun/`)**
- `catalog.go` — additive only: six new `kindCatalog*Result` envelope-kind
  constants, six `registerCatalog*Command` calls, root `Long` subcommand index
  extended, and a shared `catalogReadExit` helper (not-found → exit 6, other read
  failure → exit 3).
- `catalog_list.go` — `list`: manifest enumeration, §3 columns/filters, sorted by
  componentKey, `--json` (`CatalogListResult`).
- `catalog_describe.go` — `describe`: §4 sections, ambiguity exit 4, missing exit
  6, `--json` `{manifest, executions}` (`CatalogDescribeResult`).
- `catalog_tree.go` — `tree`: dependency-graph render, `--direction in|out|both`,
  `--json` `{nodes, edges}` (`CatalogTreeResult`). Uses bare graph-kind
  `"dependencies"` (the store appends `.json`).
- `catalog_history.go` — `history`: `ComponentExecutionIndex` enumeration,
  newest-first, `--limit`/`--trigger`/`--profile`/`--environment`, read-only,
  `--json` (`CatalogHistoryResult`).
- `catalog_validate.go` — `validate`: strict-aware re-resolve, `[]ValidationIssue`,
  exit-code contract, `--json` (`CatalogValidateResult`).
- `catalog_diff.go` — registered stub, prints not-implemented, exits 5.
- `catalog_help_test.go` — golden help-fixture harness (`-update`).
- `catalog_read_test.go` — E2E read suite (refresh → list/describe/tree/history/
  validate, text + json; ambiguity/missing/bad-direction/diff exit codes).
- `testdata/help/catalog/*.txt` — nine help fixtures.

**Engine (`internal/catalogstore/`)**
- `reads.go` — `ReadCatalogGraph`, `ReadComponentExecutionIndex` (absence is
  `found=false`, not an error), `ReadComponentManifest`, `EnumerateComponentManifests`.
- `reads_test.go` — seam unit tests (decode, absent→not-found, non-not-found read
  error, invalid name/kind, sorted enumeration, integrity error, selector error
  string).

## Checks Run

```
go mod tidy                                              → clean
go build ./...                                           → ok
go vet ./...                                             → ok
go test ./... -race -count=1                             → ok (no failures)
make verify-generated                                    → ✅ no schema drift
make test-state-redesign                                 → all gates pass:
    statestore       95.7%  (floor 95%)
    revision         90.3%  (floor 90%)
    executionstate   90.0%  (floor 90%)
    catalogmodel     91.1%  (floor 90%)
    sourcectx        91.1%  (floor 90%)
    catalogresolve   90.9%  (floor 90%)
    catalogmodel Sanitize*  100.0% (floor 100%)
grep -RnE '^\s*"(os|io/ioutil|path/filepath)"' internal/catalogstore/   → empty
```

Per-package coverage (`-cover`):
- `internal/catalogstore`   **90.7%** (floor 90%, was 89.9% before the added seam
  tests — raised by error/decode/invalid-name branch tests)
- `internal/catalogresolve` **90.9%** (floor 90%)
- `internal/sourcectx`      **91.1%** (floor 90%)
- `cmd/orun`                27.3% (no floor; E2E + fixture tests added)

Manual smoke is covered by the E2E test `TestCatalog{List,Describe,Tree,History,
Validate,Diff}_E2E` which refreshes a seeded git workspace and asserts each
command in text and `--json`.

## Assumptions (durable)

- **`list`/`describe` data source = manifest walk.** `type`/`owner`/`system`/
  `domain` are derived from each resolved `ComponentManifest`, NOT from the
  `CatalogLocalIndexes` owner/system/domain/type axes (those remain empty as PR-1
  left them; data-model §9 under-specifies them). This needs no new index axis and
  matches the task's recommended approach. An O(1) `--owner`-style filter index is
  therefore not present; filters are applied by post-filtering the walked rows.
- **`diff` stub exit code = 5.** The §6 "resolver failure" code is reused as the
  not-ready signal so scripts branching on exit 0 do not mistake the stub for a
  successful empty diff. When C8 lands, 5 becomes "resolver failure" and 0 becomes
  success.
- **Dependency graph kind is the bare name `"dependencies"`** — `CatalogGraphPath`
  appends the `.json` suffix; `catalogmodel.GraphFileDependencies` is the on-disk
  filename (with extension) and must NOT be passed to the read seam.

## Spec Proposals

None filed. The empty `CatalogLocalIndexes` non-component axes are a known,
spec-acknowledged under-specification; the manifest-walk keeps CLI behavior
correct without blocking on it, so no `ai/proposals/task-0038-spec-update.md` was
needed. If a later milestone wants O(1) owner/system filtering, that axis shape is
the natural follow-up.

## Remaining Gaps

- `tree` renders only the dependency graph; the other four graph kinds
  (systems/apis/resources/owners) are persisted but not yet surfaced by a
  `--graph` selector. Not required by C5; candidate for a follow-up.
- `history` enumerates an always-empty index in C5 (executions are appended in
  C7) — the enumeration/filtering/`--limit` paths are tested against the empty
  and (unit-level) populated index, but end-to-end populated history awaits C7.
- `validate --rebuild-indexes` and real `diff` remain C8 (intentionally out of
  scope).

## Next Task Dependencies

On merge, **C5 CLOSES** and **C6 (`orun plan` integration)** unlocks. C6/C7 will
populate the execution index that `history`/`list STATUS` read; C8 will implement
`diff` for real and wire `validate --rebuild-indexes`.

## PR Number

#<PR_NUMBER>
