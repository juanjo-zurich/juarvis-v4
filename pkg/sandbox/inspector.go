package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ResultadoInspeccion contiene el resultado de inspeccionar un comando
type ResultadoInspeccion struct {
	Permitido          bool     `json:"allowed"`
	NivelBloqueo       string   `json:"blockLevel"` // "none", "warn", "block"
	Motivo             string   `json:"reason"`
	ComandosDetectados []string `json:"detectedCommands"`
	PathsInseguros     []string `json:"unsafePaths"`
	SugereReemplazo   string   `json:"suggestedReplacement,omitempty"`
	EsPeligroso       bool     `json:"isDangerous"`
	RequiereAprobacion bool    `json:"requiresApproval"`
}

// Inspector analiza comandos para detectar patrones peligrosos
type Inspector struct {
	config        *Config
	patronesPeligrosos []PatternMatcher
	regexBlacklist []*regexp.Regexp
}

// PatternMatcher representa un patrón de comando peligroso
type PatternMatcher struct {
	Patron    string
	Regex    *regexp.Regexp
	Mensaje  string
	Severidad string // "block", "warn"
}

// NewInspector crea un nuevo inspector de comandos
func NewInspector(config *Config) (*Inspector, error) {
	i := &Inspector{
		config: config,
	}

	// Compilar regex para blacklist
	for _, cmd := range config.ComandosBlacklist {
		// Crear regex que haga match del comando completo o parcial
		// Escape special chars y crear patrón flexible
		patron := escapeRegex(cmd)
		re, err := regexp.Compile(`(?i)(` + patron + `)`)
		if err != nil {
			continue
		}
		i.regexBlacklist = append(i.regexBlacklist, re)
	}

	// Patrones peligrosos adicionales
	i.patronesPeligrosos = []PatternMatcher{
		{
			Patron:    "rm.*-rf",
			Mensaje:  "Comando de eliminación recursiva muy peligroso",
			Severidad: "block",
		},
		{
			Patron:    "curl.*\\|.*(sh|bash)",
			Mensaje:  "Ejecutando script descargado - muy peligroso",
			Severidad: "block",
		},
		{
			Patron:    "wget.*\\|.*(sh|bash)",
			Mensaje:  "Ejecutando script descargado - muy peligroso",
			Severidad: "block",
		},
		{
			Patron:    "sudo.*(rm|del|mkfs)",
			Mensaje:  "Comando con privilegios de root",
			Severidad: "block",
		},
		{
			Patron:    "chmod.*777",
			Mensaje:  "Permisos demasiado abiertos",
			Severidad: "warn",
		},
		{
			Patron:    "dd.*of=.*(sda|hda)",
			Mensaje:  "Escritura directa a disco del sistema",
			Severidad: "block",
		},
		{
			Patron:    ">.*/(passwd|shadow|group)",
			Mensaje:  "Intento de modificar archivos críticos del sistema",
			Severidad: "block",
		},
		{
			Patron:    "nc.*-e",
			Mensaje:  "Reverse shell - posible intento de conexión remota",
			Severidad: "block",
		},
		{
			Patron:    "bash.*>&.*/dev/tcp",
			Mensaje:  "Reverse shell - posible intento de conexión remota",
			Severidad: "block",
		},
		{
			Patron:    ":\\(\\)*:.*\\|:.*&",
			Mensaje:  "Fork bomb - consumo masivo de recursos",
			Severidad: "block",
		},
	}

	// Compilar patrones adicionales
	for idx := range i.patronesPeligrosos {
		re, err := regexp.Compile(i.patronesPeligrosos[idx].Patron)
		if err != nil {
			continue
		}
		i.patronesPeligrosos[idx].Regex = re
	}

	return i, nil
}

