package snapshot

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func setupGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@test.com")
	runGit(t, dir, "config", "user.name", "Test")
	runGit(t, dir, "commit", "--allow-empty", "-m", "initial")

	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %v failed: %v", args, err)
	}
}

func TestCreateSnapshot_Success(t *testing.T) {
	dir := setupGitRepo(t)

	// Crear un archivo para tener cambios
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello"), 0644)
	runGit(t, dir, "add", "test.txt")

	origCwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origCwd)

	err := CreateSnapshot("test-change")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verificar que el stash persiste (porque apply no lo elimina)
	cmd := exec.Command("git", "stash", "list")
	cmd.Dir = dir
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "juarvis-snapshot|") {
		t.Error("stash should persist after apply")
	}
}

func TestCreateSnapshot_NoChanges(t *testing.T) {
	dir := setupGitRepo(t)

	origCwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origCwd)

	// Sin cambios pendientes, debería funcionar o dar warning
	err := CreateSnapshot("no-changes")
	// No importa si hay error o no, solo que no paniquea
	_ = err
}

func TestPruneSnapshots(t *testing.T) {
	dir := setupGitRepo(t)

	origCwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origCwd)

	// Crear archivos y hacer commit para tener algo que stashear
	os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("content1"), 0644)
	runGit(t, dir, "add", "file1.txt")
	runGit(t, dir, "commit", "-m", "add file1")

	os.WriteFile(filepath.Join(dir, "file2.txt"), []byte("content2"), 0644)
	runGit(t, dir, "add", "file2.txt")
	runGit(t, dir, "commit", "-m", "add file2")

	// Crear stashes de juarvis con cambios reales
	os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("modified1"), 0644)
	runGit(t, dir, "stash", "push", "-u", "-m", "juarvis-snapshot|test1")

	os.WriteFile(filepath.Join(dir, "file2.txt"), []byte("modified2"), 0644)
	runGit(t, dir, "stash", "push", "-u", "-m", "juarvis-snapshot|test2")

	// Crear un stash del usuario
	os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("user-change"), 0644)
	runGit(t, dir, "stash", "push", "-u", "-m", "mi-stash-usuario")

	pruned, err := PruneSnapshots(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pruned != 2 {
		t.Errorf("expected 2 pruned, got %d", pruned)
	}

	// Verificar que el stash del usuario sigue ahí
	cmd := exec.Command("git", "stash", "list")
	cmd.Dir = dir
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "mi-stash-usuario") {
		t.Error("user stash was incorrectly pruned")
	}
}
