package cmd

import (
	"juarvis/pkg/output"
	"juarvis/pkg/setup"
	"os"

	"github.com/spf13/cobra"
)

var ideTarget string
var useGui bool

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Distribuye configuraciones, reglas y skills de Juarvis a tu IDE (Sustituye setup.sh)",
	Run: func(cmd *cobra.Command, args []string) {
		if useGui {
			if cmd.Flags().Changed("ide") {
				output.Warning("El flag --ide se ignora en modo GUI. Selecciona el IDE desde la interfaz web.")
			}
			if err := setup.RunServer(); err != nil {
				output.Error("Fallo en el servidor GUI: %v", err)
				os.Exit(1)
			}
			return
		}
		if err := setup.RunSetup(ideTarget); err != nil {
			output.Error("Fallo en la distribución: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	setupCmd.Flags().StringVarP(&ideTarget, "ide", "i", "all", "Entorno destino: opencode, cursor, windsurf, all")
	setupCmd.Flags().BoolVar(&useGui, "gui", false, "Inicia la interfaz gráfica de configuración en el navegador")
	rootCmd.AddCommand(setupCmd)
}
