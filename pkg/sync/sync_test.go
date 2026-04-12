package sync

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"juarvis/pkg/assets"
)

func setupSyncTest(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	return tmpDir
}

func TestRunSync_NoChanges(t *testing.T) {
	tmpDir := setupSyncTest(t)

	embeddedFS, err := assets.GetEmbeddedFS()
	if err != nil {
		t.Skip("embedded assets not available")
	}

	rootFiles := []string{"AGENTS.md", "permissions.yaml"}
	for _, f := range rootFiles {
		data, err := os.ReadFile(filepath.Join(tmpDir, f))
		if err != nil && !os.IsNotExist(err) {
			t.Fatalf("error reading file: %v", err)
		}
		if err == nil {
			embeddedData, _ := fs.ReadFile(embeddedFS, f)
			if string(data) == string(embeddedData) {
				continue
			}
		}
		os.WriteFile(filepath.Join(tmpDir, f), []byte("test"), 0644)
	}

	err = RunSync(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunSync_UpdateFiles(t *testing.T) {
	tmpDir := setupSyncTest(t)

	ag := "old content"
	os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte(ag), 0644)

	err := RunSync(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("error reading: %v", err)
	}
	if string(data) == ag {
		t.Error("expected file to be updated")
	}
}

func TestRunSync_CreateFiles(t *testing.T) {
	tmpDir := setupSyncTest(t)

	err := RunSync(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files := []string{"AGENTS.md", "permissions.yaml"}
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(tmpDir, f)); os.IsNotExist(err) {
			t.Errorf("expected file %s to be created", f)
		}
	}
}
