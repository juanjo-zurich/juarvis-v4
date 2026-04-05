package hookify

import (
	"strings"
	"testing"
)

func TestExtractFrontmatter_Valid(t *testing.T) {
	content := `---
name: test-rule
enabled: true
event: bash
pattern: rm\s+-rf
action: block
---

Mensaje de advertencia
`
	fm, body, err := extractFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm["name"] != "test-rule" {
		t.Errorf("expected name 'test-rule', got %v", fm["name"])
	}
	if !strings.Contains(body, "Mensaje de advertencia") {
		t.Errorf("expected body to contain message, got: %s", body)
	}
}

func TestExtractFrontmatter_NoFrontmatter(t *testing.T) {
	content := "Just plain markdown"
	fm, body, err := extractFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm != nil {
		t.Error("expected nil frontmatter")
	}
	if body != content {
		t.Errorf("expected body to equal content")
	}
}

func TestExtractFrontmatter_Unclosed(t *testing.T) {
	content := "---\nname: test\n"
	_, _, err := extractFrontmatter(content)
	if err == nil {
		t.Fatal("expected error for unclosed frontmatter")
	}
}

func TestCheckCondition_RegexMatch(t *testing.T) {
	cond := Condition{Field: "command", Operator: "regex_match", Pattern: `rm\s+-rf`}
	toolInput := map[string]any{"command": "rm -rf /tmp"}
	inputData := map[string]any{}
	result := checkCondition(cond, "Bash", toolInput, inputData)
	if !result {
		t.Error("expected regex match to succeed")
	}
}

func TestCheckCondition_Contains(t *testing.T) {
	cond := Condition{Field: "command", Operator: "contains", Pattern: "dangerous"}
	toolInput := map[string]any{"command": "this is a dangerous command"}
	inputData := map[string]any{}
	result := checkCondition(cond, "Bash", toolInput, inputData)
	if !result {
		t.Error("expected contains to succeed")
	}
}

func TestCheckCondition_Equals(t *testing.T) {
	cond := Condition{Field: "command", Operator: "equals", Pattern: "exact"}
	toolInput := map[string]any{"command": "exact"}
	inputData := map[string]any{}
	result := checkCondition(cond, "Bash", toolInput, inputData)
	if !result {
		t.Error("expected equals to succeed")
	}
}

func TestCheckCondition_NotContains(t *testing.T) {
	cond := Condition{Field: "command", Operator: "not_contains", Pattern: "safe"}
	toolInput := map[string]any{"command": "this is dangerous"}
	inputData := map[string]any{}
	result := checkCondition(cond, "Bash", toolInput, inputData)
	if !result {
		t.Error("expected not_contains to succeed")
	}
}

func TestEvaluateRules_Block(t *testing.T) {
	rules := []Rule{
		{
			Name:    "block-test",
			Enabled: true,
			Event:   "bash",
			Conditions: []Condition{
				{Field: "command", Operator: "regex_match", Pattern: `dangerous`},
			},
			Action:  "block",
			Message: "Blocked!",
		},
	}

	inputData := map[string]any{
		"hook_event_name": "PreToolUse",
		"tool_name":       "Bash",
		"tool_input":      map[string]any{"command": "dangerous command"},
	}
	result := EvaluateRules(rules, inputData)
	if result.HookSpecificOutput["permissionDecision"] != "deny" {
		t.Errorf("expected permissionDecision 'deny', got '%v'", result.HookSpecificOutput["permissionDecision"])
	}
	if !strings.Contains(result.SystemMessage, "Blocked!") {
		t.Errorf("expected system message to contain 'Blocked!', got: %s", result.SystemMessage)
	}
}

func TestEvaluateRules_Warn(t *testing.T) {
	rules := []Rule{
		{
			Name:    "warn-test",
			Enabled: true,
			Event:   "bash",
			Conditions: []Condition{
				{Field: "command", Operator: "contains", Pattern: "warn"},
			},
			Action:  "warn",
			Message: "Warning!",
		},
	}

	inputData := map[string]any{
		"hook_event_name": "PreToolUse",
		"tool_name":       "Bash",
		"tool_input":      map[string]any{"command": "this will warn"},
	}
	result := EvaluateRules(rules, inputData)
	if result.Decision == "deny" {
		t.Error("expected rule to warn, not block")
	}
	if result.SystemMessage == "" {
		t.Error("expected warning message")
	}
}

func TestEvaluateRules_NoMatch(t *testing.T) {
	rules := []Rule{
		{
			Name:    "no-match",
			Enabled: true,
			Event:   "bash",
			Conditions: []Condition{
				{Field: "command", Operator: "equals", Pattern: "exact-match"},
			},
			Action:  "block",
			Message: "Should not match",
		},
	}

	inputData := map[string]any{
		"hook_event_name": "PreToolUse",
		"tool_name":       "Bash",
		"tool_input":      map[string]any{"command": "something else"},
	}
	result := EvaluateRules(rules, inputData)
	if result.Decision == "deny" {
		t.Error("expected no match")
	}
}

func TestEvaluateRules_DisabledRule(t *testing.T) {
	rules := []Rule{
		{
			Name:    "disabled",
			Enabled: false,
			Event:   "bash",
			Conditions: []Condition{
				{Field: "command", Operator: "regex_match", Pattern: `.*`},
			},
			Action:  "block",
			Message: "Should not trigger",
		},
	}

	inputData := map[string]any{
		"hook_event_name": "PreToolUse",
		"tool_name":       "Bash",
		"tool_input":      map[string]any{"command": "anything"},
	}
	result := EvaluateRules(rules, inputData)
	if result.Decision == "deny" {
		t.Error("expected disabled rule to not block")
	}
}

func TestRuleFromDict_SimplePattern(t *testing.T) {
	fm := map[string]interface{}{
		"name":    "test",
		"enabled": true,
		"event":   "bash",
		"pattern": `rm\s+-rf`,
		"action":  "block",
	}
	rule := ruleFromDict(fm, "Don't do that")
	if rule.Name != "test" {
		t.Errorf("expected name 'test', got %s", rule.Name)
	}
	if len(rule.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(rule.Conditions))
	}
	if rule.Conditions[0].Field != "command" {
		t.Errorf("expected field 'command', got %s", rule.Conditions[0].Field)
	}
	if rule.Conditions[0].Operator != "regex_match" {
		t.Errorf("expected operator 'regex_match', got %s", rule.Conditions[0].Operator)
	}
}

func TestMatchesTool(t *testing.T) {
	if !matchesTool("*", "Bash") {
		t.Error("expected wildcard to match any tool")
	}
	if !matchesTool("Bash|Write", "Bash") {
		t.Error("expected Bash to match")
	}
	if !matchesTool("Bash|Write", "Write") {
		t.Error("expected Write to match")
	}
	if matchesTool("Bash|Write", "Edit") {
		t.Error("expected Edit to not match")
	}
}
