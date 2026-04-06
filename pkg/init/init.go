package initpkg

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"juarvis/pkg/assets"
	"juarvis/pkg/config"
	"juarvis/pkg/loader"
	"juarvis/pkg/output"
	"juarvis/pkg/utils"
)

// marketplaceEntry representa un plugin del marketplace.json embebido
type marketplaceEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Category    string `json:"category"`
}

type marketplaceFile struct {
	Name    string             `json:"name"`
	Plugins []marketplaceEntry `json:"plugins"`
}

// RunInit crea la estructura base de un ecosistema Juarvis desde los assets embebidos
func RunInit(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("error resolviendo path %s: %w", path, err)
	}

	// Verificar si ya existe un ecosistema
	if _, err := os.Stat(filepath.Join(absPath, "marketplace.json")); err == nil {
		return fmt.Errorf("ya existe un ecosistema Juarvis en %s", absPath)
	}

	output.Info("Inicializando ecosistema Juarvis en %s...", absPath)

	// Migrar .atl/ → .juar/ si existe un ecosistema legacy
	migrated, err := migrateAtlToJuar(absPath)
	if err != nil {
		return err
	}
	if migrated {
		output.Success("Migración .atl/ → .juar/ completada")
	}

	// Crear el directorio raíz si no existe
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return fmt.Errorf("error creando directorio %s: %w", absPath, err)
	}

	// 1. Extraer TODOS los assets embebidos al path destino
	embeddedFS, err := assets.GetEmbeddedFS()
	if err != nil {
		return fmt.Errorf("assets embebidos no disponibles: %w", err)
	}

	entries, err := fs.ReadDir(embeddedFS, ".")
	if err != nil {
		return fmt.Errorf("error leyendo assets embebidos: %w", err)
	}

	copied := 0
	for _, entry := range entries {
		srcPath := entry.Name()
		destPath := filepath.Join(absPath, srcPath)

		// No sobrescribir archivos existentes
		if _, err := os.Stat(destPath); err == nil {
			output.Warning("%s ya existe, omitiendo", srcPath)
			continue
		}

		if entry.IsDir() {
			if err := utils.CopyEmbeddedDir(embeddedFS, srcPath, destPath); err != nil {
				return fmt.Errorf("error extrayendo %s: %w", srcPath, err)
			}
		} else {
			content, err := fs.ReadFile(embeddedFS, srcPath)
			if err != nil {
				return fmt.Errorf("error leyendo %s del embed: %w", srcPath, err)
			}
			if err := os.WriteFile(destPath, content, 0644); err != nil {
				return fmt.Errorf("error escribiendo %s: %w", destPath, err)
			}
		}
		copied++
	}

	// 2. Crear .juar/ y .juarvis-plugin/ para cada plugin
	// Go embed excluye directorios que empiezan con '.', así que los creamos manualmente
	juarDir := filepath.Join(absPath, config.JuarDir)
	if err := os.MkdirAll(juarDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio .juar: %w", err)
	}

	// Leer marketplace.json para saber qué plugins existen
	marketplaceData, err := fs.ReadFile(embeddedFS, "marketplace.json")
	if err == nil {
		var market marketplaceFile
		if err := json.Unmarshal(marketplaceData, &market); err == nil {
			for _, p := range market.Plugins {
				pluginDir := filepath.Join(absPath, "plugins", strings.TrimPrefix(p.Name, "juarvis-"))
				manifestDir := filepath.Join(pluginDir, config.JuarvisPluginDir)
				if err := os.MkdirAll(manifestDir, 0755); err != nil {
					return fmt.Errorf("error creando manifest dir para %s: %w", p.Name, err)
				}

				// Crear plugin.json
				manifest := fmt.Sprintf(`{
  "name": "%s",
  "version": "%s",
  "description": "%s",
  "category": "%s"
}`, p.Name, p.Version, p.Description, p.Category)
				if err := os.WriteFile(filepath.Join(manifestDir, "plugin.json"), []byte(manifest), 0644); err != nil {
					return fmt.Errorf("error creando plugin.json para %s: %w", p.Name, err)
				}

				// Crear enabled
				if err := os.WriteFile(filepath.Join(manifestDir, "enabled"), []byte("true"), 0644); err != nil {
					return fmt.Errorf("error creando enabled para %s: %w", p.Name, err)
				}
			}
		}
	}

	// 3. Instalar pre-commit hook si existe .git/
	gitDir := filepath.Join(absPath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		hookSrc := filepath.Join(absPath, "hooks", "pre-commit")
		hookDest := filepath.Join(gitDir, "hooks", "pre-commit")
		if _, err := os.Stat(hookSrc); err == nil {
			content, readErr := os.ReadFile(hookSrc)
			if readErr == nil {
				if writeErr := os.WriteFile(hookDest, content, 0755); writeErr == nil {
					output.Success("Pre-commit hook instalado")
				}
			}
		}
	}

	// 4. Ejecutar loader para indexar los plugins extraídos
	output.Info("Indexando plugins...")
	if err := loader.RunLoader(absPath); err != nil {
		return fmt.Errorf("error indexando plugins: %w", err)
	}

	output.Success("Ecosistema Juarvis inicializado en %s", absPath)
	output.Info("%d archivos extraídos del binario", copied)
	output.Info("Ejecuta 'juarvis check' para verificar el ecosistema")
	return nil
}

// pathExists verifica si un path existe en el filesystem
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// migrateAtlToJuar renombra .atl/ a .juar/ para ecosistemas legacy.
// Retorna true si se realizó la migración, false si no era necesaria.
func migrateAtlToJuar(rootPath string) (bool, error) {
	juarPath := filepath.Join(rootPath, config.JuarDir)
	atlPath := filepath.Join(rootPath, ".atl")

	if pathExists(juarPath) {
		return false, nil // .juar ya existe
	}
	if !pathExists(atlPath) {
		return false, nil // nada que migrar
	}

	output.Warning("Directorio .atl/ renombrado a .juar/ para compatibilidad")
	if err := os.Rename(atlPath, juarPath); err != nil {
		return false, fmt.Errorf("error renombrando .atl/ a .juar/: %w", err)
	}
	return true, nil
}
