# Juarvis V4 — Sistema Operativo para Agentes IA

Juarvis V4 es una herramienta de sistema que convierte cualquier carpeta en un entorno de desarrollo profesional gobernado por agentes de IA autónomos.

No es solo un script; es un **motor global** que permite que IAs (Claude, Cursor, OpenCode, etc.) seguían protocolos de ingeniería rigurosos, autogestionen sus herramientas y aprendan de sus errores sin que tú tengas que configurar nada.

---

## 🚀 Guía Rápida

### 1. Instalación (Solo una vez)
```bash
git clone https://github.com/juanjo-zurich/juarvis-v4.git
cd juarvis-v4
make install
```

### 2. Preparar tu Proyecto
```bash
juarvis up
```
*Esto hace todo: init + setup + watch + analiza tu proyecto*

### 3. Ver el Estado
```bash
juarvis dashboard    # Panel visual del ecosistema
juarvis mode        # Ver/cambiar nivel de autonomía (0-4)
```

---

## 🧠 Niveles de Autonomía

| Nivel | Nombre | Comportamiento |
|-------|--------|---------------|
| 0 | Vibe Puro | Ejecución directa, sin planes |
| 1 | Seguro | + snapshot automático antes de cambios grandes |
| 2 | Estructurado | + descomposición de tareas (default) |
| 3 | Semi-SDD | + spec antes de implementar |
| 4 | SDD Completo | Pipeline completo: explore→spec→design→verify |

Cambia el modo con: `juarvis mode 0` a `juarvis mode 4`

---

## ⚡ Comandos Esenciales

```bash
juarvis init           # Inicializar ecosistema
juarvis up            # Onboarding completo (init + setup + watch)
juarvis watch        # Vigilancia activa de archivos
juarvis verify       # Verificar que todo esté sano
juarvis snapshot     # Puntos de restauración
juarvis analyze      # Analizar codebase
juarvis memory      # Servidor MCP de memoria
juarvis mode [n]    # Cambiar nivel de autonomía
juarvis dashboard   # Panel visual del ecosistema
```

---

## 🗃️ Memoria Persistente

Después de `juarvis init`, el agente recuerda lo que aprende:

```
🎯 LO QUE TU AGENTE AHORA SABE DE TU PROYECTO
  📦 Stack: Go, React, PostgreSQL
  📁 Archivos: 143
  📋 Convenciones: conventional commits, ESLint

💡 ANTES: cada sesión → cold context
💡 AHORA: desde el minuto 0 → contexto persistente
```

Usa las herramientas MCP:
- `mem_save(title, content, type)` — guardar observación
- `mem_search(query)` — buscar en memoria
- `mem_context(limit)` — ver sesiones recientes

---

## 🔒 Seguridad

El **Watcher** vigila cada cambio. Si una IA intenta:
- Borrar archivos clave
- Ignorar tests
- Hacer operaciones destructivas

...el watcher lo detectará y te avisará o detendrá la operación.

---

## 🏗️ Arquitectura

