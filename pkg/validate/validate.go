package validate

import (
	"bytes"
	"fmt"
	"io/fs"
	"juarvis/pkg/assets"
	"juarvis/pkg/config"
	"juarvis/pkg/output"
	"juarvis/pkg/root"
	"os"
	"os/exec"
	"path/filepath"
)

// RunHealthCheck revisa la estructura, como lo hacía `juarvis-validate check`
func RunHealthCheck() error {
	output.Info("Verificando estado del ecosistema Juarvis...")
	output.Info("----------------------------------------")

	errors := 0
	rootPath, err := root.GetRoot()
	if err != nil {
		return fmt.Errorf("error obteniendo root: %w", err)
	}

	// Validamos Git
	cmd := exec.Command("git", "--version")
	if err := cmd.Run(); err != nil {
		output.Error("[CRÍTICO] git no está instalado o no es accesible en el PATH.")
		errors++
	} else {
		output.Success("Git detectado.")
	}

	// Comprobar marketplace
	if _, err := os.Stat(filepath.Join(rootPath, "marketplace.json")); os.IsNotExist(err) {
		embeddedFS, embErr := assets.GetEmbeddedFS()
		if embErr == nil {
			if _, err := fs.ReadFile(embeddedFS, "marketplace.json"); err == nil {
				output.Warning("marketplace.json no encontrado en filesystem (disponible en binario). Ejecuta 'juarvis init'.")
			} else {
				output.Error("[CRÍTICO] marketplace.json no encontrado en ningún sitio")
				errors++
			}
		} else {
			output.Error("[CRÍTICO] marketplace.json no encontrado en filesystem")
			errors++
		}
	} else {
		output.Success("Catálogo Marketplace enlazado.")
	}

	// Comprobar Skill Registry y directorio activo
	registryPath := filepath.Join(rootPath, config.JuarDir, config.SkillRegistryFile)
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		output.Warning("No hay skill-registry.md generado. Ejecuta 'juarvis load'.")
	} else {
		// Verificar que el registry tiene contenido real
		content, readErr := os.ReadFile(registryPath)
		if readErr != nil {
			output.Warning("No se pudo leer skill-registry.md. Ejecuta 'juarvis load'.")
		} else if len(content) < 50 {
			output.Warning("skill-registry.md existe pero está vacío o corrupto. Ejecuta 'juarvis load'.")
		} else if !bytes.Contains(content, []byte("|")) {
			output.Warning("skill-registry.md existe pero no tiene entradas de skills. Ejecuta 'juarvis load'.")
		} else {
			output.Success("Base de memoria LLM (.juar) intacta.")
		}
	}

	if errors > 0 {
		return fmt.Errorf("el sistema falló la comprobación de salud (%d errores críticos encontrados)", errors)
	}

	output.Success("Verificación completada sin errores críticos.")
	return nil
}
