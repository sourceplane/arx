package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sourceplane/orun/internal/git"
	"github.com/sourceplane/orun/internal/model"
)

func TestCollectChangedComponents_ComponentManifestChangeOnlyMatchesOwningComponent(t *testing.T) {
	normalized := &model.NormalizedIntent{
		Components: map[string]model.Component{
			"docs-site-direct-upload": {
				Name:       "docs-site-direct-upload",
				Path:       "website",
				SourcePath: "website/component.yaml",
			},
			"api-edge-worker": {
				Name:       "api-edge-worker",
				Path:       "apps/api-edge",
				SourcePath: "apps/api-edge/component.yaml",
			},
		},
	}

	instances := map[string][]*model.ComponentInstance{
		"production": {
			{ComponentName: "docs-site-direct-upload", Path: "website"},
			{ComponentName: "api-edge-worker", Path: "apps/api-edge"},
		},
	}

	changed := collectChangedComponents(normalized, instances, map[string]struct{}{
		"website/component.yaml": {},
	}, "intent.yaml", git.ChangeOptions{})

	if !changed["docs-site-direct-upload"] {
		t.Fatal("expected docs-site-direct-upload to be marked changed")
	}
	if changed["api-edge-worker"] {
		t.Fatal("did not expect api-edge-worker to be marked changed by another component manifest")
	}
}

func TestCollectChangedComponents_IntentChangeMarksAllComponents(t *testing.T) {
	normalized := &model.NormalizedIntent{
		Components: map[string]model.Component{
			"docs-site-direct-upload": {
				Name:       "docs-site-direct-upload",
				Path:       "website",
				SourcePath: "website/component.yaml",
			},
			"api-edge-worker": {
				Name:       "api-edge-worker",
				Path:       "apps/api-edge",
				SourcePath: "apps/api-edge/component.yaml",
			},
		},
	}

	changed := collectChangedComponents(normalized, nil, map[string]struct{}{
		"nested/intent.yaml": {},
	}, "intent.yaml", git.ChangeOptions{})

	if !changed["docs-site-direct-upload"] || !changed["api-edge-worker"] {
		t.Fatal("expected intent.yaml change to mark all components changed")
	}
}

func TestCollectChangedComponents_AbsoluteIntentPathMatchesRelativeChangedFiles(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	normalized := &model.NormalizedIntent{
		Components: map[string]model.Component{
			"docs-site-direct-upload": {
				Name:       "docs-site-direct-upload",
				Path:       "website",
				SourcePath: "website/component.yaml",
			},
			"api-edge-worker": {
				Name:       "api-edge-worker",
				Path:       "apps/api-edge",
				SourcePath: "apps/api-edge/component.yaml",
			},
		},
	}

	// Simulate auto-discovery returning an absolute intentPath (the real scenario that was broken).
	absIntentPath := filepath.Join(cwd, "intent.yaml")
	changed := collectChangedComponents(normalized, nil, map[string]struct{}{
		"website/component.yaml": {},
	}, absIntentPath, git.ChangeOptions{})

	if !changed["docs-site-direct-upload"] {
		t.Fatal("expected docs-site-direct-upload to be changed when using absolute intentPath")
	}
	if changed["api-edge-worker"] {
		t.Fatal("did not expect api-edge-worker to be changed")
	}
}

func TestCollectChangedComponents_IntentSemanticDiff_InlineComponentOnly(t *testing.T) {
	// When intent.yaml change is detected but semantic diff says only components changed,
	// only the named components should be marked. We test this by passing Files option
	// which forces fallback to Global (safe behavior test).
	normalized := &model.NormalizedIntent{
		Components: map[string]model.Component{
			"web": {Name: "web", Path: "apps/web", SourcePath: "apps/web/component.yaml"},
			"api": {Name: "api", Path: "apps/api", SourcePath: "apps/api/component.yaml"},
		},
	}

	// With --files, semantic diff falls back to global (cannot access git refs)
	changed := collectChangedComponents(normalized, nil, map[string]struct{}{
		"intent.yaml": {},
	}, "intent.yaml", git.ChangeOptions{Files: []string{"intent.yaml"}})

	if !changed["web"] || !changed["api"] {
		t.Fatal("expected all components marked when using --files (global fallback)")
	}
}

func TestCollectChangedComponents_DiscoveredAndInlineUnion(t *testing.T) {
	// Even when intent marks some components, discovered component path changes
	// should still be detected via the normal path-based logic.
	normalized := &model.NormalizedIntent{
		Components: map[string]model.Component{
			"web":    {Name: "web", Path: "apps/web", SourcePath: "apps/web/component.yaml"},
			"api":    {Name: "api", Path: "apps/api", SourcePath: "apps/api/component.yaml"},
			"worker": {Name: "worker", Path: "apps/worker", SourcePath: "apps/worker/component.yaml"},
		},
	}

	instances := map[string][]*model.ComponentInstance{
		"production": {
			{ComponentName: "web", Path: "apps/web"},
			{ComponentName: "api", Path: "apps/api"},
			{ComponentName: "worker", Path: "apps/worker"},
		},
	}

	// Only the discovered component's files changed (no intent.yaml change)
	changed := collectChangedComponents(normalized, instances, map[string]struct{}{
		"apps/web/src/main.ts": {},
		"apps/api/handler.go":  {},
	}, "intent.yaml", git.ChangeOptions{})

	if !changed["web"] {
		t.Fatal("expected web to be changed (file under path)")
	}
	if !changed["api"] {
		t.Fatal("expected api to be changed (file under path)")
	}
	if changed["worker"] {
		t.Fatal("worker should not be changed")
	}
}
