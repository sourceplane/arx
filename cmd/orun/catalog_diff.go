package main

// catalog_diff.go registers `orun catalog diff` as a documented stub. The full
// snapshot-diff engine (changed/added/removed components + graph changes,
// cli-surface.md §6) lands in a later milestone (C8); shipping the registered
// command now keeps the subcommand index and help surface stable and gives the
// flag grammar (--base/--head/--json) a tested home.
//
// The command prints a clear "not yet implemented" line and exits 5 — the §6
// "resolver failure" code is reused as the not-ready signal so scripts that
// branch on a zero exit do not mistake the stub for a successful empty diff.

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	catalogDiffBaseFlag string
	catalogDiffHeadFlag string
)

func registerCatalogDiffCommand(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "diff [component]",
		Short: "Compare two catalog snapshots (not yet implemented — C8)",
		Long: `Compare two catalog snapshots.

NOT YET IMPLEMENTED. The diff engine (changed / added / removed components and
graph changes) lands in milestone C8. This command is registered now so the
flag grammar and subcommand index are stable; it currently prints a
not-implemented notice and exits 5.

Examples (future):
  orun catalog diff --base main --head current
  orun catalog diff api-edge --base main --head pr-139
  orun catalog diff --json

Exit codes:
  5  Not yet implemented (C8). Will become: 0 on success, 5 on resolver failure.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCatalogDiff(cmd.Context())
		},
	}

	addCatalogSelectorFlags(cmd)
	cmd.Flags().StringVar(&catalogDiffBaseFlag, "base", "", "Base snapshot selector (C8)")
	cmd.Flags().StringVar(&catalogDiffHeadFlag, "head", "", "Head snapshot selector (C8)")
	cmd.Flags().BoolVar(&catalogJSONFlag, "json", false, "Stable machine-readable output (C8)")

	parent.AddCommand(cmd)
}

func runCatalogDiff(ctx context.Context) error {
	_ = ctx
	fmt.Fprintln(os.Stderr, "orun catalog diff: not yet implemented (lands in milestone C8)")
	return exitErr(5, "catalog diff not implemented")
}
