// Package kb implementa la Knowledge Base de Juarvis.
// Retriever implementa la búsqueda y recuperación de conocimiento.
package kb

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// Retriever gestiona la búsqueda y recuperación de conocimiento.
type Retriever struct {
	kb      *KnowledgeBase
	config  *RetrieverConfig
	index   *SearchIndex
	mu      sync.RWMutex
}

// RetrieverConfig contiene la configuración del retriever.
type RetrieverConfig struct {
	MaxResults        int
	MinScore          float64
	EnableFuzzy      bool
	EnableSemantic   bool
	SimilarityThreshold float64
	ContextWindow    int
}

// DefaultRetrieverConfig devuelve la configuración por defecto.
func DefaultRetrieverConfig() *RetrieverConfig {
	return &RetrieverConfig{
		MaxResults:         10,
		MinScore:         0.1,
		EnableFuzzy:      true,
		EnableSemantic:   false, // Semántica básica
		SimilarityThreshold: 0.3,
		ContextWindow:    512,
	}
}

// NewRetriever crea un nuevo Retriever.
func NewRetriever(kb *KnowledgeBase, config *RetrieverConfig) (*Retriever, error) {
	if config == nil {
		config = DefaultRetrieverConfig()
	}

	r := &Retriever{
		kb:     kb,
		config:  config,
		index:  NewSearchIndex(),
	}

	// Inicializar índice
	if err := r.rebuildIndex(); err != nil {
		return nil, fmt.Errorf("error building search index: %w", err)
	}

	return r, nil
}

// SearchIndex es un índice para búsqueda rápida.
type SearchIndex struct {
	mu         sync.RWMutex
	entries    map[string]*KnowledgeEntry
	tfidf      *TFIDFIndex
	termFreq   map[string]map[string]int // term -> entryID -> frequency
}

// TFIDFIndex implementa el índice TF-IDF.
type TFIDFIndex struct {
	mu            sync.RWMutex
	docFreq      map[string]int       // term -> number of documents
	numDocs     int
	termIDF     map[string]float64  // term -> IDF score
}

// NewSearchIndex crea un nuevo índice de búsqueda.
func NewSearchIndex() *SearchIndex {
	return &SearchIndex{
		entries:   make(map[string]*KnowledgeEntry),
		tfidf:     &TFIDFIndex{termIDF: make(map[string]float64)},
		termFreq:  make(map[string]map[string]int),
	}
}

// rebuildIndex reconstruye el índice de búsqueda.
func (r *Retriever) rebuildIndex() error {
	entries := r.kb.GetAll()

	r.index.mu.Lock()
	defer r.index.mu.Unlock()

	r.index.entries = make(map[string]*KnowledgeEntry)
	r.index.termFreq = make(map[string]map[string]int)
	r.index.tfidf.docFreq = make(map[string]int)
	r.index.tfidf.numDocs = len(entries)
	r.index.tfidf.termIDF = make(map[string]float64)

	for _, entry := range entries {
		r.index.entries[entry.ID] = entry

		// Calcular frecuencia de términos
		text := normalizeText(entry.Title + " " + entry.Content)
		terms := strings.Fields(text)

		entryTermFreq := make(map[string]int)
		for _, term := range terms {
			if len(term) < 2 {
				continue
			}
			entryTermFreq[term]++

			// Actualizar document frequency
			if entryTermFreq[term] == 1 {
				r.index.tfidf.docFreq[term]++
			}
		}

		// Guardar frecuencias
		r.index.termFreq[entry.ID] = entryTermFreq
	}

	// Calcular IDF
	docCount := float64(r.index.tfidf.numDocs)
	if docCount > 0 {
		for term, df := range r.index.tfidf.docFreq {
			r.index.tfidf.termIDF[term] = math.Log(docCount / float64(df+1))
		}
	}

	return nil
}

// normalizeText normaliza el texto para búsqueda.
func normalizeText(text string) string {
	text = strings.ToLower(text)
	text = regexp.MustCompile(`[^\w\s]`).ReplaceAllString(text, " ")
	text = strings.Join(strings.Fields(text), " ")
	return text
}

// SearchResult representa un resultado de búsqueda.
type SearchResult struct {
	Entry  *KnowledgeEntry
	Score  float64
	Matches []string
}

