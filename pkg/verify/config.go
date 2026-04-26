// Package verify proporciona el sistema de verificación automática para Juarvis.
// Implementa un loop de verificación automático inspirado en Claude Code 2.1 xhigh.
package verify

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Modos de verificación disponibles.
type ModoVerificacion string

const (
	// Ninguna verificación automática.
	ModoNinguno ModoVerificacion = "none"
	// Verificación básica: build y sintaxis.
	ModoBasico ModoVerificacion = "basic"
	// Verificación estándar: tests y lint.
	ModoEstandar ModoVerificacion = "standard"
	// Verificación estricta: tests + coverage + security.
	ModoEstricto ModoVerificacion = "strict"
	// Verificación persistente hasta completar.
	ModoXHigh ModoVerificacion = "xhigh"
)

// ConfigVerificacion contiene la configuración del sistema de verificación.
type ConfigVerificacion struct {
	// Modo de verificación a utilizar.
	Modo ModoVerificacion `yaml:"mode"`
	// Si true, intenta aplicar correcciones automáticamente.
	AplicarArreglosAuto bool `yaml:"autoApplyFixes"`
	// Número máximo de reintentos en modo xhigh.
	MaxReintentos int `yaml:"maxRetries"`
	// Directorio de trabajo del proyecto.
	Directorio string `yaml:"-"`
	// Rutas a verificar (por defecto todas las .go en el proyecto).
	Rutas []string `yaml:"-"`
	// Paquetes específicos a verificar.
	Paquetes []string `yaml:"packages"`
	// Verificadores activos (por defecto todos).
	VerificadoresActivos []string `yaml:"activeVerifiers"`
	// Timeout para cada verificación individual.
	Timeout time.Duration `yaml:"timeout"`
	// Cobertura mínima requerida (solo para modo estricto).
	CoberturaMinima float64 `yaml:"minimumCoverage"`
	// Ignorar errores de verificación (continuar aunque fallen).
	IgnorarErrores bool `yaml:"ignoreErrors"`
	// Generar reporte de artefacto.
	GenerarReporte bool `yaml:"generateReport"`
}

// ValorPorDefecto retorna la configuración con valores por defecto.
func ValorPorDefecto() ConfigVerificacion {
	return ConfigVerificacion{
		Modo:              ModoEstandar,
		AplicarArreglosAuto: true,
		MaxReintentos:     3,
		Timeout:         5 * time.Minute,
		CoberturaMinima: 70.0,
		IgnorarErrores:   false,
		GenerarReporte:  true,
	}
}

// CargarDesdeYAML carga la configuración desde un archivo YAML sin dependencia externa.
func CargarDesdeYAML(ruta string) (ConfigVerificacion, error) {
	config := ValorPorDefecto()

	data, err := os.ReadFile(ruta)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return config, fmt.Errorf("error leyendo config: %w", err)
	}

	contenido := string(data)
	lineas := strings.Split(contenido, "\n")
	enSeccion := false

	for _, linea := range lineas {
		trimmed := strings.TrimSpace(linea)

		if strings.HasPrefix(trimmed, "verification:") {
			enSeccion = true
			continue
		}

		if enSeccion {
			if trimmed == "" {
				continue
			}
			// Nueva sección de primer nivel
			if !strings.HasPrefix(trimmed, " ") && !strings.HasPrefix(trimmed, "\t") {
				break
			}

			// Parsear clave: valor
			if strings.Contains(trimmed, "mode:") {
				partes := strings.Split(trimmed, "mode:")
				if len(partes) > 1 {
					valor := strings.TrimSpace(partes[1])
					valor = strings.Trim(valor, "\"")
					switch valor {
					case "none", "basic", "standard", "strict", "xhigh":
						config.Modo = ModoVerificacion(valor)
					}
				}
			}

			if strings.Contains(trimmed, "autoApplyFixes:") {
				partes := strings.Split(trimmed, "autoApplyFixes:")
				if len(partes) > 1 {
					config.AplicarArreglosAuto = strings.TrimSpace(partes[1]) == "true"
				}
			}

			if strings.Contains(trimmed, "maxRetries:") {
				partes := strings.Split(trimmed, "maxRetries:")
				if len(partes) > 1 {
					var retries int
					fmt.Sscanf(strings.TrimSpace(partes[1]), "%d", &retries)
					config.MaxReintentos = retries
				}
			}

			if strings.Contains(trimmed, "timeout:") {
				partes := strings.Split(trimmed, "timeout:")
				if len(partes) > 1 {
					config.Timeout = parsearDuracion(strings.TrimSpace(partes[1]))
				}
			}

			if strings.Contains(trimmed, "minimumCoverage:") {
				partes := strings.Split(trimmed, "minimumCoverage:")
				if len(partes) > 1 {
					valor := strings.TrimSpace(partes[1])
					valor = strings.Trim(valor, "%")
					fmt.Sscanf(valor, "%f", &config.CoberturaMinima)
				}
			}

			if strings.Contains(trimmed, "ignoreErrors:") {
				partes := strings.Split(trimmed, "ignoreErrors:")
				if len(partes) > 1 {
					config.IgnorarErrores = strings.TrimSpace(partes[1]) == "true"
				}
			}

			if strings.Contains(trimmed, "generateReport:") {
				partes := strings.Split(trimmed, "generateReport:")
				if len(partes) > 1 {
					config.GenerarReporte = strings.TrimSpace(partes[1]) == "true"
				}
			}
		}
	}

	config.Directorio = filepath.Dir(ruta)
	return config, nil
}

