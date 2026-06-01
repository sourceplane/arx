package revision

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sourceplane/orun/internal/catalogstore"
	"github.com/sourceplane/orun/internal/statestore"
)

// C6 — catalog-parent mirror. These tests assert the additive Phase 2 layout
// is written under sources/<srcKey>/catalogs/<catKey>/revisions/<revKey>/ and
// is byte-identical to the Phase 1 global layout, while the global layout is
// itself unaffected by the presence/absence of CatalogParent.

const (
	testSrcKey = "src-branch-main-cdef456a-t5ab21c3"
	testCatKey = "cat-c8e91d2a"
)

func writerCfgWithParent(store statestore.StateStore, now time.Time) Config {
	cfg := newWriterCfg(store, now)
	cfg.CatalogParent = CatalogParentRef{SourceKey: testSrcKey, CatalogKey: testCatKey}
	return cfg
}

// TestWriteRevision_CatalogParent_ByteIdentical verifies the three body
// files mirrored under the catalog parent match the Phase 1 globals exactly.
func TestWriteRevision_CatalogParent_ByteIdentical(t *testing.T) {
	store := newTestStore(t)
	trig := newTestTrigger(t)
	cfg := writerCfgWithParent(store, time.Date(2026, 5, 30, 18, 0, 0, 0, time.UTC))
	plan := []byte(`{"apiVersion":"orun.io/v1alpha1","kind":"Plan","jobs":[]}`)
	planHash := "feedface00112233445566778899aabbccddeeff00112233"

	rev, err := WriteRevision(context.Background(), cfg, trig, plan, planHash)
	if err != nil {
		t.Fatalf("WriteRevision: %v", err)
	}
	if err := WriteManifest(context.Background(), cfg, rev, trig); err != nil {
		t.Fatalf("WriteManifest: %v", err)
	}

	type pair struct {
		name   string
		global string
		parent func() (string, error)
	}
	parents := []pair{
		{"plan.json", statestore.PlanPath(rev.RevisionKey), func() (string, error) {
			return catalogstore.CatalogRevisionPlanPath(testSrcKey, testCatKey, rev.RevisionKey)
		}},
		{"trigger.json", statestore.TriggerPath(rev.RevisionKey), func() (string, error) {
			return catalogstore.CatalogRevisionTriggerPath(testSrcKey, testCatKey, rev.RevisionKey)
		}},
		{"revision.json", statestore.RevisionDocPath(rev.RevisionKey), func() (string, error) {
			return catalogstore.CatalogRevisionDocPath(testSrcKey, testCatKey, rev.RevisionKey)
		}},
		{"manifest.json", statestore.ManifestPath(rev.RevisionKey), func() (string, error) {
			return catalogstore.CatalogRevisionManifestPath(testSrcKey, testCatKey, rev.RevisionKey)
		}},
	}
	for _, p := range parents {
		globalRaw, _, err := store.Read(context.Background(), p.global)
		if err != nil {
			t.Fatalf("read global %s: %v", p.name, err)
		}
		parentPath, err := p.parent()
		if err != nil {
			t.Fatalf("parent path %s: %v", p.name, err)
		}
		parentRaw, _, err := store.Read(context.Background(), parentPath)
		if err != nil {
			t.Fatalf("read catalog-parent %s at %q: %v", p.name, parentPath, err)
		}
		if string(globalRaw) != string(parentRaw) {
			t.Errorf("%s diverged between layouts\n global=%s\n parent=%s", p.name, globalRaw, parentRaw)
		}
	}
}

