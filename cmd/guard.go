package cmd

import (
	"context"
	"os"
	"os/exec"

	"juarvis/pkg/output"
	"juarvis/pkg/root"
	"juarvis/pkg/security"

	"github.com/spf13/cobra"
)

var guardCmd = &cobra.Command{
	Use:   "guard",
	Short: "Seguridad unificada de Juarvis",
	Long:  `Sistema de seguridad unificado con tres capas: sandbox, permissions, hookify.`,
}

// guardRunCmd: juarvis guard run <command>
var guardRunCmd = &cobra.Command{
	Use:   "run [command...]",
	Short: "Ejecutar comando con el gate de seguridad",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, _ := root.GetRoot()

		gate, err := security.NewSecurityGate(rootPath)
		if err != nil {
			output.Error("Error creando gate: %v", err)
			return
		}

		// Evaluar todas las capas
		ctx := context.Background()
		result := gate.Eval(ctx, args[0], args[1:])

		if !result.Allowed {
			output.Error("[%s] %s", result.Layer, result.Message)
			return
		}

		output.Success("Comando permitido")

		// Ejecutar
		execCmd := exec.Command(args[0], args[1:]...)
		execCmd.Dir = rootPath
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		if err := execCmd.Run(); err != nil {
			output.Error("Comando falló: %v", err)
		}
	},
}

// guardCheckCmd: juarvis guard check <command>
var guardCheckCmd = &cobra.Command{
	Use:   "check [command]",
	Short: "Verificar si un comando pasa las tres capas",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, _ := root.GetRoot()

		gate, err := security.NewSecurityGate(rootPath)
		if err != nil {
			output.Error("Error creando gate: %v", err)
			return
		}

		ctx := context.Background()
		result := gate.Eval(ctx, args[0], nil)

		if !result.Allowed {
			output.Warning("[%s] %s", result.Layer, result.Message)
			return
		}

		output.Success("✓ Todas las capas permitieron: %s", args[0])
	},
}

// guardDisableCmd: juarvis guard disable <layer>
var guardDisableCmd = &cobra.Command{
	Use:   "disable [layer]",
	Short: "Desactivar una capa de seguridad",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, _ := root.GetRoot()

		gate, err := security.NewSecurityGate(rootPath)
		if err != nil {
			output.Error("Error creando gate: %v", err)
			return
		}

		layer := args[0]
		valid := map[string]bool{"sandbox": true, "permissions": true, "hookify": true}
		if !valid[layer] {
			output.Error("Capa inválida. Válidas: sandbox, permissions, hookify")
			return
		}

		gate.DisableLayer(layer)
		output.Success("Capa desactivada: %s", layer)
	},
}

// guardEnableCmd: juarvis guard enable <layer>
var guardEnableCmd = &cobra.Command{
	Use:   "enable [layer]",
	Short: "Activar una capa de seguridad",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, _ := root.GetRoot()

		gate, err := security.NewSecurityGate(rootPath)
		if err != nil {
			output.Error("Error creando gate: %v", err)
			return
		}

		layer := args[0]
		valid := map[string]bool{"sandbox": true, "permissions": true, "hookify": true}
		if !valid[layer] {
			output.Error("Capa inválida. Válidas: sandbox, permissions, hookify")
			return
		}

		gate.EnableLayer(layer)
		output.Success("Capa activada: %s", layer)
	},
}

func init() {
	guardRunCmd.Flags().BoolP("apply-fix", "f", false, "Aplicar auto-fix si está disponible")

	guardCmd.AddCommand(guardRunCmd)
	guardCmd.AddCommand(guardCheckCmd)
	guardCmd.AddCommand(guardDisableCmd)
	guardCmd.AddCommand(guardEnableCmd)
	rootCmd.AddCommand(guardCmd)
}