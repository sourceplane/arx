package main

// catalog_plan_resolve_test.go — C6 PR1 integration coverage for the
// plan-path catalog resolver. Drives resolvePlanCatalog against a real seeded
// git workspace (reusing the catalog-refresh harness) and asserts:
//   - a clean branch-main workspace resolves a (source, catalog) parent;
//   - the resolved snapshot is persisted (catalog.json exists);
//   - --no-catalog-refresh short-circuits to Skipped with no parent;
//   - a non-git / unresolvable workspace degrades to !Resolved without error
//     (best-effort posture) — guarding the Phase 1 compat path.

import (
	"context"
	"strings"
	"testing"

	"github.com/sourceplane/orun/internal/catalogstore"
)

func TestResolvePlanCatalog_Resolves_BranchMain(t *testing.T) {
	dir := withTempIntentRoot(t)
	seedGitCatalogWorkspace(t, dir)
	resetCatalogFlags(t)

	res, err := resolvePlanCatalog(context.Background(), planCatalogOptions{})
	if err != nil {
		t.Fatalf("resolvePlanCatalog: %v", err)
	}
	if !res.Resolved {
		t.Fatalf("expected Resolved=true for clean branch-main workspace, got %+v", res)
	}
	if res.Skipped {
		t.Errorf("Skipped should be false when resolving, got %+v", res)
	}
	if res.Parent.SourceKey == "" || res.Parent.CatalogKey == "" {
		t.Fatalf("parent keys empty: %+v", res.Parent)
	}
	if !strings.HasPrefix(res.Parent.CatalogKey, "cat-") {
		t.Errorf("catalog key not cat-prefixed: %q", res.Parent.CatalogKey)
	}
	if res.Source == nil || res.Catalog == nil {
		t.Fatalf("Source/Catalog provenance missing: %+v", res)
	}
	if res.Catalog.CatalogSnapshotKey != res.Parent.CatalogKey {
		t.Errorf("catalog key mismatch: snapshot=%q parent=%q",
			res.Catalog.CatalogSnapshotKey, res.Parent.CatalogKey)
	}

	// The resolved snapshot must have been persisted: catalog.json present.
	store, _, err := openLocalStateStore()
	if err != nil {
		t.Fatalf("open state store: %v", err)
	}
	catPath, err := catalogstore.CatalogDocPath(res.Parent.SourceKey, res.Parent.CatalogKey)
	if err != nil {
		t.Fatalf("catalog doc path: %v", err)
	}
	if _, _, err := store.Read(context.Background(), catPath); err != nil {
		t.Errorf("catalog.json not persisted at %q: %v", catPath, err)
	}
}

func TestResolvePlanCatalog_Idempotent(t *testing.T) {
	dir := withTempIntentRoot(t)
	seedGitCatalogWorkspace(t, dir)
	resetCatalogFlags(t)

	first, err := resolvePlanCatalog(context.Background(), planCatalogOptions{})
	if err != nil || !first.Resolved {
		t.Fatalf("first resolve: err=%v res=%+v", err, first)
	}
	second, err := resolvePlanCatalog(context.Background(), planCatalogOptions{})
	if err != nil || !second.Resolved {
		t.Fatalf("second resolve: err=%v res=%+v", err, second)
	}
	if first.Parent != second.Parent {
		t.Errorf("idempotent parent drift: %+v != %+v", first.Parent, second.Parent)
	}
}

func TestResolvePlanCatalog_NoRefresh_Skips(t *testing.T) {
	dir := withTempIntentRoot(t)
	seedGitCatalogWorkspace(t, dir)
	resetCatalogFlags(t)

	res, err := resolvePlanCatalog(context.Background(), planCatalogOptions{NoRefresh: true})
	if err != nil {
		t.Fatalf("resolvePlanCatalog(NoRefresh): %v", err)
	}
	if res.Resolved {
		t.Errorf("NoRefresh should not resolve, got %+v", res)
	}
	if !res.Skipped {
		t.Errorf("NoRefresh should set Skipped=true, got %+v", res)
	}
	if res.Parent.SourceKey != "" || res.Parent.CatalogKey != "" {
		t.Errorf("NoRefresh should leave parent empty, got %+v", res.Parent)
	}
}

func TestResolvePlanCatalog_BestEffort_DegradesOnNoGit(t *testing.T) {
	// A bare temp intent root with NO git repo: the resolver can still
	// produce a local-nogit snapshot in some environments, so this test only
	// asserts the best-effort contract — never an error, and the plan path is
	// never blocked. Whether Resolved is true or false depends on the
	// resolver's local-nogit handling; both are acceptable here.
	withTempIntentRoot(t)
	resetCatalogFlags(t)

	res, err := resolvePlanCatalog(context.Background(), planCatalogOptions{})
	if err != nil {
		t.Fatalf("best-effort resolve must not error on no-git workspace, got %v", err)
	}
	_ = res // either Resolved or not; the contract is "no error, no panic".
}