func TestResolveRevision_CatalogParentIndexPathSurvivesGlobalAliasLoss(t *testing.T) {
	store := newTestStore(t)
	trig := newTestTrigger(t)
	cfg := writerCfgWithParent(store, time.Date(2026, 5, 30, 18, 0, 0, 0, time.UTC)).WithCompatibilityWrites(false)
	plan := []byte(`{"apiVersion":"orun.io/v1alpha1","kind":"Plan","jobs":[]}`)
	planHash := "feedface00112233445566778899aabbccddeeff00112233"

	rev, err := WriteRevision(context.Background(), cfg, trig, plan, planHash)
	if err != nil {
		t.Fatalf("WriteRevision: %v", err)
	}
	if err := WriteManifest(context.Background(), cfg, rev, trig); err != nil {
		t.Fatalf("WriteManifest: %v", err)
	}

	idx, _, err := statestore.ReadRevisionIndex(context.Background(), store, rev.RevisionKey)
	if err != nil {
		t.Fatalf("ReadRevisionIndex: %v", err)
	}
	wantDir, err := catalogstore.CatalogRevisionDir(testSrcKey, testCatKey, rev.RevisionKey)
	if err != nil {
		t.Fatalf("CatalogRevisionDir: %v", err)
	}
	if idx.Path != wantDir {
		t.Fatalf("revision index path = %q; want catalog parent %q", idx.Path, wantDir)
	}
	if idx.SourceSnapshotKey != testSrcKey || idx.CatalogSnapshotKey != testCatKey {
		t.Fatalf("revision index parent keys = %q/%q", idx.SourceSnapshotKey, idx.CatalogSnapshotKey)
	}

	rooted, ok := store.(interface{ Root() string })
	if !ok {
		t.Fatal("test store does not expose Root")
	}
	if err := os.RemoveAll(filepath.Join(rooted.Root(), filepath.FromSlash(statestore.RevisionDir(rev.RevisionKey)))); err != nil {
		t.Fatalf("remove global revision dir: %v", err)
	}

	ref, err := ResolveRevision(context.Background(), store, rev.RevisionKey, ResolveOptions{})
	if err != nil {
		t.Fatalf("ResolveRevision via catalog parent: %v", err)
	}
	if ref.Revision.SourceSnapshotKey != testSrcKey || ref.Revision.CatalogSnapshotKey != testCatKey {
		t.Fatalf("resolved parent keys = %q/%q", ref.Revision.SourceSnapshotKey, ref.Revision.CatalogSnapshotKey)
	}
}

func TestResolveRevision_CatalogTreeFallbackWithoutIndex(t *testing.T) {
	store := newTestStore(t)
	trig := newTestTrigger(t)
	cfg := writerCfgWithParent(store, time.Date(2026, 5, 30, 18, 0, 0, 0, time.UTC)).WithCompatibilityWrites(false)
	plan := []byte(`{"apiVersion":"orun.io/v1alpha1","kind":"Plan","jobs":[]}`)
	planHash := "feedface00112233445566778899aabbccddeeff00112233"

	rev, err := WriteRevision(context.Background(), cfg, trig, plan, planHash)
	if err != nil {
		t.Fatalf("WriteRevision: %v", err)
	}
	if err := WriteManifest(context.Background(), cfg, rev, trig); err != nil {
		t.Fatalf("WriteManifest: %v", err)
	}
	deleteGlobalRevisionAlias(t, store, rev.RevisionKey)
	if err := store.Delete(context.Background(), statestore.RevisionIndexPath(rev.RevisionKey)); err != nil {
		t.Fatalf("delete revision index: %v", err)
	}

	ref, err := ResolveRevision(context.Background(), store, rev.RevisionKey, ResolveOptions{})
	if err != nil {
		t.Fatalf("ResolveRevision via catalog tree fallback: %v", err)
	}
	if ref.Revision.RevisionKey != rev.RevisionKey {
		t.Fatalf("resolved revision key = %q; want %q", ref.Revision.RevisionKey, rev.RevisionKey)
	}
	if ref.Revision.SourceSnapshotKey != testSrcKey || ref.Revision.CatalogSnapshotKey != testCatKey {
		t.Fatalf("resolved parent keys = %q/%q", ref.Revision.SourceSnapshotKey, ref.Revision.CatalogSnapshotKey)
	}
}

