package cmd

import (
	"juarvis/pkg/loader"
	"juarvis/pkg/output"
	"os"

	"github.com/spf13/cobra"
)

var loaderCmd = &cobra.Command{
	Use:   "load",
	Short: "Ejecuta el cargador de plugins y regenera enlaces dinámicos",
	Run: func(cmd *cobra.Command, args []string) {
		if err := loader.RunLoader(""); err != nil {
			output.Error("Error crítico en el cargador: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(loaderCmd)
}
