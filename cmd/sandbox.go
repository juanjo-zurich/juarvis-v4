package cmd

import (
	"context"
	"fmt"
	"time"

	"juarvis/pkg/output"
	"juarvis/pkg/root"
	"juarvis/pkg/sandbox"

	"github.com/spf13/cobra"
)

var sandboxCmd = &cobra.Command{
	Use:   "sandbox",
	Short: "Ejecutar comandos en entorno aislado",
	Long:  `Ejecuta comandos en un entorno sandbox para mayor seguridad.`,
}

// sandboxRunCmd: juarvis sandbox run <command>
var sandboxRunCmd = &cobra.Command{
	Use:   "run [command...]",
	Short: "Ejecutar comando en sandbox",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, _ := root.GetRoot()

		output.Info("Ejecutando en sandbox: %v", args)

		// Crear guardrails para seguridad
		g := sandbox.CrearGuardrailsSimple(rootPath)

		// Verificar antes de ejecutar
		if err := g.VerificarComandoBeforeRun(args[0], args[1:]); err != nil {
			output.Error("Comando bloqueado: %v", err)
			return
		}

		// Ejecutar comando
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		result := g.EjecutarComando(ctx, args[0], args[1:], sandbox.WithWorkingDir(rootPath))
		if result.Error != "" {
			output.Error("Error: %s", result.Error)
			return
		}

		if result.Bloqueado {
			output.Warning("Comando bloqueado: %s", result.RazonBloqueo)
			return
		}

		if !result.Exito {
			output.Warning("Comando no completado exitosamente")
		}

		if len(result.Salida) > 0 {
			fmt.Print(result.Salida)
		}

		output.Success("Ejecutado en %v", result.Duracion)
	},
}

// sandboxCheckCmd: juarvis sandbox check <command>
var sandboxCheckCmd = &cobra.Command{
	Use:   "check [command]",
	Short: "Verificar si un comando es seguro",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, _ := root.GetRoot()
		g := sandbox.CrearGuardrailsSimple(rootPath)

		err := g.VerificarComandoBeforeRun(args[0], nil)
		if err != nil {
			output.Warning("Comando NO permitido: %s - %v", args[0], err)
		} else {
			output.Success("Comando permitido: %s", args[0])
		}
	},
}

// sandboxStatsCmd: juarvis sandbox stats
var sandboxStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Mostrar estadísticas del sandbox",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, _ := root.GetRoot()
		g := sandbox.CrearGuardrailsSimple(rootPath)

		output.Info("Estadísticas del sandbox:")
		fmt.Printf("  Ejecuciones: %d\n", g.ObtenerContadorEjecuciones())

		cfg := g.ObtenerConfig()
		fmt.Printf("  Nivel: %s\n", cfg.Nivel)
		fmt.Printf("  Habilitado: %v\n", cfg.Enabled)
	},
}

func init() {
	sandboxCmd.AddCommand(sandboxRunCmd)
	sandboxCmd.AddCommand(sandboxCheckCmd)
	sandboxCmd.AddCommand(sandboxStatsCmd)
	rootCmd.AddCommand(sandboxCmd)
}