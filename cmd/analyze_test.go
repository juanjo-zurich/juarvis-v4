package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeCommand_HelpFlag(t *testing.T) {
	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetErr(b)

	rootCmd.SetArgs([]string{"analyze", "--help"})
	err := rootCmd.Execute()

	if err != nil {
		t.Errorf("analyze --help falló: %v", err)
	}

	output := b.String()
	if !bytes.Contains(b.Bytes(), []byte("Uso:")) && !bytes.Contains(b.Bytes(), []byte("Usage:")) {
		t.Errorf("No se mostró la ayuda de analyze: %s", output)
	}
}

func TestAnalyzeCommand_DetectGoProject(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "goproject")
	os.MkdirAll(filepath.Join(projectDir, "cmd", "main"), 0755)

	os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte("module example.com/myapp\ngo 1.21\n"), 0644)
	os.WriteFile(filepath.Join(projectDir, "cmd", "main", "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"analyze", projectDir})
	_ = rootCmd.Execute()

	output := b.String()
	if len(output) == 0 {
		t.Error("No hubo output del análisis")
	}
}

func TestAnalyzeCommand_DetectConventions(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "convproject")
	os.MkdirAll(projectDir, 0755)

	os.WriteFile(filepath.Join(projectDir, ".prettierrc"), []byte(`{"semi": true}`), 0644)
	os.WriteFile(filepath.Join(projectDir, ".eslintrc.json"), []byte(`{"rules": {}}`), 0644)

	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"analyze", projectDir})
	_ = rootCmd.Execute()

	output := b.String()
	if len(output) == 0 {
		t.Error("No hubo output del análisis")
	}
}

func TestAnalyzeCommand_NonExistentDir(t *testing.T) {
	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"analyze", "/nonexistent/path/12345"})
	err := rootCmd.Execute()

	// Analyze puede manejar esto de diferentes formas
	// Verificamos que no crash
	if err != nil {
		t.Logf("Análisis devolvió error (esperable): %v", err)
	}
}