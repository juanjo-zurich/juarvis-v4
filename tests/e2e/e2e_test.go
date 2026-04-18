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
	if _, err := os.Stat("../../juarvis"); err == nil {
		abs, _ := filepath.Abs("../../juarvis")
		return abs
	}
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

	// 1. Inicializar Juarvis
	output, err := runJuarvis(t, "init", tmpDir)
	if err != nil {
		t.Fatalf("init failed: %v\n%s", err, output)
	}

	// 2. Configurar Git y primer commit
	runGit(t, tmpDir, "init")
	runGit(t, tmpDir, "config", "user.email", "test@test.com")
	runGit(t, tmpDir, "config", "user.name", "Test")
	
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("version inicial"), 0644)
	
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "initial")

	// --- CAMBIO CLAVE AQUÍ ---
	// 3. Modificar un archivo ANTES de crear el snapshot.
	// Si el repositorio está "limpio", git stash no guarda nada.
	os.WriteFile(testFile, []byte("cambio para el snapshot"), 0644)

	// 4. Crear el snapshot (ahora sí tiene algo que guardar)
	output, err = runJuarvis(t, "--root", tmpDir, "snapshot", "create", "before-change")
	if err != nil {
		t.Fatalf("snapshot create failed: %v\n%s", err, output)
	}

	// 5. Hacer otro cambio destructivo
	os.WriteFile(testFile, []byte("cambio accidental"), 0644)

	// 6. Restaurar
	output, err = runJuarvis(t, "--root", tmpDir, "snapshot", "restore")
	if err != nil && !strings.Contains(output, "conflictos") && !strings.Contains(output, "conflicts") {
		t.Fatalf("snapshot restore failed unexpectedly: %v\n%s", err, output)
	}

	// 7. Verificar que el contenido volvió al estado del snapshot
	content, _ := os.ReadFile(testFile)
	if string(content) != "cambio para el snapshot" {
		t.Errorf("la restauración no devolvió el contenido esperado. Obtuve: %s", string(content))
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

