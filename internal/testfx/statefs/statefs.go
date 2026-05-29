// Package statefs provides test helpers for the state-redesign milestones.
//
// All helpers in this package are intended for use from *_test.go files only.
// They MUST NOT import any other internal/ package — they sit beneath the
// state-redesign packages (triggerctx, statestore, revision, executionstate,
// e2e) in the dependency graph and exist to give those packages a stable,
// dependency-light testing surface.
//
// Helpers are safe for callers using t.Parallel(): every NewWorkspace returns
// a distinct temp directory created via t.TempDir, and the package keeps no
// shared global state.
package statefs

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// NewWorkspace returns the absolute path to an isolated, empty `.orun/` root
// for the calling test. The directory lives inside t.TempDir() and is removed
// automatically when the test (and its subtests) complete; callers do not need
// to register additional cleanup.
//
// The returned path is the workspace root itself — the `.orun/` directory has
// already been created beneath it, so callers can write straight into
// filepath.Join(root, ".orun", "…") without an extra MkdirAll.
func NewWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	orunDir := filepath.Join(root, ".orun")
	if err := os.MkdirAll(orunDir, 0o755); err != nil {
		t.Fatalf("statefs.NewWorkspace: mkdir %s: %v", orunDir, err)
	}
	return root
}

// AssertJSONFile reads the JSON file at path and fails the test if its
// decoded value does not equal expected.
//
// The comparison is schema-tolerant: both sides are round-tripped through
// encoding/json into interface{} values so map key order, integer-vs-float
// ambiguity, and pointer identity do not cause false negatives. On mismatch,
// both sides are pretty-printed so the diff is human-readable.
//
// expected may be any value that can be marshalled by encoding/json — most
// commonly a map[string]any literal, a struct, or another decoded any.
//
// The parameter is typed as testing.TB so the helper composes with
// *testing.T, *testing.B, and assertion-harness wrappers used in the
// state-redesign coverage tests. *testing.T satisfies testing.TB, so callers
// pass it directly.
func AssertJSONFile(t testing.TB, path string, expected any) {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("statefs.AssertJSONFile: read %s: %v", path, err)
		return
	}

	var got any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("statefs.AssertJSONFile: decode %s: %v\nraw: %s", path, err, raw)
		return
	}

	wantBytes, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("statefs.AssertJSONFile: marshal expected: %v", err)
		return
	}
	var want any
	if err := json.Unmarshal(wantBytes, &want); err != nil {
		t.Fatalf("statefs.AssertJSONFile: re-decode expected: %v", err)
		return
	}

	if !reflect.DeepEqual(got, want) {
		gotPretty, _ := json.MarshalIndent(got, "", "  ")
		wantPretty, _ := json.MarshalIndent(want, "", "  ")
		t.Fatalf("statefs.AssertJSONFile: %s mismatch\n--- got ---\n%s\n--- want ---\n%s",
			path, gotPretty, wantPretty)
	}
}

// ReadJSON strictly decodes the JSON file at path into a value of type T and
// returns it. Any read or decode failure aborts the test via t.Fatalf.
//
// Decoding uses encoding/json with DisallowUnknownFields enabled so callers
// catch schema drift between the on-disk format and the Go type they expect.
//
// Like AssertJSONFile the helper accepts testing.TB so it composes with
// benchmarks and assertion-harness wrappers; *testing.T callers pass it
// directly.
func ReadJSON[T any](t testing.TB, path string) T {
	t.Helper()

	var out T
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("statefs.ReadJSON: read %s: %v", path, err)
		return out
	}

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&out); err != nil {
		t.Fatalf("statefs.ReadJSON: decode %s into %T: %v\nraw: %s", path, out, err, raw)
		var zero T
		return zero
	}
	return out
}
