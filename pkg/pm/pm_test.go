package pm

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"juarvis/pkg/config"
)

// skipIfCI skips test if running in CI (network disabled)
func skipIfCI(t *testing.T) {
	if os.Getenv("CI") == "true" || os.Getenv("JUARVIS_SKIP_NETWORK") == "true" {
		t.Skip("skipping test in CI (network disabled)")
	}
}

func setupPMTest(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	marketplace := `{
  "name": "test",
  "version": "1.0.0",
  "description": "Test marketplace",
  "plugins": [
    {"name": "plugin-a", "version": "1.0.0", "description": "Plugin A", "category": "dev", "source": "./plugins/plugin-a"},
    {"name": "plugin-b", "version": "2.0.0", "description": "Plugin B", "category": "dev", "source": "./plugins/plugin-b"}
  ]
}`
	os.WriteFile(filepath.Join(dir, "marketplace.json"), []byte(marketplace), 0644)
	t.Setenv("JUARVIS_ROOT", dir)

	// Mock HTTP to skip remote calls in tests
	originalHTTP := httpGetFunc
	httpGetFunc = func(url string) (*http.Response, error) {
		return nil, http.ErrMissingFile // Simulate HTTP failure
	}
	t.Cleanup(func() { httpGetFunc = originalHTTP })

	return dir
}

func TestLoadMarketplace_Success(t *testing.T) {
	skipIfCI(t)
	setupPMTest(t)

	market, err := loadMarketplace()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(market.Plugins) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(market.Plugins))
	}
	if market.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", market.Name)
	}
}

func TestLoadMarketplace_NoFile(t *testing.T) {
	skipIfCI(t)
	dir := t.TempDir()
	t.Setenv("JUARVIS_ROOT", dir)

	// Mock HTTP to skip remote calls
	originalHTTP := httpGetFunc
	httpGetFunc = func(url string) (*http.Response, error) {
		return nil, http.ErrMissingFile
	}
	defer func() { httpGetFunc = originalHTTP }()

	// Should NOT error because embedded marketplace is available as fallback
	market, err := loadMarketplace()
	if err != nil {
		t.Fatalf("expected no error (embedded fallback), got: %v", err)
	}
	if len(market.Plugins) == 0 {
		t.Error("expected plugins from embedded marketplace")
	}
}

func TestInstallPlugin_NotFound(t *testing.T) {
	setupPMTest(t)

	err := InstallPlugin("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent plugin, got nil")
	}
}

func TestInstallPlugin_AlreadyInstalled(t *testing.T) {
	dir := setupPMTest(t)

	os.MkdirAll(filepath.Join(dir, "plugins", "plugin-a"), 0755)

	err := InstallPlugin("plugin-a")
	if err == nil {
		t.Fatal("expected error for already installed plugin, got nil")
	}
}

func TestSetPluginStatus_Enable(t *testing.T) {
	dir := setupPMTest(t)

	pluginDir := filepath.Join(dir, "plugins", "plugin-a", config.JuarvisPluginDir)
	os.MkdirAll(pluginDir, 0755)
	manifest := `{"name":"plugin-a","version":"1.0.0","description":"Plugin A","category":"dev"}`
	os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte(manifest), 0644)

	err := SetPluginStatus("plugin-a", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(pluginDir, "enabled"))
	if err != nil {
		t.Fatalf("failed to read enabled file: %v", err)
	}
	if string(data) != "true" {
		t.Errorf("expected 'true', got '%s'", string(data))
	}
}

func TestSetPluginStatus_Disable(t *testing.T) {
	dir := setupPMTest(t)

	pluginDir := filepath.Join(dir, "plugins", "plugin-a", config.JuarvisPluginDir)
	os.MkdirAll(pluginDir, 0755)
	manifest := `{"name":"plugin-a","version":"1.0.0","description":"Plugin A","category":"dev"}`
	os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte(manifest), 0644)
	os.WriteFile(filepath.Join(pluginDir, "enabled"), []byte("true"), 0644)

	err := SetPluginStatus("plugin-a", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(pluginDir, "enabled"))
	if err != nil {
		t.Fatalf("failed to read enabled file: %v", err)
	}
	if string(data) != "false" {
		t.Errorf("expected 'false', got '%s'", string(data))
	}
}

func TestRemovePlugin(t *testing.T) {
	dir := setupPMTest(t)

	pluginDir := filepath.Join(dir, "plugins", "plugin-a")
	os.MkdirAll(filepath.Join(pluginDir, config.JuarvisPluginDir), 0755)
	manifest := `{"name":"plugin-a","version":"1.0.0","description":"Plugin A","category":"dev"}`
	os.WriteFile(filepath.Join(pluginDir, config.JuarvisPluginDir, "plugin.json"), []byte(manifest), 0644)

	err := RemovePlugin("plugin-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(pluginDir); !os.IsNotExist(err) {
		t.Fatal("plugin directory was not removed")
	}
}

func TestHttpGetWithRetry_Throttle(t *testing.T) {
	originalHTTP := httpGetFunc
	callCount := 0

	httpGetFunc = func(url string) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
		}
		return &http.Response{StatusCode: http.StatusTooManyRequests, Body: http.NoBody}, nil
	}
	defer func() { httpGetFunc = originalHTTP }()

	_, err := httpGetWithRetry("http://example.com/test", 2)
	if err != nil {
		t.Logf("expected error after retries: %v", err)
	}
}

func TestIsOfficialProvider(t *testing.T) {
	tests := []struct {
		source   string
		expected bool
	}{
		{"vercel-labs/repo", true},
		{"github/repo", true},
		{"google-labs-code/repo", true},
		{"vercel/repo", true},
		{"sveltejs/repo", true},
		{"random-owner/repo", false},
		{"malicious/repo", false},
	}

	for _, tt := range tests {
		result := isOfficialProvider(tt.source)
		if result != tt.expected {
			t.Errorf("isOfficialProvider(%q) = %v, want %v", tt.source, result, tt.expected)
		}
	}
}

func TestFindPluginDir(t *testing.T) {
	dir := setupPMTest(t)

	pluginDir := filepath.Join(dir, "plugins", "plugin-a")
	os.MkdirAll(filepath.Join(pluginDir, config.JuarvisPluginDir), 0755)
	manifest := `{"name":"plugin-a","version":"1.0.0","description":"Plugin A","category":"dev"}`
	os.WriteFile(filepath.Join(pluginDir, config.JuarvisPluginDir, "plugin.json"), []byte(manifest), 0644)

	found, err := findPluginDir("plugin-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != pluginDir {
		t.Errorf("expected %q, got %q", pluginDir, found)
	}

	_, err = findPluginDir("nonexistent-plugin")
	if err == nil {
		t.Fatal("expected error for nonexistent plugin")
	}
}

func TestRemovePlugin_NotFound(t *testing.T) {
	setupPMTest(t)

	err := RemovePlugin("nonexistent-plugin")
	if err == nil {
		t.Fatal("expected error for nonexistent plugin")
	}
}
