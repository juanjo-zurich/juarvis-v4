package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"juarvis/pkg/config"

	"github.com/charmbracelet/lipgloss"
)

// Monitor seguimiento en tiempo real de agentes
type Monitor struct {
	mu             sync.RWMutex
	manager       *Manager
	rootPath       string
	pollingInterval time.Duration
	historyPath    string
	events         []AgentEvent
	maxEvents      int
}

// AgentEvent representa un evento en la vida de un agente
type AgentEvent struct {
	ID        string      `json:"id"`
	AgentID   string      `json:"agent_id"`
	Type      EventType   `json:"type"`
	Message   string      `json:"message"`
	Timestamp time.Time   `json:"timestamp"`
	Metadata  interface{} `json:"metadata,omitempty"`
}

// EventType tipos de eventos
type EventType string

const (
	EventCreated   EventType = "created"
	EventStarted  EventType = "started"
	EventProgress EventType = "progress"
	EventPaused   EventType = "paused"
	EventResumed  EventType = "resumed"
	EventComplete EventType = "completed"
	EventFailed   EventType = "failed"
	EventCanceled EventType = "canceled"
	EventLog      EventType = "log"
	EventArtifact EventType = "artifact"
)

// Metrics métricas de rendimiento de agentes
type Metrics struct {
	TotalAgents    int            `json:"total_agents"`
	ActiveAgents   int            `json:"active_agents"`
	CompletedCount int            `json:"completed_count"`
	FailedCount    int            `json:"failed_count"`
	AvgDuration    time.Duration `json:"avg_duration"`
	TotalArtifacts int            `json:"total_artifacts"`
	ByType         map[AgentType]int `json:"by_type"`
}

// NewMonitor crea un nuevo monitor de agentes
func NewMonitor(rootPath string, manager *Manager) *Monitor {
	return &Monitor{
		mu:              sync.RWMutex{},
		manager:         manager,
		rootPath:        rootPath,
		pollingInterval: 1 * time.Second,
		historyPath:     filepath.Join(rootPath, config.JuarDir, "agent-history"),
		events:          make([]AgentEvent, 0),
		maxEvents:       1000,
	}
}

// StartMonitoring iniciar seguimiento
func (m *Monitor) StartMonitoring() {
	// Asegurar directorio de historial
	if err := os.MkdirAll(m.historyPath, 0755); err == nil {
		m.loadHistory()
	}
}

// RecordEvent registrar un evento
func (m *Monitor) RecordEvent(agentID string, eventType EventType, message string, metadata interface{}) {
	event := AgentEvent{
		ID:        fmt.Sprintf("%s-%d", agentID, time.Now().UnixNano()),
		AgentID:   agentID,
		Type:      eventType,
		Message:   message,
		Timestamp: time.Now().UTC(),
		Metadata:  metadata,
	}

	m.mu.Lock()
	m.events = append(m.events, event)
	if len(m.events) > m.maxEvents {
		m.events = m.events[len(m.events)-m.maxEvents:]
	}
	m.mu.Unlock()

	// Persistir evento
	m.persistEvent(event)
}

// persistEvent guardar evento en disco
func (m *Monitor) persistEvent(event AgentEvent) {
	eventFile := filepath.Join(m.historyPath, fmt.Sprintf("%s.json", event.ID))
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	os.WriteFile(eventFile, data, 0644)
}

// loadHistory cargar historial de eventos
func (m *Monitor) loadHistory() {
	entries, err := os.ReadDir(m.historyPath)
	if err != nil {
		return
	}

	var events []AgentEvent
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".json" {
			data, err := os.ReadFile(filepath.Join(m.historyPath, e.Name()))
			if err != nil {
				continue
			}
			var event AgentEvent
			if json.Unmarshal(data, &event) == nil {
				events = append(events, event)
			}
		}
	}

	if len(events) > 0 {
		m.mu.Lock()
		m.events = events
		m.mu.Unlock()
	}
}

