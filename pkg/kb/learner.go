// Package kb implementa la Knowledge Base de Juarvis.
// Learner implementa la lógica para aprender patrones del trabajo realizado.
package kb

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Learner analiza el trabajo realizado y aprende nuevos patrones.
type Learner struct {
	kb        *KnowledgeBase
	config   *LearnerConfig
	mu       sync.Mutex
	rules    []LearningRule
	autoSave bool
}

// LearnerConfig contiene la configuración del learner.
type LearnerConfig struct {
	AutoLearnEnabled   bool
	MinConfidence     float64
	MaxSuggestions    int
	SuggestionTimeout  time.Duration
	PatternsDir       string
	LearningRulesFile string
}

// DefaultLearnerConfig devuelve la configuración por defecto.
func DefaultLearnerConfig() *LearnerConfig {
	return &LearnerConfig{
		AutoLearnEnabled:   true,
		MinConfidence:   0.7,
		MaxSuggestions:  5,
		SuggestionTimeout: 30 * time.Second,
		PatternsDir:      ".juarvis/knowledge/patterns",
		LearningRulesFile: ".juarvis/knowledge/rules.json",
	}
}

// NewLearner crea un nuevo Learner.
func NewLearner(kb *KnowledgeBase, config *LearnerConfig) (*Learner, error) {
	if config == nil {
		config = DefaultLearnerConfig()
	}

	l := &Learner{
		kb:       kb,
		config:   config,
		rules:    []LearningRule{},
		autoSave: config.AutoLearnEnabled,
	}

	// Cargar reglas existentes
	if err := l.loadRules(); err != nil {
		fmt.Printf("Warning: Could not load rules: %v\n", err)
	}

	return l, nil
}

// LearningRule representa una regla de aprendizaje.
type LearningRule struct {
	ID          string          `json:"id"`
	Name       string          `json:"name"`
	Pattern    string          `json:"pattern"`
	Type      KnowledgeType   `json:"type"`
	Priority  int            `json:"priority"`
	Condition string         `json:"condition"` // Expresión regular o condición
	Action    LearningAction `json:"action"`
	Enabled   bool           `json:"enabled"`
	CreatedAt time.Time      `json:"created_at"`
}

// LearningAction define qué aprender de un patrón detectado.
type LearningAction struct {
	Type      KnowledgeType `json:"type"`
	Extract  string       `json:"extract"` // "code", "context", "steps", etc.
	Template string       `json:"template"`
}

