package main

// catalog_describe.go implements `orun catalog describe <component>`: resolve
// one ComponentManifest from the selected catalog and render the cli-surface.md
// §4 section list (text) or the full manifest + catalog-local execution rows
// (--json under {manifest, executions}).
//
// Component selectors (§4):
//   - bare name (api-edge)               → resolved within the current catalog
//   - fully-qualified key (ns/repo/name) → exact componentKey match
//   - ambiguous bare name across repos   → exit 4 with the candidate list
//
// PR-2 scope note: cross-repo ambiguity can only arise once the catalog holds
// components from multiple repos. The resolver's ResolveComponent does a direct
// name lookup, so this file additionally scans the enumerated manifest set to
// detect a fully-qualified-key request and to surface the §4 ambiguity exit.

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/sourceplane/orun/internal/catalogmodel"
	"github.com/sourceplane/orun/internal/catalogstore"
	"github.com/sourceplane/orun/internal/ui"
	"github.com/spf13/cobra"
)

// catalogDescribeData is the --json payload: the full manifest plus the
// catalog-local execution rows for the component (§4).
type catalogDescribeData struct {
	Manifest   catalogmodel.ComponentManifest       `json:"manifest"`
	Executions []catalogmodel.ComponentExecutionRow `json:"executions"`
}

