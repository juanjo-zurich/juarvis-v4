package setup

import (
	"fmt"
	"io"
	"io/fs"
	"juarvis/pkg/assets"
	"juarvis/pkg/output"
	"juarvis/pkg/root"
	"os"
	"os/exec"
	"path/filepath"
)

func extractAssetsToRoot(rootPath string) error {
	embeddedFS, err := assets.GetEmbeddedFS()
	if err != nil {
		return err
	}

	return fs.WalkDir(embeddedFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		destPath := filepath.Join(rootPath, path)
		if d.IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("error creando directorio %s: %w", destPath, err)
			}
			return nil
		}

		// Si el fichero ya existe, no lo sobreescribimos (para no pisar modificaciones locales)
		if _, statErr := os.Stat(destPath); statErr == nil {
			return nil
		}

		content, err := fs.ReadFile(embeddedFS, path)
		if err != nil {
			return err
		}

		return os.WriteFile(destPath, content, 0644)
	})
}

// Interfaz que simula el pesado setup.sh copiado multi-IDE
func RunSetup(ide string) error {
	var targets []string
	if ide == "all" {
		targets = []string{"opencode", "windsurf", "cursor", "vscode", "antigravity", "trae", "kiro"}
	} else {
		targets = []string{ide}
	}
	return RunSetupCore(targets)
}

// RunSetupCore efectúa la inyección de reglas y dependencias en los editores seleccionados
func RunSetupCore(targets []string) error {
	rootPath, err := root.GetRoot()
	if err != nil {
		return fmt.Errorf("error obteniendo root: %w", err)
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error obteniendo home directory: %w", err)
	}

	// Extraer activos embebidos al directorio raíz (auto-instalación)
	output.Info("Verificando activos del núcleo de Juarvis en el sistema local...")
	if err := extractAssetsToRoot(rootPath); err != nil {
		output.Warning("No se pudieron extraer los activos base de forma predeterminada: %v", err)
	} else {
		output.Success("Estructura base validada y asegurada.")
	}

	for _, t := range targets {
		var targetDir string
		switch t {
		case "opencode":
			targetDir = filepath.Join(homeDir, ".config", "opencode")
		case "windsurf":
			targetDir = filepath.Join(homeDir, ".windsurf", "rules")
		case "cursor":
			targetDir = filepath.Join(rootPath, ".cursor", "rules")
		case "vscode":
			targetDir = filepath.Join(rootPath, ".vscode")
		case "antigravity":
			targetDir = filepath.Join(rootPath, ".agent", "rules")
		case "trae":
			targetDir = filepath.Join(homeDir, ".trae", "rules")
		case "kiro":
			targetDir = filepath.Join(homeDir, ".kiro", "rules")
		}

		if targetDir == "" {
			continue
		}

		if err := os.MkdirAll(targetDir, 0755); err != nil {
			output.Warning("No se pudo crear directorio %s: %v", targetDir, err)
			continue
		}

		var warnings []string

		srcAgents := filepath.Join(rootPath, "AGENTS.md")
		destAgents := filepath.Join(targetDir, "AGENTS.md")
		if err := copyFile(srcAgents, destAgents); err != nil {
			warnings = append(warnings, fmt.Sprintf("no se pudo copiar AGENTS.md a %s: %v", t, err))
		} else {
			output.Success("Reglas maestras (AGENTS.md) instaladas en IDE %s", t)
		}

		srcPermissions := filepath.Join(rootPath, "permissions.yaml")
		destPermissions := filepath.Join(targetDir, "permissions.yaml")
		if err := copyFile(srcPermissions, destPermissions); err != nil {
			warnings = append(warnings, fmt.Sprintf("no se pudo copiar permissions.yaml a %s: %v", t, err))
		} else {
			output.Success("Reglas de permisos distribuidas en IDE %s", t)
		}

		if t == "opencode" {
			srcOpencode := filepath.Join(rootPath, "opencode.json")
			destOpencode := filepath.Join(targetDir, "opencode.json")
			if err := copyFile(srcOpencode, destOpencode); err != nil {
				warnings = append(warnings, fmt.Sprintf("no se pudo copiar opencode.json a %s: %v", t, err))
			} else {
				output.Success("Configuración opencode.json instalada en IDE %s", t)
			}
		}

		skillsDir := filepath.Join(rootPath, "skills")
		entries, err := os.ReadDir(skillsDir)
		if err == nil {
			for _, e := range entries {
				skillSourceMD := filepath.Join(skillsDir, e.Name(), "SKILL.md")
				if _, err := os.Stat(skillSourceMD); err == nil {
					skillDestMD := filepath.Join(targetDir, e.Name()+".md")
					if err := copyFile(skillSourceMD, skillDestMD); err != nil {
						warnings = append(warnings, fmt.Sprintf("no se pudo copiar skill %s: %v", e.Name(), err))
					} else {
						output.Success("Skill %s instalada en IDE %s", e.Name(), t)
					}
				}
			}
		}

		if len(warnings) > 0 {
			output.Warning("%d advertencias durante la distribucion para %s:", len(warnings), t)
			for _, w := range warnings {
				output.Warning("%s", w)
			}
		}
	}

	output.Success("Distribución finalizada. Tu IA absorberá estas reglas en el próximo chat.")

	// Verificar que el binario juarvis está en el PATH (necesario para hooks)
	if _, err := exec.LookPath("juarvis"); err != nil {
		output.Warning("⚠️  'juarvis' no está en el PATH. Los hooks no funcionarán.")
		output.Info("Ejecuta: make install  o  sudo cp juarvis /usr/local/bin/")
	}

	return nil
}

func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}
