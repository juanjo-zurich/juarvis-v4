package ralph

import (
	"encoding/json"
	"fmt"
	"juarvis/pkg/config"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type LoopState struct {
	Active            bool
	Iteration         int
	MaxIterations     int
	CompletionPromise string
	StartedAt         time.Time
	Prompt            string
}

const stateFile = config.OpencodeDir + "/" + config.RalphStateFile

func parseFrontmatter(content string) (map[string]string, string) {
	if !strings.HasPrefix(content, "---") {
		return nil, content
	}

	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, content
	}

	fm := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(parts[1]), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, ":"); idx >= 0 {
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			val = strings.Trim(val, "\"")
			fm[key] = val
		}
	}

	return fm, strings.TrimSpace(parts[2])
}

func LoadState() (*LoopState, error) {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, err
	}

	fm, prompt := parseFrontmatter(string(data))

	state := &LoopState{Prompt: prompt}

	if v, ok := fm["active"]; ok {
		state.Active = v == "true"
	}
	if v, ok := fm["iteration"]; ok {
		state.Iteration, _ = strconv.Atoi(v)
	}
	if v, ok := fm["max_iterations"]; ok {
		state.MaxIterations, _ = strconv.Atoi(v)
	}
	if v, ok := fm["completion_promise"]; ok {
		if v == "null" {
			state.CompletionPromise = ""
		} else {
			state.CompletionPromise = v
		}
	}
	if v, ok := fm["started_at"]; ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			state.StartedAt = t
		}
	}

	return state, nil
}

func (s *LoopState) IsActive() bool {
	return s.Active
}

func (s *LoopState) IsComplete() bool {
	if s.MaxIterations > 0 && s.Iteration >= s.MaxIterations {
		return true
	}
	return false
}

func (s *LoopState) Save() error {
	os.MkdirAll(filepath.Dir(stateFile), 0755)

	promiseYAML := "null"
	if s.CompletionPromise != "" {
		promiseYAML = fmt.Sprintf("\"%s\"", s.CompletionPromise)
	}

	content := fmt.Sprintf(`---
active: true
iteration: %d
max_iterations: %d
completion_promise: %s
started_at: "%s"
---

%s
`, s.Iteration, s.MaxIterations, promiseYAML, s.StartedAt, s.Prompt)

	return os.WriteFile(stateFile, []byte(content), 0644)
}

func (s *LoopState) Delete() {
	os.Remove(stateFile)
}

func (s *LoopState) Increment() {
	s.Iteration++
}

func CheckCompletionPromise(lastOutput, promise string) bool {
	if promise == "" {
		return false
	}

	re := regexp.MustCompile(`(?s)<promise>(.*?)</promise>`)
	matches := re.FindStringSubmatch(lastOutput)
	if len(matches) < 2 {
		return false
	}

	promiseText := strings.TrimSpace(matches[1])
	// Normalize whitespace
	promiseText = regexp.MustCompile(`\s+`).ReplaceAllString(promiseText, " ")
	return promiseText == promise
}

func ExtractLastAssistantMessage(transcriptPath string) (string, error) {
	data, err := os.ReadFile(transcriptPath)
	if err != nil {
		return "", fmt.Errorf("cannot read transcript: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var lastAssistantLine string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, `"role":"assistant"`) {
			lastAssistantLine = line
		}
	}

	if lastAssistantLine == "" {
		return "", fmt.Errorf("no assistant messages found in transcript")
	}

	var msg struct {
		Message struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"message"`
	}

	if err := json.Unmarshal([]byte(lastAssistantLine), &msg); err != nil {
		return "", fmt.Errorf("failed to parse assistant message: %w", err)
	}

	var texts []string
	for _, c := range msg.Message.Content {
		if c.Type == "text" {
			texts = append(texts, c.Text)
		}
	}

	if len(texts) == 0 {
		return "", fmt.Errorf("assistant message contains no text content")
	}

	return strings.Join(texts, "\n"), nil
}

func CreateLoopState(prompt string, maxIterations int, completionPromise string) error {
	state := &LoopState{
		Active:            true,
		Iteration:         1,
		MaxIterations:     maxIterations,
		CompletionPromise: completionPromise,
		StartedAt:         time.Now().UTC(),
		Prompt:            prompt,
	}

	return state.Save()
}

func BuildStopResponse(state *LoopState, transcriptPath string) (map[string]any, error) {
	transcriptExists := false
	if _, err := os.Stat(transcriptPath); err == nil {
		transcriptExists = true
	}

	if transcriptExists {
		lastOutput, err := ExtractLastAssistantMessage(transcriptPath)
		if err == nil && state.CompletionPromise != "" {
			if CheckCompletionPromise(lastOutput, state.CompletionPromise) {
				state.Delete()
				return map[string]any{
					"decision":      "allow",
					"systemMessage": fmt.Sprintf("✅ Ralph loop: Detected <promise>%s</promise>", state.CompletionPromise),
				}, nil
			}
		}
	}

	state.Increment()
	if err := state.Save(); err != nil {
		return nil, fmt.Errorf("failed to save state: %w", err)
	}

	var sysMsg string
	if state.CompletionPromise != "" {
		sysMsg = fmt.Sprintf("🔄 Ralph iteration %d | To stop: output <promise>%s</promise> (ONLY when statement is TRUE - do not lie to exit!)",
			state.Iteration, state.CompletionPromise)
	} else {
		sysMsg = fmt.Sprintf("🔄 Ralph iteration %d | No completion promise set - loop runs infinitely", state.Iteration)
	}

	return map[string]any{
		"decision":      "block",
		"reason":        state.Prompt,
		"systemMessage": sysMsg,
	}, nil
}
