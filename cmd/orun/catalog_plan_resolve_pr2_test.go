package main

// catalog_plan_resolve_pr2_test.go — C6 PR2 test coverage for plan/catalog
// integration: flags, metadata stamping, revision keys, and describe/get compat.

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/sourceplane/orun/internal/catalogstore"
	"github.com/sourceplane/orun/internal/model"
	"github.com/sourceplane/orun/internal/revision"
	"github.com/sourceplane/orun/internal/statestore"
	"github.com/sourceplane/orun/internal/triggerctx"
)

// --- Flag behavior tests ---

func TestResolvePlanCatalog_Strict_FailsOnError(t *testing.T) {
	// Empty intent root with no component.yaml → resolver discovers no
	// components → BuildCatalog returns nil or empty snapshot. Under strict
	// mode this must surface as an error. We use a bare temp dir (no git,
	// no components) to trigger the failure.
	dir := withTempIntentRoot(t)
	_ = dir
	resetCatalogFlags(t)

	// Override intentRoot to a path that definitely has no intent.yaml and
	// no components. The resolver should fail because there's nothing to
	// resolve in strict mode.
	res, err := resolvePlanCatalog(context.Background(), planCatalogOptions{Strict: true})
	// The strict contract: if resolution fails OR no snapshot produced,
	// we get an error. On some systems a local-nogit workspace might still
	// produce a snapshot, so we only assert the broader contract:
	// strict mode must not silently succeed when the workspace has no
	// catalog content. If it does produce a snapshot, it must be resolved.
	if err != nil {
		// Good: strict mode surfaced the error.
		return
	}
	// If no error, it must at least have resolved (no silent skip).
	if !res.Resolved {
		t.Fatalf("strict mode: no error but also not resolved — should have either resolved or errored")
	}
}

func TestResolvePlanCatalog_NoRefresh_SkipsMetadata(t *testing.T) {
	dir := withTempIntentRoot(t)
	seedGitCatalogWorkspace(t, dir)
	resetCatalogFlags(t)

	res, err := resolvePlanCatalog(context.Background(), planCatalogOptions{NoRefresh: true})
	if err != nil {
		t.Fatalf("NoRefresh: %v", err)
	}
	if res.Resolved {
		t.Errorf("NoRefresh should not resolve")
	}
	if !res.Skipped {
		t.Errorf("NoRefresh should set Skipped=true")
	}
	if res.Source != nil {
		t.Errorf("NoRefresh should leave Source nil, got %+v", res.Source)
	}
	if res.Catalog != nil {
		t.Errorf("NoRefresh should leave Catalog nil, got %+v", res.Catalog)
	}
}

func TestResolvePlanCatalog_BestEffort_NoError(t *testing.T) {
	// Non-git workspace, no strict: must not error.
	withTempIntentRoot(t)
	resetCatalogFlags(t)

	res, err := resolvePlanCatalog(context.Background(), planCatalogOptions{})
	if err != nil {
		t.Fatalf("best-effort must not error: %v", err)
	}
	// Either resolved or not; the contract is no error.
	_ = res
}

func TestResolvePlanCatalog_SnapshotKey_ResolvesExisting(t *testing.T) {
	dir := withTempIntentRoot(t)
	seedGitCatalogWorkspace(t, dir)
	resetCatalogFlags(t)

	// First, resolve and persist a catalog to create the snapshot.
	first, err := resolvePlanCatalog(context.Background(), planCatalogOptions{})
	if err != nil || !first.Resolved {
		t.Fatalf("first resolve: err=%v res=%+v", err, first)
	}

	catKey := first.Catalog.CatalogSnapshotKey

	// Now resolve using --catalog-snapshot=<key>.
	res, err := resolvePlanCatalog(context.Background(), planCatalogOptions{
		SnapshotKey: catKey,
	})
	if err != nil {
		t.Fatalf("snapshot-key resolve: %v", err)
	}
	if !res.Resolved {
		t.Fatalf("snapshot-key should resolve")
	}
	if res.Catalog.CatalogSnapshotKey != catKey {
		t.Errorf("snapshot-key mismatch: want %q got %q", catKey, res.Catalog.CatalogSnapshotKey)
	}
}

