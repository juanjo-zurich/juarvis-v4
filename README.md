# Juarvis V4

Motor CLI autocontenido para ecosistemas de agentes IA con Spec-Driven Development (SDD).

## Qué es

Juarvis V4 es una herramienta de línea de comandos escrita en Go que prepara el terreno para que IAs (Claude, Cursor, Windsurf, etc.) trabajen de forma estructurada y segura en tu proyecto. Inicializa ecosistemas, gestiona plugins y skills, distribuye configuraciones a IDEs, y ejecuta un engine de hooks de seguridad.

## Instalación

```bash
git clone https://github.com/juanjo-zurich/juarvis-v4.git
cd juarvis-v4
make build      # Compila el binario
make install    # Instala en /usr/local/bin/
```

### Requisitos

- Go 1.22+
- Git (para snapshots)

## Uso Rápido

```bash
# 1. Inicializar un ecosistema en tu proyecto
cd mi-proyecto
juarvis init

# 2. Distribuir reglas a tu IDE
juarvis setup --ide opencode   # o --gui para interfaz web

# 3. Verificar estado
juarvis check

# 4. Gestionar plugins
juarvis pm list                # Ver plugins disponibles
juarvis pm install <plugin>    # Instalar un plugin
juarvis load                   # Re-indexar skills
```

## Comandos

| Comando | Descripción |
|---------|-------------|
| `juarvis` | Sin argumentos: detecta ecosistema y muestra estado |
| `juarvis init [path]` | Inicializa un ecosistema Juarvis |
| `juarvis check` | Health-check del ecosistema |
| `juarvis setup --ide <ide>` | Distribuye configuraciones al IDE |
| `juarvis setup --gui` | Interfaz web de configuración |
| `juarvis load` / `sync` | Indexa plugins y regenera skills |
| `juarvis skill-create [name]` | Crea plantilla de nueva skill |
| `juarvis pm list` | Lista plugins del marketplace |
| `juarvis pm install <plugin>` | Instala un plugin |
| `juarvis pm enable/disable/remove` | Gestiona plugins |
| `juarvis snapshot create <name>` | Crea snapshot de seguridad |
| `juarvis snapshot restore` | Restaura último snapshot |
| `juarvis snapshot prune` | Limpia snapshots viejos |
| `juarvis ralph loop <prompt>` | Inicia bucle autónomo |
| `juarvis ralph stop` | Detiene bucle autónomo |
| `juarvis hookify list` | Lista reglas de hook activas |
| `juarvis memory` | Servidor MCP de memoria local (uso interno) |

**Flags globales:** `--root <path>`, `--json`, `--version`

## Arquitectura

```
cmd/          → Comandos CLI (Cobra)
pkg/assets/   → go:embed de todos los datos
pkg/config/   → Constantes centralizadas de paths
pkg/init/     → Extracción de assets al filesystem
pkg/setup/    → Distribución a IDEs (+ servidor GUI)
pkg/loader/   → Indexación atómica de plugins
pkg/pm/       → Package Manager (marketplace, install, search)
pkg/hookify/  → Engine de hooks de seguridad
pkg/ralph/    → Motor de bucles autónomos
pkg/validate/ → Health-check del ecosistema
pkg/snapshot/ → Snapshots via git stash
pkg/root/     → Detección del directorio raíz
pkg/output/   → Output centralizado (emojis + JSON)
pkg/memory/   → Servidor MCP de memoria local (SQLite FTS5)
pkg/utils/    → Utilidades compartidas (embed helpers)
```

## Ecosistema

- **21 plugins** embebidos (SDD, backend, frontend, testing, CI/CD, seguridad, PR review, etc.)
- **71+ skills** indexadas automáticamente
- **7 IDEs** soportados (OpenCode, Cursor, Windsurf, VS Code, Antigravity, Trae, Kiro)
- **Servidor MCP de memoria local** integrado (reemplaza dependencia externa de engram)
- **Hooks de seguridad** cableados en opencode.json (PreToolUse, PostToolUse, Stop, UserPromptSubmit)
- **73 tests** pasando con coverage completo

## Desarrollo

```bash
make build    # Compilar
make test     # Tests con coverage
make lint     # go vet
make clean    # Limpiar binario
```

## Licencia

MIT
