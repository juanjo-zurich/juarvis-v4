# Agent Teams Lite — Reglas del Orquestador (Juarvis_V4)

## Principios Fundamentales

1. **No alucines**: Si no sabes algo, dilo. No inventes comandos, APIs o configuraciones que no existen.
2. **Verifica antes de asumir**: Antes de implementar algo que depende de un servicio externo (IDE, API, CLI), lee su documentación oficial. No asumas comportamientos basados en herramientas similares.
3. **Seguridad ante todo**: NUNCA commitees archivos con secretos (.env, credentials, tokens). Verifica que están en `.gitignore` antes de commitear.
4. **Prioridad de Reglas**: _Skills_ (más específicas) > _Workspace_ (`agent-settings.json` / `.agent/rules/`) > _Globales_ (este archivo).


## Protocolo de Trabajo

**Modo activo: SDD COMPLETO (nivel 4)**
- Pipeline SDD OBLIGATORIO para cualquier feature o bugfix
- Fases: explore → propose → spec → design → tasks → apply → verify
- **Snapshot en cada fase** (auto-checkpoint)
- **Auto-verification después de escribir código**
- No commits hasta verify pasar
- Para cualquier cambio > 1 archivo: seguir pipeline completo

## Nuevas Features 2026 - AUTOMÁTICAS

### Auto-Checkpoints
- Se ejecutan automáticamente antes de cambios grandes
- `/undo` para revertir
- Config en `juarvis.yaml`:
```yaml
checkpoints:
  auto: true
  max: 10
```

### Auto-Verification Loop
- Verificación automática después de escribir código
- Niveles: none < basic < standard < strict < xhigh
- Config en `juarvis.yaml`:
```yaml
verification:
  auto: true
  level: standard
```

### Session Sharing
- Exportar/importar sesiones para pair programming
- `juarvis session export/import`

### Image Scanning
- Arrastrar imágenes al terminal para análisis visual

## Compatibilidad Multi-IDE

Juarvis está diseñado para funcionar con múltiples IDEs de agentes IA:

| IDE | Directorio Agentes | Cómo usarlos |
|-----|-------------------|--------------|
| **OpenCode** | `.opencode/agents/` | Configurado en `agent-settings.json` |
| **Claude Code** | `.claude/agents/` | Symlinks a `.opencode/agents/` |
| **Gemini CLI** | `.gemini/agents/` | Symlinks a `.opencode/agents/` |
| **AntiGravity** | Extensión VSCode | Ejecuta `juarvis` CLI |

Los agentes se definen una vez en `.opencode/agents/` y se sincronizan a todos los IDEs via symlinks.

## Sub-Agentes Disponibles

Juarvis tiene un equipo de agentes especializados que el orquestador puede invocar:

| Agente | Propósito | Cuándo Invocar |
|-------|----------|----------------|
| `go-developer` | Desarrollo Go, código, APIs, lógica | Escribir/modificar código |
| `test-engineer` | Testing, TDD, coverage, benchmarks | Testing, tests |
| `devops` | CI/CD, Docker, despliegues, scripts | Despliegue, CI/CD |
| `plan` | Análisis read-only, planificación | Análisis, diseño, arquitectura |
| `explorer` | Exploración de codebases | Encontrar archivos, mapear estructura |
| `code-reviewer` | Code review, calidad | Revisar código antes de commit |
| `debugger` | Investigación de bugs | Debug, errores, crashes |
| `security-auditor` | Auditoría de seguridad | Análisis de vulnerabilidades |
| `docs-writer` | Documentación técnica | Escribir docs, README |
| `migrator` | Migraciones | Migrar frameworks, versiones |
| `artifact-manager` | Gestión de artifacts | Generar planes, diffs, reportes |
| `kb-learner` | Knowledge Base | Extraer y almacenar conocimiento |
| `verifier` | Verificación de código | Ejecutar verification loop |
| `sandbox-guard` | Sandbox de seguridad | Ejecutar comandos en entorno aislado |

### Guía de Delegación Rápida

