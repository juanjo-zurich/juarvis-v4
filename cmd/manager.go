package cmd

import (
	"fmt"
	"strings"

	"juarvis/pkg/manager"
	"juarvis/pkg/output"
	"juarvis/pkg/root"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// Variables globales del manager
var (
	managerRootPath string
	managerWorkers  int
	agentManager    *manager.Manager
	agentScheduler  *manager.Scheduler
	agentMonitor    *manager.Monitor
)

// Inicializar el manager
func initManager(rootPath string) error {
	if managerRootPath != "" {
		rootPath = managerRootPath
	}

	if rootPath == "" {
		r, err := root.GetRoot()
		if err != nil {
			return err
		}
		rootPath = r
	}

	agentManager = manager.NewManager(rootPath, managerWorkers)
	agentScheduler = manager.NewScheduler(agentManager, managerWorkers, 100)
	agentMonitor = manager.NewMonitor(rootPath, agentManager)
	agentMonitor.StartMonitoring()

	// Cargar estado previo
	_ = agentManager.LoadState()

	return nil
}

// Estilos para el dashboard
var (
	dashTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Padding(0, 1)

	dashBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("212")).
			Padding(0, 1)

	agentBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	dashGreenStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	dashYellowStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	dashRedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	dashCyanStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	dashPurpleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("141"))
	dashDimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

// ═════════════════════════════════════════════════════════════════════════════
// COMANDO: manager (raíz)
// ═════════════════════════════════════════════════════════════════════════════

var managerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Gestión de agentes IA en paralelo",
	Long: `Agent Manager - Coordinación de múltiples agentes trabajando en paralelo.
	
Tipos de agente disponibles:
  code    - Generación y refactor de código
  test    - Escritura de tests
  review  - Code review
  debug   - Debugging
  docs    - Documentación
  custom  - Personalizado`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			output.Error("Error al mostrar ayuda: %v", err)
		}
	},
}

// Inicializar manager command
func init() {
	managerCmd.PersistentFlags().StringVar(&managerRootPath, "root", "", "Directorio raíz del ecosistema")
	managerCmd.PersistentFlags().IntVarP(&managerWorkers, "workers", "w", 4, "Número de workers paralelos")

	rootCmd.AddCommand(managerCmd)
}

// ═══════════════════════════════════════════════════════════════════���═════════
// SUBCOMANDO: manager start
// ═════════════════════════════════════════════════════════════════════════════

var managerStartCmd = &cobra.Command{
	Use:   "start <tipo> <tarea>",
	Short: "Iniciar un nuevo agente",
	Long:  "Inicia un nuevo agente del tipo especificado con la tarea dada.",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initManager(""); err != nil {
			output.Error("Ejecuta desde un proyecto con juarvis init: %v", err)
			return err
		}

		agentType := args[0]
		task := strings.Join(args[1:], " ")

		agentTypeEnum := manager.GetAgentTypeByName(agentType)
		agentConfig := manager.AgentConfig{
			Type: agentTypeEnum,
			Name: fmt.Sprintf("%s-agent-%s", agentType, shortID()),
			Task: task,
		}

		agent, err := agentManager.CreateAgent(agentConfig)
		if err != nil {
			output.Error("Error al crear agente: %v", err)
			return fmt.Errorf("error al crear agente: %w", err)
		}

		output.Success("Agente %s iniciado (ID: %s)", agentType, agent.ID)
		output.Info("Tipo: %s", manager.GetAgentTypeDescription(agentTypeEnum))
		output.Info("Task: %s", task)

		// Guardar estado
		_ = agentManager.SaveState()

		return nil
	},
}

func init() {
	managerCmd.AddCommand(managerStartCmd)
}

// ═════════════════════════════════════════════════════════════════════════════
// SUBCOMANDO: manager list
// ═════════════════════════════════════════════════════════════════════════════

