# Juarvis V4

Motor CLI autocontenido para ecosistemas de agentes IA con Spec-Driven Development (SDD), enforcement automático y aprendizaje continuo.

## Qué es

Juarvis V4 es una herramienta de línea de comandos escrita en Go que prepara el terreno para que IAs (Claude, Cursor, Windsurf, OpenCode, etc.) trabajen de forma **estructurada, segura y autónoma** en tu proyecto.

**Tres pilares:**
1. **Ecosistema autocontenido** — 21 plugins y 70+ skills embebidos en un solo binario
2. **Enforcement automático** — Watcher que vigila cambios y evalúa reglas de seguridad sin intervención humana
3. **Aprendizaje continuo** — Los agentes aprenden de sus errores usando memoria persistente (MCP local)

## Instalación

```bash
git clone https://github.com/juanjo-zurich/juarvis-v4.git
cd juarvis-v4
make build      # Compila el binario
make install    # Instala en /usr/local/bin/
```

### Requisitos

- **Go 1.23+**
- **Git** (para snapshots)

## Uso Rápido

```bash
# 1. Inicializar un ecosistema en tu proyecto
cd mi-proyecto
juarvis init

# 2. Distribuir reglas a tu IDE (incluye watcher automático)
juarvis setup --ide opencode   # o cursor, windsurf, vscode, antigravity, trae, kiro
# O configura todos a la vez:
juarvis setup --all

# 3. Verificar estado
juarvis check

# 4. El watcher arranca automáticamente al abrir el proyecto en tu IDE
# Vigila cambios, evalúa reglas de seguridad, crea auto-snapshots
```

## Comandos

### Gestión del Ecosistema

| Comando | Descripción |
|---------|-------------|
| `juarvis` | Sin argumentos: detecta ecosistema y muestra estado |
| `juarvis init [path]` | Inicializa un ecosistema completo (21 plugins, 70+ skills) |
| `juarvis check` | Health-check del ecosistema |
| `juarvis verify` | Verifica que el proyecto está sano (build, tests, configs) |
| `juarvis sync` | Actualiza archivos locales con la versión del binario |

### Watcher (Enforcement Automático)

| Comando | Descripción |
|---------|-------------|
| `juarvis watch` | Inicia daemon que vigila cambios y evalúa reglas automáticamente |
| `juarvis watch --daemon` | Ejecuta en segundo plano |
| `juarvis watch --stop` | Detiene el watcher en segundo plano |
| `juarvis watch --no-auto-snapshot` | Desactiva auto-snapshots |

> **Nota:** Al ejecutar `juarvis setup --ide <ide>`, se genera automáticamente la configuración para que el watcher arranque al abrir el proyecto en tu IDE.

### Plugins y Skills

| Comando | Descripción |
|---------|-------------|
| `juarvis load` | Indexa plugins y regenera symlinks de skills |
| `juarvis pm list` | Lista plugins del marketplace local |
| `juarvis pm search <q>` | Busca skills en skills.sh (proveedores oficiales) |
| `juarvis pm install <plugin>` | Instala un plugin (local o desde GitHub) |
| `juarvis pm enable/disable <plugin>` | Activa/desactiva un plugin |
| `juarvis pm remove <plugin>` | Elimina un plugin |
| `juarvis skill create <name>` | Crea plantilla de nueva skill |

### Seguridad y Snapshots

| Comando | Descripción |
|---------|-------------|
| `juarvis snapshot create <name>` | Crea snapshot de seguridad (git stash) |
| `juarvis snapshot restore` | Restaura último snapshot |
| `juarvis snapshot prune` | Limpia snapshots viejos |
| `juarvis hookify list` | Lista reglas de hook activas |

### Agentes Autónomos

| Comando | Descripción |
|---------|-------------|
| `juarvis ralph loop <prompt>` | Inicia bucle autónomo iterativo |
| `juarvis ralph stop` | Detiene bucle de Ralph |

### Configuración

| Comando | Descripción |
|---------|-------------|
| `juarvis setup --ide <ide>` | Distribuye reglas al IDE + watcher automático |
| `juarvis setup --all` | Distribuye a TODOS los IDEs soportados |
| `juarvis setup --gui` | Interfaz web de configuración |

**Flags globales:** `--root <path>`, `--json`, `--version`

## Cómo Funciona

### Flujo de Trabajo Completo

```
┌─────────────────────────────────────────────────────────────┐
│  1. INSTALAR (una vez por máquina)                          │
│     make install                                            │
├─────────────────────────────────────────────────────────────┤
│  2. INICIALIZAR (una vez por proyecto)                      │
│     juarvis init                                            │
│     → Extrae 21 plugins + 70+ skills del binario            │
│     → Crea .juar/ con skill-registry y memoria              │
├─────────────────────────────────────────────────────────────┤
│  3. CONFIGURAR (una vez por proyecto)                       │
│     juarvis setup --ide opencode                            │
│     → Distribuye reglas, skills y watcher al IDE            │
│     → Genera .vscode/tasks.json con runOn: folderOpen       │
├─────────────────────────────────────────────────────────────┤
│  4. TRABAJAR (diario)                                       │
│     Al abrir el proyecto → watcher arranca solo             │
│     → Vigila cambios en tiempo real                         │
│     → Evalúa reglas de hookify automáticamente              │
│     → Auto-snapshot si hay cambios masivos                  │
│     → El agente aprende de sus errores (Reflection Loop)    │
├─────────────────────────────────────────────────────────────┤
│  5. ACTUALIZAR (cuando haya nueva versión)                  │
│     make install && juarvis sync                            │
│     → Actualiza binario y sincroniza proyecto               │
└─────────────────────────────────────────────────────────────┘
```

