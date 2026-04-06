package watcher

import (
	"fmt"
	"juarvis/pkg/hookify"
	"juarvis/pkg/output"
	"os"
	"path/filepath"
)

func EvaluateFileChanges(changes map[string]string) {
	rules := hookify.LoadRules("file")
	rules = append(rules, hookify.LoadRules("all")...)

	if len(rules) == 0 {
		return
	}

	for path, eventType := range changes {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		inputData := map[string]any{
			"hook_event_name": "FileChange",
			"tool_name":       "FileChange",
			"tool_input": map[string]any{
				"file_path":  path,
				"event_type": eventType,
				"content":    string(content),
			},
			"file_path":  path,
			"event_type": eventType,
			"reason":     fmt.Sprintf("File %s: %s", eventType, filepath.Base(path)),
		}

		result := hookify.EvaluateRules(rules, inputData)
		if result.SystemMessage != "" {
			output.Warning("[Hookify] %s: %s", filepath.Base(path), result.SystemMessage)
		}
		if result.Decision == "block" {
			output.Error("[Hookify] Acción bloqueada en %s", filepath.Base(path))
		}
	}
}
