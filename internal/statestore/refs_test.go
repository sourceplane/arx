package statestore_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sourceplane/orun/internal/statestore"
	"github.com/sourceplane/orun/internal/testfx/statefs"
)

func mustReadFile(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(b)
}

func fixedTime() time.Time {
	return time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC)
}

// --- LatestRevisionRef -------------------------------------------------------

func TestLatestRevisionRef_RoundTripAndJSONStability(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := filepath.Join(statefs.NewWorkspace(t), ".orun")
	s, err := statestore.NewLocalStore(statestore.LocalConfig{Root: root})
	if err != nil {
		t.Fatalf("NewLocalStore: %v", err)
	}

	want := statestore.LatestRevisionRef{
		RevisionKey: "rev-pr139-def456a-p8f31c09",
		RevisionID:  "rev_01JABC",
		PlanHash:    "sha256:8f31c09",
		CreatedAt:   fixedTime(),
	}

	if _, err := statestore.WriteLatestRevisionRef(ctx, s, want); err != nil {
		t.Fatalf("WriteLatestRevisionRef: %v", err)
	}

	got, meta, err := statestore.ReadLatestRevisionRef(ctx, s)
	if err != nil {
		t.Fatalf("ReadLatestRevisionRef: %v", err)
	}
	if got != want {
		t.Fatalf("round-trip mismatch:\n got=%+v\nwant=%+v", got, want)
	}
	if meta.Revision == "" {
		t.Fatalf("ObjectMeta.Revision empty")
	}

	// JSON byte-for-byte stability via the testfx golden helper.
	statefs.AssertJSONFile(t,
		filepath.Join(root, "refs", "latest-revision.json"),
		map[string]any{
			"revisionKey": want.RevisionKey,
			"revisionId":  want.RevisionID,
			"planHash":    want.PlanHash,
			"createdAt":   want.CreatedAt.Format(time.RFC3339Nano),
		})

	// Two writes of identical bytes produce identical Revision (the
	// content-derived sha256 contract from store.go).
	meta2, err := statestore.WriteLatestRevisionRef(ctx, s, want)
	if err != nil {
		t.Fatalf("re-write: %v", err)
	}
	if meta2.Revision != meta.Revision {
		t.Fatalf("revision changed across identical writes: %s vs %s", meta.Revision, meta2.Revision)
	}
}

func TestLatestRevisionRef_MissingReturnsErrNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := filepath.Join(statefs.NewWorkspace(t), ".orun")
	s, _ := statestore.NewLocalStore(statestore.LocalConfig{Root: root})

	_, _, err := statestore.ReadLatestRevisionRef(ctx, s)
	if !errors.Is(err, statestore.ErrNotFound) {
		t.Fatalf("err=%v, want ErrNotFound", err)
	}
}

func TestCASLatestRevisionRef_HappyPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := filepath.Join(statefs.NewWorkspace(t), ".orun")
	s, _ := statestore.NewLocalStore(statestore.LocalConfig{Root: root})

	v1 := statestore.LatestRevisionRef{RevisionKey: "rev-1", RevisionID: "id1", PlanHash: "h1", CreatedAt: fixedTime()}
	if _, err := statestore.WriteLatestRevisionRef(ctx, s, v1); err != nil {
		t.Fatalf("write v1: %v", err)
	}
	_, meta, err := statestore.ReadLatestRevisionRef(ctx, s)
	if err != nil {
		t.Fatalf("read v1: %v", err)
	}

	v2 := statestore.LatestRevisionRef{RevisionKey: "rev-2", RevisionID: "id2", PlanHash: "h2", CreatedAt: fixedTime().Add(time.Second)}
	if _, err := statestore.CASLatestRevisionRef(ctx, s, meta, v2); err != nil {
		t.Fatalf("CAS v2: %v", err)
	}
	got, _, err := statestore.ReadLatestRevisionRef(ctx, s)
	if err != nil {
		t.Fatalf("read v2: %v", err)
	}
	if got != v2 {
		t.Fatalf("expected v2 after CAS, got %+v", got)
	}
}

