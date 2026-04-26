// Package mode implementa los modos de ejecución de Juarvis para orquestación de agentes.
// Proporciona Planning y Fast modes inspirados en Antigravity AgentKit 2.0.
package mode

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Modo representa el tipo de modo de ejecución.
type Modo int

const (
	ModoUnknown Modo = iota
	ModoFast       // Modo Fast: respuesta inmediata, baja latencia
	ModoPlanning  // Modo Planning: deliberación compleja, decomposición de tareas
	ModoAuto      // Modo Auto: selección automática basada en complejidad
)

// String returns string representation of Modo.
func (m Modo) String() string {
	switch m {
	case ModoFast:
		return "fast"
	case ModoPlanning:
		return "planning"
	case ModoAuto:
		return "auto"
	default:
		return "unknown"
	}
}

// ParseMode parses a string into Modo.
func ParseMode(s string) Modo {
	switch s {
	case "fast":
		return ModoFast
	case "planning":
		return ModoPlanning
	case "auto":
		return ModoAuto
	default:
		return ModoUnknown
	}
}

// TaskComplexity representa la complejidad estimada de una tarea.
type TaskComplexity int

const (
	ComplexityUnknown TaskComplexity = iota
	ComplexitySimple                  // Query simple, fix rápido
	ComplexityMedium                  // Tarea moderada con quelques pasos
	ComplexityComplex                 // Workflow complejo, múltiples agentes
	ComplexityVeryComplex             // Workflow muy complejo con dependencias
)

// String returns string representation of TaskComplexity.
func (c TaskComplexity) String() string {
	switch c {
	case ComplexitySimple:
		return "simple"
	case ComplexityMedium:
		return "medium"
	case ComplexityComplex:
		return "complex"
	case ComplexityVeryComplex:
		return "very_complex"
	default:
		return "unknown"
	}
}

// Task representa una tarea a ejecutar.
type Task struct {
	ID          string                 // ID único de la tarea
	Description string              // Descripción de la tarea
	Input       interface{}        // Input para la tarea
	Complexity  TaskComplexity    // Complejidad estimada
	Steps       []TaskStep        // Pasos de la tarea (para Planning mode)
	Metadata    map[string]any    // Metadatos adicionales
	CreatedAt   time.Time         // Tiempo de creación
}

// TaskStep representa un paso individual en una tarea.
type TaskStep struct {
	ID          string                 // ID único del paso
	Description string              // Descripción del paso
	AgentType   string               // Tipo de agente asignado
	Action     string               // Acción a ejecutar
	DependsOn  []string            // IDs de pasos de los que depende
	Result     any                 // Resultado del paso
	Status     StepStatus         // Estado del paso
	Error      error              // Error si ocurrió
}

// StepStatus representa el estado de un paso.
type StepStatus int

const (
	StepPending StepStatus = iota
	StepRunning
	StepCompleted
	StepFailed
)

