package artifacts

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ArtifactType representa el tipo de artifact generado por el sistema
type ArtifactType string

const (
	// TaskList representa una lista de tareas con estado
	TaskList ArtifactType = "task_list"
	// ImplementationPlan representa un plan de implementación con pasos
	ImplementationPlan ArtifactType = "implementation_plan"
	// Screenshot representa una captura de pantalla codificada
	Screenshot ArtifactType = "screenshot"
	// TestResult representa resultados de tests
	TestResult ArtifactType = "test_result"
	// VerificationReport representa un reporte de verificación
	VerificationReport ArtifactType = "verification_report"
	// CodeDiff representa un diff de código con contexto
	CodeDiff ArtifactType = "code_diff"
)

// TaskStatus representa el estado de una tarea individual
type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskInProgress TaskStatus = "in_progress"
	TaskCompleted TaskStatus = "completed"
)

// TestStatus representa el resultado de un test
type TestStatus string

const (
	TestPassed  TestStatus = "passed"
	TestFailed  TestStatus = "failed"
	TestSkipped TestStatus = "skipped"
)

// Base es la estructura base de todos los artifacts
type Base struct {
	ID        string        `json:"id"`
	Type     ArtifactType `json:"type"`
	Timestamp time.Time    `json:"timestamp"`
	Tags     []string     `json:"tags,omitempty"`
	Metadata Metadata    `json:"metadata,omitempty"`
}

// Metadata contiene metadatos adicionales del artifact
type Metadata map[string]interface{}

