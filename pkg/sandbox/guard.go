package sandbox

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Guardrails proporciona protección para la ejecución de comandos
type Guardrails struct {
	config   *Config
	inspector *Inspector
	execCount int64
	mu        sync.Mutex
}

// ResultadoEjecucion contiene el resultado de ejecutar un comando con sandbox
type ResultadoEjecucion struct {
	Exito           bool          `json:"success"`
	Salida          string        `json:"output"`
	Error          string        `json:"error,omitempty"`
	Duracion       time.Duration `json:"duration"`
	Bloqueado      bool          `json:"blocked"`
	RazonBloqueo   string        `json:"blockReason,omitempty"`
	Advertencia   bool          `json:"warning"`
	MensajeAdvertencia string    `json:"warningMessage,omitempty"`
	Cmd            string        `json:"command"`
	WorkingDir     string        `json:"workingDir"`
}

// NewGuardrails crea un nuevo Guardrails
func NewGuardrails(config *Config) (*Guardrails, error) {
	inspector, err := NewInspector(config)
	if err != nil {
		return nil, err
	}

	return &Guardrails{
		config:    config,
		inspector: inspector,
	}, nil
}

// VerificarComandoBeforeRun verifica un comando antes de ejecutarlo
func (g *Guardrails) VerificarComandoBeforeRun(comando string, args []string) error {
	// Verificar que el sandbox esté habilitado
	if !g.config.Enabled {
		return nil // Sandbox deshabilitado, permitir todo
	}

	// Validar argumentos
	if err := g.inspector.ValidateArgs(args); err != nil {
		return err
	}

	// Ejecutar inspección
	return g.inspector.ValidarComandoBeforeRun(comando, args)
}

// EjecutarComando ejecuta un comando con todas las protecciones del sandbox
func (g *Guardrails) EjecutarComando(
	ctx context.Context,
	comando string,
	args []string,
	opciones ...EjecutarOpcion,
) *ResultadoEjecucion {
	result := &ResultadoEjecucion{
		Cmd:        comando,
		WorkingDir: g.config.ObtenerDirectorioTrabajo(),
	}

	// Opciones por defecto
	opts := &ejecutarOptions{
		timeout:      g.config.ObtenerTimeout(),
		force:       false,
		workingDir:  g.config.ObtenerDirectorioTrabajo(),
		env:         os.Environ(),
	}

	for _, opt := range opciones {
		opt(opts)
	}

	startTime := time.Now()

	// 1. Verificar sandbox habilitado
	if !g.config.Enabled {
		result.Exito = true
		result.Salida = "Sandbox deshabilitado - ejecución directa"
		result.Duracion = time.Since(startTime)
		return result
	}

	// 2. Inspección del comando
	inspeccion := g.inspector.Inspect(comando, args)

	if !inspeccion.Permitido {
		result.Bloqueado = true
		result.RazonBloqueo = inspeccion.Motivo
		result.Duracion = time.Since(startTime)
		return result
	}

	// 3. Advertencias requieren aprobación (除非 force)
	if inspeccion.RequiereAprobacion && !opts.force {
		result.Advertencia = true
		result.MensajeAdvertencia = inspeccion.Motivo
		result.Duracion = time.Since(startTime)
		return result
	}

	// 4. Preparar entorno seguro
	env := g.prepararEntorno(opts.env)

	// 5. Configurar contexto con timeout
	if opts.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.timeout)
		defer cancel()
	}

	// 6. Ejecutar comando con restricciones
	cmd := exec.CommandContext(ctx, comando, args...)
	cmd.Dir = opts.workingDir
	cmd.Env = env

	// Restricciones según nivel
	if g.config.EsNivelEstrictoOSuperior() {
		// Restringir filesystem solo al workspace
		cmd.Dir = g.config.WorkspaceRoot
	}

	// Capturar输出
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 7. Ejecutar
	err := cmd.Run()

	// 8. Procesar resultado
	result.Duracion = time.Since(startTime)

	if err != nil {
		result.Error = err.Error()
		if stdout.Len() > 0 {
			result.Salida = stdout.String()
		}
		if stderr.Len() > 0 {
			result.Error += "\n" + stderr.String()
		}
	} else {
		result.Exito = true
		result.Salida = stdout.String()
	}

	// Actualizar contador
	g.mu.Lock()
	g.execCount++
	g.mu.Unlock()

	return result
}

// EjecutarString ejecuta un comando en string único
func (g *Guardrails) EjecutarString(
	ctx context.Context,
	comando string,
	opciones ...EjecutarOpcion,
) *ResultadoEjecucion {
	partes := strings.Fields(comando)
	if len(partes) == 0 {
		return &ResultadoEjecucion{
			Exito:  false,
			Error: "comando vacío",
		}
	}
	return g.EjecutarComando(ctx, partes[0], partes[1:], opciones...)
}