// Inspect analiza un comando y sus argumentos para detectar peligros
func (i *Inspector) Inspect(comando string, args []string) *ResultadoInspeccion {
	result := &ResultadoInspeccion{
		Permitido:  true,
		NivelBloqueo: "none",
	}

	// Combinar comando y args para análisis
	comandoCompleto := comando
	if len(args) > 0 {
		comandoCompleto += " " + strings.Join(args, " ")
	}

	// 1. Verificar blacklist de comandos
	for _, re := range i.regexBlacklist {
		if matches := re.FindStringSubmatch(comandoCompleto); len(matches) > 0 {
			result.ComandosDetectados = append(result.ComandosDetectados, matches[0])
			result.EsPeligroso = true
			result.NivelBloqueo = "block"
			result.Motivo = fmt.Sprintf("comando en blacklist: %s", matches[0])
			result.Permitido = false
			return result
		}
	}

	// 2. Verificar patrones peligrosos adicionales
	for _, patron := range i.patronesPeligrosos {
		if patron.Regex != nil && patron.Regex.MatchString(comandoCompleto) {
			result.ComandosDetectados = append(result.ComandosDetectados, patron.Patron)
			result.EsPeligroso = true

			if patron.Severidad == "block" {
				result.NivelBloqueo = "block"
				result.Motivo = patron.Mensaje
				result.Permitido = false
			} else {
				result.NivelBloqueo = "warn"
				result.Motivo = patron.Mensaje
				result.RequiereAprobacion = true
			}
		}
	}

	// 3. Verificar paths en el comando
	result.PathsInseguros = i.extraerPathsInseguros(comandoCompleto)

	// 4. En nivel strict o superior, verificar paths fuera del workspace
	if i.config.EsNivelEstrictoOSuperior() && len(result.PathsInseguros) > 0 {
		for _, path := range result.PathsInseguros {
			if !i.config.EnWorkspace(path) {
				result.Permitido = false
				result.NivelBloqueo = "block"
				result.Motivo = fmt.Sprintf("path fuera del workspace: %s", path)
				return result
			}
		}
	}

	// 5. Verificar si el comando intenta escapar del workspace
	if i.comandoEscapaDelWorkspace(comandoCompleto) {
		result.Permitido = false
		result.NivelBloqueo = "block"
		result.Motivo = "comando intenta acceder fuera del workspace"
		result.EsPeligroso = true
	}

	// 6. Verificar intentos de network si no está permitido
	if !i.config.PermitirRed && i.comandoUsaRed(comandoCompleto) {
		if !result.Permitido {
			result.Motivo += " + red no permitida"
		} else {
			result.Permitido = false
			result.NivelBloqueo = "block"
			result.Motivo = "comando intenta acceder a red pero allowNetwork=false"
			result.EsPeligroso = true
		}
	}

	// 7. En nivel paranoid, verificar más restricciones
	if i.config.EsNivelParanoico() {
		if i.comandoEjecutaBinario(comandoCompleto) {
			result.RequiereAprobacion = true
			if result.NivelBloqueo == "none" {
				result.NivelBloqueo = "warn"
			}
		}
	}

	// Si no es peligroso pero requiere aprobación, sugiero alternativas
	if result.RequiereAprobacion && !result.EsPeligroso {
		result.SugereReemplazo = i.sugerirAlternativa(comandoCompleto)
	}

	return result
}

// InspectString analiza un comando en string único
func (i *Inspector) InspectString(comando string) *ResultadoInspeccion {
	partes := strings.Fields(comando)
	if len(partes) == 0 {
		return &ResultadoInspeccion{
			Permitido: false,
			Motivo:    "comando vacío",
		}
	}
	return i.Inspect(partes[0], partes[1:])
}

// extraerPathsInseguros extrae paths que podrían estar fuera del workspace
func (i *Inspector) extraerPathsInseguros(comando string) []string {
	var paths []string

	partes := strings.Fields(comando)
	for idx, parte := range partes {
		// Es un path si empiexa con / o ~ o .. o ./
		if strings.HasPrefix(parte, "/") ||
			strings.HasPrefix(parte, "~") ||
			strings.HasPrefix(parte, "../") ||
			strings.HasPrefix(parte, "./") ||
			regexp.MustCompile(`^[a-zA-Z]:\\`).MatchString(parte) {
			// Verificar que no sea un flag
			if idx > 0 && !strings.HasPrefix(partes[idx-1], "-") {
				// Es un argumento de path
				expanded := os.ExpandEnv(parte)
				paths = append(paths, expanded)
			} else if idx == 0 {
				// Es el comando mismo
				expanded := os.ExpandEnv(parte)
				paths = append(paths, expanded)
			}
		}
	}

	// También verificar el directorio actual
	cwd, _ := os.Getwd()
	paths = append(paths, cwd)

return uniquePaths(paths)
}

