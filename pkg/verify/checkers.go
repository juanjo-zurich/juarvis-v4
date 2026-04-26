// Package verify proporciona el sistema de verificación automática para Juarvis.
// Implementa un loop de verificación automático inspirado en Claude Code 2.1 xhigh.
package verify

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BuildVerifier verifica que el código compila correctamente.
type BuildVerifier struct{}

func (v *BuildVerifier) Nombre() string         { return "build" }
func (v *BuildVerifier) Descripcion() string    { return "Verifica que el código compila" }

func (v *BuildVerifier) Verificar(ctx context.Context, config *ConfigVerificacion) *ResultadoVerificacion {
	result := nuevoResultadoVerificacion("build")
	result.Ruta = config.Directorio

	// Determinar paquetes a verificar
	paquetes := config.Paquetes
	if len(paquetes) == 0 {
		paquetes = []string{"./..."}
	}

	args := append([]string{"build", "-v"}, paquetes...)
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = config.Directorio
	cmd.Stdout = nil
	cmd.Stderr = nil

	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Exitoso = false
		result.Error = string(output)
		result.Sugerencias = []string{"go mod tidy", "go build ./..."}
	}
	result.Metricas["paquetes"] = len(paquetes)

	return result
}

// TypeVerifier verifica tipos con go vet.
type TypeVerifier struct{}

func (v *TypeVerifier) Nombre() string         { return "type" }
func (v *TypeVerifier) Descripcion() string      { return "Verifica tipos y go vet" }

func (v *TypeVerifier) Verificar(ctx context.Context, config *ConfigVerificacion) *ResultadoVerificacion {
	result := nuevoResultadoVerificacion("type")
	result.Ruta = config.Directorio

	paquetes := config.Paquetes
	if len(paquetes) == 0 {
		paquetes = []string{"./..."}
	}

	args := append([]string{"vet"}, paquetes...)
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = config.Directorio

	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Exitoso = false
		result.Error = string(output)
		result.Sugerencias = []string{"go vet ./..."}
	}
	result.Metricas["checked"] = len(paquetes)

	return result
}

// ImportVerifier verifica imports no usados.
type ImportVerifier struct{}

func (v *ImportVerifier) Nombre() string         { return "import" }
func (v *ImportVerifier) Descripcion() string { return "Verifica imports no utilizados" }

func (v *ImportVerifier) Verificar(ctx context.Context, config *ConfigVerificacion) *ResultadoVerificacion {
	result := nuevoResultadoVerificacion("import")
	result.Ruta = config.Directorio

	// Usar goimports para verificar imports
	args := []string{"-l", "-d", config.Directorio}
	cmd := exec.CommandContext(ctx, "goimports", args...)
	cmd.Dir = config.Directorio

	output, err := cmd.CombinedOutput()
	salida := string(output)
	
	if err != nil || strings.TrimSpace(salida) != "" {
		// Filtrar solo líneas de imports
		lineas := strings.Split(salida, "\n")
		var importsMalos []string
		for _, linea := range lineas {
			if strings.Contains(linea, "import") && strings.TrimSpace(linea) != "" {
				importsMalos = append(importsMalos, linea)
			}
		}
		if len(importsMalos) > 0 {
			result.Exitoso = false
			result.Error = strings.Join(importsMalos, "\n")
			result.Sugerencias = []string{"goimports -w"}
		}
	}

	return result
}

// LintVerifier ejecuta linters.
type LintVerifier struct{}

func (v *LintVerifier) Nombre() string         { return "lint" }
func (v *LintVerifier) Descripcion() string { return "Ejecuta golangci-lint" }

