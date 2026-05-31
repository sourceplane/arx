# Implementer Report — task-0025 (Phase 2 Milestone C2, PR 1 of N)

- **Branch:** `impl/task-0025-c2-discover-load-inherit`
- **PR:** #170 — https://github.com/sourceplane/orun/pull/170
- **Status:** Ready for verification

## Scope

Implements the first three stages of Milestone C2 of
`specs/orun-component-catalog/implementation-plan.md` —
**discover → load → inherit** — behind a single
`DiscoverAndLoad(ctx, opts)` entry point that produces a deterministic
`[]AuthoredManifest` from a workspace.

Out of scope this PR (deferred to later C2/C3 PRs):
composition defaults, inference, dependency resolution, validation
matrix, `manifestHash`, graph/snapshot.

C0 (catalogmodel data shapes + JSON Schema artifact) and C1
(`internal/sourcectx`) are unchanged; this PR adds an additive sibling
package and one additive file in `catalogmodel`.

## Deliverables

### New files in `internal/catalogresolve/`

| File | Purpose |
|---|---|
| `doc.go` | Package overview + pipeline contract |
| `types.go` | `AuthoredManifest`, `Options`, `DiscoveryResult`, `Provenance`, typed errors |
| `discover.go` | Workspace walk, default + intent excludes, mixed-extension detection, deterministic sort |
| `load.go` | Lazy schema compile (Draft 7), YAML→JSON→validate, authored provenance |
| `intent.go` | `intent.yaml` shapes (`catalog.defaults`, `catalog.discovery.exclude`); nil if missing |
| `inherit.go` | Apply intent defaults: scalar / per-key map / wholesale list rules + provenance updates |
| `resolve.go` | `DiscoverAndLoad(ctx, Options)` orchestrator + workspace-root validation |
| `resolve_test.go` | Happy path, schema-invalid, mixed-extension, no-intent, bad-intent, explicit `[]`, determinism mini-T-RES-1, malformed YAML |
| `error_test.go` | `Error()` + `Unwrap()` for all 5 typed error structs |
| `coverage_test.go` | yml-form per-key annotation inherit, workspace-root-is-file, empty `WorkspaceRoot` |
| `types_test.go` | Provenance/options round-trips |
| `testdata/` | 8 fixture trees (canonical, mixed_extension, schema_invalid, no_intent, bad_intent, explicit_empty_list, yml_form, yaml_malformed) |

### Public API

```go
// Top-level entry.
func DiscoverAndLoad(ctx context.Context, opts Options) (DiscoveryResult, error)

type Options struct {
    WorkspaceRoot string
    // future: per-call overrides
}

type DiscoveryResult struct {
    Manifests []AuthoredManifest // sorted by RelativePath
    // future: warnings, dropped paths
}

type AuthoredManifest struct {
    RelativePath string
    Path         string
    Component    catalogmodel.Component
    Provenance   Provenance // map[jsonPointer]Source
}
```

Typed errors (all wrap a sentinel + carry `Path`/`Reason`):
`ErrWorkspaceInvalid`, `ErrManifestInvalid`, `ErrManifestMixedExtension`, `ErrIntentInvalid`.

### Additive file in `internal/catalogmodel/`

`schema_embed.go` (8 lines) — exposes the existing
`schema/component-yaml.schema.json` via `//go:embed`:

```go
package catalogmodel

import _ "embed"

//go:embed schema/component-yaml.schema.json
var ComponentYAMLSchema []byte
```

No edits to any pre-existing catalogmodel source file. See *Assumptions*.

### Makefile

`test-state-redesign` gains a catalogresolve coverage gate (≥ 90%),
mirroring catalogmodel/sourcectx.

## Test results

