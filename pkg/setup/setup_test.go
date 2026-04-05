package setup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile_Success(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcFile := filepath.Join(srcDir, "test.txt")
	dstFile := filepath.Join(dstDir, "test.txt")

	os.WriteFile(srcFile, []byte("hello"), 0644)

	err := copyFile(srcFile, dstFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read dst: %v", err)
	}
	if string(content) != "hello" {
		t.Errorf("expected 'hello', got '%s'", string(content))
	}
}

func TestCopyFile_SourceNotFound(t *testing.T) {
	err := copyFile("/nonexistent/file.txt", "/tmp/dest.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent source")
	}
}

func TestCopyFile_SourceIsDirectory(t *testing.T) {
	srcDir := t.TempDir()
	dstFile := filepath.Join(t.TempDir(), "dest.txt")

	err := copyFile(srcDir, dstFile)
	if err == nil {
		t.Fatal("expected error when source is a directory")
	}
}
