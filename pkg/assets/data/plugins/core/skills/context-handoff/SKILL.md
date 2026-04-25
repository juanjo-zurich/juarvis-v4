---
name: context-handoff
description: Protocolo de handoff cuando el contexto está agotándose
trigger: Cuando el contexto de la conversación se acerca al límite de tokens o tras compaction
license: MIT
metadata:
  version: 1.0.0
  author: Juarvis
  language: es-ES
---

# Protocolo de Handoff por Contexto Agotándose

## Objetivo

Transferir el estado del trabajo a un nuevo contexto sin pérdida de información crítica cuando el contexto actual está a punto de agotarse.

## Señales de activación

- El sistema indica compaction próxima o en curso.
- Las respuestas empiezan a truncarse o perder detalle.
- El usuario reporta que el modelo "olvidó" algo reciente.
- Tokens restantes estimados < 20% de la ventana.

## Protocolo de handoff

### Fase 1: Preparación (cuando se detecta contexto bajo)

1. **Persistir** todas las observaciones pendientes con `mem_save`.
2. **Generar** resumen de sesión con `mem_session_summary` usando el formato:
   - Goal
   - Instructions (preferencias del usuario descubiertas)
   - Discoveries (hallazgos técnicos)
   - Accomplished (qué se completó)
   - Next Steps (qué falta)
   - Relevant Files (archivos clave)

### Fase 2: Transferencia

1. **Identificar** el estado exacto del trabajo:
   - ¿En qué fase SDD estamos?
   - ¿Qué artefactos existen y cuáles faltan?
   - ¿Qué subtarea estaba en progreso?
2. **Registrar** decisiones pendientes sin resolver.
3. **Listar** comandos o acciones que se iban a ejecutar a continuación.

### Fase 3: Reconexión (al inicio del nuevo contexto)

1. **Llamar** a `mem_context()` para recuperar contexto reciente.
2. **Llamar** a `mem_search()` con palabras clave del trabajo en curso.
3. **Verificar** que el estado recuperado es coherente.
4. **Continuar** desde el último punto registrado.

## Formato del resumen de handoff

```markdown
## Handoff — [fecha/hora]

### Estado actual
- Fase SDD: [explore | propose | spec | design | tasks | apply | verify]
- Tarea en progreso: [descripción]
- Bloqueos activos: [si los hay]

### Últimas acciones
1. [acción completada 1]
2. [acción completada 2]
3. [acción en progreso — dónde se quedó]

### Próximos pasos inmediatos
1. [siguiente acción]
2. [acción subsiguiente]

### Archivos relevantes
- path/to/file — [rol]
- path/to/other — [rol]

### Decisiones pendientes
- [decisión sin resolver 1]
- [decisión sin resolver 2]
```

## Reglas

- **No esperar** al último momento: iniciar handoff cuando queden ~20% de tokens.
- **Priorizar**: estado actual > próximos pasos > contexto histórico.
- **Verificar** siempre que `mem_save` tuvo éxito antes de continuar.
- En modo degradado (servidor MCP no disponible), guardar el handoff en un archivo local `.handoff.md`.
