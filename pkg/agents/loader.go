package agents

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"juarvis/pkg/config"
	"juarvis/pkg/root"
)

// ErrNoAgentsFile cuando no se encuentra AGENTS.md
var ErrNoAgentsFile = errors.New("no se encontró AGENTS.md")

// ErrAgentNotFound cuando un agente específico no existe
var ErrAgentNotFound = errors.New("agente no encontrado")

// ErrInvalidFormat cuando AGENTS.md no tiene formato válido
var ErrInvalidFormat = errors.New("formato de AGENTS.md inválido (no es YAML-frontmatter)")

// Loader maneja la carga de configuración de agentes
type Loader struct {
	rootPath string
}

// NewLoader crea un nuevo Loader
func NewLoader() *Loader {
	return &Loader{}
}

// LoadFromProject carga AGENTS.md desde la raíz del proyecto
func (l *Loader) LoadFromProject() (*AgentsConfig, error) {
	rootPath, err := root.GetRoot()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo raíz del proyecto: %w", err)
	}

	l.rootPath = rootPath
	return l.LoadFromPath(filepath.Join(rootPath, "AGENTS.md"))
}

// LoadFromPath carga AGENTS.md desde un path específico
func (l *Loader) LoadFromPath(path string) (*AgentsConfig, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoAgentsFile
		}
		return nil, fmt.Errorf("error accediendo a %s: %w", path, err)
	}

	config, err := ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("error parseando AGENTS.md: %w", err)
	}

	l.rootPath = filepath.Dir(path)
	return config, nil
}

// LoadWithPrecedence aplica la precedencia: AGENTS.md → juarvis.yaml → flags → defaults
func (l *Loader) LoadWithPrecedence(opts *LoadOptions) (*AgentsConfig, error) {
	// 1. Intentar cargar desde AGENTS.md
	agentsConfig, err := l.TryLoadAgents()
	if err == nil && agentsConfig != nil {
		// Aplicar overrides desde CLI/flags si hay opciones
		if opts != nil && len(opts.Overrides) > 0 {
			agentsConfig = applyOverrides(agentsConfig, opts.Overrides)
		}
		return agentsConfig, nil
	}

	// 2. Si el archivo existe pero no es formato válido, intentar juarvis.yaml
	if err == ErrNoAgentsFile || err == ErrInvalidFormat {
		agentsConfig, loadErr := l.loadFromJuarvisYaml()
		if loadErr == nil {
			return agentsConfig, nil
		}
	}

	// 3. Devolver configuración por defecto
	return l.defaultConfig(opts), nil
}

// TryLoadAgents intenta cargar AGENTS.md
func (l *Loader) TryLoadAgents() (*AgentsConfig, error) {
	rootPath, err := root.GetRoot()
	if err != nil {
		return nil, ErrNoAgentsFile
	}

	agentsPath := filepath.Join(rootPath, "AGENTS.md")
	if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
		return nil, ErrNoAgentsFile
	}

	config, err := ParseFile(agentsPath)
	if err != nil {
		// Si el archivo existe pero no es formato válido, retornar error específico
		if strings.Contains(err.Error(), "YAML-frontmatter") || strings.Contains(err.Error(), "frontmatter") {
			return nil, ErrInvalidFormat
		}
		return nil, err
	}

	return config, nil
}

// loadFromJuarvisYaml carga configuración de agentes desde juarvis.yaml
func (l *Loader) loadFromJuarvisYaml() (*AgentsConfig, error) {
	rootPath, err := root.GetRoot()
	if err != nil {
		return nil, err
	}

	juarvisPath := filepath.Join(rootPath, config.JuarFile)
	data, err := os.ReadFile(juarvisPath)
	if err != nil {
		return nil, err
	}

	// Parsear juarvis.yaml y extraer sección de agentes si existe
	var doc map[string]interface{}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	// Si hay sección agents, procesarla
	if agents, ok := doc["agents"]; ok {
		agentsData, err := yaml.Marshal(agents)
		if err != nil {
			return nil, err
		}

		return ParseFromYAML(string(agentsData))
	}

	return nil, errors.New("no se encontraron agentes en juarvis.yaml")
}

