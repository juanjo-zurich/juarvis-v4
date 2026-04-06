package cmd

import (
	"juarvis/pkg/output"
	"juarvis/pkg/verify"

	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verifica que el proyecto Juarvis está sano",
	Long:  `Ejecuta verificaciones de build, tests, validación de configs embebidas y comandos CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		results, err := verify.RunVerify()
		if err != nil {
			output.Error("Error ejecutando verificación: %v", err)
			return
		}

		allPassed := true
		for _, r := range results {
			if r.Passed {
				output.Success("%s: %s", r.Name, r.Message)
			} else {
				output.Error("%s: %s", r.Name, r.Message)
				allPassed = false
			}
		}

		if allPassed {
			output.Success("Todas las verificaciones pasaron")
		} else {
			output.Error("Algunas verificaciones fallaron")
		}
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}
