# Juarvis - Guía de Agentes y Comandos

## 🚀 Introducción

Juarvis es un ecosistema de agentes IA multi-IDE. Los agentes se definen una vez y funcionan en **OpenCode**, **Claude Code**, **AntiGravity** y **Gemini CLI** via symlinks.

---

## 🤖 Agentes Disponibles

### Agentes Primary (Principales)

#### 1. Orchestrator
- **Propósito**: Orquestador principal - coordina sub-agentes
- **Cuándo usarlo**: Cualquier tarea que requiera coordinación
- **Herramientas**: todas
- **Cómo invocarlo**: Es el agente por defecto

#### 2. Plan
- **Propósito**: Análisis read-only, diseño de soluciones
- **Cuándo usarlo**: 
  - "analizar la estructura del proyecto"
  - "cómo implementar X"
  - "diseñar una solución"
- **Herramientas**: read, bash (solo análisis)
- **Restricciones**: NO modifica archivos
- **Invocación manual**: `@plan` o "usa el agente plan"

---

### Agentes Subagent

#### 3. Explorer
- **Propósito**: Exploración de codebases
- **Cuándo usarlo**:
  - "dónde está la función X"
  - "cómo funciona Y"
  - "mapear la estructura"
- **Herramientas**: read, grep, glob, bash (ls)
- **Invocación**: `@explorer` - "encuentra archivos que..."

#### 4. Go Developer
- **Propósito**: Desarrollo Go
- **Cuándo usarlo**: Escribir código Go
- **Herramientas**: todas
- **Invocación**: `@go-developer` - "implementa X"

#### 5. Test Engineer
- **Propósito**: Testing, TDD, coverage
- **Cuándo usarlo**: Escribir tests, coverage
- **Herramientas**: todas
- **Invocación**: `@test-engineer` - "escribe tests para X"

#### 6. Code Reviewer
- **Propósito**: Revisar código
- **Cuándo usarlo**:
  - "revisar cambios"
  - "hay bugs en X"
  - "verificar calidad"
- **Confidence**: Solo reporta ≥75%
- **Herramientas**: read, bash (git)
- **Invocación**: `@code-reviewer` - "revisa este código"

#### 7. Debugger
- **Propósito**: Investigar bugs
- **Cuándo usarlo**:
  - "hay un error"
  - "test fallando"
  - "no funciona"
- **Herramientas**: read, bash
- **Invocación**: `@debugger` - "investiga este error"

#### 8. Security Auditor
- **Propósito**: Auditoría de seguridad
- **Cuándo usarlo**:
  - "auditar seguridad"
  - "hay vulnerabilidades"
- **Detecta**: Command injection, XSS, hardcoded secrets
- **Herramientas**: read, grep
- **Invocación**: `@security-auditor` - "audita este código"

#### 9. Docs Writer
- **Propósito**: Documentación
- **Cuándo usarlo**:
  - "escribir README"
  - "documentar X"
- **Herramientas**: read, write, edit
- **Invocación**: `@docs-writer` - "documenta X"

#### 10. Migrator
- **Propósito**: Migraciones
- **Cuándo usarlo**:
  - "migrar a Go"
  - "actualizar framework"
- **Herramientas**: todas
- **Invocación**: `@migrator` - "migrar X"

#### 11. DevOps
- **Propósito**: CI/CD, Docker
- **Cuándo usarlo**: Despliegues
- **Herramientas**: todas
- **Invocación**: `@devops` - "configura CI/CD"

---

## ⚡ Slash Commands

Juarvis incluye slash commands que pueden ejecutarse desde cualquier IDE.

### /commit
Crea un commit con mensaje generado por IA.

```bash
juarvis commit
# o simplemente
/commit
```

**Qué hace**:
1. Analiza cambios staged/unstaged
2. Examina mensajes recientes
3. Genera mensaje convencional
4. Stages archivos
5. Hace commit

**No hace**: No hace push

**Ejemplo**:
```bash
$ juarvis commit
✅ Commit created: Add: new feature
```

---

### /commit-push-pr
Commit + push + PR en un paso.

```bash
juarvis commit-push-pr
# o
/commit-push-pr
```

**Qué hace**:
1. Crea branch si está en main
2. Commit con mensaje
3. Push a origin
4. Crea PR con `gh pr create`
5. Muestra URL del PR

**Requiere**: `gh` instalado y autenticado

**Ejemplo**:
```bash
$ juarvis commit-push-pr
ℹ️ Creando branch: feature/auto-1
✅ Commit: Add: auth.go and auth_test.go
✅ Push a origin/feature/auto-1
✅ PR creado: https://github.com/user/repo/pull/123
```

---

### /code-review
Review automático con múltiples agentes paralelos.

```bash
juarvis code-review
juarvis code-review --comment
# o
/code-review
/code-review --comment
```

**Qué hace**:
1. Verifica si review es necesario
2. Lanza 4 agentes paralelos:
   - 2x CLAUDE.md compliance
   - 1x Bug detector
   - 1x History analyzer
3. Filtra confidence <80%
4. Reporta solo issues de alta confianza

**Opciones**:
- `--comment`: Post en GitHub PR

**Ejemplo**:
```bash
$ juarvis code-review
ℹ️ Ejecutando review con 4 agentes...
✅ No se encontraron issues de alta confianza
```

