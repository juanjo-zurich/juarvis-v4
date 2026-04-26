// Package mode implementa el modo Planning para tareas complejas.
package mode

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Planner ejecuta tareas en modo Planning con deliberación y descomposición.
type Planner struct {
	mu           sync.RWMutex
	taskDecomposer TaskDecomposer
	agentAssigner AgentAssigner
	executor     TaskExecutor
	timeout     time.Duration
	maxSteps    int
	deps        Dependencies
}

// TaskDecomposer descompone tareas en pasos ejecutables.
type TaskDecomposer struct {
	mu         sync.RWMutex
	strategies []DecompositionStrategy
}

// DecompositionStrategy define una estrategia de descomposición.
type DecompositionStrategy interface {
	Decompose(task *Task) []TaskStep
	Name() string
}

// AgentAssigner asigna agentes a pasos específicos.
type AgentAssigner struct {
	mu          sync.RWMutex
	agentTypes  map[string]AgentTypeConfig
	assignments []AgentAssignment
}

// AgentTypeConfig define configuración para un tipo de agente.
type AgentTypeConfig struct {
	Type          string
	MaxRetries    int
	Timeout      time.Duration
	Capabilities []string
}

// AgentAssignment representa la asignación de un agente a un paso.
type AgentAssignment struct {
	StepID   string
	AgentID  string
	AgentType string
	Status   AssignmentStatus
}

// AssignmentStatus representa el estado de una asignación.
type AssignmentStatus int

const (
	AssignmentPending AssignmentStatus = iota
	AssignmentAssigned
	AssignmentRunning
	AssignmentCompleted
	AssignmentFailed
)

// TaskExecutor ejecuta pasos de tarea.
type TaskExecutor struct {
	mu          sync.RWMutex
	executors   map[string]Executor
	results     map[string]any
	errors      map[string]error
}

// Executor define un ejecutor para un tipo de agente.
type Executor interface {
	Execute(ctx context.Context, step *TaskStep) (any, error)
	Type() string
}

// Dependencies maneja las dependencias entre pasos.
type Dependencies struct {
	mu          sync.RWMutex
	graph       map[string][]string
	completed  map[string]bool
	pending    map[string][]string
}

// NewPlanner crea un nuevo Planner.
func NewPlanner() *Planner {
	return &Planner{
		taskDecomposer: NewTaskDecomposer(),
		agentAssigner:  NewAgentAssigner(),
		executor:      NewTaskExecutor(),
		timeout:      5 * time.Minute,
		maxSteps:      50,
		deps:         NewDependencies(),
	}
}

// NewTaskDecomposer crea un nuevo descomponedor de tareas.
func NewTaskDecomposer() TaskDecomposer {
	return TaskDecomposer{
		strategies: []DecompositionStrategy{
			&SequentialDecomposition{},
			&ParallelDecomposition{},
			&ConditionalDecomposition{},
		},
	}
}

// NewAgentAssigner crea un nuevo asignador de agentes.
func NewAgentAssigner() AgentAssigner {
	return AgentAssigner{
		agentTypes: map[string]AgentTypeConfig{
			"analyzer": {
				Type:          "analyzer",
				MaxRetries:    2,
				Timeout:      30 * time.Second,
				Capabilities: []string{"analyze", "understand", "extract"},
			},
			"executor": {
				Type:          "executor",
				MaxRetries:    3,
				Timeout:      2 * time.Minute,
				Capabilities: []string{"execute", "run", "perform"},
			},
			"verifier": {
				Type:          "verifier",
				MaxRetries:    3,
				Timeout:      1 * time.Minute,
				Capabilities: []string{"verify", "check", "validate"},
			},
			"reporter": {
				Type:          "reporter",
				MaxRetries:    1,
				Timeout:      30 * time.Second,
				Capabilities: []string{"report", "summarize", "format"},
			},
		},
		assignments: make([]AgentAssignment, 0),
	}
}