func TestCASLatestRevisionRef_StalePrevReturnsErrConflict(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := filepath.Join(statefs.NewWorkspace(t), ".orun")
	s, _ := statestore.NewLocalStore(statestore.LocalConfig{Root: root})

	v1 := statestore.LatestRevisionRef{RevisionKey: "rev-1", CreatedAt: fixedTime()}
	if _, err := statestore.WriteLatestRevisionRef(ctx, s, v1); err != nil {
		t.Fatalf("write v1: %v", err)
	}
	_, stale, _ := statestore.ReadLatestRevisionRef(ctx, s)

	// Concurrent winner advances the ref out from under us.
	v2 := statestore.LatestRevisionRef{RevisionKey: "rev-2", CreatedAt: fixedTime().Add(time.Second)}
	if _, err := statestore.WriteLatestRevisionRef(ctx, s, v2); err != nil {
		t.Fatalf("winner write: %v", err)
	}

	v3 := statestore.LatestRevisionRef{RevisionKey: "rev-3", CreatedAt: fixedTime().Add(2 * time.Second)}
	_, err := statestore.CASLatestRevisionRef(ctx, s, stale, v3)
	if !errors.Is(err, statestore.ErrConflict) {
		t.Fatalf("err=%v, want ErrConflict", err)
	}
}

func TestCASLatestRevisionRef_RaceExactlyOneWinner(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := filepath.Join(statefs.NewWorkspace(t), ".orun")
	s, _ := statestore.NewLocalStore(statestore.LocalConfig{Root: root})

	seed := statestore.LatestRevisionRef{RevisionKey: "rev-seed", CreatedAt: fixedTime()}
	if _, err := statestore.WriteLatestRevisionRef(ctx, s, seed); err != nil {
		t.Fatalf("seed: %v", err)
	}
	_, prev, _ := statestore.ReadLatestRevisionRef(ctx, s)

	const N = 16
	var wins, conflicts atomic.Int64
	var wg sync.WaitGroup
	wg.Add(N)
	start := make(chan struct{})
	for i := 0; i < N; i++ {
		i := i
		go func() {
			defer wg.Done()
			<-start
			next := statestore.LatestRevisionRef{
				RevisionKey: "rev-w-" + itoa(i),
				CreatedAt:   fixedTime().Add(time.Duration(i+1) * time.Second),
			}
			_, err := statestore.CASLatestRevisionRef(ctx, s, prev, next)
			switch {
			case err == nil:
				wins.Add(1)
			case errors.Is(err, statestore.ErrConflict):
				conflicts.Add(1)
			default:
				t.Errorf("unexpected err: %v", err)
			}
		}()
	}
	close(start)
	wg.Wait()

	if wins.Load() != 1 {
		t.Fatalf("winners=%d, want exactly 1 (conflicts=%d)", wins.Load(), conflicts.Load())
	}
	if wins.Load()+conflicts.Load() != int64(N) {
		t.Fatalf("accounting: wins=%d conflicts=%d N=%d", wins.Load(), conflicts.Load(), N)
	}
}

// --- LatestExecutionRef ------------------------------------------------------

func TestLatestExecutionRef_RoundTripAndCAS(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := filepath.Join(statefs.NewWorkspace(t), ".orun")
	s, _ := statestore.NewLocalStore(statestore.LocalConfig{Root: root})

	v := statestore.LatestExecutionRef{
		RevisionKey: "rev-1", ExecutionKey: "run-001", ExecutionID: "exec_01",
		Status: "completed", CreatedAt: fixedTime(),
	}
	if _, err := statestore.WriteLatestExecutionRef(ctx, s, v); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, meta, err := statestore.ReadLatestExecutionRef(ctx, s)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got != v {
		t.Fatalf("mismatch: got=%+v want=%+v", got, v)
	}
	statefs.AssertJSONFile(t,
		filepath.Join(root, "refs", "latest-execution.json"),
		map[string]any{
			"revisionKey":  v.RevisionKey,
			"executionKey": v.ExecutionKey,
			"executionId":  v.ExecutionID,
			"status":       v.Status,
			"createdAt":    v.CreatedAt.Format(time.RFC3339Nano),
		})

	v2 := v
	v2.Status = "running"
	if _, err := statestore.CASLatestExecutionRef(ctx, s, meta, v2); err != nil {
		t.Fatalf("CAS: %v", err)
	}
	if _, err := statestore.CASLatestExecutionRef(ctx, s, meta, v2); !errors.Is(err, statestore.ErrConflict) {
		t.Fatalf("stale CAS err=%v, want ErrConflict", err)
	}
}

