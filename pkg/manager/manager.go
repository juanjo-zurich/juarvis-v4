package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"juarvis/pkg/config"

	"github.com/google/uuid"
)

// AgentType define los tipos de agente configurable
type AgentType string

const (
	AgentTypeCode   AgentType = "code"
	AgentTypeTest   AgentType = "test"
	AgentTypeReview AgentType = "review"
	AgentTypeDebug  AgentType = "debug"
	AgentTypeDocs   AgentType = "docs"
	AgentTypeCustom AgentType = "custom"
)

// AgentState representa el estado de un agente
type AgentState string

const (
	AgentStatePending   AgentState = "pending"
	AgentStateRunning  AgentState = "running"
	AgentStatePaused   AgentState = "paused"
	AgentStateComplete AgentState = "complete"
	AgentStateFailed   AgentState = "failed"
	AgentStateCanceled AgentState = "canceled"
)

// Agent representa un agente ejecutándose en el sistema
type Agent struct {
	ID          string                 `json:"id"`
	Type        AgentType              `json:"type"`
	Name        string                 `json:"name"`
	Task        string                 `json:"task"`
	State       AgentState             `json:"state"`
	Progress    float64                `json:"progress"`
	Workspace   string                 `json:"workspace"`
	Artifacts   []string               `json:"artifacts"`
	Logs        []string               `json:"logs"`
	Output      string                 `json:"output"`
	DependsOn   []string               `json:"depends_on,omitempty"`
	DependsOnBy []string              `json:"depends_on_by,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Artifact representa un artefacto generado por un agente
type Artifact struct {
	ID        string    `json:"id"`
	AgentID   string    `json:"agent_id"`
	Type      string    `json:"type"`
	Path      string    `json:"path"`
	Content   string    `json:"content,omitempty"`
	Generated time.Time `json:"generated"`
	Checksum  string    `json:"checksum"`
}

// AgentConfig configuración para crear un nuevo agente
type AgentConfig struct {
	Type      AgentType
	Name      string
	Task      string
	Workspace string
	DependsOn []string
	Metadata  map[string]interface{}
}

// Manager coordinación de múltiples agentes
type Manager struct {
	mu        sync.RWMutex
	agents    map[string]*Agent
	artifacts map[string]*Artifact
	rootPath  string
	queue     chan *Agent
	workers   int
}

// NewManager crea un nuevo Agent Manager
func NewManager(rootPath string, workers int) *Manager {
	if workers <= 0 {
		workers = 4
	}
	return &Manager{
		mu:        sync.RWMutex{},
		agents:    make(map[string]*Agent),
		artifacts: make(map[string]*Artifact),
		rootPath:  rootPath,
		queue:     make(chan *Agent, 100),
		workers:   workers,
	}
}

// CreateAgent crea y registra un nuevo agente
func (m *Manager) CreateAgent(config AgentConfig) (*Agent, error) {
	agentID := uuid.New().String()

	agent := &Agent{
		ID:          agentID,
		Type:        config.Type,
		Name:        config.Name,
		Task:        config.Task,
		State:       AgentStatePending,
		Progress:    0.0,
		Workspace:  config.Workspace,
		Artifacts:  []string{},
		Logs:        []string{},
		Output:      "",
		DependsOn:   config.DependsOn,
		CreatedAt:   time.Now().UTC(),
		Metadata:    config.Metadata,
	}

	// Validar dependencias
	for _, depID := range config.DependsOn {
		depAgent, err := m.GetAgent(depID)
		if err != nil {
			return nil, fmt.Errorf("dependencia no encontrada: %s", depID)
		}
		depAgent.DependsOnBy = append(depAgent.DependsOnBy, agentID)
	}

	m.mu.Lock()
	m.agents[agentID] = agent
	m.mu.Unlock()

	// Agregar a la cola de ejecución
	m.queue <- agent

	return agent, nil
}

// GetAgent obtener un agente por ID
func (m *Manager) GetAgent(id string) (*Agent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agent, ok := m.agents[id]
	if !ok {
		return nil, fmt.Errorf("agente no encontrado: %s", id)
	}
	return agent, nil
}

// ListAgents listar todos los agentes
func (m *Manager) ListAgents() []*Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agents := make([]*Agent, 0, len(m.agents))
	for _, agent := range m.agents {
		agents = append(agents, agent)
	}
	return agents
}

// ListAgentsByState listar agentes por estado
func (m *Manager) ListAgentsByState(state AgentState) []*Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Agent
	for _, agent := range m.agents {
		if agent.State == state {
			result = append(result, agent)
		}
	}
	return result
}

// UpdateAgentState actualizar el estado de un agente
func (m *Manager) UpdateAgentState(id string, state AgentState, progress float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent, ok := m.agents[id]
	if !ok {
		return fmt.Errorf("agente no encontrado: %s", id)
	}

	agent.State = state
	agent.Progress = progress

	if state == AgentStateRunning && agent.StartedAt == nil {
		now := time.Now().UTC()
		agent.StartedAt = &now
	}

	if state == AgentStateComplete || state == AgentStateFailed || state == AgentStateCanceled {
		now := time.Now().UTC()
		agent.CompletedAt = &now
	}

	return nil
}

// AddLog añadir un log a un agente
func (m *Manager) AddLog(id string, log string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent, ok := m.agents[id]
	if !ok {
		return fmt.Errorf("agente no encontrado: %s", id)
	}

	timestamp := time.Now().UTC().Format("15:04:05")
	agent.Logs = append(agent.Logs, fmt.Sprintf("[%s] %s", timestamp, log))
	return nil
}

// SetOutput establecer output de un agente
func (m *Manager) SetOutput(id string, output string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent, ok := m.agents[id]
	if !ok {
		return fmt.Errorf("agente no encontrado: %s", id)
	}

	agent.Output = output
	return nil
}

// AddArtifact añadir un artefacto generado por un agente
func (m *Manager) AddArtifact(agentID string, artifact *Artifact) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent, ok := m.agents[agentID]
	if !ok {
		return fmt.Errorf("agente no encontrado: %s", agentID)
	}

	artifact.AgentID = agentID
	artifact.ID = uuid.New().String()
	artifact.Generated = time.Now().UTC()

	m.artifacts[artifact.ID] = artifact
	agent.Artifacts = append(agent.Artifacts, artifact.ID)

	return nil
}

// GetArtifact obtener un artefacto por ID
func (m *Manager) GetArtifact(id string) (*Artifact, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	artifact, ok := m.artifacts[id]
	if !ok {
		return nil, fmt.Errorf("artefacto no encontrado: %s", id)
	}
	return artifact, nil
}

// GetArtifactsByAgent obtener todos los artefactos de un agente
func (m *Manager) GetArtifactsByAgent(agentID string) []*Artifact {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Artifact
	agent, ok := m.agents[agentID]
	if !ok {
		return result
	}

	for _, artID := range agent.Artifacts {
		if art, ok := m.artifacts[artID]; ok {
			result = append(result, art)
		}
	}
	return result
}

// PauseAgent pausar un agente
func (m *Manager) PauseAgent(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent, ok := m.agents[id]
	if !ok {
		return fmt.Errorf("agente no encontrado: %s", id)
	}

	if agent.State != AgentStateRunning {
		return fmt.Errorf("el agente no está en ejecución")
	}

	agent.State = AgentStatePaused
	return nil
}

// ResumeAgent reanudar un agente pausado
func (m *Manager) ResumeAgent(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent, ok := m.agents[id]
	if !ok {
		return fmt.Errorf("agente no encontrado: %s", id)
	}

	if agent.State != AgentStatePaused {
		return fmt.Errorf("el agente no está pausado")
	}

	agent.State = AgentStateRunning
	return nil
}

// CancelAgent cancelar un agente
func (m *Manager) CancelAgent(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent, ok := m.agents[id]
	if !ok {
		return fmt.Errorf("agente no encontrado: %s", id)
	}

	if agent.State == AgentStateComplete || agent.State == AgentStateFailed || agent.State == AgentStateCanceled {
		return fmt.Errorf("el agente ya ha terminado")
	}

	agent.State = AgentStateCanceled
	now := time.Now().UTC()
	agent.CompletedAt = &now
	return nil
}

// GetDependentOutput obtener el output de un agente del que depende otro
func (m *Manager) GetDependentOutput(agentID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agent, ok := m.agents[agentID]
	if !ok {
		return "", fmt.Errorf("agente no encontrado: %s", agentID)
	}

	return agent.Output, nil
}

// CanExecute verificar si un agente puede ejecutarse (dependencias resueltas)
func (m *Manager) CanExecute(agent *Agent) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, depID := range agent.DependsOn {
		depAgent, ok := m.agents[depID]
		if !ok {
			return false
		}
		if depAgent.State != AgentStateComplete {
			return false
		}
	}
	return true
}

// GetQueue returns the agent queue channel
func (m *Manager) GetQueue() chan *Agent {
	return m.queue
}

// GetWorkers returns the number of workers
func (m *Manager) GetWorkers() int {
	return m.workers
}

// SaveState guardar el estado del manager
func (m *Manager) SaveState() error {
	stateDir := filepath.Join(m.rootPath, config.JuarDir, "agent-state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return err
	}

	// Guardar estado de agentes
	stateFile := filepath.Join(stateDir, "agents.json")
	data, err := json.MarshalIndent(m.agents, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return err
	}

	// Guardar artefactos
	artifactsFile := filepath.Join(stateDir, "artifacts.json")
	artData, err := json.MarshalIndent(m.artifacts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(artifactsFile, artData, 0644)
}

// LoadState cargar el estado del manager desde disco
func (m *Manager) LoadState() error {
	stateDir := filepath.Join(m.rootPath, config.JuarDir, "agent-state")

	// Cargar agentes
	stateFile := filepath.Join(stateDir, "agents.json")
	if data, err := os.ReadFile(stateFile); err == nil {
		var agents map[string]*Agent
		if err := json.Unmarshal(data, &agents); err == nil {
			m.mu.Lock()
			m.agents = agents
			m.mu.Unlock()
		}
	}

	// Cargar artefactos
	artifactsFile := filepath.Join(stateDir, "artifacts.json")
	if data, err := os.ReadFile(artifactsFile); err == nil {
		var artifacts map[string]*Artifact
		if err := json.Unmarshal(data, &artifacts); err == nil {
			m.mu.Lock()
			m.artifacts = artifacts
			m.mu.Unlock()
		}
	}

	return nil
}

// GetAgentSummary devuelve un resumen de estadísticas del manager
func (m *Manager) GetAgentSummary() map[string]int {
	summary := make(map[string]int)

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, agent := range m.agents {
		summary[string(agent.State)]++
		summary["total"]++
	}

	return summary
}

// GetActiveAgents devuelve los agentes activos (running, pending, paused)
func (m *Manager) GetActiveAgents() []*Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var active []*Agent
	for _, agent := range m.agents {
		if agent.State == AgentStateRunning || agent.State == AgentStatePending || agent.State == AgentStatePaused {
			active = append(active, agent)
		}
	}
	return active
}

// GetChainOutput obtener output encadenado de un agente
func (m *Manager) GetChainOutput(agentID string) (string, error) {
	m.mu.RLock()
	agent, ok := m.agents[agentID]
	m.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("agente no encontrado: %s", agentID)
	}

	if agent.State != AgentStateComplete {
		return "", fmt.Errorf("el agente dependiente aún no ha terminado")
	}

	return agent.Output, nil
}

// GetAgentTypeByName obtener tipo de agente por nombre string
func GetAgentTypeByName(name string) AgentType {
	switch name {
	case "code":
		return AgentTypeCode
	case "test":
		return AgentTypeTest
	case "review":
		return AgentTypeReview
	case "debug":
		return AgentTypeDebug
	case "docs":
		return AgentTypeDocs
	case "custom":
		return AgentTypeCustom
	default:
		return AgentTypeCustom
	}
}

// GetAgentTypeDescription obtener descripción de tipo de agente
func GetAgentTypeDescription(agentType AgentType) string {
	descriptions := map[AgentType]string{
		AgentTypeCode:   "Generación y refactor de código",
		AgentTypeTest:   "Escritura de tests unitarios y de integración",
		AgentTypeReview: "Revisión de código y análisis estático",
		AgentTypeDebug:  "Depuración y resolución de errores",
		AgentTypeDocs:   "Generación de documentación técnica",
		AgentTypeCustom: "Agente personalizado configurable",
	}
	return descriptions[agentType]
}