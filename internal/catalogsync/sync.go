package catalogsync

import (
	"context"

	"github.com/sourceplane/orun/internal/catalogmodel"
)

// SyncPayload is the union of everything a remote driver needs to reconstruct
// a catalog snapshot. It mirrors the local file shapes (sync-model.md §1) so a
// Phase 3 driver can stream objects without translation. Every field is a
// catalogmodel type — the payload carries no CLI, store, or runner coupling.
//
// Indexes: the catalog-local denormalized indexes (owner/system/domain/type)
// are recomputable from the manifests and are therefore not streamed; the
// cross-source component global index is included as the canonical
// denormalized view (a future driver may also rebuild it via the store's
// RebuildIndexes, catalog-store.md §8).
type SyncPayload struct {
	Source        catalogmodel.SourceSnapshot          `json:"source"`
	Catalog       catalogmodel.CatalogSnapshot         `json:"catalog"`
	Manifests     []catalogmodel.ComponentManifest     `json:"manifests"`
	Graphs        []catalogmodel.CatalogGraph          `json:"graphs"`
	GlobalIndexes []catalogmodel.ComponentGlobalIndex  `json:"globalIndexes"`
	SourceRef     *catalogmodel.SourceRef              `json:"sourceRef,omitempty"`
	CatalogRef    *catalogmodel.CatalogRef             `json:"catalogRef,omitempty"`
	HistoryEvents []catalogmodel.ComponentHistoryEvent `json:"historyEvents"`
}

// PushOptions carries local control metadata for a push. None of these fields
// imply any network behavior in Phase 2; they shape the future audit log and
// the dirty/preview write rules (sync-model.md §5).
type PushOptions struct {
	// DryRun requests a validate-only push (no remote write) once a real
	// driver exists. NoopSyncer ignores it.
	DryRun bool `json:"dryRun"`
	// AllowDirty permits pushing a snapshot resolved from a dirty worktree.
	// The Phase 3 rule additionally gates this on intent configuration.
	AllowDirty bool `json:"allowDirty"`
	// Reason is free-form text recorded in the future audit log.
	Reason string `json:"reason,omitempty"`
	// ExtraMetadata is arbitrary key/value context attached to the push.
	ExtraMetadata map[string]string `json:"extraMetadata,omitempty"`
}

// PushResult reports the outcome of a push. In Phase 2 the only implementation
// (NoopSyncer) returns Accepted=false with a single not-configured warning.
type PushResult struct {
	// Accepted reports whether the remote accepted the snapshot. Always
	// false under NoopSyncer.
	Accepted bool `json:"accepted"`
	// RemoteSourceKey / RemoteCatalogKey are the keys the remote assigned.
	// Empty under NoopSyncer.
	RemoteSourceKey  string `json:"remoteSourceKey,omitempty"`
	RemoteCatalogKey string `json:"remoteCatalogKey,omitempty"`
	// Warnings are human-readable advisories. NoopSyncer returns exactly
	// the not-configured notice.
	Warnings []string `json:"warnings"`
}

// Syncer is the remote-sync contract. Phase 2 ships only NoopSyncer; Phase 3
// supplies a networking implementation behind the same interface.
type Syncer interface {
	// PushCatalogSnapshot pushes a resolved local snapshot to the remote.
	// Implementations must be safe to call with a fully local payload and
	// must not require network availability to return a result.
	PushCatalogSnapshot(ctx context.Context, snapshot SyncPayload, opts PushOptions) (PushResult, error)
}

// NoopWarning is the documented Phase 2 advisory returned by NoopSyncer and
// surfaced by `orun catalog refresh --sync`. Exported so callers and tests can
// reference the exact string rather than duplicating the literal.
const NoopWarning = "remote sync not configured (Phase 3)"

// NoopSyncer is the default Syncer: it performs no networking, accepts
// nothing, and returns the single not-configured warning with no error. It is
// the wired implementation for all of Phase 2.
type NoopSyncer struct{}

// PushCatalogSnapshot implements Syncer. It ignores the payload and options
// entirely and reports the snapshot was not accepted because remote sync is
// not configured. It never returns an error and never touches the network.
func (NoopSyncer) PushCatalogSnapshot(_ context.Context, _ SyncPayload, _ PushOptions) (PushResult, error) {
	return PushResult{
		Accepted: false,
		Warnings: []string{NoopWarning},
	}, nil
}

// Compile-time assertion that NoopSyncer satisfies Syncer.
var _ Syncer = NoopSyncer{}
