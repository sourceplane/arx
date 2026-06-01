package catalogsync_test

import (
	"context"
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sourceplane/orun/internal/catalogmodel"
	"github.com/sourceplane/orun/internal/catalogsync"
)

// sync_test.go covers the C9 sync seam: NoopSyncer's documented behavior, the
// catalogmodel-only composition of SyncPayload, and the import-boundary
// invariant that the package never reaches for networking, the runner, the
// store, or the CLI.

func TestNoopSyncer_NotConfigured(t *testing.T) {
	var s catalogsync.Syncer = catalogsync.NoopSyncer{}
	res, err := s.PushCatalogSnapshot(context.Background(), catalogsync.SyncPayload{}, catalogsync.PushOptions{})
	if err != nil {
		t.Fatalf("NoopSyncer must not return an error, got %v", err)
	}
	if res.Accepted {
		t.Error("NoopSyncer must report Accepted=false")
	}
	if res.RemoteSourceKey != "" || res.RemoteCatalogKey != "" {
		t.Errorf("NoopSyncer must not assign remote keys, got %+v", res)
	}
	if len(res.Warnings) != 1 || res.Warnings[0] != catalogsync.NoopWarning {
		t.Errorf("warnings = %v, want exactly [%q]", res.Warnings, catalogsync.NoopWarning)
	}
	if !strings.Contains(catalogsync.NoopWarning, "remote sync not configured") {
		t.Errorf("NoopWarning = %q, must contain the documented phrase", catalogsync.NoopWarning)
	}
}

// TestNoopSyncer_IgnoresOptionsAndPayload proves the noop result is invariant
// to the payload and options (no panic, same result).
func TestNoopSyncer_IgnoresOptionsAndPayload(t *testing.T) {
	payload := catalogsync.SyncPayload{
		Source:  catalogmodel.SourceSnapshot{SourceSnapshotKey: "src-x"},
		Catalog: catalogmodel.CatalogSnapshot{CatalogSnapshotKey: "cat-x"},
		Manifests: []catalogmodel.ComponentManifest{
			{Identity: catalogmodel.ComponentIdentity{Name: "svc-a"}},
		},
	}
	opts := catalogsync.PushOptions{
		DryRun:        true,
		AllowDirty:    true,
		Reason:        "manual push",
		ExtraMetadata: map[string]string{"k": "v"},
	}
	res, err := catalogsync.NoopSyncer{}.PushCatalogSnapshot(context.Background(), payload, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Accepted || len(res.Warnings) != 1 {
		t.Errorf("noop result changed with payload/opts: %+v", res)
	}
}

// TestSyncPayload_ComposedFromCatalogModel constructs a fully-populated
// SyncPayload using only catalogmodel types and round-trips it through JSON.
// This documents (and enforces at compile time) that the payload requires no
// runner, store, or CLI types to build.
func TestSyncPayload_ComposedFromCatalogModel(t *testing.T) {
	payload := catalogsync.SyncPayload{
		Source:  catalogmodel.SourceSnapshot{SourceSnapshotKey: "src-1"},
		Catalog: catalogmodel.CatalogSnapshot{CatalogSnapshotKey: "cat-1"},
		Manifests: []catalogmodel.ComponentManifest{
			{Identity: catalogmodel.ComponentIdentity{ComponentKey: "ns/repo/a", Name: "a"}},
		},
		Graphs: []catalogmodel.CatalogGraph{
			{Nodes: []catalogmodel.GraphNode{{Key: "ns/repo/a", Name: "a"}}},
		},
		GlobalIndexes: []catalogmodel.ComponentGlobalIndex{
			{ComponentKey: "ns/repo/a", Name: "a"},
		},
		SourceRef:     &catalogmodel.SourceRef{Name: "current", SourceSnapshotKey: "src-1"},
		CatalogRef:    &catalogmodel.CatalogRef{Name: "current", CatalogSnapshotKey: "cat-1"},
		HistoryEvents: []catalogmodel.ComponentHistoryEvent{{EventType: catalogmodel.EventCatalogResolved}},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got catalogsync.SyncPayload
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Source.SourceSnapshotKey != "src-1" || got.Catalog.CatalogSnapshotKey != "cat-1" {
		t.Errorf("round-trip lost keys: %+v", got)
	}
	if len(got.Manifests) != 1 || len(got.Graphs) != 1 || len(got.GlobalIndexes) != 1 {
		t.Errorf("round-trip lost slices: %+v", got)
	}
	if got.SourceRef == nil || got.CatalogRef == nil {
		t.Errorf("round-trip lost refs: %+v", got)
	}
}

// forbiddenImports are packages the sync seam must never pull in: any
// networking, the runner, the store, or the CLI. Keeping the list here (rather
// than only in docs) makes a regression a test failure.
var forbiddenImports = []string{
	"net/http",
	"net",
	"github.com/sourceplane/orun/internal/runner",
	"github.com/sourceplane/orun/internal/catalogstore",
	"github.com/sourceplane/orun/cmd/orun",
}

// TestImportBoundary parses every non-test .go file in this package and asserts
// none import a forbidden package. This is the load-bearing C9 invariant: the
// seam stays pure and future-driver-ready.
func TestImportBoundary(t *testing.T) {
	fset := token.NewFileSet()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, perr := parser.ParseFile(fset, filepath.Join(".", name), nil, parser.ImportsOnly)
		if perr != nil {
			t.Fatalf("parse %s: %v", name, perr)
		}
		for _, imp := range f.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			for _, bad := range forbiddenImports {
				if path == bad {
					t.Errorf("%s imports forbidden package %q", name, bad)
				}
			}
		}
	}
}