// String returns string representation of StepStatus.
func (s StepStatus) String() string {
	switch s {
	case StepPending:
		return "pending"
	case StepRunning:
		return "running"
	case StepCompleted:
		return "completed"
	case StepFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// ModeController administra los modos de ejecución.
type ModeController struct {
	mu           sync.RWMutex
	currentMode  Modo
	previousMode Modo
	metrics     ModeMetrics
	detector    ComplexityDetector
	logger     *ModeLogger
}

// ModeMetrics almacena métricas por modo.
type ModeMetrics struct {
	mu            sync.RWMutex
	Fast         ModeMetric `json:"fast"`
	Planning    ModeMetric `json:"planning"`
	autoSwitches int      `json:"auto_switches"`
}

// ModeMetric almacena métricas para un modo específico.
type ModeMetric struct {
	Executions    int           `json:"executions"`
	TotalLatency time.Duration `json:"total_latency"`
	AvgLatency   time.Duration `json:"avg_latency"`
	MinLatency   time.Duration `json:"min_latency"`
	MaxLatency   time.Duration `json:"max_latency"`
	Errors       int           `json:"errors"`
}

// ModeLogger maneja el logging de decisiones de modo.
type ModeLogger struct {
	mu     sync.RWMutex
	entries []ModeLogEntry
}

// ModeLogEntry representa una entrada de log de modo.
type ModeLogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	TaskID   string   `json:"task_id"`
	Mode    Modo    `json:"mode"`
	Action  string  `json:"action"`
	Details string  `json:"details"`
}

// ComplexityDetector detecta la complejidad de tareas.
type ComplexityDetector struct {
	simplePatterns    []string
	complexPatterns []string
}

// NewModeController crea un nuevo controlador de modo.
func NewModeController() *ModeController {
	return &ModeController{
		currentMode: ModoAuto,
		previousMode: ModoUnknown,
		detector:    NewComplexityDetector(),
		logger:     NewModeLogger(),
		metrics: ModeMetrics{
			Fast: ModeMetric{
				MinLatency: time.Hour,
				MaxLatency: 0,
			},
			Planning: ModeMetric{
				MinLatency: time.Hour,
				MaxLatency: 0,
			},
		},
	}
}

// NewComplexityDetector crea un nuevo detector de complejidad.
func NewComplexityDetector() ComplexityDetector {
	return ComplexityDetector{
		simplePatterns: []string{
			"buscar",
			"mostrar",
			"listar",
			"qué",
			"dame",
			"get",
			"show",
			"list",
			"find",
		},
		complexPatterns: []string{
			"migrar",
			"refactorizar",
			"implementar",
			"múltiples",
			"workflow",
			"pipeline",
			"dependencias",
			"integrar",
			"migrate",
			"refactor",
			"implement",
			"multiple",
			"workflow",
			"pipeline",
			"dependencies",
		},
	}
}

// NewModeLogger crea un nuevo logger de modo.
func NewModeLogger() *ModeLogger {
	return &ModeLogger{
		entries: make([]ModeLogEntry, 0),
	}
}

// CurrentMode returns the current mode.
func (c *ModeController) CurrentMode() Modo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentMode
}

// SetMode establece el modo actual.
func (c *ModeController) SetMode(mode Modo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if mode != c.currentMode {
		c.previousMode = c.currentMode
		c.currentMode = mode
		c.logger.Log(ModeLogEntry{
			Timestamp: time.Now(),
			Mode:     mode,
			Action:   "mode_changed",
			Details:  fmt.Sprintf("changed from %s to %s", c.previousMode, mode),
		})
	}
}

// DetectComplexity detecta la complejidad de una tarea.
func (c *ModeController) DetectComplexity(task *Task) TaskComplexity {
	if task.Complexity != ComplexityUnknown {
		return task.Complexity
	}
	return c.detector.Detect(task.Description)
}

// Detect implementa la detección de complejidad basada en patrones.
func (d *ComplexityDetector) Detect(description string) TaskComplexity {
	simpleCount := 0
	complexCount := 0

	for _, pattern := range d.simplePatterns {
		if containsWord(description, pattern) {
			simpleCount++
		}
	}

	for _, pattern := range d.complexPatterns {
		if containsWord(description, pattern) {
			complexCount++
		}
	}

	// heurística simple basada en longitud y patrones
	wordCount := wordCount(description)

	if wordCount < 5 && simpleCount > 0 {
		return ComplexitySimple
	}
	if wordCount > 20 || complexCount >= 2 {
		return ComplexityComplex
	}
	if wordCount > 10 || complexCount >= 1 {
		return ComplexityMedium
	}

	return ComplexitySimple
}

