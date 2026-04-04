package initpkg

import (
	"fmt"
	"os"
	"path/filepath"

	"juarvis/pkg/output"
)

// RunInit creates the base structure of a Juarvis ecosystem
func RunInit(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("error resolviendo path %s: %w", path, err)
	}

	// Create base structure
	dirs := []string{
		absPath,
		filepath.Join(absPath, "plugins", "juarvis-core"),
		filepath.Join(absPath, ".atl"),
		filepath.Join(absPath, "plugins", "juarvis-core", ".juarvis-plugin"),
		filepath.Join(absPath, "plugins", "juarvis-core", "skills"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creando directorio %s: %w", dir, err)
		}
	}

	// Create marketplace.json
	marketplacePath := filepath.Join(absPath, "marketplace.json")
	if _, err := os.Stat(marketplacePath); os.IsNotExist(err) {
		marketplace := `{
  "name": "juarvis-ecosystem",
  "version": "1.0.0",
  "description": "Ecosistema Juarvis",
  "plugins": [
    {
      "name": "juarvis-core",
      "version": "1.0.0",
      "description": "Skills basicas del sistema",
      "category": "development",
      "source": "./plugins/juarvis-core"
    }
  ]
}`
		if err := os.WriteFile(marketplacePath, []byte(marketplace), 0644); err != nil {
			return fmt.Errorf("error creando marketplace.json: %w", err)
		}
	}

	// Create plugin.json for core
	pluginJSON := filepath.Join(absPath, "plugins", "juarvis-core", ".juarvis-plugin", "plugin.json")
	if _, err := os.Stat(pluginJSON); os.IsNotExist(err) {
		manifest := `{
  "name": "juarvis-core",
  "version": "1.0.0",
  "description": "Skills basicas del sistema",
  "category": "development"
}`
		if err := os.WriteFile(pluginJSON, []byte(manifest), 0644); err != nil {
			return fmt.Errorf("error creando plugin.json: %w", err)
		}
	}

	// Create enabled file
	enabledPath := filepath.Join(absPath, "plugins", "juarvis-core", ".juarvis-plugin", "enabled")
	if _, err := os.Stat(enabledPath); os.IsNotExist(err) {
		if err := os.WriteFile(enabledPath, []byte("true"), 0644); err != nil {
			return fmt.Errorf("error creando enabled: %w", err)
		}
	}

	// Create empty SKILL.md for core
	skillPath := filepath.Join(absPath, "plugins", "juarvis-core", "skills", "juarvis-core", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(skillPath), 0755); err != nil {
		return fmt.Errorf("error creando directorio de skill: %w", err)
	}
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		skillContent := `# juarvis-core

Skills basicas del sistema Juarvis.

## Uso

Este plugin proporciona las habilidades fundamentales para el funcionamiento del ecosistema.
`
		if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
			return fmt.Errorf("error creando SKILL.md: %w", err)
		}
	}

	output.Success("Ecosistema Juarvis inicializado en %s", absPath)
	output.Info("Ejecuta 'juarvis load' para indexar los plugins")
	return nil
}