// comandoEscapaDelWorkspace verifica si el comando intenta escapar
func (i *Inspector) comandoEscapaDelWorkspace(comando string) bool {
	// Patrones que indican intento de escape
	escapePatterns := []string{
		`\.\./`,
		`\.\.$`,
		`cd\s+/`,
		`cd\s+~`,
		`~\s*`,
		`/home/`,
		`/root/`,
		`/etc/`,
		`/var/`,
		`/usr/bin`,
		`/usr/local/bin`,
		`/opt/`,
		`/tmp/`,
	}

	for _, patron := range escapePatterns {
		if regexp.MustCompile(patron).MatchString(comando) {
			return true
		}
	}

	return false
}

// comandoUsaRed verifica si el comando usa red
func (i *Inspector) comandoUsaRed(comando string) bool {
	comandosRed := []string{
		"curl", "wget", "nc", "netcat", "ssh", "scp", "sftp",
		"telnet", "ftp", "ncftp", "lynx", "elinks",
		"git clone", "git pull", "git push",
		"pip install", "npm install -g", "gem install",
		"cargo install", "go install",
		"docker pull", "docker run",
		"kubectl", "helm",
		"aws", "gcloud", "az",
	}

	comandoLower := strings.ToLower(comando)
	for _, cmd := range comandosRed {
		if strings.Contains(comandoLower, cmd) {
			return true
		}
	}

	return false
}

// comandoEjecutaBinario verifica si el comando ejecuta binarios externos
func (i *Inspector) comandoEjecutaBinario(comando string) bool {
	comandosBinarios := []string{
		"/bin/", "/usr/bin/", "/usr/local/bin/",
		".sh", ".bash", ".zsh",
		".py", ".pl", ".rb",
		".exe", ".app", ".bin",
	}

	for _, cmd := range comandosBinarios {
		if strings.Contains(comando, cmd) {
			return true
		}
	}

	return false
}

// sugerirAlternativa sugiere una alternativa más segura
func (i *Inspector) sugerirAlternativa(comando string) string {
	comandoLower := strings.ToLower(comando)

	// curl | sh -> descargar primero y revisar
	if strings.Contains(comandoLower, "curl |") || strings.Contains(comandoLower, "wget |") {
		return "Descarga el script primero, revísalo, y luego ejecútalo manualmente"
	}

	// chmod 777 -> usar permisos específicos
	if strings.Contains(comandoLower, "chmod 777") {
		return "Usa chmod 755 para directorios o chmod 644 para archivos"
	}

	// rm -rf sin verificación
	if strings.Contains(comandoLower, "rm -rf") {
		return "Usa rm -rf con precaución. Considera usar rm -ri para confirmación interactiva"
	}

	return ""
}

// uniquePaths elimina duplicados
func uniquePaths(paths []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, p := range paths {
		clean := filepath.Clean(p)
		if !seen[clean] {
			seen[clean] = true
			result = append(result, clean)
		}
	}
	return result
}

// escapeRegex escapa caracteres especiales para regex
func escapeRegex(s string) string {
	// Mapa de reemplazos
	replacements := map[string]string{
		"\\": "\\\\",
		".": "\\.",
		"*": "\\*",
		"+": "\\+",
		"?": "\\?",
		"[": "\\[",
		"]": "\\]",
		"^": "\\^",
		"$": "\\$",
		"(": "\\(",
		")": "\\)",
		"|": "\\|",
		"{": "\\{",
		"}": "\\}",
		"=": "\\=",
		" ": "\\s+",
	}

	result := s
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}
	return result
}

// ValidarComandoBeforeRun ejecuta todas las verificaciones antes de permitir ejecución
func (i *Inspector) ValidarComandoBeforeRun(comando string, args []string) error {
	result := i.Inspect(comando, args)

	if !result.Permitido {
		return fmt.Errorf("comando bloqueado: %s", result.Motivo)
	}

	if result.RequiereAprobacion && result.NivelBloqueo == "warn" {
		return fmt.Errorf("comando requiere aprobación: %s", result.Motivo)
	}

	return nil
}

