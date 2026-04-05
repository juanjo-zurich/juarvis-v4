package cmd

import (
	"fmt"
	"juarvis/pkg/loader"
	"juarvis/pkg/output"
	"juarvis/pkg/root"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var loaderCmd = &cobra.Command{
	Use:   "load",
	Short: "Ejecuta el cargador de plugins y regenera enlaces dinámicos",
	Run: func(cmd *cobra.Command, args []string) {
		if err := loader.RunLoader(); err != nil {
			output.Error("Error crítico en el cargador: %v", err)
			os.Exit(1)
		}
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Alias para load (prometido en AGENTS.md)",
	Run: func(cmd *cobra.Command, args []string) {
		output.Info("Sincronizando (alias de load)...")
		if err := loader.RunLoader(); err != nil {
			output.Error("Error crítico en el cargador: %v", err)
			os.Exit(1)
		}
	},
}

var skillCreateCmd = &cobra.Command{
	Use:   "skill-create [name]",
	Short: "Crea la plantilla base para una nueva skill",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		output.Info("Creando andamiaje para la skill '%s'...", name)

		rootPath, err := root.GetRoot()
		if err != nil {
			output.Error("%v", err)
			os.Exit(1)
		}
		pluginFolder := "juarvis-" + name
		pluginPath := filepath.Join(rootPath, "plugins", pluginFolder)
		skillsPath := filepath.Join(pluginPath, "skills", name)

		if err := os.MkdirAll(skillsPath, 0755); err != nil {
			output.Error("Error creando directorio de skills: %v", err)
			os.Exit(1)
		}
		manifestPath := filepath.Join(pluginPath, ".juarvis-plugin")
		if err := os.MkdirAll(manifestPath, 0755); err != nil {
			output.Error("Error creando directorio de manifiesto: %v", err)
			os.Exit(1)
		}

		// Escribir manifest
		manifest := fmt.Sprintf(`{
  "name": "juarvis-%s",
  "version": "1.0.0",
  "description": "Custom skill creada localmente",
  "category": "custom"
}`, name)
		if err := os.WriteFile(filepath.Join(manifestPath, "plugin.json"), []byte(manifest), 0644); err != nil {
			output.Error("Error escribiendo plugin.json: %v", err)
			os.Exit(1)
		}

		// Escribir SKILL.md
		skillMD := fmt.Sprintf("---\nname: %s\ndescription: Custom skill template\n---\n\n## Instrucciones\n\nEscribe aquí los comandos...", name)
		if err := os.WriteFile(filepath.Join(skillsPath, "SKILL.md"), []byte(skillMD), 0644); err != nil {
			output.Error("Error escribiendo SKILL.md: %v", err)
			os.Exit(1)
		}

		output.Success("Estructura base creada. Indexando...")
		if err := loader.RunLoader(); err != nil {
			output.Warning("Advertencia en indexación: %v", err)
		} else {
			output.Success("Skill integrada exitosamente!")
		}
	},
}

func init() {
	rootCmd.AddCommand(loaderCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(skillCreateCmd)
}
