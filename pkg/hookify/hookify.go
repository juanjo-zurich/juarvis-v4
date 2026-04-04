package hookify

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

func extractFrontmatter(content string) (map[string]any, string) {
	if !strings.HasPrefix(content, "---") {
		return nil, content
	}

	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, content
	}

	fmText := strings.TrimSpace(parts[1])
	message := strings.TrimSpace(parts[2])

	fm := make(map[string]any)
	lines := strings.Split(fmText, "\n")

	var currentKey string
	var currentList []any
	var currentDict map[string]string
	inList := false
	inDictItem := false

	for _, line := range lines {
		stripped := strings.TrimSpace(line)
		if stripped == "" || strings.HasPrefix(stripped, "#") {
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " \t"))

		if indent == 0 && strings.Contains(line, ":") && !strings.HasPrefix(stripped, "-") {
			if inList && currentKey != "" {
				if inDictItem && currentDict != nil {
					currentList = append(currentList, currentDict)
					currentDict = nil
				}
				fm[currentKey] = currentList
				inList = false
				inDictItem = false
				currentList = nil
			}

			idx := strings.Index(line, ":")
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])

			if value == "" {
				currentKey = key
				inList = true
				currentList = []any{}
			} else {
				value = strings.Trim(value, "\"'")
				switch strings.ToLower(value) {
				case "true":
					fm[key] = true
				case "false":
					fm[key] = false
				default:
					fm[key] = value
				}
			}
		} else if strings.HasPrefix(stripped, "-") && inList {
			if inDictItem && currentDict != nil {
				currentList = append(currentList, currentDict)
				currentDict = nil
			}

			itemText := strings.TrimSpace(stripped[1:])
			if strings.Contains(itemText, ":") && strings.Contains(itemText, ",") {
				itemDict := make(map[string]string)
				for _, part := range strings.Split(itemText, ",") {
					if idx := strings.Index(part, ":"); idx >= 0 {
						k := strings.TrimSpace(part[:idx])
						v := strings.TrimSpace(strings.Trim(part[idx+1:], "\"'"))
						itemDict[k] = v
					}
				}
				currentList = append(currentList, itemDict)
				inDictItem = false
			} else if strings.Contains(itemText, ":") {
				inDictItem = true
				idx := strings.Index(itemText, ":")
				currentDict = map[string]string{
					strings.TrimSpace(itemText[:idx]): strings.TrimSpace(strings.Trim(itemText[idx+1:], "\"'")),
				}
			} else {
				currentList = append(currentList, strings.Trim(itemText, "\"'"))
				inDictItem = false
			}
		} else if indent > 2 && inDictItem && strings.Contains(line, ":") {
			idx := strings.Index(stripped, ":")
			if idx >= 0 {
				currentDict[strings.TrimSpace(stripped[:idx])] = strings.TrimSpace(strings.Trim(stripped[idx+1:], "\"'"))
			}
		}
	}

	if inList && currentKey != "" {
		if inDictItem && currentDict != nil {
			currentList = append(currentList, currentDict)
		}
		fm[currentKey] = currentList
	}

	return fm, message
}

func ruleFromDict(fm map[string]any, message string) Rule {
	var conditions []Condition

	if condList, ok := fm["conditions"].([]any); ok {
		for _, c := range condList {
			if cm, ok := c.(map[string]string); ok {
				conditions = append(conditions, Condition{
					Field:    cm["field"],
					Operator: cm["operator"],
					Pattern:  cm["pattern"],
				})
			} else if cm, ok := c.(map[string]any); ok {
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
	fm, message := extractFrontmatter(content)
	if fm == nil {
		return Rule{}, fmt.Errorf("%s missing YAML frontmatter", filePath)
	}

	return ruleFromDict(fm, message), nil
}

func LoadRules(eventFilter string) []Rule {
	var rules []Rule

	pattern := filepath.Join(".opencode", "hookify.*.local.md")
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

var regexCache = make(map[string]*regexp.Regexp)

func compileRegex(pattern string) (*regexp.Regexp, error) {
	if re, ok := regexCache[pattern]; ok {
		return re, nil
	}
	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return nil, err
	}
	regexCache[pattern] = re
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
	if fieldValue == "" {
		return false
	}

	switch cond.Operator {
	case "regex_match":
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
