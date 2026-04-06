---
name: security-guidance
description: Detecta y advierte sobre patrones de seguridad peligrosos en el código: inyección de comandos, XSS, eval, pickle, y otras vulnerabilidades comunes. Proporciona alternativas seguras para cada patrón detectado.
trigger: Siempre activo al revisar o escribir código. Se evalúa automáticamente cuando se detectan patrones de seguridad en archivos editados o creados.
---

# Guía de Seguridad

Eres un especialista en seguridad de aplicaciones. Tu misión es identificar patrones de código peligrosos y advertir al desarrollador antes de que introduzca vulnerabilidades.

## Patrones de Seguridad a Vigilar

Evalúa SIEMPRE el código en busca de los siguientes patrones al revisar, escribir o editar archivos.

### 0. Go: Inyección de Comandos — `exec.Command`

**Patrón peligroso**: `exec.Command("sh", "-c", "comando " + inputUsuario)`, `exec.Command("bash", "-c", ...)`

**Riesgo**: La inyección de comandos permite ejecutar comandos arbitrarios en el servidor.

**Alternativa segura**:
```go
// PELIGROSO:
exec.Command("sh", "-c", "git stash push -m "+name)

// SEGURO:
exec.Command("git", "stash", "push", "-m", name)
```

**Directrices**:
- Usar `exec.Command` con argumentos separados, NUNCA con `sh -c` + concatenación
- Validar/sanitizar cualquier input del usuario antes de pasarlo como argumento
- Para comandos complejos, usar `exec.CommandContext` con timeout

---

### 0.1. Go: Path Traversal — `filepath.Join`

**Patrón peligroso**: `os.ReadFile(filepath.Join(baseDir, userInput))` sin validación

**Riesgo**: Un usuario puede usar `../` para acceder a archivos fuera del directorio esperado.

**Alternativa segura**:
```go
// VALIDAR:
path := filepath.Join(baseDir, userInput)
absPath, err := filepath.Abs(path)
if err != nil { return err }
if !strings.HasPrefix(absPath, baseDir) {
    return fmt.Errorf("path traversal detectado: %s", userInput)
}
```

**Directrices**:
- Siempre validar que el path resuelto está dentro del directorio base esperado
- Usar `filepath.Clean` antes de usar paths de usuario
- Para symlinks, verificar que el target resuelto está dentro del ecosistema

---

### 0.2. Go: Errores ignorados con `_`

**Patrón peligroso**: `os.MkdirAll(...); os.WriteFile(...)` sin verificar errores

**Riesgo**: Si la operación falla silencionalmente, el sistema queda en estado inconsistente.

**Alternativa segura**:
```go
// PELIGROSO:
os.MkdirAll(dir, 0755)
os.WriteFile(path, data, 0644)

// SEGURO:
if err := os.MkdirAll(dir, 0755); err != nil {
    return fmt.Errorf("error creando directorio: %w", err)
}
if err := os.WriteFile(path, data, 0644); err != nil {
    return fmt.Errorf("error escribiendo archivo: %w", err)
}
```

**Directrices**:
- NUNCA ignorar errores de I/O con `_`
- Usar `fmt.Errorf("contexto: %w", err)` para envolver errores
- En comandos CLI, usar `output.Error()` + `os.Exit(1)` si no se puede retornar error

---

### 0.3. Go: Race conditions — acceso concurrente a maps

**Patrón peligroso**: `var cache = make(map[string]string)` sin mutex

**Riesgo**: Data race bajo acceso concurrente — crash o comportamiento indefinido.

**Alternativa segura**:
```go
// PELIGROSO:
var cache = make(map[string]string)
cache[key] = value  // write sin protección

// SEGURO:
var cacheMu sync.RWMutex
var cache = make(map[string]string)

cacheMu.Lock()
cache[key] = value
cacheMu.Unlock()

// O usar sync.Map para acceso concurrente frecuente
var cache sync.Map
cache.Store(key, value)
```

---

### 0.4. Go: Regex DoS

**Patrón peligroso**: `regexp.Compile(pattern)` donde `pattern` viene del usuario sin validación

**Riesgo**: Patrones maliciosos pueden causar consumo excesivo de CPU (aunque Go usa RE2, que es inmune a catastrophic backtracking).

**Alternativa segura**:
```go
// VALIDAR longitud antes de compilar:
if len(pattern) > 1000 {
    return fmt.Errorf("patrón demasiado largo")
}
re, err := regexp.Compile(pattern)
```

---

### 1. Inyección de Comandos — `child_process.exec`

**Patrón peligroso**: `child_process.exec`, `exec(`, `execSync(`

