package cmd

import (
	"juarvis/pkg/output"
	"juarvis/pkg/root"
	"juarvis/pkg/sync"

	"github.com/spf13/cobra"
)

var (
	syncProvider string
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Actualiza el ecosistema local con la versión del binario",
	Long:  `Compara los archivos del proyecto (plugins, configs) con los assets embebidos en el binario y actualiza los que hayan cambiado o falten.

Cloud sync: --provider=gist sincroniza memoria con GitHub Gist,
           --provider=local solo sincroniza assets locales (default).`,
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := root.GetRoot()
		if err != nil {
			output.Fatal(output.ExitNoEcosystem,
				"Ejecuta 'juarvis init' para crear el ecosistema primero",
				"%v", err)
		}

		if err := sync.RunSync(rootPath, syncProvider); err != nil {
			output.Fatal(output.ExitGeneric,
				"Verifica que tienes permisos de escritura en el directorio del proyecto",
				"Error sincronizando: %v", err)
		}
	},
}

var cloudSyncCmd = &cobra.Command{
	Use:   "cloud sync",
	Short: "Sincroniza memoria con la cloud (Gist)",
	Long:  `Sincroniza la memoria local con un proveedor externo.
Proveedores soportados: gist (GitHub Gist), local (default).`,
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := root.GetRoot()
		if err != nil {
			output.Fatal(output.ExitNoEcosystem, "Ejecuta 'juarvis init' primero", "%v", err)
		}

		if err := sync.RunCloudSync(rootPath, syncProvider); err != nil {
			output.Fatal(output.ExitGeneric, "Error en cloud sync", "%v", err)
		}
	},
}

func init() {
	syncCmd.Flags().StringVar(&syncProvider, "provider", "local", "Proveedor de sync: local, gist, custom")
	cloudSyncCmd.Flags().StringVar(&syncProvider, "provider", "local", "Proveedor de sync: local, gist")
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(cloudSyncCmd)
}
