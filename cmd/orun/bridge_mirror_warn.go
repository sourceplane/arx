package main

// bridge_mirror_warn.go provides a shared helper for the four read-side
// commands (status, logs, describe, get) to surface bridge-mirror-failed
// events as one-line stderr warnings. The bridge writes these events under
// revisions/<revKey>/executions/<execKey>/events/ when a hardlink/copy
// mirror tick fails (data-model.md §9.1, design.md §11). Surfacing them as
// warnings — never errors — preserves the read-only contract of the
// commands while alerting operators to mirror lag.
//
// The scan is best-effort: malformed events directories are silently
// skipped, and a single warning is emitted per distinct execution even if
// many events exist. This keeps output noise-free during routine operation.

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/sourceplane/orun/internal/statestore"
)

// bridgeMirrorWarnSink is the io.Writer warnings are written to. Tests
// override it to capture warnings; production wiring leaves it as
// os.Stderr.
var bridgeMirrorWarnSink io.Writer = os.Stderr

// warnBridgeMirrorFailures emits a single stderr line per execution whose
// events/ directory contains at least one bridge-mirror-failed entry.
// The function is idempotent and safe to call from any read-side command.
//
// It is best-effort: errors enumerating events/ are dropped; the resolver
// path remains authoritative.
func warnBridgeMirrorFailures(ctx context.Context, store statestore.StateStore, revKey, execKey string) {
	if store == nil || revKey == "" || execKey == "" {
		return
	}
	dir := statestore.ExecutionDir(revKey, execKey) + "/events"
	infos, err := store.List(ctx, dir)
	if err != nil {
		return
	}
	for _, info := range infos {
		base := path.Base(info.Path)
		// Filenames are zero-padded "<seq>-<kind>.json"; we look
		// for the kind suffix.
		if strings.HasSuffix(base, "-bridge-mirror-failed.json") {
			fmt.Fprintf(bridgeMirrorWarnSink,
				"warning: bridge mirror failed for execution %s (revision %s) — see events/ for details\n",
				execKey, revKey)
			return
		}
	}
}
