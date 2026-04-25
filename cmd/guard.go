package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"juarvis/pkg/output"
	"juarvis/pkg/root"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type PermissionRule struct {
	Pattern string `yaml:"pattern"`
	Action  string `yaml:"action"`
	Reason  string `yaml:"reason"`
}

type PermissionConfig struct {
	Version string                      `yaml:"version"`
	Rules   map[string][]PermissionRule `yaml:"rules"`
	Limits  map[string]int              `yaml:"limits"`
	Audit   struct {
		Enabled       bool   `yaml:"enabled"`
		LogFile       string `yaml:"log_file"`
		LogDecisions  bool   `yaml:"log_decisions"`
		LogExecutions bool   `yaml:"log_executions"`
	} `yaml:"audit"`
}

var guardCmd = &cobra.Command{
	Use:   "guard",
	Short: "Permission guard - evalúa si un comando está permitido",
	Long:  `Servidor STDIN que evalúa permisos de comandos. Uso: echo "git push" | juarvis guard`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runGuard(); err != nil {
			output.Error("Error en guard: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(guardCmd)
}

func loadPermissions() (*PermissionConfig, error) {
	rootPath, err := root.GetRoot()
	if err != nil {
		return nil, err
	}

	permFile := filepath.Join(rootPath, "permissions.yaml")
	data, err := os.ReadFile(permFile)
	if err != nil {
		return nil, err
	}

	var config PermissionConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func evaluateCommand(cmd string, config *PermissionConfig) (string, string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return "allow", "comando vacío"
	}

	toolName := extractToolName(cmd)
	rules, ok := config.Rules[toolName]
	if !ok {
		toolName = "bash"
		rules, ok = config.Rules[toolName]
		if !ok {
			return "allow", "herramienta no encontrada en permisos, permitiendo por defecto"
		}
	}

	for _, rule := range rules {
		matched, err := filepath.Match(rule.Pattern, cmd)
		if err == nil && matched {
			return rule.Action, rule.Reason
		}

		if strings.Contains(cmd, strings.Trim(rule.Pattern, "*")) {
			return rule.Action, rule.Reason
		}
	}

	return "allow", "no hay regla específica, permitiendo por defecto"
}

func extractToolName(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "bash"
	}
	return parts[0]
}

func runGuard() error {
	config, err := loadPermissions()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		cmd := scanner.Text()
		action, reason := evaluateCommand(cmd, config)

		result := map[string]string{
			"command": cmd,
			"action":  action,
			"reason":  reason,
		}

		data, _ := json.Marshal(result)
		fmt.Println(string(data))

		if action == "deny" {
			os.Exit(1)
		}
	}

	return scanner.Err()
}

func CheckPermission(cmd string) (bool, string) {
	config, err := loadPermissions()
	if err != nil {
		return true, "no se pudieron cargar permisos"
	}

	action, reason := evaluateCommand(cmd, config)
	return action == "allow", reason
}

func CheckPermissionWithAsk(cmd string) (bool, string, bool) {
	config, err := loadPermissions()
	if err != nil {
		return true, "no se pudieron cargar permisos", false
	}

	action, reason := evaluateCommand(cmd, config)
	return action == "allow", reason, action == "ask"
}