func TestReadLatestExecutionRef_Missing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := filepath.Join(statefs.NewWorkspace(t), ".orun")
	s, _ := statestore.NewLocalStore(statestore.LocalConfig{Root: root})
	_, _, err := statestore.ReadLatestExecutionRef(ctx, s)
	if !errors.Is(err, statestore.ErrNotFound) {
		t.Fatalf("err=%v, want ErrNotFound", err)
	}
}

// --- TriggerRef --------------------------------------------------------------

func TestTriggerRef_LatestAndScopeRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := filepath.Join(statefs.NewWorkspace(t), ".orun")
	s, _ := statestore.NewLocalStore(statestore.LocalConfig{Root: root})

	v := statestore.TriggerRef{
		TriggerName:        "github-pull-request",
		TriggerKey:         "trg-pr139-def456a",
		RevisionKey:        "rev-pr139-def456a-p8f31c09",
		LatestExecutionKey: "run-001",
		HeadRevision:       "def456a1b2c3",
		CreatedAt:          fixedTime(),
	}

	latest := statestore.TriggerRefScope{Name: v.TriggerName, Latest: true}
	if _, err := statestore.WriteTriggerRef(ctx, s, latest, v); err != nil {
		t.Fatalf("write latest: %v", err)
	}
	gotLatest, _, err := statestore.ReadTriggerRef(ctx, s, latest)
	if err != nil {
		t.Fatalf("read latest: %v", err)
	}
	if gotLatest != v {
		t.Fatalf("latest mismatch: %+v", gotLatest)
	}
	statefs.AssertJSONFile(t,
		filepath.Join(root, "refs", "triggers", v.TriggerName, "latest.json"),
		map[string]any{
			"triggerName":        v.TriggerName,
			"triggerKey":         v.TriggerKey,
			"revisionKey":        v.RevisionKey,
			"latestExecutionKey": v.LatestExecutionKey,
			"headRevision":       v.HeadRevision,
			"createdAt":          v.CreatedAt.Format(time.RFC3339Nano),
		})

	scoped := statestore.TriggerRefScope{Name: v.TriggerName, Scope: "pr139"}
	if _, err := statestore.WriteTriggerRef(ctx, s, scoped, v); err != nil {
		t.Fatalf("write scoped: %v", err)
	}
	gotScoped, meta, err := statestore.ReadTriggerRef(ctx, s, scoped)
	if err != nil {
		t.Fatalf("read scoped: %v", err)
	}
	if gotScoped != v {
		t.Fatalf("scoped mismatch: %+v", gotScoped)
	}

	// CAS on scoped happy path.
	v2 := v
	v2.LatestExecutionKey = "run-002"
	if _, err := statestore.CASTriggerRef(ctx, s, scoped, meta, v2); err != nil {
		t.Fatalf("CAS scoped: %v", err)
	}
	// CAS again with stale meta -> conflict.
	if _, err := statestore.CASTriggerRef(ctx, s, scoped, meta, v2); !errors.Is(err, statestore.ErrConflict) {
		t.Fatalf("stale CAS err=%v want ErrConflict", err)
	}
}

func TestTriggerRef_MissingReturnsErrNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := filepath.Join(statefs.NewWorkspace(t), ".orun")
	s, _ := statestore.NewLocalStore(statestore.LocalConfig{Root: root})
	scope := statestore.TriggerRefScope{Name: "github-pull-request", Latest: true}
	_, _, err := statestore.ReadTriggerRef(ctx, s, scope)
	if !errors.Is(err, statestore.ErrNotFound) {
		t.Fatalf("err=%v, want ErrNotFound", err)
	}
}

func TestTriggerRef_InvalidScopeReturnsErrInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := filepath.Join(statefs.NewWorkspace(t), ".orun")
	s, _ := statestore.NewLocalStore(statestore.LocalConfig{Root: root})

	bad := statestore.TriggerRefScope{Name: "bad name!", Latest: true}
	if _, _, err := statestore.ReadTriggerRef(ctx, s, bad); !errors.Is(err, statestore.ErrInvalid) {
		t.Fatalf("read err=%v, want ErrInvalid", err)
	}
	if _, err := statestore.WriteTriggerRef(ctx, s, bad, statestore.TriggerRef{}); !errors.Is(err, statestore.ErrInvalid) {
		t.Fatalf("write err=%v, want ErrInvalid", err)
	}
	if _, err := statestore.CASTriggerRef(ctx, s, bad, statestore.ObjectMeta{}, statestore.TriggerRef{}); !errors.Is(err, statestore.ErrInvalid) {
		t.Fatalf("cas err=%v, want ErrInvalid", err)
	}

	// Non-latest with empty scope should also fail validation.
	emptyScope := statestore.TriggerRefScope{Name: "ok", Latest: false, Scope: ""}
	if _, _, err := statestore.ReadTriggerRef(ctx, s, emptyScope); !errors.Is(err, statestore.ErrInvalid) {
		t.Fatalf("empty scope err=%v, want ErrInvalid", err)
	}
}

