package cmd

import (
	"fmt"
	"os"

	initpkg "juarvis/pkg/init"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Inicializa un nuevo ecosistema Juarvis",
	Long: `Crea la estructura base de un ecosistema Juarvis en el directorio especificado.
Si no se especifica path, se usa el directorio actual.

Estructura creada:
  - marketplace.json
  - plugins/juarvis-core/
  - .juar/
`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		if err := initpkg.RunInit(path); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error inicializando ecosistema: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
