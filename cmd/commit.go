package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"juarvis/pkg/output"
)

var commitForce bool

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Crea un commit con mensaje generado por IA",
	Long: `Analiza los cambios actuales y crea un commit con mensaje convencional.

Qué hace:
1. Analiza cambios staged y unstaged
2. Examina mensajes recientes para estilo
3. Genera mensaje de commit
4. Stagea archivos relevantes
5. Crea el commit

No hace push - usa /commit-push-pr para eso.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get git status
		statusCmd := exec.Command("git", "status", "--porcelain")
		statusOut, err := statusCmd.Output()
		if err != nil {
			output.Fatal(output.ExitGeneric, "No es un repositorio git", "error: %v", err)
		}

		status := string(statusOut)
		if status == "" {
			output.Info("No hay cambios para commit")
			return
		}

		// Get diff
		diffCmd := exec.Command("git", "diff", "--staged", "HEAD")
		diffOut, err := diffCmd.Output()
		if err != nil {
			diffOut = []byte("")
		}

		// Get recent commits for style
		logCmd := exec.Command("git", "log", "--oneline", "-5")
		logOut, err := logCmd.Output()
		if err != nil {
			logOut = []byte("")
		}

		// Analyze and create commit message
		msg := analyzeCommitMessage(status, string(diffOut), string(logOut))

		// Stage all
		addCmd := exec.Command("git", "add", "-A")
		if err := addCmd.Run(); err != nil {
			output.Warning("Algunos archivos no pudieron ser stageados: %v", err)
		}

		// Commit
		commitCmd := exec.Command("git", "commit", "-m", msg)
		if err := commitCmd.Run(); err != nil {
			output.Fatal(output.ExitGeneric, "No se pudo hacer commit", "error: %v", err)
		}

		output.Success("Commit creado: %s", msg)

		// Show status
		exec.Command("git", "status", "--short").Run()
	},
}

func analyzeCommitMessage(status, diff, log string) string {
	// Simple heuristic for commit message
	var changes []string
	for _, line := range strings.Split(status, "\n") {
		if line == "" {
			continue
		}
		statusCode := strings.TrimSpace(line[:2])
		file := strings.TrimSpace(line[3:])

		switch statusCode {
		case "M", "A":
			changes = append(changes, file)
		case "D":
			changes = append(changes, "removed "+file)
		default:
			changes = append(changes, file)
		}
	}

	if len(changes) == 0 {
		return "Update"
	}

	// Generate conventional commit message
	var verb string
	if strings.Contains(status, "A") {
		verb = "Add"
	} else if strings.Contains(status, "D") {
		verb = "Remove"
	} else if strings.Contains(status, "M") {
		verb = "Update"
	} else {
		verb = "Modify"
	}

	// Summarize main change
	mainChange := changes[0]
	if len(changes) > 1 {
		mainChange += fmt.Sprintf(" and %d more files", len(changes)-1)
	}

	return fmt.Sprintf("%s: %s", verb, mainChange)
}

func init() {
	rootCmd.AddCommand(commitCmd)
}
