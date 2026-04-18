//go:build e2e

package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// findJuarvisBinary intenta localizar el binario compilado de Juarvis
func findJuarvisBinary(t *testing.T) string {
	t.Helper()
	// Buscar en la raíz del proyecto desde tests/e2e
	if _, err := os.Stat("../../juarvis"); err == nil {
		abs, _ := filepath.Abs("../../juarvis")
		return abs
	}
	// Buscar en el directorio actual
	if _, err := os.Stat("./juarvis"); err == nil {
		abs, _ := filepath.Abs("./juarvis")
		return abs
	}
	// Buscar en el PATH del sistema
	path, err := exec.LookPath("juarvis")
	if err == nil {
		return path
	}
	t.Fatal("juarvis binary not found — run 'go build -o juarvis .' first")
	return ""
}

// runJuarvis ejecuta el comando juarvis con argumentos
func runJuarvis(t *testing.T, args ...string) (string, error) {
	t.Helper()
	bin := findJuarvisBinary(t)
	cmd := exec.Command(bin, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// runGit es un helper para ejecutar comandos de git en un directorio específico
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
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

	// 1. Inicializar Juarvis en el directorio temporal
	output, err := runJuarvis(t, "init", tmpDir)
	if err != nil {
		t.Fatalf("init failed: %v\n%s", err, output)
	}

	// 2. Configurar Git y crear un estado base limpio
	runGit(t, tmpDir, "init")
	runGit(t, tmpDir, "config", "user.email", "test@test.com")
	runGit(t, tmpDir, "config", "user.name", "Test Runner")
	
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("contenido original\n"), 0644)
	runGit(t, tmpDir, "add", "test.txt")
	runGit(t, tmpDir, "commit", "-m", "initial commit")

	// 3. REALIZAR UN CAMBIO Y AGREGARLO AL INDEX
	// Esto es crucial para que git stash (usado por juarvis) no ignore el cambio
	os.WriteFile(testFile, []byte("contenido modificado para snapshot\n"), 0644)
	runGit(t, tmpDir, "add", "test.txt") 

	// 4. Crear el snapshot
	output, err = runJuarvis(t, "--root", tmpDir, "snapshot", "create", "fix-test")
	if err != nil {
		t.Fatalf("snapshot create failed: %v\n%s", err, output)
	}

	// 5. Simular un desastre o cambio accidental
	os.WriteFile(testFile, []byte("esto no deberia estar aqui\n"), 0644)

	// 6. Restaurar el snapshot
	output, err = runJuarvis(t, "--root", tmpDir, "snapshot", "restore")
	if err != nil {
		t.Fatalf("snapshot restore failed: %v\n%s", err, output)
	}

	// 7. Verificar que el contenido es el que guardamos en el snapshot
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("no se pudo leer el archivo tras restaurar: %v", err)
	}
	
	expected := "contenido modificado para snapshot\n"
	if string(content) != expected {
		t.Errorf("La restauración falló.\nEsperado: %q\nObtenido: %q", expected, string(content))
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