// NewTaskExecutor crea un nuevo ejecutor de tareas.
func NewTaskExecutor() TaskExecutor {
	return TaskExecutor{
		executors: make(map[string]Executor),
		results:  make(map[string]any),
		errors:   make(map[string]error),
	}
}

// NewDependencies crea un nuevo manejador de dependencias.
func NewDependencies() Dependencies {
	return Dependencies{
		graph:      make(map[string][]string),
		completed:  make(map[string]bool),
		pending:   make(map[string][]string),
	}
}

// Plan crea un plan de ejecución para una tarea.
func (p *Planner) Plan(ctx context.Context, task *Task) (*Plan, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}

	// Step 1: Descomponer la tarea en pasos
	steps, err := p.taskDecomposer.Decompose(task)
	if err != nil {
		return nil, fmt.Errorf("failed to decompose task: %w", err)
	}

	if len(steps) == 0 {
		return nil, fmt.Errorf("no steps generated for task")
	}

	if len(steps) > p.maxSteps {
		return nil, fmt.Errorf("task decomposition resulted in too many steps (%d > %d)", len(steps), p.maxSteps)
	}

	// Step 2: Construir grafo de dependencias
	p.deps.Build(steps)

	// Step 3: Identificar dependencias entre pasos
	dependencies := p.deps.Identify(steps)

	// Step 4: Asignar agentes a pasos
	assignments, err := p.agentAssigner.Assign(ctx, steps)
	if err != nil {
		return nil, fmt.Errorf("failed to assign agents: %w", err)
	}

	// Step 5: Calcular orden de ejecución (topological sort)
	executionOrder, err := p.deps.TopologicalSort()
	if err != nil {
		return nil, fmt.Errorf("failed to compute execution order: %w", err)
	}

	return &Plan{
		TaskID:         task.ID,
		Steps:          steps,
		Dependencies:   dependencies,
		Assignments:   assignments,
		ExecutionOrder: executionOrder,
		Status:         PlanPending,
		CreatedAt:      time.Now(),
	}, nil
}

// Execute ejecuta un plan.
func (p *Planner) Execute(ctx context.Context, plan *Plan) (*PlanResult, error) {
	if plan == nil {
		return nil, fmt.Errorf("plan cannot be nil")
	}

	result := &PlanResult{
		PlanID:    plan.TaskID,
		Status:    PlanStatusRunning,
		StartTime: time.Now(),
		Steps:     make([]StepResult, 0, len(plan.Steps)),
	}

	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Execute steps in order
	for _, stepID := range plan.ExecutionOrder {
		step := p.findStep(plan.Steps, stepID)
		if step == nil {
			result.Status = PlanStatusFailed
			result.Error = fmt.Errorf("step not found: %s", stepID)
			return result, result.Error
		}

		// Check dependencies
		if !p.deps.Completed(stepID) {
			result.Status = PlanStatusFailed
			result.Error = fmt.Errorf("dependencies not met for step: %s", stepID)
			return result, result.Error
		}

		// Execute step
		stepResult, err := p.executeStep(ctx, plan, step)
		result.Steps = append(result.Steps, stepResult)

		if err != nil {
			result.Status = PlanStatusFailed
			result.Error = err
			result.EndTime = time.Now()
			return result, err
		}

		// Mark as completed
		p.deps.MarkCompleted(stepID)
	}

	result.Status = PlanStatusCompleted
	result.EndTime = time.Now()
	return result, nil
}