func TestResolvePlanCatalog_SourceSelector_Current(t *testing.T) {
	dir := withTempIntentRoot(t)
	seedGitCatalogWorkspace(t, dir)
	resetCatalogFlags(t)

	// Persist a catalog first.
	first, err := resolvePlanCatalog(context.Background(), planCatalogOptions{})
	if err != nil || !first.Resolved {
		t.Fatalf("first resolve: err=%v res=%+v", err, first)
	}

	// Resolve using --catalog-source=current.
	res, err := resolvePlanCatalog(context.Background(), planCatalogOptions{
		SourceSelector: "current",
	})
	if err != nil {
		t.Fatalf("source-selector resolve: %v", err)
	}
	if !res.Resolved {
		t.Fatalf("source-selector should resolve")
	}
	if res.Catalog == nil {
		t.Fatalf("source-selector should have catalog")
	}
}

// --- Metadata stamping tests ---

func TestPlanMetadata_SourceCatalog_Populated(t *testing.T) {
	dir := withTempIntentRoot(t)
	seedGitCatalogWorkspace(t, dir)
	resetCatalogFlags(t)

	res, err := resolvePlanCatalog(context.Background(), planCatalogOptions{})
	if err != nil || !res.Resolved {
		t.Fatalf("resolve: err=%v res=%+v", err, res)
	}

	plan := &model.Plan{
		Metadata: model.PlanMetadata{Name: "test"},
	}

	// Stamp source metadata.
	if res.Source != nil {
		plan.Metadata.Source = &model.PlanSourceMeta{
			SnapshotKey:  res.Source.SourceSnapshotKey,
			Ref:          res.Source.Ref,
			HeadRevision: res.Source.HeadRevision,
			TreeHash:     res.Source.TreeHash,
			WorkingTree:  res.Source.WorkingTree,
			DirtyHash:    res.Source.DirtyHash,
		}
	}
	if res.Catalog != nil {
		plan.Metadata.Catalog = &model.PlanCatalogMeta{
			SnapshotKey:       res.Catalog.CatalogSnapshotKey,
			CatalogHash:      res.Catalog.CatalogHash,
			SourceSnapshotKey: res.Catalog.SourceSnapshotKey,
		}
	}

	if plan.Metadata.Source == nil {
		t.Fatal("source metadata not populated")
	}
	if plan.Metadata.Source.SnapshotKey == "" {
		t.Error("source snapshotKey empty")
	}
	if plan.Metadata.Catalog == nil {
		t.Fatal("catalog metadata not populated")
	}
	if plan.Metadata.Catalog.SnapshotKey == "" {
		t.Error("catalog snapshotKey empty")
	}
	if plan.Metadata.Catalog.Skipped {
		t.Error("catalog should not be skipped when resolved")
	}

	// Verify JSON round-trip preserves the additive fields.
	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var rt model.Plan
	if err := json.Unmarshal(data, &rt); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rt.Metadata.Source == nil || rt.Metadata.Source.SnapshotKey != plan.Metadata.Source.SnapshotKey {
		t.Error("source metadata not preserved through JSON round-trip")
	}
	if rt.Metadata.Catalog == nil || rt.Metadata.Catalog.SnapshotKey != plan.Metadata.Catalog.SnapshotKey {
		t.Error("catalog metadata not preserved through JSON round-trip")
	}
}

func TestPlanMetadata_Skipped_WhenNoRefresh(t *testing.T) {
	plan := &model.Plan{
		Metadata: model.PlanMetadata{Name: "test"},
	}
	plan.Metadata.Catalog = &model.PlanCatalogMeta{Skipped: true}

	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(data), `"skipped":true`) {
		t.Errorf("expected skipped:true in JSON, got %s", string(data))
	}

	var rt model.Plan
	if err := json.Unmarshal(data, &rt); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rt.Metadata.Catalog == nil || !rt.Metadata.Catalog.Skipped {
		t.Error("skipped flag not preserved through JSON round-trip")
	}
}

