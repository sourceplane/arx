package statestore

import (
	"context"
	"fmt"
	"time"
)

// Typed index helpers — thin marshal/unmarshal wrappers over CreateIfAbsent
// for the two index doc shapes in data-model.md §7.
//
// Index entries are immutable once written: WriteRevisionIndex /
// WriteExecutionIndex use CreateIfAbsent and surface ErrExists on a
// duplicate write. Idempotent re-writes are the caller's responsibility
// (M3 will dedupe before write); a true rebuild path goes through
// delete-then-create, or the dedicated RebuildIndexes helper once it lands.
//
// JSON encoding matches refs.go: 2-space indent, no HTML escaping, trailing
// newline, struct fields in declaration order. Round-trip tests compare
// bytes for stability across Go versions.
//
// Every path is constructed via paths.go (RevisionIndexPath /
// ExecutionIndexPath) — zero string concatenation in this file.

// RevisionIndexEntry encodes indexes/revisions/<revisionKey>.json
// (data-model.md §7.1).
type RevisionIndexEntry struct {
	RevisionKey        string    `json:"revisionKey"`
	RevisionID         string    `json:"revisionId"`
	TriggerKey         string    `json:"triggerKey"`
	PlanHash           string    `json:"planHash"`
	SourceSnapshotKey  string    `json:"sourceSnapshotKey,omitempty"`
	CatalogSnapshotKey string    `json:"catalogSnapshotKey,omitempty"`
	CreatedAt          time.Time `json:"createdAt"`
	Path               string    `json:"path"`
}

// ExecutionIndexEntry encodes indexes/executions/<executionKey>.json
// (data-model.md §7.2).
type ExecutionIndexEntry struct {
	ExecutionKey       string    `json:"executionKey"`
	ExecutionID        string    `json:"executionId"`
	RevisionKey        string    `json:"revisionKey"`
	SourceSnapshotKey  string    `json:"sourceSnapshotKey,omitempty"`
	CatalogSnapshotKey string    `json:"catalogSnapshotKey,omitempty"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"createdAt"`
	Path               string    `json:"path"`
}

// ReadRevisionIndex reads indexes/revisions/<revisionKey>.json. Returns
// ErrNotFound if no entry has been written for the key, ErrInvalid if
// the key violates path policy.
func ReadRevisionIndex(ctx context.Context, store StateStore, revisionKey string) (RevisionIndexEntry, ObjectMeta, error) {
	if err := ValidateComponent(revisionKey); err != nil {
		return RevisionIndexEntry{}, ObjectMeta{}, err
	}
	var out RevisionIndexEntry
	meta, err := readRef(ctx, store, RevisionIndexPath(revisionKey), &out)
	return out, meta, err
}

// WriteRevisionIndex creates indexes/revisions/<entry.RevisionKey>.json via
// CreateIfAbsent. Re-writing an existing index entry returns ErrExists; the
// caller handles dedup before this call.
func WriteRevisionIndex(ctx context.Context, store StateStore, entry RevisionIndexEntry) (ObjectMeta, error) {
	if err := ValidateComponent(entry.RevisionKey); err != nil {
		return ObjectMeta{}, err
	}
	return store.CreateIfAbsent(ctx, RevisionIndexPath(entry.RevisionKey), marshalCanonicalJSON(entry))
}

// ReadExecutionIndex reads indexes/executions/<executionKey>.json.
func ReadExecutionIndex(ctx context.Context, store StateStore, executionKey string) (ExecutionIndexEntry, ObjectMeta, error) {
	if err := ValidateComponent(executionKey); err != nil {
		return ExecutionIndexEntry{}, ObjectMeta{}, err
	}
	var out ExecutionIndexEntry
	meta, err := readRef(ctx, store, ExecutionIndexPath(executionKey), &out)
	return out, meta, err
}

// WriteExecutionIndex creates indexes/executions/<entry.ExecutionKey>.json
// via CreateIfAbsent.
func WriteExecutionIndex(ctx context.Context, store StateStore, entry ExecutionIndexEntry) (ObjectMeta, error) {
	if err := ValidateComponent(entry.ExecutionKey); err != nil {
		return ObjectMeta{}, err
	}
	return store.CreateIfAbsent(ctx, ExecutionIndexPath(entry.ExecutionKey), marshalCanonicalJSON(entry))
}

// RebuildIndexes is reserved for M3+. It exists so consumers can wire
// against the final symbol now; calling it always returns an error wrapping
// ErrInvalid. The real implementation lands when revisions actually
// populate (data-model.md §7, design.md §5.1 ordered writes).
//
// Do not advertise this helper as functional in package docs.
func RebuildIndexes(ctx context.Context, store StateStore) error {
	_ = ctx
	_ = store
	return fmt.Errorf("%w: RebuildIndexes is deferred to M3+", ErrInvalid)
}
