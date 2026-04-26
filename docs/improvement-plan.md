# Plan de Mejoras Juarvis V4

> Documento de planificación técnica para evolutivos del ecosistema.

---

## Estado Actual (Análisis)

### Fortalezas Identificadas

| Área | Descripción |
|------|------------|
| **Arquitectura CLI** | Comandos bien aislados con Cobra, estructura limpia |
| **Assets Embebidos** | Todo el ecosistema cabe en el binario |
| **Sistema de Memoria MCP** | stdio-based, IDE-agnostic |
| **Seguridad Zero-Trust** | Whitelist de proveedores oficiales |
| **Watcher** | Auto-snapshot, scoring de archivos |
| **Plugins** | Sistema de módulos con enabled/disabled |

### Limitaciones y Deuda Técnica

| severidad | Área | Problema |
|----------|------|----------|
| 🔴 alta | Memoria | Solo file-based (JSON), sin búsqueda eficiente |
| 🔴 alta | Persistencia | No hay sync cloud, atado a la máquina |
| 🟡 media | Permisos | Declarativos, el IDE debe respetarlos |
| 🟡 media | Watcher | No es un servicio real (no hay restart, log rotation) |
| 🟡 media | Tests | Pocos e2e, coverage limitado |
| 🟢 baja | Plugins | Sin versionado semver |
| 🟢 baja | Telemetry | No hay analytics de uso |
| 🟢 baixa | Windows | Soporte parcial (fsnotify) |

---

## Hoja de Ruta

### Fase 1: Estabilidad (v4.1.x)

> **Objetivo**: Corregir bugs críticos y mejorar robustness.

#### 1.1 Sistema de Memoria

```
PROBLEMA: Búsqueda lenta en proyectos grandes (miles de observaciones)
SOLUCIÓN: SQLite embebido + índices FTS5
```

- [ ] Reemplazar almacenamiento JSON por SQLite (`modernc.org/sqlite`)
- [ ] Añadir índices full-text search para mem_search
- [ ] Migrar datos existentes `.juar/memory/` → `.juar/memory.db`
- [ ] Añadir `mem_gc` command (garbage collection automática)

```go
// pkg/memory/storage.go
type Storage struct {
    db *sql.DB  // en vez de map[string]*Observation
}
```

#### 1.2 Tests y Coverage

```
PROBLEMA: Pocos tests, miedo a refactors
SOLUCIÓN: Aumentar coverage a 70%
```

- [ ] Añadir tests unitarios para `pkg/hookify/` (regex matching)
- [ ] Añadir tests de integración para `juarvis init`
- [ ] Añadir tests e2e para flujo completo `juarvis up`
- [ ] Configurar CI con GitHub Actions
- [ ] Badge de coverage en README

#### 1.3 Watcher Robustness

```
PROBLEMA: Muere silenciosamente, no hace restart
SOLUCIÓN: Mejorar gestión de errores y logging
```

- [ ] Añadir log rotation (`/tmp/juarvis-watch.log`)
- [ ] Health check endpoint (opcional, para monitoring)
- [ ] Auto-restart con exponential backoff
- [ ] Signal handling mejorado (SIGUSR1 para reload config)

---

### Fase 2: Funcionalidad (v4.2.x)

> **Objetivo**: Añadir features solicitadas.

#### 2.1 Permisos Enforcement

```
PROBLEMA: permissions.yaml es declarativo
SOLUCIÓN: Hook de pre-ejecución real
```

- [ ] Crear `juarvis guard` command (serve STDIN, evalúa permisos)
- [ ] Integrar con MCP del IDE como tool
- [ ] Soportar `permissions.yaml` más expresivo (AND/OR)

```yaml
permission:
  bash:
    "*": "allow"
    "rm -rf *": "deny"
    "git push *": "ask"
    # Nueva sintaxis
    "npm install *": "allow_if:package.json"
```

#### 2.2 Plugin Versioning

```
PROBLEMA: No hay forma de actualizar plugins
SOLUCIÓN: Semver + update command
```