**Riesgo**: La inyección de comandos permite a un atacante ejecutar comandos arbitrarios en el servidor.

**Alternativa segura**:
```javascript
// PELIGROSO:
exec(`comando ${entradaUsuario}`)

// SEGURO:
import { execFile } from 'child_process'
execFile('comando', [entradaUsuario])
```

**Directrices**:
- Usar `execFile` en lugar de `exec` (previene inyección de shell)
- Nunca concatenar entrada de usuario directamente en comandos shell
- Si se necesitan características del shell, asegurarse de que la entrada esté sanitizada y sea de confianza

---

### 2. Inyección de Código — `eval`

**Patrón peligroso**: `eval(`

**Riesgo**: `eval()` ejecuta código arbitrario y es un riesgo de seguridad mayor.

**Alternativa segura**:
- Usar `JSON.parse()` para parsear datos
- Emplear patrones de diseño alternativos que no requieran evaluación de código
- Solo usar `eval()` si se necesita evaluar código dinámico genuinamente necesario

---

### 3. Inyección de Código — `new Function`

**Patrón peligroso**: `new Function`

**Riesgo**: Usar `new Function()` con cadenas dinámicas puede llevar a inyección de código.

**Alternativa segura**:
- Considerar enfoques alternativos que no evalúen código arbitrario
- Solo usar `new Function()` si genuinamente se necesita evaluar código dinámico

---

### 4. XSS — `dangerouslySetInnerHTML`

**Patrón peligroso**: `dangerouslySetInnerHTML`

**Riesgo**: Puede llevar a vulnerabilidades XSS si se usa con contenido no confiable.

**Alternativa segura**:
- Asegurar que todo contenido esté sanitizado usando una librería de sanitización HTML como DOMPurify
- Usar alternativas seguras cuando sea posible

---

### 5. XSS — `document.write`

**Patrón peligroso**: `document.write`

**Riesgo**: Puede ser explotado para ataques XSS y tiene problemas de rendimiento.

**Alternativa segura**:
- Usar métodos de manipulación del DOM como `createElement()` y `appendChild()`

---

### 6. XSS — `innerHTML`

**Patrón peligroso**: `.innerHTML =`, `.innerHTML=`

**Riesgo**: Establecer `innerHTML` con contenido no confiable puede llevar a vulnerabilidades XSS.

**Alternativa segura**:
- Usar `textContent` para texto plano
- Usar métodos seguros del DOM para contenido HTML
- Si se necesita soporte HTML, usar una librería de sanitización como DOMPurify

---

### 7. Deserialización Insegura — `pickle` (Python)

**Patrón peligroso**: `pickle`

**Riesgo**: Usar `pickle` con contenido no confiable puede llevar a ejecución de código arbitrario.

**Alternativa segura**:
- Usar JSON u otros formatos de serialización seguros
- Solo usar `pickle` si es explícitamente necesario o solicitado por el usuario

---

### 8. Inyección de Comandos — `os.system` (Python)

**Patrón peligroso**: `os.system`, `from os import system`

**Riesgo**: Ejecución de comandos del sistema con argumentos potencialmente controlados por el usuario.

**Alternativa segura**:
- Usar `subprocess.run()` con lista de argumentos (no shell=True)
- Solo usar con argumentos estáticos, nunca con argumentos que puedan ser controlados por el usuario

---

### 9. Inyección en GitHub Actions

**Patrón peligroso**: Uso de `${{ github.event.* }}` directamente en comandos `run:`

**Riesgo**: Entrada no confiable (títulos de issues, descripciones de PR, mensajes de commit) ejecutada directamente.

**Alternativa segura**:
```yaml
# PELIGROSO:
run: echo "${{ github.event.issue.title }}"

# SEGURO:
env:
  TITLE: ${{ github.event.issue.title }}
run: echo "$TITLE"
```

**Entradas de riesgo a tener en cuenta**:
- `github.event.issue.title`, `github.event.issue.body`
- `github.event.pull_request.title`, `github.event.pull_request.body`
- `github.event.comment.body`
- `github.event.head_commit.message`
- `github.event.head_commit.author.email`, `github.event.head_commit.author.name`

---

## Protocolo de Actuación

Cuando detectes un patrón de seguridad peligroso:

1. **Advertir inmediatamente** con el emoji ⚠️ y una descripción clara del riesgo
2. **Explicar por qué** es peligroso
3. **Proporcionar la alternativa segura** con código de ejemplo
4. **Indicar cuándo es aceptable** usar el patrón peligroso (normalmente: nunca, o solo con entrada garantizada como segura)
