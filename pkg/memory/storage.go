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
}

func NewStorage(rootPath string) (*Storage, error) {
	memoryDir := filepath.Join(rootPath, config.JuarDir, config.MemoryDir)
	if err := os.MkdirAll(filepath.Join(memoryDir, "observations"), 0755); err != nil {
		return nil, fmt.Errorf("error creando directorio de observaciones: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(memoryDir, "sessions"), 0755); err != nil {
		return nil, fmt.Errorf("error creando directorio de sesiones: %w", err)
	}
	return &Storage{memoryDir: memoryDir}, nil
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

	data, err := json.MarshalIndent(obs, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializando observación: %w", err)
	}

	path := filepath.Join(s.memoryDir, "observations", obs.ID+".json")
	return os.WriteFile(path, data, 0644)
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

	entries, err := os.ReadDir(filepath.Join(s.memoryDir, "observations"))
	if err != nil {
		return nil, fmt.Errorf("error leyendo observaciones: %w", err)
	}

	query = strings.ToLower(query)
	var results []Observation

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
		if project != "" && obs.Project != project {
			continue
		}
		if obsType != "" && obs.Type != obsType {
			continue
		}
		if scope != "" && obs.Scope != scope {
			continue
		}

		content := strings.ToLower(obs.Title + " " + obs.Content)
		if query != "" && !strings.Contains(content, query) {
			continue
		}

		results = append(results, obs)
		if len(results) >= limit {
			break
		}
	}

	return results, nil
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
	obs["updated_at"] = time.Now().Format(time.RFC3339)
	if rev, ok := obs["revision_count"].(float64); ok {
		obs["revision_count"] = rev + 1
	} else {
		obs["revision_count"] = 1
	}

	data, err = json.MarshalIndent(obs, "", "  ")
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
		return os.Remove(path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("observación no encontrada: %s", id)
	}

	var obs map[string]interface{}
	if err := json.Unmarshal(data, &obs); err != nil {
		return fmt.Errorf("error parseando: %w", err)
	}

	obs["deleted_at"] = time.Now().Format(time.RFC3339)
	data, _ = json.MarshalIndent(obs, "", "  ")
	return os.WriteFile(path, data, 0644)
}

func (s *Storage) SaveSession(sess *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(sess, "", "  ")
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
