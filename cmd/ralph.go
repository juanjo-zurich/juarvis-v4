package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"juarvis/pkg/output"
	"juarvis/pkg/ralph"
	"juarvis/pkg/root"

	"github.com/spf13/cobra"
)

var ralphCmd = &cobra.Command{
	Use:   "ralph",
	Short: "Ralph Wiggum loop engine - Self-referential development loops",
}

var ralphLoopCmd = &cobra.Command{
	Use:   "loop [prompt...]",
	Short: "Start a Ralph self-referential loop in the current session",
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt := strings.Join(args, " ")
		if prompt == "" {
			output.Error("No prompt provided. Ralph needs a task description.")
			output.Info("Example: juarvis ralph loop Build a REST API --max-iterations 20")
			return fmt.Errorf("empty prompt")
		}

		maxIter, _ := cmd.Flags().GetInt("max-iterations")
		completionPromise, _ := cmd.Flags().GetString("completion-promise")

		rootPath, _ := root.GetRoot()
		if err := ralph.CreateLoopState(rootPath, prompt, maxIter, completionPromise); err != nil {
			output.Error("Error en el bucle de Ralph: %v", err)
			os.Exit(1)
		}

		iterLabel := "unlimited"
		if maxIter > 0 {
			iterLabel = strconv.Itoa(maxIter)
		}

		promiseLabel := "none (runs forever)"
		if completionPromise != "" {
			promiseLabel = completionPromise
		}

		output.Success("Ralph loop activated in this session!")
		output.Info("Iteration: 1")
		output.Info("Max iterations: %s", iterLabel)
		output.Info("Completion promise: %s", promiseLabel)
		output.Warning("This loop cannot be stopped manually! It will run infinitely unless you set --max-iterations or --completion-promise.")

		return nil
	},
}

var ralphStopCmd = &cobra.Command{
	Use:    "stop",
	Short:  "Evaluate Ralph stop hook (called by the Stop event)",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var inputData map[string]any
		if err := json.NewDecoder(os.Stdin).Decode(&inputData); err != nil {
			return fmt.Errorf("invalid JSON from stdin: %w", err)
		}

		rootPath, _ := root.GetRoot()
		state, err := ralph.LoadState(rootPath)
		if err != nil {
			os.Exit(0)
		}

		if !state.IsActive() {
			os.Exit(0)
		}

		if state.IsComplete() {
			output.Error("Ralph loop: Max iterations (%d) reached.", state.MaxIterations)
			state.Delete()
			os.Exit(0)
		}

		transcriptPath, _ := inputData["transcript_path"].(string)
		if transcriptPath == "" {
			output.Warning("Ralph loop: No transcript_path in hook input. Stopping.")
			state.Delete()
			os.Exit(0)
		}

		result, err := ralph.BuildStopResponse(state, transcriptPath)
		if err != nil {
			output.Warning("Ralph loop error: %v", err)
			state.Delete()
			os.Exit(0)
		}

		if result["decision"] == "allow" {
			fmt.Println(result["systemMessage"])
			os.Exit(0)
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(result)

		os.Exit(0)
		return nil
	},
}

var ralphStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current Ralph loop status",
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, _ := root.GetRoot()
		state, err := ralph.LoadState(rootPath)
		if err != nil {
			output.Info("No active Ralph loop")
			return
		}

		output.Info("Ralph loop status:")
		output.Info("  Active: %v", state.Active)
		output.Info("  Iteration: %d", state.Iteration)
		output.Info("  Max iterations: %d", state.MaxIterations)
		output.Info("  Completion promise: %s", state.CompletionPromise)
		output.Info("  Started at: %s", state.StartedAt)
		output.Info("  Prompt: %s", state.Prompt)
	},
}

var ralphResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset/cancel the current Ralph loop",
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, _ := root.GetRoot()
		state, err := ralph.LoadState(rootPath)
		if err != nil {
			output.Info("No active Ralph loop to reset")
			return
		}
		state.Delete()
		output.Success("Ralph loop cancelled")
	},
}

func init() {
	ralphLoopCmd.Flags().Int("max-iterations", 0, "Maximum iterations before auto-stop (0 = unlimited)")
	ralphLoopCmd.Flags().String("completion-promise", "", "Promise phrase to detect completion")
	ralphCmd.AddCommand(ralphLoopCmd)
	ralphCmd.AddCommand(ralphStopCmd)
	ralphCmd.AddCommand(ralphStatusCmd)
	ralphCmd.AddCommand(ralphResetCmd)
	rootCmd.AddCommand(ralphCmd)
}
