package main

// catalog_list.go implements `orun catalog list`: enumerate the components in
// the selected catalog with the cli-surface.md §3 columns
// (COMPONENT, TYPE, OWNER, SYSTEM, LAST EXEC, STATUS) and the
// --owner/--system/--domain/--type/--status filters.
//
// Data source (task-0038 Integration Note): the non-component axes of the
// catalog-local indexes are intentionally empty in PR-1, so type/owner/system/
// domain are derived directly from each resolved ComponentManifest (a manifest
// walk over the resolved CatalogSnapshot), and the lastExecution*/STATUS
// columns from the catalog-local ComponentExecutionIndex newest row. This
// needs no new index axis. Rows are sorted by componentKey for byte-stable
// output (the EnumerateComponentManifests seam guarantees this).

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/sourceplane/orun/internal/catalogmodel"
	"github.com/sourceplane/orun/internal/catalogstore"
	"github.com/sourceplane/orun/internal/statestore"
	"github.com/sourceplane/orun/internal/ui"
	"github.com/spf13/cobra"
)

// catalog list filter flag values. Package scope so the cobra bindings and
// the RunE body share them; reset per-invocation by cobra.
var (
	catalogListOwnerFlag  string
	catalogListSystemFlag string
	catalogListDomainFlag string
	catalogListTypeFlag   string
	catalogListStatusFlag string
)

// catalogListRow is one row of the CatalogListResult envelope `data` array.
// Field names are the stable §3 JSON contract.
type catalogListRow struct {
	ComponentKey        string `json:"componentKey"`
	Name                string `json:"name"`
	Type                string `json:"type"`
	Owner               string `json:"owner"`
	System              string `json:"system"`
	LastRevisionKey     string `json:"lastRevisionKey"`
	LastExecutionKey    string `json:"lastExecutionKey"`
	LastExecutionStatus string `json:"lastExecutionStatus"`
	SourceSnapshotKey   string `json:"sourceSnapshotKey"`
	CatalogSnapshotKey  string `json:"catalogSnapshotKey"`
}