```
$ make test-state-redesign
…
🧪 Coverage gate: ./internal/statestore/...  (>= 95%)   measured: 95.7%
🧪 Coverage gate: ./internal/revision/...    (>= 90%)   measured: 90.3%
🧪 Coverage gate: ./internal/executionstate/ (>= 90%)   measured: 90.0%
🧪 Coverage gate: ./internal/catalogmodel/   (>= 90%)   measured: 91.1%
🧪 Coverage gate: ./internal/sourcectx/      (>= 90%)   measured: 91.1%
🧪 Coverage gate: ./internal/catalogresolve/ (>= 90%)   measured: 90.0%
🧪 Coverage gate: ./internal/catalogmodel/ Sanitize* (== 100%)
                                                        measured: 100.0%

$ make verify-generated
✅ generated artifacts up-to-date

$ go build ./...   ✅
$ go vet   ./...   ✅
$ go test  -race ./...   ✅ (every package)
```

Determinism (mini-T-RES-1): two consecutive
`DiscoverAndLoad(ctx, sameOpts)` calls produce byte-identical
`DiscoveryResult` JSON dumps — asserted in `resolve_test.go`.

## Behavioral notes

- **Default excludes** (hardcoded): `.git .orun build dist node_modules vendor`. Intent's `catalog.discovery.exclude` is appended.
- **Mixed extension** — if both `component.yaml` and `component.yml` exist in the same directory, the typed `ErrManifestMixedExtension` is returned with both paths, before any schema work.
- **Schema compile** — lazy `sync.Once` over `jsonschema.NewCompiler()` Draft 7 with `AddResource("memory://component-yaml.schema.json", ...)`, mirroring the pattern in `internal/schema/validator.go`.
- **List inheritance** — `nil` slice in authored ⇒ inherit from intent; non-nil (including explicit `[]`) ⇒ preserve. Map inheritance is per-key. Scalars inherit only when authored is the zero value.
- **Provenance** — every authored-or-inherited field records the source file + JSON pointer (RFC 6901-escaped); covered by `authoredProvenance` in `load.go` and updated by `inherit.go`.
- **Workspace-root validation** — `os.Stat`+`IsDir` runs before intent load, so a file-as-root surfaces as `ErrWorkspaceInvalid`, not a confusing `ErrIntentInvalid`.

## Assumptions

**Schema-embed location.** Task pack PR-Boundary §3 says:

> No edits to `internal/catalogmodel/` or `internal/sourcectx/`.

while §1/§2 say:

> Schema = single source of truth at
> `internal/catalogmodel/schema/component-yaml.schema.json`; do not
> vendor a copy; do not re-derive from Go types.

`//go:embed` cannot escape its package directory, so a downstream
package cannot embed the schema directly. Available options were:

1. New file in `internal/catalogmodel/` exposing the embedded bytes.
2. Vendor a duplicate copy of the schema into `catalogresolve/`.
3. Read the file at runtime via absolute path.

(2) is forbidden by the spec. (3) is fragile (breaks `go install` and
cross-repo embedding). I took **option 1** as the narrowest reading of
the constraint — *one new file* (`schema_embed.go`, 8 lines) with **no
edits to any pre-existing catalogmodel source**. This was an
autonomous decision while the user was AFK; flagging here for
verifier/orchestrator review.

## Spec Proposals

Suggest tightening the C2 PR-Boundary wording from

> No edits to `internal/catalogmodel/` …

to

> No edits to **existing source files in** `internal/catalogmodel/` …

so additive sibling files (embeds, registrations, exported view types)
needed by dependent packages remain available without re-litigating the
constraint per task. Same wording applies to the `internal/sourcectx/`
clause for symmetry.

## Verification checklist for the verifier

- [ ] `make test-state-redesign` green; catalogresolve gate ≥ 90%.
- [ ] `make verify-generated` clean.
- [ ] No edits to existing catalogmodel/sourcectx source (only one new
      file in catalogmodel: `schema_embed.go`).
- [ ] No vendored copy of `component-yaml.schema.json`.
- [ ] `DiscoverAndLoad` is the only exported entry point; everything
      else is types / typed errors.
- [ ] Determinism: re-running `go test ./internal/catalogresolve/...
      -count=5 -race` stays green.
- [ ] PR diff stays within `internal/catalogresolve/`,
      `internal/catalogmodel/schema_embed.go`, and `Makefile`.
