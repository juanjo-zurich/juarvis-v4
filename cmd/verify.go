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
		skipBuild, _ := cmd.Flags().GetBool("skip-build")
		skipVet, _ := cmd.Flags().GetBool("skip-vet")
		skipTest, _ := cmd.Flags().GetBool("skip-test")
		skipJSON, _ := cmd.Flags().GetBool("skip-json")
		skipPlugins, _ := cmd.Flags().GetBool("skip-plugins")
		skipCLI, _ := cmd.Flags().GetBool("skip-cli")

		opts := verify.VerifyOptions{
			SkipBuild:   skipBuild,
			SkipVet:     skipVet,
			SkipTest:    skipTest,
			SkipJSON:    skipJSON,
			SkipPlugins: skipPlugins,
			SkipCLI:     skipCLI,
		}

		results, err := verify.RunVerify(opts)
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
	verifyCmd.Flags().Bool("skip-build", false, "Saltar verificación de compilación")
	verifyCmd.Flags().Bool("skip-vet", false, "Saltar verificación de go vet")
	verifyCmd.Flags().Bool("skip-test", false, "Saltar verificación de tests")
	verifyCmd.Flags().Bool("skip-json", false, "Saltar verificación de JSONs embebidos")
	verifyCmd.Flags().Bool("skip-plugins", false, "Saltar verificación de manifiestos de plugins")
	verifyCmd.Flags().Bool("skip-cli", false, "Saltar verificación de comandos CLI")

	rootCmd.AddCommand(verifyCmd)
}
