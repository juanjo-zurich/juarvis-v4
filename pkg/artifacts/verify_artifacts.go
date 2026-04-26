package artifacts

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"juarvis/pkg/verify"
)

// SaveVerificationFromVerify guarda automáticamente los resultados de verify como un artifact
func SaveVerificationFromVerify(rootDir string, results []verify.CheckResult, duration string) (*VerificationReportArtifact, error) {
	// Convertir los resultados de verify al formato de artifact
	checks := make([]CheckResult, len(results))
	for i, r := range results {
		checks[i] = CheckResult{
			Name:    r.Name,
			Passed:  r.Passed,
			Message: r.Message,
		}
	}
	artifact := NewVerificationReportArtifact(checks, duration)
	artifact.Tags = []string{"verify", "automatic"}
	manager := NewManager(rootDir)
	if err := manager.SaveVerificationReport(artifact); err != nil {
		return nil, fmt.Errorf("error guardando verification report: %w", err)
	}
	return artifact, nil
}

// ConvertVerifyResults convierte resultados de verify.CheckResult a artifacts.CheckResult
func ConvertVerifyResults(results []verify.CheckResult) []CheckResult {
	checks := make([]CheckResult, len(results))
	for i, r := range results {
		checks[i] = CheckResult{
			Name:    r.Name,
			Passed:  r.Passed,
			Message: r.Message,
		}
	}
	return checks
}

// GenerateCodeDiffFromGit.genera un CodeDiffArtifact a partir de git diff
func GenerateCodeDiffFromGit(rootDir, baseBranch, headBranch string) (*CodeDiffArtifact, error) {
	// Obtener el diff
	cmd := exec.Command("git", "diff", baseBranch+"..."+headBranch, "--stat")
	_, err := cmd.CombinedOutput()
	if err != nil {
		// Intentar con diff normal
		cmd = exec.Command("git", "diff", baseBranch, headBranch, "--stat")
		_, err = cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("error obteniendo diff: %w", err)
		}
	}
	// Obtener diff detallado
	cmd = exec.Command("git", "diff", baseBranch, headBranch)
	diffOutput, err := cmd.CombinedOutput()
	if err != nil {
		cmd = exec.Command("git", "diff", baseBranch+"..."+headBranch)
		diffOutput, err = cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("error obteniendo diff detallado: %w", err)
		}
	}
	// Parsear el diff
	files := parseGitDiff(string(diffOutput))
	stats := DiffStats{
		FilesChanged: len(files),
	}
	for _, f := range files {
		for _, h := range f.Hunks {
			stats.Insertions += h.NewLines
			stats.Deletions += h.OldLines
		}
	}
	artifact := NewCodeDiffArtifact(baseBranch, headBranch, files, stats)
	manager := NewManager(rootDir)
	if err := manager.SaveCodeDiff(artifact); err != nil {
		return nil, fmt.Errorf("error guardando code diff: %w", err)
	}
	return artifact, nil
}

// parseGitDiff parsea la salida de git diff
func parseGitDiff(output string) []FileDiff {
	var files []FileDiff
	lines := strings.Split(output, "\n")
	var currentFile *FileDiff
	var currentHunk *Hunk
	oldStart := 0
	oldLines := 0
	newStart := 0
	newLines := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Nueva sección de archivo
		if strings.HasPrefix(line, "diff --git") {
			if currentFile != nil && currentHunk != nil {
				currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
			}
			if currentFile != nil {
				files = append(files, *currentFile)
			}
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				oldPath := strings.TrimPrefix(parts[2], "a/")
				newPath := strings.TrimPrefix(parts[3], "b/")
				mode := "modified"
				if strings.HasPrefix(line, "diff --git a/") && !strings.Contains(line, " b/") {
					mode = "added"
				}
				currentFile = &FileDiff{
					OldPath: oldPath,
					NewPath: newPath,
					Mode:   mode,
				}
			}
			currentHunk = nil
			continue
		}
		// Línea de hunk
		if strings.HasPrefix(line, "@@") {
			if currentHunk != nil && currentFile != nil {
				currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
			}
			// Parsear posición del hunk
			// @@ -oldStart,oldLines +newStart,newLines @@
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				rangePart := strings.Trim(parts[1], "@")
				oldRange := strings.Split(rangePart, " ")[0]
				newRange := strings.Split(rangePart, " ")[1]
				fmt.Sscanf(oldRange, "%d,%d", &oldStart, &oldLines)
				fmt.Sscanf(newRange, "%d,%d", &newStart, &newLines)
				currentHunk = &Hunk{
					OldStart: oldStart,
					OldLines: oldLines,
					NewStart: newStart,
					NewLines: newLines,
				}
			}
			continue
		}
		// Agregar contenido al hunk
		if currentHunk != nil && (strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, " ")) {
			currentHunk.Content += line + "\n"
		}
	}

	// Cerrar últimas estructuras
	if currentHunk != nil && currentFile != nil {
		currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
	}
	if currentFile != nil {
		files = append(files, *currentFile)
	}

	return files
}