// SelectMode selecciona el modo basado en la complejidad de la tarea.
func (c *ModeController) SelectMode(task *Task) Modo {
	complexity := c.DetectComplexity(task)

	c.logger.Log(ModeLogEntry{
		Timestamp: time.Now(),
		TaskID:    task.ID,
		Action:   "mode_selection",
		Details:  fmt.Sprintf("detected complexity: %s", complexity),
	})

	switch complexity {
	case ComplexitySimple:
		return ModoFast
	case ComplexityMedium:
		return ModoFast //也可以是 ModoPlanning según preferencia
	case ComplexityComplex, ComplexityVeryComplex:
		return ModoPlanning
	default:
		return ModoFast
	}
}

// ExecuteTask ejecuta una tarea en el modo actual.
func (c *ModeController) ExecuteTask(ctx context.Context, task *Task) (any, error) {
	startTime := time.Now()
	defer func() {
		c.recordLatency(c.CurrentMode(), time.Since(startTime))
	}()

	mode := c.CurrentMode()
	if mode == ModoAuto {
		mode = c.SelectMode(task)
		c.SetMode(mode)
	}

	c.logger.Log(ModeLogEntry{
		Timestamp: time.Now(),
		TaskID:   task.ID,
		Mode:    mode,
		Action:  "execution_start",
		Details: fmt.Sprintf("executing task in %s mode", mode),
	})

	switch mode {
	case ModoFast:
		return c.executeFast(ctx, task)
	case ModoPlanning:
		return c.executePlanning(ctx, task)
	default:
		return c.executeFast(ctx, task)
	}
}

// executeFast ejecuta una tarea en modo Fast.
func (c *ModeController) executeFast(ctx context.Context, task *Task) (any, error) {
	// Simulación de ejecución rápida
	// En implementación real, aquí se ejecutaría la tarea directamente

	c.metrics.mu.Lock()
	c.metrics.Fast.Executions++
	c.metrics.mu.Unlock()

	c.logger.Log(ModeLogEntry{
		Timestamp: time.Now(),
		TaskID:   task.ID,
		Mode:    ModoFast,
		Action:  "execution_complete",
		Details: "task executed in fast mode",
	})

	return map[string]any{
		"mode":   "fast",
		"taskID": task.ID,
		"result": "completed",
	}, nil
}

// executePlanning ejecuta una tarea en modo Planning.
func (c *ModeController) executePlanning(ctx context.Context, task *Task) (any, error) {
	// Primero, decompose la tarea en pasos
	if task.Steps == nil {
		task.Steps = c.decomposeTask(task)
	}

	// Ejecución de pasos en orden
	for i := range task.Steps {
		step := &task.Steps[i]

		// Verificar dependencias
		if !c.dependenciesMet(step, task.Steps) {
			step.Status = StepFailed
			step.Error = fmt.Errorf("dependencies not met for step %s", step.ID)
			c.recordError(ModoPlanning)
			return nil, step.Error
		}

		step.Status = StepRunning
		// Aquí se ejecutaría el paso...

		step.Status = StepCompleted
	}

	c.metrics.mu.Lock()
	c.metrics.Planning.Executions++
	c.metrics.mu.Unlock()

	c.logger.Log(ModeLogEntry{
		Timestamp: time.Now(),
		TaskID:   task.ID,
		Mode:    ModoPlanning,
		Action:  "execution_complete",
		Details: "task executed in planning mode",
	})

	return map[string]any{
		"mode":  "planning",
		"steps": len(task.Steps),
	}, nil
}

// decomposeTask descompone una tarea en pasos.
func (c *ModeController) decomposeTask(task *Task) []TaskStep {
	// Esta es una implementación básica
	// En implementación real, usar parsing NL o IA
	return []TaskStep{
		{
			ID:          fmt.Sprintf("%s-step-1", task.ID),
			Description: "Análisis de la tarea",
			AgentType:   "analyzer",
			Action:      "analyze",
		},
		{
			ID:          fmt.Sprintf("%s-step-2", task.ID),
			Description: "Ejecución de la tarea",
			AgentType:   "executor",
			Action:      "execute",
			DependsOn:   []string{fmt.Sprintf("%s-step-1", task.ID)},
		},
	}
}

