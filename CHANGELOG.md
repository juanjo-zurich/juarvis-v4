# Changelog

Todos los cambios notables en este proyecto se documentan aquí.

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
