---
name: todo-tracking
description: Tracking de progreso con checklist durante tareas complejas
trigger: Cuando se necesita mantener un seguimiento de subtareas o verificar el progreso de una implementación
license: MIT
metadata:
  version: 1.0.0
  author: Juarvis
  language: es-ES
---

# Tracking de Progreso con Checklist

## Objetivo

Mantener una lista visible de subtareas con estado (`completado`, `en progreso`, `pendiente`) durante cualquier trabajo que implique múltiples pasos.

## Protocolo

### Al iniciar una tarea compleja

1. **Desglosar** la tarea en subtareas accionables (máximo 7 ± 2).
2. **Marcar** cada subtarea como `🔲 pendiente`.
3. **Priorizar** por dependencias: lo que bloquea a otros va primero.

### Durante la ejecución

1. **Actualizar** el estado en tiempo real:
   - `🔲 pendiente` — no iniciada
   - `🔄 en progreso` — trabajando ahora
   - `✅ completada` — verificada y cerrada
   - `⚠️ bloqueada` — depende de otro recurso
2. **Nunca** tener más de una subtarea en `🔄 en progreso`.
3. **Registrar** hallazgos inesperados como nuevas subtareas.

### Al completar una subtarea

1. **Verificar** contra el criterio de aceptación original.
2. **Marcar** como `✅ completada` solo tras verificación.
3. **Desbloquear** la siguiente subtarea dependiente.

## Formato del checklist

```markdown
## Progreso: [nombre de la tarea]

- [x] Subtarea 1 — completada
- [ ] Subtarea 2 — en progreso
- [ ] Subtarea 3 — pendiente (bloqueada por subtarea 2)
- [ ] Subtarea 4 — pendiente

**Estado global**: 1/4 completadas (25%)
```

## Reglas

- Una subtarea solo se marca completada si tiene verificación (test, revisión manual, o confirmación del usuario).
- Si una subtarea se complica, dividirla en sub-subtareas antes de continuar.
- Al finalizar, el checklist completo queda como registro del trabajo realizado.

## Salida esperada

Checklist visible en el chat, actualizado tras cada subtarea completada, con porcentaje de avance global.
