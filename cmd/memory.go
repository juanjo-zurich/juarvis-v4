package cmd

import (
	"github.com/spf13/cobra"
	"juarvis/pkg/memory"
	"juarvis/pkg/output"
	"juarvis/pkg/root"
)

var memoryRoot string

var memoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Servidor MCP de memoria persistente del agente",
	Long: `Servidor MCP que gestiona la memoria persistente del agente usando el filesystem local (.juar/memory/). Se invoca via stdio desde OpenCode.

Ejemplo de uso:
  juarvis memory                    # Usar directorio actual
  juarvis memory --root /path/to/project  # Usar proyecto específico`,
	Hidden: false,
	Run: func(cmd *cobra.Command, args []string) {
		rootPath := memoryRoot
		if rootPath == "" {
			var err error
			rootPath, err = root.GetRoot()
			if err != nil {
				output.Fatal(output.ExitNoEcosystem,
					"Ejecuta 'juarvis init' para inicializar el ecosistema primero",
					"error: %v", err)
			}
		}

		if err := memory.ServeStdio(rootPath); err != nil {
			output.Fatal(output.ExitGeneric,
				"Comprueba los permisos de escritura en el directorio .juar/",
				"error: %v", err)
		}
	},
}

func init() {
	memoryCmd.Flags().StringVar(&memoryRoot, "root", "", "Directorio raíz del ecosistema")
	rootCmd.AddCommand(memoryCmd)
}
