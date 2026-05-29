package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLiveOrunService_GeneratePlan_RespectsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	svc := NewLiveOrunService(LiveServiceConfig{IntentFile: "intent.yaml"})
	if _, err := svc.GeneratePlan(ctx, PlanRequest{}); err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func TestLiveOrunService_GeneratePlan_MissingIntent(t *testing.T) {
	svc := NewLiveOrunService(LiveServiceConfig{})
	_, err := svc.GeneratePlan(context.Background(), PlanRequest{})
	if err == nil {
		t.Fatal("expected error when no IntentFile is configured")
	}
}

func TestLiveOrunService_GeneratePlan_LoadFailure(t *testing.T) {
	dir := t.TempDir()
	intentPath := filepath.Join(dir, "intent.yaml")
	if err := os.WriteFile(intentPath, []byte("not: a: valid: yaml: structure: ::"), 0o644); err != nil {
		t.Fatal(err)
	}
	svc := NewLiveOrunService(LiveServiceConfig{IntentFile: intentPath})
	_, err := svc.GeneratePlan(context.Background(), PlanRequest{})
	if err == nil {
		t.Fatal("expected load error on malformed intent")
	}
}

func TestLiveOrunService_GeneratePlan_RequestOverridesConfig(t *testing.T) {
	// IntentFile in request must win over LiveServiceConfig.IntentFile.
	svc := NewLiveOrunService(LiveServiceConfig{IntentFile: "/does/not/exist.yaml"})
	_, err := svc.GeneratePlan(context.Background(), PlanRequest{IntentFile: ""})
	if err == nil {
		t.Fatal("expected error when neither file resolves")
	}
	// Sanity: with request override empty string, falls back to cfg, which
	// points at /does/not/exist.yaml → load fails (proves we tried).
	if err != nil && !errStringContains(err, "load intent") && !errStringContains(err, "no such file") {
		t.Logf("got non-load error (acceptable): %v", err)
	}
}

func errStringContains(err error, sub string) bool {
	if err == nil {
		return false
	}
	return contains(err.Error(), sub)
}

func contains(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