// executeStep ejecuta un paso individual.
func (p *Planner) executeStep(ctx context.Context, plan *Plan, step *TaskStep) (StepResult, error) {
	result := StepResult{
		StepID:   step.ID,
		Status:   StepResultRunning,
		StartTime: time.Now(),
	}

	// Find assigned agent
	assignment := p.findAssignment(plan.Assignments, step.ID)
	if assignment == nil {
		result.Status = StepResultFailed
		result.Error = fmt.Errorf("no agent assigned to step: %s", step.ID)
		return result, result.Error
	}

	// Execute with executor
	execResult, err := p.executor.Execute(ctx, step)
	result.Result = execResult
	result.EndTime = time.Now()

	if err != nil {
		result.Status = StepResultFailed
		result.Error = err
		return result, err
	}

	result.Status = StepResultCompleted
	return result, nil
}

// decompose implementa la descomposición de tareas.
func (d *TaskDecomposer) Decompose(task *Task) ([]TaskStep, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Usar estrategia según complejidad
	var strategy DecompositionStrategy
	switch {
	case task.Complexity >= ComplexityComplex:
		strategy = &ConditionalDecomposition{}
	case len(task.Steps) > 5:
		strategy = &ParallelDecomposition{}
	default:
		strategy = &SequentialDecomposition{}
	}

	return strategy.Decompose(task), nil
}

// DecomposeTask descompone una tarea en pasos usando el descomponedor interno.
func (p *Planner) DecomposeTask(task *Task) ([]TaskStep, error) {
	return p.taskDecomposer.Decompose(task)
}

// Assign asigna agentes a pasos.
func (a *AgentAssigner) Assign(ctx context.Context, steps []TaskStep) ([]AgentAssignment, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	assignments := make([]AgentAssignment, 0, len(steps))

	for i := range steps {
		agentType := a.selectAgentType(&steps[i])
		assignment := AgentAssignment{
			StepID:    steps[i].ID,
			AgentType: agentType,
			Status:    AssignmentAssigned,
		}
		assignments = append(assignments, assignment)
	}

	a.assignments = append(a.assignments, assignments...)
	return assignments, nil
}

// selectAgentType selecciona el tipo de agente apropiado para un paso.
func (a *AgentAssigner) selectAgentType(step *TaskStep) string {
	switch step.AgentType {
	case "analyzer", "analizar":
		return "analyzer"
	case "executor", "ejecutar":
		return "executor"
	case "verifier", "verificar":
		return "verifier"
	case "reporter", "reportar":
		return "reporter"
	default:
		// Default to executor
		return "executor"
	}
}

// findStep encuentra un paso por ID.
func (p *Planner) findStep(steps []TaskStep, id string) *TaskStep {
	for i := range steps {
		if steps[i].ID == id {
			return &steps[i]
		}
	}
	return nil
}

// findAssignment encuentra una asignación por ID de paso.
func (p *Planner) findAssignment(assignments []AgentAssignment, stepID string) *AgentAssignment {
	for i := range assignments {
		if assignments[i].StepID == stepID {
			return &assignments[i]
		}
	}
	return nil
}

// Build construye el grafo de dependencias.
func (d *Dependencies) Build(steps []TaskStep) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.graph = make(map[string][]string)
	d.pending = make(map[string][]string)

	for _, step := range steps {
		d.graph[step.ID] = step.DependsOn
		if len(step.DependsOn) == 0 {
			d.pending[step.ID] = nil // No dependencies
		} else {
			d.pending[step.ID] = step.DependsOn
		}
	}
}

// Identify devuelve las dependencias entre pasos.
func (d *Dependencies) Identify(steps []TaskStep) map[string][]string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	deps := make(map[string][]string)
	for _, step := range steps {
		deps[step.ID] = step.DependsOn
	}
	return deps
}

