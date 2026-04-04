---
name: memory-update
description: Persistencia estructurada de descubrimientos usando formato What/Why/Where/Learned
trigger: Tras completar un bugfix, decisión arquitectónica, descubrimiento o cambio de configuración
license: MIT
metadata:
  version: 1.0.0
  author: Juarvis
  language: es-ES
---

# Persistencia Estructurada — What/Why/Where/Learned

## Objetivo

Asegurar que todo conocimiento relevante del proyecto se persiste de forma estructurada y recuperable entre sesiones.

## Cuándo persistir

Guardar observación inmediatamente después de:

- Completar un bugfix
- Tomar una decisión arquitectónica
- Descubrir un comportamiento inesperado del código
- Cambiar configuración o entorno
- Establecer un patrón o convención
- Aprender una preferencia del usuario

## Formato de la observación

```yaml
title: "Verbo + qué — corto y buscable"
type: bugfix | decision | architecture | discovery | pattern | config | preference
scope: project (por defecto) | personal
topic_key: "categoria/tema-estable"  # opcional, recomendado para temas evolutivos
content: |
  **Qué**: Una frase — qué se hizo
  **Por qué**: Qué lo motivó (petición del usuario, bug, rendimiento, etc.)
  **Dónde**: Archivos o rutas afectadas
  **Aprendizaje**: Gotchas, casos extremo, cosas que sorprendieron (omitir si no hay)
```

## Ejemplos

### Bugfix
```yaml
title: "Corregido error de concurrencia en cola de tareas"
type: bugfix
content: |
  **Qué**: Añadido mutex al acceso compartido de la cola de tareas
  **Por qué**: Race condition causaba pérdida de tareas bajo carga concurrente
  **Dónde**: internal/queue/worker.go, internal/queue/manager.go
  **Aprendizaje**: El mutex debe englobar tanto la lectura como la escritura del slice
```

### Decisión
```yaml
title: "Elegido Zustand sobre Redux para estado global"
type: decision
topic_key: "architecture/state-management"
content: |
  **Qué**: Adoptado Zustand como librería de estado global
  **Por qué**: Menor boilerplate, mejor rendimiento en renders, TypeScript nativo
  **Dónde**: src/stores/*.ts
  **Aprendizaje**: Zustand no soporta middleware de forma tan robusta como Redux Toolkit
```

## Reglas de topic_key

- Diferentes topics no deben sobreescribirse entre sí.
- Reusar el mismo `topic_key` para actualizar un tema evolutivo.
- Si no se está seguro de la clave, usar `mem_suggest_topic_key` primero.

## Recuperación

1. `mem_context()` — historial reciente de la sesión (rápido).
2. `mem_search(query)` — búsqueda de texto completo FTS5.
3. `mem_get_observation(id)` — contenido completo sin truncar.

## Reglas

- **Nunca** omitir el guardado tras un bugfix o decisión importante.
- **Nunca** guardar información trivial o redundante.
- **Siempre** incluir la ruta del archivo afectado en "Dónde".
- El título debe ser buscable: usar verbo + objeto (`Corregido X`, `Elegido Y`, `Descubierto Z`).