// prepararEntorno prepara un entorno seguro para la ejecución
func (g *Guardrails) prepararEntorno(baseEnv []string) []string {
	env := make([]string, 0, len(baseEnv)+10)

	// Filtrar variables peligrosas
	dangerousVars := []string{
		"LD_PRELOAD",
		"LD_LIBRARY_PATH",
		"LD_AUDIT",
		"LD_DEBUG",
		"BASH_ENV",
		"ENV",
		"CDPATH",
		"GLOBIGNORE",
		"BASH_FUNC_*",
	}

	for _, v := range baseEnv {
		skip := false
		for _, dv := range dangerousVars {
			if strings.HasPrefix(v, dv+"=") {
				skip = true
				break
			}
		}
		if !skip {
			env = append(env, v)
		}
	}

	// Agregar PATH seguro
	seguroPath := g.obtenerSafePath()
	env = append(env, "PATH="+seguroPath)

	// Restringir TMPDIR al workspace
	if g.config.WorkspaceRoot != "" {
		tmpDir := filepath.Join(g.config.WorkspaceRoot, ".tmp")
		os.MkdirAll(tmpDir, 0755)
		env = append(env, "TMPDIR="+tmpDir)
		env = append(env, "TMP="+tmpDir)
		env = append(env, "TEMP="+tmpDir)
	}

	// Deshabilitar variables de red si no está permitido
	if !g.config.PermitirRed {
		env = append(env, "HTTP_PROXY=")
		env = append(env, "HTTPS_PROXY=")
		env = append(env, "http_proxy=")
		env = append(env, "https_proxy=")
	}

	return env
}

// obtenerSafePath retorna un PATH seguro
func (g *Guardrails) obtenerSafePath() string {
	paths := []string{}

	// Agregar paths comunes seguros
	safePaths := []string{
		"/usr/local/bin",
		"/usr/bin",
		"/bin",
		"/usr/local/sbin",
		"/usr/sbin",
		"/sbin",
	}

	for _, p := range safePaths {
		if _, err := os.Stat(p); err == nil {
			paths = append(paths, p)
		}
	}

	// Agregar workspace bin si existe
	if g.config.WorkspaceRoot != "" {
		wsBin := filepath.Join(g.config.WorkspaceRoot, "bin")
		if _, err := os.Stat(wsBin); err == nil {
			paths = append(paths, wsBin)
		}
	}

	// Agregar node_modules/.bin si existe
	if g.config.WorkspaceRoot != "" {
		nmBin := filepath.Join(g.config.WorkspaceRoot, "node_modules", ".bin")
		if _, err := os.Stat(nmBin); err == nil {
			paths = append(paths, nmBin)
		}
	}

	return strings.Join(paths, string(filepath.ListSeparator))
}

// GenerarPromptAprobacion genera un mensaje de aprobación para comandos peligrosos
func (g *Guardrails) GenerarPromptAprobacion(comando string, args []string) string {
	inspeccion := g.inspector.Inspect(comando, args)

	var buf bytes.Buffer
	buf.WriteString("⚠️  COMANDO PELIGROSO DETECTADO\n\n")
	buf.WriteString("El sandbox ha detectado un comando que podría ser peligroso:\n\n")
	buf.WriteString("Comando: " + comando + " " + strings.Join(args, " ") + "\n\n")

	if inspeccion.Motivo != "" {
		buf.WriteString("Razón: " + inspeccion.Motivo + "\n\n")
	}

	if len(inspeccion.ComandosDetectados) > 0 {
		buf.WriteString("Patrones peligrosos detectados:\n")
		for _, c := range inspeccion.ComandosDetectados {
			buf.WriteString("  - " + c + "\n")
		}
		buf.WriteString("\n")
	}

	if len(inspeccion.PathsInseguros) > 0 {
		buf.WriteString("Paths involucrados:\n")
		for _, p := range inspeccion.PathsInseguros {
			buf.WriteString("  - " + p + "\n")
		}
		buf.WriteString("\n")
	}

	if inspeccion.SugereReemplazo != "" {
		buf.WriteString("Sugerencia: " + inspeccion.SugereReemplazo + "\n\n")
	}

	buf.WriteString("¿Deseas continuar de todas formas? (--force para omitir esta verificación)\n")
	buf.WriteString("Escribe 'sí' para aprobar o 'no' para cancelar: ")

	return buf.String()
}

// LeerAprobacion lee la aprobación del usuario desde stdin
func LeerAprobacion(prompt string) (bool, error) {
	fmt.Print(prompt)

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return false, scanner.Err()
	}

	respuesta := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return respuesta == "sí" || respuesta == "si" || respuesta == "s" || respuesta == "yes" || respuesta == "y", nil
}

// ObtenerContadorEjecuciones retorna el número de comandos ejecutados
func (g *Guardrails) ObtenerContadorEjecuciones() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.execCount
}