func TestResolveRevision_CatalogTreeFallbackErrors(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		store := newTestStore(t)
		_, err := resolveFromCatalogTree(context.Background(), store, "rev-main-abcdef0-pfeedface")
		if !errors.Is(err, statestore.ErrNotFound) {
			t.Fatalf("err=%v; want ErrNotFound", err)
		}
	})

	t.Run("conflict", func(t *testing.T) {
		store := newTestStore(t)
		trig := newTestTrigger(t)
		cfg := writerCfgWithParent(store, time.Date(2026, 5, 30, 18, 0, 0, 0, time.UTC)).WithCompatibilityWrites(false)
		plan := []byte(`{"apiVersion":"orun.io/v1alpha1","kind":"Plan","jobs":[]}`)
		planHash := "feedface00112233445566778899aabbccddeeff00112233"

		rev, err := WriteRevision(context.Background(), cfg, trig, plan, planHash)
		if err != nil {
			t.Fatalf("WriteRevision: %v", err)
		}
		if err := WriteManifest(context.Background(), cfg, rev, trig); err != nil {
			t.Fatalf("WriteManifest: %v", err)
		}
		duplicateCatalogRevision(t, store, testCatKey, "cat-deadbee", rev.RevisionKey)
		deleteGlobalRevisionAlias(t, store, rev.RevisionKey)
		if err := store.Delete(context.Background(), statestore.RevisionIndexPath(rev.RevisionKey)); err != nil {
			t.Fatalf("delete revision index: %v", err)
		}

		_, err = ResolveRevision(context.Background(), store, rev.RevisionKey, ResolveOptions{})
		if !errors.Is(err, statestore.ErrConflict) {
			t.Fatalf("err=%v; want ErrConflict", err)
		}
	})
}

func TestWriteRevision_CatalogParentInvalidPath(t *testing.T) {
	store := newTestStore(t)
	trig := newTestTrigger(t)
	cfg := newWriterCfg(store, time.Date(2026, 5, 30, 18, 0, 0, 0, time.UTC)).WithCompatibilityWrites(false)
	cfg.CatalogParent = CatalogParentRef{SourceKey: "bad/source", CatalogKey: testCatKey}
	plan := []byte(`{"apiVersion":"orun.io/v1alpha1","kind":"Plan","jobs":[]}`)
	planHash := "feedface00112233445566778899aabbccddeeff00112233"

	if _, err := WriteRevision(context.Background(), cfg, trig, plan, planHash); err == nil {
		t.Fatal("WriteRevision unexpectedly succeeded with invalid catalog parent")
	}
}

func TestWriteCatalogParentManifestInvalidPath(t *testing.T) {
	store := newTestStore(t)
	err := writeCatalogParentManifest(
		context.Background(),
		store,
		CatalogParentRef{SourceKey: "bad/source", CatalogKey: testCatKey},
		RevisionManifest{},
		"rev-main-abcdef0-pfeedface",
	)
	if err == nil {
		t.Fatal("writeCatalogParentManifest unexpectedly succeeded with invalid catalog parent")
	}
}

// TestWriteRevision_CatalogParent_Inactive verifies that with no
// CatalogParent (the Phase 1 / --no-catalog-refresh case) NO files appear
// under sources/ and the global layout is still written.
func TestWriteRevision_CatalogParent_Inactive(t *testing.T) {
	store := newTestStore(t)
	trig := newTestTrigger(t)
	cfg := newWriterCfg(store, time.Date(2026, 5, 30, 18, 0, 0, 0, time.UTC)) // no CatalogParent
	plan := []byte(`{"apiVersion":"orun.io/v1alpha1","kind":"Plan","jobs":[]}`)
	planHash := "feedface00112233445566778899aabbccddeeff00112233"

	rev, err := WriteRevision(context.Background(), cfg, trig, plan, planHash)
	if err != nil {
		t.Fatalf("WriteRevision: %v", err)
	}
	if err := WriteManifest(context.Background(), cfg, rev, trig); err != nil {
		t.Fatalf("WriteManifest: %v", err)
	}
	// Global layout present.
	if _, _, err := store.Read(context.Background(), statestore.PlanPath(rev.RevisionKey)); err != nil {
		t.Fatalf("global plan.json missing: %v", err)
	}
	// Catalog parent absent.
	parentPlan, err := catalogstore.CatalogRevisionPlanPath(testSrcKey, testCatKey, rev.RevisionKey)
	if err != nil {
		t.Fatalf("parent path: %v", err)
	}
	if _, _, err := store.Read(context.Background(), parentPlan); !errors.Is(err, statestore.ErrNotFound) {
		t.Errorf("catalog-parent plan.json unexpectedly present when CatalogParent inactive: err=%v", err)
	}
}

