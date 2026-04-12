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
func RunLoader(rootPath string) error {
	if rootPath == "" {
		var err error
		rootPath, err = root.GetRoot()
		if err != nil {
			return fmt.Errorf("error obteniendo root: %w", err)
		}
	}
	pluginDir := filepath.Join(rootPath, "plugins")
	skillsDir := filepath.Join(rootPath, "skills")
	juarDir := filepath.Join(rootPath, config.JuarDir)
	registryPath := filepath.Join(juarDir, config.SkillRegistryFile)

	output.Info("Iniciando carga e indexación de Plugins (Juarvis Engine en Go)")

	// Leer plugins primero (necesario para validación y para el loader)
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return fmt.Errorf("error leyendo carpeta plugins: %w", err)
	}

	// Verificación incremental: si skillsDir y registry existen y todos los symlinks son válidos, skip
	if valid, _ := areSkillsValid(rootPath, skillsDir, registryPath, entries); valid {
		output.Info("Skills ya actualizadas. No se requiere recarga.")
		return nil
	}

	// Crear directorio temporal en el mismo filesystem para atomicidad
	tmpDir, err := os.MkdirTemp(filepath.Dir(skillsDir), "juarvis-loader-*")
	if err != nil {
		return fmt.Errorf("error creando directorio temporal: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if err := os.MkdirAll(juarDir, 0755); err != nil {
		return fmt.Errorf("error creando .juar dir: %w", err)
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

					// Security: validate symlink target stays within ecosystem
					cleanSource := filepath.Clean(source)
					absSource, err := filepath.Abs(filepath.Join(tmpDir, cleanSource))
					if err != nil {
						output.Warning("Skill %s saltada: no se pudo resolver path", skName)
						continue
					}
					absRoot, _ := filepath.Abs(rootPath)
					if !strings.HasPrefix(absSource, absRoot) {
						output.Warning("Skill %s saltada: symlink apunta fuera del ecosistema", skName)
						continue
					}

					if err := os.Symlink(cleanSource, dest); err != nil {
						// Skill duplicada — otro plugin ya la tiene. Skip sin error.
						output.Warning("Skill '%s' duplicada en plugin '%s' (ya existe en otro plugin). Se omite.", skName, pName)
						continue
					}
					registryRows = append(registryRows, fmt.Sprintf("| %s | %s | %s | enabled |", skName, pName, filepath.Join("plugins", pName, "skills", skName)))
				}
			}
		}
	}

	// ==== User Skills (.agent/skills/) ====
	userSkillsDir := filepath.Join(rootPath, ".agent", "skills")
	if userFiles, err := os.ReadDir(userSkillsDir); err == nil {
		output.Info("Indexando skills de usuario...")
		for _, userFile := range userFiles {
			if !userFile.IsDir() {
				continue
			}
			skillName := userFile.Name()
			skillDir := filepath.Join(userSkillsDir, skillName)

			// Buscar SKILL.md
			skillFile := filepath.Join(skillDir, "SKILL.md")
			if _, err := os.Stat(skillFile); os.IsNotExist(err) {
				continue
			}

			// Crear symlink en skills/
			source := filepath.Join("..", ".agent", "skills", skillName)
			dest := filepath.Join(tmpDir, skillName)

			// Security: validate symlink target stays within ecosystem
			cleanSource := filepath.Clean(source)
			absSource, err := filepath.Abs(filepath.Join(rootPath, cleanSource))
			if err != nil {
				output.Warning("Skill de usuario %s saltada: no se pudo resolver path", skillName)
				continue
			}
			absRoot, _ := filepath.Abs(rootPath)
			if !strings.HasPrefix(absSource, absRoot) {
				output.Warning("Skill de usuario %s saltada: symlink apunta fuera del ecosistema", skillName)
				continue
			}

			if err := os.Symlink(cleanSource, dest); err != nil {
				// Skill duplicada — ya existe. Skip sin error.
				output.Warning("Skill de usuario '%s' duplicada. Se omite.", skillName)
				continue
			}
			registryRows = append(registryRows, fmt.Sprintf("| %s | user | .agent/skills/%s | enabled |", skillName, skillName))
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

// areSkillsValid verifica si los symlinks existentes son válidos y el registry existe.
func areSkillsValid(rootPath, skillsDir, registryPath string, pluginEntries []os.DirEntry) (bool, error) {
	// Verificar que el registry existe
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		return false, nil
	}

	// Verificar que skillsDir existe
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		return false, nil
	}

	// Verificar que cada plugin habilitado tiene sus symlinks válidos
	for _, e := range pluginEntries {
		if !e.IsDir() {
			continue
		}
		pName := e.Name()
		pPath := filepath.Join(rootPath, "plugins", pName)
		enabledFile := filepath.Join(pPath, config.JuarvisPluginDir, "enabled")
		if content, err := os.ReadFile(enabledFile); err == nil && strings.TrimSpace(string(content)) == "false" {
			continue
		}
		skillPath := filepath.Join(pPath, "skills")
		skillFolders, err := os.ReadDir(skillPath)
		if err != nil {
			continue
		}
		for _, sk := range skillFolders {
			if !sk.IsDir() {
				continue
			}
			linkPath := filepath.Join(skillsDir, sk.Name())
			target, err := os.Readlink(linkPath)
			if err != nil {
				return false, nil
			}
			// Verificar que el target resuelve
			absTarget, err := filepath.Abs(filepath.Join(skillsDir, target))
			if err != nil {
				return false, nil
			}
			if _, err := os.Stat(absTarget); err != nil {
				return false, nil
			}
		}
	}
	return true, nil
}