// Search busca conocimiento relevancia.
func (r *Retriever) Search(query string, filters *SearchFilters) ([]*SearchResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Parsear query
	terms := parseQuery(query)

	// Obtener candidatos
	candidates := r.getCandidates(terms)

	// Calcular scores
	results := make([]*SearchResult, 0)

	for _, entry := range candidates {
		score := r.calculateScore(entry, terms)
		if score < r.config.MinScore {
			continue
		}

		matches := findMatches(entry, terms, r.config.ContextWindow)

		results = append(results, &SearchResult{
			Entry:   entry,
			Score:  score,
			Matches: matches,
		})
	}

	// Ordenar por score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Aplicar límite
	if r.config.MaxResults > 0 && len(results) > r.config.MaxResults {
		results = results[:r.config.MaxResults]
	}

	// Aplicar filtros adicionales
	if filters != nil {
		results = r.applyResultFilters(results, filters)
	}

	return results, nil
}

// parseQuery parsea una query en términos individuales.
func parseQuery(query string) []string {
	query = normalizeText(query)
	terms := strings.Fields(query)

	// Filtrar stop words
	stopWords := map[string]bool{
		"el": true, "la": true, "los": true, "las": true, "un": true, "una": true,
		"de": true, "del": true, "en": true, "y": true, "o": true, "a": true,
		"que": true, "es": true, "son": true, "para": true, "por": true,
		"como": true, "the": true, "and": true, "or": true, "of": true, "to": true,
	}

	filtered := make([]string, 0)
	for _, term := range terms {
		if !stopWords[term] && len(term) >= 2 {
			filtered = append(filtered, term)
		}
	}

	return filtered
}

// getCandidates obtiene candidatos para los términos dados.
func (r *Retriever) getCandidates(terms []string) []*KnowledgeEntry {
	r.index.mu.RLock()
	defer r.index.mu.RUnlock()

	candidates := make(map[string]*KnowledgeEntry)

	for _, term := range terms {
		// Buscar en término directo
		if termEntries, ok := r.index.termFreq[term]; ok {
			for id := range termEntries {
				if entry, ok := r.index.entries[id]; ok {
					candidates[id] = entry
				}
			}
		}

		// Búsqueda fuzzy si está habilitada
		if r.config.EnableFuzzy {
			fuzzyMatches := r.findFuzzyMatches(term)
			for id, entry := range fuzzyMatches {
				candidates[id] = entry
			}
		}
	}

	// Convertir a slice
	ret := make([]*KnowledgeEntry, 0, len(candidates))
	for _, entry := range candidates {
		ret = append(ret, entry)
	}

	return ret
}

// findFuzzyMatches encuentra coincidencias fuzzy para un término.
func (r *Retriever) findFuzzyMatches(term string) map[string]*KnowledgeEntry {
	r.index.mu.RLock()
	defer r.index.mu.RUnlock()

	matches := make(map[string]*KnowledgeEntry)

	// Búsqueda por prefijo
	for id, entry := range r.index.entries {
		title := strings.ToLower(entry.Title)
		content := strings.ToLower(entry.Content)

		// Comprobar si el término es prefijo de alguna palabra
		if strings.Contains(title, term) || strings.Contains(content, term) {
			matches[id] = entry
			continue
		}

		// Similitud básica con palabras del título
		words := strings.Fields(title)
		for _, word := range words {
			if strings.HasPrefix(word, term) || levenshteinDistance(word, term) <= 2 {
				matches[id] = entry
				break
			}
		}
	}

	return matches
}

// levenshteinDistance calcula la distancia de Levenshtein entre dos strings.
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,
				min(matrix[i][j-1]+1, matrix[i-1][j-1]+cost),
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// calculateScore calcula la relevancia de una entrada para una query.
func (r *Retriever) calculateScore(entry *KnowledgeEntry, terms []string) float64 {
	r.index.mu.RLock()
	defer r.index.mu.RUnlock()

	score := 0.0

	termFreq := r.index.termFreq[entry.ID]
	if termFreq == nil {
		return 0
	}

	for _, term := range terms {
		// TF-IDF score
		tf := float64(termFreq[term])
		idf := r.index.tfidf.termIDF[term]
		tfidfScore := tf * idf
		score += tfidfScore

		// Bonus por coincidencia exacta en título
		if strings.Contains(strings.ToLower(entry.Title), term) {
			score += 0.5
		}

		// Bonus por coincidencia en keywords
		for _, kw := range entry.Keywords {
			if strings.ToLower(kw) == term {
				score += 0.3
			}
		}
	}

	// Normalizar por longitud del contenido
	contentLen := float64(len(entry.Content))
	if contentLen > 0 {
		score = score / math.Sqrt(contentLen)
	}

	// Multiplicar por confianza
	score *= entry.Confidence

	// Bonus por uso frecuente
	usageBonus := float64(entry.UsageCount) * 0.01
	score += usageBonus

	return score
}