// ResetContador resetea el contador de ejecuciones
func (g *Guardrails) ResetContador() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.execCount = 0
}

// Tipo de opción para EjecutarComando
type EjecutarOpcion func(*ejecutarOptions)

type ejecutarOptions struct {
	timeout     time.Duration
	force      bool
	workingDir  string
	env        []string
	stdin      *bytes.Buffer
	stdout     *bytes.Buffer
	stderr     *bytes.Buffer
}

// WithTimeout establece el timeout para la ejecución
func WithTimeout(timeout time.Duration) EjecutarOpcion {
	return func(o *ejecutarOptions) {
		o.timeout = timeout
	}
}

// WithForce fuerza la ejecución sin aprobaciones
func WithForce(force bool) EjecutarOpcion {
	return func(o *ejecutarOptions) {
		o.force = force
	}
}

// WithWorkingDir establece el directorio de trabajo
func WithWorkingDir(dir string) EjecutarOpcion {
	return func(o *ejecutarOptions) {
		o.workingDir = dir
	}
}

// WithEnv establece variables de entorno
func WithEnv(env []string) EjecutarOpcion {
	return func(o *ejecutarOptions) {
		o.env = env
	}
}

// ObtenerStats retorna estadísticas del sandbox
func (g *Guardrails) ObtenerStats() map[string]interface{} {
	g.mu.Lock()
	defer g.mu.Unlock()

	return map[string]interface{}{
		"execCount":      g.execCount,
		"enabled":      g.config.Enabled,
		"level":        g.config.Nivel,
		"timeout":      g.config.Timeout,
		"allowNetwork": g.config.PermitirRed,
	}
}

// VerificarWorkspace verifica la integridad del workspace
func (g *Guardrails) VerificarWorkspace() ([]string, error) {
	if g.config.WorkspaceRoot == "" {
		return nil, fmt.Errorf("workspace no configurado")
	}

	var problemas []string

	// Verificar que el workspace existe
	if _, err := os.Stat(g.config.WorkspaceRoot); err != nil {
		problemas = append(problemas, "workspace no existe: "+g.config.WorkspaceRoot)
		return problemas, nil
	}

	// Verificar que tenemos acceso de lectura
	if !puedeLeer(g.config.WorkspaceRoot) {
		problemas = append(problemas, "sin acceso de lectura al workspace")
	}

	// Verificar que tenemos acceso de escritura
	if !puedeEscribir(g.config.WorkspaceRoot) {
		problemas = append(problemas, "sin acceso de escritura al workspace")
	}

	return problemas, nil
}

// puedeLeer verifica si se puede leer un directorio
func puedeLeer(path string) bool {
	testFile := filepath.Join(path, ".sandbox_check_read")
	if err := os.WriteFile(testFile, []byte("test"), 0444); err == nil {
		os.Remove(testFile)
		return true
	}
	return false
}

// puedeEscribir verifica si se puede escribir en un directorio
func puedeEscribir(path string) bool {
	testFile := filepath.Join(path, ".sandbox_check_write")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err == nil {
		os.Remove(testFile)
		return true
	}
	return false
}

// CrearGuardrailsSimple crea un Guardrails con configuración básica (para testing)
func CrearGuardrailsSimple(workspace string) *Guardrails {
	config := DefaultConfig(workspace)
	inspector, _ := NewInspector(config)
	return &Guardrails{
		config:    config,
		inspector: inspector,
	}
}

// ObtenerConfig retorna la configuración del sandbox
func (g *Guardrails) ObtenerConfig() *Config {
	return g.config
}

// EstablecerNivel cambia el nivel del sandbox
func (g *Guardrails) EstablecerNivel(nivel NivelSandbox) error {
	if err := g.config.VerificarNivel(); err != nil {
		return err
	}
	g.config.Nivel = nivel
	return nil
}

// HabilitarSandbox habilita o deshabilita el sandbox
func (g *Guardrails) HabilitarSandbox(enabled bool) {
	g.config.Enabled = enabled
}

// VerificarComandoSimple verifica un comando de forma simple
func VerificarComandoSimple(comando string) (bool, string) {
	config := DefaultConfig("/workspace")
	inspector, _ := NewInspector(config)
	result := inspector.InspectString(comando)
	return result.Permitido, result.Motivo
}

// EjecutarConTimeout ejecuta un comando con timeout específico
func EjecutarConTimeout(
	ctx context.Context,
	comando string,
	args []string,
	timeout time.Duration,
) *ResultadoEjecucion {
	config := DefaultConfig("/workspace")
	g, _ := NewGuardrails(config)
	return g.EjecutarComando(ctx, comando, args, WithTimeout(timeout))
}

// NewMockGuardrails crea un Guardrails para testing
func NewMockGuardrails() *Guardrails {
	return CrearGuardrailsSimple("/test/workspace")
}