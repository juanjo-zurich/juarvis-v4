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
	Use:    "hookify",
	Short:  "Hookify engine - Evaluates user-defined rules against hook events",
	Hidden: true,
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
	Short: "List loaded hookify rules",
	Run: func(cmd *cobra.Command, args []string) {
		rules := hookify.LoadRules("")
		rules = append(rules, hookify.LoadRules("all")...)

		if len(rules) == 0 {
			output.Info("No hookify rules loaded")
			return
		}

		output.Info("%d hookify rules loaded", len(rules))
		headers := []string{"NAME", "EVENT", "ACTION", "ENABLED"}
		rows := [][]string{}
		for _, r := range rules {
			rows = append(rows, []string{r.Name, r.Event, r.Action, fmt.Sprint(r.Enabled)})
		}
		output.PrintTable(headers, rows)
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
