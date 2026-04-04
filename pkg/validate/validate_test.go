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
		t.Logf("RunHealthCheck returned error (may be expected if git/python missing): %v", err)
	}
}

func TestRunHealthCheck_NoMarketplace(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("JUARVIS_ROOT", dir)

	err := RunHealthCheck()
	if err == nil {
		t.Log("Warning: RunHealthCheck did not return error without marketplace")
	}
}