---

### /clean-gone
Limpia branches locales eliminadas en remote.

```bash
juarvis clean-gone
# o
/clean-gone
```

**Qué hace**:
1. Lista branches [gone]
2. Elimina worktrees asociados
3. Borra branches stale
4. Reporta limpieza

**Ejemplo**:
```bash
$ juarvis clean-gone
ℹ️ Actualizando remote tracking...
⚠️ Encontradas 2 branches [gone]: old-feature, deprecated
ℹ️ Eliminando: old-feature
ℹ️ Eliminando: deprecated
✅ Eliminado 2 branches

Workspace limpio ✅
```

---

## 🔄 Uso Automático de Agentes

### Cómo el Orquestador Delega

El `orchestrator` detecta cuándo delegar automáticamente:

```
Usuario: "dónde está la función auth?"
    ↓
Orquestador → "Usa explorer para encontrar eso"
    ↓
Explorer retorna: "en pkg/auth/auth.go:L45"
```

### Table de Delegación Automática

| Solicitud | → Agente |
|-----------|----------|
| "dónde está X" | → explorer |
| "cómo funciona Y" | → explorer |
| "analizar estructura" | → plan |
| "escribir código" | → go-developer |
| "escribir tests" | → test-engineer |
| "revisar código" | → code-reviewer |
| "hay un error" | → debugger |
| "auditar seguridad" | → security-auditor |
| "documentar" | → docs-writer |
| "migrar" | → migrator |
| "despliegue" | → devops |

---

## 📁 Estructura de Archivos

```
.juar/                           # Raíz universal
├── agents/                       # Agentes (para todos los IDEs)
│   ├── orchestrator.md
│   ├── plan.md
│   ├── explorer.md
│   └── ...
├── commands/                     # Slash commands
│   ├── commit.md
│   └── code-review.md
└── context/                    # Context files
    └── CONTEXT.md

.opencode/                       # OpenCode específico
├── agents/                      # Agentes OpenCode
│   └── *.md
├── commands/                    # Commands OpenCode
└── skills/

.claude/                         # Claude Code
└── agents/ → ../.opencode/agents/  # Symlinks

.gemini/                         # Gemini CLI  
└── agents/ → ../.opencode/agents/  # Symlinks
```

---

## ⚙️ Configuración

### agent-settings.json (OpenCode)

```json
{
  "agent": {
    "orchestrator": { ... },
    "plan": { 
      "mode": "primary",
      "prompt": "{file:.opencode/agents/plan.md}",
      "tools": { "read": true, "write": false }
    },
    "explorer": {
      "mode": "subagent",
      "prompt": "{file:.opencode/agents/explorer.md}",
      "tools": { "read": true, "write": false }
    }
  }
}
```

---

## 🪝 Comandos de Hooks

### hooks list
Lista todas las reglas de hooks cargadas.

```bash
juarvis hooks list
# Muestra tabla con NOMBRE, EVENTO, ACCIÓN, ESTADO
```

### hooks create
Crea una nueva regla hook.

```bash
juarvis hooks create --name mi-regla --pattern "rm -rf" --action block
```

**Opciones**:
- `--name`: Nombre de la regla
- `--pattern`: Patrón regex a buscar
- `--action`: Acción (warn, block)
- `--event`: Evento (bash, file, stop, all)
- `--disabled`: Crear deshabilitada

### hooks enable/disable
Habilita o deshabilita una regla.

```bash
juarvis hooks enable mi-regla
juarvis hooks disable mi-regla
```

---

## 💾 Comandos de Sesión

### session save
Guarda el estado actual de la sesión.

```bash
juarvis session save                # Auto-nombre con timestamp
juarvis session save mi-sesión      # Nombre personalizado
```

Guarda:
- Estado git (staged/unstaged)
- Diff completo
- Metadata (branch, timestamp)

### session list
Lista sesiones guardadas.

```bash
juarvis session list
# Muestra NOMBRE, FECHA, BRANCH
```

### session resume
Restaura una sesión guardada.

```bash
juarvis session resume mi-sesión
```

### session export
Exporta sesión a JSON.

```bash
juarvis session export mi-sesión
```

---

## 🧪 Verificación

```bash
# Verificar agentes disponibles
ls .opencode/agents/

# Verificar symlinks
ls -la .claude/agents/

# Verificar build
juarvis verify
```

---

## 📋 Checklist de Uso

| Necesidad | Acción | Agente/Comando |
|----------|--------|---------------|
| Encontrar archivo | "dónde está X" | `@explorer` |
| Entender código | "cómo funciona" | `@explorer` |
| Planificar | "cómo implementar" | `@plan` |
| Escribir código | "implementa X" | `@go-developer` |
| Escribir tests | "escribe tests" | `@test-engineer` |
| Revisar código | "revisa cambios" | `@code-reviewer` / `/code-review` |
| Investigar error | "hay error" | `@debugger` |
| Auditar seguridad | "audita" | `@security-auditor` |
| Documentar | "escribe docs" | `@docs-writer` |
| Hacer commit | "haz commit" | `/commit` |
| Commit+PR | "crea PR" | `/commit-push-pr` |
| Limpiar branches | "limpia" | `/clean-gone` |