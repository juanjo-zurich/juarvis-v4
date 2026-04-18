package hookify

import (
	"fmt"
	"juarvis/pkg/config"
	"juarvis/pkg/utils"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

type Condition struct {
	Field    string
	Operator string
	Pattern  string
}

type Rule struct {
	Name        string
	Enabled     bool
	Event       string
	Pattern     string
	Conditions  []Condition
	Action      string
	ToolMatcher string
	Message     string
}

// extractFrontmatter extrae el frontmatter YAML de un archivo markdown.
// El frontmatter debe estar entre --- y --- al inicio del archivo.
func extractFrontmatter(content string) (map[string]interface{}, string, error) {
	fmStr, body, found := utils.ExtractFrontmatterBlock(content)
	if !found {
		return nil, content, nil
	}

	var fm map[string]interface{}
	if err := yaml.Unmarshal([]byte(fmStr), &fm); err != nil {
		return nil, content, fmt.Errorf("error parseando frontmatter: %w", err)
	}

	return fm, body, nil
}

func ruleFromDict(fm map[string]interface{}, message string) Rule {
	var conditions []Condition

	if condList, ok := fm["conditions"].([]interface{}); ok {
		for _, c := range condList {
			if cm, ok := c.(map[string]interface{}); ok {
				conditions = append(conditions, Condition{
					Field:    fmt.Sprint(cm["field"]),
					Operator: fmt.Sprint(cm["operator"]),
					Pattern:  fmt.Sprint(cm["pattern"]),
				})
			}
		}
	}

	simplePattern, _ := fm["pattern"].(string)
	if simplePattern != "" && len(conditions) == 0 {
		event, _ := fm["event"].(string)
		field := "content"
		switch event {
		case "bash":
			field = "command"
		case "file":
			field = "new_text"
		}
		conditions = append(conditions, Condition{
			Field:    field,
			Operator: "regex_match",
			Pattern:  simplePattern,
		})
	}

	name, _ := fm["name"].(string)
	enabled := true
	if v, ok := fm["enabled"].(bool); ok {
		enabled = v
	}
	event, _ := fm["event"].(string)
	action, _ := fm["action"].(string)
	if action == "" {
		action = "warn"
	}
	toolMatcher, _ := fm["tool_matcher"].(string)

	return Rule{
		Name:        name,
		Enabled:     enabled,
		Event:       event,
		Pattern:     simplePattern,
		Conditions:  conditions,
		Action:      action,
		ToolMatcher: toolMatcher,
		Message:     message,
	}
}

func loadRuleFile(filePath string) (Rule, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Rule{}, fmt.Errorf("cannot read %s: %w", filePath, err)
	}

	content := string(data)
	fm, message, err := extractFrontmatter(content)
	if err != nil {
		return Rule{}, fmt.Errorf("%s: %w", filePath, err)
	}
	if fm == nil {
		return Rule{}, fmt.Errorf("%s missing YAML frontmatter", filePath)
	}

	return ruleFromDict(fm, message), nil
}

func LoadRules(eventFilter string) []Rule {
	var rules []Rule

	pattern := filepath.Join(config.JuarvisDir, config.HookifyPattern)
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil
	}

	for _, f := range files {
		rule, err := loadRuleFile(f)
		if err != nil {
			continue
		}
		if !rule.Enabled {
			continue
		}
		if eventFilter != "" && rule.Event != "all" && rule.Event != eventFilter {
			continue
		}
		rules = append(rules, rule)
	}

	return rules
}

const maxPatternLength = 1000

var regexCache sync.Map

func compileRegex(pattern string) (*regexp.Regexp, error) {
	if val, ok := regexCache.Load(pattern); ok {
		return val.(*regexp.Regexp), nil
	}
	if len(pattern) > maxPatternLength {
		return nil, fmt.Errorf("patrón de regex demasiado largo (%d caracteres, máximo %d)", len(pattern), maxPatternLength)
	}
	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return nil, fmt.Errorf("patrón de regex inválido %q: %w", pattern, err)
	}
	regexCache.Store(pattern, re)
	return re, nil
}

func matchesTool(matcher, toolName string) bool {
	if matcher == "*" {
		return true
	}
	for _, p := range strings.Split(matcher, "|") {
		if strings.TrimSpace(p) == toolName {
			return true
		}
	}
	return false
}