// dependenciesMet verifica si las dependencias de un paso están completas.
func (c *ModeController) dependenciesMet(step *TaskStep, steps []TaskStep) bool {
	for _, depID := range step.DependsOn {
		found := false
		for _, s := range steps {
			if s.ID == depID && s.Status == StepCompleted {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// recordLatency registra la latencia de una ejecución.
func (c *ModeController) recordLatency(mode Modo, latency time.Duration) {
	c.metrics.mu.Lock()
	defer c.metrics.mu.Unlock()

	switch mode {
	case ModoFast:
		c.metrics.Fast.TotalLatency += latency
		if c.metrics.Fast.Executions > 0 {
			c.metrics.Fast.AvgLatency = c.metrics.Fast.TotalLatency / time.Duration(c.metrics.Fast.Executions)
		}
		if latency < c.metrics.Fast.MinLatency {
			c.metrics.Fast.MinLatency = latency
		}
		if latency > c.metrics.Fast.MaxLatency {
			c.metrics.Fast.MaxLatency = latency
		}
	case ModoPlanning:
		c.metrics.Planning.TotalLatency += latency
		if c.metrics.Planning.Executions > 0 {
			c.metrics.Planning.AvgLatency = c.metrics.Planning.TotalLatency / time.Duration(c.metrics.Planning.Executions)
		}
		if latency < c.metrics.Planning.MinLatency {
			c.metrics.Planning.MinLatency = latency
		}
		if latency > c.metrics.Planning.MaxLatency {
			c.metrics.Planning.MaxLatency = latency
		}
	}
}

// recordError registra un error.
func (c *ModeController) recordError(mode Modo) {
	c.metrics.mu.Lock()
	defer c.metrics.mu.Unlock()

	switch mode {
	case ModoFast:
		c.metrics.Fast.Errors++
	case ModoPlanning:
		c.metrics.Planning.Errors++
	}
}

// Metrics devuelve las métricas actuales.
func (c *ModeController) Metrics() ModeMetrics {
	c.metrics.mu.RLock()
	defer c.metrics.mu.RUnlock()
	return c.metrics
}

// Log registra una entrada de log.
func (l *ModeLogger) Log(entry ModeLogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, entry)

	// También escribir con log estándar
	log.Printf("[MODE] %s | Task: %s | Mode: %s | Action: %s | Details: %s",
		entry.Timestamp.Format(time.RFC3339),
		entry.TaskID,
		entry.Mode,
		entry.Action,
		entry.Details,
	)
}

// Entries devuelve las entradas de log.
func (l *ModeLogger) Entries() []ModeLogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	entries := make([]ModeLogEntry, len(l.entries))
	copy(entries, l.entries)
	return entries
}

// SwitchMode permite cambiar entre modos durante la ejecución.
func (c *ModeController) SwitchMode(ctx context.Context, newMode Modo) error {
	if newMode == ModoAuto {
		return fmt.Errorf("cannot switch to auto mode directly")
	}

	oldMode := c.CurrentMode()
	if oldMode == newMode {
		return nil
	}

	c.SetMode(newMode)

	log.Printf("[MODE] Switched from %s to %s", oldMode, newMode)
	return nil
}

// GetModeName returns the name of the mode for CLI display.
func (m Modo) GetModeName() string {
	switch m {
	case ModoFast:
		return "Fast"
	case ModoPlanning:
		return "Planning"
	case ModoAuto:
		return "Auto"
	default:
		return "Unknown"
	}
}

// Helper functions
func containsWord(s, word string) bool {
	// Simple word matching
	return len(s) > 0 && len(word) > 0
	// In real implementation, use proper word boundaries
}

func wordCount(s string) int {
	count := 0
	inWord := false
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' {
			inWord = false
		} else if !inWord {
			inWord = true
			count++
		}
	}
	return count
}