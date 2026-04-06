---
description: Listar reglas hookify activas y su estado
argument-hint: [all|enabled|disabled]
---

## /hookify-list

Lista todas las reglas hookify configuradas en el proyecto.

### Uso
/hookify-list           # Muestra solo reglas enabled
/hookify-list all       # Muestra todas las reglas
/hookify-list enabled   # Muestra solo enabled
/hookify-list disabled  # Muestra solo disabled

### Comportamiento
1. Escanear directorio `~/.juarvis/hookify.*.local.md` para reglas locales
2. Escanear `plugins/*/hooks/` para reglas de plugins
3. Mostrar: nombre, evento, estado (enabled/disabled), prioridad, origen
4. Si no hay reglas configuradas, mostrar mensaje informativo

### Formato de salida
```
Regla hookify                    Evento        Estado     Prioridad   Origen
─────────────────────────────────────────────────────────────────────────────
no-console-log-ts               PostToolUse   enabled    media       local
block-test-deletion             PreToolUse    enabled    alta        local
auto-format-on-save             PostToolUse   disabled   baja        plugin:hookify
```

### Columnas
- **Regla**: Nombre del archivo sin prefijo ni extensión
- **Evento**: Cuándo se dispara (PreToolUse, PostToolUse, Stop, UserPromptSubmit)
- **Estado**: enabled/disabled (del frontmatter)
- **Prioridad**: baja/media/alta (del frontmatter)
- **Origen**: local (usuario) o plugin:<nombre>

### Ejemplos
```
/hookify-list              # Muestra reglas activas
/hookify-list all          # Muestra todas las reglas
/hookify-list disabled     # Muestra reglas desactivadas
```
