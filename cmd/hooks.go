package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"juarvis/pkg/hookify"
	"juarvis/pkg/output"
)

var (
	createName     string
	createEvent    string
	createPattern  string
	createAction   string
	createDisabled bool
)

var hooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Gestión de hooks de comportamiento",
	Long: `Comandos para gestionar hooks:
  hooks list       - Lista todas las reglas
  hooks create     - Crea una nueva regla
  hooks enable     - Habilita una regla
  hooks disable   - Deshabilita una regla`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
		}
	},
}

var hooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista todas las reglas hooks",
	Run: func(cmd *cobra.Command, args []string) {
		rules := hookify.LoadRules("")
		rules = append(rules, hookify.LoadRules("all")...)

		if len(rules) == 0 {
			output.Info("No hay reglas hooks cargadas")
			return
		}

		output.Success("%d reglas:", len(rules))
		rows := [][]string{}
		for _, r := range rules {
			enabled := "✅"
			if !r.Enabled {
				enabled = "❌"
			}
			rows = append(rows, []string{r.Name, r.Event, r.Action, enabled})
		}
		output.PrintTable([]string{"NOMBRE", "EVENTO", "ACCIÓN", "ESTADO"}, rows)
	},
}

var hooksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Crea una nueva regla hook",
	Run: func(cmd *cobra.Command, args []string) {
		name := createName
		if name == "" {
			output.Fatal(output.ExitGeneric, "Especifica --name", "")
		}
		pattern := createPattern
		if pattern == "" {
			output.Fatal(output.ExitGeneric, "Especifica --pattern", "")
		}
		event := createEvent
		if event == "" {
			event = "all"
		}
		action := createAction
		if action == "" {
			action = "warn"
		}

		enabled := !createDisabled

		content := `---
name: ` + name + `
enabled: ` + fmt.Sprintf("%t", enabled) + `
event: ` + event + `
pattern: ` + pattern + `
action: ` + action + `

# Rule: ` + name + `
Pattern: ` + pattern + `
Action: ` + action
		// Create file
		ruleFile := ".juar/hookify." + name + ".md"
		os.WriteFile(ruleFile, []byte(content), 0644)

		output.Success("Regla creada: %s", name)
	},
}

var hooksEnableCmd = &cobra.Command{
	Use:   "enable [nombre]",
	Short: "Habilita una regla",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setHookEnabled(args[0], true)
	},
}

var hooksDisableCmd = &cobra.Command{
	Use:   "disable [nombre]",
	Short: "Deshabilita una regla",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setHookEnabled(args[0], false)
	},
}

func setHookEnabled(name string, enabled bool) {
	matches, _ := filepath.Glob(".juar/hookify." + name + "*.md")
	if len(matches) == 0 {
		output.Fatal(output.ExitGeneric, "Regla no encontrada", "name: %v", name)
	}

	content, _ := os.ReadFile(matches[0])
	enabledStr := "enabled: true"
	if !enabled {
		enabledStr = "enabled: false"
	}
	newContent := strings.Replace(string(content), enabledStr, "enabled: "+fmt.Sprintf("%t", enabled), 1)
	os.WriteFile(matches[0], []byte(newContent), 0644)

	state := "habilitada"
	if !enabled {
		state = "deshabilitada"
	}
	output.Success("Regla %s %s", name, state)
}

func init() {
	hooksCmd.AddCommand(hooksListCmd)
	hooksCmd.AddCommand(hooksCreateCmd)
	hooksCmd.AddCommand(hooksEnableCmd)
	hooksCmd.AddCommand(hooksDisableCmd)

	hooksCreateCmd.Flags().StringVar(&createName, "name", "", "Nombre de la regla")
	hooksCreateCmd.Flags().StringVar(&createEvent, "event", "all", "Evento (bash, file, stop, all)")
	hooksCreateCmd.Flags().StringVar(&createPattern, "pattern", "", "Patrón regex")
	hooksCreateCmd.Flags().StringVar(&createAction, "action", "warn", "Acción (warn, block)")
	hooksCreateCmd.Flags().BoolVar(&createDisabled, "disabled", false, "Crear deshabilitada")

	rootCmd.AddCommand(hooksCmd)
}
