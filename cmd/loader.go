package cmd

import (
	"juarvis/pkg/loader"
	"juarvis/pkg/output"

	"github.com/spf13/cobra"
)

var loaderCmd = &cobra.Command{
	Use:   "load",
	Short: "Ejecuta el cargador de plugins y regenera enlaces dinámicos",
	Run: func(cmd *cobra.Command, args []string) {
		if err := loader.RunLoader(""); err != nil {
			output.Fatal(output.ExitPluginError,
				"Ejecuta 'juarvis check' para diagnosticar el ecosistema",
				"Error crítico en el cargador: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(loaderCmd)
}
