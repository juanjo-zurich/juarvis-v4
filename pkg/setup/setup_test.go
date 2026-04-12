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

func TestRunSetup_ValidIDE(t *testing.T) {
	tmpRoot := t.TempDir()
	t.Setenv("JUARVIS_ROOT", tmpRoot)

	os.MkdirAll(filepath.Join(tmpRoot, "skills"), 0755)
	os.WriteFile(filepath.Join(tmpRoot, "AGENTS.md"), []byte("# Test agents"), 0644)
	os.WriteFile(filepath.Join(tmpRoot, "permissions.yaml"), []byte("permissions: test"), 0644)
	os.WriteFile(filepath.Join(tmpRoot, "opencode.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(tmpRoot, "marketplace.json"), []byte("{}"), 0644)

	err := RunSetupCore([]string{"opencode"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunSetup_AllIDEs(t *testing.T) {
	tmpRoot := t.TempDir()
	t.Setenv("JUARVIS_ROOT", tmpRoot)

	os.MkdirAll(filepath.Join(tmpRoot, "skills"), 0755)
	os.WriteFile(filepath.Join(tmpRoot, "AGENTS.md"), []byte("# Test agents"), 0644)
	os.WriteFile(filepath.Join(tmpRoot, "permissions.yaml"), []byte("permissions: test"), 0644)
	os.WriteFile(filepath.Join(tmpRoot, "opencode.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(tmpRoot, "marketplace.json"), []byte("{}"), 0644)

	err := RunSetupCore([]string{"opencode", "windsurf", "vscode"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunSetup_InvalidIDE(t *testing.T) {
	tmpRoot := t.TempDir()
	t.Setenv("JUARVIS_ROOT", tmpRoot)

	os.MkdirAll(filepath.Join(tmpRoot, "skills"), 0755)
	os.WriteFile(filepath.Join(tmpRoot, "AGENTS.md"), []byte("# Test agents"), 0644)
	os.WriteFile(filepath.Join(tmpRoot, "permissions.yaml"), []byte("permissions: test"), 0644)
	os.WriteFile(filepath.Join(tmpRoot, "marketplace.json"), []byte("{}"), 0644)

	err := RunSetupCore([]string{"nonexistent-ide"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunSetup_DirectoriesExist(t *testing.T) {
	tmpRoot := t.TempDir()
	t.Setenv("JUARVIS_ROOT", tmpRoot)

	opencodeDir := filepath.Join(os.Getenv("HOME"), ".config", "opencode")
	os.MkdirAll(opencodeDir, 0755)

	os.MkdirAll(filepath.Join(tmpRoot, "skills"), 0755)
	os.WriteFile(filepath.Join(tmpRoot, "AGENTS.md"), []byte("# Test agents"), 0644)
	os.WriteFile(filepath.Join(tmpRoot, "permissions.yaml"), []byte("permissions: test"), 0644)
	os.WriteFile(filepath.Join(tmpRoot, "opencode.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(tmpRoot, "marketplace.json"), []byte("{}"), 0644)

	err := RunSetupCore([]string{"opencode"})
	if err != nil {
		t.Fatalf("unexpected error when dirs exist: %v", err)
	}
}
