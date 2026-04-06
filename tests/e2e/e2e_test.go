//go:build e2e

package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func findJuarvisBinary(t *testing.T) string {
	t.Helper()
	// Try project root relative to test dir (tests/e2e -> ../../juarvis)
	if _, err := os.Stat("../../juarvis"); err == nil {
		abs, _ := filepath.Abs("../../juarvis")
		return abs
	}
	// Try current directory
	if _, err := os.Stat("./juarvis"); err == nil {
		abs, _ := filepath.Abs("./juarvis")
		return abs
	}
	// Try PATH
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

func TestE2E_FullUserFlow(t *testing.T) {
	tmpDir := t.TempDir()

	output, err := runJuarvis(t, "init", tmpDir)
	if err != nil {
		t.Fatalf("init failed: %v\n%s", err, output)
	}

	output, err = runJuarvis(t, "--root", tmpDir, "check")
	if err != nil {
		t.Fatalf("check failed: %v\n%s", err, output)
	}

	output, err = runJuarvis(t, "--root", tmpDir, "load")
	if err != nil {
		t.Fatalf("load failed: %v\n%s", err, output)
	}

	registryPath := filepath.Join(tmpDir, ".juar", "skill-registry.md")
	data, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatalf("skill-registry.md not found: %v", err)
	}
	if len(data) < 100 {
		t.Errorf("skill-registry.md too small (%d bytes)", len(data))
	}
}

func TestE2E_SnapshotFlow(t *testing.T) {
	tmpDir := t.TempDir()

	output, err := runJuarvis(t, "init", tmpDir)
	if err != nil {
		t.Fatalf("init failed: %v\n%s", err, output)
	}

	runGit(t, tmpDir, "init")
	runGit(t, tmpDir, "config", "user.email", "test@test.com")
	runGit(t, tmpDir, "config", "user.name", "Test")
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "initial")

	output, err = runJuarvis(t, "--root", tmpDir, "snapshot", "create", "before-change")
	if err != nil {
		t.Fatalf("snapshot create failed: %v\n%s", err, output)
	}

	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("hello"), 0644)

	output, err = runJuarvis(t, "--root", tmpDir, "snapshot", "restore")
	if err != nil && !strings.Contains(output, "conflictos") && !strings.Contains(output, "conflicts") {
		t.Fatalf("snapshot restore failed unexpectedly: %v\n%s", err, output)
	}

	if !strings.Contains(output, "snapshot") && !strings.Contains(output, "stash") && !strings.Contains(output, "restaur") {
		t.Errorf("expected snapshot restore output, got:\n%s", output)
	}
}

func TestE2E_SkillCreateFlow(t *testing.T) {
	tmpDir := t.TempDir()

	output, err := runJuarvis(t, "init", tmpDir)
	if err != nil {
		t.Fatalf("init failed: %v\n%s", err, output)
	}

	output, err = runJuarvis(t, "--root", tmpDir, "skill", "create", "my-test-skill")
	if err != nil {
		t.Fatalf("skill create failed: %v\n%s", err, output)
	}

	skillPath := filepath.Join(tmpDir, ".agent", "skills", "my-test-skill", "SKILL.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		t.Errorf("expected skill %s to exist", skillPath)
	}
}

func TestE2E_OutsideEcosystem(t *testing.T) {
	tmpDir := t.TempDir()

	output, err := runJuarvis(t, "--root", tmpDir)
	if err != nil {
		t.Fatalf("juarvis outside ecosystem failed: %v\n%s", err, output)
	}

	if !strings.Contains(output, "No se detecto un ecosistema") {
		t.Errorf("expected 'no ecosystem' message, got:\n%s", output)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}
