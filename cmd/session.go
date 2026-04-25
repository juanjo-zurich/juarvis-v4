package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"juarvis/pkg/output"
	"juarvis/pkg/root"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Gestión de sesiones y checkpoints",
	Long: `Guarda y restaura sesiones:
  session save [nombre] - Guarda estado actual
  session list         - Lista sesiones guardadas
  session resume [id]  - Restaura una sesión
  session export      - Exporta a JSON`,

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
		}
	},
}

var sessionSaveCmd = &cobra.Command{
	Use:   "save [nombre]",
	Short: "Guarda el estado actual de la sesión",
	Args:  cobra.RangeArgs(0, 1),

	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := root.GetRoot()
		if err != nil {
			output.Fatal(output.ExitNoEcosystem, "No hay ecosistema", "")
		}

		// Get name
		name := "session"
		if len(args) > 0 {
			name = args[0]
		}
		timestamp := time.Now().Format("2006-01-02T15-04-05")
		if len(args) == 0 {
			name = name + "-" + timestamp
		}

		// Create session directory
		sessionDir := filepath.Join(rootPath, ".juar", "sessions", name)
		os.MkdirAll(sessionDir, 0755)

		// Save git status
		saveGitOutput("git", []string{"status", "--short"}, filepath.Join(sessionDir, "git-status.txt"))
		saveGitOutput("git", []string{"diff", "--staged"}, filepath.Join(sessionDir, "git-staged.txt"))
		saveGitOutput("git", []string{"diff"}, filepath.Join(sessionDir, "git-diff.txt"))

		// Save metadata
		metadata := map[string]interface{}{
			"name":          name,
			"timestamp":     timestamp,
			"root":          rootPath,
			"currentBranch": getCurrentBranch(),
		}
		metaJSON, _ := json.MarshalIndent(metadata, "", "  ")
		os.WriteFile(filepath.Join(sessionDir, "metadata.json"), metaJSON, 0644)

		output.Success("Sesión guardada: %s", name)
		output.Info("Ubicación: %s", sessionDir)
	},
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista sesiones guardadas",
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := root.GetRoot()
		if err != nil {
			output.Fatal(output.ExitNoEcosystem, "No hay ecosistema", "")
		}

		sessionsDir := filepath.Join(rootPath, ".juar", "sessions")
		entries, err := os.ReadDir(sessionsDir)
		if err != nil {
			output.Info("No hay sesiones guardadas")
			return
		}

		var sessions [][]string
		for _, e := range entries {
			if e.IsDir() {
				metaFile := filepath.Join(sessionsDir, e.Name(), "metadata.json")
				if data, err := os.ReadFile(metaFile); err == nil {
					var meta map[string]interface{}
					if json.Unmarshal(data, &meta) == nil {
						ts := meta["timestamp"]
						branch := meta["currentBranch"]
						sessions = append(sessions, []string{e.Name(), fmt.Sprintf("%v", ts), fmt.Sprintf("%v", branch)})
					}
				}
			}
		}

		if len(sessions) == 0 {
			output.Info("No hay sesiones guardadas")
			return
		}

		output.Success("%d sesiones:", len(sessions))
		output.PrintTable([]string{"NOMBRE", "FECHA", "BRANCH"}, sessions)
	},
}

var sessionResumeCmd = &cobra.Command{
	Use:   "resume [nombre|id]",
	Short: "Restaura una sesión guardada",
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		rootPath, _ := root.GetRoot()
		name := args[0]

		sessionDir := filepath.Join(rootPath, ".juar", "sessions", name)
		if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
			output.Fatal(output.ExitGeneric, "Sesión no encontrada", "name: %v", name)
		}

		output.Info("Restaurando sesión: %s", name)

		// Read git status
		statusFile := filepath.Join(sessionDir, "git-status.txt")
		if data, err := os.ReadFile(statusFile); err == nil {
			output.Info("Cambios en la sesión:")
			fmt.Println(string(data))
		}

		output.Success("Sesión %s restaurada", name)
	},
}

var sessionExportCmd = &cobra.Command{
	Use:   "export [nombre]",
	Short: "Exporta sesión a JSON",
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		rootPath, _ := root.GetRoot()
		name := args[0]

		sessionDir := filepath.Join(rootPath, ".juar", "sessions", name)
		metaFile := filepath.Join(sessionDir, "metadata.json")
		data, err := os.ReadFile(metaFile)
		if err != nil {
			output.Fatal(output.ExitGeneric, "Sesión no encontrada", "name: %v", name)
		}

		fmt.Println(string(data))
	},
}

func getCurrentBranch() string {
	cmd := exec.Command("git", "branch", "--show-current")
	out, _ := cmd.Output()
	return string(out)
}

func saveGitOutput(name string, args []string, path string) {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	if err == nil {
		os.WriteFile(path, out, 0644)
	}
}

func init() {
	sessionCmd.AddCommand(sessionSaveCmd)
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionResumeCmd)
	sessionCmd.AddCommand(sessionExportCmd)

	rootCmd.AddCommand(sessionCmd)
}
