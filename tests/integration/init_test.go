//go:build integration

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func findJuarvisBinary(t *testing.T) string {
	t.Helper()
	if _, err := os.Stat("./juarvis"); err == nil {
		abs, _ := filepath.Abs("./juarvis")
		return abs
	}
	path, err := exec.LookPath("juarvis")
	if err == nil {
		return path
	}
	t.Fatal("juarvis binary not found — run 'go build -o juarvis .' first")
	return ""
}

func runJuarvis(t *testing.T, args ...string) (string, error) {
	t.Helper()
	bin := findJuarvisBinary(t)
	cmd := exec.Command(bin, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func TestInit_CreatesEcosystem(t *testing.T) {
	tmpDir := t.TempDir()

	output, err := runJuarvis(t, "init", tmpDir)
	if err != nil {
		t.Fatalf("juarvis init failed: %v\n%s", err, output)
	}

	expectedFiles := []string{
		"marketplace.json",
		"AGENTS.md",
		"opencode.json",
		"permissions.yaml",
	}
	for _, f := range expectedFiles {
		if _, err := os.Stat(filepath.Join(tmpDir, f)); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", f)
		}
	}

	expectedDirs := []string{"plugins", ".juar", "skills"}
	for _, d := range expectedDirs {
		if _, err := os.Stat(filepath.Join(tmpDir, d)); os.IsNotExist(err) {
			t.Errorf("expected directory %s to exist", d)
		}
	}
}

func TestInit_ThenCheck(t *testing.T) {
	tmpDir := t.TempDir()

	output, err := runJuarvis(t, "init", tmpDir)
	if err != nil {
		t.Fatalf("juarvis init failed: %v\n%s", err, output)
	}

	output, err = runJuarvis(t, "--root", tmpDir, "check")
	if err != nil {
		t.Fatalf("juarvis check failed: %v\n%s", err, output)
	}

	if !strings.Contains(output, "git") {
		t.Error("expected check output to mention git")
	}
}

func TestInit_ThenLoad(t *testing.T) {
	tmpDir := t.TempDir()

	output, err := runJuarvis(t, "init", tmpDir)
	if err != nil {
		t.Fatalf("juarvis init failed: %v\n%s", err, output)
	}

	output, err = runJuarvis(t, "--root", tmpDir, "load")
	if err != nil {
		t.Fatalf("juarvis load failed: %v\n%s", err, output)
	}

	registryPath := filepath.Join(tmpDir, ".juar", "skill-registry.md")
	data, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatalf("skill-registry.md not found: %v", err)
	}
	if len(data) < 100 {
		t.Errorf("skill-registry.md seems too small (%d bytes)", len(data))
	}
}
