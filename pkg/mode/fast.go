// Package mode implementa el modo Fast para respuesta inmediata.
package mode

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// FastExecutor ejecuta tareas en modo Fast con baja latencia.
type FastExecutor struct {
	mu           sync.RWMutex
	maxLatency    time.Duration
	agentCache   *AgentCache
	executorPool *ExecutorPool
	shortcuts   *ShortcutRegistry
}

// AgentCache cache de agentes reutilizables.
type AgentCache struct {
	mu      sync.RWMutex
	agents  map[string]any
	ttl     time.Duration
	expires map[string]time.Time
}

// ExecutorPool pool de ejecutores para ejecución paralela.
type ExecutorPool struct {
	mu       sync.RWMutex
	workers int
	active  int
	queue   chan struct{}
}

// ShortcutRegistry registro de atajos para queries comunes.
type ShortcutRegistry struct {
	mu        sync.RWMutex
	shortcuts map[string]Shortcut
}

// Shortcut representa un atajo para ejecución rápida.
type Shortcut struct {
	Pattern  string
	Action  string
	Handler func(ctx context.Context, input any) (any, error)
}

// NewFastExecutor crea un nuevo ejecutor Fast.
func NewFastExecutor() *FastExecutor {
	return &FastExecutor{
		maxLatency:    5 * time.Second,
		agentCache:   NewAgentCache(),
		executorPool: NewExecutorPool(10),
		shortcuts:   NewShortcutRegistry(),
	}
}

// NewAgentCache crea un nuevo cache de agentes.
func NewAgentCache() *AgentCache {
	return &AgentCache{
		agents:  make(map[string]any),
		ttl:     5 * time.Minute,
		expires: make(map[string]time.Time),
	}
}

// NewExecutorPool crea un nuevo pool de ejecutores.
func NewExecutorPool(workers int) *ExecutorPool {
	return &ExecutorPool{
		workers: workers,
		queue:  make(chan struct{}, workers),
	}
}

// NewShortcutRegistry crea un nuevo registro de atajos.
func NewShortcutRegistry() *ShortcutRegistry {
	return &ShortcutRegistry{
		shortcuts: make(map[string]Shortcut),
	}
}

// Execute ejecuta una tarea en modo Fast.
func (f *FastExecutor) Execute(ctx context.Context, task *Task) (any, error) {
	startTime := time.Now()
	defer func() {
		latency := time.Since(startTime)
		if latency > f.maxLatency {
			// Log warning for slow execution
			fmt.Printf("[FAST] Warning: execution exceeded max latency (%v > %v)\n", latency, f.maxLatency)
		}
	}()

	// Skip deliberación compleja - ejecución directa
	// Verificar si hay un atajo disponible
	if shortcut := f.shortcuts.Find(task.Description); shortcut != nil {
		return shortcut.Handler(ctx, task.Input)
	}

	// Ejecución directa del agente más apropiado
	agent, err := f.getAgent(ctx, task)
	if err != nil {
		return nil, err
	}

	// Ejecutar directamente
	return f.executeAgent(ctx, agent, task)
}

// ExecuteImmediate ejecuta una tarea de forma inmediata sin deliberación.
func (f *FastExecutor) ExecuteImmediate(ctx context.Context, input string) (any, error) {
	// Crear tarea mínima
	task := &Task{
		ID:          fmt.Sprintf("fast-%d", time.Now().UnixNano()),
		Description: input,
		Complexity:  ComplexitySimple,
		Input:      input,
		CreatedAt:  time.Now(),
	}

	return f.Execute(ctx, task)
}

// executeAgent ejecuta un agente directamente.
func (f *FastExecutor) executeAgent(ctx context.Context, agent any, task *Task) (any, error) {
	// Simulación de ejecución directa
	// En implementación real, aquí se ejecutaría el agente

	return map[string]any{
		"mode":    "fast",
		"taskID": task.ID,
		"result": "completed",
	}, nil
}

// getAgent obtiene un agente apropiado para la tarea.
func (f *FastExecutor) getAgent(ctx context.Context, task *Task) (any, error) {
	// Intentar obtener del cache
	agentKey := "default"
	if agent, found := f.agentCache.Get(agentKey); found {
		return agent, nil
	}

	// Crear nuevo agente
	agent := map[string]any{
		"type": "fast",
		"name": "FastAgent",
	}

	// Guardar en cache
	f.agentCache.Set(agentKey, agent)

	return agent, nil
}

// Get implementa la obtención de un agente del cache.
func (c *AgentCache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if exp, ok := c.expires[key]; ok {
		if time.Now().After(exp) {
			// Expirado
			return nil, false
		}
	}

	agent, ok := c.agents[key]
	return agent, ok
}

// Set implementa el guardado de un agente en cache.
func (c *AgentCache) Set(key string, agent any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.agents[key] = agent
	c.expires[key] = time.Now().Add(c.ttl)
}

