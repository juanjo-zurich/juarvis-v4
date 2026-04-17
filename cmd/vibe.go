package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"juarvis/pkg/config"
	"juarvis/pkg/output"
	"juarvis/pkg/root"

	"github.com/spf13/cobra"
)

var vibeCmd = &cobra.Command{
	Use:   "vibe",
	Short: "Vibe Check: Evalúa la salud creativa y el flujo de tu proyecto",
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := root.GetRoot()
		if err != nil {
			output.Error("No estás en un ecosistema Juarvis. ¡Haz un 'juarvis up' para empezar!")
			return
		}

		output.Banner("VIBE CHECK - ESTADO DEL FLUJO")

		// 1. Skills
		skillsCount := 0
		entries, _ := os.ReadDir(filepath.Join(rootPath, "plugins"))
		for _, e := range entries {
			if e.IsDir() {
				skillsCount++
			}
		}
		output.Info("Capacidades: %s", output.Styled("cyan", "%d plugins activos", skillsCount))

		// 2. Snapshots
		snapshotsCount := 0
		out, err := exec.Command("git", "stash", "list").CombinedOutput()
		if err == nil {
			snapshotsCount = strings.Count(string(out), "juarvis-snapshot|")
		}
		output.Info("Seguridad: %s", output.Styled("green", "%d snapshots guardados", snapshotsCount))

		// 3. Health
		output.Info("Salud Técnica:")
		buildCmd := exec.Command("go", "build", "./...")
		if err := buildCmd.Run(); err != nil {
			fmt.Printf("  - Build: %s\n", output.Styled("red", "ROTO ❌"))
		} else {
			fmt.Printf("  - Build: %s\n", output.Styled("green", "SANO ✅"))
		}

		testCmd := exec.Command("go", "test", "-short", "./...")
		if err := testCmd.Run(); err != nil {
			fmt.Printf("  - Tests: %s\n", output.Styled("yellow", "FALLANDO ⚠️"))
		} else {
			fmt.Printf("  - Tests: %s\n", output.Styled("green", "PASANDO ✅"))
		}

		// 4. Memory
		memFile := filepath.Join(rootPath, config.JuarDir, config.SkillRegistryFile)
		if info, err := os.Stat(memFile); err == nil {
			output.Info("Memoria: %s", output.Styled("purple", "%d bytes de conocimiento persistente", info.Size()))
		}

		fmt.Println()
		if snapshotsCount > 0 && skillsCount > 5 {
			output.Success("¡LA VIBRA ES EXCELENTE! Estás en el flow creativo. 🚀")
		} else {
			output.Warning("Sigue construyendo. El flujo está empezando a tomar forma. 🌱")
		}
	},
}

func init() {
	rootCmd.AddCommand(vibeCmd)
}