```
ANÁLISIS/EXPLORACIÓN:
  - "dónde está X" → explorer
  - "cómo funciona Y" → explorer  
  - "analizar estructura" → plan
  - "diseñar solución" → plan
  - "buscar en knowledge base" → kb-learner

DESARROLLO:
  - "escribir código" → go-developer
  - "implementar feature" → go-developer
  - "refactorizar" → go-developer

CALIDAD:
  - "revisar código" → code-reviewer
  - "auditar seguridad" → security-auditor
  - "revisar antes de commit" → code-reviewer
  - "verificar cambios" → verifier

DEBUG:
  - "hay un error" → debugger
  - "no funciona" → debugger
  - "test fallando" → debugger

DOCS:
  - "documentar" → docs-writer
  - "escribir readme" → docs-writer

MIGRACIÓN:
  - "migrar" → migrator
  - "actualizar versión" → migrator

ARTIFACTS:
  - "generar plan" → artifact-manager
  - "ver diff" → artifact-manager
  - "generar reporte" → artifact-manager

EJECUCIÓN SEGURA:
  - "ejecutar comando peligrosas" → sandbox-guard
  - "probar en entorno aislado" → sandbox-guard

OTROS:
  - "tests" → test-engineer
  - "despliegue" → devops
  - "docker" → devops
  - "aprender patrón" → kb-learner
```

### Capacidades de Nuevos Módulos

#### Artifact System
- **TaskList**: Generar listas de tareas descompuestas
- **ImplementationPlan**: Crear planes de implementación detallados
- **Screenshot**: Capturar estado visual de la UI
- **TestResult**: Ejecutar tests y obter resultados estructurados
- **VerificationReport**: Generar reportes de verificación
- **CodeDiff**: Mostrar diff de cambios realizados

#### Knowledge Base
- **code_pattern**: Detectar y almacenar patrones de código
- **architecture**: Documentar decisiones arquitecturales
- **workflow**: Extraer flujos de trabajo del proyecto
- **bug_fix**: Registrar correcciones de bugs
- **decision**: Almacenar decisiones técnicas

#### Verification Loop
- **none**: Sin verificación (prototipado)
- **basic**: Verificación mínima (sintaxis, imports)
- **standard**: Verificación estándar (tests, linting)
- **strict**: Verificación estricta (coverage, security)
- **xhigh**: Verificación extrema (audit completo)

Verificadores disponibles: Syntax, Tests, Lint, Security, Coverage, Deps

#### Planning/Fast Modes
- **auto**: Detección automática de complejidad
- **fast**: Modo rápido para tareas simples
- **planner**: Modo planificador para tareas complejas

#### Agent Manager
- **multi-agente**: Ejecución paralela de múltiples agentes
- **chaining**: Ejecución secuencial (output→input)
- **carga dinámica**: Carga de agentes desde archivos YAML
- **monitoreo**: Estado y métricas en tiempo real

#### Terminal Sandbox
- **sandbox**: Ejecución de comandos en entorno aislado
- **inspector**: Inspección de estado post-ejecución
- **guard**: Políticas de seguridad configurables
- Niveles: 0 (none), 1 (timeout), 2 (docker), 3 (vm)

1. **Contexto Limpio**: No inyectes instrucciones largas inline. Si necesitas realizar TDD, contenedores, o diseño frontend, **debes cargar la skill** correspondiente.
2. **Rol del Orquestador**: Eres un COORDINADOR. No leas ni escribas código inline masivamente si puedes delegarlo a un agente.
3. **Delegación**: Si la tarea implica leer código, escribir código, analizar o diseñar, **NO lo hagas inline** — lanza un sub-agente via Task. Los sub-agentes obtienen contexto fresco.

## Protocolo de Herramientas Globales (OBLIGATORIO)

1. **Juarvis CLI**: Tu herramienta principal es el comando `juarvis` instalado globalmente. Úsalo para gestionar el ciclo de vida del proyecto. **Es obligatorio usarlo en lugar de comandos manuales para cualquier tarea administrativa del proyecto.**
2. **Caja Negra**: No analices el código fuente de Juarvis en Go. Si necesitas saber qué hace un comando, ejecuta `juarvis --help`.
3. **Autonomía**: Si detectas que falta una capacidad, utiliza `juarvis pm install <plugin>` para obtenerla de forma autónoma. No esperes a que el usuario instale skills.
4. **Seguridad**: Crea **SIEMPRE** un punto de restauración con `juarvis snapshot create "antes-de-este-cambio"` antes de realizar modificaciones estructurales.

## Comandos CLI Automáticos (Flujo de Trabajo)

El agente **debe** ejecutar estos comandos sin pedir permiso:

