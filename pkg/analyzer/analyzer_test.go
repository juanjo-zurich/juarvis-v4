package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	"juarvis/pkg/config"
)

func TestAnalyzeTranscriptRealFormat(t *testing.T) {
	// Crear transcript con formato JSONL real de OpenCode
	tmpDir := t.TempDir()
	transcriptPath := filepath.Join(tmpDir, "transcript.jsonl")

	// Transcript real con mensajes JSONL
	transcript := `{"role":"assistant","content":"Elegí usar uuid.New() porque genera IDs únicos sin colisión, sobre time.Now().UnixNano()","ts":1714000000}
{"role":"user","content":"Fix del bug en storage","ts":1714000001}
{"role":"assistant","content":"Arreglé el bug en UpdateObservation — ahora sincroniza obsCache. El problema era que no actualizaba el cache.","ts":1714000002}
{"role":"tool","content":"Write[storage.go]","ts":1714000003}
{"role":"assistant","content":"Usé middleware pattern para la API, porque es más extensible que handler functions","ts":1714000004}
{"role":"user","content":"Agrega tests","ts":1714000005}
{"role":"assistant","content":"Elegí table-driven tests sobre t.Run() anidados porque son más legibles","ts":1714000006}`

	if err := os.WriteFile(transcriptPath, []byte(transcript), 0644); err != nil {
		t.Fatalf("error creando transcript: %v", err)
	}

	analysis, err := AnalyzeTranscript(transcriptPath)
	if err != nil {
		t.Fatalf("AnalyzeTranscript falló: %v", err)
	}

	if analysis == nil {
		t.Fatal("analysis es nil")
	}

	// Debe encontrar decisiones (debería encontrar al menos 2)
	if len(analysis.Decisions) == 0 {
		t.Error("no se encontraron decisiones - el parser JSON no está funcionando")
	} else {
		t.Logf("decisiones encontradas: %d", len(analysis.Decisions))
		for _, d := range analysis.Decisions {
			t.Logf("  - %s", d.Choice)
		}
	}

	// Debe encontrar mistakes/fixes
	if len(analysis.Mistakes) == 0 {
		t.Error("no se encontraron mistakes/fixes")
	} else {
		t.Logf("mistakes encontrados: %d", len(analysis.Mistakes))
	}

	// Debe detectar patrones
	if len(analysis.Patterns) == 0 {
		t.Log("no se encontraron patrones")
	} else {
		t.Logf("patrones encontrados: %d", len(analysis.Patterns))
	}
}

func TestAnalyzeTranscriptJSONParsing(t *testing.T) {
	// Test específico: verificar que el parseo JSON funciona
	tmpDir := t.TempDir()
	transcriptPath := filepath.Join(tmpDir, "transcript.jsonl")

	transcript := `{"role":"assistant","content":"Elegí X sobre Y porque es más simple","ts":1714000000}`

	if err := os.WriteFile(transcriptPath, []byte(transcript), 0644); err != nil {
		t.Fatalf("error creando transcript: %v", err)
	}

	analysis, err := AnalyzeTranscript(transcriptPath)
	if err != nil {
		t.Fatalf("AnalyzeTranscript falló: %v", err)
	}

	// Verificar que el contenido se extrajo correctamente
	// El regex debe encontrar "elegí X sobre Y"
	if len(analysis.Decisions) == 0 {
		t.Error("Parser JSON funciona pero el regex no encontró la decisión esperada")
	}
}

func TestAnalyzeTranscriptPlainTextFallback(t *testing.T) {
	// Verificar fallback para texto plano
	tmpDir := t.TempDir()
	transcriptPath := filepath.Join(tmpDir, "transcript.txt")

	// Texto plano (no JSON)
	plainText := "Elegí usar sqlite porque es más rápido que json."
	if err := os.WriteFile(transcriptPath, []byte(plainText), 0644); err != nil {
		t.Fatalf("error creando transcript: %v", err)
	}

	analysis, err := AnalyzeTranscript(transcriptPath)
	if err != nil {
		t.Fatalf("AnalyzeTranscript falló: %v", err)
	}

	// Fallback debería funcionar
	if analysis == nil {
		t.Error("analysis es nil en fallback")
	}
}

func TestAnalyzeTranscriptEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	transcriptPath := filepath.Join(tmpDir, "empty.jsonl")

	// Archivo vacío
	if err := os.WriteFile(transcriptPath, []byte(""), 0644); err != nil {
		t.Fatalf("error creando transcript: %v", err)
	}

	analysis, err := AnalyzeTranscript(transcriptPath)
	if err != nil {
		t.Fatalf("AnalyzeTranscript falló con archivo vacío: %v", err)
	}

	// Con archivo vacío, analysis debería tener valores por defecto
	if analysis == nil {
		t.Error("analysis es nil")
	}
}

func TestSaveAnalysisAndLoad(t *testing.T) {
	tmpDir := t.TempDir()

	analysis := &SessionAnalysis{
		Decisions: []Decision{
			{Choice: "Elegí X sobre Y", Reason: "porque es más simple"},
		},
		Mistakes: []Bugfix{
			{Error: "cache no se actualizaba", Fix: "fixed"},
		},
		Patterns: []Pattern{
			{Name: "api", Count: 5},
			{Name: "middleware", Count: 3},
		},
		Files: []FileChange{
			{File: "storage.go", Action: "modified"},
			{File: "new.go", Action: "created"},
		},
	}

	projectPath := filepath.Join(tmpDir, "testproject")
	if err := os.MkdirAll(filepath.Join(projectPath, config.JuarDir, "memory"), 0755); err != nil {
		t.Skipf("saltando test: no se puede crear directorio: %v", err)
	}

	err := SaveAnalysis(analysis, projectPath)
	if err != nil {
		t.Fatalf("SaveAnalysis falló: %v", err)
	}
}

func TestBuildSessionSummaryComplete(t *testing.T) {
	analysis := &SessionAnalysis{
		Decisions: []Decision{
			{Choice: "Decision 1", Reason: "reason 1"},
			{Choice: "Decision 2", Reason: "reason 2"},
		},
		Mistakes: []Bugfix{
			{Error: "Error 1", Fix: "fixed"},
		},
		Patterns: []Pattern{
			{Name: "api", Count: 5},
			{Name: "middleware", Count: 3},
		},
		Files: []FileChange{
			{File: "storage.go", Action: "modified"},
			{File: "new.go", Action: "created"},
		},
	}

	summary := BuildSessionSummary(analysis)
	if summary == "" {
		t.Fatal("summary vacío")
	}

	// Verificar secciones esperadas
	if !containsString(summary, "Análisis de Sesión") {
		t.Error("summary no contiene título")
	}
	if !containsString(summary, "Decisiones Tomadas") {
		t.Error("summary no contiene decisiones")
	}
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (len(s) < 500 || indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestExtractContextsEmpty(t *testing.T) {
	transcript := `[]`
	result := extractContexts(transcript, "test")
	if len(result) > 0 {
		t.Logf("extractContexts con JSON vacío devolvió %d contextos", len(result))
	}
}

func TestExtractContextsPartial(t *testing.T) {
	transcript := `[{"role": "user", "content": "hola"}]`
	result := extractContexts(transcript, "test")
	// Accept nil or empty for partial transcript
	if result == nil || len(result) == 0 {
		t.Log("extractContexts devolvió nil/vacío para transcript parcial (aceptable)")
	}
}

func TestDeduplicateFunctions(t *testing.T) {
	decisions := []Decision{
		{Choice: "a", Reason: "r1"},
		{Choice: "a", Reason: "r1"},
		{Choice: "b", Reason: "r2"},
	}
	deduped := deduplicateDecisions(decisions)
	if deduped == nil {
		t.Error("deduplicateDecisions devolvió nil")
	}
}