// Task representa una tarea individual en una TaskList
type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      TaskStatus `json:"status"`
	Priority    int        `json:"priority,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TaskListArtifact representa una lista de tareas con estado
type TaskListArtifact struct {
	Base
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Tasks       []Task `json:"tasks"`
}

// ImplementationStep representa un paso en un plan de implementación
type ImplementationStep struct {
	StepNumber int    `json:"step_number"`
	Title     string `json:"title"`
	Description string `json:"description"`
	File      string `json:"file,omitempty"`
	Duration  string `json:"duration,omitempty"`
	Completed bool   `json:"completed"`
	DependsOn []int  `json:"depends_on,omitempty"`
}

// ImplementationPlanArtifact representa un plan de implementación con pasos
type ImplementationPlanArtifact struct {
	Base
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Steps       []ImplementationStep `json:"steps"`
	Goal        string              `json:"goal"`
}

// ScreenshotArtifact representa una captura de pantalla codificada
type ScreenshotArtifact struct {
	Base
	Format   string `json:"format"` // png, jpeg, webp
	Content string `json:"content"` // base64 encoded
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	URL     string `json:"url,omitempty"`
}

// TestCase representa un caso de test individual
type TestCase struct {
	Name      string     `json:"name"`
	Status   TestStatus `json:"status"`
	Duration float64   `json:"duration,omitempty"`
	Message  string    `json:"message,omitempty"`
}

// TestResultArtifact representa resultados de tests
type TestResultArtifact struct {
	Base
	TotalTests   int        `json:"total_tests"`
	Passed      int        `json:"passed"`
	Failed      int        `json:"failed"`
	Skipped     int        `json:"skipped"`
	Coverage    float64    `json:"coverage,omitempty"`
	TestCases   []TestCase  `json:"test_cases,omitempty"`
	Timestamp_ time.Time  `json:"test_timestamp"`
}

// CheckResult representa el resultado de una verificación individual
type CheckResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

// VerificationReportArtifact representa un reporte de verificación
type VerificationReportArtifact struct {
	Base
	Passed     bool          `json:"passed"`
	Total      int           `json:"total"`
	PassedCount int          `json:"passed_count"`
	FailedCount int          `json:"failed_count"`
	Checks    []CheckResult `json:"checks"`
	Duration  string       `json:"duration,omitempty"`
}

// Hunk representa un hunk (bloque) en un diff
type Hunk struct {
	OldStart int    `json:"old_start"`
	OldLines int    `json:"old_lines"`
	NewStart int    `json:"new_start"`
	NewLines int    `json:"new_lines"`
	Content string `json:"content"`
}

// FileDiff representa los cambios en un archivo
type FileDiff struct {
	OldPath string `json:"old_path,omitempty"`
	NewPath string `json:"new_path,omitempty"`
	Mode   string `json:"mode,omitempty"` // added, deleted, modified, renamed
	Hunks  []Hunk `json:"hunks"`
}

// CodeDiffArtifact representa un diff de código con contexto
type CodeDiffArtifact struct {
	Base
	BaseBranch string     `json:"base_branch"`
	HeadBranch string    `json:"head_branch"`
	Files      []FileDiff `json:"files"`
	Stats      DiffStats  `json:"stats"`
}

// DiffStats contiene estadísticas del diff
type DiffStats struct {
	FilesChanged int `json:"files_changed"`
	Insertions  int `json:"insertions"`
	Deletions   int `json:"deletions"`
}

// NewBase crea una nueva base de artifact con ID único
func NewBase(artifactType ArtifactType, tags ...string) Base {
	return Base{
		ID:        uuid.New().String(),
		Type:     artifactType,
		Timestamp: time.Now().UTC(),
		Tags:     tags,
		Metadata: make(Metadata),
	}
}

// NewTaskListArtifact crea un nuevo artifact de tipo TaskList
func NewTaskListArtifact(title, description string, tasks []Task) *TaskListArtifact {
	base := NewBase(TaskList)
	return &TaskListArtifact{
		Base:        base,
		Title:       title,
		Description: description,
		Tasks:       tasks,
	}
}

// NewImplementationPlanArtifact crea un nuevo artifact de tipo ImplementationPlan
func NewImplementationPlanArtifact(title, description, goal string, steps []ImplementationStep) *ImplementationPlanArtifact {
	base := NewBase(ImplementationPlan)
	return &ImplementationPlanArtifact{
		Base:        base,
		Title:       title,
		Description: description,
		Goal:        goal,
		Steps:       steps,
	}
}

// NewScreenshotArtifact crea un nuevo artifact de tipo Screenshot
func NewScreenshotArtifact(format, content string, width, height int) *ScreenshotArtifact {
	base := NewBase(Screenshot)
	return &ScreenshotArtifact{
		Base:    base,
		Format:  format,
		Content: content,
		Width:   width,
		Height:  height,
	}
}

// NewTestResultArtifact crea un nuevo artifact de tipo TestResult
func NewTestResultArtifact(total, passed, failed, skipped int, coverage float64, cases []TestCase) *TestResultArtifact {
	base := NewBase(TestResult)
	return &TestResultArtifact{
		Base:       base,
		TotalTests:  total,
		Passed:     passed,
		Failed:    failed,
		Skipped:   skipped,
		Coverage:  coverage,
		TestCases: cases,
		Timestamp_: time.Now().UTC(),
	}
}

// NewVerificationReportArtifact crea un nuevo artifact de tipo VerificationReport
func NewVerificationReportArtifact(checks []CheckResult, duration string) *VerificationReportArtifact {
	passedCount := 0
	failedCount := 0
	for _, c := range checks {
		if c.Passed {
			passedCount++
		} else {
			failedCount++
		}
	}
	base := NewBase(VerificationReport)
	return &VerificationReportArtifact{
		Base:        base,
		Passed:     failedCount == 0,
		Total:       len(checks),
		PassedCount: passedCount,
		FailedCount: failedCount,
		Checks:     checks,
		Duration:   duration,
	}
}

// NewCodeDiffArtifact crea un nuevo artifact de tipo CodeDiff
func NewCodeDiffArtifact(baseBranch, headBranch string, files []FileDiff, stats DiffStats) *CodeDiffArtifact {
	base := NewBase(CodeDiff)
	return &CodeDiffArtifact{
		Base:       base,
		BaseBranch: baseBranch,
		HeadBranch: headBranch,
		Files:      files,
		Stats:      stats,
	}
}

// GetArtifactTypeString devuelve el tipo como string
func (a ArtifactType) String() string {
	return string(a)
}

// ParseArtifactType parsea un string a ArtifactType
func ParseArtifactType(s string) (ArtifactType, error) {
	switch ArtifactType(s) {
	case TaskList:
		return TaskList, nil
	case ImplementationPlan:
		return ImplementationPlan, nil
	case Screenshot:
		return Screenshot, nil
	case TestResult:
		return TestResult, nil
	case VerificationReport:
		return VerificationReport, nil
	case CodeDiff:
		return CodeDiff, nil
	default:
		return "", fmt.Errorf("tipo de artifact desconocido: %s", s)
	}
}