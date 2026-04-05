package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"juarvis/pkg/config"
	"juarvis/pkg/output"
	"juarvis/pkg/pm"
	"juarvis/pkg/root"
)

// RunLoader simula el plugin-loader.sh: Recrea symlinks, genera registry.
func RunLoader() error {
	rootPath, err := root.GetRoot()
	if err != nil {
		return fmt.Errorf("error obteniendo root: %w", err)
	}
	pluginDir := filepath.Join(rootPath, "plugins")
	skillsDir := filepath.Join(rootPath, "skills")
	juarDir := filepath.Join(rootPath, config.JuarDir)
	registryPath := filepath.Join(juarDir, "skill-registry.md")

	output.Info("Iniciando carga e indexación de Plugins (Juarvis Engine en Go)")

	// Crear directorio temporal en el mismo filesystem para atomicidad
	tmpDir, err := os.MkdirTemp(filepath.Dir(skillsDir), "juarvis-loader-*")
	if err != nil {
		return fmt.Errorf("error creando directorio temporal: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.MkdirAll(juarDir, 0755); err != nil {
		return fmt.Errorf("error creando .juar dir: %w", err)
	}

	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return fmt.Errorf("error leyendo carpeta plugins: %v", err)
	}

	var registryRows []string

	enabledCount := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		pName := e.Name()
		pPath := filepath.Join(pluginDir, pName)

		// Verificar si está deshabilitado
		enabledFile := filepath.Join(pPath, config.JuarvisPluginDir, "enabled")
		if content, err := os.ReadFile(enabledFile); err == nil && strings.TrimSpace(string(content)) == "false" {
			continue // Saltado
		}

		enabledCount++
		// Leer manifiesto
		manifestFile := filepath.Join(pPath, config.JuarvisPluginDir, "plugin.json")
		var plug pm.Plugin
		if data, err := os.ReadFile(manifestFile); err == nil {
			if err := json.Unmarshal(data, &plug); err != nil {
				return fmt.Errorf("error parseando manifest de %s: %w", pName, err)
			}
		} else {
			plug.Name = "juarvis-" + pName
		}

		// Leer skills reales de carpeta
		skillFolders, err := os.ReadDir(filepath.Join(pPath, "skills"))
		if err == nil {
			for _, sk := range skillFolders {
				if sk.IsDir() {
					skName := sk.Name()
					source := filepath.Join("..", "plugins", pName, "skills", skName)
					dest := filepath.Join(tmpDir, skName)
					if err := os.Symlink(source, dest); err != nil {
						return fmt.Errorf("error creando symlink para skill %s: %w", skName, err)
					}
					registryRows = append(registryRows, fmt.Sprintf("| %s | %s | %s | enabled |", skName, pName, filepath.Join("plugins", pName, "skills", skName)))
				}
			}
		}
	}

	// Reemplazo atómico: eliminar skillsDir y renombrar tmpDir
	if err := os.RemoveAll(skillsDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error eliminando skills antiguo: %w", err)
	}
	if err := os.Rename(tmpDir, skillsDir); err != nil {
		return fmt.Errorf("error aplicando cambios atómicos: %w", err)
	}

	// Construir Registry MD
	registryMD := "# Skill Registry\n\n> Generado dinámicamente por Juarvis V4 (Go)\n\n"
	registryMD += "| Skill | Plugin | Source | Status |\n|-------|--------|--------|--------|\n"
	registryMD += strings.Join(registryRows, "\n")

	if err := os.WriteFile(registryPath, []byte(registryMD), 0644); err != nil {
		return fmt.Errorf("error escribiendo skill-registry.md: %w", err)
	}

	output.Success("Cargador finalizado. %d Plugins leídos. %d Skills indexadas y enlazadas.", enabledCount, len(registryRows))
	return nil
}
