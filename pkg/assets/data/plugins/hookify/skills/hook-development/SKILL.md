---
name: Desarrollo de Hooks
description: Esta skill debe usarse cuando el usuario pide "crear un hook", "añadir un hook PreToolUse", "validar uso de herramientas", "implementar hooks basados en prompts", "automatización basada en eventos", "bloquear comandos peligrosos", o menciona eventos de hook (PreToolUse, PostToolUse, Stop, SessionStart, SessionEnd). Proporciona orientación completa para crear hooks de plugins.
version: 0.1.0
---

# Desarrollo de Hooks para Plugins

## Visión General

Los hooks son scripts de automatización basados en eventos que se ejecutan en respuesta a acciones. Usar hooks para validar operaciones, aplicar políticas, añadir contexto e integrar herramientas externas.

**Capacidades clave:**
- Validar llamadas a herramientas antes de ejecutarlas (PreToolUse)
- Reaccionar a resultados de herramientas (PostToolUse)
- Aplicar estándares de completitud (Stop, SubagentStop)
- Cargar contexto de proyecto (SessionStart)
- Automatizar flujos de trabajo en todo el ciclo de desarrollo

## Tipos de Hook

### Hooks Basados en Prompt (Recomendado)

Usar toma de decisiones basada en LLM para validación consciente del contexto:

```json
{
  "type": "prompt",
  "prompt": "Evaluar si este uso de herramienta es apropiado: $TOOL_INPUT",
  "timeout": 30
}
```

**Eventos soportados:** Stop, SubagentStop, UserPromptSubmit, PreToolUse

**Beneficios:**
- Decisiones conscientes del contexto basadas en razonamiento en lenguaje natural
- Lógica de evaluación flexible sin scripts bash
- Mejor manejo de casos límite
- Más fácil de mantener y extender

### Hooks de Comando

Ejecutar comandos bash para comprobaciones deterministas:

```json
{
  "type": "command",
  "command": "bash ${PLUGIN_ROOT}/scripts/validate.sh",
  "timeout": 60
}
```

**Usar para:**
- Validaciones deterministas rápidas
- Operaciones del sistema de archivos
- Integraciones con herramientas externas
- Comprobaciones críticas de rendimiento

## Formatos de Configuración

### Formato Plugin hooks.json

Para hooks de plugin en `hooks/hooks.json`, usar formato envolvente:

```json
{
  "description": "Hooks de validación para calidad de código",
  "hooks": {
    "PreToolUse": [...],
    "Stop": [...],
    "SessionStart": [...]
  }
}
```

### Formato Settings (Directo)

Para configuraciones de usuario en `.opencode/settings.json`, usar formato directo:

```json
{
  "PreToolUse": [...],
  "Stop": [...],
  "SessionStart": [...]
}
```

## Eventos de Hook

### PreToolUse

Ejecutar antes de cualquier herramienta. Usar para aprobar, denegar o modificar llamadas a herramientas.

**Ejemplo (basado en prompt):**
```json
{
  "PreToolUse": [
    {
      "matcher": "Write|Edit",
      "hooks": [
        {
          "type": "prompt",
          "prompt": "Validar seguridad de escritura de archivo. Comprobar: rutas de sistema, credenciales, path traversal, contenido sensible. Devolver 'approve' o 'deny'."
        }
      ]
    }
  ]
}
```

**Salida para PreToolUse:**
```json
{
  "hookSpecificOutput": {
    "permissionDecision": "allow|deny|ask",
    "updatedInput": {"field": "modified_value"}
  },
  "systemMessage": "Explicación"
}
```

### PostToolUse

Ejecutar después de que la herramienta complete. Usar para reaccionar a resultados y proporcionar retroalimentación.

### Stop

Ejecutar cuando el agente principal considera detenerse. Usar para validar completitud.

**Ejemplo:**
```json
{
  "Stop": [
    {
      "matcher": "*",
      "hooks": [
        {
          "type": "prompt",
          "prompt": "Verificar completitud de tarea: tests ejecutados, build exitoso, preguntas respondidas. Devolver 'approve' para parar o 'block' con razón para continuar."
        }
      ]
    }
  ]
}
```

### SessionStart

Ejecutar al inicio de sesión. Usar para cargar contexto y configurar entorno.

**Capacidad especial:** Persistir variables de entorno usando `$ENV_FILE`:
```bash
echo "export PROJECT_TYPE=nodejs" >> "$ENV_FILE"
```

### SessionEnd, PreCompact, Notification, UserPromptSubmit

Ver documentación de referencia para detalles de cada evento.

## Formato de Salida de Hook

### Salida Estándar (Todos los Hooks)

```json
{
  "continue": true,
  "suppressOutput": false,
  "systemMessage": "Mensaje para el agente"
}
```

### Códigos de Salida

- `0` - Éxito (stdout se muestra en transcripción)
- `2` - Error de bloqueo (stderr se devuelve al agente)
- Otro - Error no bloqueante

## Variables de Entorno

Disponibles en todos los hooks de comando:

- `$PROJECT_DIR` - Ruta raíz del proyecto
- `$PLUGIN_ROOT` - Directorio del plugin (usar para rutas portables)
- `$ENV_FILE` - Solo SessionStart: persistir variables de entorno aquí

## Matchers

**Coincidencia exacta:** `"matcher": "Write"`
**Múltiples herramientas:** `"matcher": "Read|Write|Edit"`
**Comodín (todas):** `"matcher": "*"`
**Patrones regex:** `"matcher": "mcp__.*__delete.*"`

## Mejores Prácticas de Seguridad

### Validación de Entrada

Siempre validar entradas en hooks de comando:

```bash
#!/bin/bash
set -euo pipefail

input=$(cat)
tool_name=$(echo "$input" | jq -r '.tool_name')

if [[ ! "$tool_name" =~ ^[a-zA-Z0-9_]+$ ]]; then
  echo '{"decision": "deny", "reason": "Nombre de herramienta inválido"}' >&2
  exit 2
fi
```

### Seguridad de Rutas

Comprobar path traversal y archivos sensibles:

```bash
file_path=$(echo "$input" | jq -r '.tool_input.file_path')

if [[ "$file_path" == *".."* ]]; then
  echo '{"decision": "deny", "reason": "Path traversal detectado"}' >&2
  exit 2
fi
```

## Consideraciones de Rendimiento

Todos los hooks coincidentes se ejecutan **en paralelo**. Diseñar para independencia.

**Optimización:**
1. Usar hooks de comando para comprobaciones deterministas rápidas
2. Usar hooks de prompt para razonamiento complejo
3. Cachear resultados de validación en archivos temporales
4. Minimizar I/O en rutas críticas

## Ciclo de Vida de los Hooks

Los hooks se cargan al inicio de la sesión. Los cambios en la configuración requieren reiniciar la sesión.

## Referencia Rápida

| Evento | Cuándo | Usar para |
|--------|--------|-----------|
| PreToolUse | Antes de herramienta | Validación, modificación |
| PostToolUse | Después de herramienta | Retroalimentación, logging |
| Stop | Agente deteniéndose | Comprobación de completitud |
| SessionStart | Inicio de sesión | Carga de contexto |
| SessionEnd | Fin de sesión | Limpieza, logging |

## Flujo de Implementación

1. Identificar eventos a los que conectar hooks
2. Decidir entre hooks basados en prompt (flexibles) o de comando (deterministas)
3. Escribir configuración en `hooks/hooks.json`
4. Para hooks de comando, crear scripts
5. Usar `${PLUGIN_ROOT}` para todas las referencias de archivos
6. Validar configuración
7. Probar exhaustivamente
8. Documentar en README del plugin
