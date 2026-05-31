package catalogstore_test

// reads_test.go covers the C5 PR-2 read seams added in reads.go:
//
//   - ReadCatalogGraph: decodes a persisted graph; absent graph chains
//     statestore.ErrNotFound; an invalid kind is a path-input error.
//   - ReadComponentExecutionIndex: decodes the catalog-local execution
//     index; ABSENCE IS NOT AN ERROR (found=false, zero value).
//   - ReadComponentManifest: decodes one manifest; absence chains
//     statestore.ErrNotFound.
//   - EnumerateComponentManifests: reads every manifest the snapshot's
//     objects.components inventory references, sorted by componentKey; a
//     referenced-but-missing manifest is a surfaced integrity error.
//
// The seams are pure reads over a statestore, so the tests reuse the
// in-package spyStore (writer_test.go) seeded directly via the exported path
// helpers.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/sourceplane/orun/internal/catalogmodel"
	"github.com/sourceplane/orun/internal/catalogstore"
	"github.com/sourceplane/orun/internal/statestore"
)

const (
	readsSrcKey = "src-branch-main-cabcdef-tabcdef0"
	readsCatKey = "cat-deadbeef"
)

// seedObject marshals v and plants it at path in the spy's object map so the
// read seams find it. Uses preExisting so it is treated as already-persisted.
func seedObject(t *testing.T, spy *spyStore, path string, v any) {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal seed %s: %v", path, err)
	}
	spy.preExisting[path] = b
	spy.revisions[path] = spy.nextRev()
}

func TestReadCatalogGraph_DecodesPersisted(t *testing.T) {
	spy := newSpyStore()
	p, err := catalogstore.CatalogGraphPath(readsSrcKey, readsCatKey, "dependencies")
	if err != nil {
		t.Fatal(err)
	}
	want := catalogmodel.CatalogGraph{
		APIVersion: catalogmodel.APIVersionV1Alpha1,
		Kind:       "CatalogGraph",
		Nodes: []catalogmodel.GraphNode{
			{Key: "k/r/a", Kind: "Component", Name: "a"},
			{Key: "k/r/b", Kind: "Component", Name: "b"},
		},
		Edges: []catalogmodel.GraphEdge{
			{From: "k/r/a", To: "k/r/b", Type: "calls"},
		},
	}
	seedObject(t, spy, p, want)

	got, err := catalogstore.ReadCatalogGraph(context.Background(), spy, readsSrcKey, readsCatKey, "dependencies")
	if err != nil {
		t.Fatalf("ReadCatalogGraph: %v", err)
	}
	if len(got.Nodes) != 2 || len(got.Edges) != 1 {
		t.Fatalf("decoded graph mismatch: %+v", got)
	}
	if got.Edges[0].From != "k/r/a" || got.Edges[0].To != "k/r/b" {
		t.Errorf("edge mismatch: %+v", got.Edges[0])
	}
}

func TestReadCatalogGraph_AbsentChainsNotFound(t *testing.T) {
	spy := newSpyStore()
	_, err := catalogstore.ReadCatalogGraph(context.Background(), spy, readsSrcKey, readsCatKey, "dependencies")
	if err == nil {
		t.Fatal("expected error for absent graph")
	}
	if !errors.Is(err, statestore.ErrNotFound) {
		t.Errorf("absent graph must chain statestore.ErrNotFound, got %v", err)
	}
}

func TestReadCatalogGraph_InvalidKind(t *testing.T) {
	spy := newSpyStore()
	_, err := catalogstore.ReadCatalogGraph(context.Background(), spy, readsSrcKey, readsCatKey, "bogus")
	if err == nil {
		t.Fatal("expected error for invalid graph kind")
	}
	// Must NOT be a not-found — it's an input-validation error.
	if errors.Is(err, statestore.ErrNotFound) {
		t.Errorf("invalid kind should be a path-input error, not not-found: %v", err)
	}
}

func TestReadComponentExecutionIndex_AbsentIsNotError(t *testing.T) {
	spy := newSpyStore()
	idx, found, err := catalogstore.ReadComponentExecutionIndex(context.Background(), spy, readsSrcKey, readsCatKey, "svc-a")
	if err != nil {
		t.Fatalf("absent index must not error, got %v", err)
	}
	if found {
		t.Error("found should be false for an absent index")
	}
	if len(idx.Executions) != 0 {
		t.Errorf("absent index should be zero value, got %+v", idx)
	}
}

