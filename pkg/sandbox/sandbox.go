package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// NivelSandbox define el nivel de restricción del sandbox
type NivelSandbox string

const (
	// Ninguno: Sin restricciones
	Ninguno NivelSandbox = "none"
	// Basico: Solo path restriction
	Basico NivelSandbox = "basic"
	// Estricto: Path + blacklist + timeout
	Estricto NivelSandbox = "strict"
	// Paranoico: Máximo security
	Paranoico NivelSandbox = "paranoid"
)

// Config contiene la configuración del sandbox
type Config struct {
	Enabled             bool          `json:"enabled" yaml:"enabled"`
	Nivel              NivelSandbox `json:"level" yaml:"level"`
	PermitirRed        bool          `json:"allowNetwork" yaml:"allowNetwork"`
	Timeout            time.Duration `json:"timeout" yaml:"timeout"`
	ComandosBlacklist  []string      `json:"blacklistedCommands" yaml:"blacklistedCommands"`
	LimiteCPU          int           `json:"cpuLimit" yaml:"cpuLimit"`           // porcentaje (0 = sin límite)
	LimiteMemoria      int           `json:"memoryLimit" yaml:"memoryLimit"`     // MB (0 = sin límite)
	MaxArchivosEscritura int         `json:"maxFilesWrite" yaml:"maxFilesWrite"` // max archivos a escribir
	WorkspaceRoot      string        `json:"-" yaml:"-"`
}

// Valores por defecto para comandos blacklistados
var ComandosBlacklistPredeterminados = []string{
	"rm -rf",
	"rm -rf /",
	"sudo",
	"chmod 777",
	"chmod -R 777",
	"chown",
	"chgrp",
	":(){:|:&};:",          // fork bomb
	"mkfs",
	"dd if=/dev/zero of=",
	"> /dev/sda",
	"curl | sh",
	"wget -O- | sh",
	"curl -sL | bash",
	"wget -qO- | bash",
	"> ~/.bashrc",
	"> ~/.bash_profile",
	"> /etc/passwd",
	"> /etc/shadow",
	"nc -e /bin/sh",
	"bash -i >& /dev/tcp/",
	"python.*-m http.server",
	"-d background",
}

// AllowedExtensions extensiones permitidas para ejecución
var AllowedExtensions = []string{
	".sh", ".bash",
	".py",
	".js", ".ts",
	".go",
	".rb",
	".pl",
}

// DefaultConfig retorna una configuración por defecto
func DefaultConfig(workspaceRoot string) *Config {
	return &Config{
		Enabled:             true,
		Nivel:              Estricto,
		PermitirRed:        false,
		Timeout:            30 * time.Second,
		ComandosBlacklist:  ComandosBlacklistPredeterminados,
		LimiteCPU:          50, // 50% CPU
		LimiteMemoria:       512, // 512MB
		MaxArchivosEscritura: 100,
		WorkspaceRoot:      workspaceRoot,
	}
}

// Load configura el sandbox desde juarvis.yaml
func Load(rootPath string) (*Config, error) {
	configPath := filepath.Join(rootPath, "juarvis.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Si no existe, usar configuración por defecto
		return DefaultConfig(rootPath), nil
	}

	// Parsear YAML directamente buscando la sección sandbox
	var rawConfig struct {
		Sandbox *yaml.Node `yaml:"sandbox"`
	}

	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return DefaultConfig(rootPath), nil
	}

	if rawConfig.Sandbox == nil {
		return DefaultConfig(rootPath), nil
	}

	// Decodificar la sección sandbox
	cfg := &Config{WorkspaceRoot: rootPath}
	if err := rawConfig.Sandbox.Decode(cfg); err != nil {
		return nil, fmt.Errorf("error decodificando configuración de sandbox: %w", err)
	}

	// Aplicar valores por defecto si no se especificaron
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if len(cfg.ComandosBlacklist) == 0 {
		cfg.ComandosBlacklist = ComandosBlacklistPredeterminados
	}
	if cfg.Nivel == "" {
		cfg.Nivel = Estricto
	}

	return cfg, nil
}

// Save guarda la configuración en juarvis.yaml
func (c *Config) Save(rootPath string) error {
	configPath := filepath.Join(rootPath, "juarvis.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Crear archivo nuevo
		data = []byte{}
	}

	var doc map[string]interface{}
	yaml.Unmarshal(data, &doc)
	if doc == nil {
		doc = make(map[string]interface{})
	}

	// Actualizar sección sandbox
	doc["sandbox"] = map[string]interface{}{
		"enabled":             c.Enabled,
		"level":              c.Nivel,
		"allowNetwork":       c.PermitirRed,
		"timeout":            c.Timeout.String(),
		"blacklistedCommands": c.ComandosBlacklist,
		"cpuLimit":           c.LimiteCPU,
		"memoryLimit":        c.LimiteMemoria,
		"maxFilesWrite":     c.MaxArchivosEscritura,
	}

	out, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("error serializando config: %w", err)
	}

	return os.WriteFile(configPath, out, 0644)
}

// VerificarNivel valida que el nivel sea válido
func (c *Config) VerificarNivel() error {
	switch c.Nivel {
	case Ninguno, Basico, Estricto, Paranoico:
		return nil
	default:
		return fmt.Errorf("nivel de sandbox inválido: %s (válidos: none, basic, strict, paranoid)", c.Nivel)
	}
}

// EsNivelParanoico retorna true si el nivel es paranoid
func (c *Config) EsNivelParanoico() bool {
	return c.Nivel == Paranoico
}

// EsNivelEstrictoOSuperior retorna true si el nivel es strict o paranoid
func (c *Config) EsNivelEstrictoOSuperior() bool {
	return c.Nivel == Estricto || c.Nivel == Paranoico
}

// ObtenerTimeout retorna el timeout configurado
func (c *Config) ObtenerTimeout() time.Duration {
	if c.Timeout == 0 {
		return 30 * time.Second
	}
	return c.Timeout
}

// NormalizarPath convierte un path a su forma canónica y verifica que esté dentro del workspace
func (c *Config) NormalizarPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path vacío")
	}

	// Convertir a path absoluto
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("error obteniendo path absoluto: %w", err)
	}

	// Resolver symlinks para evitar escapes
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		realPath = absPath
	}

	// Normalizar separadores
	realPath = filepath.Clean(realPath)

	// Verificar que está dentro del workspace
	if c.WorkspaceRoot != "" {
		wsRoot, err := filepath.EvalSymlinks(c.WorkspaceRoot)
		if err != nil {
			wsRoot = c.WorkspaceRoot
		}
		wsRoot = filepath.Clean(wsRoot)

		// Verificar prefijo
		if !strings.HasPrefix(realPath, wsRoot+string(filepath.Separator)) &&
			realPath != wsRoot {
			return "", fmt.Errorf("path fuera del workspace: %s (workspace: %s)", realPath, wsRoot)
		}
	}

	return realPath, nil
}

// EnWorkspace verifica si un path está dentro del workspace
func (c *Config) EnWorkspace(path string) bool {
	normalized, err := c.NormalizarPath(path)
	if err != nil {
		return false
	}
	return normalized != ""
}

// ObtenerDirectorioTrabajo retorna el directorio de trabajo seguro
func (c *Config) ObtenerDirectorioTrabajo() string {
	if c.WorkspaceRoot != "" {
		return c.WorkspaceRoot
	}
	// Fallback al directorio actual
	cwd, _ := os.Getwd()
	return cwd
}