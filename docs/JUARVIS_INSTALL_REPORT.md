# INFORME: Qué hace Juarvis y qué instala

## 1. PROCESO DE INSTALACIÓN GLOBAL

### `make install` o `scripts/install.sh`

| Paso | Qué hace | Dónde (archivo:línea) |
|------|-----------|------------------------|
| 1. Detecta OS/Arch | macOS (amd64/arm64), Linux, Windows | `scripts/install.sh:8-28` |
| 2. Determina directorio | `~/.local/bin/`, `/usr/local/bin/`, o `~/bin/` | `scripts/install.sh:30-44` |
| 3. **Si hay binario local** `juarvis` | Lo copia a `$install_dir/juarvis` | `scripts/install.sh:103-107` |
| 4. **Si NO hay binario** | Descarga desde GitHub Releases: `juarvis-${os}-${arch}` | `scripts/install.sh:110-123` |
| 5. Hace ejecutable | `chmod +x $install_dir/juarvis` | `scripts/install.sh:122,125` |
| 6. Añade al PATH | Actualiza `~/.bashrc`, `~/.zshrc`, o `~/.config/fish/` | `scripts/install.sh:48-76` |

**Resultado**: Binario `juarvis` instalado globalmente y disponible en terminal.

---

## 2. QUÉ CREA `juarvis init` EN UN PROYECTO

### `juarvis init [path]` (llama a `pkg/init/init.go:32-147`)

| Paso | Qué crea | Dónde (archivo:línea) |
|------|-----------|------------------------|
| 1. Extrae **assets embebidos** | Lee `pkg/assets/data/*` y copia al path destino | `init.go:60-96` |
| 2. Crea **`.juar/`** | Directorio raíz del ecosistema | `init.go:100` |
| 3. Crea **`.juar/skills/`** | Skills del usuario y symlinks | `init.go:111` |
| 4. Crea **`.juar/rules/`** | Reglas del agente | `init.go:114` |
| 5. Extrae **`marketplace.json`** | Lista de 26 plugins disponibles | Desde `pkg/assets/data/marketplace.json` |
| 6. Extrae **`AGENTS.md`** | Constitución del proyecto (reglas orquestador) | Desde `pkg/assets/data/AGENTS.md` |
| 7. Extrae **`agent-settings.json`** | Config de agentes (orquestador, go-developer, etc.) | Desde `pkg/assets/data/agent-settings.json` |
| 8. Extrae **`permissions.yaml`** | Sistema de permisos | Desde `pkg/assets/data/permissions.yaml` |
| 9. Crea **`plugins/`** + **`.juarvis-plugin/`** | Por cada plugin en marketplace.json | `init.go:118-146` |
| 10. Crea **`plugins/<p>/skills/`** + **`plugin.json`** | Manifiesto de cada plugin | `init.go:130-139` |
| 11. Crea **`plugins/<p>/.juarvis-plugin/enabled`** | Habilita el plugin | `init.go:141-144` |
| 12. Instala **pre-commit hook** | Si existe `.git/`, copia hook de `hooks/` | `init.go:149-159` |

**Resultado**: Ecosistema Juarvis inicializado con plugins, skills, y configuración.

---

## 3. QUÉ COMANDOS PRINCIPALES HAY

### Registrados en `cmd/root.go:24-121`

