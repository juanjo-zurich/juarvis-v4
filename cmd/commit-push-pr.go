package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"juarvis/pkg/output"
)

var cppDryRun bool

var commitPushPrCmd = &cobra.Command{
	Use:   "commit-push-pr",
	Short: "Commit, push y crea PR en un paso",
	Long: `Workflow completo: commit + push + PR.

Qué hace:
1. Crea nueva branch si está en main
2. Commit con mensaje convencional
3. Push a origin
4. Crea PR con gh pr create

Requiere:
- gh instalado y autenticado
- Repositorio con remote origin`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check git
		if _, err := exec.LookPath("gh"); err != nil {
			output.Fatal(output.ExitGeneric,
				"GitHub CLI no está instalado",
				"Instala con: brew install gh")
		}

		// Get current branch
		branchCmd := exec.Command("git", "branch", "--show-current")
		branch, _ := branchCmd.Output()
		currentBranch := strings.TrimSpace(string(branch))

		// Create branch if on main
		newBranch := ""
		if currentBranch == "main" || currentBranch == "master" {
			newBranch = fmt.Sprintf("feature/auto-%d", 1)
			output.Info("Creando branch: %s", newBranch)
			
			checkoutCmd := exec.Command("git", "checkout", "-b", newBranch)
			if err := checkoutCmd.Run(); err != nil {
				output.Fatal(output.ExitGeneric, "No se pudo crear branch", "error: %v", err)
			}
		} else {
			newBranch = currentBranch
		}

		// Run /commit logic
		statusCmd := exec.Command("git", "status", "--porcelain")
		statusOut, _ := statusCmd.Output()
		status := string(statusOut)
		
		if status == "" {
			output.Info("No hay cambios")
			return
		}

		// Stage all
		exec.Command("git", "add", "-A").Run()

		// Generate message
		msg := analyzeCommitMessage(status, "", "")
		
		// Commit
		commitCmd := exec.Command("git", "commit", "-m", msg)
		if err := commitCmd.Run(); err != nil {
			output.Fatal(output.ExitGeneric, "Commit falló", "error: %v", err)
		}
		output.Success("Commit: %s", msg)

		// Push
		pushCmd := exec.Command("git", "push", "-u", "origin", newBranch)
		if err := pushCmd.Run(); err != nil {
			output.Fatal(output.ExitGeneric, "Push falló", "error: %v", err)
		}
		output.Success("Push a origin/%s", newBranch)

		// Create PR
		prCmd := exec.Command("gh", "pr", "create",
			"--fill", 
			"--title", msg,
			"--body", "Automated PR via juarvis")
		prOut, err := prCmd.Output()
		if err != nil {
			output.Warning("PR no creado (puede que ya exista): %v", err)
			return
		}

		output.Success("PR creado a las %s", strings.TrimSpace(string(prOut)))
	},
}

func init() {
	rootCmd.AddCommand(commitPushPrCmd)
}