var managerListCmd = &cobra.Command{
	Use:   "list",
	Short: "Listar agentes activos",
	Long:  "Muestra todos los agentes registrados y su estado.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initManager(""); err != nil {
			output.Error("Ejecuta desde un proyecto con juarvis init: %v", err)
			return err
		}

		agents := agentManager.ListAgents()

		if len(agents) == 0 {
			output.Info("No hay agentes registrados")
			return nil
		}

		// Generar output formateado
		var sb strings.Builder
		sb.WriteString("\n")
		sb.WriteString(dashTitleStyle.Render("═══ AGENTES REGISTRADOS ═══\n"))

		// Tabla de agentes
		header := fmt.Sprintf("%-8s %-10s %-12s %-8s %s\n",
			"ID", "TIPO", "ESTADO", "PROGRESO", "NOMBRE")
		sb.WriteString(dashDimStyle.Render(header))
		sb.WriteString(dashDimStyle.Render(strings.Repeat("─", 70)) + "\n")

		for _, agent := range agents {
			id := agent.ID
			if len(id) > 8 {
				id = id[:8]
			}

			var stateColor string
			switch agent.State {
			case manager.AgentStateRunning:
				stateColor = "82"
			case manager.AgentStatePending:
				stateColor = "226"
			case manager.AgentStatePaused:
				stateColor = "141"
			case manager.AgentStateComplete:
				stateColor = "39"
			case manager.AgentStateFailed:
				stateColor = "203"
			default:
				stateColor = "245"
			}

			stateStr := lipgloss.NewStyle().Foreground(lipgloss.Color(stateColor)).Render(string(agent.State))
			progressStr := fmt.Sprintf("%.0f%%", agent.Progress*100)

			line := fmt.Sprintf("%-8s %-10s %-12s %-8s %s\n",
				id,
				string(agent.Type),
				stateStr,
				progressStr,
				truncateName(agent.Name, 25),
			)
			sb.WriteString(line)
		}

		sb.WriteString(dashDimStyle.Render("\nTotal: ") + dashCyanStyle.Render(fmt.Sprintf("%d agentes", len(agents))))

		output.Info("%s", sb.String())
		return nil
	},
}

func init() {
	managerCmd.AddCommand(managerListCmd)
}

// ═════════════════════════════════════════════════════════════════════════════
// SUBCOMANDO: manager status
// ═════════════════════════════════════════════════════════════════════════════

var managerStatusCmd = &cobra.Command{
	Use:   "status <agent-id>",
	Short: "Ver estado de un agente",
	Long:  "Muestra información detallada de un agente específico.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initManager(""); err != nil {
			output.Error("Ejecuta desde un proyecto con juarvis init: %v", err)
			return err
		}

		agentID := args[0]
		agent, err := agentManager.GetAgent(agentID)
		if err != nil {
			output.Error("Agente no encontrado: %s", agentID)
			return fmt.Errorf("agente no encontrado: %s", agentID)
		}

		// Mostrar detalles
		var sb strings.Builder
		sb.WriteString("\n")
		sb.WriteString(dashTitleStyle.Render("═══ ESTADO DEL AGENTE ═══\n\n"))

		// Info básica
		sb.WriteString(dashBoxStyle.Width(30).Render("ID: " + agent.ID))
		sb.WriteString("  ")
		sb.WriteString(dashBoxStyle.Width(30).Render("Tipo: " + string(agent.Type)))
		sb.WriteString("\n")

		sb.WriteString(dashBoxStyle.Width(30).Render("Nombre: " + agent.Name))
		sb.WriteString("  ")
		sb.WriteString(dashBoxStyle.Width(30).Render("Estado: " + string(agent.State)))
		sb.WriteString("\n\n")

		// Progreso
		sb.WriteString(dashTitleStyle.Render("Progreso\n"))
		progressBar := renderProgressBar(agent.Progress, 40)
		sb.WriteString(progressBar)
		sb.WriteString("\n\n")

		// Tarea
		sb.WriteString(dashTitleStyle.Render("Tarea\n"))
		sb.WriteString(dashBoxStyle.Width(70).Render(truncateName(agent.Task, 60)))
		sb.WriteString("\n\n")

		// Artefactos
		artifacts := agentManager.GetArtifactsByAgent(agentID)
		sb.WriteString(dashTitleStyle.Render(fmt.Sprintf("Artefactos (%d)\n", len(artifacts))))
		if len(artifacts) == 0 {
			sb.WriteString(dashDimStyle.Render("  Sin artefactos\n"))
		} else {
			for _, art := range artifacts {
				sb.WriteString(fmt.Sprintf("  • %s (%s)\n", art.Type, art.Path))
			}
		}

		// Logs recientes
		sb.WriteString(dashTitleStyle.Render("\nLogs Recientes\n"))
		if len(agent.Logs) > 0 {
			logLines := agent.Logs
			if len(logLines) > 10 {
				logLines = logLines[len(logLines)-10:]
			}
			for _, log := range logLines {
				sb.WriteString(dashDimStyle.Render("  " + log + "\n"))
			}
		} else {
			sb.WriteString(dashDimStyle.Render("  Sin logs\n"))
		}

		output.Info("%s", sb.String())
		return nil
	},
}

func init() {
	managerCmd.AddCommand(managerStatusCmd)
}

// ═════════════════════════════════════════════════════════════════════════════
// SUBCOMANDO: manager pause
// ═════════════════════════════════════════════════════════════════════════════