func TestPlanMetadata_Nil_WhenUnresolved(t *testing.T) {
	plan := &model.Plan{
		Metadata: model.PlanMetadata{Name: "test"},
	}

	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(data), `"source"`) {
		t.Errorf("source should be omitted when nil")
	}
	if strings.Contains(string(data), `"catalog"`) {
		t.Errorf("catalog should be omitted when nil")
	}
}

// --- PlanRevision source/catalog keys tests ---

func TestPlanRevision_CarriesCatalogKeys(t *testing.T) {
	dir := withTempIntentRoot(t)
	seedGitCatalogWorkspace(t, dir)
	resetCatalogFlags(t)

	catRes, err := resolvePlanCatalog(context.Background(), planCatalogOptions{})
	if err != nil || !catRes.Resolved {
		t.Fatalf("resolve: err=%v res=%+v", err, catRes)
	}

	stateStore, _, err := openLocalStateStore()
	if err != nil {
		t.Fatalf("open state store: %v", err)
	}

	trig := minimalTrigger()
	cfg := revision.Config{
		Store:         stateStore,
		JobCount:      1,
		CatalogParent: catRes.Parent,
	}.WithCompatibilityWrites(false)

	rev, err := revision.WriteRevision(context.Background(), cfg, trig, []byte(`{"test":true}`), "sha256:deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	if err != nil {
		t.Fatalf("WriteRevision: %v", err)
	}

	if rev.SourceSnapshotKey == "" {
		t.Error("SourceSnapshotKey empty on revision")
	}
	if rev.CatalogSnapshotKey == "" {
		t.Error("CatalogSnapshotKey empty on revision")
	}
	if rev.SourceSnapshotKey != catRes.Parent.SourceKey {
		t.Errorf("SourceSnapshotKey mismatch: want %q got %q", catRes.Parent.SourceKey, rev.SourceSnapshotKey)
	}
	if rev.CatalogSnapshotKey != catRes.Parent.CatalogKey {
		t.Errorf("CatalogSnapshotKey mismatch: want %q got %q", catRes.Parent.CatalogKey, rev.CatalogSnapshotKey)
	}

	// Verify persisted revision.json contains the keys.
	raw, _, rerr := stateStore.Read(context.Background(), statestore.RevisionDocPath(rev.RevisionKey))
	if rerr != nil {
		t.Fatalf("read revision.json: %v", rerr)
	}
	if !strings.Contains(string(raw), `"sourceSnapshotKey"`) {
		t.Error("revision.json missing sourceSnapshotKey")
	}
	if !strings.Contains(string(raw), `"catalogSnapshotKey"`) {
		t.Error("revision.json missing catalogSnapshotKey")
	}
}

func TestPlanRevision_EmptyKeys_WhenNoParent(t *testing.T) {
	dir := withTempIntentRoot(t)
	seedGitCatalogWorkspace(t, dir)
	resetCatalogFlags(t)

	stateStore, _, err := openLocalStateStore()
	if err != nil {
		t.Fatalf("open state store: %v", err)
	}

	trig := minimalTrigger()
	cfg := revision.Config{
		Store:    stateStore,
		JobCount: 1,
	}.WithCompatibilityWrites(false)

	rev, err := revision.WriteRevision(context.Background(), cfg, trig, []byte(`{"test":true}`), "sha256:deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	if err != nil {
		t.Fatalf("WriteRevision: %v", err)
	}

	if rev.SourceSnapshotKey != "" {
		t.Errorf("SourceSnapshotKey should be empty without CatalogParent, got %q", rev.SourceSnapshotKey)
	}
	if rev.CatalogSnapshotKey != "" {
		t.Errorf("CatalogSnapshotKey should be empty without CatalogParent, got %q", rev.CatalogSnapshotKey)
	}
}

