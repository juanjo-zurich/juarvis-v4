package memory

import (
	"os"
	"path/filepath"
	"testing"

	"juarvis/pkg/config"
)

func setupStorage(t *testing.T) *Storage {
	t.Helper()
	tmpDir := t.TempDir()
	s, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("error creating storage: %v", err)
	}
	return s
}

func TestSaveAndGetObservation(t *testing.T) {
	s := setupStorage(t)

	obs := &Observation{
		Title:   "Test observation",
		Type:    "test",
		Project: "test-project",
		Content: "test content",
	}

	if err := s.SaveObservation(obs); err != nil {
		t.Fatalf("error saving: %v", err)
	}

	if obs.ID == "" {
		t.Fatal("expected ID to be generated")
	}

	got, err := s.GetObservation(obs.ID)
	if err != nil {
		t.Fatalf("error getting: %v", err)
	}

	if got.Title != obs.Title {
		t.Errorf("expected title %q, got %q", obs.Title, got.Title)
	}
}

func TestSearchObservations(t *testing.T) {
	s := setupStorage(t)

	s.SaveObservation(&Observation{Title: "Fix bug in auth", Type: "bugfix", Project: "proj-a", Content: "authentication fix"})
	s.SaveObservation(&Observation{Title: "Add new feature", Type: "feature", Project: "proj-a", Content: "new dashboard"})
	s.SaveObservation(&Observation{Title: "Fix bug in db", Type: "bugfix", Project: "proj-b", Content: "database fix"})

	results, err := s.SearchObservations("bug", "proj-a", "", "", 10)
	if err != nil {
		t.Fatalf("error searching: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Title != "Fix bug in auth" {
		t.Errorf("expected 'Fix bug in auth', got %q", results[0].Title)
	}
}

func TestDeleteObservation(t *testing.T) {
	s := setupStorage(t)

	obs := &Observation{Title: "To delete", Type: "test", Content: "content"}
	s.SaveObservation(obs)

	if err := s.DeleteObservation(obs.ID, false); err != nil {
		t.Fatalf("error soft deleting: %v", err)
	}

	results, err := s.SearchObservations("", "", "", "", 10)
	if err != nil {
		t.Fatalf("error searching: %v", err)
	}

	for _, r := range results {
		if r.ID == obs.ID {
			t.Error("deleted observation should not appear in search")
		}
	}

	if err := s.DeleteObservation(obs.ID, true); err != nil {
		t.Fatalf("error hard deleting: %v", err)
	}

	_, err = s.GetObservation(obs.ID)
	if err == nil {
		t.Fatal("expected error getting hard-deleted observation")
	}
}

func TestSessionLifecycle(t *testing.T) {
	s := setupStorage(t)

	sess := &Session{
		ID:        "session-1",
		Project:   "test-project",
		Directory: "/tmp/test",
	}

	if err := s.SaveSession(sess); err != nil {
		t.Fatalf("error saving session: %v", err)
	}

	sessions, err := s.ListSessions("test-project", 10)
	if err != nil {
		t.Fatalf("error listing sessions: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
}

func TestStorageCreatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("error creating storage: %v", err)
	}

	obsDir := filepath.Join(tmpDir, config.JuarDir, config.MemoryDir, "observations")
	sessDir := filepath.Join(tmpDir, config.JuarDir, config.MemoryDir, "sessions")

	if _, err := os.Stat(obsDir); os.IsNotExist(err) {
		t.Error("observations directory was not created")
	}
	if _, err := os.Stat(sessDir); os.IsNotExist(err) {
		t.Error("sessions directory was not created")
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"hello world", []string{"hello", "world"}},
		{"Hello, World!", []string{"hello", "world"}},
		{"test  test", []string{"test"}},
		{"a", []string{}},                 // menos de 2 chars
		{"a bc de", []string{"bc", "de"}}, // solo >1 char
	}

	for _, tt := range tests {
		result := tokenize(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("tokenize(%q) = %v, want %v", tt.input, result, tt.expected)
			continue
		}
		for i, v := range result {
			if v != tt.expected[i] {
				t.Errorf("tokenize(%q)[%d] = %v, want %v", tt.input, i, v, tt.expected[i])
			}
		}
	}
}

func TestUpdateObservation(t *testing.T) {
	s := setupStorage(t)

	obs := &Observation{
		Title:   "Original title",
		Type:    "test",
		Project: "test-project",
		Content: "original content",
	}
	if err := s.SaveObservation(obs); err != nil {
		t.Fatalf("error saving: %v", err)
	}

	updates := map[string]interface{}{
		"title": "Updated title",
	}
	if err := s.UpdateObservation(obs.ID, updates); err != nil {
		t.Fatalf("error updating: %v", err)
	}

	updated, err := s.GetObservation(obs.ID)
	if err != nil {
		t.Fatalf("error getting: %v", err)
	}
	if updated.Title != "Updated title" {
		t.Errorf("expected title 'Updated title', got %q", updated.Title)
	}
}

func TestSearchObservations_Filters(t *testing.T) {
	s := setupStorage(t)

	s.SaveObservation(&Observation{Title: "Bug fix A", Type: "bugfix", Project: "proj-a", Content: "fix auth"})
	s.SaveObservation(&Observation{Title: "Feature B", Type: "feature", Project: "proj-a", Content: "new dashboard"})
	s.SaveObservation(&Observation{Title: "Bug fix C", Type: "bugfix", Project: "proj-b", Content: "fix db"})

	results, err := s.SearchObservations("", "proj-a", "", "", 10)
	if err != nil {
		t.Fatalf("error searching: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results for proj-a, got %d", len(results))
	}

	results, err = s.SearchObservations("", "", "bugfix", "", 10)
	if err != nil {
		t.Fatalf("error searching: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 bugfix results, got %d", len(results))
	}

	results, err = s.SearchObservations("", "", "", "", 1)
	if err != nil {
		t.Fatalf("error searching: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result with limit, got %d", len(results))
	}
}

func TestGetObservation_NotFound(t *testing.T) {
	s := setupStorage(t)

	_, err := s.GetObservation("nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent observation")
	}
}