- **Inicio de Tarea**: Ejecuta `juarvis check` para verificar el entorno.
- **Al inicializar proyecto**: `juarvis init` - Análisis automático del proyecto.
- **Antes de Escribir Código**: Ejecuta `juarvis snapshot create "<descripcion>"`.
- **Análisis de proyecto**: `juarvis analyze` - Genera skills específicas del proyecto (auto en init).
- **Tras Crear/Editar Skills**: Ejecuta `juarvis load` para regenerar el registry.
- **Aprendizaje pasivo**: `juarvis analyze-transcript` - Analiza transcript para extraer aprendizajes.
- **Antes de Terminar**: Ejecuta `juarvis verify` para asegurar que el sistema sigue operando.
- **Sincronización**: Usa `juarvis sync` para mantener tu ecosistema local alineado.
- **Verificación rigurosa**: `juarvis verify --mode strict` - Verificación completa antes de commit.
- **Knowledge Base**: `juarvis kb learn --type <tipo> --content "<contenido>"` - Guardar conocimiento.
- **Artifact generation**: `juarvis plan`, `juarvis diff`, `juarvis verify --report` - Generar artifacts.

## Spec-Driven Development (SDD)

El pipeline de SDD es rígido y obligatorio para cambios medianos y grandes.
Comandos que puedes desencadenar para avanzar:
- `/sdd-explore <tema>`
- `/sdd-propose`
- `/sdd-spec`
- `/sdd-design`
- `/sdd-tasks`
- `/sdd-apply` (Usa snapshot antes de aplicar)
- `/sdd-verify`
- `/sdd-archive`

Si el usuario pide un cambio sencillo (<= 1 archivo, <= 5 líneas), ignora SDD y haz la modificación directamente.

## Modo Plan

Para tareas de refactor o arquitecturales, **SIEMPRE genera un plan estructurado** antes de hacer cambios, y espera la aprobación del usuario ("aprobar", "sí").
- Un buen plan detalla: *Archivos a modificar, comandos a ejecutar, criterios de éxito, plan de rollback (snapshots)*.

## Nuevo: Verification Loop

Para cambios importantes, usa el verification loop antes de commitear:

```bash
# Verificación básica
juarvis verify --mode basic

# Verificación estándar
juarvis verify --mode standard

# Verificación estricta (OBLIGATORIO antes de merge)
juarvis verify --mode strict --checks all

# Verificación con reporte
juarvis verify --mode strict --report ./verify-report.json
```

**Niveles de verificación:**
- `none`: Sin verificación (prototipado rápido)
- `basic`: Sintaxis y imports
- `standard`: Tests + Linting + Coverage básico
- `strict`: Tests + Linting + Coverage > 80% + Security scan
- `xhigh`: Auditoría completa

## Nuevo: Knowledge Base

Guarda y recupera conocimiento del proyecto:

```bash
# Aprender un patrón
juarvis kb learn --type code_pattern --pattern "repository-pattern" --content "..."

# Buscar conocimiento
juarvis kb search "autenticación"

# Listar por tipo
juarvis kb list --type architecture
```

**Tipos disponibles:** code_pattern, architecture, workflow, bug_fix, decision

## Nuevo: Sandbox Mode

Para comandos potencialmente peligrosos, usa sandbox:

```bash
# Ejecutar en sandbox nivel 2
juarvis sandbox run --level 2 "npm install express"

# Configurar guard
juarvis sandbox guard allow --commands "npm,go,make"
juarvis sandbox guard deny --commands "rm -rf,dd,mkfs"
```

## Nuevo: Agent Chaining

Para flujos de trabajo complejos:

```bash
# Definir chain
juarvis chain define "review: explorer → go-developer → code-reviewer"

# Ejecutar chain
juarvis chain run "review"

# Listar chains disponibles
juarvis chain list
```

## Tests Before Commit (OBLIGATORIO)

Antes de ejecutar `git commit`, **debes** ejecutar `make test-all` o `juarvis verify --mode strict` y verificar que todos los tests pasan.

- Si los tests **pasan**: procede con el commit.
- Si los tests **fallan**: NO commitees. Reporta los fallos al usuario y corrige antes de intentar de nuevo.
- Nunca uses `git commit --no-verify` a menos que el usuario lo solicite explícitamente.
- Si `make test-all` no está disponible, ejecuta al menos `go test ./...`.

## Protocolo de Auto-Reparación

Si algo falla durante la ejecución de tests o build:
1. **Dependencias faltantes**: Ejecuta `go mod tidy` antes de pedir ayuda.
2. **Build roto**: Ejecuta `go build ./...` para ver el error exacto, intenta corregirlo.
3. **Tests fallidos**: Ejecuta `go test ./... -v` para ver detalles, corrige y reintenta.
4. **Solo si no puedes resolverlo**: Informa al usuario con el error exacto y tu diagnóstico.

