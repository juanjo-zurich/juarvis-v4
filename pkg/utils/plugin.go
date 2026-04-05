package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreatePluginManifest crea el directorio .juarvis-plugin con plugin.json y enabled.
func CreatePluginManifest(pluginDir, name, version, desc, cat string) error {
	manifestDir := filepath.Join(pluginDir, ".juarvis-plugin")
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio de manifiesto: %w", err)
	}

	manifest := fmt.Sprintf(`{
  "name": "%s",
  "version": "%s",
  "description": "%s",
  "category": "%s"
}`, name, version, desc, cat)

	if err := os.WriteFile(filepath.Join(manifestDir, "plugin.json"), []byte(manifest), 0644); err != nil {
		return fmt.Errorf("error escribiendo plugin.json: %w", err)
	}

	if err := os.WriteFile(filepath.Join(manifestDir, "enabled"), []byte("true"), 0644); err != nil {
		return fmt.Errorf("error escribiendo enabled: %w", err)
	}

	return nil
}
