package statefs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewWorkspace_CreatesIsolatedOrunRoot(t *testing.T) {
	t.Parallel()

	a := NewWorkspace(t)
	b := NewWorkspace(t)

	if a == b {
		t.Fatalf("expected distinct workspace roots, got identical paths: %s", a)
	}

	for _, root := range []string{a, b} {
		orun := filepath.Join(root, ".orun")
		info, err := os.Stat(orun)
		if err != nil {
			t.Fatalf("stat %s: %v", orun, err)
		}
		if !info.IsDir() {
			t.Fatalf("%s: expected directory, got mode %v", orun, info.Mode())
		}
	}
}

func TestNewWorkspace_CleanupRemovesDir(t *testing.T) {
	t.Parallel()

	var captured string
	t.Run("sub", func(t *testing.T) {
		captured = NewWorkspace(t)
		if _, err := os.Stat(captured); err != nil {
			t.Fatalf("workspace should exist inside sub: %v", err)
		}
	})

	// After the subtest finishes, t.TempDir's cleanup runs and the
	// workspace must be gone.
	if _, err := os.Stat(captured); !os.IsNotExist(err) {
		t.Fatalf("expected workspace removed after subtest, got err=%v", err)
	}
}

func TestAssertJSONFile_MatchAndMismatch(t *testing.T) {
	t.Parallel()

	root := NewWorkspace(t)
	path := filepath.Join(root, ".orun", "obj.json")

	payload := map[string]any{
		"name":  "alpha",
		"count": 3,
		"tags":  []any{"x", "y"},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Match using a differently-ordered map and float-vs-int values for
	// the same logical content; schema-tolerant compare should pass.
	AssertJSONFile(t, path, map[string]any{
		"tags":  []string{"x", "y"},
		"count": float64(3),
		"name":  "alpha",
	})

	// Mismatch path: run inside a fake testing.T-like wrapper so the
	// failure does not fail the outer test.
	fake := &fakeT{TB: t}
	AssertJSONFile(fake, path, map[string]any{
		"name":  "alpha",
		"count": 4, // wrong
		"tags":  []any{"x", "y"},
	})
	if !fake.failed {
		t.Fatalf("expected AssertJSONFile to flag mismatch")
	}
}

func TestAssertJSONFile_MissingFile(t *testing.T) {
	t.Parallel()

	fake := &fakeT{TB: t}
	AssertJSONFile(fake, filepath.Join(NewWorkspace(t), "nope.json"), map[string]any{})
	if !fake.failed {
		t.Fatalf("expected failure on missing file")
	}
}

type sampleDoc struct {
	Name  string   `json:"name"`
	Count int      `json:"count"`
	Tags  []string `json:"tags"`
}

func TestReadJSON_HappyPath(t *testing.T) {
	t.Parallel()

	root := NewWorkspace(t)
	path := filepath.Join(root, ".orun", "doc.json")
	if err := os.WriteFile(path, []byte(`{"name":"alpha","count":3,"tags":["x","y"]}`), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	got := ReadJSON[sampleDoc](t, path)
	if got.Name != "alpha" || got.Count != 3 || len(got.Tags) != 2 || got.Tags[0] != "x" || got.Tags[1] != "y" {
		t.Fatalf("ReadJSON mismatch: got=%+v", got)
	}
}

func TestReadJSON_UnknownFieldRejected(t *testing.T) {
	t.Parallel()

	root := NewWorkspace(t)
	path := filepath.Join(root, ".orun", "drift.json")
	if err := os.WriteFile(path, []byte(`{"name":"alpha","count":3,"tags":["x"],"extra":true}`), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	fake := &fakeT{TB: t}
	_ = ReadJSON[sampleDoc](fake, path)
	if !fake.failed {
		t.Fatalf("expected ReadJSON to fail on unknown field")
	}
}

func TestReadJSON_MissingFile(t *testing.T) {
	t.Parallel()

	fake := &fakeT{TB: t}
	_ = ReadJSON[sampleDoc](fake, filepath.Join(NewWorkspace(t), "missing.json"))
	if !fake.failed {
		t.Fatalf("expected ReadJSON to fail on missing file")
	}
}

// fakeT intercepts Fatalf calls so tests can assert that a helper failed
// without aborting the outer test. It embeds *testing.T to satisfy the
// helper's *testing.T parameter; we only override the failure path.
type fakeT struct {
	testing.TB
	failed bool
}

func (f *fakeT) Fatalf(format string, args ...any) {
	f.failed = true
	// Swallow: do not call the embedded TB.Fatalf — that would abort the
	// outer test. The helpers under test return immediately after invoking
	// Fatalf (we added explicit `return` after each call), so it is safe
	// to let control flow continue past the failing call site.
	_, _ = format, args
}
