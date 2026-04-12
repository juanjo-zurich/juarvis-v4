package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"juarvis/pkg/config"
)

type Observation struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	Type          string     `json:"type"`
	Scope         string     `json:"scope"`
	Project       string     `json:"project"`
	TopicKey      string     `json:"topic_key,omitempty"`
	Content       string     `json:"content"`
	SessionID     string     `json:"session_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at,omitempty"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
	RevisionCount int        `json:"revision_count"`
}

type Session struct {
	ID        string     `json:"id"`
	Project   string     `json:"project"`
	Directory string     `json:"directory"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	Summary   string     `json:"summary,omitempty"`
}

type Storage struct {
	mu        sync.RWMutex
	memoryDir string
	index     map[string][]string     // token -> observation IDs
	obsCache  map[string]*Observation // observation cache for fast lookup
}

func NewStorage(rootPath string) (*Storage, error) {
	memoryDir := filepath.Join(rootPath, config.JuarDir, config.MemoryDir)
	if err := os.MkdirAll(filepath.Join(memoryDir, "observations"), 0755); err != nil {
		return nil, fmt.Errorf("error creando directorio de observaciones: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(memoryDir, "sessions"), 0755); err != nil {
		return nil, fmt.Errorf("error creando directorio de sesiones: %w", err)
	}
	s := &Storage{memoryDir: memoryDir, index: make(map[string][]string)}
	s.buildIndex()
	return s, nil
}

func tokenize(text string) []string {
	text = strings.ToLower(text)
	words := strings.Fields(text)
	seen := make(map[string]bool)
	var result []string
	for _, w := range words {
		w = strings.Trim(w, ".,;:!?()[]{}\"'")
		if len(w) >= 2 && !seen[w] {
			seen[w] = true
			result = append(result, w)
		}
	}
	return result
}

func (s *Storage) buildIndex() {
	s.obsCache = make(map[string]*Observation)
	entries, err := os.ReadDir(filepath.Join(s.memoryDir, "observations"))
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.memoryDir, "observations", entry.Name()))
		if err != nil {
			continue
		}
		var obs Observation
		if err := json.Unmarshal(data, &obs); err != nil {
			continue
		}
		if obs.DeletedAt != nil {
			continue
		}
		// Cache the observation for fast lookup
		s.obsCache[obs.ID] = &obs
		tokens := tokenize(obs.Title + " " + obs.Content)
		for _, token := range tokens {
			found := false
			for _, id := range s.index[token] {
				if id == obs.ID {
					found = true
					break
				}
			}
			if !found {
				s.index[token] = append(s.index[token], obs.ID)
			}
		}
	}
}

func (s *Storage) indexObservation(obs *Observation) {
	tokens := tokenize(obs.Title + " " + obs.Content)
	for _, token := range tokens {
		found := false
		for _, id := range s.index[token] {
			if id == obs.ID {
				found = true
				break
			}
		}
		if !found {
			s.index[token] = append(s.index[token], obs.ID)
		}
	}
}

func (s *Storage) SaveObservation(obs *Observation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if obs.ID == "" {
		obs.ID = generateID()
	}
	now := time.Now()
	if obs.CreatedAt.IsZero() {
		obs.CreatedAt = now
	}
	obs.UpdatedAt = now
	obs.RevisionCount++

	data, err := json.Marshal(obs)
	if err != nil {
		return fmt.Errorf("error serializando observación: %w", err)
	}

	path := filepath.Join(s.memoryDir, "observations", obs.ID+".json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error escribiendo observación: %w", err)
	}
	// Update cache
	s.obsCache[obs.ID] = obs
	s.indexObservation(obs)
	return nil
}

func (s *Storage) GetObservation(id string) (*Observation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.memoryDir, "observations", id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("observación no encontrada: %s", id)
	}

	var obs Observation
	if err := json.Unmarshal(data, &obs); err != nil {
		return nil, fmt.Errorf("error parseando observación: %w", err)
	}
	return &obs, nil
}

