package sync

import (
	"bytes"
	"fmt"
	"io/fs"
	"juarvis/pkg/assets"
	"juarvis/pkg/output"
	"os"
	"path/filepath"
	"strings"
)

// RunSync actualiza los archivos del ecosistema local con las versiones del binario.
func RunSync(rootPath string) error {
	embeddedFS, err := assets.GetEmbeddedFS()
	if err != nil {
		return fmt.Errorf("error accediendo a assets embebidos: %w", err)
	}

	updatedCount := 0

	// 1. Sincronizar archivos de configuración raíz
	rootFiles := []string{"AGENTS.md", "permissions.yaml", "agent-settings.json", "marketplace.json"}
	for _, f := range rootFiles {
		srcData, err := fs.ReadFile(embeddedFS, f)
		if err != nil {
			continue // Archivo no existe en el binario
		}

		destPath := filepath.Join(rootPath, f)
		destData, readErr := os.ReadFile(destPath)

		if readErr == nil && bytes.Equal(srcData, destData) {
			continue // Ya está actualizado
		}

		if readErr == nil {
			output.Warning("Actualizando %s", f)
		} else {
			output.Info("Creando %s", f)
		}

		if err := os.WriteFile(destPath, srcData, 0644); err != nil {
			return fmt.Errorf("error escribiendo %s: %w", f, err)
		}
		updatedCount++
	}

	// 2. Sincronizar plugins
	entries, err := fs.ReadDir(embeddedFS, "plugins")
	if err != nil {
		return fmt.Errorf("error leyendo plugins embebidos: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pluginName := entry.Name()
		srcPluginPath := "plugins/" + pluginName
		destPluginPath := filepath.Join(rootPath, "plugins", pluginName)

		// Asegurar que el directorio del plugin existe
		os.MkdirAll(destPluginPath, 0755)

		// Recorrer archivos del plugin
		fs.WalkDir(embeddedFS, srcPluginPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				relPath := strings.TrimPrefix(path, srcPluginPath)
				os.MkdirAll(filepath.Join(destPluginPath, relPath), 0755)
				return nil
			}

			relPath := strings.TrimPrefix(path, srcPluginPath)
			destFile := filepath.Join(destPluginPath, relPath)

			// NO sobrescribir el estado del usuario (enabled)
			if relPath == "/.juarvis-plugin/enabled" {
				return nil
			}

			srcData, _ := fs.ReadFile(embeddedFS, path)
			destData, readErr := os.ReadFile(destFile)

			if readErr == nil && bytes.Equal(srcData, destData) {
				return nil // Igual
			}

			if readErr == nil {
				output.Warning("Actualizando plugin %s%s", pluginName, relPath)
			} else {
				output.Info("Añadiendo archivo %s%s", pluginName, relPath)
			}

			if err := os.WriteFile(destFile, srcData, 0644); err != nil {
				return fmt.Errorf("error escribiendo %s: %w", destFile, err)
			}
			updatedCount++
			return nil
		})
	}

	if updatedCount == 0 {
		output.Success("El ecosistema ya está actualizado con la versión del binario.")
	} else {
		output.Success("%d archivos actualizados.", updatedCount)
		output.Info("Ejecuta 'juarvis load' para regenerar el índice de skills.")
	}

	return nil
}
