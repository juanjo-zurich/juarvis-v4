package cmd

import (
	"juarvis/pkg/output"
	"juarvis/pkg/setup"

	"github.com/spf13/cobra"
)

var ideTarget string
var useGui bool
var setupAll bool

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Distribuye configuraciones, reglas y skills de Juarvis a tu IDE (Sustituye setup.sh)",
	Run: func(cmd *cobra.Command, args []string) {
		if useGui {
			if cmd.Flags().Changed("ide") || cmd.Flags().Changed("all") {
				output.Warning("El flag --ide/--all se ignora en modo GUI. Selecciona el IDE desde la interfaz web.")
			}
			if err := setup.RunServer(); err != nil {
				output.Fatal(output.ExitGeneric,
					"Comprueba que el puerto 3000 no está en uso con: lsof -i :3000",
					"Fallo en el servidor GUI: %v", err)
			}
			return
		}
		target := ideTarget
		if setupAll {
			target = "all"
		} else if target == "" {
			target = "all"
		}
		if err := setup.RunSetup(target); err != nil {
			output.Fatal(output.ExitConfigError,
				"Verifica que el ecosistema está inicializado con 'juarvis check'",
				"Fallo en la distribución: %v", err)
		}
	},
}

func init() {
	setupCmd.Flags().StringVarP(&ideTarget, "ide", "i", "", "Entorno destino: opencode, cursor, windsurf, vscode, antigravity, trae, kiro")
	setupCmd.Flags().BoolVar(&setupAll, "all", false, "Distribuir a TODOS los IDEs soportados")
	setupCmd.Flags().BoolVar(&useGui, "gui", false, "Inicia la interfaz gráfica de configuración en el navegador")
	rootCmd.AddCommand(setupCmd)
}