// TopologicalSort devuelve el orden topológico de ejecución.
func (d *Dependencies) TopologicalSort() ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// copia del grafo
	graph := make(map[string][]string)
	for k, v := range d.graph {
		graph[k] = v
	}

	// Indegree map
	inDegree := make(map[string]int)
	for node := range graph {
		inDegree[node] = 0
	}
	for _, deps := range graph {
		for _, dep := range deps {
			inDegree[dep]++
		}
	}

	// Queue for nodes with no dependencies
	queue := make([]string, 0)
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	result := make([]string, 0)
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		// Remove edges from this node
		for nextNode, deps := range graph {
			newDeps := make([]string, 0)
			for _, dep := range deps {
				if dep != node {
					newDeps = append(newDeps, dep)
				}
			}
			graph[nextNode] = newDeps
			if len(newDeps) == 0 {
				queue = append(queue, nextNode)
			}
		}
	}

	if len(result) != len(inDegree) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return result, nil
}

// Completed verifica si un paso está completado.
func (d *Dependencies) Completed(id string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.completed[id]
}

// MarkCompleted marca un paso como completado.
func (d *Dependencies) MarkCompleted(id string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.completed[id] = true
}

// Execute implementa la ejecución de pasos.
func (e *TaskExecutor) Execute(ctx context.Context, step *TaskStep) (any, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Simulación de ejecución
	// En implementación real, usar el executor registrado

	e.results[step.ID] = map[string]any{
		"stepID":  step.ID,
		"status": "completed",
	}

	return e.results[step.ID], nil
}

// RegisterExecutor registra un ejecutor para un tipo de agente.
func (e *TaskExecutor) RegisterExecutor(exec Executor) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.executors[exec.Type()] = exec
}

// Plan representa un plan de ejecución.
type Plan struct {
	TaskID         string
	Steps          []TaskStep
	Dependencies   map[string][]string
	Assignments    []AgentAssignment
	ExecutionOrder []string
	Status        PlanStatus
	CreatedAt     time.Time
}

// PlanStatus representa el estado de un plan.
type PlanStatus int

const (
	PlanPending PlanStatus = iota
	PlanReady
	PlanStatusRunning
	PlanStatusCompleted
	PlanStatusFailed
)