// --- NamedRef ----------------------------------------------------------------

func TestNamedRef_RoundTripAndCASAndValidation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := filepath.Join(statefs.NewWorkspace(t), ".orun")
	s, _ := statestore.NewLocalStore(statestore.LocalConfig{Root: root})

	v := statestore.NamedRef{
		Name: "release-candidate", RevisionKey: "rev-1", RevisionID: "id1",
		PlanHash: "h1", CreatedAt: fixedTime(),
	}
	if _, err := statestore.WriteNamedRef(ctx, s, v.Name, v); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, meta, err := statestore.ReadNamedRef(ctx, s, v.Name)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got != v {
		t.Fatalf("mismatch: %+v", got)
	}
	statefs.AssertJSONFile(t,
		filepath.Join(root, "refs", "named", v.Name+".json"),
		map[string]any{
			"name":        v.Name,
			"revisionKey": v.RevisionKey,
			"revisionId":  v.RevisionID,
			"planHash":    v.PlanHash,
			"createdAt":   v.CreatedAt.Format(time.RFC3339Nano),
		})

	v2 := v
	v2.RevisionKey = "rev-2"
	if _, err := statestore.CASNamedRef(ctx, s, v.Name, meta, v2); err != nil {
		t.Fatalf("CAS: %v", err)
	}

	// invalid name → ErrInvalid on every entry point.
	if _, _, err := statestore.ReadNamedRef(ctx, s, "bad name!"); !errors.Is(err, statestore.ErrInvalid) {
		t.Fatalf("read err=%v want ErrInvalid", err)
	}
	if _, err := statestore.WriteNamedRef(ctx, s, "bad/name", statestore.NamedRef{}); !errors.Is(err, statestore.ErrInvalid) {
		t.Fatalf("write err=%v want ErrInvalid", err)
	}
	if _, err := statestore.CASNamedRef(ctx, s, "..", statestore.ObjectMeta{}, statestore.NamedRef{}); !errors.Is(err, statestore.ErrInvalid) {
		t.Fatalf("cas err=%v want ErrInvalid", err)
	}

	// missing → ErrNotFound on read.
	if _, _, err := statestore.ReadNamedRef(ctx, s, "never-written"); !errors.Is(err, statestore.ErrNotFound) {
		t.Fatalf("missing err=%v want ErrNotFound", err)
	}
}

// --- JSON encoder edge cases -------------------------------------------------

func TestRefs_DecodeRejectsUnknownFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := filepath.Join(statefs.NewWorkspace(t), ".orun")
	s, _ := statestore.NewLocalStore(statestore.LocalConfig{Root: root})

	// Plant a doc with an extra field via the raw Write primitive.
	bad := []byte(`{"revisionKey":"x","revisionId":"y","planHash":"z","createdAt":"2026-05-29T00:00:00Z","bogus":1}` + "\n")
	if _, err := s.Write(ctx, statestore.LatestRevisionRefPath(), bad, statestore.WriteOptions{}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	_, _, err := statestore.ReadLatestRevisionRef(ctx, s)
	if !errors.Is(err, statestore.ErrInvalid) {
		t.Fatalf("err=%v, want ErrInvalid", err)
	}
}

func TestRefs_NoStringConcatenationInPaths(t *testing.T) {
	// Sanity guard: every helper's path must come from paths.go. This
	// test scans the source of refs.go/indexes.go for forbidden idioms.
	// The repo-relative path resolves against the package dir at test time.
	for _, name := range []string{"refs.go", "indexes.go"} {
		body := mustReadFile(t, name)
		if strings.Contains(body, `"refs/`) || strings.Contains(body, `"indexes/`) {
			t.Fatalf("%s contains a literal 'refs/' or 'indexes/' string — paths must go through paths.go", name)
		}
	}
}

// itoa avoids dragging strconv into a hot test loop.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}