// ObtenerNivelBloqueo retorna el nivel de bloqueo formateado para usuario
func (r *ResultadoInspeccion) ObtenerNivelBloqueo() string {
	switch r.NivelBloqueo {
	case "block":
		return "🔴 BLOQUEADO"
	case "warn":
		return "🟡 ADVERTENCIA"
	default:
		return "🟢 PERMITIDO"
	}
}

// ToUserSummary retorna un resumen para mostrar al usuario
func (r *ResultadoInspeccion) ToUserSummary() string {
	if r.Permitido && !r.RequiereAprobacion {
		return fmt.Sprintf("%s - %s", r.ObtenerNivelBloqueo(), "Comando seguro para ejecutar")
	}

	lines := []string{r.ObtenerNivelBloqueo()}

	if r.Motivo != "" {
		lines = append(lines, "Razón: "+r.Motivo)
	}

	if len(r.ComandosDetectados) > 0 {
		lines = append(lines, "Comandos detectados: "+strings.Join(r.ComandosDetectados, ", "))
	}

	if len(r.PathsInseguros) > 0 {
		lines = append(lines, "Paths involucrados: "+strings.Join(r.PathsInseguros, ", "))
	}

	if r.SugereReemplazo != "" {
		lines = append(lines, "Sugerencia: "+r.SugereReemplazo)
	}

	return strings.Join(lines, "\n")
}

// GetExitCode retorna el código de salida según el resultado
func (r *ResultadoInspeccion) GetExitCode() int {
	if !r.Permitido {
		return ExitBlocked
	}
	if r.RequiereAprobacion {
		return ExitWarning
	}
	return ExitOK
}

// Códigos de salida semánticos para el inspector
const (
	ExitOK = 0
	ExitWarning = 1
	ExitBlocked = 2
)

// NewMockInspector crea un inspector para testing
func NewMockInspector() *Inspector {
	config := DefaultConfig("/test/workspace")
	insp, _ := NewInspector(config)
	return insp
}

// ParseArgumentCount intenta parsear el número de argumentos
func ParseArgumentCount(args []string) (int, error) {
	count := 0
	for _, arg := range args {
		// Ignorar flags
		if strings.HasPrefix(arg, "-") {
			continue
		}
		count++
	}
	return count, nil
}

// ValidateArgs verifica que los argumentos sean seguros
func (i *Inspector) ValidateArgs(args []string) error {
	count, err := ParseArgumentCount(args)
	if err != nil {
		return err
	}

	// Verificar argumentos excesivos (posible ataque)
	if count > 100 {
		return fmt.Errorf("demasiados argumentos (%d) - posible intento de buffer overflow", count)
	}

	// Verificar argumentos dangerously large
	for _, arg := range args {
		if len(arg) > 10000 {
			return fmt.Errorf("argumento demasiado largo (%d caracteres)", len(arg))
		}
	}

	return nil
}

// ObtenerCmdPath retorna el path del comando en el sistema
func ObtenerCmdPath(comando string) string {
	path, err := exec.LookPath(comando)
	if err != nil {
		return ""
	}
	return path
}

// Verificar_cmd existe verifica si un comando existe en el sistema
func VerificarCmdExiste(comando string) bool {
	path := ObtenerCmdPath(comando)
	return path != ""
}

// ObtenerInfoComando retorna información sobre un comando
func ObtenerInfoComando(comando string) map[string]string {
	info := make(map[string]string)

	// Buscar el comando
	path := ObtenerCmdPath(comando)
	if path != "" {
		info["path"] = path

		// Obtener tamaño
		if stat, err := os.Stat(path); err == nil {
			info["size"] = strconv.FormatInt(stat.Size(), 10)
			info["mode"] = stat.Mode().String()
		}
	}

	return info
}

// SanitizarComando limpia un comando de caracteres peligrosos
func SanitizarComando(comando string, args []string) []string {
	limpio := []string{}

	for _, arg := range args {
		// Filtrar caracteres nulos
		arg = strings.ReplaceAll(arg, string(rune(0)), "")

		// Filtrar newlines多余
		arg = strings.TrimSpace(arg)

		// No agregar argumentos vacíos
		if arg != "" {
			limpio = append(limpio, arg)
		}
	}

	return limpio
}