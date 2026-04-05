package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"juarvis/pkg/memory"
)

var memoryCmd = &cobra.Command{
	Use:    "memory",
	Short:  "Servidor MCP de memoria persistente del agente",
	Long:   `Servidor MCP que gestiona la memoria persistente del agente usando el filesystem local (.juar/memory/). Se invoca via stdio desde OpenCode.`,
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := getMemoryRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		if err := memory.ServeStdio(rootPath); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
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
