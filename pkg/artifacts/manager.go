package artifacts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"juarvis/pkg/output"
)

// Manager gestiona el almacenamiento y recuperación de artifacts
type Manager struct {
	rootDir string
}

// NewManager crea un nuevo gestor de artifacts
func NewManager(rootDir string) *Manager {
	return &Manager{
		rootDir: rootDir,
	}
}

// ensureDir asegura que el directorio de artifacts existe
func (m *Manager) ensureDir() error {
	artifactsDir := filepath.Join(m.rootDir, ".juarvis", "artifacts")
	if err := os.MkdirAll(artifactsDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio de artifacts: %w", err)
	}
	return nil
}

// artifactPath devuelve la ruta del archivo de un artifact
func (m *Manager) artifactPath(id string) string {
	return filepath.Join(m.rootDir, ".juarvis", "artifacts", id+".json")
}

// metadataPath devuelve la ruta del archivo de metadatos
func (m *Manager) metadataPath(id string) string {
	return filepath.Join(m.rootDir, ".juarvis", "artifacts", id+".meta.json")
}

// SaveTaskList guarda un TaskListArtifact
func (m *Manager) SaveTaskList(artifact *TaskListArtifact) error {
	if err := m.ensureDir(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializando artifact: %w", err)
	}
	path := m.artifactPath(artifact.ID)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error guardando artifact: %w", err)
	}
	if err := m.saveMetadata(artifact.Base); err != nil {
		return err
	}
	output.Debug("Artifact guardado: %s (%s)", artifact.ID, artifact.Type)
	return nil
}

// SaveImplementationPlan guarda un ImplementationPlanArtifact
func (m *Manager) SaveImplementationPlan(artifact *ImplementationPlanArtifact) error {
	if err := m.ensureDir(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializando artifact: %w", err)
	}
	path := m.artifactPath(artifact.ID)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error guardando artifact: %w", err)
	}
	if err := m.saveMetadata(artifact.Base); err != nil {
		return err
	}
	output.Debug("Artifact guardado: %s (%s)", artifact.ID, artifact.Type)
	return nil
}

// SaveScreenshot guarda un ScreenshotArtifact
func (m *Manager) SaveScreenshot(artifact *ScreenshotArtifact) error {
	if err := m.ensureDir(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializando artifact: %w", err)
	}
	path := m.artifactPath(artifact.ID)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error guardando artifact: %w", err)
	}
	if err := m.saveMetadata(artifact.Base); err != nil {
		return err
	}
	output.Debug("Artifact guardado: %s (%s)", artifact.ID, artifact.Type)
	return nil
}

// SaveTestResult guarda un TestResultArtifact
func (m *Manager) SaveTestResult(artifact *TestResultArtifact) error {
	if err := m.ensureDir(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializando artifact: %w", err)
	}
	path := m.artifactPath(artifact.ID)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error guardando artifact: %w", err)
	}
	if err := m.saveMetadata(artifact.Base); err != nil {
		return err
	}
	output.Debug("Artifact guardado: %s (%s)", artifact.ID, artifact.Type)
	return nil
}

// SaveVerificationReport guarda un VerificationReportArtifact
func (m *Manager) SaveVerificationReport(artifact *VerificationReportArtifact) error {
	if err := m.ensureDir(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializando artifact: %w", err)
	}
	path := m.artifactPath(artifact.ID)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error guardando artifact: %w", err)
	}
	if err := m.saveMetadata(artifact.Base); err != nil {
		return err
	}
	output.Debug("Artifact guardado: %s (%s)", artifact.ID, artifact.Type)
	return nil
}

// SaveCodeDiff guarda un CodeDiffArtifact
func (m *Manager) SaveCodeDiff(artifact *CodeDiffArtifact) error {
	if err := m.ensureDir(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializando artifact: %w", err)
	}
	path := m.artifactPath(artifact.ID)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error guardando artifact: %w", err)
	}
	if err := m.saveMetadata(artifact.Base); err != nil {
		return err
	}
	output.Debug("Artifact guardado: %s (%s)", artifact.ID, artifact.Type)
	return nil
}

// saveMetadata guarda metadatos del artifact
func (m *Manager) saveMetadata(base Base) error {
	meta := struct {
		ID        string        `json:"id"`
		Type      ArtifactType `json:"type"`
		Timestamp time.Time    `json:"timestamp"`
		Tags     []string     `json:"tags"`
	}{
		ID:        base.ID,
		Type:      base.Type,
		Timestamp: base.Timestamp,
		Tags:     base.Tags,
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializando metadatos: %w", err)
	}
	path := m.metadataPath(base.ID)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error guardando metadatos: %w", err)
	}
	return nil
}

// Get obtiene un artifact por ID
func (m *Manager) Get(id string) (interface{}, error) {
	path := m.artifactPath(id)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("artifact no encontrado: %s", id)
	}
	// Detectar tipo parseando primero la base
	var base Base
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, fmt.Errorf("error parseando artifact: %w", err)
	}
	switch base.Type {
	case TaskList:
		var a TaskListArtifact
		if err := json.Unmarshal(data, &a); err != nil {
			return nil, err
		}
		return &a, nil
	case ImplementationPlan:
		var a ImplementationPlanArtifact
		if err := json.Unmarshal(data, &a); err != nil {
			return nil, err
		}
		return &a, nil
	case Screenshot:
		var a ScreenshotArtifact
		if err := json.Unmarshal(data, &a); err != nil {
			return nil, err
		}
		return &a, nil
	case TestResult:
		var a TestResultArtifact
		if err := json.Unmarshal(data, &a); err != nil {
			return nil, err
		}
		return &a, nil
	case VerificationReport:
		var a VerificationReportArtifact
		if err := json.Unmarshal(data, &a); err != nil {
			return nil, err
		}
		return &a, nil
	case CodeDiff:
		var a CodeDiffArtifact
		if err := json.Unmarshal(data, &a); err != nil {
			return nil, err
		}
		return &a, nil
	default:
		return nil, fmt.Errorf("tipo de artifact desconocido: %s", base.Type)
	}
}

