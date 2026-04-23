package cmd

import (
	"juarvis/pkg/output"
	"juarvis/pkg/pm"

	"github.com/spf13/cobra"
)

var pmCmd = &cobra.Command{
	Use:   "pm",
	Short: "Gestor de Paquetes (Marketplace) de Juarvis",
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista de plugins disponibles en tu catálogo",
	Run: func(cmd *cobra.Command, args []string) {
		pm.ListPlugins()
	},
}

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Busca en el directorio global (skills.sh) plugins de cualquier proveedor",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pm.SearchPlugins(args[0])
	},
}

var enableCmd = &cobra.Command{
	Use:   "enable [plugin]",
	Short: "Habilita un plugin instalado",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pName := args[0]
		if err := pm.SetPluginStatus(pName, true); err != nil {
			output.Error("Error al habilitar: %v", err)
		} else {
			output.Success("Plugin '%s' habilitado exitosamente.", pName)
		}
	},
}

var disableCmd = &cobra.Command{
	Use:   "disable [plugin]",
	Short: "Deshabilita un plugin para que no se carge en Juarvis",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pName := args[0]
		if err := pm.SetPluginStatus(pName, false); err != nil {
			output.Error("Error al deshabilitar: %v", err)
		} else {
			output.Info("Plugin '%s' ha sido deshabilitado.", pName)
		}
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove [plugin]",
	Short: "Elimina por completo un plugin del sistema",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pName := args[0]
		if err := pm.RemovePlugin(pName); err != nil {
			output.Error("Error al eliminar: %v", err)
		} else {
			output.Success("Plugin '%s' eliminado completamente de Juarvis.", pName)
		}
	},
}

var installCmd = &cobra.Command{
	Use:   "install [plugin]",
	Short: "Instala un plugin desde el marketplace",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pluginName := args[0]

		output.Info("Instalando plugin '%s'...", pluginName)

		if err := pm.InstallPlugin(pluginName); err != nil {
			output.Fatal(output.ExitPluginError,
				"Ejecuta 'juarvis pm search "+pluginName+"' para verificar que el plugin existe",
				"Error instalando plugin: %v", err)
		}

		output.Success("Plugin '%s' instalado correctamente", pluginName)
		output.Info("Ejecuta 'juarvis load' para indexar el nuevo plugin")
	},
}

var checkUpdatesCmd = &cobra.Command{
	Use:   "check",
	Short: "Busca actualizaciones disponibles para los plugins instalados",
	Run: func(cmd *cobra.Command, args []string) {
		if err := pm.CheckUpdates(); err != nil {
			output.Error("Error verificando actualizaciones: %v", err)
		}
	},
}

var updateCmd = &cobra.Command{
	Use:   "update [plugin]",
	Short: "Actualiza un plugin a la última versión",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pName := args[0]
		if err := pm.UpdatePlugin(pName); err != nil {
			output.Error("Error actualizando plugin: %v", err)
		} else {
			output.Success("Plugin '%s' actualizado correctamente", pName)
		}
	},
}

var updateAllCmd = &cobra.Command{
	Use:   "update-all",
	Short: "Actualiza todos los plugins instalados a su última versión",
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")
		updated, err := pm.UpdateAllPlugins(force)
		if err != nil {
			output.Error("Error actualizando plugins: %v", err)
		} else {
			output.Success("%d plugin(s) actualizado(s)", updated)
		}
	},
}

func init() {
	updateAllCmd.Flags().BoolP("force", "f", false, "Fuerza actualización incluso si ya está actualizado")
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback [plugin]",
	Short: "Revierte un plugin a la versión anterior",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pName := args[0]
		if err := pm.RollbackPlugin(pName); err != nil {
			output.Error("Error en rollback: %v", err)
		} else {
			output.Success("Plugin '%s' revertido", pName)
		}
	},
}

func init() {
	pmCmd.AddCommand(listCmd)
	pmCmd.AddCommand(searchCmd)
	pmCmd.AddCommand(enableCmd)
	pmCmd.AddCommand(disableCmd)
	pmCmd.AddCommand(removeCmd)
	pmCmd.AddCommand(installCmd)
	pmCmd.AddCommand(checkUpdatesCmd)
	pmCmd.AddCommand(updateCmd)
	pmCmd.AddCommand(updateAllCmd)
	pmCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(pmCmd)
}
