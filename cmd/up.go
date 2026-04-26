package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	pkginit "juarvis/pkg/init"
	"juarvis/pkg/output"
	"juarvis/pkg/root"
	"juarvis/pkg/setup"
	"juarvis/pkg/watcher"

	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Vibecoding Start: Inicializa, configura e inicia el watcher de una vez",
	Long:  `El comando definitivo para empezar a trabajar. Ejecuta init, setup --all y watch de forma secuencial.`,
	Run: func(cmd *cobra.Command, args []string) {
		output.Banner("JUARVIS UP - ENTRANDO EN EL FLOW")

		// 1. Init
		output.Info("Paso 1/3: Inicializando ecosistema...")
		if err := pkginit.RunInit(""); err != nil {
			output.Fatal(output.ExitGeneric,
				"Fallo en la inicialización",
				"Error: %v", err)
		}
		output.Success("Ecosistema listo.")

		// 2. Setup
		output.Info("Paso 2/3: Configurando todos los IDEs...")
		if err := setup.RunSetup("all"); err != nil {
			output.Warning("Algunos IDEs no pudieron ser configurados: %v", err)
		} else {
			output.Success("IDEs configurados.")
		}

		// 3. Watch
		output.Info("Paso 3/3: Arrancando el Watcher...")
		rootPath, err := root.GetRoot()
		if err != nil {
			output.Fatal(output.ExitNoEcosystem,
				"Ejecuta 'juarvis init' primero",
				"Error: %v", err)
		}

		cfg := watcher.DefaultWatcherConfig(rootPath)
		w, err := watcher.NewWatcher(cfg)
		if err != nil {
			output.Fatal(output.ExitWatcherError,
				"Error iniciando watcher",
				"Error: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			cancel()
		}()

		output.Banner("WATCHER ACTIVO - ¡A PROGRAMAR!")
		if err := w.Start(ctx); err != nil {
			output.Fatal(output.ExitWatcherError,
				"Error en watcher",
				"Error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
}
