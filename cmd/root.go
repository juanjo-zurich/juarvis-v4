package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"juarvis/pkg/output"
)

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

var GlobalRoot string
var GlobalJSON bool

var rootCmd = &cobra.Command{
	Use:   "juarvis",
	Short: "Juarvis V4 CLI - El motor de tu ecosistema de agentes IA",
	Long:  `Juarvis V4 gestiona Spec-Driven Development, integradores MCP y control de paquetes de skills (Marketplace) directamente compilado en Go.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if GlobalRoot != "" {
			os.Setenv("JUARVIS_ROOT", GlobalRoot)
		}
		output.SetJSONMode(GlobalJSON)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&GlobalRoot, "root", "", "Directorio raíz del ecosistema Juarvis (opcional)")
	rootCmd.PersistentFlags().BoolVar(&GlobalJSON, "json", false, "Salida en formato JSON para consumo programatico")
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, BuildDate)
	rootCmd.SetVersionTemplate("juarvis version {{.Version}}\n")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
