package ralph

import (
	"encoding/json"
	"fmt"
	"juarvis/pkg/config"
	"juarvis/pkg/utils"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type LoopState struct {
	RootPath          string `yaml:"-"` // path base del ecosistema
	Active            bool   `yaml:"active"`
	Iteration         int    `yaml:"iteration"`
	MaxIterations     int    `yaml:"max_iterations"`
	CompletionPromise string `yaml:"completion_promise"`
	StartedAt         time.Time `yaml:"started_at"`
	Prompt            string `yaml:"-"`
}

func LoadState(rootPath string) (*LoopState, error) {
	stateFile := filepath.Join(rootPath, config.JuarvisDir, config.RalphStateFile)
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, err
	}

	content := string(data)
	fmRaw, prompt, found := utils.ExtractFrontmatterBlock(content)
	if !found {
		return nil, fmt.Errorf("no frontmatter found in Ralph state file")
	}

	state := &LoopState{
		RootPath: rootPath,
		Prompt:   prompt,
	}

	if err := yaml.Unmarshal([]byte(fmRaw), state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Ralph state: %w", err)
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
	stateFile := filepath.Join(s.RootPath, config.JuarvisDir, config.RalphStateFile)
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		return fmt.Errorf("failed to create directory for Ralph state: %w", err)
	}

	fmData, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal Ralph state: %w", err)
	}

	content := fmt.Sprintf("---\n%s---\n\n%s", string(fmData), s.Prompt)

	return os.WriteFile(stateFile, []byte(content), 0644)
}

func (s *LoopState) Delete() {
	stateFile := s.RootPath + "/" + config.JuarvisDir + "/" + config.RalphStateFile
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

func CreateLoopState(rootPath, prompt string, maxIterations int, completionPromise string) error {
	state := &LoopState{
		RootPath:          rootPath,
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
