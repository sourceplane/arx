package statestore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Typed ref helpers — thin marshal/unmarshal wrappers over the StateStore
// primitives for the four ref doc shapes in data-model.md §6.
//
// All ref documents are encoded with the canonical encoder defined here:
// 2-space indent, no HTML escaping, trailing newline, struct fields emitted
// in declaration order. Two writes of identical struct values produce
// byte-identical output across Go versions; round-trip tests compare bytes.
//
// Reader helpers return (value, ObjectMeta, error). The ObjectMeta carries
// the content-derived Revision the caller must thread into a CAS helper
// (see state-store.md §3.3, §6 — CAS is used only on refs and indexes,
// loser-retries live in the caller per the design's writer-order).
//
// Writer helpers come in two flavors:
//   - Write*(ctx, store, value)                       — unconditional Write
//   - CAS*(ctx, store, prev ObjectMeta, value)        — CompareAndSwap with
//     prev.Revision as the oldRev. The helper does NOT re-read inside; the
//     caller threaded prev from a prior Read.
//
// All helpers route paths through paths.go; there is zero string
// concatenation for ref paths in this file.

// LatestRevisionRef encodes refs/latest-revision.json (data-model.md §6.1).
type LatestRevisionRef struct {
	RevisionKey        string    `json:"revisionKey"`
	RevisionID         string    `json:"revisionId"`
	PlanHash           string    `json:"planHash"`
	SourceSnapshotKey  string    `json:"sourceSnapshotKey,omitempty"`
	CatalogSnapshotKey string    `json:"catalogSnapshotKey,omitempty"`
	CreatedAt          time.Time `json:"createdAt"`
}

// LatestExecutionRef encodes refs/latest-execution.json (data-model.md §6.2).
type LatestExecutionRef struct {
	RevisionKey        string    `json:"revisionKey"`
	ExecutionKey       string    `json:"executionKey"`
	ExecutionID        string    `json:"executionId"`
	SourceSnapshotKey  string    `json:"sourceSnapshotKey,omitempty"`
	CatalogSnapshotKey string    `json:"catalogSnapshotKey,omitempty"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"createdAt"`
}

// TriggerRef encodes refs/triggers/<name>/latest.json and
// refs/triggers/<name>/<scope>.json (data-model.md §6.3). The two file
// locations share a single shape; the path is selected by the writer.
type TriggerRef struct {
	TriggerName        string    `json:"triggerName"`
	TriggerKey         string    `json:"triggerKey"`
	RevisionKey        string    `json:"revisionKey"`
	LatestExecutionKey string    `json:"latestExecutionKey"`
	HeadRevision       string    `json:"headRevision"`
	CreatedAt          time.Time `json:"createdAt"`
}

// NamedRef encodes refs/named/<name>.json (data-model.md §6.4). Phase 1 ships
// the helpers; CLI sugar to manage these aliases is reserved for later.
type NamedRef struct {
	Name        string    `json:"name"`
	RevisionKey string    `json:"revisionKey"`
	RevisionID  string    `json:"revisionId"`
	PlanHash    string    `json:"planHash"`
	CreatedAt   time.Time `json:"createdAt"`
}

// marshalCanonicalJSON renders v as 2-space-indented JSON with no HTML
// escaping and a trailing newline. The struct types in this file have only
// named fields, so encoding/json emits keys in declaration order — making
// the output byte-stable across Go versions.
//
// The encoder cannot fail for the typed values we marshal (no channels,
// funcs, or cyclic structures), so any error here would be a programmer
// mistake (e.g. adding an unsupported field type to one of the ref/index
// structs). Surface it as a panic rather than smuggling a sentinel out
// through the writer signatures.
func marshalCanonicalJSON(v any) []byte {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		panic(fmt.Sprintf("statestore: canonical JSON encode failed: %v", err))
	}
	// json.Encoder.Encode appends a trailing '\n' already — that's the
	// canonical terminator we want; do not strip it.
	return buf.Bytes()
}

// unmarshalRef strictly decodes raw into out. Decode errors wrap ErrInvalid
// so callers see a sentinel-routed failure rather than a raw json error.
func unmarshalRef(raw []byte, out any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("%w: decode ref: %v", ErrInvalid, err)
	}
	return nil
}

// readRef is the shared Read+decode path. It propagates ErrNotFound and
// ErrInvalid through unchanged.
func readRef(ctx context.Context, store StateStore, path string, out any) (ObjectMeta, error) {
	raw, meta, err := store.Read(ctx, path)
	if err != nil {
		return ObjectMeta{}, err
	}
	if err := unmarshalRef(raw, out); err != nil {
		return meta, err
	}
	return meta, nil
}

// writeRef is the shared marshal+Write path used by every unconditional ref
// writer.
func writeRef(ctx context.Context, store StateStore, path string, v any) (ObjectMeta, error) {
	return store.Write(ctx, path, marshalCanonicalJSON(v), WriteOptions{})
}

// casRef is the shared marshal+CAS path used by every conditional ref
// writer. prev.Revision is forwarded verbatim; the caller is responsible
// for supplying a non-empty oldRev (an empty oldRev will surface as a
// driver-side ErrConflict against any existing object).
func casRef(ctx context.Context, store StateStore, path string, prev ObjectMeta, v any) (ObjectMeta, error) {
	return store.CompareAndSwap(ctx, path, prev.Revision, marshalCanonicalJSON(v))
}

// ReadLatestRevisionRef reads refs/latest-revision.json. ErrNotFound is
// returned unchanged when the ref has never been written.
func ReadLatestRevisionRef(ctx context.Context, store StateStore) (LatestRevisionRef, ObjectMeta, error) {
	var out LatestRevisionRef
	meta, err := readRef(ctx, store, LatestRevisionRefPath(), &out)
	return out, meta, err
}

