package ralph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"juarvis/pkg/config"
)

func TestCreateLoopState(t *testing.T) {
	dir := t.TempDir()
	// Create .juarvis directory
	os.MkdirAll(dir+"/"+config.JuarvisDir, 0755)

	err := CreateLoopState(dir, "test-prompt", 5, "promise")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		s, err := LoadState(dir)
		if err == nil {
			s.Delete()
		}
	}()

	state, err := LoadState(dir)
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}
	if state.Prompt != "test-prompt" {
		t.Errorf("expected prompt 'test-prompt', got %s", state.Prompt)
	}
	if state.MaxIterations != 5 {
		t.Errorf("expected max 5, got %d", state.MaxIterations)
	}
	if state.CompletionPromise != "promise" {
		t.Errorf("expected promise 'promise', got %s", state.CompletionPromise)
	}
	if state.Iteration != 1 {
		t.Errorf("expected iteration 1, got %d", state.Iteration)
	}
}

func TestCheckCompletionPromise_Found(t *testing.T) {
	output := "I'm done with the task <promise>done</promise>"
	found := CheckCompletionPromise(output, "done")
	if !found {
		t.Error("expected to find completion promise")
	}
}

func TestCheckCompletionPromise_NotFound(t *testing.T) {
	output := "still working on it"
	found := CheckCompletionPromise(output, "done")
	if found {
		t.Error("expected not to find completion promise")
	}
}

func TestCheckCompletionPromise_WrongValue(t *testing.T) {
	output := "<promise>something-else</promise>"
	found := CheckCompletionPromise(output, "done")
	if found {
		t.Error("expected not to match wrong promise value")
	}
}

func TestExtractLastAssistantMessage(t *testing.T) {
	tmpDir := t.TempDir()
	transcriptPath := filepath.Join(tmpDir, "transcript.jsonl")
	transcriptContent := `{"message":{"content":[{"type":"text","text":"hello"}]},"role":"user"}
{"message":{"content":[{"type":"text","text":"response"}]},"role":"assistant"}`
	os.WriteFile(transcriptPath, []byte(transcriptContent), 0644)

	msg, err := ExtractLastAssistantMessage(transcriptPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(msg, "response") {
		t.Errorf("expected message to contain 'response', got: %s", msg)
	}
}

func TestExtractLastAssistantMessage_NoAssistant(t *testing.T) {
	tmpDir := t.TempDir()
	transcriptPath := filepath.Join(tmpDir, "transcript.jsonl")
	transcriptContent := `{"message":{"content":[{"type":"text","text":"hello"}]},"role":"user"}`
	os.WriteFile(transcriptPath, []byte(transcriptContent), 0644)

	_, err := ExtractLastAssistantMessage(transcriptPath)
	if err == nil {
		t.Fatal("expected error when no assistant messages")
	}
}

func TestExtractLastAssistantMessage_FileNotFound(t *testing.T) {
	_, err := ExtractLastAssistantMessage("/nonexistent/transcript.jsonl")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestBuildStopResponse_CompletionPromiseMatch(t *testing.T) {
	tmpDir := t.TempDir()
	transcriptPath := filepath.Join(tmpDir, "transcript.jsonl")

	state := &LoopState{
		Active:            true,
		Iteration:         1,
		MaxIterations:     5,
		CompletionPromise: "done",
		Prompt:            "test prompt",
	}

	// Write transcript with matching promise
	transcriptContent := `{"message":{"content":[{"type":"text","text":"<promise>done</promise>"}]},"role":"assistant"}`
	os.WriteFile(transcriptPath, []byte(transcriptContent), 0644)

	resp, err := BuildStopResponse(state, transcriptPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["decision"] != "allow" {
		t.Errorf("expected decision 'allow', got %v", resp["decision"])
	}
}

func TestBuildStopResponse_NoPromise(t *testing.T) {
	tmpDir := t.TempDir()
	transcriptPath := filepath.Join(tmpDir, "transcript.jsonl")
	os.WriteFile(transcriptPath, []byte(`{"message":{"content":[{"type":"text","text":"working"}]},"role":"assistant"}`), 0644)

	state := &LoopState{
		Active:        true,
		Iteration:     1,
		MaxIterations: 5,
		Prompt:        "test prompt",
		RootPath:      tmpDir,
	}

	resp, err := BuildStopResponse(state, transcriptPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["decision"] != "block" {
		t.Errorf("expected decision 'block', got %v", resp["decision"])
	}
}

func TestLoopState_IsComplete(t *testing.T) {
	state := &LoopState{Iteration: 5, MaxIterations: 5}
	if !state.IsComplete() {
		t.Error("expected state to be complete")
	}

	state2 := &LoopState{Iteration: 4, MaxIterations: 5}
	if state2.IsComplete() {
		t.Error("expected state to not be complete")
	}
}

func TestLoopState_Increment(t *testing.T) {
	state := &LoopState{Iteration: 1}
	state.Increment()
	if state.Iteration != 2 {
		t.Errorf("expected iteration 2, got %d", state.Iteration)
	}
}
