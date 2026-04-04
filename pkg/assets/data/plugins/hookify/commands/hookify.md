---
description: Crear regla de hook por lenguaje natural
argument-hint: Descripción de la regla (ej: "Avísame cuando use console.log en TypeScript")
---

# Hookify — Crear regla por lenguaje natural

Interpreta la descripción del usuario y genera un archivo de regla hookify.

## Proceso

1. **Analizar la petición** del usuario para extraer:
   - **Evento**: ¿En qué momento debe dispararse? (PreToolUse, PostToolUse, Stop, UserPromptSubmit)
   - **Condición**: ¿Qué debe detectar? (herramienta, lenguaje, patrón de código, ruta)
   - **Acción**: ¿Qué debe hacer? (bloquear, avisar, registrar, transformar)
   - **Nombre**: Generar un nombre descriptivo en kebab-case (ej: `no-console-log-ts`)

2. **Generar el archivo** en `~/.opencode/hookify.{nombre}.local.md` con este formato:

```markdown
---
name: {nombre}
description: {descripción corta}
event: {evento}
enabled: true
priority: {baja|media|alta}
---

## Condición

{condición en lenguaje natural clara}

## Acción

{acción a ejecutar: block|warn|log|transform}

## Mensaje

{mensaje que verá el usuario cuando se dispare la regla}
```

3. **Confirmar** al usuario:
   - Nombre del archivo creado
   - Resumen de cuándo se disparará
   - Cómo desactivarla (`enabled: false` en el frontmatter)
   - Cómo eliminarla (`rm ~/.opencode/hookify.{nombre}.local.md`)

## Ejemplos de peticiones

| Petición | Evento | Condición | Acción |
|----------|--------|-----------|--------|
| "Avísame cuando use console.log en TypeScript" | PostToolUse | Archivo `.ts` o `.tsx` contiene `console.log` | warn |
| "Bloquea que se borren archivos de test" | PreToolUse | Herramienta Bash con `rm` sobre `*test*` o `*spec*` | block |
| "Registra cuando se modifiquen migraciones" | PostToolUse | Archivo modificado bajo `migrations/` | log |
| "No permitas commits con 'WIP' en el mensaje" | PreToolUse | Herramienta Bash con `git commit` y mensaje contiene `WIP` | block |

## Validaciones

- Si el nombre generado ya existe, añadir sufijo numérico (`-2`, `-3`, ...)
- Si no se puede determinar el evento, preguntar al usuario
- Si la condición es ambigua, mostrar interpretación y pedir confirmación antes de crear
