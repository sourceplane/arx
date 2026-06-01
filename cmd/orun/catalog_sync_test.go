package main

// catalog_sync_test.go is the C9 CLI suite for `orun catalog refresh --sync`.
// It proves the local refresh still runs and persists, the wired
// catalogsync.NoopSyncer reports the documented not-configured warning, the
// command exits 0, and the --json envelope exposes the sync result
// deterministically alongside the unchanged refresh fields.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sourceplane/orun/internal/catalogsync"
)

func TestCatalogRefresh_Sync_Text_Exit0(t *testing.T) {
	dir := withTempIntentRoot(t)
	seedGitCatalogWorkspace(t, dir)
	resetCatalogFlags(t)
	catalogSyncFlag = true

	out := captureStdout(t, func() error {
		if err := runCatalogRefresh(nil); err != nil {
			t.Fatalf("refresh --sync must exit 0, got %v", err)
		}
		return nil
	})

	// Local refresh still happened.
	if !strings.Contains(out, "Catalog snapshot created") {
		t.Errorf("missing local refresh summary, got:\n%s", out)
	}
	// Sync section + documented warning present.
	if !strings.Contains(out, "Sync:") {
		t.Errorf("missing Sync section, got:\n%s", out)
	}
	if !strings.Contains(out, "remote sync not configured") {
		t.Errorf("missing not-configured warning, got:\n%s", out)
	}
}

func TestCatalogRefresh_Sync_JSON_StableEnvelope(t *testing.T) {
	dir := withTempIntentRoot(t)
	seedGitCatalogWorkspace(t, dir)
	resetCatalogFlags(t)
	catalogSyncFlag = true
	catalogJSONFlag = true

	out := captureStdout(t, func() error {
		if err := runCatalogRefresh(nil); err != nil {
			t.Fatalf("refresh --sync --json must exit 0, got %v", err)
		}
		return nil
	})

	var env catalogEnvelope
	var data catalogRefreshData
	env.Data = &data
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("refresh envelope: %v\n%s", err, out)
	}
	if env.Kind != kindCatalogRefreshResult {
		t.Errorf("kind = %q, want %q", env.Kind, kindCatalogRefreshResult)
	}
	// Existing fields are unchanged / present.
	if data.CatalogSnapshotKey == "" || !data.Created {
		t.Errorf("expected a created snapshot with a key, got %+v", data)
	}
	// Sync block is present and carries the noop result deterministically.
	if data.Sync == nil {
		t.Fatal("expected data.sync to be present under --sync --json")
	}
	if !data.Sync.Requested {
		t.Error("data.sync.requested should be true")
	}
	if data.Sync.Accepted {
		t.Error("noop syncer must report accepted=false")
	}
	if len(data.Sync.Warnings) != 1 || data.Sync.Warnings[0] != catalogsync.NoopWarning {
		t.Errorf("data.sync.warnings = %v, want [%q]", data.Sync.Warnings, catalogsync.NoopWarning)
	}
}

// TestCatalogRefresh_NoSync_OmitsSyncField proves the JSON envelope omits the
// sync field entirely when --sync is not requested, so existing consumers see
// no change.
func TestCatalogRefresh_NoSync_OmitsSyncField(t *testing.T) {
	dir := withTempIntentRoot(t)
	seedGitCatalogWorkspace(t, dir)
	resetCatalogFlags(t)
	catalogJSONFlag = true

	out := captureStdout(t, func() error { return runCatalogRefresh(nil) })

	// Raw JSON must not contain a "sync" key when --sync is off.
	if strings.Contains(out, "\"sync\"") {
		t.Errorf("envelope must omit sync when --sync is off, got:\n%s", out)
	}

	var env catalogEnvelope
	var data catalogRefreshData
	env.Data = &data
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("refresh envelope: %v\n%s", err, out)
	}
	if data.Sync != nil {
		t.Errorf("data.sync must be nil without --sync, got %+v", data.Sync)
	}
}

// TestCatalogRefresh_Sync_Deterministic proves the --sync --json envelope is
// byte-identical across repeated runs (idempotent reuse path included).
func TestCatalogRefresh_Sync_Deterministic(t *testing.T) {
	dir := withTempIntentRoot(t)
	seedGitCatalogWorkspace(t, dir)
	resetCatalogFlags(t)
	catalogSyncFlag = true
	catalogJSONFlag = true

	// First run creates; subsequent runs reuse. The sync block is identical
	// in both, so compare the two reuse runs for byte stability.
	_ = captureStdout(t, func() error { return runCatalogRefresh(nil) })
	first := captureStdout(t, func() error { return runCatalogRefresh(nil) })
	second := captureStdout(t, func() error { return runCatalogRefresh(nil) })
	if first != second {
		t.Errorf("reuse --sync output non-deterministic:\n first=%s\nsecond=%s", first, second)
	}
	if !strings.Contains(first, catalogsync.NoopWarning) {
		t.Errorf("reuse path lost the sync warning:\n%s", first)
	}
}
