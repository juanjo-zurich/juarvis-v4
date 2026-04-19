package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"juarvis/pkg/memory"
	"juarvis/pkg/output"
)

var memoryCmd = &cobra.Command{
	Use:    "memory",
	Short:  "Servidor MCP de memoria persistente del agente",
	Long:   `Servidor MCP que gestiona la memoria persistente del agente usando el filesystem local (.juar/memory/). Se invoca via stdio desde OpenCode.`,
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := getMemoryRoot()
		if err != nil {
			output.Fatal(output.ExitNoEcosystem,
				"Ejecuta 'juarvis init' para inicializar el ecosistema primero",
				"error: %v", err)
		}

		if err := memory.ServeStdio(rootPath); err != nil {
			output.Fatal(output.ExitGeneric,
				"Comprueba los permisos de escritura en el directorio .juar/",
				"error: %v", err)
		}
	},
}

func getMemoryRoot() (string, error) {
	if root := os.Getenv("JUARVIS_ROOT"); root != "" {
		return root, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("no se pudo obtener directorio actual: %w", err)
	}
	return cwd, nil
}

func init() {
	rootCmd.AddCommand(memoryCmd)
}