func registerCatalogListCommand(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the components in the selected catalog",
		Long: `List the components in the selected catalog.

Resolves the catalog via the shared source selector (default 'current') and
prints one row per component with its type, owner, system, and last execution
status. The filter flags narrow the set; output is sorted by component key.

Examples:
  orun catalog list
  orun catalog list --source main
  orun catalog list --owner team/platform-edge
  orun catalog list --type cloudflare-worker
  orun catalog list --json

Exit codes:
  0  Listing rendered (possibly empty).
  1  Invalid selector.
  3  StateStore failure.
  6  Catalog not found.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCatalogList(cmd.Context())
		},
	}

	addCatalogSelectorFlags(cmd)
	cmd.Flags().StringVar(&catalogListOwnerFlag, "owner", "", "Only components with this owner")
	cmd.Flags().StringVar(&catalogListSystemFlag, "system", "", "Only components in this system")
	cmd.Flags().StringVar(&catalogListDomainFlag, "domain", "", "Only components in this domain")
	cmd.Flags().StringVar(&catalogListTypeFlag, "type", "", "Only components of this type")
	cmd.Flags().StringVar(&catalogListStatusFlag, "status", "", "Only components whose last execution has this status")
	cmd.Flags().BoolVar(&catalogJSONFlag, "json", false, "Stable machine-readable output")

	parent.AddCommand(cmd)
}

func runCatalogList(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	sel, err := parseCatalogSelector()
	if err != nil {
		return err
	}

	stateStore, _, err := openLocalStateStore()
	if err != nil {
		return exitErr(3, "open state store: %w", err)
	}
	store := catalogstore.New(stateStore)

	cat, err := store.ResolveCatalog(ctx, sel)
	if err != nil {
		return catalogReadExit(err, "resolve catalog")
	}

	manifests, err := catalogstore.EnumerateComponentManifests(ctx, stateStore, cat)
	if err != nil {
		return exitErr(3, "enumerate components: %w", err)
	}

	rows := make([]catalogListRow, 0, len(manifests))
	for _, m := range manifests {
		lastRev, lastExec, lastStatus := lastExecutionFor(ctx, stateStore, cat, m.Identity.Name)
		row := catalogListRow{
			ComponentKey:        m.Identity.ComponentKey,
			Name:                m.Identity.Name,
			Type:                m.Spec.Type,
			Owner:               m.Metadata.Owner,
			System:              m.Spec.System,
			LastRevisionKey:     lastRev,
			LastExecutionKey:    lastExec,
			LastExecutionStatus: lastStatus,
			SourceSnapshotKey:   cat.SourceSnapshotKey,
			CatalogSnapshotKey:  cat.CatalogSnapshotKey,
		}
		if !catalogListRowMatches(row, m) {
			continue
		}
		rows = append(rows, row)
	}

	if catalogJSONFlag {
		return writeCatalogEnvelope(kindCatalogListResult, rows, nil)
	}
	return renderCatalogListText(rows)
}

// catalogListRowMatches applies the §3 filter flags. domain is matched
// against the manifest's spec.domain (not surfaced as a column but a valid
// filter); the other axes match the rendered row fields.
func catalogListRowMatches(row catalogListRow, m catalogmodel.ComponentManifest) bool {
	if catalogListOwnerFlag != "" && row.Owner != catalogListOwnerFlag {
		return false
	}
	if catalogListSystemFlag != "" && row.System != catalogListSystemFlag {
		return false
	}
	if catalogListTypeFlag != "" && row.Type != catalogListTypeFlag {
		return false
	}
	if catalogListDomainFlag != "" && m.Spec.Domain != catalogListDomainFlag {
		return false
	}
	if catalogListStatusFlag != "" && row.LastExecutionStatus != catalogListStatusFlag {
		return false
	}
	return true
}

// lastExecutionFor returns the newest execution row's revision/execution/status
// for a component from the catalog-local ComponentExecutionIndex. The index is
// absent (and the triple empty) until executions are appended in C7; absence is
// not an error. Rows are sorted newest-first by CreatedAt before the head is
// taken, matching the §7 history ordering so list and history agree.
func lastExecutionFor(ctx context.Context, stateStore statestore.StateStore, cat catalogmodel.CatalogSnapshot, name string) (string, string, string) {
	idx, found, err := catalogstore.ReadComponentExecutionIndex(ctx, stateStore, cat.SourceSnapshotKey, cat.CatalogSnapshotKey, name)
	if err != nil || !found || len(idx.Executions) == 0 {
		return "", "", ""
	}
	rows := append([]catalogmodel.ComponentExecutionRow(nil), idx.Executions...)
	sort.SliceStable(rows, func(a, b int) bool {
		return rows[a].CreatedAt > rows[b].CreatedAt
	})
	head := rows[0]
	return head.RevisionKey, head.ExecutionKey, head.Status
}

func renderCatalogListText(rows []catalogListRow) error {
	color := ui.ColorEnabledForWriter(os.Stdout)
	if len(rows) == 0 {
		fmt.Fprintln(os.Stdout, "No components in the selected catalog.")
		return nil
	}
	fmt.Fprintf(os.Stdout, "%s\n\n", ui.Bold(color, "Catalog components"))
	fmt.Fprintf(os.Stdout, "%-28s %-22s %-22s %-16s %-24s %s\n",
		"COMPONENT", "TYPE", "OWNER", "SYSTEM", "LAST EXEC", "STATUS")
	for _, r := range rows {
		lastExec := r.LastExecutionKey
		if lastExec == "" {
			lastExec = "-"
		}
		status := r.LastExecutionStatus
		if status == "" {
			status = "-"
		}
		fmt.Fprintf(os.Stdout, "%-28s %-22s %-22s %-16s %-24s %s\n",
			truncField(r.Name, 28), truncField(r.Type, 22), truncField(r.Owner, 22),
			truncField(r.System, 16), truncField(lastExec, 24), status)
	}
	return nil
}

// truncField clamps a column value so the fixed-width table stays aligned for
// pathologically long values. Values within width are returned unchanged.
func truncField(s string, width int) string {
	if len(s) <= width {
		return s
	}
	if width <= 1 {
		return s[:width]
	}
	return s[:width-1] + "…"
}