| Comando | Descripción | Archivo |
|---------|-------------|---------|
| `juarvis init` | Inicializa ecosistema | `cmd/init.go` |
| `juarvis load` | Ejecuta cargador: crea symlinks de skills, genera `skill-registry.md` | `cmd/loader.go` → `pkg/loader/loader.go` |
| `juarvis check` | Health check completo | `cmd/verify.go` |
| `juarvis verify` | Verificación de salud (build, test, plugins, CLI) | `pkg/verify/verify.go` |
| `juarvis pm` | Gestor de paquetes (Marketplace) | `cmd/pm.go` |
| `juarvis hooks` | Gestión de hooks (list, create, enable, disable) | `cmd/hooks.go` |
| `juarvis session` | Gestión de sesiones (save, list, resume, export) | `cmd/session.go` |
| `juarvis commit` | Commit con mensaje IA | `cmd/commit.go` |
| `juarvis commit-push-pr` | Commit + push + PR workflow | `cmd/commit-push-pr.go` |
| `juarvis code-review` | Review automático con multi-agentes | `cmd/code-review.go` |
| `juarvis clean-gone` | Limpia branches stale | `cmd/clean-gone.go` |
| `juarvis memory` | Servidor MCP de memoria persistente | `cmd/memory.go` |
| `juarvis schedule` | Gestión de jobs programados | `cmd/schedule.go` |
| `juarvis snapshot` | Snapshots de seguridad para SDD | `cmd/snapshot.go` |
| `juarvis sync` | Sincronización con cloud (Gist) | `cmd/sync.go` |
| `juarvis watch` | Vigila cambios y ejecuta reglas | `cmd/watch.go` |
| `juarvis analyze` | Analiza codebase y genera skills | `cmd/analyze.go` |
| `juarvis analyze-transcript` | Analiza transcript y extrae learnings | `cmd/analyze_transcript.go` |
| `juarvis mode` | Muestra/cambia nivel de autonomía | `cmd/mode.go` |
| `juarvis doctor` | Diagnóstico completo del ecosistema | `cmd/doctor.go` |
| `juarvis setup` | Distribuye configuraciones a IDEs | `cmd/setup.go` |
| `juarvis vibe` | Vibe check (salud creativa) | `cmd/vibe.go` |
| `juarvis ralph` | Ralph Wiggum loop engine | `cmd/ralph.go` |
| `juarvis guard` | Permission guard | `cmd/guard.go` |
| `juarvis hookify` | Hookify engine (pre-tool-use, post-tool-use) | `cmd/hookify.go` |

---

## 4. QUÉ PLUGINS/SKILLS SE CARGAN

### `juarvis load` → `pkg/loader/loader.go:16-229`

| Paso | Qué hace | Dónde (archivo:línea) |
|------|-----------|------------------------|
| 1. Lee **`plugins/`** del proyecto | Lista directorios de plugins | `loader.go:33-36` |
| 2. Verifica si **ya están actualizados** | Si `skills/` y `skill-registry.md` son válidos → skip | `loader.go:38-42` |
| 3. Crea **directorio temporal** en mismo filesystem (atomicidad) | `loader.go:44-48` |
| 4. Por cada plugin **habilitado** (lee `.juarvis-plugin/enabled`) | Verifica `enabled != "false"` | `loader.go:67-70` |
| 5. Lee **`plugin.json`** del manifiesto | Obtiene nombre, versión del plugin | `loader.go:74-82` |
| 6. Lee **`<plugin>/skills/`** | Lista carpetas de skills del plugin | `loader.go:84-86` |
| 7. **Crea symlinks** de cada skill: `skills/<skill>` → `../plugins/<plugin>/skills/<skill>` | Verifica seguridad: no salen del ecosistema | `loader.go:90-112` |
| 8. **Carga skills de usuario** de `.agent/skills/` | Crea symlinks también | `loader.go:117-158` |
| 9. **Genera `.juar/skill-registry.md`** | Tabla con: `| Skill \| Plugin \| Source \| Status |` | `loader.go:168-177` |
| 10. **Reemplazo atómico**: elimina `skills/` viejo, renombra `tmpDir` → `skills/` | `loader.go:160-166` |

### Assets Embebidos en `pkg/assets/data/`

| Archivo | Propósito |
|--------|----------|
| `marketplace.json` | 26 plugins disponibles con nombre, versión, categoría |
| `AGENTS.md` | Constitución del proyecto (reglas orquestador) |
| `agent-settings.json` | Config de 14 agentes (orquestador, go-developer, etc.) |
| `permissions.yaml` | Sistema de permisos (bash: allow/deny, format strings) |
| `hooks/` | Pre-commit hooks para seguridad |

