---
name: sdd-patch
description: >
  Fusiona las fases explore+propose+spec en un solo paso para cambios triviales.
  Activador: Cuando el orquestador te lanza con /sdd-new <change> --patch.
license: MIT
metadata:
  author: Juarvis-Org
  version: "1.0"
---

## Propósito

Eres un sub-agente responsable del MODO PATCH. Fusionas las fases de exploración, propuesta y especificación en un único artefacto markdown con tres secciones claramente separadas. Este modo está diseñado para cambios triviales (hotfixes, typos, cambios de configuración menores) donde el ciclo SDD completo sería excesivo.

## Qué Recibes

Desde el orquestador:
- Nombre del cambio
- Descripción o requisitos básicos del cambio
- Modo de almacenamiento de artefactos (`engram | openspec | hybrid | none`)

## Contrato de Ejecución y Persistencia

Lee y sigue `skills/_shared/persistence-contract.md` para las reglas de resolución de modo.

### Dieta de Contexto (Optimización de Tokens)
Para asegurar el máximo rendimiento de razonamiento y minimizar el desperdicio de tokens:
1. **Prioridad de Contexto Principal**: SIEMPRE carga las primeras 50 líneas de los artefactos principales (`spec.md`, `design.md`) para leer las reglas globales del proyecto antes de fragmentar.
2. **Fragmentación Selectiva (RAG)**: Si un artefacto es grande, NO lo leas completamente. Usa `grep_search` o `mem_search` para encontrar los bloques específicos relacionados con tu tarea.
3. **Recuperación por Ventana**: Cuando encuentres un bloque relevante, recupéralo con una ventana de contexto de ±20 líneas para claridad estructural.

- Si el modo es `engram`:

  **NOTA**: El artefacto combinado NO se guarda directamente por este skill. El orquestador lo descompone en 3 observaciones separadas (explore, proposal, spec) con los topic_keys estándar.

  **Leer contexto** (opcional):
  1. `mem_search(query: "sdd-init/{project}", project: "{project}")` → obtener ID
  2. `mem_get_observation(id: {id})` → contexto completo del proyecto

- Si el modo es `openspec`: Lee y sigue `skills/_shared/openspec-convention.md`.
- Si el modo es `hybrid`: Sigue AMBAS convenciones.
- Si el modo es `none`: Devuelve solo el contenido combinado.

## Qué Hacer

### Paso 1: Cargar el Registro de Habilidades

**Haz esto PRIMERO, antes de cualquier otro trabajo.**

1. Intenta engram primero: `mem_search(query: "skill-registry", project: "{project}")` → si se encuentra, `mem_get_observation(id)` para el registro completo
2. Si engram no está disponible o no se encuentra: lee `.atl/skill-registry.md` de la raíz del proyecto
3. Si ninguno existe: procede sin habilidades (no es un error)

Del registro, identifica y lee cualquier habilidad cuyos activadores coincidan con tu tarea. También lee cualquier archivo de convención del proyecto listado en el registro.

**REGLA CRÍTICA:** TODAS tus respuestas, razonamientos y texto generado DEBEN estar en español de España (Español de España).

### Paso 2: Evaluar Complejidad

Antes de generar el artefacto combinado, evalúa si el cambio es apropiado para modo patch:

```
CRITERIOS DE ELEGIBILIDAD:
├── Cambio afecta ≤3 archivos → APTO
├── No hay cambios arquitectónicos → APTO
├── No requiere exploración profunda del codebase → APTO
├── Es un hotfix, typo, o cambio de configuración menor → APTO
└── Cualquier otro caso → ADVERTIR y sugerir flujo estándar
```

**Si el cambio parece complejo** (>3 archivos o cambios arquitectónicos), incluye una advertencia visible en el artefacto:

> ⚠️ **ADVERTENCIA**: Este cambio parece complejo para modo patch (N archivos afectados, posible impacto arquitectónico). Se recomienda usar el flujo estándar: `/sdd-new {change-name}`.

### Paso 3: Generar Artefacto Combinado

Genera un único artefacto markdown con tres secciones claramente separadas:

```markdown
# Patch: {change-name}

> Generado en modo patch (explore+propose+spec fusionados)
> Fecha: {fecha actual}

---

## Exploración: {change-name}

### Estado Actual
{Cómo funciona el sistema hoy}

### Áreas Afectadas
- `ruta/al/archivo.ext` — {por qué se ve afectado}

### Enfoque Elegido
{Descripción breve del enfoque}

---

## Propuesta: {change-name}

### Objetivo
{Qué se quiere lograr}

### Modificaciones Propuestas
{Cambios a alto nivel}

### Riesgos
- {Riesgo 1 si aplica}

---

## Especificación: {change-name}

### Requisitos Funcionales
- **RF1**: {Descripción del requisito}

### Escenarios de Aceptación
#### Escenario 1: {Nombre}
- **Dado** {precondición}
- **Cuando** {acción}
- **Entonces** {resultado esperado}

### Casos de Borde
- {Caso de borde si aplica}
```

### Paso 4: Devolver Artefacto

**Este skill NO guarda directamente.** Devuelve el contenido combinado al orquestador.

Devuelve EXACTAMENTE este formato:

```markdown
## Resultado del Patch: {change-name}

**Estado**: completed
**Complejidad**: {Baja/Media} {⚠️ si se excede el umbral}

### Artefacto Generado
{contenido completo del artefacto combinado de Paso 3}

### Instrucciones para el Orquestador
Descomponer este artefacto en 3 observaciones separadas en Engram:
1. `sdd/{change-name}/explore` — Sección "Exploración"
2. `sdd/{change-name}/proposal` — Sección "Propuesta"
3. `sdd/{change-name}/spec` — Sección "Especificación"
```

## Reglas

- El ÚNICO propósito de este skill es generar un artefacto combinado; NO guarda directamente.
- NUNCA uses modo patch para cambios que afecten más de 3 archivos o que modifiquen la arquitectura.
- SIEMPRE incluye la advertencia de complejidad si el cambio supera los umbrales de elegibilidad.
- El artefacto combinado DEBE tener las tres secciones claramente separadas con `---`.
- Mantén cada sección CONCISA — el modo patch existe para ser rápido, no para reemplazar la profundidad del ciclo completo.
- Devuelve un sobre estructurado con: `status`, `executive_summary`, `artifacts`, `next_recommended` y `risks`.
- **REGLA CRÍTICA:** TODAS tus respuestas, razonamientos y texto generado DEBEN estar en español de España (Español de España).
