package discovery

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed in %s: %v\n%s", dir, err, out)
	}
}

func TestFindIntentFile_InCurrentDir(t *testing.T) {
	root := t.TempDir()
	initGitRepo(t, root)
	writeFile(t, filepath.Join(root, "intent.yaml"), "kind: Intent")

	path, dir, err := FindIntentFile(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(path) != "intent.yaml" {
		t.Errorf("expected intent.yaml, got %s", path)
	}
	absRoot, _ := filepath.Abs(root)
	if resolved, evalErr := filepath.EvalSymlinks(absRoot); evalErr == nil {
		absRoot = resolved
	}
	if dir != absRoot {
		t.Errorf("expected dir %s, got %s", absRoot, dir)
	}
}

func TestFindIntentFile_WalksUp(t *testing.T) {
	root := t.TempDir()
	initGitRepo(t, root)
	writeFile(t, filepath.Join(root, "intent.yaml"), "kind: Intent")

	subdir := filepath.Join(root, "services", "api", "src")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	path, _, err := FindIntentFile(subdir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(path) != "intent.yaml" {
		t.Errorf("expected intent.yaml, got %s", path)
	}
}

func TestFindIntentFile_StopsAtGitRoot(t *testing.T) {
	// Create a parent dir with intent.yaml but NO git repo
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "intent.yaml"), "kind: Intent")

	// Create a sub-directory that IS a git repo (its own root)
	subRepo := filepath.Join(root, "sub")
	if err := os.MkdirAll(subRepo, 0755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, subRepo)

	// Verify sub is actually its own git root
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = subRepo
	out, err := cmd.Output()
	if err != nil {
		t.Skipf("could not verify git root in sub dir: %v", err)
	}
	t.Logf("sub repo git root: %s", string(out))

	subdir := filepath.Join(subRepo, "deep", "nested")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	// FindIntentFile from deep/nested should stop at subRepo (the git root)
	// and NOT find the intent.yaml in root (which is above the git boundary)
	_, _, err = FindIntentFile(subdir)
	if err == nil {
		t.Fatal("expected error when intent.yaml is above git root")
	}
}

func TestFindIntentFile_NotFound(t *testing.T) {
	root := t.TempDir()
	initGitRepo(t, root)

	_, _, err := FindIntentFile(root)
	if err == nil {
		t.Fatal("expected error when no intent.yaml exists")
	}
}

func TestFindIntentFile_YmlVariant(t *testing.T) {
	root := t.TempDir()
	initGitRepo(t, root)
	writeFile(t, filepath.Join(root, "intent.yml"), "kind: Intent")

	path, _, err := FindIntentFile(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(path) != "intent.yml" {
		t.Errorf("expected intent.yml, got %s", path)
	}
}

func TestFindIntentFile_PrefersYamlOverYml(t *testing.T) {
	root := t.TempDir()
	initGitRepo(t, root)
	writeFile(t, filepath.Join(root, "intent.yaml"), "kind: Intent")
	writeFile(t, filepath.Join(root, "intent.yml"), "kind: Intent")

	path, _, err := FindIntentFile(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(path) != "intent.yaml" {
		t.Errorf("expected intent.yaml to be preferred, got %s", path)
	}
}

const sampleComponentYAML = `apiVersion: sourceplane.io/v1
kind: Component

metadata:
  name: web-app

spec:
  type: helm
`

func TestFindComponentFile_InCurrentDir(t *testing.T) {
	root := t.TempDir()
	initGitRepo(t, root)
	writeFile(t, filepath.Join(root, "component.yaml"), sampleComponentYAML)

	name, filePath, err := FindComponentFile(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "web-app" {
		t.Errorf("expected 'web-app', got %q", name)
	}
	if filepath.Base(filePath) != "component.yaml" {
		t.Errorf("expected component.yaml path, got %s", filePath)
	}
}

func TestFindComponentFile_WalksUp(t *testing.T) {
	root := t.TempDir()
	initGitRepo(t, root)

	compDir := filepath.Join(root, "services", "api")
	writeFile(t, filepath.Join(compDir, "component.yaml"), sampleComponentYAML)

	srcDir := filepath.Join(compDir, "src", "handlers")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	name, _, err := FindComponentFile(srcDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "web-app" {
		t.Errorf("expected 'web-app', got %q", name)
	}
}

func TestFindComponentFile_NotFound(t *testing.T) {
	root := t.TempDir()
	initGitRepo(t, root)

	subdir := filepath.Join(root, "services", "api")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	name, filePath, err := FindComponentFile(subdir)
	if err != nil {
		t.Fatalf("expected nil error when no component.yaml found, got: %v", err)
	}
	if name != "" {
		t.Errorf("expected empty name, got %q", name)
	}
	if filePath != "" {
		t.Errorf("expected empty path, got %q", filePath)
	}
}

func TestFindComponentFile_StopsAtGitRoot(t *testing.T) {
	// component.yaml lives above the git root — should NOT be found
	outer := t.TempDir()
	writeFile(t, filepath.Join(outer, "component.yaml"), sampleComponentYAML)

	gitRoot := filepath.Join(outer, "repo")
	if err := os.MkdirAll(gitRoot, 0755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, gitRoot)

	subdir := filepath.Join(gitRoot, "src")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	name, _, err := FindComponentFile(subdir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "" {
		t.Errorf("expected empty name (component.yaml is above git root), got %q", name)
	}
}

func TestFindComponentFile_YmlVariant(t *testing.T) {
	root := t.TempDir()
	initGitRepo(t, root)
	writeFile(t, filepath.Join(root, "component.yml"), sampleComponentYAML)

	name, filePath, err := FindComponentFile(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "web-app" {
		t.Errorf("expected 'web-app', got %q", name)
	}
	if filepath.Base(filePath) != "component.yml" {
		t.Errorf("expected component.yml, got %s", filePath)
	}
}

func TestExtractMetadataName(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard component.yaml",
			input:    sampleComponentYAML,
			expected: "web-app",
		},
		{
			name: "quoted name",
			input: `metadata:
  name: "my-component"
`,
			expected: "my-component",
		},
		{
			name: "single-quoted name",
			input: `metadata:
  name: 'another-comp'
`,
			expected: "another-comp",
		},
		{
			name:     "no metadata section",
			input:    "kind: Component\nspec:\n  type: helm\n",
			expected: "",
		},
		{
			name:     "empty content",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractMetadataName(tc.input)
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}
