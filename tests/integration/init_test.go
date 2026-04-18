//go:build integration

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func findJuarvisBinary(t *testing.T) string {
	t.Helper()
	// Busca el binario subiendo dos niveles (de tests/integration/ a la raíz)
	if _, err := os.Stat("../../juarvis"); err == nil {
		abs, _ := filepath.Abs("../../juarvis")
		return abs
	}
	// Intento secundario: buscar en el directorio actual
	if _, err := os.Stat("./juarvis"); err == nil {
		abs, _ := filepath.Abs("./juarvis")
		return abs
	}
	// Intento final: buscar en el PATH
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
		t.Fatalf("init failed: %v\n%s", err, output)
	}

	expectedPaths := []string{
		filepath.Join(tmpDir, ".juar"),
		// .agent/ NO se crea automaticamente
		filepath.Join(tmpDir, ".juar", "skill-registry.md"),
	}

	for _, p := range expectedPaths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("expected path %s to exist", p)
		}
	}
}

func TestInit_ThenCheck(t *testing.T) {
	tmpDir := t.TempDir()
	runJuarvis(t, "init", tmpDir)

	output, err := runJuarvis(t, "--root", tmpDir, "check")
	if err != nil {
		t.Fatalf("check failed: %v\n%s", err, output)
	}
}

func TestInit_ThenLoad(t *testing.T) {
	tmpDir := t.TempDir()
	runJuarvis(t, "init", tmpDir)

	output, err := runJuarvis(t, "--root", tmpDir, "load")
	if err != nil {
		t.Fatalf("load failed: %v\n%s", err, output)
	}
}