// String returns string representation of PlanStatus.
func (s PlanStatus) String() string {
	switch s {
	case PlanPending:
		return "pending"
	case PlanReady:
		return "ready"
	case PlanStatusRunning:
		return "running"
	case PlanStatusCompleted:
		return "completed"
	case PlanStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// PlanResult representa el resultado de ejecutar un plan.
type PlanResult struct {
	PlanID    string
	Status    PlanStatus
	Error     error
	Steps     []StepResult
	StartTime time.Time
	EndTime   time.Time
}

// StepResult representa el resultado de un paso.
type StepResult struct {
	StepID    string
	Status   StepResultStatus
	Result   any
	Error    error
	StartTime time.Time
	EndTime  time.Time
}

// StepResultStatus representa el estado de un resultado de paso.
type StepResultStatus int

const (
	StepResultPending StepResultStatus = iota
	StepResultRunning
	StepResultCompleted
	StepResultFailed
)

// SequentialDecomposition descompone tareas secuencialmente.
type SequentialDecomposition struct{}

// Decompose descompone una tarea en pasos secuenciales.
func (d *SequentialDecomposition) Decompose(task *Task) []TaskStep {
	steps := []TaskStep{
		{
			ID:          fmt.Sprintf("%s-analyze", task.ID),
			Description: "Analizar la tarea: " + task.Description,
			AgentType:   "analyzer",
			Action:      "analyze",
		},
		{
			ID:          fmt.Sprintf("%s-execute", task.ID),
			Description: "Ejecutar la tarea",
			AgentType:   "executor",
			Action:      "execute",
			DependsOn:   []string{fmt.Sprintf("%s-analyze", task.ID)},
		},
		{
			ID:          fmt.Sprintf("%s-verify", task.ID),
			Description: "Verificar el resultado",
			AgentType:   "verifier",
			Action:      "verify",
			DependsOn:   []string{fmt.Sprintf("%s-execute", task.ID)},
		},
		{
			ID:          fmt.Sprintf("%s-report", task.ID),
			Description: "Reportar el resultado",
			AgentType:   "reporter",
			Action:      "report",
			DependsOn:   []string{fmt.Sprintf("%s-verify", task.ID)},
		},
	}
	return steps
}

// Name returns the name of the strategy.
func (d *SequentialDecomposition) Name() string {
	return "sequential"
}

// ParallelDecomposition descompone tareas permitiendo ejecución paralela.
type ParallelDecomposition struct{}

// Decompose descompone una tarea en pasos paralelos.
func (d *ParallelDecomposition) Decompose(task *Task) []TaskStep {
	steps := []TaskStep{
		{
			ID:          fmt.Sprintf("%s-analyze", task.ID),
			Description: "Analizar la tarea",
			AgentType:   "analyzer",
			Action:      "analyze",
		},
		{
			ID:          fmt.Sprintf("%s-execute-1", task.ID),
			Description: "Ejecutar parte 1",
			AgentType:   "executor",
			Action:      "execute",
		},
		{
			ID:          fmt.Sprintf("%s-execute-2", task.ID),
			Description: "Ejecutar parte 2",
			AgentType:   "executor",
			Action:      "execute",
		},
		{
			ID:          fmt.Sprintf("%s-merge", task.ID),
			Description: "Fusionar resultados",
			AgentType:   "executor",
			Action:      "merge",
			DependsOn:   []string{
				fmt.Sprintf("%s-execute-1", task.ID),
				fmt.Sprintf("%s-execute-2", task.ID),
			},
		},
		{
			ID:          fmt.Sprintf("%s-verify", task.ID),
			Description: "Verificar resultado",
			AgentType:   "verifier",
			Action:      "verify",
			DependsOn:   []string{fmt.Sprintf("%s-merge", task.ID)},
		},
	}
	return steps
}

// Name returns the name of the strategy.
func (d *ParallelDecomposition) Name() string {
	return "parallel"
}

// ConditionalDecomposition descompone tareas con condicionales.
type ConditionalDecomposition struct{}

// Decompose descompone una tarea con lógica condicional.
func (d *ConditionalDecomposition) Decompose(task *Task) []TaskStep {
	steps := []TaskStep{
		{
			ID:          fmt.Sprintf("%s-analyze", task.ID),
			Description: "Analizar la tarea",
			AgentType:   "analyzer",
			Action:      "analyze",
		},
		{
			ID:          fmt.Sprintf("%s-branch-1", task.ID),
			Description: "Ejecutar rama condicional 1",
			AgentType:   "executor",
			Action:      "execute_branch_1",
			DependsOn:   []string{fmt.Sprintf("%s-analyze", task.ID)},
		},
		{
			ID:          fmt.Sprintf("%s-branch-2", task.ID),
			Description: "Ejecutar rama condicional 2",
			AgentType:   "executor",
			Action:      "execute_branch_2",
			DependsOn:   []string{fmt.Sprintf("%s-analyze", task.ID)},
		},
		{
			ID:          fmt.Sprintf("%s-resolve", task.ID),
			Description: "Resolver resultado",
			AgentType:   "executor",
			Action:      "resolve",
			DependsOn:   []string{
				fmt.Sprintf("%s-branch-1", task.ID),
				fmt.Sprintf("%s-branch-2", task.ID),
			},
		},
		{
			ID:          fmt.Sprintf("%s-verify", task.ID),
			Description: "Verificar resultado final",
			AgentType:   "verifier",
			Action:      "verify",
			DependsOn:   []string{fmt.Sprintf("%s-resolve", task.ID)},
		},
	}
	return steps
}

// Name returns the name of the strategy.
func (d *ConditionalDecomposition) Name() string {
	return "conditional"
}

// SetMaxSteps establece el número máximo de pasos.
func (p *Planner) SetMaxSteps(max int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.maxSteps = max
}

// SetTimeout establece el timeout para ejecuciones.
func (p *Planner) SetTimeout(timeout time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.timeout = timeout
}