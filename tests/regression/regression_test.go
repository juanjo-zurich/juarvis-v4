//go:build regression

package regression

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func findJuarvisBinary(t *testing.T) string {
	t.Helper()
	// Try project root relative to test dir (tests/regression -> ../../juarvis)
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

func loadGolden(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("golden", name))
	if err != nil {
		t.Fatalf("golden file %s not found: %v", name, err)
	}
	return string(data)
}

func TestHelp_Output(t *testing.T) {
	output, err := runJuarvis(t, "--help")
	if err != nil {
		t.Fatalf("juarvis --help failed: %v", err)
	}

	golden := loadGolden(t, "help_output.golden")
	if !strings.Contains(output, golden) {
		t.Errorf("help output does not contain expected text.\nExpected: %q\nGot: %q", golden, output)
	}
}

func TestVersion_Output(t *testing.T) {
	output, err := runJuarvis(t, "--version")
	if err != nil {
		t.Fatalf("juarvis --version failed: %v", err)
	}

	golden := loadGolden(t, "version_output.golden")
	if !strings.HasPrefix(output, golden) {
		t.Errorf("version output does not start with expected text.\nExpected prefix: %q\nGot: %q", golden, output)
	}
}

func TestVerify_Runs(t *testing.T) {
	output, _ := runJuarvis(t, "verify")
	if len(output) == 0 {
		t.Fatal("juarvis verify produced no output")
	}
	if !strings.Contains(output, "go build") && !strings.Contains(output, "embedded JSON") {
		t.Errorf("expected verify output to contain verification results, got:\n%s", output)
	}
}

func TestInit_Output(t *testing.T) {
	tmpDir := t.TempDir()
	output, err := runJuarvis(t, "init", tmpDir)
	if err != nil {
		t.Fatalf("juarvis init failed: %v\n%s", err, output)
	}

	if !strings.Contains(output, "Ecosistema Juarvis inicializado") {
		t.Errorf("expected init success message, got:\n%s", output)
	}
}
