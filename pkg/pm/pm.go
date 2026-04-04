package pm

import (
	"encoding/json"
	"fmt"
	"juarvis/pkg/output"
	"juarvis/pkg/root"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Marketplace representa la estructura del JSON del catálogo
type Marketplace struct {
	Name    string   `json:"name"`
	Plugins []Plugin `json:"plugins"`
}

type Plugin struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Source      string `json:"source"`
	Category    string `json:"category"`
}

func loadMarketplace() (*Marketplace, error) {
	rootPath, err := root.GetRoot()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo root: %w", err)
	}
	file, err := os.ReadFile(filepath.Join(rootPath, "marketplace.json"))
	if err != nil {
		return nil, fmt.Errorf("no se encontró marketplace.json en %s", rootPath)
	}

	var market Marketplace
	if err := json.Unmarshal(file, &market); err != nil {
		return nil, fmt.Errorf("JSON corrupto: %v", err)
	}
	return &market, nil
}

func ListPlugins() {
	market, err := loadMarketplace()
	if err != nil {
		output.Error("%v", err)
		return
	}

	output.Info("Catálogo: %s", market.Name)
	headers := []string{"NAME", "CATEGORY", "VERSION", "DESCRIPTION"}
	rows := [][]string{}
	for _, p := range market.Plugins {
		rows = append(rows, []string{p.Name, p.Category, p.Version, p.Description})
	}
	output.PrintTable(headers, rows)
}

func findPluginDir(name string) (string, error) {
	rootPath, err := root.GetRoot()
	if err != nil {
		return "", fmt.Errorf("error obteniendo root: %w", err)
	}
	pluginDir := filepath.Join(rootPath, "plugins")

	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return "", err
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(pluginDir, e.Name())
		manifestFile := filepath.Join(path, ".juarvis-plugin", "plugin.json")
		var plug Plugin
		if data, err := os.ReadFile(manifestFile); err == nil {
			json.Unmarshal(data, &plug)
			if plug.Name == name {
				return path, nil
			}
		} else {
			// Fallback si no tiene manifest pero la carpeta se llama igual
			if e.Name() == name || "juarvis-"+e.Name() == name {
				return path, nil
			}
		}
	}
	return "", fmt.Errorf("plugin '%s' no encontrado en el sistema de archivos", name)
}

// SetPluginStatus cambia el estado (habilitado/deshabilitado) creando el fichero 'enabled'
func SetPluginStatus(name string, enabled bool) error {
	pluginDir, err := findPluginDir(name)
	if err != nil {
		return err
	}

	targetFile := filepath.Join(pluginDir, ".juarvis-plugin", "enabled")

	val := "false"
	if enabled {
		val = "true"
	}

	errWrite := os.WriteFile(targetFile, []byte(val), 0644)
	if errWrite != nil {
		return fmt.Errorf("error al escribir estado: %v", errWrite)
	}
	return nil
}

// RemovePlugin borra recursivamente un plugin instalado
func RemovePlugin(name string) error {
	pluginDir, err := findPluginDir(name)
	if err != nil {
		return err
	}

	return os.RemoveAll(pluginDir)
}

// InstallPlugin instala un plugin desde el marketplace
func InstallPlugin(pluginName string) error {
	market, err := loadMarketplace()
	if err != nil {
		return fmt.Errorf("error cargando marketplace: %w", err)
	}

	var targetPlugin *Plugin
	for _, p := range market.Plugins {
		if p.Name == pluginName {
			targetPlugin = &p
			break
		}
	}

	if targetPlugin == nil {
		return fmt.Errorf("plugin '%s' no encontrado en el marketplace", pluginName)
	}

	rootPath, err := root.GetRoot()
	if err != nil {
		return fmt.Errorf("error obteniendo root: %w", err)
	}

	pluginDir := filepath.Join(rootPath, "plugins", targetPlugin.Name)

	// Verificar si ya está instalado
	if _, err := os.Stat(pluginDir); err == nil {
		return fmt.Errorf("plugin '%s' ya instalado. Usa 'juarvis pm remove %s' primero", pluginName, pluginName)
	}

	// Determinar tipo de fuente
	if strings.HasPrefix(targetPlugin.Source, "http") {
		return installFromGit(targetPlugin.Source, pluginDir, targetPlugin.Name)
	}
	return installFromLocal(targetPlugin.Source, pluginDir, rootPath)
}

func installFromGit(url, destDir, pluginName string) error {
	cmd := exec.Command("git", "clone", "--depth", "1", url, destDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error clonando repositorio: %s", string(output))
	}

	// Crear estructura .juarvis-plugin si no existe
	manifestDir := filepath.Join(destDir, ".juarvis-plugin")
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio de manifiesto: %w", err)
	}

	// Crear plugin.json básico si no existe
	pluginJSON := filepath.Join(manifestDir, "plugin.json")
	if _, err := os.Stat(pluginJSON); os.IsNotExist(err) {
		manifest := fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "Plugin instalado desde %s",
  "category": "external"
}`, pluginName, url)
		if err := os.WriteFile(pluginJSON, []byte(manifest), 0644); err != nil {
			return fmt.Errorf("error creando plugin.json: %w", err)
		}
	}

	return nil
}

func installFromLocal(source, destDir, rootPath string) error {
	srcPath := filepath.Join(rootPath, source)
	if _, err := os.Stat(srcPath); err != nil {
		return fmt.Errorf("fuente local no encontrada: %s", srcPath)
	}

	// Copiar directorio recursivamente
	return copyDir(srcPath, destDir)
}

// copyDir copia un directorio recursivamente
func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("error leyendo directorio fuente: %w", err)
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("error creando directorio destino: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("error leyendo %s: %w", srcPath, err)
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return fmt.Errorf("error escribiendo %s: %w", dstPath, err)
			}
		}
	}

	return nil
}