func deleteGlobalRevisionAlias(t *testing.T, store statestore.StateStore, revKey string) {
	t.Helper()
	rooted, ok := store.(interface{ Root() string })
	if !ok {
		t.Fatal("test store does not expose Root")
	}
	if err := os.RemoveAll(filepath.Join(rooted.Root(), filepath.FromSlash(statestore.RevisionDir(revKey)))); err != nil {
		t.Fatalf("remove global revision dir: %v", err)
	}
}

func duplicateCatalogRevision(t *testing.T, store statestore.StateStore, fromCatKey, toCatKey, revKey string) {
	t.Helper()
	for _, paths := range []struct {
		from func(string, string, string) (string, error)
		to   func(string, string, string) (string, error)
	}{
		{catalogstore.CatalogRevisionPlanPath, catalogstore.CatalogRevisionPlanPath},
		{catalogstore.CatalogRevisionDocPath, catalogstore.CatalogRevisionDocPath},
		{catalogstore.CatalogRevisionTriggerPath, catalogstore.CatalogRevisionTriggerPath},
	} {
		fromPath, err := paths.from(testSrcKey, fromCatKey, revKey)
		if err != nil {
			t.Fatalf("source catalog path: %v", err)
		}
		toPath, err := paths.to(testSrcKey, toCatKey, revKey)
		if err != nil {
			t.Fatalf("dest catalog path: %v", err)
		}
		raw, _, err := store.Read(context.Background(), fromPath)
		if err != nil {
			t.Fatalf("read %s: %v", fromPath, err)
		}
		if _, err := store.Write(context.Background(), toPath, raw, statestore.WriteOptions{}); err != nil {
			t.Fatalf("write %s: %v", toPath, err)
		}
	}
}

// TestWriteRevision_CatalogParent_PartialKeysSkip verifies that a half-filled
// CatalogParent (only one key set) is treated as inactive — no parent writes,
// no error. Guards the Active() invariant.
func TestWriteRevision_CatalogParent_PartialKeysSkip(t *testing.T) {
	cases := []CatalogParentRef{
		{SourceKey: testSrcKey},  // catalog key empty
		{CatalogKey: testCatKey}, // source key empty
		{},                       // both empty
	}
	for i, parent := range cases {
		store := newTestStore(t)
		trig := newTestTrigger(t)
		cfg := newWriterCfg(store, time.Date(2026, 5, 30, 18, 0, 0, 0, time.UTC))
		cfg.CatalogParent = parent
		plan := []byte(`{"apiVersion":"orun.io/v1alpha1","kind":"Plan"}`)
		planHash := "feedface00112233445566778899aabbccddeeff00112233"
		rev, err := WriteRevision(context.Background(), cfg, trig, plan, planHash)
		if err != nil {
			t.Fatalf("case %d WriteRevision: %v", i, err)
		}
		// Use a fully-valid pair only to construct the would-be path; with a
		// partial parent nothing should have been written under sources/.
		parentPlan, perr := catalogstore.CatalogRevisionPlanPath(testSrcKey, testCatKey, rev.RevisionKey)
		if perr != nil {
			t.Fatalf("case %d parent path: %v", i, perr)
		}
		if _, _, err := store.Read(context.Background(), parentPlan); !errors.Is(err, statestore.ErrNotFound) {
			t.Errorf("case %d: parent written for partial CatalogParent %+v: err=%v", i, parent, err)
		}
	}
}

// TestCatalogParentRef_Active is a direct unit check on the gating predicate.
func TestCatalogParentRef_Active(t *testing.T) {
	if !(CatalogParentRef{SourceKey: "s", CatalogKey: "c"}).Active() {
		t.Error("both keys set should be active")
	}
	for _, c := range []CatalogParentRef{{}, {SourceKey: "s"}, {CatalogKey: "c"}} {
		if c.Active() {
			t.Errorf("%+v should be inactive", c)
		}
	}
}
