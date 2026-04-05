package memory

import (
	"testing"
)

func TestServeStdio_Initialization(t *testing.T) {
	tmpDir := t.TempDir()

	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("error creating storage: %v", err)
	}

	if storage == nil {
		t.Fatal("expected non-nil storage")
	}

	if storage.memoryDir == "" {
		t.Fatal("expected non-empty memoryDir")
	}
}