func extractField(field, toolName string, toolInput map[string]any, inputData map[string]any) string {
	if v, ok := toolInput[field]; ok {
		switch val := v.(type) {
		case string:
			return val
		default:
			return fmt.Sprint(val)
		}
	}

	switch field {
	case "reason":
		if v, ok := inputData["reason"].(string); ok {
			return v
		}
	case "transcript":
		if path, ok := inputData["transcript_path"].(string); ok && path != "" {
			data, err := os.ReadFile(path)
			if err == nil {
				return string(data)
			}
		}
		return ""
	case "user_prompt":
		if v, ok := inputData["user_prompt"].(string); ok {
			return v
		}
	}

	switch toolName {
	case "Bash":
		if field == "command" {
			if v, ok := toolInput["command"].(string); ok {
				return v
			}
		}
	case "Write", "Edit":
		switch field {
		case "content", "new_text", "new_string":
			for _, k := range []string{"content", "new_string"} {
				if v, ok := toolInput[k].(string); ok && v != "" {
					return v
				}
			}
		case "old_text", "old_string":
			if v, ok := toolInput["old_string"].(string); ok {
				return v
			}
		case "file_path":
			if v, ok := toolInput["file_path"].(string); ok {
				return v
			}
		}
	case "MultiEdit":
		if field == "file_path" {
			if v, ok := toolInput["file_path"].(string); ok {
				return v
			}
		}
		if field == "new_text" || field == "content" {
			if edits, ok := toolInput["edits"].([]any); ok {
				var parts []string
				for _, e := range edits {
					if em, ok := e.(map[string]any); ok {
						if s, ok := em["new_string"].(string); ok {
							parts = append(parts, s)
						}
					}
				}
				return strings.Join(parts, " ")
			}
		}
	}

	return ""
}

func checkCondition(cond Condition, toolName string, toolInput map[string]any, inputData map[string]any) bool {
	fieldValue := extractField(cond.Field, toolName, toolInput, inputData)

	switch cond.Operator {
	case "script":
		// Ejecutar script externo
		// El script recibe el valor del campo por Stdin y debe devolver exit code 0 para PASS, != 0 para FAIL
		cmdParts := strings.Fields(cond.Pattern)
		if len(cmdParts) == 0 {
			return false
		}
		cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
		cmd.Stdin = strings.NewReader(fieldValue)
		err := cmd.Run()
		return err == nil

	case "regex_match":
		if fieldValue == "" {
			return false
		}
		re, err := compileRegex(cond.Pattern)
		if err != nil {
			return false
		}
		return re.MatchString(fieldValue)
	case "contains":
		return strings.Contains(fieldValue, cond.Pattern)
	case "equals":
		return fieldValue == cond.Pattern
	case "not_contains":
		return !strings.Contains(fieldValue, cond.Pattern)
	case "starts_with":
		return strings.HasPrefix(fieldValue, cond.Pattern)
	case "ends_with":
		return strings.HasSuffix(fieldValue, cond.Pattern)
	default:
		return false
	}
}

func ruleMatches(rule Rule, inputData map[string]any) bool {
	toolName, _ := inputData["tool_name"].(string)
	toolInput, _ := inputData["tool_input"].(map[string]any)

	if rule.ToolMatcher != "" && !matchesTool(rule.ToolMatcher, toolName) {
		return false
	}

	if len(rule.Conditions) == 0 {
		return false
	}

	for _, cond := range rule.Conditions {
		if !checkCondition(cond, toolName, toolInput, inputData) {
			return false
		}
	}

	return true
}

type HookResult struct {
	SystemMessage      string
	HookSpecificOutput map[string]any
	Decision           string
}

func EvaluateRules(rules []Rule, inputData map[string]any) HookResult {
	hookEvent, _ := inputData["hook_event_name"].(string)
	var blocking, warning []Rule

	for _, rule := range rules {
		if ruleMatches(rule, inputData) {
			// Auto-Fixer: si la acción es 'fix', intentamos reparar
			if strings.HasPrefix(rule.Action, "fix:") {
				fixCmd := strings.TrimPrefix(rule.Action, "fix:")
				cmdParts := strings.Fields(fixCmd)
				if len(cmdParts) > 0 {
					// El fix recibe los datos del campo (content por defecto) y puede modificar el entorno
					cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
					_ = cmd.Run()
					// Tras el fix, el check vuelve a evaluar si sigue rompiendo
					if !ruleMatches(rule, inputData) {
						continue // Reparado con éxito
					}
				}
			}

			if rule.Action == "block" {
				blocking = append(blocking, rule)
			} else {
				warning = append(warning, rule)
			}
		}
	}

	if len(blocking) > 0 {
		var msgs []string
		for _, r := range blocking {
			msgs = append(msgs, fmt.Sprintf("**[%s]**\n%s", r.Name, r.Message))
		}
		combined := strings.Join(msgs, "\n\n")

		result := HookResult{SystemMessage: combined}

		switch hookEvent {
		case "Stop":
			result.Decision = "block"
		case "PreToolUse", "PostToolUse":
			result.HookSpecificOutput = map[string]any{
				"hookEventName":      hookEvent,
				"permissionDecision": "deny",
			}
		}
		return result
	}

	if len(warning) > 0 {
		var msgs []string
		for _, r := range warning {
			msgs = append(msgs, fmt.Sprintf("**[%s]**\n%s", r.Name, r.Message))
		}
		return HookResult{SystemMessage: strings.Join(msgs, "\n\n")}
	}

	return HookResult{}
}
