// Package catalogsync defines the future remote-sync contract for the
// component catalog as a local-only Go seam. Phase 2 (this milestone, C9)
// ships the interface, the value types, and a NoopSyncer default; it ships no
// networking, no auth, and no remote persistence — those belong to a Phase 3
// spec.
//
// # Authority
//
// Contract source: specs/orun-component-catalog/sync-model.md. The interface
// and payload shapes here mirror the local on-disk model so a future remote
// driver can stream the same objects without translation.
//
// # Boundary
//
// catalogsync is a pure contract package. It imports only catalogmodel (the
// persisted data types) and the standard library. It MUST NOT import
// net/http, internal/runner, internal/catalogstore, or cmd/orun — verified by
// the import-boundary test in this package. Keeping the seam free of the CLI,
// the store, the runner, and any network/auth concern is what makes it safe
// for Phase 3 to drop in a real Syncer without touching callers.
//
// # Replaceability
//
// The CLI depends only on the Syncer interface (wired to NoopSyncer today).
// Phase 3 replaces the constructed implementation; no caller signature
// changes.
package catalogsync
