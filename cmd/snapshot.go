package cmd

import (
	"juarvis/pkg/output"
	"juarvis/pkg/snapshot"
	"os"

	"github.com/spf13/cobra"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Motor de Snapshots de seguridad locales para SDD",
}

var snapshotCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Toma un backup instantáneo del repositorio",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		output.Info("Inicializando motor de snapshots locales...")
		if err := snapshot.CreateSnapshot(args[0]); err != nil {
			output.Error("%v", err)
			os.Exit(1)
		}
	},
}

var snapshotRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restaura el último snapshot de seguridad tomado por Juarvis",
	Run: func(cmd *cobra.Command, args []string) {
		if err := snapshot.RestoreLatestSnapshot(); err != nil {
			output.Error("Error al restaurar: %v", err)
			os.Exit(1)
		}
	},
}

var snapshotPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Elimina snapshots antiguos de juarvis",
	Long: `Elimina todos los stashes de juarvis acumulados.
Solo afecta a stashes creados por juarvis (prefijo "juarvis-snapshot|").
Los stashes del usuario no se tocan.`,
	Run: func(cmd *cobra.Command, args []string) {
		pruned, err := snapshot.PruneSnapshots(true)
		if err != nil {
			output.Error("Error eliminando snapshots: %v", err)
			os.Exit(1)
		}

		if pruned == 0 {
			output.Info("No hay snapshots de juarvis para eliminar")
		} else {
			output.Success("%d snapshots eliminados", pruned)
		}
	},
}

func init() {
	snapshotCmd.AddCommand(snapshotCreateCmd)
	snapshotCmd.AddCommand(snapshotRestoreCmd)
	snapshotCmd.AddCommand(snapshotPruneCmd)
	rootCmd.AddCommand(snapshotCmd)
}