var managerPauseCmd = &cobra.Command{
	Use:   "pause <agent-id>",
	Short: "Pausar un agente en ejecución",
	Long:  "Pausa temporalmente un agente que está en ejecución.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initManager(""); err != nil {
			output.Error("Ejecuta desde un proyecto con juarvis init: %v", err)
			return err
		}

		agentID := args[0]
		if err := agentManager.PauseAgent(agentID); err != nil {
			output.Error("Error al pausar agente: %v", err)
			return fmt.Errorf("error al pausar agente: %w", err)
		}

		output.Success("Agente %s pausado", agentID[:8])
		_ = agentManager.SaveState()
		return nil
	},
}

func init() {
	managerCmd.AddCommand(managerPauseCmd)
}

// ═════════════════════════════════════════════════════════════════════════════
// SUBCOMANDO: manager cancel
// ═════════════════════════════════════════════════════════════════════════════

var managerCancelCmd = &cobra.Command{
	Use:   "cancel <agent-id>",
	Short: "Cancelar un agente",
	Long:  "Cancela definitivamente un agente.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initManager(""); err != nil {
			output.Error("Ejecuta desde un proyecto con juarvis init: %v", err)
			return err
		}

		agentID := args[0]
		if err := agentManager.CancelAgent(agentID); err != nil {
			output.Error("Error al cancelar agente: %v", err)
			return fmt.Errorf("error al cancelar agente: %w", err)
		}

		output.Warning("Agente %s cancelado", agentID[:8])
		_ = agentManager.SaveState()
		return nil
	},
}

func init() {
	managerCmd.AddCommand(managerCancelCmd)
}

// ═════════════════════════════════════════════════════════════════════════════
// SUBCOMANDO: manager dashboard
// ═════════════════════════════════════════════════════════════════════════════

var managerDashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Abrir dashboard visual de agentes",
	Long:  "Muestra un panel de control visual con el estado de todos los agentes.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initManager(""); err != nil {
			output.Error("Ejecuta desde un proyecto con juarvis init: %v", err)
			return err
		}

		// Renderizar dashboard usando el monitor
		dashboard := agentMonitor.RenderDashboard()
		fmt.Println(dashboard)

		return nil
	},
}

func init() {
	managerCmd.AddCommand(managerDashboardCmd)
}

// ═════════════════════════════════════════════════════════════════════════════
// SUBCOMANDO: manager chain (encadenar agentes)
// ═════════════════════════════════════════════════════════════════════════════

var managerChainCmd = &cobra.Command{
	Use:   "chain <agent-id> <task>",
	Short: "Crear agente encadenado",
	Long:  "Crea un nuevo agente que usa la salida de otro como entrada.",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initManager(""); err != nil {
			output.Error("Ejecuta desde un proyecto con juarvis init: %v", err)
			return err
		}

		dependsOnID := args[0]
		task := strings.Join(args[1:], " ")

		// Verificar que el agente dependencia existe
		_, err := agentManager.GetAgent(dependsOnID)
		if err != nil {
			output.Error("Agente dependencia no encontrado: %s", dependsOnID)
			return fmt.Errorf("agente dependencia no encontrado: %s", dependsOnID)
		}

		agentConfig := manager.AgentConfig{
			Type:      manager.AgentTypeCustom,
			Name:      fmt.Sprintf("chained-agent-%s", shortID()),
			Task:      task,
			DependsOn: []string{dependsOnID},
		}

		agent, err := agentManager.CreateAgent(agentConfig)
		if err != nil {
			output.Error("Error al crear agente encadenado: %v", err)
			return fmt.Errorf("error al crear agente encadenado: %w", err)
		}

		output.Success("Agente encadenado creado (ID: %s)", agent.ID)
		output.Info("Depende de: %s", dependsOnID[:8])
		_ = agentManager.SaveState()

		return nil
	},
}

func init() {
	managerCmd.AddCommand(managerChainCmd)
}

// ═════════════════════════════════════════════════════════════════════════════
// Funciones helper
// ═════════════════════════════════════════════════════════════════════════════

func shortID() string {
	var count int
	if agentManager != nil {
		count = len(agentManager.ListAgents())
	}
	return fmt.Sprintf("%04x", count)
}

func truncateName(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func renderProgressBar(progress float64, width int) string {
	filled := int(progress * float64(width))
	empty := width - filled

	var bar strings.Builder
	bar.WriteString(dashDimStyle.Render("["))
	for i := 0; i < filled; i++ {
		bar.WriteString(dashGreenStyle.Render("█"))
	}
	for i := 0; i < empty; i++ {
		bar.WriteString(dashDimStyle.Render("░"))
	}
	bar.WriteString(dashDimStyle.Render("]"))
	bar.WriteString(" " + fmt.Sprintf("%.1f%%", progress*100))

	return bar.String()
}

// Config representa la configuración del manager almacenada en config.yaml
type Config struct {
	Manager ManagerConfig `yaml:"manager"`
}

// ManagerConfig configuración para el agent manager
type ManagerConfig struct {
	Workers int    `yaml:"workers"`
	Root    string `yaml:"root"`
}