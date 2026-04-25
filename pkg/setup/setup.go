package setup

import (
	"encoding/json"
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
		targets = []string{"opencode", "windsurf", "cursor", "vscode", "antigravity", "trae", "kiro", "claude"}
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
		var mcpDest string
		switch t {
		case "opencode":
			targetDir = filepath.Join(homeDir, ".config", "opencode")
		case "windsurf":
			targetDir = filepath.Join(homeDir, ".windsurf", "rules")
			mcpDest = filepath.Join(homeDir, ".codeium", "windsurf", "mcp_config.json")
		case "cursor":
			targetDir = filepath.Join(rootPath, ".cursor", "rules")
			mcpDest = filepath.Join(rootPath, ".cursor", "mcp.json")
		case "vscode":
			targetDir = filepath.Join(rootPath, ".vscode")
		case "antigravity":
			targetDir = filepath.Join(rootPath, ".agent", "rules")
		case "trae":
			targetDir = filepath.Join(homeDir, ".trae", "rules")
		case "kiro":
			targetDir = filepath.Join(homeDir, ".kiro", "rules")
		case "claude":
			targetDir = filepath.Join(rootPath, ".claude")
			mcpDest = filepath.Join(rootPath, ".mcp.json")
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

		// Cargar manifiesto universal para generación dinámica
		var manifest UniversalManifest
		manifestPath := filepath.Join(rootPath, "agent-settings.json")
		if data, err := os.ReadFile(manifestPath); err == nil {
			_ = json.Unmarshal(data, &manifest)
		}

		if t == "opencode" {
			configData, err := manifest.GenerateOpenCodeConfig()
			destOpencode := filepath.Join(targetDir, "opencode.json")
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("no se pudo generar opencode.json: %v", err))
			} else {
				if err := os.WriteFile(destOpencode, configData, 0644); err != nil {
					warnings = append(warnings, fmt.Sprintf("no se pudo escribir opencode.json: %v", err))
				} else {
					output.Success("Configuración opencode.json generada para IDE %s", t)
				}
			}
		}

		if t == "cursor" {
			cursorRules := manifest.GenerateCursorConfig()
			if cursorRules != "" {
				destRules := filepath.Join(targetDir, ".cursorrules")
				if err := os.WriteFile(destRules, []byte(cursorRules), 0644); err != nil {
					warnings = append(warnings, fmt.Sprintf("no se pudo escribir .cursorrules: %v", err))
				} else {
					output.Success("Reglas específicas .cursorrules generadas para Cursor")
				}
			}
		}

		// Distribuir configuración MCP de memoria local
		if mcpDest != "" {
			mcpDir := filepath.Dir(mcpDest)
			if err := os.MkdirAll(mcpDir, 0755); err != nil {
				warnings = append(warnings, fmt.Sprintf("no se pudo crear directorio MCP para %s: %v", t, err))
			} else {
				var srcMCP string
				switch t {
				case "cursor":
					srcMCP = filepath.Join(rootPath, "mcp-cursor.json")
				case "windsurf":
					srcMCP = filepath.Join(rootPath, "mcp-windsurf.json")
				case "claude":
					srcMCP = filepath.Join(rootPath, "mcp-claude.json")
				}
				if srcMCP != "" {
					if _, err := os.Stat(srcMCP); err == nil {
						if err := copyFile(srcMCP, mcpDest); err != nil {
							warnings = append(warnings, fmt.Sprintf("no se pudo copiar MCP config a %s: %v", t, err))
						} else {
							output.Success("Servidor MCP de memoria configurado en IDE %s", t)
						}
					}
				}
			}
		}

		skillsDir := filepath.Join(rootPath, "skills")
		entries, err := os.ReadDir(skillsDir)
		if err == nil {
			for _, e := range entries {
				skillSourceMD := filepath.Join(skillsDir, e.Name(), "SKILL.md")
				if _, err := os.Stat(skillSourceMD); err == nil {
					// Para IDEs tipo OpenCode, copiar a subdirectorio skills/
					skillDestDir := filepath.Join(targetDir, "skills")
					os.MkdirAll(skillDestDir, 0755)
					skillDestMD := filepath.Join(skillDestDir, e.Name(), "SKILL.md")
					os.MkdirAll(filepath.Join(skillDestDir, e.Name()), 0755)
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

		switch t {
		case "vscode", "cursor", "windsurf", "opencode", "antigravity", "trae", "kiro":
			vscodeDir := filepath.Join(rootPath, ".vscode")
			if err := setupWatcherTask(rootPath, vscodeDir, t); err != nil {
				output.Warning("No se pudo añadir watcher task para %s: %v", t, err)
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

	destination, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

type vscodeTask struct {
	Label          string           `json:"label"`
	Type           string           `json:"type"`
	Command        string           `json:"command"`
	RunOptions     vscodeRunOptions `json:"runOptions"`
	IsBackground   bool             `json:"isBackground"`
	ProblemMatcher []string         `json:"problemMatcher"`
}

type vscodeRunOptions struct {
	RunOn string `json:"runOn"`
}

type vscodeTasksFile struct {
	Version string       `json:"version"`
	Tasks   []vscodeTask `json:"tasks"`
}

func setupWatcherTask(rootPath string, targetDir string, ide string) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("no se pudo crear %s: %w", targetDir, err)
	}

	tasksFile := filepath.Join(targetDir, "tasks.json")

	var tasks vscodeTasksFile

	if _, err := os.Stat(tasksFile); err == nil {
		data, readErr := os.ReadFile(tasksFile)
		if readErr != nil {
			return fmt.Errorf("no se pudo leer %s: %w", tasksFile, readErr)
		}
		if err := json.Unmarshal(data, &tasks); err != nil {
			return fmt.Errorf("no se pudo parsear %s: %w", tasksFile, err)
		}
	} else {
		tasks = vscodeTasksFile{
			Version: "2.0.0",
			Tasks:   []vscodeTask{},
		}
	}

	for _, task := range tasks.Tasks {
		if task.Label == "Juarvis Watcher" {
			return nil
		}
	}

	newTask := vscodeTask{
		Label:          "Juarvis Watcher",
		Type:           "shell",
		Command:        "juarvis watch",
		RunOptions:     vscodeRunOptions{RunOn: "folderOpen"},
		IsBackground:   true,
		ProblemMatcher: []string{},
	}

	tasks.Tasks = append(tasks.Tasks, newTask)

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("no se pudo serializar tasks.json: %w", err)
	}

	if err := os.WriteFile(tasksFile, append(data, '\n'), 0644); err != nil {
		return fmt.Errorf("no se pudo escribir %s: %w", tasksFile, err)
	}

	output.Success("Watcher task añadida a %s/tasks.json para %s", targetDir, ide)
	return nil
}
