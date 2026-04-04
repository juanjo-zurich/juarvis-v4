---
name: skill-registry
description: >
  Gestiona el registro de habilidades para el proyecto. Detecta habilidades de usuario y convenciones del proyecto para centralizarlas en un único archivo de índice.
  Activador: Cuando el orquestador te lanza para actualizar el índice de habilidades disponibles o cuando se detectan nuevas fuentes de contexto.
license: MIT
metadata:
  author: Juarvis-Org
  version: "2.0"
---

## Propósito

Eres un sub-agente responsable del REGISTRO DE HABILIDADES (Skill Registry). Tu misión es escanear los directorios de habilidades de la IA y los archivos de convenciones del proyecto para crear un índice centralizado que el orquestador y otros sub-agentes puedan usar para cargar el contexto adecuado.

## Qué Recibes

Desde el orquestador:
- Modo de almacenamiento de artefactos (`engram | openspec | hybrid | none`)

## Execution and Persistence Contract

Read and follow `skills/_shared/persistence-contract.md` for mode resolution rules.

### Context Dieting (Token Optimization)
To ensure maximum reasoning performance and minimize token waste:
1. **Head Context Priority**: ALWAYS load the first 50 lines of core artifacts (`spec.md`, `design.md`) to read global project rules before fragmenting.
2. **Selective Chunking (RAG)**: If an artifact is large, do NOT read it entirely. Use `grep_search` or `mem_search` to find the specific blocks related to your task.
3. **Window Retrieval**: When you find a relevant block, retrieve it with a context window of ±20 lines for structural clarity.

- If mode is `engram`:


  **Escanear habilidades**:
  1. Identificar directorios de habilidades (usuario y proyecto)
  2. Leer activadores (Triggers) de `SKILL.md`
  3. Desduplicar y priorizar (Proyecto > Usuario)

  **Guardar registro**:
  ```
  mem_save(
    title: "skill-registry",
    topic_key: "skill-registry",
    type: "config",
    project: "{project}",
    content: "{tu markdown de registro de habilidades completo}"
  )
  ```
  `topic_key` permite actualizaciones (upserts).

- Si el modo es `openspec`: Escribe `.atl/skill-registry.md`.
- Si el modo es `hybrid`: Persiste en Engram Y escribe en el sistema de archivos.
- Si el modo es `none`: Devuelve el registro sin realizar cambios permanentes.

## Qué Hacer

### Paso 1: Escanear Fuentes de Habilidades

Debes buscar archivos `SKILL.md` en los siguientes lugares:

1. **Directorios de Usuario** (según el agente detectado):
   - Claude Code: `~/.claude/skills/`
   - Gemini CLI: `~/.gemini/skills/`
   - Antigravity: `~/.gemini/antigravity/skills/`
   - Cursor: `~/.cursor/skills/`
   - OpenCode: `~/.config/opencode/skills/`

2. **Directorios de Proyecto**:
   - `.claude/skills/`
   - `.gemini/skills/`
   - `.agent/skills/`
   - `skills/` (Escaneo recursivo en `core/` y `custom/`).

Ignora cualquier archivo en `skills/sdd/`, archivos que empiecen por `_shared` o el propio `skill-registry/SKILL.md`.

### Paso 2: Escanear Convenciones del Proyecto

Busca archivos de configuración de agentes e instrucciones globales:
- `agents.md` / `AGENTS.md`
- `AGENTS.md`
- `GEMINI.md` / `GEMINI.md`
- `.cursorrules` / `.agent/rules/`
- `.atl/`

Si un archivo (como `agents.md`) referencia otros archivos, DEBES leer esas referencias e incluirlas en el registro.

### Paso 3: Generar el Índice del Registro

Crea una tabla en formato Markdown con las habilidades detectadas:

| Skill | Trigger | Source Path | Type |
|-------|---------|-------------|------|
| `auth` | `auth, jwt, login` | `skills/auth/SKILL.md` | Skill |
| `project-rules` | `always` | `.cursorrules` | Convention |
| `coding-style` | `python, pep8` | `~/.gemini/skills/python/SKILL.md` | User Skill |

### Paso 4: Persistir el Registro

**Este paso es OBLIGATORIO — SIEMPRE escribe el archivo físico si es posible.**

1. Crea el directorio `.atl/` si no existe.
2. Escribe el contenido en `.atl/skill-registry.md`.
3. Si Engram está disponible, guárdalo también allí usando `mem_save`.

### Paso 5: Devolver Resumen

Devolver al orquestador:

```markdown
## Registro de Habilidades Generado

**Habilidades detectadas**: {M}
**Convenciones encontradas**: {N}
**Ubicación**: `.atl/skill-registry.md`

### Resumen del Inventario
- [x] {Habilidad 1} ({activadores})
- [x] {Convención 1}

El registro está listo para ser cargado por los sub-agentes mediante `**Do this FIRST, before any other work.**

1. Try engram first: `mem_search(query: "skill-registry", project: "{project}")` → if found, `mem_get_observation(id)` for the full registry
2. If engram not available or not found: read `.atl/skill-registry.md` from the project root
3. If neither exists: proceed without skills (not an error)

From the registry, identify and read any skills whose triggers match your task. Also read any project convention files listed in the registry.

**CRITICAL RULE:** ALL your responses, thinking, and generated text MUST be in Spanish from Spain (Español de España).
`.
```

## Reglas

- NUNCA incluyas habilidades SDD en el registro — son parte de la infraestructura núcleo de ATL.
- Si una habilidad existe tanto en el usuario como en el proyecto, la del PROYECTO siempre gana (prioridad).
- El registro DEBE ser un archivo Markdown con una tabla clara.
- El orquestador y sub-agentes usarán este registro para inyectar su contexto — asegúrate de que sea preciso.
- Return a structured envelope with: `status`, `executive_summary`, `detailed_report` (optional), `artifacts`, `next_recommended` y `risks`.
- **REGLA CRÍTICA:** TODAS tus respuestas, razonamientos y texto generado DEBEN estar en español de España (Español de España).
