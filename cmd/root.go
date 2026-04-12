package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"juarvis/pkg/config"
	"juarvis/pkg/output"
	"juarvis/pkg/root"
)

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

var GlobalRoot string
var GlobalJSON bool

var rootCmd = &cobra.Command{
	Use:   "juarvis",
	Short: "Juarvis V4 CLI - El motor de tu ecosistema de agentes IA",
	Long:  `Juarvis V4 gestiona Spec-Driven Development, integradores MCP y control de paquetes de skills (Marketplace) directamente compilado en Go.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if GlobalRoot != "" {
			_ = os.Setenv("JUARVIS_ROOT", GlobalRoot)
		}
		output.SetJSONMode(GlobalJSON)
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Detectar si estamos dentro de un ecosistema
		rootPath, err := root.GetRoot()
		if err != nil {
			// No estamos en un ecosistema
			output.Info("No se detecto un ecosistema Juarvis en este directorio")
			fmt.Println()
			fmt.Println("Comandos disponibles:")
			fmt.Printf("  %-20s %s\n", "juarvis init", "Inicializa un nuevo ecosistema en este directorio")
			fmt.Printf("  %-20s %s\n", "juarvis init [path]", "Inicializa un ecosistema en el path especificado")
			fmt.Printf("  %-20s %s\n", "juarvis --help", "Ver todos los comandos disponibles")
			fmt.Println()
			output.Info("Ejecuta 'juarvis init' para empezar")
			return
		}

		// Estamos en un ecosistema — mostrar estado
		output.Info("Ecosistema Juarvis detectado en: %s", rootPath)
		fmt.Println()

		// Verificar componentes
		checks := []struct {
			name  string
			check func(string) bool
		}{
			{"marketplace.json", func(r string) bool {
				_, err := os.Stat(filepath.Join(r, "marketplace.json"))
				return err == nil
			}},
			{"AGENTS.md", func(r string) bool {
				_, err := os.Stat(filepath.Join(r, "AGENTS.md"))
				return err == nil
			}},
			{"permissions.yaml", func(r string) bool {
				_, err := os.Stat(filepath.Join(r, "permissions.yaml"))
				return err == nil
			}},
			{"opencode.json", func(r string) bool {
				_, err := os.Stat(filepath.Join(r, "opencode.json"))
				return err == nil
			}},
			{"plugins/", func(r string) bool {
				entries, err := os.ReadDir(filepath.Join(r, "plugins"))
				return err == nil && len(entries) > 0
			}},
			{"skills/", func(r string) bool {
				_, err := os.Stat(filepath.Join(r, "skills"))
				return err == nil
			}},
			{config.JuarDir + "/" + config.SkillRegistryFile, func(r string) bool {
				_, err := os.Stat(filepath.Join(r, config.JuarDir, config.SkillRegistryFile))
				return err == nil
			}},
		}

		for _, c := range checks {
			if c.check(rootPath) {
				output.Success("%s", c.name)
			} else {
				output.Warning("%s no encontrado", c.name)
			}
		}

		fmt.Println()
		fmt.Println("Comandos rapidos:")
		fmt.Printf("  %-25s %s\n", "juarvis check", "Health check completo")
		fmt.Printf("  %-25s %s\n", "juarvis pm list", "Listar plugins del marketplace")
		fmt.Printf("  %-25s %s\n", "juarvis setup --ide <ide>", "Distribuir reglas al IDE")
		fmt.Printf("  %-25s %s\n", "juarvis --help", "Ver todos los comandos")
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&GlobalRoot, "root", "", "Directorio raíz del ecosistema Juarvis (opcional)")
	rootCmd.PersistentFlags().BoolVar(&GlobalJSON, "json", false, "Salida en formato JSON para consumo programatico")
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, BuildDate)
	rootCmd.SetVersionTemplate("juarvis version {{.Version}}\n")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