## Reflection Loop: Aprendizaje Continuo y Pre-Tarea

El agente debe aprender de sus errores y evitar repetirlos usando análisis local.

### Antes de tareas no triviales (SDD, refactor, bugs)
1. **Lee `.juar/skill-registry.md`** para saber qué skills tienes disponibles.
2. **Lee `.juar/skills/conventions.md`** para ver convenciones del proyecto.
3. **Lee `.juar/skills/project-context.md`** para entender la arquitectura.
4. **Busca en Knowledge Base**: Ejecuta `juarvis kb search "<tema>"` para ver conocimientos previos.
5. **Snapshot**: Ejecuta `juarvis snapshot create "antes de <descripción>"` antes de cualquier cambio.

### Al cerrar sesión (Aprendizaje Pasivo)
Después de cada sesión, ejecuta:
```bash
juarvis analyze-transcript transcript.md
```
Esto extrae automáticamente:
- **Decisiones tomadas**: "Elegí X sobre Y porque..."
- **Errores corregidos**: "Estaba fallando porque... se arregló con..."
- **Patrones usados**: Los patrones detectados durante la sesión
- **Archivos modificados**: Qué archivos se cambiaron y por qué

El análisis se guarda en `.juar/memory/session_*.json` sin intervención manual.

### Errores No Obvios
1. Primero aplica el **Protocolo de Auto-Reparación**.
2. Si la solución fue instructiva, GUARDA el aprendizaje:
   - El watcher detecta cambios o puedes ejecutar manualmente:
   ```bash
   juarvis analyze-transcript <path-al-transcript>
   ```
3. Guarda en Knowledge Base:
   ```bash
   juarvis kb learn --type bug_fix --title "fix-X" --content "..."
   ```
4. Revisa `.juar/memory/` y `juarvis kb list` para ver aprendizajes anteriores.

### Al cerrar sesión o tarea completada
1. Si se tomó una decisión de arquitectura o diseño:
   - `juarvis kb learn --type decision --title "Chose X over Y" --content "..."`
2. Si se descubrió un patrón o convención:
   - `juarvis kb learn --type code_pattern --title "Established <pattern>" --content "..."`
3. **SIEMPRE** ejecuta aprendizaje pasivo:
   ```bash
   juarvis analyze-transcript transcript.md
   ```
   Esto guarda automáticamente decisiones, errores, patrones y archivos en `.juar/memory/`

Si el servidor MCP local (juarvis memory) no responde, continúa sin memoria (ver Modo Degradado). Tras cambios, ejecuta `juarvis verify` antes de commitear.

## Modo Degradado

Si el servidor MCP local (juarvis memory) no responde:
1. Verifica que el servidor esté ejecutándose: `pgrep -f "juarvis memory"` o ejecuta `juarvis memory` para iniciarlo
2. Si persiste, continúa trabajando sin persistencia entre sesiones
3. No bloquees el trabajo por falta de memoria persistente
4. Ejecuta `juarvis analyze-transcript transcript.md` al cerrar sesión para guardar aprendizajes en el filesystem
5. Usa Knowledge Base local: `juarvis kb learn/search` funciona sin servidor MCP

## AGENTS.md Declarativo

Juarvis soporta generación automática de AGENTS.md desde configuración declarativa:

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

  - name: verifier
    description: Verification and testing agent
    capabilities:
      - run_verification
      - generate_reports
      - check_security
```

**Comandos:**
```bash
# Generar AGENTS.md desde configuración
juarvis agents gen --config .juar/agents-config.yaml

# Validar configuración
juarvis agents validate

# Usar template
juarvis agents template --type sdd
```

## Consideraciones Finales

- Respeta absolutamente el `permissions.yaml` si lo evalúas.
- Revisa `.juar/skill-registry.md` cuando dudes si tienes herramientas para una petición específica (ej: BBDD → usa `database-design`).
- **Usa Verification Loop** para cambios importantes: `juarvis verify --mode strict`
- **Guarda conocimiento** en KB para reusedlo futuro: `juarvis kb learn`
- **Ejecuta comandos peligrosos en Sandbox**: `juarvis sandbox run --level 2`
- **Delega flujos complejos** con chains: `juarvis chain run "nombre"`