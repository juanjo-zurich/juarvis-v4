---
name: juarvis-meta
description: Como usar el CLI juarvis — tu propia herramienta de trabajo
---

# Juarvis — Tu propia herramienta

## Comandos disponibles

| Comando | Qué hace | Cuándo usarlo |
|---------|----------|---------------|
| `juarvis check` | Health check del ecosistema | Antes de empezar a trabajar |
| `juarvis snapshot create "descripción"` | Backup con git stash | **ANTES** de cualquier cambio de código |
| `juarvis snapshot restore` | Revertir al último snapshot | Si algo falla tras un cambio |
| `juarvis snapshot prune` | Limpiar snapshots viejos | Mantenimiento periódico |
| `juarvis load` | Regenerar skill-registry y symlinks | Tras crear/editar/eliminar skills o plugins |
| `juarvis pm list` | Ver plugins del marketplace | Para ver qué plugins hay disponibles |
| `juarvis pm install <plugin>` | Instalar un plugin | Para añadir funcionalidad nueva |
| `juarvis pm enable/disable <plugin>` | Activar/desactivar plugin | Gestión de plugins |
| `juarvis skill create <name>` | Crear nueva skill en .agent/skills/ | Para añadir conocimiento nuevo al agente |
| `juarvis verify` | Verificar que el proyecto está sano | Antes de commitear |
| `juarvis ralph loop "tarea" --max-iterations N` | Bucle iterativo autónomo | Para tareas que requieren iteración |
| `juarvis ralph stop` | Detener bucle de Ralph | Cuando el bucle debe parar |
| `juarvis setup --ide <ide>` | Distribuir reglas al IDE | Tras inicializar ecosistema |
| `juarvis setup --gui` | Interfaz web de configuración | Alternativa visual a --ide |

## Protocolo obligatorio

1. **Antes de editar código**: `juarvis snapshot create "antes de X"`
2. **Tras crear/editar skills**: `juarvis load`
3. **Antes de commitear**: `juarvis verify`
4. **Si algo falla**: `juarvis snapshot restore`

## Hookify — crear reglas de comportamiento

Las reglas de hookify van en `.opencode/hookify.*.local.md`:

```markdown
---
name: mi-regla
enabled: true
event: bash
pattern: rm\s+-rf
action: block
---

No uses rm -rf.
```

Eventos disponibles: `bash`, `file`, `stop`, `userpromptsubmit`
Acciones: `warn` (advertencia), `block` (bloquea con exit code 2)

## Estructura del ecosistema

```
.juar/
  skill-registry.md  # Índice de todas las skills disponibles
  memory/            # Memoria persistente del agente (MCP)
plugins/             # Plugins con sus skills
skills/              # Symlinks a las skills de plugins habilitados
```

## Si Engram (MCP memory) no responde

1. Intenta reconectar — puede ser temporal
2. Si persiste, informa al usuario
3. Continúa trabajando sin persistencia entre sesiones
4. No bloquees el trabajo por falta de memoria persistente
