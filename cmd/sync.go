package cmd

import (
	"juarvis/pkg/output"
	"juarvis/pkg/root"
	"juarvis/pkg/sync"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Actualiza el ecosistema local con la versión del binario",
	Long:  `Compara los archivos del proyecto (plugins, configs) con los assets embebidos en el binario y actualiza los que hayan cambiado o falten.`,
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := root.GetRoot()
		if err != nil {
			output.Fatal(output.ExitNoEcosystem,
				"Ejecuta 'juarvis init' para crear el ecosistema primero",
				"%v", err)
		}

		if err := sync.RunSync(rootPath); err != nil {
			output.Fatal(output.ExitGeneric,
				"Verifica que tienes permisos de escritura en el directorio del proyecto",
				"Error sincronizando: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