// defaultConfig retorna configuración por defecto
func (l *Loader) defaultConfig(opts *LoadOptions) *AgentsConfig {
	agents := []Agent{
		{
			Name:        "default",
			Description: "Agente por defecto de Juarvis",
			Skills:      []string{},
			Tools:       []string{"read", "write", "edit", "glob", "grep", "bash"},
			Mode:        ModeInteractive,
			MaxTurns:    50,
		},
	}

	if opts != nil && opts.DefaultAgent != "" {
		agents[0].Name = opts.DefaultAgent
	}

	return &AgentsConfig{
		Version: CurrentVersion,
		Agents:  agents,
	}
}

// applyOverrides aplica overrides desde CLI/flags
func applyOverrides(config *AgentsConfig, overrides map[string]string) *AgentsConfig {
	// Por ahora solo soportamos cambiar el modo y max_turns globalmente
	for key, value := range overrides {
		switch key {
		case "mode":
			for i := range config.Agents {
				config.Agents[i].Mode = AgentMode(value)
			}
		case "max_turns":
			for i := range config.Agents {
				config.Agents[i].MaxTurns = parseInt(value, config.Agents[i].MaxTurns)
			}
		}
	}
	return config
}

func parseInt(s string, def int) int {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return def
	}
	return n
}

// LoadOptions opciones para LoadWithPrecedence
type LoadOptions struct {
	DefaultAgent string
	Overrides    map[string]string
}

// FindAgentFile busca AGENTS.md en el directorio actual o en la raíz
func FindAgentFile(startDir string) (string, error) {
	// Primero buscar en el directorio actual
	agentsPath := filepath.Join(startDir, "AGENTS.md")
	if _, err := os.Stat(agentsPath); err == nil {
		return agentsPath, nil
	}

	// Buscar en la raíz del proyecto
	rootPath, err := root.GetRoot()
	if err != nil {
		return "", ErrNoAgentsFile
	}

	agentsPath = filepath.Join(rootPath, "AGENTS.md")
	if _, err := os.Stat(agentsPath); err == nil {
		return agentsPath, nil
	}

	return "", ErrNoAgentsFile
}

// GetRootPath retorna el path raíz donde se cargó la configuración
func (l *Loader) GetRootPath() string {
	return l.rootPath
}

// EnsureAgentsFile crea un AGENTS.md de ejemplo si no existe
func EnsureAgentsFile(rootPath string) error {
	agentsPath := filepath.Join(rootPath, "AGENTS.md")

	if _, err := os.Stat(agentsPath); err == nil {
		return nil // Ya existe
	}

	exampleConfig := `---
version: "1.0"
agents:
  - name: developer
    description: Desarrollador principal
    skills:
      - go
      - code-review
    tools:
      - read
      - edit
      - write
      - grep
      - glob
      - bash
    mode: autonomous

  - name: reviewer
    description: Revisor de código
    skills:
      - code-review
      - security
    tools:
      - read
      - grep
      - glob
    mode: interactive

  - name: explorer
    description: Explorador de código base
    skills:
      - analysis
    tools:
      - read
      - grep
      - glob
    mode: read-only

defaults:
  max_turns: 50
  mode: interactive
  tools:
    - read
    - grep
---

# AGENTS.md
# Este archivo define los agentes disponibles para Juarvis.
#
# Estructura:
# - version: Versión del formato (1.0)
# - agents: Lista de agentes
#   - name: Nombre único del agente
#   - description: Descripción breve
#   - skills: Skills disponibles para el agente
#   - tools: Herramientas permitidas
#   - mode: Modo de operación (interactive|autonomous|read-only)
#   - max_turns: Límite de turnos por sesión
#
# Precedencia de configuración:
# AGENTS.md > juarvis.yaml > flags > defaults
`

	return os.WriteFile(agentsPath, []byte(exampleConfig), 0644)
}

// IsValidVersion verifica si la versión es compatible
func IsValidVersion(version string) bool {
	return version != "" && (version == "1.0" || len(version) >= 3 && version[:2] == "1.")
}