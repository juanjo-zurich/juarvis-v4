package initpkg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunInit_CreatesStructure(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "test-project")

	err := RunInit(target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files := []string{
		"marketplace.json",
		"AGENTS.md",
		"opencode.json",
		"permissions.yaml",
	}
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(target, f)); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", f)
		}
	}

	dirs := []string{"plugins", ".juar", "skills"}
	for _, d := range dirs {
		if _, err := os.Stat(filepath.Join(target, d)); os.IsNotExist(err) {
			t.Errorf("expected directory %s to exist", d)
		}
	}
}

func TestRunInit_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "marketplace.json"), []byte(`{"name":"test","plugins":[]}`), 0644)

	err := RunInit(tmpDir)
	if err == nil {
		t.Fatal("expected error for existing ecosystem, got nil")
	}
}