- **Binario**: `juarvis` — CLI en Go, zero runtime deps
- **.juar/** en proyecto: memoria, skills, config
- **AGENTS.md**: protocolo activo (se actualiza con `juarvis mode`)
- **Servidor MCP**: memoria local (JSON + índice token en RAM)

---

## ✅ Instalación Verificada

```bash
go build .                    # Build binario
CGO_ENABLED=0 go build .     # Pure Go (sin C)
go test ./...                 # Tests
juarvis verify               # Verificar ecosistema
```

---

## 📦 Nuevas Features

### 🔧 Artifact System

Sistema de artifacts estructurados para representar outputs de tareas de agentes.

| Artifact Type | Descripción | Comando CLI |
|--------------|-------------|-------------|
| **TaskList** | Lista de tareas descompuestas | `juarvis tasks` |
| **ImplementationPlan** | Plan de implementación detallado | `juarvis plan` |
| **Screenshot** | Captura visual de estado de UI | `juarvis screenshot` |
| **TestResult** | Resultado de ejecución de tests | `juarvis test --output json` |
| **VerificationReport** | Reporte de verificación del sistema | `juarvis verify --report` |
| **CodeDiff** | Diff de cambios realizados | `juarvis diff` |

**Ejemplo de uso:**
```bash
# Generar un plan de implementación
juarvis plan --feature "autenticación"

# Verificar cambios con diff
juarvis diff --since "snapshot-001"

# Generar reporte de verificación
juarvis verify --mode strict --report artifacts/verify-report.json
```

---

### 📚 Knowledge Base

Sistema de aprendizaje continuo que extrae y almacena conocimiento del proyecto.

| Tipo de Conocimiento | Descripción | Comando CLI |
|---------------------|-------------|-------------|
| **code_pattern** | Patrones de código detectados | `juarvis kb learn --pattern` |
| **architecture** | Decisiones arquitecturales | `juarvis kb learn --arch` |
| **workflow** | Flujos de trabajo del proyecto | `juarvis kb learn --workflow` |
| **bug_fix** | Correcciones de bugs documentadas | `juarvis kb learn --bug` |
| **decision** | Decisiones técnicas tomadas | `juarvis kb learn --decision` |

**Ejemplo de uso:**
```bash
# Aprender un nuevo patrón de código
juarvis kb learn --pattern "repository-pattern" --files "pkg/db/"

# Buscar conocimiento previo
juarvis kb search "cómo se maneja la autenticación"

# Listar arquitectura conocida
juarvis kb list --type architecture
```

---

### 🔄 Verification Loop

Sistema de verificación automático con múltiples niveles de rigurosidad.

| Modo | Descripción | Casos de Uso |
|------|-------------|--------------|
| **none** | Sin verificación | Prototipado rápido |
| **basic** | Verificación mínima (sintaxis, imports) | Desarrollo rápido |
| **standard** | Verificación estándar (tests, linting) | Desarrollo normal |
| **strict** | Verificación estricta (coverage, security) | Pre-production |
| **xhigh** | Verificación extrema (audit completo) | Producción |

**Verificadores disponibles (6):**

| Verificador | Qué verifica | Comando |
|-------------|--------------|---------|
| **Syntax** | Sintaxis y parseo del código | `juarvis verify --check syntax` |
| **Tests** | Ejecución y resultado de tests | `juarvis verify --check tests` |
| **Lint** | Estilo y convenciones (golint, govet) | `juarvis verify --check lint` |
| **Security** | Vulnerabilidades conocidas | `juarvis verify --check security` |
| **Coverage** | Cobertura de tests | `juarvis verify --check coverage` |
| **deps** | Dependencias actualizadas | `juarvis verify --check deps` |

**Ejemplo de uso:**
```bash
# Verificación básica
juarvis verify --mode basic

# Verificación estricta con todos los checks
juarvis verify --mode strict --checks all

# Verificación de producción
juarvis verify --mode xhigh --report ./verify-report.json
```

---

### ⚡ Planning / Fast Modes

Sistema de detección automática de complejidad para elegir el nivel de planificación apropiado.

| Modo | Descripción | Activación |
|------|-------------|------------|
| **auto** | Detección automática de complejidad | `-` |
| **fast** | Modo rápido para tareas simples | `juarvis fast` |
| **planner** | Modo planificador para tareas complejas | `juarvis plan` |

**Detección de complejidad:**

El sistema analiza automáticamente:
- Número de archivos a modificar
- Dependencias entre componentes
- Presencia de tests existentes
- Cambios en la API pública

**Ejemplo de uso:**
```bash
# Modo automático (selecciona fast o planner según complejidad)
juarvis task "corregir bug en auth" --auto

# Forzar modo rápido
juarvis fast "agregar comment"

# Forzar modo planificación
juarvis planner "refactorizar módulo users"
```

---

### 🤖 Agent Manager

Sistema de gestión de múltiples agentes con capacidad de chaining.

| Capacidad | Descripción | Comando CLI |
|-----------|-------------|-------------|
| **Multi-agente** | Ejecución paralela de múltiples agentes | `juarvis agents run --all` |
| **Chaining** | Ejecución secuencial de agentes (output→input) | `juarvis chain run` |
| **Carga dinámica** | Carga de agentes desde archivos declarativos | `juarvis agents load` |
| **Monitoreo** | Estado y métricas de agentes | `juarvis agents status` |

**Configuración de agentes:**
```yaml
# .juar/agents.yaml
agents:
  - name: go-developer
    role: coder
    enabled: true
  
  - name: code-reviewer
    role: reviewer
    enabled: true
    chain_after: go-developer
```

**Ejemplo de uso:**
```bash
# Ejecutar todos los agentes activos
juarvis agents run --all

# Definir y ejecutar un chain
juarvis chain define "review-flow: explorer → go-developer → code-reviewer"
juarvis chain run "review-flow"

# Monitorear estado de agentes
juarvis agents status
```

---

### 🖥️ Terminal Sandbox

Entorno aislado para ejecución de comandos con múltiples niveles de aislamiento.

| Nivel | Descripción | Aislamiento |
|-------|-------------|--------------|
| **0** | Sin aislamiento | None |
| **1** | Aislamiento básico (timeout) | Timeout, memoria |
| **2** | Aislamiento medio (docker) | Containers读过 |
| **3** | Aislamiento completo (vm) | VM completa |

| Compponente | Descripción | Comando |
|-------------|-------------|---------|
| **sandbox** | Ejecución de comandos en entorno aislado | `juarvis sandbox run` |
| **inspector** | Inspección de estado post-ejecución | `juarvis sandbox inspect` |
| **guard** | Políticas de seguridad de ejecución | `juarvis sandbox guard` |

**Ejemplo de uso:**
```bash
# Ejecutar comando en sandbox nivel 2
juarvis sandbox run --level 2 "npm install"

# Inspeccionar resultado
juarvis sandbox inspect --last

# Configurar políticas de guard
juarvis sandbox guard allow --commands "npm,go,make"
juarvis sandbox guard deny --commands "rm -rf"
```

---

### 📋 AGENTS.md Declarativo

Sistema de generación y gestión de archivos AGENTS.md de forma declarativa.

| Feature | Descripción | Comando CLI |
|---------|-------------|-------------|
| **Generación automática** | Crea AGENTS.md basado en configuración | `juarvis agents gen` |
| **Templates** | Plantillas predefinidas para diferentes flujos | `juarvis agents template` |
| **Validación** | Valida sintaxis del archivo AGENTS.md | `juarvis agents validate` |
| **Migración** | Migra formato antiguo a nuevo | `juarvis agents migrate` |

**Archivo de configuración:**
```yaml
# .juar/agents-config.yaml
version: "2.0"
agents:
  - name: explorer
    description: Exploration agent for codebase analysis
    capabilities:
      - read_files
      - search_code
      - analyze_structure
  
  - name: developer
    description: Main development agent
    capabilities:
      - write_code
      - run_tests
      - git_operations
    requires: [explorer]
```

**Ejemplo de uso:**
```bash
# Generar AGENTS.md desde configuración
juarvis agents gen --config .juar/agents-config.yaml

# Usar template
juarvis agents template --type sdd

# Validar AGENTS.md existente
juarvis agents validate
```

---

## 📋 Resumen de Comandos

| Categoría | Comandos |
|-----------|----------|
| **Ciclo de Vida** | `init`, `up`, `watch`, `verify`, `cleanup` |
| **Memoria** | `memory`, `kb`, `analyze`, `analyze-transcript` |
| **Artifacts** | `plan`, `tasks`, `diff`, `screenshot` |
| **Verificación** | `verify` (con --mode, --checks, --report) |
| **Modos** | `mode`, `fast`, `planner` |
| **Agentes** | `agents` (run, load, status, gen, validate) |
| **Chaining** | `chain` (define, run, list) |
| **Sandbox** | `sandbox` (run, inspect, guard) |
| **Snapshot** | `snapshot` (create, restore, list) |
| **Dashboard** | `dashboard`, `status` |

---

Más info: `juarvis --help`