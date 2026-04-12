package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractFrontmatterBlock(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantFM    string
		wantBody  string
		wantFound bool
	}{
		{
			name:      "valid frontmatter",
			content:   "---\nname: test\ndescription: test desc\n---\n# Body",
			wantFM:    "\nname: test\ndescription: test desc",
			wantBody:  "# Body",
			wantFound: true,
		},
		{
			name:      "no frontmatter",
			content:   "# Just a header\nNo frontmatter here",
			wantFM:    "",
			wantBody:  "# Just a header\nNo frontmatter here",
			wantFound: false,
		},
		{
			name:      "empty content",
			content:   "",
			wantFM:    "",
			wantBody:  "",
			wantFound: false,
		},
		{
			name:      "frontmatter with colons in values",
			content:   "---\nname: test:value\ndescription: another: value\n---\nBody content",
			wantFM:    "\nname: test:value\ndescription: another: value",
			wantBody:  "Body content",
			wantFound: true,
		},
		{
			name:      "malformed frontmatter (no closing)",
			content:   "---\nname: test\nNo closing delimiter",
			wantFM:    "",
			wantBody:  "---\nname: test\nNo closing delimiter",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body, found := ExtractFrontmatterBlock(tt.content)
			if fm != tt.wantFM {
				t.Errorf("frontmatter = %q, want %q", fm, tt.wantFM)
			}
			if body != tt.wantBody {
				t.Errorf("body = %q, want %q", body, tt.wantBody)
			}
			if found != tt.wantFound {
				t.Errorf("found = %v, want %v", found, tt.wantFound)
			}
		})
	}
}

func TestCreatePluginManifest(t *testing.T) {
	// Create temp directory
	dir := t.TempDir()
	pluginName := "test-plugin"
	version := "1.0.0"
	description := "Test plugin description"
	category := "utilities"

	// Call the function
	err := CreatePluginManifest(dir, pluginName, version, description, category)
	if err != nil {
		t.Fatalf("CreatePluginManifest() error = %v", err)
	}

	// Verify files were created
	pluginDir := filepath.Join(dir, ".juarvis-plugin")
	manifestPath := filepath.Join(pluginDir, "plugin.json")

	// Check manifest exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Errorf("plugin.json not created at %s", manifestPath)
	}

	// Read and verify content
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to read plugin.json: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, `"name": "test-plugin"`) {
		t.Errorf("manifest missing plugin name, content: %s", content)
	}
	if !strings.Contains(content, `"version": "1.0.0"`) {
		t.Errorf("manifest missing version, content: %s", content)
	}
	if !strings.Contains(content, `"description": "Test plugin description"`) {
		t.Errorf("manifest missing description, content: %s", content)
	}
	if !strings.Contains(content, `"category": "utilities"`) {
		t.Errorf("manifest missing category, content: %s", content)
	}

	// Check enabled file exists
	enabledPath := filepath.Join(pluginDir, "enabled")
	if _, err := os.Stat(enabledPath); os.IsNotExist(err) {
		t.Errorf("enabled file not created at %s", enabledPath)
	}
}

func TestCreatePluginManifest_EmptyCategory(t *testing.T) {
	dir := t.TempDir()
	pluginName := "test-plugin"
	version := "1.0.0"
	description := "Test plugin"
	category := ""

	// Call with empty category - should still work
	err := CreatePluginManifest(dir, pluginName, version, description, category)
	if err != nil {
		t.Fatalf("CreatePluginManifest() error = %v", err)
	}
}