func TestReadComponentExecutionIndex_DecodesPersisted(t *testing.T) {
	spy := newSpyStore()
	p, err := catalogstore.ComponentLocalIndexPath(readsSrcKey, readsCatKey, "svc-a")
	if err != nil {
		t.Fatal(err)
	}
	want := catalogmodel.ComponentExecutionIndex{
		APIVersion:         catalogmodel.APIVersionV1Alpha1,
		Kind:               catalogmodel.KindComponentExecIndex,
		ComponentKey:       "k/r/svc-a",
		SourceSnapshotKey:  readsSrcKey,
		CatalogSnapshotKey: readsCatKey,
		Executions: []catalogmodel.ComponentExecutionRow{
			{ExecutionKey: "exec-1", Status: "succeeded", CreatedAt: "2026-01-01T00:00:00Z"},
		},
	}
	seedObject(t, spy, p, want)

	idx, found, err := catalogstore.ReadComponentExecutionIndex(context.Background(), spy, readsSrcKey, readsCatKey, "svc-a")
	if err != nil {
		t.Fatalf("ReadComponentExecutionIndex: %v", err)
	}
	if !found {
		t.Fatal("found should be true for a persisted index")
	}
	if len(idx.Executions) != 1 || idx.Executions[0].ExecutionKey != "exec-1" {
		t.Errorf("decoded index mismatch: %+v", idx)
	}
}

func TestReadComponentManifest_DecodesAndAbsent(t *testing.T) {
	spy := newSpyStore()
	p, err := catalogstore.ComponentManifestPath(readsSrcKey, readsCatKey, "svc-a")
	if err != nil {
		t.Fatal(err)
	}
	want := makeReadManifest("k/r/svc-a", "svc-a")
	seedObject(t, spy, p, want)

	got, err := catalogstore.ReadComponentManifest(context.Background(), spy, readsSrcKey, readsCatKey, "svc-a")
	if err != nil {
		t.Fatalf("ReadComponentManifest: %v", err)
	}
	if got.Identity.ComponentKey != "k/r/svc-a" {
		t.Errorf("componentKey = %q", got.Identity.ComponentKey)
	}

	// Absent manifest chains not-found.
	_, err = catalogstore.ReadComponentManifest(context.Background(), spy, readsSrcKey, readsCatKey, "ghost")
	if err == nil || !errors.Is(err, statestore.ErrNotFound) {
		t.Errorf("absent manifest must chain ErrNotFound, got %v", err)
	}
}

func TestEnumerateComponentManifests_SortedByKey(t *testing.T) {
	spy := newSpyStore()
	// Seed three manifests out of key order.
	for _, kv := range []struct{ key, name string }{
		{"k/r/charlie", "charlie"},
		{"k/r/alpha", "alpha"},
		{"k/r/bravo", "bravo"},
	} {
		p, err := catalogstore.ComponentManifestPath(readsSrcKey, readsCatKey, kv.name)
		if err != nil {
			t.Fatal(err)
		}
		seedObject(t, spy, p, makeReadManifest(kv.key, kv.name))
	}

	cat := catalogmodel.CatalogSnapshot{
		SourceSnapshotKey:  readsSrcKey,
		CatalogSnapshotKey: readsCatKey,
		Objects: catalogmodel.CatalogObjects{
			Components: []catalogmodel.ManifestRef{
				{Name: "charlie"}, {Name: "alpha"}, {Name: "bravo"},
			},
		},
	}

	got, err := catalogstore.EnumerateComponentManifests(context.Background(), spy, cat)
	if err != nil {
		t.Fatalf("EnumerateComponentManifests: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 manifests, got %d", len(got))
	}
	wantOrder := []string{"k/r/alpha", "k/r/bravo", "k/r/charlie"}
	for i, w := range wantOrder {
		if got[i].Identity.ComponentKey != w {
			t.Errorf("order[%d] = %q, want %q", i, got[i].Identity.ComponentKey, w)
		}
	}
}

func TestEnumerateComponentManifests_MissingIsIntegrityError(t *testing.T) {
	spy := newSpyStore()
	cat := catalogmodel.CatalogSnapshot{
		SourceSnapshotKey:  readsSrcKey,
		CatalogSnapshotKey: readsCatKey,
		Objects: catalogmodel.CatalogObjects{
			Components: []catalogmodel.ManifestRef{{Name: "ghost"}},
		},
	}
	_, err := catalogstore.EnumerateComponentManifests(context.Background(), spy, cat)
	if err == nil {
		t.Fatal("a referenced-but-missing manifest must be a surfaced error")
	}
}

// TestReadSeams_NonNotFoundReadErrorSurfaces drives the spy's one-shot read
// error injection so each reader's "read failed for a reason other than
// not-found" branch is exercised (these are distinct from the absent-object
// paths and must propagate, not be swallowed).
func TestReadSeams_NonNotFoundReadErrorSurfaces(t *testing.T) {
	boom := errors.New("disk on fire")

	t.Run("graph", func(t *testing.T) {
		spy := newSpyStore()
		p, _ := catalogstore.CatalogGraphPath(readsSrcKey, readsCatKey, "dependencies")
		spy.readErr[p] = boom
		_, err := catalogstore.ReadCatalogGraph(context.Background(), spy, readsSrcKey, readsCatKey, "dependencies")
		if err == nil || errors.Is(err, statestore.ErrNotFound) {
			t.Fatalf("expected a non-not-found read error to surface, got %v", err)
		}
	})

	t.Run("execIndex", func(t *testing.T) {
		spy := newSpyStore()
		p, _ := catalogstore.ComponentLocalIndexPath(readsSrcKey, readsCatKey, "svc-a")
		spy.readErr[p] = boom
		_, found, err := catalogstore.ReadComponentExecutionIndex(context.Background(), spy, readsSrcKey, readsCatKey, "svc-a")
		if err == nil {
			t.Fatal("expected a non-not-found read error to surface")
		}
		if found {
			t.Error("found must be false on a read error")
		}
	})

	t.Run("manifest", func(t *testing.T) {
		spy := newSpyStore()
		p, _ := catalogstore.ComponentManifestPath(readsSrcKey, readsCatKey, "svc-a")
		spy.readErr[p] = boom
		_, err := catalogstore.ReadComponentManifest(context.Background(), spy, readsSrcKey, readsCatKey, "svc-a")
		if err == nil || errors.Is(err, statestore.ErrNotFound) {
			t.Fatalf("expected a non-not-found read error to surface, got %v", err)
		}
	})
}

