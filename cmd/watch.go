package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"juarvis/pkg/output"
	"juarvis/pkg/root"
	"juarvis/pkg/watcher"

	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Vigila cambios en el proyecto y ejecuta reglas de seguridad automaticamente",
	Long: `Inicia un daemon que monitoriza el sistema de archivos del proyecto.
Cuando detecta cambios, evalua reglas de hookify automaticamente y crea
snapshots de seguridad si se detectan modificaciones masivas.

Esto proporciona un "cinturon de seguridad" automatico que no depende de
la voluntad del agente de IA.`,
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := root.GetRoot()
		if err != nil {
			output.Error("%v", err)
			os.Exit(1)
		}

		debounceMs, _ := cmd.Flags().GetInt("debounce-ms")
		threshold, _ := cmd.Flags().GetInt("auto-snapshot-threshold")
		noAutoSnapshot, _ := cmd.Flags().GetBool("no-auto-snapshot")

		cfg := watcher.DefaultWatcherConfig(rootPath)
		if debounceMs > 0 {
			cfg.DebounceMs = debounceMs
		}
		if threshold > 0 {
			cfg.AutoSnapshotThreshold = threshold
		}
		if noAutoSnapshot {
			cfg.NoAutoSnapshot = true
		}

		w, err := watcher.NewWatcher(cfg)
		if err != nil {
			output.Error("Error iniciando watcher: %v", err)
			os.Exit(1)
		}

		ctx, cancel := context.WithCancel(context.Background())
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			cancel()
		}()

		if err := w.Start(ctx); err != nil {
			output.Error("Error en watcher: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	watchCmd.Flags().Int("debounce-ms", 500, "Ventana de debounce en milisegundos")
	watchCmd.Flags().Int("auto-snapshot-threshold", 5, "Numero de cambios para auto-snapshot")
	watchCmd.Flags().Bool("no-auto-snapshot", false, "Desactivar auto-snapshots")
	rootCmd.AddCommand(watchCmd)
}