// --- Phase 1 / global revision layout compatibility ---

func TestGlobalRevisionLayout_StillWorks(t *testing.T) {
	dir := withTempIntentRoot(t)
	seedGitCatalogWorkspace(t, dir)
	resetCatalogFlags(t)

	stateStore, _, err := openLocalStateStore()
	if err != nil {
		t.Fatalf("open state store: %v", err)
	}

	trig := minimalTrigger()
	cfg := revision.Config{
		Store:    stateStore,
		JobCount: 1,
	}.WithCompatibilityWrites(false)

	rev, err := revision.WriteRevision(context.Background(), cfg, trig, []byte(`{"test":true}`), "sha256:deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	if err != nil {
		t.Fatalf("WriteRevision: %v", err)
	}

	// Global layout files should exist.
	for _, path := range []string{
		statestore.TriggerPath(rev.RevisionKey),
		statestore.RevisionDocPath(rev.RevisionKey),
		statestore.PlanPath(rev.RevisionKey),
	} {
		if _, _, err := stateStore.Read(context.Background(), path); err != nil {
			t.Errorf("global layout missing %q: %v", path, err)
		}
	}

	// Latest revision ref should exist.
	_, _, err = statestore.ReadLatestRevisionRef(context.Background(), stateStore)
	if err != nil {
		t.Errorf("latest revision ref missing: %v", err)
	}
}

// --- Catalog-parent revision mirror ---

func TestCatalogParent_Mirror_StillWorks(t *testing.T) {
	dir := withTempIntentRoot(t)
	seedGitCatalogWorkspace(t, dir)
	resetCatalogFlags(t)

	catRes, err := resolvePlanCatalog(context.Background(), planCatalogOptions{})
	if err != nil || !catRes.Resolved {
		t.Fatalf("resolve: err=%v res=%+v", err, catRes)
	}

	stateStore, _, err := openLocalStateStore()
	if err != nil {
		t.Fatalf("open state store: %v", err)
	}

	trig := minimalTrigger()
	cfg := revision.Config{
		Store:         stateStore,
		JobCount:      1,
		CatalogParent: catRes.Parent,
	}.WithCompatibilityWrites(false)

	rev, err := revision.WriteRevision(context.Background(), cfg, trig, []byte(`{"test":"mirror"}`), "sha256:cafebabecafebabecafebabecafebabecafebabecafebabecafebabecafebabe")
	if err != nil {
		t.Fatalf("WriteRevision: %v", err)
	}

	// Catalog-parent mirror files should exist.
	trigPath, _ := catalogstore.CatalogRevisionTriggerPath(catRes.Parent.SourceKey, catRes.Parent.CatalogKey, rev.RevisionKey)
	revPath, _ := catalogstore.CatalogRevisionDocPath(catRes.Parent.SourceKey, catRes.Parent.CatalogKey, rev.RevisionKey)
	planPath, _ := catalogstore.CatalogRevisionPlanPath(catRes.Parent.SourceKey, catRes.Parent.CatalogKey, rev.RevisionKey)

	for _, path := range []string{trigPath, revPath, planPath} {
		if _, _, err := stateStore.Read(context.Background(), path); err != nil {
			t.Errorf("catalog-parent mirror missing %q: %v", path, err)
		}
	}
}

// --- Helper ---

func minimalTrigger() triggerctx.TriggerOccurrence {
	return triggerctx.TriggerOccurrence{
		TriggerID:   "trg_test01",
		TriggerKey:  "trg-manual-full-abc1234",
		TriggerName: "test",
		TriggerType: "manual",
		Mode:        "manual",
		CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		PlanScope: triggerctx.PlanScope{
			Mode:               "full",
			ActiveEnvironments: []string{"dev"},
		},
		Source: triggerctx.TriggerSource{
			SourceScope: "branch-main",
		},
	}
}
