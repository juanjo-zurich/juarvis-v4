package cmd

import (
	"os"

	"juarvis/pkg/analyzer"
	"juarvis/pkg/output"

	"github.com/spf13/cobra"
)

var analyzeTranscriptCmd = &cobra.Command{
	Use:   "analyze-transcript [path]",
	Short: "Analiza un transcript de sesión y extrae aprendizajes",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "transcript.md"
		if len(args) > 0 {
			path = args[0]
		}

		output.Info("Analizando transcript: %s", path)

		analysis, err := analyzer.AnalyzeTranscript(path)
		if err != nil {
			output.Fatal(output.ExitGeneric,
				"Verifica que el transcript existe: %v",
				"%v", err)
		}

		summary := analyzer.BuildSessionSummary(analysis)
		output.Info("%s", "\n"+summary)

		// Guardar análisis
		cwd, err := os.Getwd()
		if err != nil {
			output.Warning("No se pudo obtener directorio: %v", err)
			return
		}
		if err := analyzer.SaveAnalysis(analysis, cwd); err != nil {
			output.Warning("Error guardando: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(analyzeTranscriptCmd)
}
