package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func setupValidateTest(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	marketplace := `{"name":"test","plugins":[]}`
	os.WriteFile(filepath.Join(dir, "marketplace.json"), []byte(marketplace), 0644)

	os.MkdirAll(filepath.Join(dir, "plugins"), 0755)
	os.MkdirAll(filepath.Join(dir, ".juar"), 0755)
	os.WriteFile(filepath.Join(dir, ".juar", "skill-registry.md"), []byte("# Skill Registry\n"), 0644)

	t.Setenv("JUARVIS_ROOT", dir)

	return dir
}

func TestRunHealthCheck_Success(t *testing.T) {
	setupValidateTest(t)

	err := RunHealthCheck()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRunHealthCheck_NoRegistry(t *testing.T) {
	dir := t.TempDir()

	// Create valid ecosystem without running load (no registry)
	marketplace := `{"name":"test","plugins":[]}`
	os.WriteFile(filepath.Join(dir, "marketplace.json"), []byte(marketplace), 0644)
	os.MkdirAll(filepath.Join(dir, "plugins"), 0755)
	t.Setenv("JUARVIS_ROOT", dir)

	err := RunHealthCheck()
	// Should pass because embedded marketplace is available as fallback
	if err != nil {
		t.Fatalf("expected no error (embedded marketplace fallback), got: %v", err)
	}
}
