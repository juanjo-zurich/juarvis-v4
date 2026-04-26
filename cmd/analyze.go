package cmd

import (
	"juarvis/pkg/analyze"
	"juarvis/pkg/output"

	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analiza la codebase y genera skills específicas del proyecto",
	Long: `Analiza el proyecto para entender:
- Stack tecnológico
- Convenciones de código
- Patrones y antipatrones
- Arquitectura del proyecto

Genera skills propias del proyecto que el agente puede usar.`,
	Run: func(cmd *cobra.Command, args []string) {
		update, _ := cmd.Flags().GetBool("update")
		verbose, _ := cmd.Flags().GetBool("verbose")

		if err := analyze.RunAnalyze(update, verbose); err != nil {
			output.Fatal(output.ExitGeneric,
				"Verifica que estás en un proyecto con código fuente",
				"Error analizando proyecto: %v", err)
		}
	},
}

func init() {
	analyzeCmd.Flags().BoolP("update", "u", false, "Actualiza skills existentes (diff only)")
	analyzeCmd.Flags().BoolP("verbose", "v", false, "Salida detallada")
	rootCmd.AddCommand(analyzeCmd)
}
