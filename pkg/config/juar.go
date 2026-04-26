package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AutonomyLevel int            `json:"autonomy_level" yaml:"autonomy_level"`
	ProjectName   string         `json:"project_name" yaml:"project_name"`
	Version       string         `json:"version" yaml:"version"`
	Features      FeaturesConfig `json:"features" yaml:"features"`
}

type FeaturesConfig struct {
	AutoSnapshot bool `json:"auto_snapshot" yaml:"auto_snapshot"`
	AutoAnalyze  bool `json:"auto_analyze" yaml:"auto_analyze"`
	LearningLoop bool `json:"learning_loop" yaml:"learning_loop"`
	SDDPipeline  bool `json:"sdd_pipeline" yaml:"sdd_pipeline"`
}

var LevelNames = []string{
	"vibe-puro",
	"vibe-seguro",
	"vibe-estructurado",
	"semi-sdd",
	"sdd-completo",
}

const (
	LevelVibePuro = iota
	LevelVibeSeguro
	LevelVibeEstructurado
	LevelSemiSDD
	LevelSDDCompleto
)

var LevelDescriptions = []string{
	" Solo memoria + seguridad. Cero proceso. Máxima velocidad.",
	" + snapshot automático antes de cambios grandes.",
	" + descomposición de tareas antes de codear (default).",
	" + spec de aprobación antes de implementar.",
	" Pipeline SDD completo (Explore→propose→spec→design→tasks→apply→verify).",
}

func LoadOrCreate(rootPath string) (*Config, error) {
	// First: check versioned config in root (team-shared)
	configPath := filepath.Join(rootPath, JuarFile)
	if data, err := os.ReadFile(configPath); err == nil {
		var cfg Config
		if err := yaml.Unmarshal(data, &cfg); err == nil {
			return &cfg, nil
		}
	}

	// Second: check legacy local config (personal preferences)
	legacyPath := filepath.Join(rootPath, JuarvisDir, "config.yaml")
	if data, err := os.ReadFile(legacyPath); err == nil {
		var cfg Config
		if err := yaml.Unmarshal(data, &cfg); err == nil {
			// Migrate to versioned config
			cfg.Save(rootPath)
			return &cfg, nil
		}
	}

	// Create default config in root (versioned)
	cfg := &Config{
		AutonomyLevel: LevelVibeEstructurado,
		ProjectName:   filepath.Base(rootPath),
		Version:       "1.0",
		Features: FeaturesConfig{
			AutoSnapshot: true,
			AutoAnalyze:  true,
			LearningLoop: true,
			SDDPipeline:  false,
		},
	}

	if err := cfg.Save(rootPath); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Save(rootPath string) error {
	// Save to versioned config in root (team-shared)
	configPath := filepath.Join(rootPath, JuarFile)
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("error serializing config: %w", err)
	}

	return os.WriteFile(configPath, data, 0644)
}

func GetLevelName(level int) string {
	if level < 0 || level > 4 {
		return "unknown"
	}
	return LevelNames[level]
}

func (c *Config) SetLevel(level int) error {
	if level < 0 || level > 4 {
		return fmt.Errorf("nivel inválido: debe ser 0-4")
	}
	c.AutonomyLevel = level
	return nil
}

func (c *Config) ShouldAutoSnapshot() bool {
	return c.AutonomyLevel >= LevelVibeSeguro
}

func (c *Config) ShouldTaskDecomposition() bool {
	return c.AutonomyLevel >= LevelVibeEstructurado
}

func (c *Config) ShouldRequireSpec() bool {
	return c.AutonomyLevel >= LevelSemiSDD
}

func (c *Config) ShouldUseSDD() bool {
	return c.AutonomyLevel >= LevelSDDCompleto
}

var protocolTemplates = []string{
	0: `## Protocolo de Trabajo

**Modo activo: VIBE PURO (nivel 0)**
- Ejecuta directamente sin planes ni specs
- Minimiza iteración, maximiza velocidad
- Solo usa memoria + seguridad como gates
- Si algo falla, aplica Protocolo de Auto-Reparación`,
	1: `## Protocolo de Trabajo

**Modo activo: VIBE SEGURO (nivel 1)**
- Ejecuta directamente
- Crea snapshot antes de cambios grandes (≥5 líneas en ≥2 archivos)
- Si algo falla, aplica Protocolo de Auto-Reparación
- Minimiza iteración`,
	2: `## Protocolo de Trabajo

**Modo activo: ESTRUCTURADO (nivel 2)** - DEFAULT
- Descompón tareas antes de codear (≥3 pasos)
- Crea snapshot antes de cambios grandes
- Si algo falla, aplica Protocolo de Auto-Reparación
- Para refactors o arquitectura: genera plan estructurado antes de actuar`,
	3: `## Protocolo de Trabajo

**Modo activo: SEMI-SDD (nivel 3)**
- Requiere spec de approval antes de implementar features nuevas
- Descompón tareas antes de codear
- Snapshot obligatorio antes de cambios grandes
- Para arquitectura: pipeline SDD parcial (spec → apply → verify)`,
	4: `## Protocolo de Trabajo

**Modo activo: SDD COMPLETO (nivel 4)**
- Pipeline SDD OBLIGATORIO para cualquier feature o bugfix
- Fases: explore → propose → spec → design → tasks → apply → verify
- Snapshot en cada fase
- No commits hasta verify pasar
- Para cualquier cambio > 1 archivo: seguir pipeline completo`,
}

func (c *Config) GenerateAgentsSection() string {
	if c.AutonomyLevel >= 0 && c.AutonomyLevel < len(protocolTemplates) {
		return protocolTemplates[c.AutonomyLevel]
	}
	return "## Protocolo de Trabajo\n\n**Modo activo: DESCONOCIDO**\nConfiguracion invalida."
}

func (c *Config) UpdateAgentsSection(rootPath string) error {
	agentsPath := filepath.Join(rootPath, "AGENTS.md")
	if _, err := os.Stat(agentsPath); err != nil {
		return fmt.Errorf("AGENTS.md no encontrado en %s: %w", rootPath, err)
	}

	data, err := os.ReadFile(agentsPath)
	if err != nil {
		return fmt.Errorf("error leyendo AGENTS.md: %w", err)
	}

	content := string(data)
	newSection := c.GenerateAgentsSection()

	marker := "## Protocolo de Trabajo"
	idx := strings.Index(content, marker)
	if idx >= 0 {
		end := len(content) // default: reemplazar hasta el final si no hay siguiente ##
		for i := idx + len(marker); i < len(content)-1; i++ {
			if content[i] == '\n' && content[i+1] == '#' {
				end = i
				break
			}
		}
		content = content[:idx] + newSection + content[end:]
	} else {
		marker = "Prioridad de Reglas"
		idx := strings.Index(content, marker)
		if idx >= 0 {
			insertPoint := idx + len(marker)
			for i := insertPoint; i < len(content); i++ {
				if content[i] == '\n' && i+1 < len(content) && content[i+1] == '\n' {
					insertPoint = i + 2
					break
				}
			}
			content = content[:insertPoint] + "\n" + newSection + "\n" + content[insertPoint:]
		}
	}

	if err := os.WriteFile(agentsPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error escribiendo AGENTS.md: %w", err)
	}

	return nil
}
