package analyzer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"juarvis/pkg/config"
	"juarvis/pkg/output"
)

type Decision struct {
	Choice   string `json:"choice"`
	Reason   string `json:"reason"`
	File    string `json:"file,omitempty"`
}

type Bugfix struct {
	Error   string `json:"error"`
	Fix     string `json:"fix"`
	File   string `json:"file"`
}

type Pattern struct {
	Name      string `json:"name"`
	Count    int    `json:"count"`
	Contexts []string `json:"contexts"`
}

type FileChange struct {
	File   string `json:"file"`
	Action string `json:"action"` // created, modified, deleted
}

type SessionAnalysis struct {
	Decisions  []Decision  `json:"decisions"`
	Mistakes   []Bugfix   `json:"mistakes"`
	Patterns  []Pattern  `json:"patterns"`
	Files     []FileChange `json:"files_changed"`
	Timestamp time.Time  `json:"timestamp"`
}

var (
	decisionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(elegí|escogí|usé|preferí|decidí)\s+(\w+)\s+sobre\s+(\w+)`),
		regexp.MustCompile(`(?i)(porque|ya que|dado que)\s+([^\.]+)`),
		regexp.MustCompile(`(?i)decision:\s*([^\n]+)`),
		regexp.MustCompile(`(?i)elegante de\s+([^\.]+)`),
	}

	bugPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(error|fallo|failing|broken|bug)\s+([^\.]{5,50})`),
		regexp.MustCompile(`(?i)(arreglé|fixed|corregí)\s+([^\.]+)`),
		regexp.MustCompile(`(?i)(no funcionaba porque)\s+([^\.]+)`),
		regexp.MustCompile(`(?i)(estaba|was)\s+([^\.]{5,30})(error|fallo|broken)`),
	}

	fileActionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(creó|created)\s+([^\.]+\.(go|ts|js|py|md|json))`),
		regexp.MustCompile(`(?i)(modificó|modified|editó)\s+([^\.]+\.(go|ts|js|py|md|json))`),
		regexp.MustCompile(`(?i)(eliminó|deleted)\s+([^\.]+\.(go|ts|js|py|md|json))`),
		regexp.MustCompile(`(?i)Write\[([^\]]+)\]`),
		regexp.MustCompile(`(?i)Edit\(([^\)]+)`),
	}

	patternKeywords = []string{
		"hook", "context", "component", "service", "api", "middleware",
		"mutation", "query", "store", "slice", "reducer",
		"plugin", "skill", "mcp", "memory",
	}
)

func AnalyzeTranscript(path string) (*SessionAnalysis, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error abriendo transcript: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error leyendo transcript: %w", err)
	}

	analysis := &SessionAnalysis{
		Timestamp: time.Now(),
	}

	fullText := strings.Join(lines, "\n")

	// Extraer decisiones
	analysis.Decisions = extractDecisions(fullText)

	// Extraer errores/arreglos
	analysis.Mistakes = extractBugfixes(fullText)

	// Extraer cambios de archivos
	analysis.Files = extractFileChanges(fullText)

	// Detectar patrones
	analysis.Patterns = detectPatterns(fullText)

	return analysis, nil
}

func extractDecisions(text string) []Decision {
	var decisions []Decision

	for _, pattern := range decisionPatterns {
		matches := pattern.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 1 && len(match[1]) > 3 {
				decision := Decision{
					Choice: match[0],
					Reason: extractReason(match[0]),
				}
				decisions = append(decisions, decision)
			}
		}
	}

	// Deduplicar decisiones similares
	return deduplicateDecisions(decisions)
}

func extractReason(text string) string {
	text = strings.ToLower(text)
	if strings.Contains(text, "porque") {
		idx := strings.Index(text, "porque")
		if idx > 0 && idx < len(text)-8 {
			return strings.Trim(text[idx:idx+50], ".")
		}
	}
	if strings.Contains(text, "ya que") {
		idx := strings.Index(text, "ya que")
		if idx > 0 && idx < len(text)-8 {
			return strings.Trim(text[idx:idx+50], ".")
		}
	}
	return ""
}

func extractBugfixes(text string) []Bugfix {
	var fixes []Bugfix

	for _, pattern := range bugPatterns {
		matches := pattern.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 1 && len(match[1]) > 3 {
				fix := Bugfix{
					Error: match[0],
					Fix:   extractFix(match[0]),
				}
				fixes = append(fixes, fix)
			}
		}
	}

	return deduplicateBugfixes(fixes)
}

func extractFix(text string) string {
	text = strings.ToLower(text)
	if strings.Contains(text, "arreglé") || strings.Contains(text, "fixed") || strings.Contains(text, "corregí") {
		return "fixed"
	}
	return "detected"
}

func extractFileChanges(text string) []FileChange {
	var changes []FileChange

	for _, pattern := range fileActionPatterns {
		matches := pattern.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 1 {
				filename := strings.Trim(match[1], "[]()\"' ")
				if filename != "" && !strings.Contains(filename, "$") {
					action := "modified"
					if strings.Contains(match[0], "creó") || strings.Contains(match[0], "created") || strings.Contains(match[0], "Write") {
						action = "created"
					}
					if strings.Contains(match[0], "eliminó") || strings.Contains(match[0], "deleted") {
						action = "deleted"
					}
					changes = append(changes, FileChange{File: filename, Action: action})
				}
			}
		}
	}

	return deduplicateFiles(changes)
}

func detectPatterns(text string) []Pattern {
	text = strings.ToLower(text)
	var patterns []Pattern

	for _, keyword := range patternKeywords {
		count := strings.Count(text, keyword)
		if count >= 2 {
			pattern := Pattern{
				Name:   keyword,
				Count:  count,
				Contexts: extractContexts(text, keyword),
			}
			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

func extractContexts(text string, keyword string) []string {
	var contexts []string
	re := regexp.MustCompile(fmt.Sprintf(`(?i)(.{0,30}%s.{0,30})`, keyword))
	matches := re.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		if len(match) > 1 {
			contexts = append(contexts, strings.Trim(match[1], " \t"))
		}
	}
	if len(contexts) > 3 {
		contexts = contexts[:3]
	}
	return contexts
}

func deduplicateDecisions(decisions []Decision) []Decision {
	seen := make(map[string]bool)
	var result []Decision
	for _, d := range decisions {
		key := strings.TrimSpace(d.Choice)
		if key != "" && !seen[key] {
			seen[key] = true
			result = append(result, d)
		}
	}
	return result
}

func deduplicateBugfixes(fixes []Bugfix) []Bugfix {
	seen := make(map[string]bool)
	var result []Bugfix
	for _, f := range fixes {
		key := strings.TrimSpace(f.Error)
		if key != "" && !seen[key] {
			seen[key] = true
			result = append(result, f)
		}
	}
	return result
}

func deduplicateFiles(files []FileChange) []FileChange {
	seen := make(map[string]bool)
	var result []FileChange
	for _, f := range files {
		if !seen[f.File] {
			seen[f.File] = true
			result = append(result, f)
		}
	}
	return result
}

func SaveAnalysis(analysis *SessionAnalysis, projectPath string) error {
	juarDir := filepath.Join(projectPath, config.JuarDir, "memory")
	if err := os.MkdirAll(juarDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio memory: %w", err)
	}

	filename := filepath.Join(juarDir, fmt.Sprintf("session_%s.json", analysis.Timestamp.Format("2006-01-02_15-04-05")))
	data, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializando analysis: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("error guardando analysis: %w", err)
	}

	output.Info("✅ Análisis de sesión guardado: %s", filepath.Base(filename))
	return nil
}

func BuildSessionSummary(analysis *SessionAnalysis) string {
	var summary strings.Builder

	summary.WriteString("# Análisis de Sesión\n\n")

	if len(analysis.Decisions) > 0 {
		summary.WriteString("## Decisiones Tomadas\n")
		for _, d := range analysis.Decisions[:min(5, len(analysis.Decisions))] {
			summary.WriteString(fmt.Sprintf("- %s\n", d.Choice))
		}
		summary.WriteString("\n")
	}

	if len(analysis.Mistakes) > 0 {
		summary.WriteString("## Errores y Arreglos\n")
		for _, m := range analysis.Mistakes[:min(3, len(analysis.Mistakes))] {
			summary.WriteString(fmt.Sprintf("- %s → %s\n", m.Error, m.Fix))
		}
		summary.WriteString("\n")
	}

	if len(analysis.Patterns) > 0 {
		summary.WriteString("## Patrones Usados\n")
		for _, p := range analysis.Patterns[:min(5, len(analysis.Patterns))] {
			summary.WriteString(fmt.Sprintf("- %s (%d veces)\n", p.Name, p.Count))
		}
		summary.WriteString("\n")
	}

	if len(analysis.Files) > 0 {
		summary.WriteString("## Archivos Modificados\n")
		for _, f := range analysis.Files[:min(10, len(analysis.Files))] {
			summary.WriteString(fmt.Sprintf("- %s: %s\n", f.Action, f.File))
		}
	}

	return summary.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}