package setup

import "encoding/json"

// UniversalManifest representa la configuración agnóstica del IDE
type UniversalManifest struct {
	Schema       string                 `json:"$schema,omitempty"`
	Agent        map[string]AgentConfig `json:"agent"`
	MCP          map[string]MCPConfig   `json:"mcp,omitempty"`
	Permission   PermissionConfig       `json:"permission,omitempty"`
	ContextPaths []string               `json:"contextPaths,omitempty"`
}

type AgentConfig struct {
	Description string          `json:"description,omitempty"`
	Mode        string          `json:"mode,omitempty"`
	Prompt      string          `json:"prompt"`
	Tools       map[string]bool `json:"tools,omitempty"`
}

type MCPConfig struct {
	Enabled bool     `json:"enabled"`
	Type    string   `json:"type"`
	Command []string `json:"command,omitempty"`
	URL     string   `json:"url,omitempty"`
}

type PermissionConfig struct {
	Bash map[string]string `json:"bash,omitempty"`
	Read map[string]string `json:"read,omitempty"`
}

// GenerateOpenCodeConfig produce un opencode.json válido a partir del manifiesto universal
func (m *UniversalManifest) GenerateOpenCodeConfig() ([]byte, error) {
	// Clonamos para no mutar el original
	data, _ := json.Marshal(m)
	var opencode map[string]interface{}
	json.Unmarshal(data, &opencode)

	// OpenCode espera $schema específico
	opencode["$schema"] = "https://opencode.ai/config.json"

	// Aquí podríamos filtrar claves no soportadas si fuera necesario
	// Por ahora mantenemos contextPaths solo si el usuario lo necesita,
	// pero lo manejamos de forma que no rompa versiones antiguas si es posible.

	return json.MarshalIndent(opencode, "", "  ")
}

// GenerateCursorConfig genera un .cursorrules a partir del prompt principal
func (m *UniversalManifest) GenerateCursorConfig() string {
	if orch, ok := m.Agent["orchestrator"]; ok {
		return orch.Prompt
	}
	return ""
}
