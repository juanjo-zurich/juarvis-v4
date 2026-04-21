package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AutonomyLevel int `json:"autonomy_level" yaml:"autonomy_level"`
	ProjectName  string `json:"project_name" yaml:"project_name"`
	Version     string `json:"version" yaml:"version"`
	Features    FeaturesConfig `json:"features" yaml:"features"`
}

type FeaturesConfig struct {
	AutoSnapshot  bool `json:"auto_snapshot" yaml:"auto_snapshot"`
	AutoAnalyze   bool `json:"auto_analyze" yaml:"auto_analyze"`
	LearningLoop  bool `json:"learning_loop" yaml:"learning_loop"`
	SDDPipeline   bool `json:"sdd_pipeline" yaml:"sdd_pipeline"`
}

const (
	LevelVibePuro = iota
	LevelVibeSeguro
	LevelVibeEstructurado
	LevelSemiSDD
	LevelSDDCompleto
)

var LevelNames = []string{
	"vibe-puro",
	"vibe-seguro", 
	"vibe-estructurado",
	"semi-sdd",
	"sdd-completo",
}

var LevelDescriptions = []string{
	" Solo memoria + seguridad. Cero proceso. Máxima velocidad.",
	" + snapshot automático antes de cambios grandes.",
	" + descomposición de tareas antes de codear (default).",
	" + spec de aprobación antes de implementar.",
	" Pipeline SDD completo (Explore→propose→spec→design→tasks→apply→verify).",
}

func LoadOrCreate(rootPath string) (*Config, error) {
	configPath := filepath.Join(rootPath, JuarvisDir, "config.yaml")

	// Si existe, cargar
	if data, err := os.ReadFile(configPath); err == nil {
		var cfg Config
		if err := yaml.Unmarshal(data, &cfg); err == nil {
			return &cfg, nil
		}
	}

	// Crear default
	cfg := &Config{
		AutonomyLevel: LevelVibeEstructurado,
		ProjectName:  filepath.Base(rootPath),
		Version:      "1.0",
		Features: FeaturesConfig{
			AutoSnapshot: true,
			AutoAnalyze:  true,
			LearningLoop: true,
			SDDPipeline:  false,
		},
	}

	// Guardar
	if err := cfg.Save(rootPath); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Save(rootPath string) error {
	configDir := filepath.Join(rootPath, JuarvisDir)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio config: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("error serializando config: %w", err)
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