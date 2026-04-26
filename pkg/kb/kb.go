// Package kb implementa la Knowledge Base de Juarvis.
// Permite a los agentes aprender de su trabajo pasado y guardar conocimiento para tareas futuras.
package kb

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// KnowledgeType representa los tipos de conocimiento soportados.
type KnowledgeType string

const (
	KnowledgeTypeCodePattern  KnowledgeType = "code_pattern"
	KnowledgeTypeArchitecture KnowledgeType = "architecture"
	KnowledgeTypeWorkflow     KnowledgeType = "workflow"
	KnowledgeTypeBugFix       KnowledgeType = "bug_fix"
	KnowledgeTypeDecision     KnowledgeType = "decision"
)

// KnowledgeEntry representa una entrada de conocimiento en la base de datos.
type KnowledgeEntry struct {
	ID          string        `json:"id"`
	Type       KnowledgeType `json:"type"`
	Title      string        `json:"title"`
	Content    string        `json:"content"`
	Language   string        `json:"language,omitempty"`
	Framework  string        `json:"framework,omitempty"`
	Keywords  []string     `json:"keywords"`
	Context   string        `json:"context,omitempty"`
	Justification string    `json:"justification,omitempty"` // Para decisiones
	Problem    string       `json:"problem,omitempty"`      // Para bug fixes
	Solution   string       `json:"solution,omitempty"`   // Para bug fixes
	Steps      []string     `json:"steps,omitempty"`    // Para workflows
	Pattern    string      `json:"pattern,omitempty"`  // Para code patterns
	Tags       []string    `json:"tags"`
	Source     string      `json:"source,omitempty"`   // De dónde se obtuvo
	Confidence float64     `json:"confidence"`          // 0.0 - 1.0
	UsageCount int         `json:"usage_count"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	LearnedFrom string   `json:"learned_from,omitempty"` // Tarea de dónde se aprendió
}

// NewKnowledgeEntry crea una nueva entrada de conocimiento con ID único.
func NewKnowledgeEntry(ktype KnowledgeType, title, content string) *KnowledgeEntry {
	now := time.Now()
	return &KnowledgeEntry{
		ID:          uuid.New().String(),
		Type:        ktype,
		Title:       title,
		Content:     content,
		Keywords:    []string{},
		Tags:        []string{},
		Confidence:  0.5,
		UsageCount:  0,
		Metadata:   make(map[string]any),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// KnowledgeBase gestiona el almacenamiento y recuperación de conocimiento.
type KnowledgeBase struct {
	mu       sync.RWMutex
	entries  map[string]*KnowledgeEntry
	index    *InvertedIndex
	storage  *JSONStorage
	config   *Config
}

// Config contiene la configuración de la Knowledge Base.
type Config struct {
	StoragePath     string
	MaxEntries      int
	IndexEnabled    bool
	AutoSave        bool
	AutoSaveInterval time.Duration
}

// DefaultConfig devuelve la configuración por defecto.
func DefaultConfig() *Config {
	return &Config{
		StoragePath:     ".juarvis/knowledge",
		MaxEntries:     10000,
		IndexEnabled:   true,
		AutoSave:       true,
		AutoSaveInterval: 5 * time.Minute,
	}
}

// NewKnowledgeBase crea una nueva Knowledge Base.
func NewKnowledgeBase(config *Config) (*KnowledgeBase, error) {
	if config == nil {
		config = DefaultConfig()
	}

	kb := &KnowledgeBase{
		entries: make(map[string]*KnowledgeEntry),
		index:  NewInvertedIndex(),
		config: config,
	}

	// Inicializar almacenamiento
	storage, err := NewJSONStorage(config.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("error initializing storage: %w", err)
	}
	kb.storage = storage

	// Cargar entradas existentes
	if err := kb.load(); err != nil {
		fmt.Printf("Warning: Could not load existing knowledge: %v\n", err)
	}

	return kb, nil
}

// Add añade una nueva entrada de conocimiento.
func (kb *KnowledgeBase) Add(entry *KnowledgeEntry) error {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	entry.UpdatedAt = time.Now()

	// Añadir al mapa de entradas
	kb.entries[entry.ID] = entry

	// Indexar para búsqueda
	if kb.config.IndexEnabled {
		kb.index.Add(entry)
	}

	// Auto-guardar si está habilitado
	if kb.config.AutoSave {
		go kb.save()
	}

	return nil
}

// Update actualiza una entrada existente.
func (kb *KnowledgeBase) Update(entry *KnowledgeEntry) error {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	existing, ok := kb.entries[entry.ID]
	if !ok {
		return fmt.Errorf("entry not found: %s", entry.ID)
	}

	// Re-indexar si es necesario
	if kb.config.IndexEnabled {
		kb.index.Remove(existing)
	}

	entry.UpdatedAt = time.Now()
	entry.UsageCount = existing.UsageCount
	entry.CreatedAt = existing.CreatedAt
	kb.entries[entry.ID] = entry

	if kb.config.IndexEnabled {
		kb.index.Add(entry)
	}

	if kb.config.AutoSave {
		go kb.save()
	}

	return nil
}

// Delete elimina una entrada de conocimiento.
func (kb *KnowledgeBase) Delete(id string) error {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	entry, ok := kb.entries[id]
	if !ok {
		return fmt.Errorf("entry not found: %s", id)
	}

	if kb.config.IndexEnabled {
		kb.index.Remove(entry)
	}

	delete(kb.entries, id)

	if kb.config.AutoSave {
		go kb.save()
	}

	return nil
}

// Get recupera una entrada por su ID.
func (kb *KnowledgeBase) Get(id string) (*KnowledgeEntry, bool) {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	entry, ok := kb.entries[id]
	if !ok {
		return nil, false
	}

	// Incrementar contador de uso
	entry.UsageCount++

	return entry, true
}

// GetAll devuelve todas las entradas.
func (kb *KnowledgeBase) GetAll() []*KnowledgeEntry {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	entries := make([]*KnowledgeEntry, 0, len(kb.entries))
	for _, entry := range kb.entries {
		entries = append(entries, entry)
	}

	slices.SortFunc(entries, func(a, b *KnowledgeEntry) int {
		if a.UpdatedAt.Before(b.UpdatedAt) {
			return 1
		}
		if a.UpdatedAt.After(b.UpdatedAt) {
			return -1
		}
		return 0
	})

	return entries
}

// Search busca entradas que coincidan con la query.
func (kb *KnowledgeBase) Search(query string, filters *SearchFilters) ([]*KnowledgeEntry, error) {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	var results []*KnowledgeEntry

	if kb.config.IndexEnabled && kb.index != nil {
		// Usar índice invertido para búsqueda rápida
		results = kb.index.Search(query)
	} else {
		// Búsqueda lineal fallback
		lowerQuery := strings.ToLower(query)
		for _, entry := range kb.entries {
			if strings.Contains(strings.ToLower(entry.Title), lowerQuery) ||
				strings.Contains(strings.ToLower(entry.Content), lowerQuery) {
				results = append(results, entry)
			}
		}
	}

	// Aplicar filtros
	if filters != nil {
		results = kb.applyFilters(results, filters)
	}

	// Ordenar por relevancia
	slices.SortFunc(results, func(a, b *KnowledgeEntry) int {
		if a.Confidence != b.Confidence {
			if a.Confidence > b.Confidence {
				return -1
			}
			return 1
		}
		if a.UsageCount != b.UsageCount {
			if a.UsageCount > b.UsageCount {
				return -1
			}
			return 1
		}
		return 0
	})

	return results, nil
}

// SearchFilters contiene filtros para la búsqueda.
type SearchFilters struct {
	Types      []KnowledgeType
	Languages  []string
	Frameworks []string
	Keywords  []string
	Tags      []string
	MinConfidence float64
	Limit     int
}

// applyFilters aplica los filtros a los resultados.
func (kb *KnowledgeBase) applyFilters(entries []*KnowledgeEntry, filters *SearchFilters) []*KnowledgeEntry {
	var filtered []*KnowledgeEntry

	for _, entry := range entries {
		// Filtrar por tipo
		if len(filters.Types) > 0 {
			if !slices.Contains(filters.Types, entry.Type) {
				continue
			}
		}

		// Filtrar por lenguaje
		if len(filters.Languages) > 0 {
			if !slices.Contains(filters.Languages, entry.Language) {
				continue
			}
		}

		// Filtrar por framework
		if len(filters.Frameworks) > 0 {
			if !slices.Contains(filters.Frameworks, entry.Framework) {
				continue
			}
		}

		// Filtrar por keywords
		if len(filters.Keywords) > 0 {
			found := false
			for _, kw := range filters.Keywords {
				kwLower := strings.ToLower(kw)
				for _, entryKw := range entry.Keywords {
					if strings.Contains(strings.ToLower(entryKw), kwLower) {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filtrar por tags
		if len(filters.Tags) > 0 {
			found := false
			for _, tag := range filters.Tags {
				if slices.Contains(entry.Tags, tag) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filtrar por confianza mínima
		if filters.MinConfidence > 0 && entry.Confidence < filters.MinConfidence {
			continue
		}

		filtered = append(filtered, entry)
	}

	// Aplicar límite
	if filters.Limit > 0 && len(filtered) > filters.Limit {
		filtered = filtered[:filters.Limit]
	}

	return filtered
}

// GetByType devuelve entradas de un tipo específico.
func (kb *KnowledgeBase) GetByType(ktype KnowledgeType) []*KnowledgeEntry {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	var results []*KnowledgeEntry
	for _, entry := range kb.entries {
		if entry.Type == ktype {
			results = append(results, entry)
		}
	}

	return results
}

// GetSuggestions devuelve sugerencias basadas en el contexto actual.
func (kb *KnowledgeBase) GetSuggestions(context *Context) []*KnowledgeEntry {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	var suggestions []*KnowledgeEntry

	// Buscar por keywords del contexto
	for _, kw := range context.Keywords {
		if results := kb.index.Search(kw); len(results) > 0 {
			for _, entry := range results {
				// Añadir si no está ya en sugerencias
				found := false
				for _, s := range suggestions {
					if s.ID == entry.ID {
						found = true
						break
					}
				}
				if !found {
					suggestions = append(suggestions, entry)
				}
			}
		}
	}

	// Buscar por lenguaje/framework
	if context.Language != "" {
		for _, entry := range kb.entries {
			if entry.Language == context.Language {
				found := false
				for _, s := range suggestions {
					if s.ID == entry.ID {
						found = true
						break
					}
				}
				if !found {
					suggestions = append(suggestions, entry)
				}
			}
		}
	}

	// Ordenar por confianza y uso
	slices.SortFunc(suggestions, func(a, b *KnowledgeEntry) int {
		if a.Confidence != b.Confidence {
			if a.Confidence > b.Confidence {
				return -1
			}
			return 1
		}
		if a.UsageCount != b.UsageCount {
			if a.UsageCount > b.UsageCount {
				return -1
			}
			return 1
		}
		return 0
	})

	// Limitar sugerencias
	if len(suggestions) > 10 {
		suggestions = suggestions[:10]
	}

	return suggestions
}

// Context representa el contexto para sugerencias.
type Context struct {
	Task       string
	Language  string
	Framework string
	Keywords []string
	Tags     []string
}

// GetStats devuelve estadísticas de la Knowledge Base.
func (kb *KnowledgeBase) GetStats() *Stats {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	stats := &Stats{
		TotalEntries:    len(kb.entries),
		ByType:        make(map[KnowledgeType]int),
		ByLanguage:    make(map[string]int),
		ByFramework:  make(map[string]int),
		AvgConfidence: 0,
		TotalUsage:    0,
	}

	var totalConfidence float64

	for _, entry := range kb.entries {
		stats.ByType[entry.Type]++
		if entry.Language != "" {
			stats.ByLanguage[entry.Language]++
		}
		if entry.Framework != "" {
			stats.ByFramework[entry.Framework]++
		}
		totalConfidence += entry.Confidence
		stats.TotalUsage += entry.UsageCount
	}

	if stats.TotalEntries > 0 {
		stats.AvgConfidence = totalConfidence / float64(stats.TotalEntries)
	}

	return stats
}

// Stats contiene estadísticas de la Knowledge Base.
type Stats struct {
	TotalEntries    int
	ByType          map[KnowledgeType]int
	ByLanguage      map[string]int
	ByFramework    map[string]int
	AvgConfidence  float64
	TotalUsage     int
}

// load carga la Knowledge Base desde almacenamiento.
func (kb *KnowledgeBase) load() error {
	entries, err := kb.storage.Load()
	if err != nil {
		return err
	}

	kb.mu.Lock()
	defer kb.mu.Unlock()

	for _, entry := range entries {
		kb.entries[entry.ID] = entry
		if kb.config.IndexEnabled {
			kb.index.Add(entry)
		}
	}

	return nil
}

// save guarda la Knowledge Base en almacenamiento.
func (kb *KnowledgeBase) save() error {
	kb.mu.RLock()
	entries := make([]*KnowledgeEntry, 0, len(kb.entries))
	for _, entry := range kb.entries {
		entries = append(entries, entry)
	}
	kb.mu.RUnlock()

	return kb.storage.Save(entries)
}

// ExportJSON exports la Knowledge Base a JSON.
func (kb *KnowledgeBase) ExportJSON() ([]byte, error) {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	entries := make([]*KnowledgeEntry, 0, len(kb.entries))
	for _, entry := range kb.entries {
		entries = append(entries, entry)
	}

	return json.MarshalIndent(entries, "", "  ")
}

// ImportJSON importa una Knowledge Base desde JSON.
func (kb *KnowledgeBase) ImportJSON(data []byte) error {
	var entries []*KnowledgeEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	kb.mu.Lock()
	defer kb.mu.Unlock()

	for _, entry := range entries {
		if _, exists := kb.entries[entry.ID]; exists {
			// Actualizar entrada existente
			kb.entries[entry.ID] = entry
		} else {
			// Añadir nueva entrada
			kb.entries[entry.ID] = entry
		}

		if kb.config.IndexEnabled {
			kb.index.Add(entry)
		}
	}

	return kb.save()
}

// InvertedIndex es un índice invertido para búsqueda rápida.
type InvertedIndex struct {
	mu        sync.RWMutex
	byKeyword map[string]map[string]*KnowledgeEntry // keyword -> entry ID -> entry
	byTitle   map[string]map[string]*KnowledgeEntry
	byContent map[string]map[string]*KnowledgeEntry
}

// NewInvertedIndex crea un nuevo índice invertido.
func NewInvertedIndex() *InvertedIndex {
	return &InvertedIndex{
		byKeyword: make(map[string]map[string]*KnowledgeEntry),
		byTitle:   make(map[string]map[string]*KnowledgeEntry),
		byContent: make(map[string]map[string]*KnowledgeEntry),
	}
}

// Add añade una entrada al índice.
func (idx *InvertedIndex) Add(entry *KnowledgeEntry) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Indexar por keywords
	for _, kw := range entry.Keywords {
		kwLower := strings.ToLower(kw)
		if idx.byKeyword[kwLower] == nil {
			idx.byKeyword[kwLower] = make(map[string]*KnowledgeEntry)
		}
		idx.byKeyword[kwLower][entry.ID] = entry
	}

	// Indexar por palabras del título
	titleWords := strings.Fields(strings.ToLower(entry.Title))
	for _, word := range titleWords {
		if len(word) >= 3 {
			if idx.byTitle[word] == nil {
				idx.byTitle[word] = make(map[string]*KnowledgeEntry)
			}
			idx.byTitle[word][entry.ID] = entry
		}
	}

	// Indexar por palabras del contenido (solo las primeras 1000 palabras)
	contentWords := strings.Fields(strings.ToLower(entry.Content))
	if len(contentWords) > 1000 {
		contentWords = contentWords[:1000]
	}
	for _, word := range contentWords {
		if len(word) >= 3 {
			if idx.byContent[word] == nil {
				idx.byContent[word] = make(map[string]*KnowledgeEntry)
			}
			idx.byContent[word][entry.ID] = entry
		}
	}
}

// Remove elimina una entrada del índice.
func (idx *InvertedIndex) Remove(entry *KnowledgeEntry) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Eliminar de byKeyword
	for _, kw := range entry.Keywords {
		kwLower := strings.ToLower(kw)
		if idx.byKeyword[kwLower] != nil {
			delete(idx.byKeyword[kwLower], entry.ID)
			if len(idx.byKeyword[kwLower]) == 0 {
				delete(idx.byKeyword, kwLower)
			}
		}
	}

	// Eliminar de byTitle
	titleWords := strings.Fields(strings.ToLower(entry.Title))
	for _, word := range titleWords {
		if len(word) >= 3 {
			if idx.byTitle[word] != nil {
				delete(idx.byTitle[word], entry.ID)
				if len(idx.byTitle[word]) == 0 {
					delete(idx.byTitle, word)
				}
			}
		}
	}

	// Eliminar de byContent
	contentWords := strings.Fields(strings.ToLower(entry.Content))
	for _, word := range contentWords {
		if len(word) >= 3 {
			if idx.byContent[word] != nil {
				delete(idx.byContent[word], entry.ID)
				if len(idx.byContent[word]) == 0 {
					delete(idx.byContent, word)
				}
			}
		}
	}
}

// Search busca entradas que contengan la query.
func (idx *InvertedIndex) Search(query string) []*KnowledgeEntry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	results := make(map[string]*KnowledgeEntry)

	queryLower := strings.ToLower(query)
	queryWords := strings.Fields(queryLower)

	// Buscar en keywords
	for _, word := range queryWords {
		if entries, ok := idx.byKeyword[word]; ok {
			for id, entry := range entries {
				results[id] = entry
			}
		}
	}

	// Buscar en título
	for _, word := range queryWords {
		if entries, ok := idx.byTitle[word]; ok {
			for id, entry := range entries {
				results[id] = entry
			}
		}
	}

	// Buscar en contenido
	for _, word := range queryWords {
		if entries, ok := idx.byContent[word]; ok {
			for id, entry := range entries {
				results[id] = entry
			}
		}
	}

	// Convertir a slice
	ret := make([]*KnowledgeEntry, 0, len(results))
	for _, entry := range results {
		ret = append(ret, entry)
	}

	return ret
}

// JSONStorage implementsa almacenamiento en formato JSON.
type JSONStorage struct {
	basePath string
}

// NewJSONStorage crea un nuevo almacenamiento JSON.
func NewJSONStorage(basePath string) (*JSONStorage, error) {
	storage := &JSONStorage{
		basePath: basePath,
	}

	// Crear directorio si no existe
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("error creating directory: %w", err)
	}

	return storage, nil
}

// Load carga las entradas desde JSON.
func (s *JSONStorage) Load() ([]*KnowledgeEntry, error) {
	filePath := filepath.Join(s.basePath, "knowledge.json")

	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		return []*KnowledgeEntry{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var entries []*KnowledgeEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return entries, nil
}

// Save guarda las entradas en JSON.
func (s *JSONStorage) Save(entries []*KnowledgeEntry) error {
	filePath := filepath.Join(s.basePath, "knowledge.json")

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

// StoragePath devuelve la ruta de almacenamiento.
func (s *JSONStorage) StoragePath() string {
	return s.basePath
}