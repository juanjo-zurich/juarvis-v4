package cmd

import (
	"github.com/spf13/cobra"
	"juarvis/pkg/config"
	"juarvis/pkg/output"
	"juarvis/pkg/root"
)

var modeCmd = &cobra.Command{
	Use:   "mode [level]",
	Short: "Muestra o cambia el nivel de autonomía (0-4)",
	Long: `Niveles de autonomía:
0 (vibe): Solo memoria + seguridad
1 (seguro): + snapshot automático  
2 (estructurado): + descomposición de tareas (default)
3 (semi): + spec antes de implementar
4 (sdd): Pipeline SDD completo

Ejemplos:
juarvis mode          # Ver nivel actual
juarvis mode 0        # Cambiar a Vibe Puro
juarvis mode sdd       # Cambiar a SDD Completo`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := root.GetRoot()
		if err != nil {
			output.Fatal(output.ExitNoEcosystem,
				"Ejecuta desde un proyecto con juarvis init",
				"Error: %v", err)
		}

		cfg, err := config.LoadOrCreate(rootPath)
		if err != nil {
			output.Fatal(output.ExitGeneric,
				"Error cargando configuración",
				"Error: %v", err)
		}

		// Sin args = mostrar nivel actual
		if len(args) == 0 {
			level := cfg.AutonomyLevel
			output.Info("📊 Nivel de autonomía: %d (%s)%s",
				level,
				config.GetLevelName(level),
				config.LevelDescriptions[level])
			return
		}

		// Con args = cambiar nivel
		level := args[0]

		levels := map[string]int{
			"0": 0, "vibe": 0,
			"1": 1, "seguro": 1,
			"2": 2, "estructurado": 2,
			"3": 3, "semi": 3,
			"4": 4, "sdd": 4,
		}
		newLevel, ok := levels[level]
		if !ok {
			output.Fatal(output.ExitGeneric,
				"Nivel inválido. Usa: 0 (vibe), 1 (seguro), 2 (estructurado), 3 (semi), 4 (sdd)",
				"Nivel '%s' no reconocido", level)
		}

		cfg.AutonomyLevel = newLevel
		if err := cfg.Save(rootPath); err != nil {
			output.Fatal(output.ExitGeneric,
				"Error guardando configuración",
				"Error: %v", err)
		}

		// Actualizar AGENTS.md con el nuevo protocolo
		if err := cfg.UpdateAgentsSection(rootPath); err != nil {
			output.Warning("No se pudo actualizar AGENTS.md: %v", err)
		}

		output.Success("✅ Modo %s activado.%s",
			config.GetLevelName(newLevel),
			config.LevelDescriptions[newLevel])
	},
}

func init() {
	rootCmd.AddCommand(modeCmd)
}