// GetEvents obtener eventos de agentes específicos
func (m *Monitor) GetEvents(agentID string) []AgentEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []AgentEvent
	for _, event := range m.events {
		if event.AgentID == agentID {
			result = append(result, event)
		}
	}
	return result
}

// GetRecentEvents obtener eventos recientes
func (m *Monitor) GetRecentEvents(count int) []AgentEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if count > len(m.events) {
		count = len(m.events)
	}
	start := len(m.events) - count
	return m.events[start:]
}

// GetMetrics obtener métricas de agentes
func (m *Monitor) GetMetrics() Metrics {
	agents := m.manager.ListAgents()

	metrics := Metrics{
		TotalAgents:  len(agents),
		ActiveAgents: 0,
		ByType:      make(map[AgentType]int),
	}

	var totalDuration time.Duration
	var completedCount int

	for _, agent := range agents {
		metrics.ByType[agent.Type]++

		switch agent.State {
		case AgentStateRunning, AgentStatePending, AgentStatePaused:
			metrics.ActiveAgents++
		case AgentStateComplete:
			completedCount++
			if agent.StartedAt != nil && agent.CompletedAt != nil {
				totalDuration += agent.CompletedAt.Sub(*agent.StartedAt)
			}
		case AgentStateFailed:
			metrics.FailedCount++
		}

		metrics.TotalArtifacts += len(agent.Artifacts)
	}

	metrics.CompletedCount = completedCount
	if completedCount > 0 {
		metrics.AvgDuration = totalDuration / time.Duration(completedCount)
	}

	return metrics
}

// RenderDashboard renderizar dashboard visual
func (m *Monitor) RenderDashboard() string {
	agents := m.manager.ListAgents()
	metrics := m.GetMetrics()

	// Estilos
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true).
		Padding(0, 1)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("212")).
		Padding(0, 1)

	agentBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	cyanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	purpleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	var sb strings.Builder

	// Título
	sb.WriteString("\n")
	sb.WriteString(titleStyle.Render("═══════════════════════════════════════════════"))
	sb.WriteString(titleStyle.Render("\n  🤖 AGENT MANAGER - Juarvis"))
	sb.WriteString(titleStyle.Render("\n═══════════════════════════════════════════════\n"))

	// Métricas generales
	sb.WriteString(boxStyle.Width(25).Render(fmt.Sprintf("TOTAL: %d", metrics.TotalAgents)))
	sb.WriteString("  ")
	sb.WriteString(boxStyle.Width(25).Render(fmt.Sprintf("ACTIVOS: %d", metrics.ActiveAgents)))
	sb.WriteString("  ")
	sb.WriteString(boxStyle.Width(25).Render(fmt.Sprintf("COMPLETADOS: %s", greenStyle.Render(fmt.Sprintf("%d", metrics.CompletedCount)))))
	sb.WriteString("  ")
	sb.WriteString(boxStyle.Width(25).Render(fmt.Sprintf("FALLIDOS: %s", redStyle.Render(fmt.Sprintf("%d", metrics.FailedCount)))))
	sb.WriteString("\n\n")

	// Leyenda de tipos de agente
	typeLegend := fmt.Sprintf("CODE: %d | TEST: %d | REVIEW: %d | DEBUG: %d | DOCS: %d | CUSTOM: %d",
		metrics.ByType[AgentTypeCode],
		metrics.ByType[AgentTypeTest],
		metrics.ByType[AgentTypeReview],
		metrics.ByType[AgentTypeDebug],
		metrics.ByType[AgentTypeDocs],
		metrics.ByType[AgentTypeCustom],
	)
	sb.WriteString(dimStyle.Render(typeLegend))
	sb.WriteString("\n\n")

	// Lista de agentes
	sb.WriteString(titleStyle.Render("── AGENTES ──\n"))

	for i, agent := range agents {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(m.renderAgentBox(agent, agentBoxStyle, greenStyle, yellowStyle, redStyle, cyanStyle, purpleStyle, dimStyle))
	}

	// Footer
	sb.WriteString("\n\n")
	sb.WriteString(dimStyle.Render("[q] salir · [r] refresh · [Enter] detalles · [p] pausar · [c] cancelar · [l] logs"))

	return sb.String()
}