func registerCatalogDescribeCommand(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "describe <component>",
		Short: "Show the full resolved manifest for one component",
		Long: `Show the full resolved manifest for one component in the selected catalog.

The component may be given as a bare name (resolved within the catalog) or as
a fully-qualified component key (namespace/repo/name) for an exact match. A
bare name that matches components in more than one repo exits 4 with the list
of candidate keys.

Examples:
  orun catalog describe api-edge
  orun catalog describe api-edge --source main
  orun catalog describe sourceplane/orun/api-edge
  orun catalog describe api-edge --json

Exit codes:
  0  Component found and rendered.
  1  Invalid selector or missing component argument.
  3  StateStore failure.
  4  Ambiguous bare name across repos.
  6  Catalog or component not found.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCatalogDescribe(cmd.Context(), args[0])
		},
	}

	addCatalogSelectorFlags(cmd)
	cmd.Flags().BoolVar(&catalogJSONFlag, "json", false, "Stable machine-readable output")

	parent.AddCommand(cmd)
}

func runCatalogDescribe(ctx context.Context, arg string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return exitErr(1, "describe requires a component name")
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

	manifest, err := selectComponent(manifests, arg)
	if err != nil {
		return err
	}

	idx, _, ierr := catalogstore.ReadComponentExecutionIndex(ctx, stateStore, cat.SourceSnapshotKey, cat.CatalogSnapshotKey, manifest.Identity.Name)
	if ierr != nil {
		return exitErr(3, "read execution index: %w", ierr)
	}
	execs := idx.Executions
	if execs == nil {
		execs = []catalogmodel.ComponentExecutionRow{}
	}

	if catalogJSONFlag {
		return writeCatalogEnvelope(kindCatalogDescribeResult, catalogDescribeData{
			Manifest:   manifest,
			Executions: execs,
		}, nil)
	}
	return renderCatalogDescribeText(manifest, execs)
}

// selectComponent resolves the §4 component selector against the enumerated
// manifest set. A fully-qualified key (containing '/') is matched exactly; a
// bare name matches on Identity.Name and exits 4 when more than one repo
// supplies that name.
func selectComponent(manifests []catalogmodel.ComponentManifest, arg string) (catalogmodel.ComponentManifest, error) {
	if strings.Contains(arg, "/") {
		for _, m := range manifests {
			if m.Identity.ComponentKey == arg {
				return m, nil
			}
		}
		return catalogmodel.ComponentManifest{}, exitErr(6, "component %q not found in catalog", arg)
	}

	var matches []catalogmodel.ComponentManifest
	for _, m := range manifests {
		if m.Identity.Name == arg {
			matches = append(matches, m)
		}
	}
	switch len(matches) {
	case 0:
		return catalogmodel.ComponentManifest{}, exitErr(6, "component %q not found in catalog", arg)
	case 1:
		return matches[0], nil
	default:
		keys := make([]string, 0, len(matches))
		for _, m := range matches {
			keys = append(keys, m.Identity.ComponentKey)
		}
		sort.Strings(keys)
		return catalogmodel.ComponentManifest{}, exitErr(4,
			"component %q is ambiguous across repos; qualify with a full key: %s",
			arg, strings.Join(keys, ", "))
	}
}

func renderCatalogDescribeText(m catalogmodel.ComponentManifest, execs []catalogmodel.ComponentExecutionRow) error {
	color := ui.ColorEnabledForWriter(os.Stdout)
	out := os.Stdout

	section := func(title string) { fmt.Fprintf(out, "\n%s\n", ui.Bold(color, title)) }
	kv := func(k, v string) {
		if v != "" {
			fmt.Fprintf(out, "  %-14s %s\n", k+":", v)
		}
	}

	fmt.Fprintf(out, "%s\n", ui.Bold(color, m.Identity.Name))

	section("Component")
	kv("Key", m.Identity.ComponentKey)
	kv("Name", m.Identity.Name)
	kv("Type", m.Spec.Type)
	kv("Lifecycle", m.Spec.Lifecycle)
	kv("Title", m.Metadata.Title)
	kv("Description", m.Metadata.Description)

	section("Ownership")
	kv("Owner", m.Metadata.Owner)
	kv("System", m.Spec.System)
	kv("Domain", m.Spec.Domain)
	if len(m.Metadata.Maintainers) > 0 {
		kv("Maintainers", strings.Join(m.Metadata.Maintainers, ", "))
	}

	section("Source")
	kv("Source", m.Source.SourceSnapshotKey)
	kv("Catalog", m.Source.CatalogSnapshotKey)
	kv("Ref", m.Source.Ref)
	kv("Branch", m.Source.Branch)
	kv("Tree", m.Source.TreeHash)
	kv("WorkingTree", m.Source.WorkingTree)
	kv("ManifestHash", m.Source.ManifestHash)

	section("Environments")
	if len(m.Spec.Environments) == 0 {
		fmt.Fprintln(out, "  (none)")
	} else {
		for _, name := range sortedEnvKeys(m.Spec.Environments) {
			env := m.Spec.Environments[name]
			active := "inactive"
			if env.Active {
				active = "active"
			}
			fmt.Fprintf(out, "  %-14s profile=%s (%s)\n", name, env.Profile, active)
		}
	}

	section("Dependencies")
	if len(m.Spec.Dependencies.Components) == 0 {
		fmt.Fprintln(out, "  (none)")
	} else {
		for _, d := range m.Spec.Dependencies.Components {
			opt := ""
			if d.Optional {
				opt = " (optional)"
			}
			fmt.Fprintf(out, "  %s → %s [%s]%s\n", m.Identity.Name, d.Name, d.Relationship, opt)
		}
	}

	section("APIs")
	if len(m.Spec.Dependencies.APIs.Provides) > 0 {
		kv("Provides", strings.Join(m.Spec.Dependencies.APIs.Provides, ", "))
	}
	if len(m.Spec.Dependencies.APIs.Consumes) > 0 {
		kv("Consumes", strings.Join(m.Spec.Dependencies.APIs.Consumes, ", "))
	}
	if len(m.Spec.Dependencies.APIs.Provides) == 0 && len(m.Spec.Dependencies.APIs.Consumes) == 0 {
		fmt.Fprintln(out, "  (none)")
	}

	section("Resources")
	if len(m.Spec.Dependencies.Resources.Uses) == 0 {
		fmt.Fprintln(out, "  (none)")
	} else {
		kv("Uses", strings.Join(m.Spec.Dependencies.Resources.Uses, ", "))
	}

	section("Runtime inference")
	kv("Languages", strings.Join(m.Runtime.Inferred.Languages, ", "))
	kv("PackageMgrs", strings.Join(m.Runtime.Inferred.PackageManagers, ", "))
	kv("Frameworks", strings.Join(m.Runtime.Inferred.Frameworks, ", "))
	kv("Infra", strings.Join(m.Runtime.Inferred.Infra, ", "))

	section("Latest executions")
	if len(execs) == 0 {
		fmt.Fprintln(out, "  (none)")
	} else {
		rows := append([]catalogmodel.ComponentExecutionRow(nil), execs...)
		sort.SliceStable(rows, func(a, b int) bool { return rows[a].CreatedAt > rows[b].CreatedAt })
		for _, r := range rows {
			fmt.Fprintf(out, "  %-22s %-10s %s\n", r.ExecutionKey, r.Status, r.CreatedAt)
		}
	}

	section("Resolution provenance")
	if len(m.Resolution.InheritedFrom) == 0 && len(m.Resolution.InferredFrom) == 0 {
		fmt.Fprintln(out, "  (none)")
	} else {
		for _, k := range sortedStringKeys(m.Resolution.InheritedFrom) {
			fmt.Fprintf(out, "  %-22s inherited from %s\n", k, m.Resolution.InheritedFrom[k])
		}
		for _, k := range sortedSliceMapKeys(m.Resolution.InferredFrom) {
			fmt.Fprintf(out, "  %-22s inferred from %s\n", k, strings.Join(m.Resolution.InferredFrom[k], ", "))
		}
	}
	return nil
}

func sortedEnvKeys(m map[string]catalogmodel.ComponentEnvironment) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedStringKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedSliceMapKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