// findMatches encuentra las partes del contenido que coinciden.
func findMatches(entry *KnowledgeEntry, terms []string, contextWindow int) []string {
	matches := make([]string, 0)

	content := strings.ToLower(entry.Content)

	for _, term := range terms {
		idx := strings.Index(content, term)
		if idx >= 0 {
			// Extraer contexto alrededor del término
			start := max(0, idx-contextWindow/2)
			end := min(len(content), idx+len(term)+contextWindow/2)
			context := content[start:end]
			context = strings.TrimSpace(context)
			matches = append(matches, context)
		}
	}

	return matches
}

// applyResultFilters aplica filtros adicionales a los resultados.
func (r *Retriever) applyResultFilters(results []*SearchResult, filters *SearchFilters) []*SearchResult {
	filtered := make([]*SearchResult, 0)

	for _, result := range results {
		entry := result.Entry

		// Filtrar por tipo
		if len(filters.Types) > 0 {
			found := false
			for _, t := range filters.Types {
				if entry.Type == t {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filtrar por lenguaje
		if filters.Languages != nil && len(filters.Languages) > 0 {
			if entry.Language == "" {
				continue
			}
			found := false
			for _, lang := range filters.Languages {
				if entry.Language == lang {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filtrar por framework
		if filters.Frameworks != nil && len(filters.Frameworks) > 0 {
			if entry.Framework == "" {
				continue
			}
			found := false
			for _, fw := range filters.Frameworks {
				if entry.Framework == fw {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		filtered = append(filtered, result)
	}

	return filtered
}

// RetrieveByID recupera una entrada por su ID.
func (r *Retriever) RetrieveByID(id string) (*KnowledgeEntry, error) {
	entry, ok := r.kb.Get(id)
	if !ok {
		return nil, fmt.Errorf("entry not found: %s", id)
	}

	return entry, nil
}

// RetrieveByType recupera entradas por tipo.
func (r *Retriever) RetrieveByType(ktype KnowledgeType) []*KnowledgeEntry {
	return r.kb.GetByType(ktype)
}

// RetrieveRelated recupera entradas relacionadas.
func (r *Retriever) RetrieveRelated(entry *KnowledgeEntry, maxResults int) ([]*SearchResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Usar keywords para encontrar relacionados
	terms := make([]string, 0)

	// Añadir keywords del entry
	terms = append(terms, entry.Keywords...)

	// Añadir keywords de tipos similares
	related := r.kb.GetByType(entry.Type)
	for _, re := range related {
		if re.ID != entry.ID {
			terms = append(terms, re.Keywords...)
		}
	}

	// Buscar con los términos combinados
	query := strings.Join(terms, " ")

	results, err := r.Search(query, &SearchFilters{
		Limit: maxResults,
	})

	if err != nil {
		return nil, err
	}

	// Filtrar el propio entry
	filtered := make([]*SearchResult, 0)
	for _, result := range results {
		if result.Entry.ID != entry.ID {
			filtered = append(filtered, result)
		}
	}

	return filtered, nil
}

// GetSuggestions devuelve sugerencias basadas en el contexto actual.
func (r *Retriever) GetSuggestions(context *Context) ([]*KnowledgeEntry, error) {
	kbContext := &Context{
		Task:       context.Task,
		Language:  context.Language,
		Framework: context.Framework,
		Keywords:  context.Keywords,
		Tags:      context.Tags,
	}

	entries := r.kb.GetSuggestions(kbContext)

	// Convertir a sugerencias con scores
	results := make([]*SearchResult, 0)
	for _, entry := range entries {
		results = append(results, &SearchResult{
			Entry: entry,
			Score: entry.Confidence,
		})
	}

	// Ordenar por score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limitar resultados
	if r.config.MaxResults > 0 && len(results) > r.config.MaxResults {
		results = results[:r.config.MaxResults]
	}

	// Convertir de vuelta a entries
	ret := make([]*KnowledgeEntry, len(results))
	for i, result := range results {
		ret[i] = result.Entry
	}

	return ret, nil
}

// SuggestPatternsForTask sugiere patrones para una tarea específica.
func (r *Retriever) SuggestPatternsForTask(task *TaskContext) ([]*SearchResult, error) {
	// Crear query desde el contexto
	query := task.TaskName + " " + task.Description

	// Añadir keywords del contexto
	if len(task.Tags) > 0 {
		query += " " + strings.Join(task.Tags, " ")
	}

	//搜索
	results, err := r.Search(query, &SearchFilters{
		Types: []KnowledgeType{KnowledgeTypeCodePattern, KnowledgeTypeWorkflow},
		Limit: 5,
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetAutoCompleteSuggestions devuelve sugerencias de autocompletado.
func (r *Retriever) GetAutoCompleteSuggestions(partial string, maxResults int) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(partial) < 2 {
		return nil
	}

	partialLower := strings.ToLower(partial)
	suggestions := make(map[string]bool)

	r.index.mu.RLock()
	defer r.index.mu.RUnlock()

	// Buscar en títulos
	for _, entry := range r.index.entries {
		title := strings.ToLower(entry.Title)
		if strings.HasPrefix(title, partialLower) {
			suggestions[entry.Title] = true
		}

		// Buscar en keywords
		for _, kw := range entry.Keywords {
			kwLower := strings.ToLower(kw)
			if strings.HasPrefix(kwLower, partialLower) {
				suggestions[kw] = true
			}
		}
	}

	// Convertir a slice y ordenar
	result := make([]string, 0)
	for s := range suggestions {
		result = append(result, s)
	}

	sort.Strings(result)

	// Limitar
	if maxResults > 0 && len(result) > maxResults {
		result = result[:maxResults]
	}

	return result
}

// SearchByPattern busca entradas que coincidan con un patrón regex.
func (r *Retriever) SearchByPattern(pattern string, filters *SearchFilters) ([]*SearchResult, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	entries := r.kb.GetAll()
	results := make([]*SearchResult, 0)

	for _, entry := range entries {
		// Buscar en título y contenido
		matches := re.FindString(entry.Title)
		if matches == "" {
			matches = re.FindString(entry.Content)
		}

		if matches != "" {
			results = append(results, &SearchResult{
				Entry:  entry,
				Score:  1.0,
				Matches: []string{matches},
			})
		}
	}

	// Aplicar filtros
	if filters != nil {
		results = r.applyResultFilters(results, filters)
	}

	// Ordenar por usage count
	sort.Slice(results, func(i, j int) bool {
		return results[i].Entry.UsageCount > results[j].Entry.UsageCount
	})

	return results, nil
}

// GetRecentEntries devuelve las entradas más recientes.
func (r *Retriever) GetRecentEntries(limit int) []*KnowledgeEntry {
	entries := r.kb.GetAll()

	// Ordenar por fecha de actualización
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].UpdatedAt.After(entries[j].UpdatedAt)
	})

	// Limitar
	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return entries
}

// GetPopularEntries devuelve las entradas más usadas.
func (r *Retriever) GetPopularEntries(limit int) []*KnowledgeEntry {
	entries := r.kb.GetAll()

	// Ordenar por usage count
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].UsageCount > entries[j].UsageCount
	})

	// Limitar
	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return entries
}

// GetEntriesByTag recupera entradas por tag.
func (r *Retriever) GetEntriesByTag(tag string) []*KnowledgeEntry {
	entries := r.kb.GetAll()
	tagLower := strings.ToLower(tag)

	result := make([]*KnowledgeEntry, 0)
	for _, entry := range entries {
		for _, t := range entry.Tags {
			if strings.ToLower(t) == tagLower {
				result = append(result, entry)
				break
			}
		}
	}

	return result
}

// SearchByLanguage busca entradas por lenguaje de programación.
func (r *Retriever) SearchByLanguage(language string) []*KnowledgeEntry {
	entries := r.kb.GetAll()
	result := make([]*KnowledgeEntry, 0)
	for _, entry := range entries {
		if entry.Language == language {
			result = append(result, entry)
		}
	}

	return result
}

// SearchByFramework busca entradas por framework.
func (r *Retriever) SearchByFramework(framework string) []*KnowledgeEntry {
	entries := r.kb.GetAll()
	result := make([]*KnowledgeEntry, 0)
	for _, entry := range entries {
		if entry.Framework == framework {
			result = append(result, entry)
		}
	}

	return result
}

// SearchBySource busca entradas por fuente.
func (r *Retriever) SearchBySource(source string) []*KnowledgeEntry {
	entries := r.kb.GetAll()
	result := make([]*KnowledgeEntry, 0)

	for _, entry := range entries {
		if entry.Source == source {
			result = append(result, entry)
		}
	}

	return result
}

// GetEntriesLearnedFrom recupera entradas aprendidas de una tarea.
func (r *Retriever) GetEntriesLearnedFrom(taskID string) []*KnowledgeEntry {
	entries := r.kb.GetAll()
	result := make([]*KnowledgeEntry, 0)

	for _, entry := range entries {
		if entry.LearnedFrom == taskID {
			result = append(result, entry)
		}
	}

	return result
}

// IncremenUsage incrementa el contador de uso de una entrada.
func (r *Retriever) IncrementUsage(id string) error {
	entry, ok := r.kb.Get(id)
	if !ok {
		return fmt.Errorf("entry not found: %s", id)
	}

	entry.UsageCount++

	return r.kb.Update(entry)
}

// RebuildIndex reconstruye el índice de búsqueda.
func (r *Retriever) RebuildIndex() error {
	return r.rebuildIndex()
}