---

## 5. QUÉ EJECUTAN LOS AGENTES CUANDO TRABAJAN

### Agentes trabajan en **PROYECTO DEL USUARIO** (donde está instalado Juarvis, NO en código fuente de Juarvis)

| Agente | Qué lee | Qué escribe | Qué ejecuta |
|--------|----------|-------------|--------------|
| **orchestrator** (primary) | `AGENTS.md`, `agent-settings.json` | Coordina, no escribe | `juarvis verify`, `juarvis snapshot`, `juarvis commit` |
| **go-developer** (subagent) | Código del proyecto (detecta lenguaje: Go, React, Python) | Archivos del proyecto | `juarvis verify`, `go test`/`npm test`/`pytest` según proyecto |
| **test-engineer** (subagent) | Tests del proyecto | Tests del proyecto | `go test ./...`/`npm test`/`pytest` |
| **code-reviewer** (subagent) | Código del proyecto | No escribe (solo reporta) | `juarvis code-review`, `juarvis verify` |
| **debugger** (subagent) | Código, logs, stack traces | No escribe (solo diagnostica) | `go build`/`npm run build` según proyecto |
| **security-auditor** (subagent) | Código del proyecto | No escribe (solo reporta) | `grep` para patrones inseguros |
| **docs-writer** (subagent) | Código del proyecto | `README.md`, `docs/` del proyecto | `juarvis verify` |
| **frontend-designer** (subagent) | HTML/CSS/JS del proyecto | Crea UIs con aesthetics (evita "AI slop") | Detecta framework (React, Vue, vanilla) |
| **explorer** (subagent) | Estructura del proyecto | No escribe | `glob`, `grep`, `ls` |
| **migrator** (subagent) | Código + configs (package.json, go.mod, etc.) | Migra código entre versiones/frameworks | Comandos de migración según tecnología |
| **devops** (subagent) | CI/CD, Docker, scripts del proyecto | Configura CI/CD | `juarvis verify`, detecta herramientas del proyecto |

### Comandos Juarvis que auto-ejecutan los agentes:

| Comando | Cuándo lo ejecutan |
|---------|-------------------|
| `juarvis verify` | Siempre que agente termine una tarea |
| `juarvis commit` | Antes de commit (si hay cambios) |
| `juarvis code-review` | Antes de commit (revisión de calidad) |
| `juarvis session save` | Antes de cambios estructurales |
| `go test ./...` / `npm test` / `pytest` | Según lenguaje del proyecto |

---

## 6. ARCHIVOS .GO RELEVANTES (ANÁLISIS LÍNEA A LÍNEA)

### `cmd/root.go` (129 líneas)
- **Líneas 24-37**: Define comando principal `juarvis`, flags `--root`, `--json`, `--debug`
- **Líneas 38-54**: `Run`: Detecta ecosistema con `root.GetRoot()`, muestra ayuda si no hay
- **Líneas 63-103**: Verifica componentes: `marketplace.json`, `AGENTS.md`, `permissions.yaml`, `agent-settings.json`, `plugins/`, `skills/`

### `cmd/init.go` (39 líneas)
- **Líneas 11-22**: Define `juarvis init [path]`, llama a `initpkg.RunInit(path)`
- **Líneas 23-34**: Ejecuta `initpkg.RunInit()` que crea estructura base

### `pkg/init/init.go` (285 líneas)
- **Líneas 32-47**: `RunInit()`: Extrae assets embebidos, crea `.juar/`, verifica si ya existe ecosistema
- **Líneas 60-96**: Copia assets: `marketplace.json`, `AGENTS.md`, `permissions.yaml`, `agent-settings.json`, `hooks/`
- **Líneas 98-115**: Crea `.juar/`, `.agent/skills/`, `.agent/rules/`
- **Líneas 118-146**: Por cada plugin en marketplace: crea `plugins/<p>/`, `plugin.json`, `enabled`
- **Líneas 149-159**: Instala pre-commit hook si hay `.git/`

