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
		// Pasamos verify.VerifyOptions{} para que ejecute todas las pruebas
		results, err := verify.RunVerify(verify.VerifyOptions{})
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
	rootCmd.AddCommand(verifyCmd)
}

