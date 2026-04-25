package cmd

import (
	"juarvis/pkg/output"
	"juarvis/pkg/validate"

	"github.com/spf13/cobra"
)

var healthCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Ejecuta un health-check del ecosistema (Sustituye juarvis-validate check)",
	Run: func(cmd *cobra.Command, args []string) {
		if err := validate.RunHealthCheck(); err != nil {
			output.Fatal(output.ExitBuildFailed,
				"Ejecuta 'juarvis doctor' para un diagnóstico detallado",
				"%v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(healthCheckCmd)
}
