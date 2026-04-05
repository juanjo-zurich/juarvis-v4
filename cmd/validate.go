package cmd

import (
	"juarvis/pkg/output"
	"juarvis/pkg/validate"
	"os"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Ejecuta un health-check del ecosistema (Sustituye juarvis-validate check)",
	Run: func(cmd *cobra.Command, args []string) {
		if err := validate.RunHealthCheck(); err != nil {
			output.Error("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