- [ ] Añadir campo `version` en `plugin.json`
- [ ] `juarvis pm check` (check updates disponibles)
- [ ] `juarvis pm update <plugin>` (update a latest)
- [ ] `juarvis pm rollback <plugin>` (downgrade)

#### 2.3 Cloud Sync (Opcional)

```
PROBLEMA: Memoria atada a la máquina
SOLUCIÓN: Sync opcional con proveedor
```

- [ ] Diseño: providers interface (no hardcoded)
- [ ] Provider local-file (default, actual)
- [ ] Provider: GitHub Gist (opcional)
- [ ] Provider: Custom endpoint (opcional)

```yaml
# agent-settings.json
memory:
  provider: "local"  # "gist" | "custom"
  gist:
    token: "${GITHUB_TOKEN}"
    id: "optional-id"
```

#### 2.4 Hot Reload de Config

```
PROBLEMA: Cambios en AGENTS.md requieren reiniciar IDE
SOLUCIÓN: Auto-reload con fsnotify
```

- [ ] Watcher detecta cambios en `AGENTS.md`, `permissions.yaml`
- [ ] Broadcast a IDE via MCP notification
- [ ] Soporte para `.juarvis/reload` trigger

---

### Fase 3: Ecosistema (v4.3.x)

> **Objetivo**: Mejorar DX y comunidad.

#### 3.1 Plugin Marketplace Mejorado

```
PROBLEMA: solo search básico en skills.sh
SOLUCIÓN: marketplace propio con ratings
```

- [ ] `juarvis pm publish <plugin>` (publicar al registry)
- [ ] Ratings y reviews (local file, no cloud)
- [ ] Dependencias entre plugins (`peerDependencies`)
- [ ] Plugin templates repository

#### 3.2 IDE Integrations

```
PROBLEMA: config manual para cada IDE
SOLUCIÓN: Auto-detect + config
```

- [ ] `juarvis setup --auto` (detecta IDEs instalados)
- [ ] Soporte para Windsurf (actualmente parcial)
- [ ] Soporte para Zed
- [ ] VS Code extension (futuro)

#### 3.3 Developer Experience

```
PROBLEMA: Debugging oscuro
SOLUCIÓN: Mejor logging y debugging
```

- [ ] `juarvis debug` command (verbose mode)
- [ ] JSON output en todos los comandos
- [ ] `--dry-run` flag para comandos destructivos
- [ ] `juarvis doctor` (diagnose system)

---

## Plan de Implementación

### Sprint Structure

```
Sprint 1 (1 semana):
├── [1.1.1] SQLite migration skeleton
└── [1.2.1] Tests para hookify

Sprint 2 (1 semana):
├── [1.1.2] Migration script
└── [1.2.2] CI setup

Sprint 3 (1 semana):
├── [1.3.1] Log rotation
└── [1.3.2] Signal handling

Sprint 4 (1 semana):
└── [2.1.1] juarvis guard prototype
```

### Criterios de Éxito

| Feature | Métrica |
|---------|--------|
| Memoria | Búsqueda <100ms en 10k obs |
| Tests | Coverage >70% |
| Watcher | Uptime >99% en 24h |
| Guard | Latencia <50ms por check |

### Rollback Plan

- **Antes de chaque fase**: `juarvis snapshot create "antes-fase-X"`
- **Migration**: siempre backwards-compatible
- **Feature flags**: para features experimentales

---

## Priorities Inmediatas (Next 2 semanas)

1. **SQLite para memoria** - El más necesario
2. **Más tests** - Base para refactors seguros
3. **CI/CD** - Automatizar releases

---

## Apéndice: Tech Stack Suggestions

| Área | Opción | Razón |
|------|-------|-------|
| DB | `modernc.org/sqlite` | Embeddable, single file |
| UI | `tview` | TUI para `juarvis gui` |
| Auth | No necesario | Sistema local |
| Cloud | Provider pattern | No lock-in |