func (v *LintVerifier) Verificar(ctx context.Context, config *ConfigVerificacion) *ResultadoVerificacion {
	result := nuevoResultadoVerificacion("lint")
	result.Ruta = config.Directorio

	// Intentar con golangci-lint primero
	args := []string{"run", "--issues-exit-code=1", "./..."}
	cmd := exec.CommandContext(ctx, "golangci-lint", args...)
	cmd.Dir = config.Directorio

	output, err := cmd.CombinedOutput()
	salida := string(output)

	// También intentar con go vet como respaldo
	if err != nil {
		vetCmd := exec.CommandContext(ctx, "go", "vet", "./...")
		vetCmd.Dir = config.Directorio
		vetOutput, vetErr := vetCmd.CombinedOutput()

		if err != nil && vetErr != nil {
			result.Exitoso = false
			result.Error = "golangci-lint: " + salida + "\ngo vet: " + string(vetOutput)
			result.Sugerencias = []string{"golangci-lint run --fix", "go vet"}
		}
	}

	// Verificar advertencias de estilo
	if strings.Contains(salida, "warning") || strings.Contains(salida, "warning") {
		result.Advertencias = append(result.Advertencias, "Existen advertencias de lint")
	}

	result.Metricas["linters"] = "golangci-lint,go vet"
	return result
}

// TestVerifier ejecuta tests.
type TestVerifier struct{}

func (v *TestVerifier) Nombre() string         { return "test" }
func (v *TestVerifier) Descripcion() string { return "Ejecuta tests y verifica resultados" }

func (v *TestVerifier) Verificar(ctx context.Context, config *ConfigVerificacion) *ResultadoVerificacion {
	result := nuevoResultadoVerificacion("test")
	result.Ruta = config.Directorio

	paquetes := config.Paquetes
	if len(paquetes) == 0 {
		paquetes = []string{"./..."}
	}

	// Construir argumentos para go test
	args := []string{"test", "-v", "-race", "-coverprofile=coverage.out"}
	args = append(args, paquetes...)

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = config.Directorio

	output, err := cmd.CombinedOutput()
	salida := string(output)

	if err != nil {
		result.Exitoso = false
		result.Error = salida
		result.Sugerencias = []string{"go test -v ./...", "Revisar logs de tests"}
	} else {
		// Extraer métricas de coverage si está disponible
		if strings.Contains(salida, "coverage:") {
			result.Advertencias = append(result.Advertencias, "Tests incompletos")
		}
	}

	// Intentar leer coverage
	if coverageData, err := os.ReadFile(filepath.Join(config.Directorio, "coverage.out")); err == nil {
		result.Metricas["coverage"] = string(coverageData)
	}

	result.Metricas["paquetes"] = len(paquetes)

	return result
}

// SecurityVerifier análisis de seguridad básico.
type SecurityVerifier struct{}

func (v *SecurityVerifier) Nombre() string         { return "security" }
func (v *SecurityVerifier) Descripcion() string  { return "Análisis de seguridad con gosec" }

func (v *SecurityVerifier) Verificar(ctx context.Context, config *ConfigVerificacion) *ResultadoVerificacion {
	result := nuevoResultadoVerificacion("security")
	result.Ruta = config.Directorio

	// Intentar con gosec
	args := []string{"./..."}
	cmd := exec.CommandContext(ctx, "gosec", args...)
	cmd.Dir = config.Directorio

	output, err := cmd.CombinedOutput()
	salida := string(output)

	if err != nil {
		// Si gosec no está instalado, intentar con basic go vet security
		vetArgs := []string{"vet", "-tags", "security", "./..."}
		vetCmd := exec.CommandContext(ctx, "go", vetArgs...)
		vetCmd.Dir = config.Directorio

		vetOutput, vetErr := vetCmd.CombinedOutput()
		if vetErr != nil && vetOutput != nil {
			result.Exitoso = false
			result.Error = "gosec no disponible: " + string(vetOutput)
			result.Advertencias = append(result.Advertencias, "Análisis de seguridad limitado")
		} else {
			result.Metricas["scanner"] = "go vet"
			return result
		}
	}

	if strings.Contains(salida, "Issues found") || strings.Contains(salida, "High") {
		result.Exitoso = false
		result.Error = salida
		result.Sugerencias = []string{"Revisar vulnerabilidades encontradas"}
		result.Advertencias = append(result.Advertencias, "Se encontraron posibles vulnerabilidades")
	}

	result.Metricas["scanner"] = "gosec"
	return result
}