// renderAgentBox renderizar un agente individual
func (m *Monitor) renderAgentBox(agent *Agent, boxStyle, greenStyle, yellowStyle, redStyle, cyanStyle, purpleStyle, dimStyle lipgloss.Style) string {
	// Determinar estilo según estado
	var statusColor string
	var statusText string
	switch agent.State {
	case AgentStateRunning:
		statusColor = "82" // verde
		statusText = "● RUNNING"
	case AgentStatePending:
		statusColor = "226" // amarillo
		statusText = "○ PENDING"
	case AgentStatePaused:
		statusColor = "141" // morado
		statusText = "◐ PAUSED"
	case AgentStateComplete:
		statusColor = "39" // cyan
		statusText = "✓ COMPLETE"
	case AgentStateFailed:
		statusColor = "203" // rojo
		statusText = "✗ FAILED"
	case AgentStateCanceled:
		statusColor = "245" // gris
		statusText = "⊘ CANCELED"
	}

	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Bold(true)

	// Barra de progreso
	progressBar := m.renderProgressBar(agent.Progress, 30)

	// Artefactos
	artifactsText := ""
	if len(agent.Artifacts) > 0 {
		artifactsText = fmt.Sprintf(" | %d artifacts", len(agent.Artifacts))
	}

	// Tiempo transcurrido
	durationText := ""
	if agent.StartedAt != nil {
		elapsed := time.Since(*agent.StartedAt)
		if agent.CompletedAt != nil {
			elapsed = agent.CompletedAt.Sub(*agent.StartedAt)
		}
		durationText = fmt.Sprintf(" | %s", elapsed.Round(time.Second))
	}

	// Construir box
	idPrefix := agent.ID
	if len(idPrefix) > 8 {
		idPrefix = idPrefix[:8]
	}

	box := boxStyle.Width(60).Render(
		fmt.Sprintf("%s %s\n",
			statusStyle.Render(statusText),
			dimStyle.Render(idPrefix),
		) +
			fmt.Sprintf("%s\n", agent.Name) +
			fmt.Sprintf("Task: %s\n", truncate(agent.Task, 40)) +
			fmt.Sprintf("%s%s%s",
				progressBar,
				artifactsText,
				durationText,
			),
	)

	return box
}

// renderProgressBar renderizar barra de progreso
func (m *Monitor) renderProgressBar(progress float64, width int) string {
	filled := int(progress * float64(width))
	empty := width - filled

	var bar strings.Builder
	bar.WriteString("[")
	for i := 0; i < filled; i++ {
		bar.WriteString("█")
	}
	for i := 0; i < empty; i++ {
		bar.WriteString("░")
	}
	bar.WriteString("]")
	bar.WriteString(fmt.Sprintf(" %.0f%%", progress*100))

	return bar.String()
}

// GetAgentDetails obtener detalles completos de un agente
func (m *Monitor) GetAgentDetails(agentID string) (map[string]interface{}, error) {
	agent, err := m.manager.GetAgent(agentID)
	if err != nil {
		return nil, err
	}

	artifacts := m.manager.GetArtifactsByAgent(agentID)
	events := m.GetEvents(agentID)

	details := map[string]interface{}{
		"agent":     agent,
		"artifacts": artifacts,
		"events":    events,
		"metrics":   m.GetMetrics(),
	}

	return details, nil
}

// ExportMetrics exportar métricas a archivo
func (m *Monitor) ExportMetrics(path string) error {
	metrics := m.GetMetrics()
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// GetLogLines obtener líneas de log de un agente
func (m *Monitor) GetLogLines(agentID string, lines int) ([]string, error) {
	agent, err := m.manager.GetAgent(agentID)
	if err != nil {
		return nil, err
	}

	logs := agent.Logs
	if lines > 0 && lines < len(logs) {
		logs = logs[len(logs)-lines:]
	}
	return logs, nil
}

// truncate string helper
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}