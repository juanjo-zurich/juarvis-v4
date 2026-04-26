package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TaskStatus estado de una tarea en el scheduler
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCanceled  TaskStatus = "canceled"
)

// Task representa una tarea programada
type Task struct {
	ID          string                 `json:"id"`
	AgentID     string                 `json:"agent_id"`
	Type        AgentType              `json:"type"`
	Description string                 `json:"description"`
	Status      TaskStatus             `json:"status"`
	Priority    int                    `json:"priority"`     // Mayor número = mayor prioridad
	Retries     int                    `json:"retries"`      // Reintentos restantes
	MaxRetries  int                    `json:"max_retries"`  // Máximo de reintentos
	CreatedAt   time.Time             `json:"created_at"`
	StartedAt   *time.Time            `json:"started_at,omitempty"`
	CompletedAt *time.Time            `json:"completed_at,omitempty"`
	Result      string                `json:"result,omitempty"`
	Error       string                `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WorkItem representa un elemento de trabajo
type WorkItem struct {
	ID          string
	AgentID     string
	Task        string
	Type        AgentType
	Priority    int
	Ctx         context.Context
	Cancel      context.CancelFunc
	Result      chan<- WorkResult
}

// WorkResult resultado de un work item
type WorkResult struct {
	AgentID    string
	Success    bool
	Output     string
	Artifacts  []string
	Error      error
	Duration   time.Duration
}

// Scheduler distribución y programación de trabajo
type Scheduler struct {
	mu           sync.RWMutex
	manager      *Manager
	tasks        map[string]*Task
	workQueue    chan WorkItem
	workerPool   []Worker
	maxRetries   int
	retryDelay   time.Duration
	shutdownCh  chan struct{}
	wg           sync.WaitGroup
}

// Workerpool worker pool configuration
type Worker struct {
	ID         int
	TaskChan   chan WorkItem
	quit       chan struct{}
}

// NewScheduler crea un nuevo scheduler
func NewScheduler(manager *Manager, workers int, queueSize int) *Scheduler {
	if workers <= 0 {
		workers = 4
	}
	if queueSize <= 0 {
		queueSize = 100
	}

	s := &Scheduler{
		mu:          sync.RWMutex{},
		manager:     manager,
		tasks:       make(map[string]*Task),
		workQueue:   make(chan WorkItem, queueSize),
		workerPool:  make([]Worker, workers),
		maxRetries:  3,
		retryDelay:  2 * time.Second,
		shutdownCh:  make(chan struct{}),
	}

	// Inicializar workers
	for i := 0; i < workers; i++ {
		s.workerPool[i] = Worker{
			ID:       i,
			TaskChan: make(chan WorkItem, 1),
			quit:     make(chan struct{}),
		}
	}

	return s
}

// Start iniciar el scheduler y sus workers
func (s *Scheduler) Start(ctx context.Context) {
	s.wg.Add(len(s.workerPool))
	for i := range s.workerPool {
		go s.runWorker(ctx, &s.workerPool[i])
	}

	// Goroutine para procesar la cola del manager
	s.wg.Add(1)
	go s.processQueue(ctx)
}

// Stop detener el scheduler
func (s *Scheduler) Stop() {
	close(s.shutdownCh)
	s.wg.Wait()
}

// runWorker ejecutar worker
func (s *Scheduler) runWorker(ctx context.Context, w *Worker) {
	defer s.wg.Done()

	for {
		select {
		case <-w.quit:
			return
		case work := <-w.TaskChan:
			s.executeWorkItem(ctx, work)
		case <-ctx.Done():
			return
		}
	}
}

// processQueue procesar cola de agentes del manager
func (s *Scheduler) processQueue(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-s.shutdownCh:
			return
		case <-ctx.Done():
			return
		case agent := <-s.manager.GetQueue():
			if agent != nil {
				s.scheduleAgent(ctx, agent)
			}
		}
	}
}

// scheduleAgent programar un agente para ejecución
func (s *Scheduler) scheduleAgent(ctx context.Context, agent *Agent) {
	work := WorkItem{
		ID:      uuid.New().String(),
		AgentID: agent.ID,
		Task:    agent.Task,
		Type:    agent.Type,
		Priority: 0,
		Ctx:     ctx,
	}

	// Verificar si puede ejecutarse (dependencias resueltas)
	if !s.manager.CanExecute(agent) {
		// Dependencias no resueltas, esperar
		s.manager.UpdateAgentState(agent.ID, AgentStatePending, 0)
		s.manager.AddLog(agent.ID, "Esperando dependencias...")
		
		// Re-intentar en unos segundos
		go func() {
			select {
			case <-time.After(5 * time.Second):
				s.manager.GetQueue() <- agent
			case <-s.shutdownCh:
			}
		}()
		return
	}

	// Obtener output de dependencias si existen
	if len(agent.DependsOn) > 0 {
		for _, depID := range agent.DependsOn {
			output, err := s.manager.GetChainOutput(depID)
			if err == nil && output != "" {
				work.Task = output + "\n\n" + agent.Task
				break
			}
		}
	}

	result := make(chan WorkResult, 1)
	work.Result = result

	s.manager.UpdateAgentState(agent.ID, AgentStateRunning, 0)
	s.manager.AddLog(agent.ID, "Agente iniciado")

	// Asignar al worker con menor carga
	s.assignWork(work)
	
	// Esperar resultado
	select {
	case res := <-result:
		s.handleWorkResult(agent.ID, res)
	case <-ctx.Done():
		s.manager.CancelAgent(agent.ID)
	case <-s.shutdownCh:
		return
	}
}

// assignWork asignar trabajo al worker con menor carga
func (s *Scheduler) assignWork(work WorkItem) {
	// Simple round-robin: asignar al primer worker disponible
	workerIdx := int(work.AgentID[0]) % len(s.workerPool) // Hash simple
	s.workerPool[workerIdx].TaskChan <- work
}

// executeWorkItem ejecutar un work item
func (s *Scheduler) executeWorkItem(ctx context.Context, work WorkItem) {
	startTime := time.Now()

	// Simular ejecución del agente
	s.manager.AddLog(work.AgentID, "Ejecutando tarea...")

	// Ejecutar según tipo de agente
	output, artifacts, err := s.executeAgent(ctx, work)

	duration := time.Since(startTime)

	result := WorkResult{
		AgentID:   work.AgentID,
		Success:   err == nil,
		Output:    output,
		Artifacts: artifacts,
		Error:     err,
		Duration:  duration,
	}

	select {
	case work.Result <- result:
	case <-ctx.Done():
	}
}

// executeAgent ejecutar lógica específica según tipo de agente
func (s *Scheduler) executeAgent(ctx context.Context, work WorkItem) (string, []string, error) {
	// Aquí iría la lógica específica para cada tipo de agente
	// Por ahora es un placeholder - en implementación real se conectaría con los agentes existentes

	switch work.Type {
	case AgentTypeCode:
		return s.executeCodeAgent(ctx, work)
	case AgentTypeTest:
		return s.executeTestAgent(ctx, work)
	case AgentTypeReview:
		return s.executeReviewAgent(ctx, work)
	case AgentTypeDebug:
		return s.executeDebugAgent(ctx, work)
	case AgentTypeDocs:
		return s.executeDocsAgent(ctx, work)
	default:
		return s.executeCustomAgent(ctx, work)
	}
}

// Placeholder methods que se implementarían con los agentes reales de Juarvis
func (s *Scheduler) executeCodeAgent(ctx context.Context, work WorkItem) (string, []string, error) {
	s.manager.AddLog(work.AgentID, "Ejecutando agente de código...")
	// TODO: Integrar con agente go-developer de Juarvis
	return "Código generado", []string{}, nil
}

func (s *Scheduler) executeTestAgent(ctx context.Context, work WorkItem) (string, []string, error) {
	s.manager.AddLog(work.AgentID, "Ejecutando agente de tests...")
	// TODO: Integrar con agente test-engineer de Juarvis
	return "Tests generados", []string{}, nil
}

func (s *Scheduler) executeReviewAgent(ctx context.Context, work WorkItem) (string, []string, error) {
	s.manager.AddLog(work.AgentID, "Ejecutando agente de review...")
	// TODO: Integrar con agente code-reviewer de Juarvis
	return "Review completado", []string{}, nil
}

func (s *Scheduler) executeDebugAgent(ctx context.Context, work WorkItem) (string, []string, error) {
	s.manager.AddLog(work.AgentID, "Ejecutando agente de debug...")
	// TODO: Integrar con agente debugger de Juarvis
	return "Debug completado", []string{}, nil
}

func (s *Scheduler) executeDocsAgent(ctx context.Context, work WorkItem) (string, []string, error) {
	s.manager.AddLog(work.AgentID, "Ejecutando agente de documentación...")
	// TODO: Integrar con agente docs-writer de Juarvis
	return "Documentación generada", []string{}, nil
}

func (s *Scheduler) executeCustomAgent(ctx context.Context, work WorkItem) (string, []string, error) {
	s.manager.AddLog(work.AgentID, "Ejecutando agente personalizado...")
	return "Ejecución completada", []string{}, nil
}

// handleWorkResult manejar resultado de trabajo
func (s *Scheduler) handleWorkResult(agentID string, result WorkResult) {
	if result.Success {
		s.manager.UpdateAgentState(agentID, AgentStateComplete, 1.0)
		s.manager.SetOutput(agentID, result.Output)
		s.manager.AddLog(agentID, fmt.Sprintf("Agente completado en %v", result.Duration))
		
		// Añadir artefactos
		for _, artPath := range result.Artifacts {
			artifact := &Artifact{
				Type: "output",
				Path: artPath,
			}
			_ = s.manager.AddArtifact(agentID, artifact)
		}
	} else {
		s.manager.UpdateAgentState(agentID, AgentStateFailed, 0)
		s.manager.SetOutput(agentID, result.Error.Error())
		s.manager.AddLog(agentID, fmt.Sprintf("Error: %v", result.Error))
	}

	// Notificar a agentes dependientes
	s.notifyDependents(agentID)
}

// notifyDependents notificar a agentes dependientes que pueden ejecutarse
func (s *Scheduler) notifyDependents(agentID string) {
	agents := s.manager.ListAgents()
	for _, agent := range agents {
		for _, dep := range agent.DependsOn {
			if dep == agentID && s.manager.CanExecute(agent) && agent.State == AgentStatePending {
				s.manager.GetQueue() <- agent
			}
		}
	}
}

// ScheduleTask programar una tarea explícitamente
func (s *Scheduler) ScheduleTask(agentType AgentType, description string, priority int) (*Task, error) {
	task := &Task{
		ID:          uuid.New().String(),
		Type:        agentType,
		Description: description,
		Status:      TaskStatusPending,
		Priority:    priority,
		MaxRetries:  s.maxRetries,
		Retries:     s.maxRetries,
		CreatedAt:   time.Now().UTC(),
	}

	s.mu.Lock()
	s.tasks[task.ID] = task
	s.mu.Unlock()

	return task, nil
}

// GetTask obtener tarea por ID
func (s *Scheduler) GetTask(id string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[id]
	if !ok {
		return nil, fmt.Errorf("tarea no encontrada: %s", id)
	}
	return task, nil
}

// ListTasks listar todas las tareas
func (s *Scheduler) ListTasks() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// CancelTask cancelar una tarea
func (s *Scheduler) CancelTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return fmt.Errorf("tarea no encontrada: %s", id)
	}

	task.Status = TaskStatusCanceled
	task.CompletedAt = func() *time.Time { t := time.Now().UTC(); return &t }()

	return nil
}

// GetStats obtener estadísticas del scheduler
func (s *Scheduler) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]interface{}{
		"workers":    len(s.workerPool),
		"queue_size": len(s.workQueue),
		"total_tasks": len(s.tasks),
	}

	// Contar por estado
	taskCounts := make(map[TaskStatus]int)
	for _, task := range s.tasks {
		taskCounts[task.Status]++
	}
	stats["tasks_by_status"] = taskCounts

	return stats
}

// SetMaxRetries establecer reintentos máximos
func (s *Scheduler) SetMaxRetries(max int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maxRetries = max
}

// SetRetryDelay establecer delay de reintentos
func (s *Scheduler) SetRetryDelay(delay time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.retryDelay = delay
}