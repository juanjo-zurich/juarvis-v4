package assets

import (
	"os"
	"testing"
)

func TestGetEmbeddedFS(t *testing.T) {
	fs, err := GetEmbeddedFS()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fs == nil {
		t.Fatal("expected fs.FS to not be nil")
	}
}

func TestCopyEmbeddedToDisk(t *testing.T) {
	tmpDir := t.TempDir()

	err := CopyEmbeddedToDisk(".", tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("error reading dir: %v", err)
	}
	if len(entries) == 0 {
		t.Log("no files copied - may be empty embed or no data/")
	}
}

func TestCopyEmbeddedToDisk_NonExistentPath(t *testing.T) {
	tmpDir := t.TempDir()

	err := CopyEmbeddedToDisk("nonexistent/path", tmpDir)
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}
