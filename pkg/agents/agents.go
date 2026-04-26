package agents

import (
	"fmt"
	"strings"
)

// Version actual del formato AGENTS.md
const CurrentVersion = "1.0"

// AgentMode define el modo de operación del agente
type AgentMode string

const (
	ModeInteractive AgentMode = "interactive"
	ModeAutonomous  AgentMode = "autonomous"
	ModeReadOnly    AgentMode = "read-only"
)

// ToolPermission define los permisos de herramientas para un agente
type ToolPermission struct {
	Name       string   `yaml:"name" json:"name"`
	Permission string   `yaml:"permission" json:"permission,omitempty"`
	Args       []string `yaml:"args,omitempty" json:"args,omitempty"`
}

// Agent representa un agente definido en AGENTS.md
type Agent struct {
	Name        string           `yaml:"name" json:"name"`
	Description string           `yaml:"description" json:"description"`
	Skills      []string         `yaml:"skills,omitempty" json:"skills,omitempty"`
	Tools       []string         `yaml:"tools,omitempty" json:"tools,omitempty"`
	Permissions []ToolPermission `yaml:"permissions,omitempty" json:"permissions,omitempty"`
	MaxTurns    int              `yaml:"max_turns,omitempty" json:"max_turns,omitempty"`
	Mode        AgentMode        `yaml:"mode,omitempty" json:"mode,omitempty"`
	Prompt      string           `yaml:"prompt,omitempty" json:"prompt,omitempty"`
	Env         map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
}

// AgentsConfig representa la configuración completa de AGENTS.md
type AgentsConfig struct {
	Version  string   `yaml:"version" json:"version"`
	Agents   []Agent `yaml:"agents" json:"agents"`
	Defaults *DefaultsConfig `yaml:"defaults,omitempty" json:"defaults,omitempty"`
}

// DefaultsConfig define valores por defecto para agentes
type DefaultsConfig struct {
	MaxTurns int      `yaml:"max_turns,omitempty" json:"max_turns,omitempty"`
	Mode     string   `yaml:"mode,omitempty" json:"mode,omitempty"`
	Tools    []string `yaml:"tools,omitempty" json:"tools,omitempty"`
}

// ValidationError representa errores de validación
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationResult resultado de la validación
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// GetAgent retorna un agente por nombre
func (c *AgentsConfig) GetAgent(name string) *Agent {
	for i := range c.Agents {
		if c.Agents[i].Name == name {
			return &c.Agents[i]
		}
	}
	return nil
}

// GetAgentNames retorna los nombres de todos los agentes
func (c *AgentsConfig) GetAgentNames() []string {
	names := make([]string, len(c.Agents))
	for i, a := range c.Agents {
		names[i] = a.Name
	}
	return names
}

// HasTool verifica si un agente tiene acceso a una herramienta
func (a *Agent) HasTool(tool string) bool {
	for _, t := range a.Tools {
		if strings.EqualFold(t, tool) {
			return true
		}
	}
	return false
}

// HasSkill verifica si un agente tiene una skill asignada
func (a *Agent) HasSkill(skill string) bool {
	for _, s := range a.Skills {
		if strings.EqualFold(s, skill) {
			return true
		}
	}
	return false
}

// String implementa fmt.Stringer para AgentMode
func (m AgentMode) String() string {
	return string(m)
}

// Validate verifica la estructura de un agente
func (a *Agent) Validate() []ValidationError {
	var errs []ValidationError

	if strings.TrimSpace(a.Name) == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "el nombre del agente es obligatorio"})
	}

	if strings.TrimSpace(a.Description) == "" {
		errs = append(errs, ValidationError{Field: "description", Message: "la descripción es obligatoria"})
	}

	// Validar modo si está presente
	if a.Mode != "" && a.Mode != ModeInteractive && a.Mode != ModeAutonomous && a.Mode != ModeReadOnly {
		errs = append(errs, ValidationError{
			Field:   "mode",
			Message: fmt.Sprintf("modo inválido: %s (debe ser: interactive, autonomous, read-only)", a.Mode),
		})
	}

	// Validar MaxTurns
	if a.MaxTurns < 0 {
		errs = append(errs, ValidationError{Field: "max_turns", Message: "max_turns no puede ser negativo"})
	}

	return errs
}

// Validate verifica la estructura completa del配置
func (c *AgentsConfig) Validate() ValidationResult {
	var errs []ValidationError

	// Validar versión
	if c.Version == "" {
		errs = append(errs, ValidationError{Field: "version", Message: "la versión es obligatoria"})
	}

	// Validar que hay agentes definidos
	if len(c.Agents) == 0 {
		errs = append(errs, ValidationError{Field: "agents", Message: "debe definir al menos un agente"})
	}

	// Validar cada agente
	seenNames := make(map[string]bool)
	for i, agent := range c.Agents {
		agentErrs := agent.Validate()
		for _, e := range agentErrs {
			e.Field = fmt.Sprintf("agents[%d].%s", i, e.Field)
			errs = append(errs, e)
		}

		// Verificar nombres duplicados
		if seenNames[agent.Name] {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("agents[%d].name", i),
				Message: fmt.Sprintf("nombre duplicado: %s", agent.Name),
			})
		}
		seenNames[agent.Name] = true
	}

	return ValidationResult{
		Valid:  len(errs) == 0,
		Errors: errs,
	}
}