// Package verify proporciona el sistema de verificación automática para Juarvis.
// Implementa un loop de verificación automático inspirado en Claude Code 2.1 xhigh.
package verify

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Verificador define la interfaz para todos los verificadores.
type Verificador interface {
	// Nombre retorna el nombre del verificador.
	Nombre() string
	// Verificar ejecuta la verificación y retorna el resultado.
	Verificar(ctx context.Context, config *ConfigVerificacion) *ResultadoVerificacion
	// Descripcion retorna una descripción de lo que verifica.
	Descripcion() string
}

// ResultadoVerificacion contiene el resultado de una verificación individual.
type ResultadoVerificacion struct {
	// Nombre del verificador.
	Nombre string `json:"nombre"`
	// true si la verificación pasó.
	Exitoso bool `json:"exitoso"`
	// Mensaje de error si falló.
	Error string `json:"error,omitempty"`
	// Advertencias encontradas.
	Advertencias []string `json:"advertencias,omitempty"`
	// Métricas adicionales (tiempo, coverage, etc.).
	Metricas map[string]interface{} `json:"metricas,omitempty"`
	// Timestamp de la verificación.
	Timestamp time.Time `json:"timestamp"`
	// Ruta del archivo o paquete verificado.
	Ruta string `json:"ruta,omitempty"`
	// Sugerencias de arreglo.
	Sugerencias []string `json:"sugerencias,omitempty"`
}