// WriteLatestRevisionRef unconditionally writes refs/latest-revision.json.
// Use CASLatestRevisionRef for the loser-retry path described in
// state-store.md §3.3.
func WriteLatestRevisionRef(ctx context.Context, store StateStore, v LatestRevisionRef) (ObjectMeta, error) {
	return writeRef(ctx, store, LatestRevisionRefPath(), v)
}

// CASLatestRevisionRef performs a CompareAndSwap against
// refs/latest-revision.json using prev.Revision as the oldRev. Returns
// ErrConflict on revision mismatch and ErrNotFound if the ref does not
// exist. The helper does NOT re-read the ref; the caller must supply
// prev from a prior Read so loser-retries observe a stable revision per
// state-store.md §6.
func CASLatestRevisionRef(ctx context.Context, store StateStore, prev ObjectMeta, next LatestRevisionRef) (ObjectMeta, error) {
	return casRef(ctx, store, LatestRevisionRefPath(), prev, next)
}

// ReadLatestExecutionRef reads refs/latest-execution.json.
func ReadLatestExecutionRef(ctx context.Context, store StateStore) (LatestExecutionRef, ObjectMeta, error) {
	var out LatestExecutionRef
	meta, err := readRef(ctx, store, LatestExecutionRefPath(), &out)
	return out, meta, err
}

// WriteLatestExecutionRef unconditionally writes refs/latest-execution.json.
func WriteLatestExecutionRef(ctx context.Context, store StateStore, v LatestExecutionRef) (ObjectMeta, error) {
	return writeRef(ctx, store, LatestExecutionRefPath(), v)
}

// CASLatestExecutionRef performs a CompareAndSwap against
// refs/latest-execution.json using prev.Revision as the oldRev.
func CASLatestExecutionRef(ctx context.Context, store StateStore, prev ObjectMeta, next LatestExecutionRef) (ObjectMeta, error) {
	return casRef(ctx, store, LatestExecutionRefPath(), prev, next)
}

// TriggerRefScope identifies which of the two trigger ref locations a helper
// addresses: the per-trigger latest pointer (Latest=true) or a per-scope
// pointer (Latest=false, Scope set).
type TriggerRefScope struct {
	// Name is the trigger name (single path component, e.g. "github-pull-request").
	Name string
	// Latest selects refs/triggers/<Name>/latest.json. When true, Scope is
	// ignored.
	Latest bool
	// Scope selects refs/triggers/<Name>/<Scope>.json. Required when
	// Latest is false; both Name and Scope are validated as single
	// components by the path helper.
	Scope string
}

func (s TriggerRefScope) path() (string, error) {
	if err := ValidateComponent(s.Name); err != nil {
		return "", err
	}
	if s.Latest {
		return TriggerLatestRefPath(s.Name), nil
	}
	if err := ValidateComponent(s.Scope); err != nil {
		return "", err
	}
	return TriggerScopeRefPath(s.Name, s.Scope), nil
}

// ReadTriggerRef reads either the per-trigger latest pointer or a per-scope
// pointer, selected by scope.
func ReadTriggerRef(ctx context.Context, store StateStore, scope TriggerRefScope) (TriggerRef, ObjectMeta, error) {
	path, err := scope.path()
	if err != nil {
		return TriggerRef{}, ObjectMeta{}, err
	}
	var out TriggerRef
	meta, err := readRef(ctx, store, path, &out)
	return out, meta, err
}

// WriteTriggerRef unconditionally writes the trigger ref selected by scope.
func WriteTriggerRef(ctx context.Context, store StateStore, scope TriggerRefScope, v TriggerRef) (ObjectMeta, error) {
	path, err := scope.path()
	if err != nil {
		return ObjectMeta{}, err
	}
	return writeRef(ctx, store, path, v)
}

// CASTriggerRef performs a CompareAndSwap against the trigger ref selected
// by scope using prev.Revision as the oldRev.
func CASTriggerRef(ctx context.Context, store StateStore, scope TriggerRefScope, prev ObjectMeta, next TriggerRef) (ObjectMeta, error) {
	path, err := scope.path()
	if err != nil {
		return ObjectMeta{}, err
	}
	return casRef(ctx, store, path, prev, next)
}

// ReadNamedRef reads refs/named/<name>.json.
func ReadNamedRef(ctx context.Context, store StateStore, name string) (NamedRef, ObjectMeta, error) {
	if err := ValidateComponent(name); err != nil {
		return NamedRef{}, ObjectMeta{}, err
	}
	var out NamedRef
	meta, err := readRef(ctx, store, NamedRefPath(name), &out)
	return out, meta, err
}

// WriteNamedRef unconditionally writes refs/named/<name>.json.
func WriteNamedRef(ctx context.Context, store StateStore, name string, v NamedRef) (ObjectMeta, error) {
	if err := ValidateComponent(name); err != nil {
		return ObjectMeta{}, err
	}
	return writeRef(ctx, store, NamedRefPath(name), v)
}

// CASNamedRef performs a CompareAndSwap against refs/named/<name>.json using
// prev.Revision as the oldRev.
func CASNamedRef(ctx context.Context, store StateStore, name string, prev ObjectMeta, next NamedRef) (ObjectMeta, error) {
	if err := ValidateComponent(name); err != nil {
		return ObjectMeta{}, err
	}
	return casRef(ctx, store, NamedRefPath(name), prev, next)
}

// (no package-level guard needed; imports are all used by exported helpers.)
