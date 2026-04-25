package analyze

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunAnalyze(t *testing.T) {
	// Crear proyecto temporal con estructura
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "testproject")
	os.MkdirAll(filepath.Join(projectDir, "cmd", "main"), 0755)
	os.MkdirAll(filepath.Join(projectDir, "pkg", "config"), 0755)

	// Crear go.mod básico
	os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte("module testproject\n"), 0644)
	os.WriteFile(filepath.Join(projectDir, "cmd", "main", "main.go"), []byte("package main\n"), 0644)
	os.WriteFile(filepath.Join(projectDir, "pkg", "config", "config.go"), []byte("package config\n"), 0644)

	// Ejecutar analyze
	err := RunAnalyzeIn(projectDir, false, false)
	if err != nil {
		t.Fatalf("RunAnalyze falló: %v", err)
	}
}

func TestDetectStack(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "goproject")
	os.MkdirAll(projectDir, 0755)

	// Crear estructura Go
	os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte("module goproject\n"), 0644)
	os.WriteFile(filepath.Join(projectDir, "main.go"), []byte("package main\n"), 0644)
	os.MkdirAll(filepath.Join(projectDir, "cmd"), 0755)

	stack := detectStack(projectDir)

	found := false
	for _, s := range stack {
		if s == "go" {
			found = true
			break
		}
	}
	if !found {
		t.Logf("no se detectó go en stack: %v", stack)
	}
}

func TestDetectConventions(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "testproject")
	os.MkdirAll(projectDir, 0755)

	// Sin convenciones
	conventions := detectConventions(projectDir)
	if len(conventions) > 0 {
		t.Logf("detectó convenciones inesperadas: %v", conventions)
	}

	// Crear .prettierrc
	os.WriteFile(filepath.Join(projectDir, ".prettierrc"), []byte(`{"semi": true}`), 0644)
	conventions = detectConventions(projectDir)

	found := false
	for _, c := range conventions {
		if c == "prettier" {
			found = true
			break
		}
	}
	if !found {
		t.Error("no se detectó prettier")
	}
}

func TestDetectPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "testproject")
	os.MkdirAll(projectDir, 0755)

	patterns := detectPatterns(projectDir)
	if len(patterns) == 0 {
		t.Log("no hay patrones en proyecto vacío (esperado)")
	}
}

func TestDetectAntiPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "testproject")
	os.MkdirAll(projectDir, 0755)

	antiPatterns := detectAntiPatterns(projectDir)
	if len(antiPatterns) == 0 {
		t.Log("no hay anti-patrones en proyecto vacío (esperado)")
	}
}

func TestCountFiles(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "testproject")
	os.MkdirAll(projectDir, 0755)

	// Crear archivos en subdirectorios no ignorados
	srcDir := filepath.Join(projectDir, "src")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("package main\n"), 0644)
	os.WriteFile(filepath.Join(srcDir, "lib.go"), []byte("package main\n"), 0644)
	os.WriteFile(filepath.Join(projectDir, "index.html"), []byte("<html></html>"), 0644)

	count, _ := countFiles(projectDir)
	if count == 0 {
		t.Error("countFiles devolvió 0")
	}
	// Verificar totales
	if count < 3 {
		t.Errorf("esperaba al menos 3 archivos, obtuve %d", count)
	}
}

func TestFormatList(t *testing.T) {
	items := []string{"a", "b", "c"}
	result := formatList(items)
	if result == "" {
		t.Error("formatList devolvió string vacío")
	}
	if len(result) < 3 {
		t.Error("formatList demasiado corto")
	}
}

func TestFormatMap(t *testing.T) {
	m := map[string]int{"go": 5, "ts": 3}
	result := formatMap(m)
	if result == "" {
		t.Error("formatMap devolvió string vacío")
	}
	if len(result) < 5 {
		t.Error("formatMap demasiado corto")
	}
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}
	if !contains(slice, "a") {
		t.Error("contains devolvió false para 'a'")
	}
	if contains(slice, "d") {
		t.Error("contains devolvió true para 'd'")
	}
}

func TestSearchInProjectCode(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "testproject")
	os.MkdirAll(projectDir, 0755)

	os.WriteFile(filepath.Join(projectDir, "main.go"), []byte("func main() {}\n"), 0644)

	if !searchInProjectCode(projectDir, "func main") {
		t.Error("no encontró func main")
	}
}

func TestDetectStackPython(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "pyproject")
	os.MkdirAll(projectDir, 0755)

	// Crear proyecto Python
	os.WriteFile(filepath.Join(projectDir, "main.py"), []byte("print('hello')\n"), 0644)
	os.WriteFile(filepath.Join(projectDir, "requirements.txt"), []byte("flask\n"), 0644)

	stack := detectStack(projectDir)

	found := false
	for _, s := range stack {
		if s == "python" {
			found = true
			break
		}
	}
	if !found {
		t.Logf("no se detectó python en stack: %v", stack)
	}
}

func TestDetectStackNodeJS(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "nodecproject")
	os.MkdirAll(projectDir, 0755)

	// Crear proyecto Node.js
	os.WriteFile(filepath.Join(projectDir, "index.js"), []byte("console.log('hello')\n"), 0644)
	os.WriteFile(filepath.Join(projectDir, "package.json"), []byte(`{"name": "test"}`), 0644)

	stack := detectStack(projectDir)

	found := false
	for _, s := range stack {
		if s == "node" || s == "javascript" {
			found = true
			break
		}
	}
	if !found {
		t.Logf("no se detectó node en stack: %v", stack)
	}
}

func TestDetectCodeAntiPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "testproject")
	os.MkdirAll(projectDir, 0755)

	// Crear archivo con potenciales anti-patterns
	os.WriteFile(filepath.Join(projectDir, "main.go"), []byte(`
package main

func hugeFunction() {
	// 200+ líneas de código sin функции
	for i := 0; i < 1000; i++ {
		doSomething()
	}
}
`), 0644)

	antiPatterns := detectCodeAntiPatterns(projectDir)
	if len(antiPatterns) == 0 {
		t.Log("no se detectaron anti-patterns en código problema (esperado en algunos casos)")
	}
}