// LearnFromTask aprende de una tarea completada exitosamente.
func (l *Learner) LearnFromTask(task *TaskContext) error {
	if task == nil {
		return fmt.Errorf("task context is nil")
	}

	if !l.config.AutoLearnEnabled {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Determinar tipo de conocimiento a extraer
	ktype := l.determineKnowledgeType(task)

	// Extraer conocimiento según el tipo
	entry := l.extractKnowledge(task, ktype)
	if entry == nil {
		return fmt.Errorf("could not extract knowledge from task")
	}

	entry.Confidence = l.calculateConfidence(task)
	entry.LearnedFrom = task.TaskID

	// Aplicar reglas de aprendizaje
	l.applyRules(task, entry)

	// Añadir a la Knowledge Base
	if err := l.kb.Add(entry); err != nil {
		return fmt.Errorf("error adding knowledge: %w", err)
	}

	return nil
}

// TaskContext contiene el contexto de una tarea completada.
type TaskContext struct {
	TaskID       string
	TaskName     string
	Description string
	Files        []FileContext
	Language    string
	Framework   string
	Tags        []string
	Success     bool
	Duration    time.Duration
	Metadata    map[string]any
}

// FileContext contiene el contexto de un archivo procesado.
type FileContext struct {
	Path         string
	Content     string
	Language    string
	Framework   string
	Changes     []ChangeContext
	ChangeCount int
}

// ChangeContext contiene el contexto de un cambio realizado.
type ChangeContext struct {
	Type        string // "add", "modify", "delete"
	Before      string
	After       string
	LineNumbers string
	Description string
}

// determineKnowledgeType determina el tipo de conocimiento apropiada.
func (l *Learner) determineKnowledgeType(task *TaskContext) KnowledgeType {
	// Analizar el contexto para determinar el tipo
	description := strings.ToLower(task.Description)

	switch {
	case strings.Contains(description, "bug") || strings.Contains(description, "fix") ||
		strings.Contains(description, "error") || strings.Contains(description, "corregir"):
		return KnowledgeTypeBugFix
	case strings.Contains(description, "estructura") || strings.Contains(description, "architecture") ||
		strings.Contains(description, "arquitectura"):
		return KnowledgeTypeArchitecture
	case strings.Contains(description, "workflow") || strings.Contains(description, "flujo") ||
		strings.Contains(description, "pasos") || strings.Contains(description, "proceso"):
		return KnowledgeTypeWorkflow
	case strings.Contains(description, "decisión") || strings.Contains(description, "decision") ||
		strings.Contains(description, "diseño") || strings.Contains(description, "design"):
		return KnowledgeTypeDecision
	default:
		// Por defecto, intentar detectar código
		if len(task.Files) > 0 && hasCode(task.Files[0].Content) {
			return KnowledgeTypeCodePattern
		}
		return KnowledgeTypeCodePattern
	}
}

// hasCode detecta si el contenido contiene código.
func hasCode(content string) bool {
	codeIndicators := []string{"func ", "function ", "class ", "struct ", "def ", "const ", "var ",
		"if ", "for ", "while ", "return ", "import ", "package "}

	contentLower := strings.ToLower(content)
	for _, indicator := range codeIndicators {
		if strings.Contains(contentLower, indicator) {
			return true
		}
	}
	return false
}

// extractKnowledge extrae conocimiento del contexto de la tarea.
func (l *Learner) extractKnowledge(task *TaskContext, ktype KnowledgeType) *KnowledgeEntry {
	var entry *KnowledgeEntry

	switch ktype {
	case KnowledgeTypeBugFix:
		entry = l.extractBugFix(task)
	case KnowledgeTypeArchitecture:
		entry = l.extractArchitecture(task)
	case KnowledgeTypeWorkflow:
		entry = l.extractWorkflow(task)
	case KnowledgeTypeDecision:
		entry = l.extractDecision(task)
	case KnowledgeTypeCodePattern:
		entry = l.extractCodePattern(task)
	default:
		entry = l.extractCodePattern(task)
	}

	// Añadir metadata común
	if entry != nil {
		entry.Language = task.Language
		entry.Framework = task.Framework
		entry.Tags = task.Tags
		entry.Source = "auto-learned"
	}

	return entry
}

// extractBugFix extrae un bug fix del contexto.
func (l *Learner) extractBugFix(task *TaskContext) *KnowledgeEntry {
	entry := NewKnowledgeEntry(KnowledgeTypeBugFix, task.TaskName, task.Description)

	// Buscar cambios que describan un fix
	var problem, solution string
	for _, file := range task.Files {
		for _, change := range file.Changes {
			if change.Type == "modify" {
				if problem == "" && change.Before != "" {
					problem = change.Before
				}
				if solution == "" && change.After != "" {
					solution = change.After
				}
			}
		}
	}

	entry.Problem = problem
	entry.Solution = solution

	// Extraer keywords del contenido
	entry.Keywords = extractKeywords(task.Description)
	entry.Keywords = append(entry.Keywords, task.Tags...)

	return entry
}

// extractArchitecture extrae una decisión arquitectónica.
func (l *Learner) extractArchitecture(task *TaskContext) *KnowledgeEntry {
	entry := NewKnowledgeEntry(KnowledgeTypeArchitecture, task.TaskName, task.Description)

	// Añadir contexto arquitectónico
	var context strings.Builder
	for _, file := range task.Files {
		if file.ChangeCount > 0 {
			context.WriteString(fmt.Sprintf("\n- %s: %d cambios", file.Path, file.ChangeCount))
		}
	}
	entry.Context = context.String()

	// Extraer keywords
	entry.Keywords = extractKeywords(task.Description)
	entry.Keywords = append(entry.Keywords, "arquitectura", "estructura")

	return entry
}

// extractWorkflow extrae un workflow.
func (l *Learner) extractWorkflow(task *TaskContext) *KnowledgeEntry {
	entry := NewKnowledgeEntry(KnowledgeTypeWorkflow, task.TaskName, task.Description)

	// Extraer pasos del contexto
	var steps []string
	for _, file := range task.Files {
		if file.ChangeCount > 0 {
			steps = append(steps, fmt.Sprintf("Modificar %s (%d cambios)", file.Path, file.ChangeCount))
		}
	}

	if len(steps) == 0 {
		steps = append(steps, task.Description)
	}

	entry.Steps = steps
	entry.Keywords = extractKeywords(task.Description)
	entry.Keywords = append(entry.Keywords, "workflow", "flujo", "proceso")

	return entry
}

// extractDecision extrae una decisión de diseño.
func (l *Learner) extractDecision(task *TaskContext) *KnowledgeEntry {
	entry := NewKnowledgeEntry(KnowledgeTypeDecision, task.TaskName, task.Description)

	// Extraer justificación del contexto
	var justification strings.Builder
	for _, file := range task.Files {
		if file.ChangeCount > 0 {
			justification.WriteString(fmt.Sprintf("\nArchivo afectado: %s", file.Path))
		}
	}
	entry.Justification = justification.String()
	entry.Keywords = extractKeywords(task.Description)
	entry.Keywords = append(entry.Keywords, "decisión", "diseño", "design")

	return entry
}

// extractCodePattern extrae un patrón de código.
func (l *Learner) extractCodePattern(task *TaskContext) *KnowledgeEntry {
	entry := NewKnowledgeEntry(KnowledgeTypeCodePattern, task.TaskName, task.Description)

	// Buscar código en los cambios
	var pattern strings.Builder
	for _, file := range task.Files {
		for _, change := range file.Changes {
			if change.Type == "add" && change.After != "" {
				pattern.WriteString(change.After)
				break
			}
		}
		if pattern.Len() > 0 {
			break
		}
	}

	if pattern.Len() > 0 {
		entry.Pattern = pattern.String()
	}

	// Extraer keywords
	entry.Keywords = extractKeywords(task.Description)
	if task.Language != "" {
		entry.Keywords = append(entry.Keywords, task.Language)
	}
	if task.Framework != "" {
		entry.Keywords = append(entry.Keywords, task.Framework)
	}

	return entry
}

// extractKeywords extrae keywords del texto.
func extractKeywords(text string) []string {
	// Palabras comunes a excluir
	stopWords := map[string]bool{
		"el": true, "la": true, "los": true, "las": true, "un": true, "una": true,
		"de": true, "del": true, "en": true, "y": true, "o": true, "a": true,
		"que": true, "es": true, "son": true, "para": true, "por": true, "con": true,
		"sin": true, "sobre": true, "entre": true, "como": true, "pero": true,
		"este": true, "esta": true, "estos": true, "estas": true, "ese": true,
"esa": true, "al": true, "se": true, "le": true, "lo": true, "me": true,
	"the": true, "and": true, "or": true, "of": true, "to": true, "in": true,
	"an": true, "is": true, "are": true, "was": true, "were": true,
		"for": true, "with": true, "on": true, "at": true, "by": true, "from": true,
		"it": true, "this": true, "that": true, "be": true, "as": true, "have": true,
		"has": true, "had": true, "not": true, "all": true, "we": true, "you": true,
		"do": true, "does": true, "did": true, "can": true, "could": true, "will": true,
		"would": true, "should": true, "may": true, "might": true, "must": true,
		"implementar": true, "crear": true, "hacer": true, "añadir": true, "agregar": true,
		"modificar": true, "eliminar": true, "actualizar": true, "buscar": true,
		"obtener": true, "devolver": true, "guardar": true, "cargar": true,
	}

	// Limpiar y tokenizar
	cleanText := strings.ToLower(text)
	cleanText = regexp.MustCompile(`[^\w\s]`).ReplaceAllString(cleanText, " ")
	words := strings.Fields(cleanText)

	// Filtrar keywords
	keywords := make([]string, 0)
	seen := make(map[string]bool)

	for _, word := range words {
		// Excluir palabras cortas o stop words
		if len(word) < 3 {
			continue
		}
		if stopWords[word] {
			continue
		}
		if seen[word] {
			continue
		}
		if _, err := strconv.Atoi(word); err == nil {
			continue // Excluir números
		}

		seen[word] = true
		keywords = append(keywords, word)
	}

	// Limitar keywords
	if len(keywords) > 20 {
		keywords = keywords[:20]
	}

	return keywords
}

// calculateConfidence calcula la confianza del conocimiento extraído.
func (l *Learner) calculateConfidence(task *TaskContext) float64 {
	confidence := 0.5

	// Más archivos = más confianza
	if len(task.Files) >= 3 {
		confidence += 0.2
	} else if len(task.Files) >= 1 {
		confidence += 0.1
	}

	// Más cambios = más confianza
	totalChanges := 0
	for _, file := range task.Files {
		totalChanges += file.ChangeCount
	}
	if totalChanges >= 10 {
		confidence += 0.15
	} else if totalChanges >= 5 {
		confidence += 0.1
	}

	// Éxito conocido = más confianza
	if task.Success {
		confidence += 0.1
	}

	// Duración razonable = más confianza
	if task.Duration > time.Minute && task.Duration < time.Hour {
		confidence += 0.05
	}

	// Clamp a 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// applyRules aplica las reglas de aprendizaje.
func (l *Learner) applyRules(task *TaskContext, entry *KnowledgeEntry) {
	for _, rule := range l.rules {
		if !rule.Enabled {
			continue
		}

		// Verificar condición
		if rule.Condition != "" {
			matched, err := regexp.MatchString(rule.Condition, task.Description)
			if err != nil || !matched {
				continue
			}
		}

		// Aplicar acción
		switch rule.Action.Extract {
		case "title":
			entry.Title = fmt.Sprintf(rule.Action.Template, task.TaskName)
		case "keywords":
			newKeywords := extractKeywords(rule.Action.Template)
			entry.Keywords = append(entry.Keywords, newKeywords...)
		case "confidence":
			entry.Confidence += 0.1
		}
	}
}

// AddRule añade una regla de aprendizaje.
func (l *Learner) AddRule(rule *LearningRule) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now()
	}

	l.rules = append(l.rules, *rule)

	return l.saveRules()
}

