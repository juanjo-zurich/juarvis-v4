package agents

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
	"juarvis/pkg/utils"
)

// ParseErrors errores encontrados durante el parseo
type ParseError struct {
	Line   int
	Column int
	Message string
}

// ParseFile parsea un archivo AGENTS.md y retorna la configuración
func ParseFile(path string) (*AgentsConfig, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error leyendo archivo: %w", err)
	}

	return Parse(string(content))
}

// Parse parsea contenido YAML-frontmatter y retorna la configuración
func Parse(content string) (*AgentsConfig, error) {
	frontmatter, _, found := utils.ExtractFrontmatterBlock(content)
	if !found {
		return nil, errors.New("no se encontró bloque YAML-frontmatter (delimitado por ---)")
	}

	// Quitar los primeros 3 caracteres del inicio (después del primer ---)
	frontmatter = strings.TrimPrefix(frontmatter, "---")
	frontmatter = strings.TrimPrefix(frontmatter, "\n")
	frontmatter = strings.Trim(frontmatter, "\n")

	var config AgentsConfig
	if err := yaml.Unmarshal([]byte(frontmatter), &config); err != nil {
		return nil, fmt.Errorf("error parseando YAML: %w", err)
	}

	// Validar que hay agentes definidos
	if len(config.Agents) == 0 {
		return nil, fmt.Errorf("no se encontraron agentes en la configuración")
	}

	// Normalizar valores
	config.normalize()

	// Validar estructura básica
	if err := config.validateBasic(); err != nil {
		return nil, err
	}

	return &config, nil
}

// normalize normaliza los valores de la configuración
func (c *AgentsConfig) normalize() {
	// Normalizar versión
	c.Version = strings.TrimSpace(c.Version)

	// Normalizar agentes
	for i := range c.Agents {
		agent := &c.Agents[i]
		agent.Name = strings.TrimSpace(agent.Name)
		agent.Description = strings.TrimSpace(agent.Description)

		// Normalizar modo
		if agent.Mode != "" {
			agent.Mode = AgentMode(strings.ToLower(string(agent.Mode)))
		}

		// Normalizar skills
		for j := range agent.Skills {
			agent.Skills[j] = strings.TrimSpace(agent.Skills[j])
		}

		// Normalizar tools
		for j := range agent.Tools {
			agent.Tools[j] = strings.TrimSpace(agent.Tools[j])
		}
	}
}

// validateBasic validaciones básicas después del parseo
func (c *AgentsConfig) validateBasic() error {
	if c.Version == "" {
		return fmt.Errorf("versión no especificada en el frontmatter")
	}

	// Soportar múltiples versiones menores de la misma versión mayor
	if !strings.HasPrefix(c.Version, "1.") && c.Version != "1.0" {
		return fmt.Errorf("versión no soportada: %s (soportado: 1.0)", c.Version)
	}

	if len(c.Agents) == 0 {
		return fmt.Errorf("no se definieron agentes")
	}

	return nil
}

// Marshal serializa la configuración a formato YAML-frontmatter
func (c *AgentsConfig) Marshal() (string, error) {
	// Agregar versión por defecto si no existe
	if c.Version == "" {
		c.Version = CurrentVersion
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("error serializando YAML: %w", err)
	}

	return fmt.Sprintf("---\n%s\n---\n", string(data)), nil
}

// ToYAML serializa solo el contenido YAML (sin frontmatter)
func (c *AgentsConfig) ToYAML() (string, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("error serializando YAML: %w", err)
	}
	return string(data), nil
}

// ParseFromYAML parsea directamente YAML sin frontmatter
func ParseFromYAML(yamlContent string) (*AgentsConfig, error) {
	var config AgentsConfig
	if err := yaml.Unmarshal([]byte(yamlContent), &config); err != nil {
		return nil, fmt.Errorf("error parseando YAML: %w", err)
	}

	config.normalize()

	if err := config.validateBasic(); err != nil {
		return nil, err
	}

	return &config, nil
}

// DetectFormat detecta si el contenido es YAML-frontmatter o YAML plano
func DetectFormat(content string) string {
	frontmatter, _, found := utils.ExtractFrontmatterBlock(content)
	if found && strings.Contains(frontmatter, "version:") {
		return "frontmatter"
	}
	if strings.Contains(content, "version:") {
		return "yaml"
	}
	return "unknown"
}