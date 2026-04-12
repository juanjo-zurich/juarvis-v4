package verify

import (
	"testing"
)

func TestCheckGoBuild(t *testing.T) {
	result := checkGoBuild()
	// Just verify the result has the expected fields
	if result.Name != "go build" {
		t.Errorf("Name = %q, want %q", result.Name, "go build")
	}
	// Message should not be empty
	if result.Message == "" {
		t.Error("Message should not be empty")
	}
	// Passed should be true if Go is installed
	if result.Passed && result.Message != "Compilación exitosa" {
		t.Errorf("When Passed is true, Message should be 'Compilación exitosa', got %q", result.Message)
	}
}

func TestCheckGoVet(t *testing.T) {
	result := checkGoVet()
	if result.Name != "go vet" {
		t.Errorf("Name = %q, want %q", result.Name, "go vet")
	}
	if result.Message == "" {
		t.Error("Message should not be empty")
	}
}

func TestCheckGoTest(t *testing.T) {
	result := checkGoTest()
	if result.Name != "go test" {
		t.Errorf("Name = %q, want %q", result.Name, "go test")
	}
	if result.Message == "" {
		t.Error("Message should not be empty")
	}
}

func TestCheckEmbeddedJSONs(t *testing.T) {
	result := checkEmbeddedJSONs()
	if result.Name != "embedded JSON" {
		t.Errorf("Name = %q, want %q", result.Name, "embedded JSON")
	}
	// If embedded assets work, should pass
	if !result.Passed && result.Message != "" {
		// Failed for some reason, check message
		t.Logf("Embedded JSON check failed: %s", result.Message)
	}
}

func TestCheckPluginManifests(t *testing.T) {
	result := checkPluginManifests()
	if result.Name != "plugin manifests" {
		t.Errorf("Name = %q, want %q", result.Name, "plugin manifests")
	}
	// Should typically pass as manifests are optional
	if !result.Passed {
		t.Logf("Plugin manifests check: %s", result.Message)
	}
}

func TestRunVerify(t *testing.T) {
	results, err := RunVerify()
	if err != nil {
		t.Fatalf("RunVerify() error = %v", err)
	}
	// Should have returned results for all checks
	if len(results) == 0 {
		t.Error("RunVerify() should return at least one result")
	}
	// Verify all results have required fields
	for i, r := range results {
		if r.Name == "" {
			t.Errorf("Result[%d].Name is empty", i)
		}
		if r.Message == "" {
			t.Errorf("Result[%d].Message is empty", i)
		}
	}
}