// Delete elimina un artifact por ID
func (m *Manager) Delete(id string) error {
	artifactsPath := m.artifactPath(id)
	metadataPath := m.metadataPath(id)
	if err := os.Remove(artifactsPath); err != nil {
		return fmt.Errorf("error eliminando artifact: %w", err)
	}
	if err := os.Remove(metadataPath); err != nil {
		output.Debug("Metadatos no encontrados para: %s", id)
	}
	output.Debug("Artifact eliminado: %s", id)
	return nil
}

// ListArtifactInfo representa información resumida de un artifact
type ListArtifactInfo struct {
	ID        string        `json:"id"`
	Type     ArtifactType `json:"type"`
	Timestamp time.Time    `json:"timestamp"`
	Tags     []string     `json:"tags"`
	Summary  string       `json:"summary"`
}

// List lista todos los artifacts, opcionalmente filtrados por tipo
func (m *Manager) List(filterType ArtifactType) ([]ListArtifactInfo, error) {
	artifactsDir := filepath.Join(m.rootDir, ".juarvis", "artifacts")
	entries, err := os.ReadDir(artifactsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []ListArtifactInfo{}, nil
		}
		return nil, fmt.Errorf("error leyendo directorio: %w", err)
	}
	var artifacts []ListArtifactInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".meta.json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(artifactsDir, entry.Name()))
		if err != nil {
			continue
		}
		var base Base
		if err := json.Unmarshal(data, &base); err != nil {
			continue
		}
		if filterType != "" && base.Type != filterType {
			continue
		}
		summary := m.generateSummary(&base)
		artifacts = append(artifacts, ListArtifactInfo{
			ID:        base.ID,
			Type:     base.Type,
			Timestamp: base.Timestamp,
			Tags:     base.Tags,
			Summary:  summary,
		})
	}
	// Ordenar por timestamp descendente
	sort.Slice(artifacts, func(i, j int) bool {
		return artifacts[i].Timestamp.After(artifacts[j].Timestamp)
	})
	return artifacts, nil
}

// generateSummary genera un resumen del contents del artifact
func (m *Manager) generateSummary(base *Base) string {
	data, err := os.ReadFile(m.artifactPath(base.ID))
	if err != nil {
		return ""
	}
	switch base.Type {
	case TaskList:
		var a TaskListArtifact
		if err := json.Unmarshal(data, &a); err != nil {
			return ""
		}
		completed := 0
		for _, t := range a.Tasks {
			if t.Status == TaskCompleted {
				completed++
			}
		}
		return fmt.Sprintf("%s (%d/%d tareas)", a.Title, completed, len(a.Tasks))
	case ImplementationPlan:
		var a ImplementationPlanArtifact
		if err := json.Unmarshal(data, &a); err != nil {
			return ""
		}
		completed := 0
		for _, s := range a.Steps {
			if s.Completed {
				completed++
			}
		}
		return fmt.Sprintf("%s (%d/%d pasos)", a.Title, completed, len(a.Steps))
	case Screenshot:
		var a ScreenshotArtifact
		if err := json.Unmarshal(data, &a); err != nil {
			return ""
		}
		return fmt.Sprintf("%dx%d %s", a.Width, a.Height, a.Format)
	case TestResult:
		var a TestResultArtifact
		if err := json.Unmarshal(data, &a); err != nil {
			return ""
		}
		return fmt.Sprintf("%d/%d passed (%.1f%%)", a.Passed, a.TotalTests, float64(a.Passed)/float64(a.TotalTests)*100)
	case VerificationReport:
		var a VerificationReportArtifact
		if err := json.Unmarshal(data, &a); err != nil {
			return ""
		}
		status := "PASSED"
		if !a.Passed {
			status = "FAILED"
		}
		return fmt.Sprintf("%s (%d/%d checks)", status, a.PassedCount, a.Total)
	case CodeDiff:
		var a CodeDiffArtifact
		if err := json.Unmarshal(data, &a); err != nil {
			return ""
		}
		return fmt.Sprintf("%s → %s (%d archivos)", a.BaseBranch, a.HeadBranch, len(a.Files))
	default:
		return ""
	}
}

// FindByTag busca artifacts por tag
func (m *Manager) FindByTag(tag string) ([]ListArtifactInfo, error) {
	all, err := m.List("")
	if err != nil {
		return nil, err
	}
	var results []ListArtifactInfo
	for _, a := range all {
		for _, t := range a.Tags {
			if t == tag {
				results = append(results, a)
				break
			}
		}
	}
	return results, nil
}

// Count devuelve el número total de artifacts
func (m *Manager) Count() (int, error) {
	artifactsDir := filepath.Join(m.rootDir, ".juarvis", "artifacts")
	entries, err := os.ReadDir(artifactsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".json") && !strings.HasSuffix(entry.Name(), ".meta.json") {
			count++
		}
	}
	return count, nil
}

// GetArtifactsDir devuelve el directorio de artifacts
func (m *Manager) GetArtifactsDir() string {
	return filepath.Join(m.rootDir, ".juarvis", "artifacts")
}