func (s *Storage) SearchObservations(query, project, obsType, scope string, limit int) ([]Observation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 10
	}

	// Use index to find candidate IDs
	candidateIDs := make(map[string]bool)
	if query != "" {
		queryTokens := tokenize(query)
		for _, token := range queryTokens {
			if ids, ok := s.index[token]; ok {
				for _, id := range ids {
					candidateIDs[id] = true
				}
			}
		}
	} else {
		// No query: use all observations
		entries, _ := os.ReadDir(filepath.Join(s.memoryDir, "observations"))
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".json") {
				candidateIDs[strings.TrimSuffix(entry.Name(), ".json")] = true
			}
		}
	}

	var results []Observation
	for id := range candidateIDs {
		if len(results) >= limit {
			break
		}
		// Use cache instead of reading from disk
		obs, ok := s.obsCache[id]
		if !ok {
			continue
		}
		if obs.DeletedAt != nil {
			continue
		}
		if project != "" && obs.Project != project {
			continue
		}
		if obsType != "" && obs.Type != obsType {
			continue
		}
		if scope != "" && obs.Scope != scope {
			continue
		}
		results = append(results, *obs)
	}

	return results, nil
}

// getObservationLocked is reserved for future use with transactions
// nolint:unused
func (s *Storage) getObservationLocked(id string) (*Observation, error) {
	path := filepath.Join(s.memoryDir, "observations", id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("observación no encontrada: %s", id)
	}
	var obs Observation
	if err := json.Unmarshal(data, &obs); err != nil {
		return nil, fmt.Errorf("error parseando observación: %w", err)
	}
	return &obs, nil
}

func (s *Storage) UpdateObservation(id string, updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.memoryDir, "observations", id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("observación no encontrada: %s", id)
	}

	var obs map[string]interface{}
	if err := json.Unmarshal(data, &obs); err != nil {
		return fmt.Errorf("error parseando observación: %w", err)
	}

	for k, v := range updates {
		if v != nil {
			obs[k] = v
		}
	}
	obs["updated_at"] = time.Now()
	if rev, ok := obs["revision_count"].(float64); ok {
		obs["revision_count"] = rev + 1
	} else {
		obs["revision_count"] = 1
	}

	data, err = json.Marshal(obs)
	if err != nil {
		return fmt.Errorf("error serializando: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

func (s *Storage) DeleteObservation(id string, hard bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.memoryDir, "observations", id+".json")
	if hard {
		_ = os.Remove(path)
		// Remove from cache
		delete(s.obsCache, id)
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("observación no encontrada: %s", id)
	}

	var obs map[string]interface{}
	if err := json.Unmarshal(data, &obs); err != nil {
		return fmt.Errorf("error parseando: %w", err)
	}

	obs["deleted_at"] = time.Now()
	data, err = json.Marshal(obs)
	if err != nil {
		return fmt.Errorf("error serializando: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}
	// Update cache with soft-deleted observation
	if existing, ok := s.obsCache[id]; ok {
		now := time.Now()
		existing.DeletedAt = &now
	}
	return nil
}

func (s *Storage) SaveSession(sess *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	path := filepath.Join(s.memoryDir, "sessions", sess.ID+".json")
	return os.WriteFile(path, data, 0644)
}

func (s *Storage) GetSession(id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.memoryDir, "sessions", id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("sesión no encontrada: %s", id)
	}

	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *Storage) ListSessions(project string, limit int) ([]Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 20
	}

	entries, err := os.ReadDir(filepath.Join(s.memoryDir, "sessions"))
	if err != nil {
		return nil, err
	}

	var results []Session
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(s.memoryDir, "sessions", entry.Name()))
		if err != nil {
			continue
		}

		var sess Session
		if err := json.Unmarshal(data, &sess); err != nil {
			continue
		}

		if project != "" && sess.Project != project {
			continue
		}

		results = append(results, sess)
		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

func generateID() string {
	return fmt.Sprintf("obs_%d", time.Now().UnixNano())
}