// GenerateTestResultFromGoTest genera un TestResultArtifact a partir de go test
func GenerateTestResultFromGoTest(rootDir string) (*TestResultArtifact, error) {
	cmd := exec.Command("go", "test", "./...", "-json", "-cover", "-timeout", "5m")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error ejecutando tests: %w", err)
	}
	// Parsear resultados JSON
	var testCases []TestCase
	passed := 0
	failed := 0
	skipped := 0

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var testEvent struct {
			Action  string  `json:"Action"`
			Test   string  `json:"Test"`
			Output string  `json:"Output,omitempty"`
			Time   string  `json:"Time,omitempty"`
		}
		if err := json.Unmarshal([]byte(line), &testEvent); err != nil {
			continue
		}
		switch testEvent.Action {
		case "pass":
			passed++
			testCases = append(testCases, TestCase{
				Name:     testEvent.Test,
				Status:   TestPassed,
				Duration: 0,
			})
		case "fail":
			failed++
			testCases = append(testCases, TestCase{
				Name:    testEvent.Test,
				Status:  TestFailed,
				Message: testEvent.Output,
			})
		case "skip":
			skipped++
			testCases = append(testCases, TestCase{
				Name:    testEvent.Test,
				Status:  TestSkipped,
				Message: testEvent.Output,
			})
		case "output":
			// Ignorar output
		}
	}

	total := passed + failed + skipped
	coverage := 0.0
	// Intentar obtener coverage
	coverCmd := exec.Command("go", "test", "./...", "-cover")
	coverOutput, _ := coverCmd.CombinedOutput()
	for _, line := range strings.Split(string(coverOutput), "\n") {
		if strings.Contains(line, "coverage:") {
			var cov float64
			fmt.Sscanf(strings.ReplaceAll(line, "%", ""), "coverage: %f", &cov)
			coverage = cov
			break
		}
	}

	artifact := NewTestResultArtifact(total, passed, failed, skipped, coverage, testCases)
	artifact.Tags = []string{"test", "automatic"}
	manager := NewManager(rootDir)
	if err := manager.SaveTestResult(artifact); err != nil {
		return nil, fmt.Errorf("error guardando test result: %w", err)
	}
	return artifact, nil
}

// GetArtifactsDirectory asegura que existe el directorio de artifacts y lo devuelve
func GetArtifactsDirectory(rootDir string) (string, error) {
	artifactsDir := filepath.Join(rootDir, ".juarvis", "artifacts")
	if err := os.MkdirAll(artifactsDir, 0755); err != nil {
		return "", fmt.Errorf("error creando directorio de artifacts: %w", err)
	}
	return artifactsDir, nil
}

// ListByType lista todos los artifacts de un tipo específico
func ListByType(rootDir string, artifactType ArtifactType) ([]ListArtifactInfo, error) {
	manager := NewManager(rootDir)
	return manager.List(artifactType)
}

// ListAll lista todos los artifacts
func ListAll(rootDir string) ([]ListArtifactInfo, error) {
	manager := NewManager(rootDir)
	return manager.List("")
}

// GetArtifact obtiene un artifact por ID
func GetArtifact(rootDir, id string) (interface{}, error) {
	manager := NewManager(rootDir)
	return manager.Get(id)
}

// DeleteArtifact elimina un artifact por ID
func DeleteArtifact(rootDir, id string) error {
	manager := NewManager(rootDir)
	return manager.Delete(id)
}

// GetArtifactCount devuelve el número de artifacts
func GetArtifactCount(rootDir string) (int, error) {
	manager := NewManager(rootDir)
	return manager.Count()
}

// GetLatestArtifacts devuelve los últimos N artifacts
func GetLatestArtifacts(rootDir string, n int) ([]ListArtifactInfo, error) {
	manager := NewManager(rootDir)
	all, err := manager.List("")
	if err != nil {
		return nil, err
	}
	if n > len(all) {
		n = len(all)
	}
	return all[:n], nil
}

// ArtifactSummary representa un resumen de un artifact
type ArtifactSummary struct {
	ID        string    `json:"id"`
	Type     string    `json:"type"`
	Summary  string    `json:"summary"`
	Timestamp time.Time `json:"timestamp"`
	Tags     []string  `json:"tags"`
}

// GetSummary obtiene un resumen de un artifact por ID
func GetSummary(rootDir, id string) (*ArtifactSummary, error) {
	manager := NewManager(rootDir)
	info, err := manager.List("")
	if err != nil {
		return nil, err
	}
	for _, a := range info {
		if a.ID == id {
			return &ArtifactSummary{
				ID:        a.ID,
				Type:     string(a.Type),
				Summary:  a.Summary,
				Timestamp: a.Timestamp,
				Tags:     a.Tags,
			}, nil
		}
	}
	return nil, fmt.Errorf("artifact no encontrado: %s", id)
}

// AutoSaveVerification guarda automáticamente un VerificationReport después de verify
func AutoSaveVerification(rootDir string) error {
	// Ejecutar verify
	opts := verify.VerifyOptions{}
	results, err := verify.RunVerify(opts)
	if err != nil {
		return fmt.Errorf("error ejecutando verify: %w", err)
	}
	// Guardar como artifact
	_, err = SaveVerificationFromVerify(rootDir, results, "auto")
	return err
}