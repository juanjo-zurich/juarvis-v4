package cmd

import (
	"juarvis/pkg/output"
	"juarvis/pkg/root"
	"juarvis/pkg/sync"
	"os"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Actualiza el ecosistema local con la versión del binario",
	Long:  `Compara los archivos del proyecto (plugins, configs) con los assets embebidos en el binario y actualiza los que hayan cambiado o falten.`,
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := root.GetRoot()
		if err != nil {
			output.Error("%v", err)
			os.Exit(1)
		}

		if err := sync.RunSync(rootPath); err != nil {
			output.Error("Error sincronizando: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
