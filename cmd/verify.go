package cmd

import (
	"fmt"
	"os"

	"juarvis/pkg/verify"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Ejecuta verificaciones de salud del ecosistema",
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
			fmt.Printf("❌ Error al ejecutar verify: %v\n", err)
			os.Exit(1)
		}

		passed := true
		for _, res := range results {
			status := "✅"
			if !res.Passed {
				status = "❌"
				passed = false
			}
			fmt.Printf("%s %-20s: %s\n", status, res.Name, res.Message)
		}

		if !passed {
			os.Exit(1)
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

