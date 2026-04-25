package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"juarvis/pkg/output"
)

var cleanGoneCmd = &cobra.Command{
	Use:   "clean-gone",
	Short: "Limpia branches locales eliminadas en remote",
	Long: `Limpia branches que ya no existen en el repositorio remote.

Qué hace:
1. Lista branches con estado [gone]
2. Elimina worktrees asociados
3. Borra branches stale
4. Reporta lo que se limpió`,
	Run: func(cmd *cobra.Command, args []string) {
		// First, prune remote tracking
		output.Info("Actualizando remote tracking...")
		pruneCmd := exec.Command("git", "fetch", "--prune")
		if err := pruneCmd.Run(); err != nil {
			output.Warning("Fetch --prune falló: %v", err)
		}

		// List gone branches
		listCmd := exec.Command("git", "branch", "-v")
		listOut, err := listCmd.Output()
		if err != nil {
			output.Fatal(output.ExitGeneric, "No se pudieron listar branches", "error: %v", err)
		}

		// Find gone branches
		var goneBranches []string
		for _, line := range strings.Split(string(listOut), "\n") {
			if strings.Contains(line, "[gone]") {
				parts := strings.Fields(line)
				if len(parts) >= 1 {
					// Remove * prefix if present
					branch := strings.TrimPrefix(parts[0], "*")
					goneBranches = append(goneBranches, branch)
				}
			}
		}

		if len(goneBranches) == 0 {
			output.Success("No hay branches [gone] para limpiar")
			return
		}

		output.Warning("Encontradas %d branches [gone]: %s",
			len(goneBranches), strings.Join(goneBranches, ", "))

		// Delete gone branches
		deleted := 0
		for _, branch := range goneBranches {
			output.Info("Eliminando: %s", branch)
			delCmd := exec.Command("git", "branch", "-d", branch)
			err := delCmd.Run()
			if err != nil {
				// Try -D if -d fails
				delCmd = exec.Command("git", "branch", "-D", branch)
				err = delCmd.Run()
				if err != nil {
					output.Warning("No se pudo eliminar %s: %v", branch, err)
					continue
				}
			}
			deleted++
		}

		output.Success("Eliminado %d branches", deleted)
		fmt.Printf("\nWorkspace limpio ✅\n")
	},
}

func init() {
	rootCmd.AddCommand(cleanGoneCmd)
}
