package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"juarvis/pkg/output"
	"juarvis/pkg/root"
)

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Padding(0, 1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Foreground(lipgloss.Color("212"))

	greenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82"))

	cyanStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))

	purpleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("141"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82"))

	boldStyle = lipgloss.NewStyle().
			Bold(true)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("212"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Background(lipgloss.Color("236"))

	tabBorder = lipgloss.RoundedBorder()
)

type dashboardData struct {
	projectName   string
	autonomyLevel int
	watcherStatus string
	watcherPID    string
	memCount      int
	lastMemAgo    string
	sessionCount  int
	globalSkills  int
	projectSkills int
	hookRules     int
	snapCount     int
	lastSnapAgo   string
}

func getDashboardData(rootPath string) dashboardData {
	d := dashboardData{
		projectName:   filepath.Base(rootPath),
		autonomyLevel: 2,
		watcherStatus: "○ INACTIVO",
		watcherPID:    "-",
		memCount:      0,
		sessionCount:  0,
		globalSkills:  73,
		projectSkills: 0,
		hookRules:     0,
		snapCount:     0,
	}

	// Config - autonomy level
	configPath := filepath.Join(rootPath, ".juarvis", "config.yaml")
	if data, err := os.ReadFile(configPath); err == nil {
		content := string(data)
		if strings.Contains(content, "autonomy_level: 0") {
			d.autonomyLevel = 0
		} else if strings.Contains(content, "autonomy_level: 1") {
			d.autonomyLevel = 1
		} else if strings.Contains(content, "autonomy_level: 3") {
			d.autonomyLevel = 3
		} else if strings.Contains(content, "autonomy_level: 4") {
			d.autonomyLevel = 4
		}
	}

	// Watcher
	watcherFile := filepath.Join(rootPath, ".juarvis", ".watcher.pid")
	if data, err := os.ReadFile(watcherFile); err == nil {
		d.watcherStatus = "● ACTIVO"
		d.watcherPID = strings.TrimSpace(string(data))
		if len(d.watcherPID) > 6 {
			d.watcherPID = d.watcherPID[:6]
		}
	}

	// Memory observations
	memPath := filepath.Join(rootPath, ".juar", "memory", "observations")
	if entries, err := os.ReadDir(memPath); err == nil {
		var lastTime time.Time
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
				d.memCount++
				if info, err := e.Info(); err == nil && info.ModTime().After(lastTime) {
					lastTime = info.ModTime()
				}
			}
		}
		if !lastTime.IsZero() {
			d.lastMemAgo = time.Since(lastTime).Round(time.Minute).String()
		}
	}

	// Sessions
	sessPath := filepath.Join(rootPath, ".juar", "memory", "sessions")
	if entries, err := os.ReadDir(sessPath); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
				d.sessionCount++
			}
		}
	}

	// Project skills
	skillPath := filepath.Join(rootPath, ".juar", "skills")
	if entries, err := os.ReadDir(skillPath); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				d.projectSkills++
			}
		}
	}

	// Hook rules
	hookPath := filepath.Join(rootPath, ".juar", "hooks")
	if entries, err := os.ReadDir(hookPath); err == nil {
		d.hookRules = len(entries)
	}

	// Snapshots
	snapPath := filepath.Join(rootPath, ".juar", "snapshots")
	if entries, err := os.ReadDir(snapPath); err == nil {
		var snaps []os.FileInfo
		for _, e := range entries {
			if info, err := e.Info(); err == nil {
				snaps = append(snaps, info)
			}
		}
		d.snapCount = len(snaps)
		if len(snaps) > 0 {
			sort.Slice(snaps, func(i, j int) bool {
				return snaps[i].ModTime().After(snaps[j].ModTime())
			})
			d.lastSnapAgo = time.Since(snaps[0].ModTime()).Round(time.Minute).String()
		}
	}

	return d
}