### Reflection Loop: Aprendizaje Continuo

Los agentes siguen un ciclo de 3 fases:

| Fase | Cuándo | Qué hace |
|------|--------|----------|
| **Pre-tarea** | Antes de tareas no triviales | Busca errores pasados en memoria (`mem_context`, `mem_search`) |
| **Durante error** | Cuando algo falla | Guarda causa raíz y solución (`mem_save`) |
| **Post-tarea** | Al cerrar sesión | Guarda decisiones, patrones y resumen (`mem_session_summary`) |

### Enforcement Automático (Watcher)

El watcher (`juarvis watch`) proporciona un "cinturón de seguridad" que no depende de la voluntad del agente:

1. **Vigila** cambios en archivos del proyecto (ignora `.git/`, `.juar/`, `node_modules/`)
2. **Evalúa** reglas de hookify automáticamente en cada cambio
3. **Alerta** si detecta patrones peligrosos (secretos, comandos destructivos)
4. **Auto-snapshot** si detecta cambios masivos (configurable)

## Arquitectura

```
cmd/          → Comandos CLI (Cobra)
pkg/assets/   → go:embed de todos los datos (plugins, skills, configs)
pkg/config/   → Constantes centralizadas de paths
pkg/init/     → Extracción de assets al filesystem
pkg/setup/    → Distribución a IDEs (+ servidor GUI + watcher task)
pkg/loader/   → Indexación atómica de plugins (temp dir + rename)
pkg/pm/       → Package Manager (marketplace, install, search, HTTP retry)
pkg/hookify/  → Engine de hooks de seguridad (YAML rules, regex evaluation)
pkg/ralph/    → Motor de bucles autónomos
pkg/validate/ → Health-check del ecosistema
pkg/snapshot/ → Snapshots via git stash (apply, no pop)
pkg/root/     → Detección del directorio raíz (3 niveles máx)
pkg/output/   → Output centralizado (emojis + JSON mode)
pkg/memory/   → Servidor MCP de memoria local (JSON + índice en memoria)
pkg/watcher/  → Daemon de file watching (fsnotify + debounce + hookify)
pkg/sync/     → Sincronización de assets embebidos con proyecto local
pkg/utils/    → Utilidades compartidas (embed helpers, frontmatter)
```

## Ecosistema

| Característica | Detalle |
|----------------|---------|
| **Plugins embebidos** | 21 (SDD, backend, frontend, testing, CI/CD, seguridad, PR review, etc.) |
| **Skills indexadas** | 70+ |
| **IDEs soportados** | 8 (OpenCode, Cursor, Windsurf, VS Code, Antigravity, Trae, Kiro, Claude) |
| **Memoria MCP** | Servidor local integrado (JSON + índice en memoria, sin dependencias externas) |
| **Tests** | 73 pasando con coverage en 11 paquetes |
| **CI/CD** | GitHub Actions con 5 jobs paralelos (unit, integration, regression, e2e, verify) |
| **Marketplace** | skills.sh con filtrado de proveedores oficiales (vercel-labs, github, google-labs) |

## Seguridad

- **Permissions.yaml** — Reglas granulares (allow/deny/ask) para comandos bash, git, Go, etc.
- **Hookify** — Motor de hooks evalúa reglas YAML en cada cambio de archivo
- **Watcher** — Daemon automático que vigila y evalúa sin intervención humana
- **Auto-snapshot** — Backup automático ante cambios masivos
- **Validación de symlinks** — Previene path traversal en plugins
- **Validación de inputs** — Sanitización de nombres de snapshot, URLs de git, etc.
- **Auditoría** — Log de decisiones en `~/.config/juarvis/audit.log`

## Desarrollo

```bash
make build              # Compilar con ldflags (version, commit, date)
make test               # Tests unitarios con coverage
make test-integration   # Tests de integración (binario real)
make test-regression    # Tests de regresión (golden files)
make test-e2e           # Tests E2E (flujos completos)
make test-verify        # Ejecuta juarvis verify
make test-all           # Todos los tests
make lint               # go vet
make clean              # Limpiar binario
```

### Estructura de Tests

| Tipo | Dónde | Qué cubre |
|------|-------|-----------|
| **Unitarios** | `pkg/*_test.go` | Funciones individuales (73 tests) |
| **Integración** | `tests/integration/` | Binario CLI real (init, check, load, verify) |
| **Regresión** | `tests/regression/` | Golden files de output CLI |
| **E2E** | `tests/e2e/` | Flujos completos (init→check→load, snapshot, skill-create) |

### CI/CD

5 jobs paralelos en GitHub Actions:
1. **unit-tests** → build + vet + test -race + coverage
2. **integration-tests** → binario real + tests de integración
3. **regression-tests** → golden files de output
4. **e2e-tests** → flujos completos de usuario
5. **verify** → `juarvis verify` completo

## Licencia

MIT