func nuevoResultadoVerificacion(nombre string) *ResultadoVerificacion {
	return &ResultadoVerificacion{
		Nombre:     nombre,
		Exitoso:    true,
		Metricas:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

// ReporteVerificacion reporte completo de verificación.
type ReporteVerificacion struct {
	// Estado general de la verificación.
	Exitoso bool `json:"exitoso"`
	// Modo de verificación utilizado.
	Modo ModoVerificacion `json:"modo"`
	// Resultados individuales.
	Resultados []*ResultadoVerificacion `json:"resultados"`
	// Número total de verificaciones.
	Total int `json:"total"`
	// Número de verificaciones exitosas.
	Exitosas int `json:"exitosas"`
	// Número de fallos.
	Fallos int `json:"fallos"`
	// Número de reintentos realizados.
	Reintentos int `json:"reintentos"`
	// Directorio del proyecto.
	Directorio string `json:"directorio"`
	// Timestamp de inicio.
	Inicio time.Time `json:"inicio"`
	// Timestamp de fin.
	Fin time.Time `json:"fin"`
	// Duración total.
	Duracion time.Duration `json:"duracion"`
	// Errores encontrados.
	Errores []string `json:"errores,omitempty"`
	// Advertencias acumuladas.
	Advertencias []string `json:"advertencias,omitempty"`
}

func nuevoReporteVerificacion(config *ConfigVerificacion) *ReporteVerificacion {
	return &ReporteVerificacion{
		Modo:       config.Modo,
		Resultados: make([]*ResultadoVerificacion, 0),
		Total:      0,
		Exitosas:   0,
		Fallos:     0,
		Reintentos: 0,
		Directorio: config.Directorio,
		Inicio:    time.Now(),
	}
}

// LoopVerificacion implementa el loop de verificación automática.
type LoopVerificacion struct {
	config        *ConfigVerificacion
	verificadores map[string]Verificador
	mu            sync.RWMutex
}

// NuevaInstanciaLoop crea una nueva instancia del loop de verificación.
func NuevaInstanciaLoop(config *ConfigVerificacion) *LoopVerificacion {
	if config == nil {
		cfg := ValorPorDefecto()
		config = &cfg
	}

	// Establecer directorio por defecto si no se especificó
	if config.Directorio == "" {
		dir, err := os.Getwd()
		if err == nil {
			config.Directorio = dir
		}
	}

	loop := &LoopVerificacion{
		config:        config,
		verificadores: make(map[string]Verificador),
	}

	// Registrar verificadores por defecto
	loop.RegistrarVerificador(&BuildVerifier{})
	loop.RegistrarVerificador(&TypeVerifier{})
	loop.RegistrarVerificador(&ImportVerifier{})
	loop.RegistrarVerificador(&LintVerifier{})
	loop.RegistrarVerificador(&TestVerifier{})
	loop.RegistrarVerificador(&SecurityVerifier{})

	return loop
}

// RegistrarVerificador registra un verificador.
func (l *LoopVerificacion) RegistrarVerificador(v Verificador) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.verificadores[v.Nombre()] = v
}

// Ejecutar ejecuta el loop de verificación.
func (l *LoopVerificacion) Ejecutar(ctx context.Context) (*ReporteVerificacion, error) {
	reporte := nuevoReporteVerificacion(l.config)
	verificadores := l.config.ObtenerVerificadores()

	if len(verificadores) == 0 {
		reporte.Exitoso = true
		reporte.Fin = time.Now()
		reporte.Duracion = time.Since(reporte.Inicio)
		return reporte, nil
	}

	// En modo xhigh, reintentar hasta que pase o se agoten los reintentos
	maxReintentos := 1
	if l.config.EsModoXHigh() {
		maxReintentos = l.config.MaxReintentos
		if maxReintentos < 1 {
			maxReintentos = 1
		}
	}

	for intento := 0; intento < maxReintentos; intento++ {
		if intento > 0 {
			reporte.Reintentos++
			fmt.Printf("Intento %d/%d...\n", intento+1, maxReintentos)
			time.Sleep(500 * time.Millisecond)
		}

		// Ejecutar cada verificador
		for _, nombre := range verificadores {
			resultado := l.ejecutarVerificador(ctx, nombre)
			reporte.Resultados = append(reporte.Resultados, resultado)
			reporte.Total++

			if resultado.Exitoso {
				reporte.Exitosas++
			} else {
				reporte.Fallos++
				if resultado.Error != "" {
					reporte.Errores = append(reporte.Errores,
						fmt.Sprintf("[%s] %s", nombre, resultado.Error))
				}
				if len(resultado.Advertencias) > 0 {
					reporte.Advertencias = append(reporte.Advertencias,
						resultado.Advertencias...)
				}
			}
		}

		// Verificar si todo pasó
		if reporte.Fallos == 0 {
			reporte.Exitoso = true
			break
		}

		// En modo xhigh, intentar arreglar automáticamente
		if l.config.EsModoXHigh() && !reporte.Exitoso && l.config.AplicarArreglosAuto {
			l.aplicarArreglos(reporte)
		}
	}

	reporte.Fin = time.Now()
	reporte.Duracion = time.Since(reporte.Inicio)
	reporte.Exitoso = reporte.Fallos == 0

	return reporte, nil
}

// ejecutarVerificador ejecuta un verificador específico.
func (l *LoopVerificacion) ejecutarVerificador(ctx context.Context, nombre string) *ResultadoVerificacion {
	l.mu.RLock()
	v, ok := l.verificadores[nombre]
	l.mu.RUnlock()

	if !ok {
		result := nuevoResultadoVerificacion(nombre)
		result.Exitoso = false
		result.Error = fmt.Sprintf("verificador '%s' no encontrado", nombre)
		return result
	}

	// Timeout específico para cada verificación
	ctx, cancel := context.WithTimeout(ctx, l.config.Timeout)
	defer cancel()

	return v.Verificar(ctx, l.config)
}

// aplicarArreglos intenta aplicar arreglos automáticos a los fallos.
func (l *LoopVerificacion) aplicarArreglos(reporte *ReporteVerificacion) {
	sugerenciasAgrupadas := make(map[string][]string)

	for _, resultado := range reporte.Resultados {
		if !resultado.Exitoso && len(resultado.Sugerencias) > 0 {
			for _, sug := range resultado.Sugerencias {
				sugerenciasAgrupadas[sug] = append(sugerenciasAgrupadas[sug], resultado.Nombre)
			}
		}
	}

	// Aplicar arreglos comunes
	for sugerencia, verificadores := range sugerenciasAgrupadas {
		fmt.Printf("Aplicando arreglo: %s (afecta a: %s)\n", sugerencia, strings.Join(verificadores, ", "))
		switch {
		case strings.Contains(sugerencia, "go mod tidy"):
			l.ejecutarComando("go", "mod", "tidy")
		case strings.Contains(sugerencia, "go fmt"):
			l.ejecutarComando("gofmt", "-w", l.config.Directorio)
		case strings.Contains(sugerencia, "goimports"):
			l.ejecutarComando("goimports", "-w", l.config.Directorio)
		}
	}
}

// ejecutarComando exec un comando en el directorio del proyecto.
func (l *LoopVerificacion) ejecutarComando(nombre string, args ...string) error {
	cmd := exec.Command(nombre, args...)
	cmd.Dir = l.config.Directorio
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// AgregarVerificadorPorNombre es un método público para agregar verificadores.
func (l *LoopVerificacion) AgregarVerificadorPorNombre(nombre string, v Verificador) {
	l.RegistrarVerificador(v)
}

// ObtenerVerificador retorna un verificador por nombre.
func (l *LoopVerificacion) ObtenerVerificador(nombre string) Verificador {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.verificadores[nombre]
}

// Config retorna la configuración actual.
func (l *LoopVerificacion) Config() *ConfigVerificacion {
	return l.config
}

// ActualizarConfig actualiza la configuración.
func (l *LoopVerificacion) ActualizarConfig(config *ConfigVerificacion) {
	l.config = config
}

// EjecutarVerificacion es una función de conveniencia para ejecutar verificación única.
func EjecutarVerificacion(ctx context.Context, config *ConfigVerificacion) (*ReporteVerificacion, error) {
	loop := NuevaInstanciaLoop(config)
	return loop.Ejecutar(ctx)
}

// BuscarYAML busca un archivo juarvis.yaml en el directorio o ancestros.
func BuscarYAML(dir string) (string, error) {
	buscar := dir
	for {
		ruta := filepath.Join(buscar, "juarvis.yaml")
		if _, err := os.Stat(ruta); err == nil {
			return ruta, nil
		}

		// Ir al directorio padre
		anterior := buscar
		buscar = filepath.Dir(buscar)
		if buscar == anterior || buscar == "." {
			break
		}
	}

	return "", fmt.Errorf("no se encontró juarvis.yaml en %s ni en ancestros", dir)
}

// CargarConfigDesdeProyecto carga la configuración desde el proyecto.
func CargarConfigDesdeProyecto(dir string) (*ConfigVerificacion, error) {
	ruta, err := BuscarYAML(dir)
	if err != nil {
		// Usar configuración por defecto
		config := ValorPorDefecto()
		config.Directorio = dir
		return &config, nil
	}

	config, err := CargarDesdeYAML(ruta)
	if err != nil {
		return nil, err
	}

	config.Directorio = filepath.Dir(ruta)
	return &config, nil
}