// parsearDuracion parsea una duración string simple.
func parsearDuracion(s string) time.Duration {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\"")

	switch {
	case strings.HasSuffix(s, "m"):
		var m int
		fmt.Sscanf(s, "%dm", &m)
		return time.Duration(m) * time.Minute
	case strings.HasSuffix(s, "s"):
		var sec int
		fmt.Sscanf(s, "%ds", &sec)
		return time.Duration(sec) * time.Second
	case strings.HasSuffix(s, "h"):
		var h int
		fmt.Sscanf(s, "%dh", &h)
		return time.Duration(h) * time.Hour
	}

	return 5 * time.Minute
}

// ObtenerVerificadores retornar la lista de verificadores según el modo.
func (c *ConfigVerificacion) ObtenerVerificadores() []string {
	if len(c.VerificadoresActivos) > 0 {
		return c.VerificadoresActivos
	}

	switch c.Modo {
	case ModoNinguno:
		return []string{}
	case ModoBasico:
		return []string{"build", "type", "import"}
	case ModoEstandar:
		return []string{"build", "type", "import", "lint", "test"}
	case ModoEstricto:
		return []string{"build", "type", "import", "lint", "test", "security"}
	case ModoXHigh:
		return []string{"build", "type", "import", "lint", "test", "security"}
	default:
		return []string{"build", "type", "import", "lint", "test"}
	}
}

// EsModoXHigh verifica si el modo es xhigh (persistente).
func (c *ConfigVerificacion) EsModoXHigh() bool {
	return c.Modo == ModoXHigh
}

// NecesitaTests verifica si el modo actual incluye tests.
func (c *ConfigVerificacion) NecesitaTests() bool {
	return c.Modo == ModoEstandar || c.Modo == ModoEstricto || c.Modo == ModoXHigh
}

// NecesitaLint verifica si el modo actual incluye linting.
func (c *ConfigVerificacion) NecesitaLint() bool {
	return c.Modo == ModoEstandar || c.Modo == ModoEstricto || c.Modo == ModoXHigh
}

// NecesitaSecurity verifica si el modo actual incluye análisis de seguridad.
func (c *ConfigVerificacion) NecesitaSecurity() bool {
	return c.Modo == ModoEstricto || c.Modo == ModoXHigh
}

// NecesitaCobertura verifica si el modo actual requiere análisis de coverage.
func (c *ConfigVerificacion) NecesitaCobertura() bool {
	return c.Modo == ModoEstricto || c.Modo == ModoXHigh
}

const VersionGoRequerida = "1.25"

// VerificarRequisitos verifica los requisitos del sistema.
func VerificarRequisitos() error {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("go no encontrado: %w", err)
	}

	version := strings.TrimSpace(string(output))
	if !strings.Contains(version, "go1.2") && !strings.Contains(version, "go1.1") && !strings.Contains(version, "go1.") {
		return fmt.Errorf("versión de Go incompatible: %s", version)
	}

	return nil
}