func renderDashboard(d dashboardData) string {
	levelNames := []string{"vibe-puro", "vibe-seguro", "vibe-estructurado", "semi-sdd", "sdd-completo"}
	levelName := levelNames[d.autonomyLevel]

	// Build output
	var sb strings.Builder

	// Header
	sb.WriteString("\n")
	sb.WriteString(boldStyle.Foreground(lipgloss.Color("86")).Render("─ JUARVIS ─ "))
	sb.WriteString(boldStyle.Render(d.projectName))
	sb.WriteString(boldStyle.Foreground(lipgloss.Color("86")).Render(" ─ modo: "))
	sb.WriteString(cyanStyle.Render(levelName))
	sb.WriteString(boldStyle.Foreground(lipgloss.Color("86")).Render(" ─\n"))

	// Status boxes
	sb.WriteString("\n")
	sb.WriteString(borderStyle.Width(35).Render("WATCHER"))
	sb.WriteString("  ")
	sb.WriteString(borderStyle.Width(35).Render("SEGURIDAD"))
	sb.WriteString("\n")

	watcherStatusDisplay := d.watcherStatus
	if d.watcherStatus == "● ACTIVO" {
		watcherStatusDisplay = successStyle.Render("● ACTIVO")
	}
	sb.WriteString(borderStyle.Width(35).Render(fmt.Sprintf("%s %s", watcherStatusDisplay, d.watcherPID)))
	sb.WriteString("  ")
	sb.WriteString(borderStyle.Width(35).Render(fmt.Sprintf("%s · %d reglas", successStyle.Render("0 alertas"), d.hookRules)))
	sb.WriteString("\n")

	// Memory & Skills
	sb.WriteString("\n")
	sb.WriteString(borderStyle.Width(35).Render("MEMORIA"))
	sb.WriteString("  ")
	sb.WriteString(borderStyle.Width(35).Render("SKILLS"))
	sb.WriteString("\n")
	sb.WriteString(borderStyle.Width(35).Render(fmt.Sprintf("%s obs · %s", purpleStyle.Render(fmt.Sprintf("%d", d.memCount)), d.lastMemAgo)))
	sb.WriteString("  ")
	sb.WriteString(borderStyle.Width(35).Render(fmt.Sprintf("%s globales + %s propias", cyanStyle.Render(fmt.Sprintf("%d", d.globalSkills)), cyanStyle.Render(fmt.Sprintf("%d", d.projectSkills)))))
	sb.WriteString("\n")

	// Sessions & Snapshots
	sb.WriteString("\n")
	sb.WriteString(borderStyle.Width(35).Render("SESIONES"))
	sb.WriteString("  ")
	sb.WriteString(borderStyle.Width(35).Render("SNAPSHOTS"))
	sb.WriteString("\n")
	sb.WriteString(borderStyle.Width(35).Render(fmt.Sprintf("%d guardadas", d.sessionCount)))
	sb.WriteString("  ")
	sb.WriteString(borderStyle.Width(35).Render(fmt.Sprintf("%d guardados · %s", d.snapCount, d.lastSnapAgo)))
	sb.WriteString("\n")

	// Footer
	sb.WriteString("\n")
	sb.WriteString(infoStyle.Render("[q] salir · [r] refresh · [m] memoria · [s] skills"))
	sb.WriteString("\n")

	// Tip
	sb.WriteString("\n")
	sb.WriteString(infoStyle.Render("💡 Ejecuta 'juarvis watch --daemon' para activar watcher"))
	sb.WriteString("\n")

	return sb.String()
}

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Panel visual del ecosistema Juarvis",
	Long:  `Muestra el estado del ecosistema con formato visual.`,
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := root.GetRoot()
		if err != nil {
			output.Fatal(output.ExitNoEcosystem,
				"Ejecuta desde un proyecto con juarvis init",
				"Error: %v", err)
		}

		d := getDashboardData(rootPath)
		output.Info("%s", renderDashboard(d))
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}
