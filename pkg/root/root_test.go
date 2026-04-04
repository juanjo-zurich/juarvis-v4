package root

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetRoot_FromEnv_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "marketplace.json"), []byte(`{"name":"test","plugins":[]}`), 0644)
	t.Setenv("JUARVIS_ROOT", tmpDir)

	root, err := GetRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root != tmpDir {
		t.Errorf("expected %s, got %s", tmpDir, root)
	}
}

func TestGetRoot_FromEnv_Invalid(t *testing.T) {
	t.Setenv("JUARVIS_ROOT", "/nonexistent/path")

	_, err := GetRoot()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetRoot_FromEnv_NoMarketplace(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("JUARVIS_ROOT", tmpDir)

	_, err := GetRoot()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetRoot_FromCwd(t *testing.T) {
	tmpDir := t.TempDir()
	resolvedDir, _ := filepath.EvalSymlinks(tmpDir)
	os.WriteFile(filepath.Join(resolvedDir, "marketplace.json"), []byte(`{"name":"test","plugins":[]}`), 0644)
	os.Unsetenv("JUARVIS_ROOT")

	origCwd, _ := os.Getwd()
	os.Chdir(resolvedDir)
	defer os.Chdir(origCwd)

	root, err := GetRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root != resolvedDir {
		t.Errorf("expected %s, got %s", resolvedDir, root)
	}
}

func TestGetRoot_FromSubdir(t *testing.T) {
	tmpDir := t.TempDir()
	resolvedDir, _ := filepath.EvalSymlinks(tmpDir)
	os.WriteFile(filepath.Join(resolvedDir, "marketplace.json"), []byte(`{"name":"test","plugins":[]}`), 0644)
	subDir := filepath.Join(resolvedDir, "plugins", "core")
	os.MkdirAll(subDir, 0755)
	os.Unsetenv("JUARVIS_ROOT")

	origCwd, _ := os.Getwd()
	os.Chdir(subDir)
	defer os.Chdir(origCwd)

	root, err := GetRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root != resolvedDir {
		t.Errorf("expected %s, got %s", resolvedDir, root)
	}
}

func TestGetRoot_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	os.Unsetenv("JUARVIS_ROOT")

	origCwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origCwd)

	_, err := GetRoot()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