// TestReadSeams_DecodeErrorSurfaces plants a non-JSON body so each reader's
// json.Unmarshal failure branch is exercised.
func TestReadSeams_DecodeErrorSurfaces(t *testing.T) {
	plantGarbage := func(spy *spyStore, p string) {
		spy.preExisting[p] = []byte("{not json")
		spy.revisions[p] = spy.nextRev()
	}

	t.Run("graph", func(t *testing.T) {
		spy := newSpyStore()
		p, _ := catalogstore.CatalogGraphPath(readsSrcKey, readsCatKey, "dependencies")
		plantGarbage(spy, p)
		if _, err := catalogstore.ReadCatalogGraph(context.Background(), spy, readsSrcKey, readsCatKey, "dependencies"); err == nil {
			t.Fatal("expected decode error")
		}
	})

	t.Run("execIndex", func(t *testing.T) {
		spy := newSpyStore()
		p, _ := catalogstore.ComponentLocalIndexPath(readsSrcKey, readsCatKey, "svc-a")
		plantGarbage(spy, p)
		if _, _, err := catalogstore.ReadComponentExecutionIndex(context.Background(), spy, readsSrcKey, readsCatKey, "svc-a"); err == nil {
			t.Fatal("expected decode error")
		}
	})

	t.Run("manifest", func(t *testing.T) {
		spy := newSpyStore()
		p, _ := catalogstore.ComponentManifestPath(readsSrcKey, readsCatKey, "svc-a")
		plantGarbage(spy, p)
		if _, err := catalogstore.ReadComponentManifest(context.Background(), spy, readsSrcKey, readsCatKey, "svc-a"); err == nil {
			t.Fatal("expected decode error")
		}
	})
}

// TestReadComponentExecutionIndex_InvalidName exercises the path-input error
// branch (an empty / illegal component name fails ComponentLocalIndexPath
// before any Read).
func TestReadComponentExecutionIndex_InvalidName(t *testing.T) {
	spy := newSpyStore()
	if _, _, err := catalogstore.ReadComponentExecutionIndex(context.Background(), spy, readsSrcKey, readsCatKey, ""); err == nil {
		t.Fatal("expected a path-input error for an empty component name")
	}
}

// TestReadComponentManifest_InvalidName exercises the manifest path-input
// error branch.
func TestReadComponentManifest_InvalidName(t *testing.T) {
	spy := newSpyStore()
	if _, err := catalogstore.ReadComponentManifest(context.Background(), spy, readsSrcKey, readsCatKey, ""); err == nil {
		t.Fatal("expected a path-input error for an empty component name")
	}
}

// TestErrInvalidSelector_ErrorString covers the ErrInvalidSelector.Error()
// formatter (the selector parser returns *ErrInvalidSelector; this exercises
// its message rendering directly).
func TestErrInvalidSelector_ErrorString(t *testing.T) {
	_, err := catalogstore.ParseRefSelector("branches/", "")
	if err == nil {
		t.Fatal("expected an invalid-selector error")
	}
	var ise *catalogstore.ErrInvalidSelector
	if !errors.As(err, &ise) {
		t.Fatalf("expected *ErrInvalidSelector, got %T", err)
	}
	msg := ise.Error()
	if msg == "" || !strings.Contains(msg, "invalid catalog selector") {
		t.Errorf("unexpected error string: %q", msg)
	}
}

// makeReadManifest builds a minimal valid ComponentManifest for seeding the
// read-seam tests with an explicit componentKey (the in-package makeManifest
// derives the key from a fixed namespace/repo, which these enumeration-order
// tests need to control directly).
func makeReadManifest(key, name string) catalogmodel.ComponentManifest {
	return catalogmodel.ComponentManifest{
		APIVersion: catalogmodel.APIVersionV1Alpha1,
		Kind:       "ComponentManifest",
		Identity: catalogmodel.ComponentIdentity{
			ComponentKey: key,
			Name:         name,
		},
		Source: catalogmodel.ComponentSource{
			SourceSnapshotKey:  readsSrcKey,
			CatalogSnapshotKey: readsCatKey,
		},
		Spec: catalogmodel.ComponentSpec{Type: "service"},
	}
}
