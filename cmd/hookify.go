package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"juarvis/pkg/hookify"
	"juarvis/pkg/output"

	"github.com/spf13/cobra"
)

var hookifyCmd = &cobra.Command{
	Use:   "hookify",
	Short: "Hookify engine - Gestiona reglas de comportamiento",
	Long: `Sistema de hooks para prevenir comportamientos no deseados.

Comandos:
  juarvis hookify list             - Lista reglas
  juarvis hookify create         - Crea regla
  juarvis hookify enable [nombre]  - Habilita regla
  juarvis hookify disable [nombre] - Deshabilita regla`,
}

var hookifyPreToolUseCmd = &cobra.Command{
	Use:    "pretooluse",
	Short:  "Evaluate PreToolUse hooks",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHookifyHook("PreToolUse", "bash")
	},
}

var hookifyPostToolUseCmd = &cobra.Command{
	Use:    "posttooluse",
	Short:  "Evaluate PostToolUse hooks",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHookifyHook("PostToolUse", "bash")
	},
}

var hookifyStopCmd = &cobra.Command{
	Use:    "stop",
	Short:  "Evaluate Stop hooks",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHookifyHook("Stop", "stop")
	},
}

var hookifyUserPromptSubmitCmd = &cobra.Command{
	Use:    "userpromptsubmit",
	Short:  "Evaluate UserPromptSubmit hooks",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHookifyHook("UserPromptSubmit", "prompt")
	},
}

func runHookifyHook(eventName, eventFilter string) error {
	var inputData map[string]any
	if err := json.NewDecoder(os.Stdin).Decode(&inputData); err != nil {
		return fmt.Errorf("invalid JSON from stdin: %w", err)
	}

	inputData["hook_event_name"] = eventName

	rules := hookify.LoadRules(eventFilter)
	rules = append(rules, hookify.LoadRules("all")...)

	if len(rules) == 0 {
		os.Exit(0)
	}

	result := hookify.EvaluateRules(rules, inputData)

	if result.SystemMessage != "" {
		fmt.Fprintln(os.Stderr, result.SystemMessage)
	}

	if result.HookSpecificOutput != nil {
		if decision, ok := result.HookSpecificOutput["permissionDecision"].(string); ok && decision == "deny" {
			os.Exit(2)
		}
	}

	if result.Decision == "block" {
		os.Exit(2)
	}

	os.Exit(0)
	return nil
}

var hookifyListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista todas las reglas hookify",
	Long: `Lista todas las reglas de hookify cargadas.

Muestra: nombre, evento, acción, habilitada`,
	Run: func(cmd *cobra.Command, args []string) {
		rules := hookify.LoadRules("")
		rules = append(rules, hookify.LoadRules("all")...)

		if len(rules) == 0 {
			output.Info("No hay reglas hookify cargadas")
			return
		}

		output.Success("%d reglas cargadas:", len(rules))
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

func init() {
	hookifyCmd.AddCommand(hookifyPreToolUseCmd)
	hookifyCmd.AddCommand(hookifyPostToolUseCmd)
	hookifyCmd.AddCommand(hookifyStopCmd)
	hookifyCmd.AddCommand(hookifyUserPromptSubmitCmd)
	hookifyCmd.AddCommand(hookifyListCmd)
	rootCmd.AddCommand(hookifyCmd)
}
