package watcher

import (
	"fmt"
	"juarvis/pkg/hookify"
	"juarvis/pkg/output"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	rulesCache     []hookify.Rule
	rulesCacheTime time.Time
	rulesCacheMu   sync.Mutex
	rulesTTL       = 30 * time.Second // Cache rules for 30 seconds
)

// reloadRulesIfNeeded recarg reglas si han pasado más de 30 segundos
func reloadRulesIfNeeded() {
	rulesCacheMu.Lock()
	defer rulesCacheMu.Unlock()

	if time.Since(rulesCacheTime) > rulesTTL {
		rules := hookify.LoadRules("file")
		rules = append(rules, hookify.LoadRules("all")...)
		rulesCache = rules
		rulesCacheTime = time.Now()
		if len(rules) > 0 {
			output.Info("Reglas de security recargadas")
		}
	}
}

func getCachedRules() []hookify.Rule {
	rulesCacheMu.Lock()
	defer rulesCacheMu.Unlock()

	if time.Since(rulesCacheTime) > rulesTTL {
		// Unlock briefly to reload
		rulesCacheMu.Unlock()
		rules := hookify.LoadRules("file")
		rules = append(rules, hookify.LoadRules("all")...)
		rulesCacheMu.Lock()
		rulesCache = rules
		rulesCacheTime = time.Now()
	}

	if rulesCache == nil {
		rules := hookify.LoadRules("file")
		rules = append(rules, hookify.LoadRules("all")...)
		rulesCache = rules
		rulesCacheTime = time.Now()
	}

	return rulesCache
}

func EvaluateFileChanges(changes map[string]string) {
	// Recargar reglas si es necesario
	reloadRulesIfNeeded()

	rules := getCachedRules()

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
