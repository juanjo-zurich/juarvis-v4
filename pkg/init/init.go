package initpkg

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"juarvis/pkg/assets"
	"juarvis/pkg/loader"
	"juarvis/pkg/output"
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
			if err := copyEmbeddedDir(embeddedFS, srcPath, destPath); err != nil {
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

	// 2. Crear .atl/ y .juarvis-plugin/ para cada plugin
	// Go embed excluye directorios que empiezan con '.', así que los creamos manualmente
	atlDir := filepath.Join(absPath, ".atl")
	if err := os.MkdirAll(atlDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio .atl: %w", err)
	}

	// Leer marketplace.json para saber qué plugins existen
	marketplaceData, err := fs.ReadFile(embeddedFS, "marketplace.json")
	if err == nil {
		var market marketplaceFile
		if err := json.Unmarshal(marketplaceData, &market); err == nil {
			for _, p := range market.Plugins {
				pluginDir := filepath.Join(absPath, "plugins", strings.TrimPrefix(p.Name, "juarvis-"))
				manifestDir := filepath.Join(pluginDir, ".juarvis-plugin")
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

	// 3. Ejecutar loader para indexar los plugins extraídos
	// IMPORTANTE: Setear JUARVIS_ROOT para que el loader use el nuevo ecosistema
	os.Setenv("JUARVIS_ROOT", absPath)
	output.Info("Indexando plugins...")
	if err := loader.RunLoader(); err != nil {
		return fmt.Errorf("error indexando plugins: %w", err)
	}

	output.Success("Ecosistema Juarvis inicializado en %s", absPath)
	output.Info("%d archivos extraídos del binario", copied)
	output.Info("Ejecuta 'juarvis check' para verificar el ecosistema")
	return nil
}

// copyEmbeddedDir copia un directorio completo del embed.FS al filesystem
func copyEmbeddedDir(targetFS fs.FS, srcPath, destPath string) error {
	return fs.WalkDir(targetFS, srcPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath := strings.TrimPrefix(path, srcPath)
		if relPath == "" {
			relPath = "."
		}
		dest := filepath.Join(destPath, strings.TrimPrefix(relPath, string(filepath.Separator)))
		if relPath == "." {
			dest = destPath
		}
		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		content, err := fs.ReadFile(targetFS, path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, content, 0644)
	})
}
