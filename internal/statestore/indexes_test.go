package statestore_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/sourceplane/orun/internal/statestore"
	"github.com/sourceplane/orun/internal/testfx/statefs"
)

func sampleRevisionIndexEntry() statestore.RevisionIndexEntry {
	return statestore.RevisionIndexEntry{
		RevisionKey: "rev-pr139-def456a-p8f31c09",
		RevisionID:  "rev_01JABC",
		TriggerKey:  "trg-pr139-def456a",
		PlanHash:    "sha256:8f31c09",
		CreatedAt:   fixedTime(),
		Path:        "revisions/rev-pr139-def456a-p8f31c09",
	}
}

func sampleExecutionIndexEntry() statestore.ExecutionIndexEntry {
	return statestore.ExecutionIndexEntry{
		ExecutionKey: "run-001",
		ExecutionID:  "exec_01JXYZ",
		RevisionKey:  "rev-pr139-def456a-p8f31c09",
		Status:       "completed",
		CreatedAt:    fixedTime(),
		Path:         "revisions/rev-pr139-def456a-p8f31c09/executions/run-001",
	}
}

func TestRevisionIndex_RoundTripAndJSONStability(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := filepath.Join(statefs.NewWorkspace(t), ".orun")
	s, err := statestore.NewLocalStore(statestore.LocalConfig{Root: root})
	if err != nil {
		t.Fatalf("NewLocalStore: %v", err)
	}

	want := sampleRevisionIndexEntry()
	if _, err := statestore.WriteRevisionIndex(ctx, s, want); err != nil {
		t.Fatalf("WriteRevisionIndex: %v", err)
	}

	got, _, err := statestore.ReadRevisionIndex(ctx, s, want.RevisionKey)
	if err != nil {
		t.Fatalf("ReadRevisionIndex: %v", err)
	}
	if got != want {
		t.Fatalf("mismatch:\n got=%+v\nwant=%+v", got, want)
	}

	statefs.AssertJSONFile(t,
		filepath.Join(root, "indexes", "revisions", want.RevisionKey+".json"),
		map[string]any{
			"revisionKey": want.RevisionKey,
			"revisionId":  want.RevisionID,
			"triggerKey":  want.TriggerKey,
			"planHash":    want.PlanHash,
			"createdAt":   want.CreatedAt.Format(time.RFC3339Nano),
			"path":        want.Path,
		})
}

func TestRevisionIndex_DoubleWriteReturnsErrExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := filepath.Join(statefs.NewWorkspace(t), ".orun")
	s, _ := statestore.NewLocalStore(statestore.LocalConfig{Root: root})

	want := sampleRevisionIndexEntry()
	if _, err := statestore.WriteRevisionIndex(ctx, s, want); err != nil {
		t.Fatalf("first write: %v", err)
	}
	_, err := statestore.WriteRevisionIndex(ctx, s, want)
	if !errors.Is(err, statestore.ErrExists) {
		t.Fatalf("err=%v, want ErrExists", err)
	}
}

func TestRevisionIndex_MissingReturnsErrNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := filepath.Join(statefs.NewWorkspace(t), ".orun")
	s, _ := statestore.NewLocalStore(statestore.LocalConfig{Root: root})

	_, _, err := statestore.ReadRevisionIndex(ctx, s, "rev-never-written")
	if !errors.Is(err, statestore.ErrNotFound) {
		t.Fatalf("err=%v, want ErrNotFound", err)
	}
}

func TestRevisionIndex_InvalidKeyReturnsErrInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := filepath.Join(statefs.NewWorkspace(t), ".orun")
	s, _ := statestore.NewLocalStore(statestore.LocalConfig{Root: root})

	if _, _, err := statestore.ReadRevisionIndex(ctx, s, "bad/key"); !errors.Is(err, statestore.ErrInvalid) {
		t.Fatalf("read err=%v want ErrInvalid", err)
	}
	bad := sampleRevisionIndexEntry()
	bad.RevisionKey = "bad key!"
	if _, err := statestore.WriteRevisionIndex(ctx, s, bad); !errors.Is(err, statestore.ErrInvalid) {
		t.Fatalf("write err=%v want ErrInvalid", err)
	}
}

func TestExecutionIndex_RoundTripAndJSONStability(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := filepath.Join(statefs.NewWorkspace(t), ".orun")
	s, _ := statestore.NewLocalStore(statestore.LocalConfig{Root: root})

	want := sampleExecutionIndexEntry()
	if _, err := statestore.WriteExecutionIndex(ctx, s, want); err != nil {
		t.Fatalf("WriteExecutionIndex: %v", err)
	}
	got, _, err := statestore.ReadExecutionIndex(ctx, s, want.ExecutionKey)
	if err != nil {
		t.Fatalf("ReadExecutionIndex: %v", err)
	}
	if got != want {
		t.Fatalf("mismatch:\n got=%+v\nwant=%+v", got, want)
	}

	statefs.AssertJSONFile(t,
		filepath.Join(root, "indexes", "executions", want.ExecutionKey+".json"),
		map[string]any{
			"executionKey": want.ExecutionKey,
			"executionId":  want.ExecutionID,
			"revisionKey":  want.RevisionKey,
			"status":       want.Status,
			"createdAt":    want.CreatedAt.Format(time.RFC3339Nano),
			"path":         want.Path,
		})
}

func TestExecutionIndex_DoubleWriteReturnsErrExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := filepath.Join(statefs.NewWorkspace(t), ".orun")
	s, _ := statestore.NewLocalStore(statestore.LocalConfig{Root: root})

	want := sampleExecutionIndexEntry()
	if _, err := statestore.WriteExecutionIndex(ctx, s, want); err != nil {
		t.Fatalf("first: %v", err)
	}
	if _, err := statestore.WriteExecutionIndex(ctx, s, want); !errors.Is(err, statestore.ErrExists) {
		t.Fatalf("err=%v, want ErrExists", err)
	}
}

func TestExecutionIndex_MissingAndInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := filepath.Join(statefs.NewWorkspace(t), ".orun")
	s, _ := statestore.NewLocalStore(statestore.LocalConfig{Root: root})

	if _, _, err := statestore.ReadExecutionIndex(ctx, s, "run-never"); !errors.Is(err, statestore.ErrNotFound) {
		t.Fatalf("missing err=%v want ErrNotFound", err)
	}
	if _, _, err := statestore.ReadExecutionIndex(ctx, s, "bad/key"); !errors.Is(err, statestore.ErrInvalid) {
		t.Fatalf("invalid err=%v want ErrInvalid", err)
	}
	bad := sampleExecutionIndexEntry()
	bad.ExecutionKey = "bad key!"
	if _, err := statestore.WriteExecutionIndex(ctx, s, bad); !errors.Is(err, statestore.ErrInvalid) {
		t.Fatalf("invalid write err=%v want ErrInvalid", err)
	}
}

func TestRebuildIndexes_StubReturnsErrInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := filepath.Join(statefs.NewWorkspace(t), ".orun")
	s, _ := statestore.NewLocalStore(statestore.LocalConfig{Root: root})

	err := statestore.RebuildIndexes(ctx, s)
	if !errors.Is(err, statestore.ErrInvalid) {
		t.Fatalf("err=%v, want ErrInvalid", err)
	}
}