// RemoveRule elimina una regla de aprendizaje.
func (l *Learner) RemoveRule(id string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i, rule := range l.rules {
		if rule.ID == id {
			l.rules = append(l.rules[:i], l.rules[i+1:]...)
			return l.saveRules()
		}
	}

	return fmt.Errorf("rule not found: %s", id)
}

// EnableRule habilita o deshabilita una regla.
func (l *Learner) EnableRule(id string, enabled bool) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i, rule := range l.rules {
		if rule.ID == id {
			l.rules[i].Enabled = enabled
			return l.saveRules()
		}
	}

	return fmt.Errorf("rule not found: %s", id)
}

// loadRules carga las reglas desde archivo.
func (l *Learner) loadRules() error {
	data, err := os.ReadFile(l.config.LearningRulesFile)
	if os.IsNotExist(err) {
		l.rules = l.defaultRules()
		return nil
	}
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &l.rules)
}

// saveRules guarda las reglas en archivo.
func (l *Learner) saveRules() error {
	dir := filepath.Dir(l.config.LearningRulesFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(l.rules, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(l.config.LearningRulesFile, data, 0644)
}

// defaultRules devuelve las reglas por defecto.
func (l *Learner) defaultRules() []LearningRule {
	return []LearningRule{
		{
			ID:     "default-bugfix",
			Name:  "Detectar Bug Fixes",
			Pattern: `(?i)(fix|bug|corregir|error)`,
			Type: KnowledgeTypeBugFix,
			Priority: 10,
			Condition: `(?i)(fix|bug|corregir|error)`,
			Action: LearningAction{
				Type:     KnowledgeTypeBugFix,
				Extract: "keywords",
			},
			Enabled:   true,
			CreatedAt: time.Now(),
		},
		{
			ID:     "default-architecture",
			Name:  "Detectar Arquitectura",
			Pattern: `(?i)(estructura|architecture|arquitectura|diseño)`,
			Type: KnowledgeTypeArchitecture,
			Priority: 8,
			Condition: `(?i)(estructura|architecture|arquitectura)`,
			Action: LearningAction{
				Type:     KnowledgeTypeArchitecture,
				Extract: "keywords",
			},
			Enabled:   true,
			CreatedAt: time.Now(),
		},
	}
}

// GetRules devuelve las reglas de aprendizaje.
func (l *Learner) GetRules() []LearningRule {
	l.mu.Lock()
	defer l.mu.Unlock()

	rules := make([]LearningRule, len(l.rules))
	copy(rules, l.rules)

	return rules
}

// AnalyzeTask analiza una tarea y sugiere qué aprender.
func (l *Learner) AnalyzeTask(task *TaskContext) *AnalysisResult {
	result := &AnalysisResult{
		TaskID:     task.TaskID,
		Suggestions: []LearningSuggestion{},
	}

	// Determinar tipo
	ktype := l.determineKnowledgeType(task)
	result.DetectedType = ktype

	// Analizar keywords
	result.Keywords = extractKeywords(task.Description)

	// Analizar archivos
	result.FileCount = len(task.Files)
	result.TotalChanges = 0
	for _, file := range task.Files {
		result.TotalChanges += file.ChangeCount
	}

	// Hacer sugerencias
	if result.FileCount >= 1 && result.TotalChanges >= 3 {
		result.Suggestions = append(result.Suggestions, LearningSuggestion{
			Type:       ktype,
			Confidence: l.calculateConfidence(task),
			Reason:    "Suficiente contenido para aprender",
		})
	}

	return result
}

// AnalysisResult contiene el resultado del análisis.
type AnalysisResult struct {
	TaskID       string
	DetectedType KnowledgeType
	Keywords    []string
	FileCount   int
	TotalChanges int
	Suggestions []LearningSuggestion
}

// LearningSuggestion es una sugerencia de aprendizaje.
type LearningSuggestion struct {
	Type       KnowledgeType
	Confidence float64
	Reason    string
}

// SuggestPatterns sugiere patrones relevantes para una tarea.
func (l *Learner) SuggestPatterns(context *TaskContext) ([]*KnowledgeEntry, error) {
	kbContext := &Context{
		Task:       context.TaskName,
		Language:  context.Language,
		Framework: context.Framework,
		Keywords: extractKeywords(context.Description),
		Tags:     context.Tags,
	}

	return l.kb.GetSuggestions(kbContext), nil
}

// AutoLearnFromCompleteTask aprende automáticamente de tareas completadas.
func (l *Learner) AutoLearnFromCompleteTask(task *TaskContext) {
	if task == nil || !task.Success {
		return
	}

	if err := l.LearnFromTask(task); err != nil {
		fmt.Printf("Auto-learn error: %v\n", err)
	}
}