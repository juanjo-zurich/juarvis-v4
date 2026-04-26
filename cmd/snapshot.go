package cmd

import (
	"juarvis/pkg/output"
	"juarvis/pkg/snapshot"

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
			output.Fatal(output.ExitGeneric,
				"Verifica que estás en un repositorio git con 'git status'",
				"%v", err)
		}
	},
}

var snapshotRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restaura el último snapshot de seguridad tomado por Juarvis",
	Run: func(cmd *cobra.Command, args []string) {
		if err := snapshot.RestoreLatestSnapshot(); err != nil {
			output.Fatal(output.ExitGeneric,
				"Ejecuta 'git stash list' para ver los snapshots disponibles",
				"Error al restaurar: %v", err)
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
			output.Fatal(output.ExitGeneric,
				"Verifica que estás en un repositorio git con 'git status'",
				"Error eliminando snapshots: %v", err)
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
