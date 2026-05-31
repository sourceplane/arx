package catalogstore

// reads.go adds the C5 PR-2 read accessors the Resolver interface does
// not already expose: the per-kind catalog graph body and the
// catalog-local ComponentExecutionIndex. Both are package-level pure
// reads over statestore (mirroring the ListRefs seam) so the Resolver
// interface stays byte-for-byte as PR-1/PR-3 shipped it. No raw
// os / io / filepath imports (the no-raw-FS lint must stay green).
//
// The read CLI (`orun catalog tree` / `history`) consumes these so
// graph-walking and history enumeration live behind a tested seam
// rather than buried in a cobra RunE.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/sourceplane/orun/internal/catalogmodel"
	"github.com/sourceplane/orun/internal/statestore"
)

// ReadCatalogGraph reads a single relationship graph
// (sources/<srcKey>/catalogs/<catKey>/graph/<kind>.json) and decodes it.
// `kind` is one of the five recognized graph kinds (dependencies,
// systems, apis, resources, owners). An absent graph is surfaced as a
// chained statestore.ErrNotFound so callers can errors.Is against it.
func ReadCatalogGraph(ctx context.Context, state statestore.StateStore, srcKey, catKey, kind string) (catalogmodel.CatalogGraph, error) {
	p, err := CatalogGraphPath(srcKey, catKey, kind)
	if err != nil {
		return catalogmodel.CatalogGraph{}, fmt.Errorf("ReadCatalogGraph: %w", err)
	}
	body, _, rerr := state.Read(ctx, p)
	if rerr != nil {
		if errors.Is(rerr, statestore.ErrNotFound) {
			return catalogmodel.CatalogGraph{}, errNotFoundChain(fmt.Errorf("ReadCatalogGraph: graph %q absent", kind))
		}
		return catalogmodel.CatalogGraph{}, fmt.Errorf("ReadCatalogGraph: Read %s: %w", p, rerr)
	}
	var g catalogmodel.CatalogGraph
	if err := json.Unmarshal(body, &g); err != nil {
		return catalogmodel.CatalogGraph{}, fmt.Errorf("ReadCatalogGraph: decode %s: %w", p, err)
	}
	return g, nil
}

// ReadComponentExecutionIndex reads the catalog-local execution-history
// index for a component
// (sources/<srcKey>/catalogs/<catKey>/indexes/components/<name>.json)
// and decodes it. This is the read source for `orun catalog history`.
//
// A missing index is NOT an error: a component that has never executed
// has no execution-index file, and history must enumerate an empty list
// rather than fail. Absence therefore returns a zeroed
// ComponentExecutionIndex with found=false.
func ReadComponentExecutionIndex(ctx context.Context, state statestore.StateStore, srcKey, catKey, name string) (catalogmodel.ComponentExecutionIndex, bool, error) {
	p, err := ComponentLocalIndexPath(srcKey, catKey, name)
	if err != nil {
		return catalogmodel.ComponentExecutionIndex{}, false, fmt.Errorf("ReadComponentExecutionIndex: %w", err)
	}
	body, _, rerr := state.Read(ctx, p)
	if rerr != nil {
		if errors.Is(rerr, statestore.ErrNotFound) {
			return catalogmodel.ComponentExecutionIndex{}, false, nil
		}
		return catalogmodel.ComponentExecutionIndex{}, false, fmt.Errorf("ReadComponentExecutionIndex: Read %s: %w", p, rerr)
	}
	var idx catalogmodel.ComponentExecutionIndex
	if err := json.Unmarshal(body, &idx); err != nil {
		return catalogmodel.ComponentExecutionIndex{}, false, fmt.Errorf("ReadComponentExecutionIndex: decode %s: %w", p, err)
	}
	return idx, true, nil
}

// ReadComponentManifest reads a single ComponentManifest under a catalog
// (sources/<srcKey>/catalogs/<catKey>/components/<name>/manifest.json) and
// decodes it. Unlike Resolver.ResolveComponent it takes the already-resolved
// (srcKey, catKey) pair, so enumerations that have resolved the catalog once
// can read each manifest without re-walking the ref layer per component.
//
// An absent manifest is surfaced as a chained statestore.ErrNotFound so
// callers can errors.Is against it.
func ReadComponentManifest(ctx context.Context, state statestore.StateStore, srcKey, catKey, name string) (catalogmodel.ComponentManifest, error) {
	p, err := ComponentManifestPath(srcKey, catKey, name)
	if err != nil {
		return catalogmodel.ComponentManifest{}, fmt.Errorf("ReadComponentManifest: %w", err)
	}
	body, _, rerr := state.Read(ctx, p)
	if rerr != nil {
		if errors.Is(rerr, statestore.ErrNotFound) {
			return catalogmodel.ComponentManifest{}, errNotFoundChain(fmt.Errorf("ReadComponentManifest: manifest %q absent", name))
		}
		return catalogmodel.ComponentManifest{}, fmt.Errorf("ReadComponentManifest: Read %s: %w", p, rerr)
	}
	var m catalogmodel.ComponentManifest
	if err := json.Unmarshal(body, &m); err != nil {
		return catalogmodel.ComponentManifest{}, fmt.Errorf("ReadComponentManifest: decode %s: %w", p, err)
	}
	return m, nil
}

// EnumerateComponentManifests reads every ComponentManifest referenced by a
// resolved CatalogSnapshot's objects.components inventory and returns them
// sorted by Identity.ComponentKey for byte-stable output. This is the read
// source for `orun catalog list` / `tree`: the catalog is resolved once by
// the caller, then each manifest is read directly by (srcKey, catKey, name).
//
// A manifest referenced by the inventory but missing on disk is a real
// integrity error (the catalog is immutable after write) and is surfaced.
func EnumerateComponentManifests(ctx context.Context, state statestore.StateStore, cat catalogmodel.CatalogSnapshot) ([]catalogmodel.ComponentManifest, error) {
	out := make([]catalogmodel.ComponentManifest, 0, len(cat.Objects.Components))
	for _, ref := range cat.Objects.Components {
		m, err := ReadComponentManifest(ctx, state, cat.SourceSnapshotKey, cat.CatalogSnapshotKey, ref.Name)
		if err != nil {
			return nil, fmt.Errorf("EnumerateComponentManifests: %w", err)
		}
		out = append(out, m)
	}
	sort.SliceStable(out, func(a, b int) bool {
		return out[a].Identity.ComponentKey < out[b].Identity.ComponentKey
	})
	return out, nil
}
