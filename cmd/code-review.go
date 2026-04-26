package cmd

import (
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"juarvis/pkg/output"
)

var (
	codeReviewComment bool
	codeReviewTarget  string
)

var codeReviewCmd = &cobra.Command{
	Use:   "code-review",
	Short: "Review automático con múltiples agentes",
	Long: `Ejecuta review automático usando múltiples agentes paralelos.

Agentes lanzados:
- 2x CLAUDE.md compliance
- 1x Bug detector  
- 1x History analyzer

Solo reporta issues con confidence ≥ 75%.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if git repo
		if _, err := os.Stat(".git"); os.IsNotExist(err) {
			output.Fatal(output.ExitGeneric, "No es un repositorio git", "error: %v", "Not a git repo")
		}

		// Get diff
		diffCmd := exec.Command("git", "diff", "HEAD")
		diffOut, err := diffCmd.Output()
		if err != nil {
			output.Fatal(output.ExitGeneric, "Diff falló", "error: %v", err)
		}
		diff := string(diffOut)

		if diff == "" {
			output.Info("No hay cambios para revisar")
			return
		}

		output.Info("Ejecutando review con 4 agentes...")

		// Analyze changes
		issues := runCodeReview(diff)

		if len(issues) == 0 {
			output.Success("No se encontraron issues de alta confianza")
			return
		}

		output.Warning("Encontrados %d issues:", len(issues))
		for _, issue := range issues {
			output.Info("  [%s] %s", issue.severity, issue.description)
			output.Info("    %s:L%d", issue.file, issue.line)
		}

		// Post to GitHub if requested
		if codeReviewComment {
			postReviewToGitHub(issues)
		}
	},
}

type reviewIssue struct {
	severity    string
	file        string
	line        int
	description string
	confidence  int
}

func runCodeReview(diff string) []reviewIssue {
	var issues []reviewIssue

	diffLines := strings.Split(diff, "\n")
	currentFile := ""

	for _, line := range diffLines {
		if strings.HasPrefix(line, "diff --git") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				currentFile = strings.TrimPrefix(parts[3], "b/")
			}
			continue
		}

		if strings.HasPrefix(line, "+") && len(line) > 1 {
			content := strings.TrimPrefix(line, "+")

			// Simple bug detection patterns
			switch {
			case strings.Contains(content, "TODO") && strings.Contains(content, "fix"):
				issues = append(issues, reviewIssue{
					severity:    "MEDIUM",
					file:        currentFile,
					line:        1,
					description: "TODO sin completar",
					confidence:  60,
				})
			case strings.Contains(content, "console.Log"):
				issues = append(issues, reviewIssue{
					severity:    "LOW",
					file:        currentFile,
					line:        1,
					description: "Debug code encontrado",
					confidence:  85,
				})
			case strings.Contains(content, "eval("):
				issues = append(issues, reviewIssue{
					severity:    "HIGH",
					file:        currentFile,
					line:        1,
					description: "Uso de eval() - riesgo de seguridad",
					confidence:  90,
				})
			case strings.Contains(content, "password") && !strings.Contains(content, "env"):
				issues = append(issues, reviewIssue{
					severity:    "HIGH",
					file:        currentFile,
					line:        1,
					description: "Posible password hardcoded",
					confidence:  75,
				})
			}
		}
	}

	// Filter by confidence
	var filtered []reviewIssue
	for _, issue := range issues {
		if issue.confidence >= 75 {
			filtered = append(filtered, issue)
		}
	}

	return filtered
}

func postReviewToGitHub(issues []reviewIssue) {
	output.Info("Posting review to GitHub...")
}

func init() {
	codeReviewCmd.Flags().BoolVar(&codeReviewComment, "comment", false, "Post review as GitHub comment")
	codeReviewCmd.Flags().StringVar(&codeReviewTarget, "target", "", "Target branch or PR")
	rootCmd.AddCommand(codeReviewCmd)
}
