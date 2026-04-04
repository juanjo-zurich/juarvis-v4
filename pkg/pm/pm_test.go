package pm

import (
	"os"
	"path/filepath"
	"testing"
)

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

	return dir
}

func TestLoadMarketplace_Success(t *testing.T) {
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
	dir := t.TempDir()
	t.Setenv("JUARVIS_ROOT", dir)

	_, err := loadMarketplace()
	if err == nil {
		t.Fatal("expected error, got nil")
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

	pluginDir := filepath.Join(dir, "plugins", "plugin-a", ".juarvis-plugin")
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

	pluginDir := filepath.Join(dir, "plugins", "plugin-a", ".juarvis-plugin")
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
	os.MkdirAll(filepath.Join(pluginDir, ".juarvis-plugin"), 0755)
	manifest := `{"name":"plugin-a","version":"1.0.0","description":"Plugin A","category":"dev"}`
	os.WriteFile(filepath.Join(pluginDir, ".juarvis-plugin", "plugin.json"), []byte(manifest), 0644)

	err := RemovePlugin("plugin-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(pluginDir); !os.IsNotExist(err) {
		t.Fatal("plugin directory was not removed")
	}
}
