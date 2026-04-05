# Changelog

Todos los cambios notables en este proyecto se documentan aquí.

## [Unreleased]

### Added
- Servidor MCP de memoria local (`pkg/memory/`) con SQLite FTS5, reemplaza dependencia externa de engram
- Comando `juarvis memory` para servir como MCP server
- 9 herramientas MCP: mem_save, mem_search, mem_context, mem_session_summary, mem_get_observation, mem_suggest_topic_key, mem_update, mem_delete, mem_session_start, mem_session_end
- 4 nuevos plugins: backend-patterns, api-error-handling, frontend-design, frontend-ui
- 3 nuevos plugins de PR review: pr-review, pr-code-reviewer, pr-comment-analyzer
- Hooks de seguridad cableados en opencode.json (PreToolUse, PostToolUse, Stop, UserPromptSubmit)
- Verificación de PATH en setup command
- `juarvis` sin argumentos detecta ecosistema y muestra estado
- Migración automática `.atl/` → `.juar/` durante init
- Comando `juarvis ralph stop`
- Paquete `pkg/config/paths.go` con constantes centralizadas de paths
- Tests offline deterministas para `pkg/pm/`

### Fixed
- Snapshot create destruía el snapshot (stash pop → stash apply)
- Hookify buscaba reglas en ruta incorrecta (.opencode/ en lugar de .juar/)
- Tests sin assertions (validate_test.go, loader_test.go)
- Error handling en skill-create
- Race condition en regexCache (sync.Map)
- Font externa en GUI (offline)
- Tests frágiles con HTTP real
- Código duplicado (copyEmbeddedDir)
- Parsing YAML manual → gopkg.in/yaml.v3
- Git argument injection en snapshot.go y pm.go
- Error handling inconsistente en 6 comandos cmd/
- Paths .opencode/ hardcodeados en hookify.go
- Deduplicación de skills por nombre
- YAML parsing manual reemplazado por librería estándar

### Security
- Servidor GUI ahora escucha solo en 127.0.0.1
- Validación de contenido en skill-registry
- Git argument injection mitigado en snapshot y pm

### Changed
- Reemplazada dependencia externa de engram por servidor MCP de memoria local integrado
- Centralizados todos los paths hardcodeados en `pkg/config/paths.go`
- Hooks de seguridad movidos de archivos sueltos a opencode.json
- Migración automática de directorio `.atl/` a `.juar/`

## [1.0.0] - 2026-04-04

### Added
- CLI completo con 20+ comandos
- 20 plugins embebidos con 68+ skills
- Spec-Driven Development (9 fases)
- Engine de hooks de seguridad (Hookify)
- Motor de bucles autónomos (Ralph)
- Servidor GUI para configuración (--gui)
- Sistema de snapshots via git stash
- Package Manager con marketplace
- Soporte para 7 IDEs
- Makefile con targets estándar
- Tests unitarios (37+ tests)

### Fixed
- Snapshot create destruía el snapshot (stash pop → stash apply)
- Hookify buscaba reglas en ruta incorrecta
- Tests sin assertions (validate, loader)
- Error handling en skill-create
- Race condition en regexCache (sync.Map)
- Font externa en GUI (offline)
- Tests frágiles con HTTP real
- Código duplicado (copyEmbeddedDir)
- Parsing YAML manual (gopkg.in/yaml.v3)

### Security
- Servidor GUI ahora escucha solo en 127.0.0.1
- Validación de contenido en skill-registry