// Clear limpia el cache.
func (c *AgentCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.agents = make(map[string]any)
	c.expires = make(map[string]time.Time)
}

// Acquire acquires un worker del pool.
func (p *ExecutorPool) Acquire() bool {
	select {
	case p.queue <- struct{}{}:
		p.mu.Lock()
		p.active++
		p.mu.Unlock()
		return true
	default:
		return false
	}
}

// Release releases un worker del pool.
func (p *ExecutorPool) Release() {
	select {
	case <-p.queue:
		p.mu.Lock()
		p.active--
		p.mu.Unlock()
	default:
	}
}

// Active returns el número de workers activos.
func (p *ExecutorPool) Active() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.active
}

// RegisterShortcut registra un atajo.
func (r *ShortcutRegistry) Register(pattern string, action string, handler func(ctx context.Context, input any) (any, error)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.shortcuts[pattern] = Shortcut{
		Pattern:  pattern,
		Action:  action,
		Handler: handler,
	}
}

// Find encuentra un atajo para la descripción dada.
func (r *ShortcutRegistry) Find(description string) *Shortcut {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, shortcut := range r.shortcuts {
		if containsWord(description, shortcut.Pattern) {
			return &shortcut
		}
	}
	return nil
}

// SetMaxLatency establece la latencia máxima permitida.
func (f *FastExecutor) SetMaxLatency(latency time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.maxLatency = latency
}

// Verify ejecuta una verificación rápida.
func (f *FastExecutor) Verify(ctx context.Context, check string) (bool, error) {
	// Verificación rápida sin deliberación
	task := &Task{
		ID:          fmt.Sprintf("verify-%d", time.Now().UnixNano()),
		Description: check,
		Complexity:  ComplexitySimple,
		Input:      check,
		CreatedAt:  time.Now(),
	}

	result, err := f.Execute(ctx, task)
	if err != nil {
		return false, err
	}

	// Verificar resultado
	if m, ok := result.(map[string]any); ok {
		if status, ok := m["result"].(string); ok {
			return status == "completed", nil
		}
	}

	return false, nil
}

// QuickQuery ejecuta una query simple de forma rápida.
func (f *FastExecutor) QuickQuery(ctx context.Context, query string) (any, error) {
	// Queries simples: buscar, mostrar, listar
	return f.ExecuteImmediate(ctx, query)
}

// LoopVerification ejecuta un loop de verificación.
func (f *FastExecutor) LoopVerification(ctx context.Context, checks []string) ([]bool, error) {
	results := make([]bool, 0, len(checks))

	for _, check := range checks {
		result, err := f.Verify(ctx, check)
		if err != nil {
			results = append(results, false)
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// ParallelExecute ejecuta múltiples tareas en paralelo.
func (f *FastExecutor) ParallelExecute(ctx context.Context, tasks []*Task) ([]any, error) {
	if !f.executorPool.Acquire() {
		return nil, fmt.Errorf("executor pool exhausted")
	}
	defer f.executorPool.Release()

	results := make([]any, 0, len(tasks))
	type result struct {
		result any
		err    error
	}
	resultCh := make(chan result, len(tasks))

	for _, task := range tasks {
		go func(t *Task) {
			r, err := f.Execute(ctx, t)
			resultCh <- result{result: r, err: err}
		}(task)
	}

	for i := 0; i < len(tasks); i++ {
		res := <-resultCh
		if res.err != nil {
			return nil, res.err
		}
		results = append(results, res.result)
	}

	return results, nil
}

// WarmUp precalienta el ejecutor Fast.
func (f *FastExecutor) WarmUp(ctx context.Context) error {
	// Precache agente default
	agentKey := "default"
	f.agentCache.Set(agentKey, map[string]any{
		"type": "fast",
		"name": "FastAgent",
	})

	// Registrar atajos comunes
	f.shortcuts.Register("buscar", "search", func(ctx context.Context, input any) (any, error) {
		return map[string]any{"action": "search", "input": input}, nil
	})
	f.shortcuts.Register("mostrar", "show", func(ctx context.Context, input any) (any, error) {
		return map[string]any{"action": "show", "input": input}, nil
	})
	f.shortcuts.Register("listar", "list", func(ctx context.Context, input any) (any, error) {
		return map[string]any{"action": "list", "input": input}, nil
	})
	f.shortcuts.Register("verificar", "verify", func(ctx context.Context, input any) (any, error) {
		return map[string]any{"action": "verify", "input": input}, nil
	})

	return nil
}

// Stats devuelve estadísticas del ejecutor Fast.
func (f *FastExecutor) Stats() FastStats {
	return FastStats{
		MaxLatency: f.maxLatency,
		Workers:   f.executorPool.workers,
		Active:   f.executorPool.Active(),
	}
}

// FastStats devuelve estadísticas del modo Fast.
type FastStats struct {
	MaxLatency time.Duration `json:"max_latency"`
	Workers   int        `json:"workers"`
	Active    int        `json:"active"`
}