### `cmd/loader.go` (24 líneas)
- **Líneas 10-20**: Define `juarvis load`, llama a `loader.RunLoader("")`

### `pkg/loader/loader.go` (229 líneas)
- **Líneas 16-24**: `RunLoader()`: Lee plugins, verifica si skills ya están actualizados
- **Líneas 44-62**: Crea directorio temporal, lee manifests de plugins
- **Líneas 84-112**: Por cada skill de plugin: crea symlink en `tmpDir/`
- **Líneas 117-158**: Carga skills de usuario de `.agent/skills/`
- **Líneas 160-177**: Genera `.juar/skill-registry.md` con tabla
- **Líneas 181-228**: `areSkillsValid()`: Verifica symlinks válidos

### `Makefile` (75 líneas)
- **Línea 11**: `build`: `go build -ldflags ... -o juarvis .`
- **Líneas 53-55**: `install`: Ejecuta `scripts/install.sh`
- **Líneas 57-64**: `install-local`: Copia `juarvis` a `~/.local/bin/`

### `scripts/install.sh` (182 líneas)
- **Líneas 8-28**: `detect_os()`, `detect_arch()`: Detecta OS/Arch
- **Líneas 30-44**: `get_install_dir()`: Determina directorio (`~/.local/bin/`, etc.)
- **Líneas 47-76**: `add_to_path()`: Añade al PATH en shell rc
- **Líneas 81-139**: `install_juarvis()`: Copia binario local o descarga de GitHub Releases
- **Líneas 141-151**: `uninstall_juarvis()`: Elimina binario

### `pkg/assets/assets.go` (22 líneas)
- **Línea 12**: `//go:embed all:data` - Embebe TODO el directorio `data/`
- **Líneas 14-17**: `GetEmbeddedFS()`: Retorna `embed.FS` con assets
- **Línea 22**: `CopyEmbeddedToDisk()`: Copia directorio embebido al filesystem

---

## 7. RESUMEN EJECUTIVO

```
1. INSTALACIÓN:
   make install → scripts/install.sh
   ↓
   - Detecta OS/Arch (macOS, Linux, Windows)
   - Copia/descarga binario `juarvis`
   - Instala en ~/.local/bin/ o /usr/local/bin/
   - Añade al PATH en .bashrc/.zshrc
   ↓
   RESULTADO: `juarvis` disponible globalmente

2. INICIALIZACIÓN (en proyecto del usuario):
   juarvis init
   ↓
   - Extrae assets embebidos (marketplace.json, AGENTS.md, etc.)
   - Crea .juar/, plugins/, .agent/skills/
   - Crea manifests de plugins (plugin.json)
   - Instala pre-commit hook
   ↓
   RESULTADO: Ecosistema Juarvis listo para agentes IA

3. CARGA DE PLUGINS:
   juarvis load
   ↓
   - Lee plugins/ habilitados
   - Crea symlinks de skills (plugins/ → skills/)
   - Carga skills de usuario (.agent/skills/)
   - Genera .juar/skill-registry.md
   ↓
   RESULTADO: 26 plugins con skills indexadas

4. TRABAJO DE AGENTES (en proyecto del usuario):
   Agente orquestador recibe tarea
   ↓
   - go-developer: Escribe código del proyecto
   - test-engineer: Ejecuta tests del proyecto
   - code-reviewer: Revisa código
   - debugger: Investiga bugs
   - etc.
   ↓
   Auto-ejecutan: juarvis verify, juarvis commit, go test, etc.
```

---

**IMPORTANTE**: Juarvis NO es el proyecto en que trabajan los agentes. 
Juarvis es el **SISTEMA OPERATIVO** que usan los agentes para trabajar en **TU PROYECTO**.
