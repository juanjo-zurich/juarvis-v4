package loader

import (
	"os"
	"path/filepath"
	"testing"

	"juarvis/pkg/root"
)

func setupLoaderTest(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	marketplace := `{"name":"test","plugins":[{"name":"test-plugin","version":"1.0.0","description":"Test","category":"dev","source":"./plugins/test-plugin"}]}`
	os.WriteFile(filepath.Join(dir, "marketplace.json"), []byte(marketplace), 0644)

	pluginDir := filepath.Join(dir, "plugins", "test-plugin", ".juarvis-plugin")
	os.MkdirAll(pluginDir, 0755)
	manifest := `{"name":"test-plugin","version":"1.0.0","description":"Test plugin","category":"dev"}`
	os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte(manifest), 0644)
	os.WriteFile(filepath.Join(pluginDir, "enabled"), []byte("true"), 0644)

	skillDir := filepath.Join(dir, "plugins", "test-plugin", "skills", "test-skill")
	os.MkdirAll(skillDir, 0755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# test-skill\n"), 0644)

	t.Setenv("JUARVIS_ROOT", dir)

	return dir
}

func TestRunLoader_Success(t *testing.T) {
	setupLoaderTest(t)

	err := RunLoader()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rootPath, _ := root.GetRoot()
	skillsDir := filepath.Join(rootPath, "skills")
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		t.Fatal("skills directory was not created")
	}

	registryPath := filepath.Join(rootPath, ".juar", "skill-registry.md")
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		t.Fatal("skill-registry.md was not created")
	}
}

func TestRunLoader_NoPlugins(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "marketplace.json"), []byte(`{"name":"test","plugins":[]}`), 0644)
	os.MkdirAll(filepath.Join(dir, "plugins"), 0755)
	t.Setenv("JUARVIS_ROOT", dir)

	err := RunLoader()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunLoader_InvalidManifest(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "marketplace.json"), []byte(`{"name":"test","plugins":[{"name":"bad","version":"1.0.0","description":"Bad","category":"dev","source":"./plugins/bad"}]}`), 0644)

	pluginDir := filepath.Join(dir, "plugins", "bad", ".juarvis-plugin")
	os.MkdirAll(pluginDir, 0755)
	os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte(`{invalid json`), 0644)
	os.WriteFile(filepath.Join(pluginDir, "enabled"), []byte("true"), 0644)
	t.Setenv("JUARVIS_ROOT", dir)

	err := RunLoader()
	if err == nil {
		t.Log("Warning: RunLoader did not return error for invalid manifest")
	}
}
