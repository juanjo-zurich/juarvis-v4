---
name: "Agent Hook - Security Validator"
description: >
  Usa un agente para evaluar si una operación es segura antes de ejecutarla.
event: PreToolUse
matcher: "Bash|Bash(git push|Write|Edit"
action: agent
model: "claude-3-opus"
prompt_file: "./prompts/security-validator.md"
---
# Agent-Based Hooks

## Descripción
Los **Agent Hooks** usan un Agente IA para evaluar condiciones complejas antes de permitir o denegar una operación.

A diferencia de los command hooks que ejecutan scripts simples, los agent hooks pueden:
- Analizar contexto semántico
- Entender intenciones del código
- Tomar decisiones basadas en políticas complejas

## Estructura

```yaml
---
name: "Nombre del hook"
event: PreToolUse | PostToolUse | PostToolUseFailure | SessionStart | SessionEnd
matcher: "Patrón para coincidir"
type: agent
model: "claude-3-opus"  # Modelo a usar
prompt: |
  Eres un guardián de seguridad.
  Evalúa si el siguiente comando es seguro:
  
  Comando: $TOOL_NAME
  Args: $ARGUMENTS
  
  Responde ONLY: ALLOW o DENY con razón.
timeout_ms: 5000
retry: 2
---
```

## Ejemplo: Validar Git Push

```yaml
---
name: "Validate git push"
event: PreToolUse
matcher: "Bash(git push"
type: agent
model: "claude-3-sonnet"
prompt: |
  Eres un revisor de código.
  
  El agente intenta hacer 'git push'. Antes de permitirlo:
  1. Verifica que los tests pasen
  2. Verifica que no hay secretos en staging
  3. Verifica que el mensaje de commit sigue convenciones
  
  Comandos a verificar:
  $ARGUMENTS
  
  Responder en JSON:
  {"decision": "ALLOW|DENY", "reason": "..."}
timeout_ms: 10000
---
```

## Ejemplo: Validar Código Antes de Escribir

```yaml
---
name: "Validate code before write"
event: PreToolUse
matcher: "Write|Edit"
type: agent
model: "claude-3-opus"
prompt: |
  Analiza el código que el agente quiere escribir.
  
  Herramienta: $TOOL_NAME
  Archivo: $FILE_PATH
  Contenido: $ARGUMENTS.new_text
  
  Evalúa:
  1. ¿El código tiene senseionales?
  2. ¿Sigue los patrones del proyecto?
  3. ¿Es código real o placeholder/temporal?
  
  Responder:
  {"decision": "ALLOW|DENY", "reason": "...", "suggestions": [...]}
timeout_ms: 15000
---
```

## Agent Hook para Memory

```yaml
---
name: "Memory context validator"
event: PreToolUse
matcher: "Agent|mcp__memory__.*"
type: agent
model: "claude-3-opus"
prompt: |
  Antes de permitir acceso a memoria:
  
  Petición: $TOOL_NAME $ARGUMENTS
  
  Verifica:
  1. ¿Es una operación de lectura o escritura?
  2. ¿Contiene datos sensibles?
  3. ¿El agente tiene permisos?
  
  Responder: {"decision": "ALLOW|DENY", "reason": "..."}
---
```

## Variables en Agent Hooks

| Variable | Descripción |
|----------|-----------|
| `$TOOL_NAME` | Nombre de herramienta |
| `$ARGUMENTS` | JSON con argumentos completos |
| `$ARGUMENTS.command` | Comando bash |
| `$ARGUMENTS.file_path` | Path de archivo |
| `$ARGUMENTS.new_text` | Nuevo texto a escribir |
| `$SESSION_ID` | ID de sesión actual |
| `$PROJECT_ROOT` | Raíz del proyecto |
| `$USER` | Usuario ejecutando |
| `$PREVIOUS_TOOL` | Herramienta anterior |
| `$TOOL_HISTORY` | JSON con últimas 5 herramientas |

## Integración con MCP

Los Agent Hooks pueden usar herramientas MCP externas:

```yaml
---
name: "Check against security database"
event: PreToolUse  
matcher: "Bash|npm install|pip install"
type: agent
model: "claude-3-opus"
mcp_tools:
  - "security-db-check"
  - "vulnerability-scan"
prompt: |
  Usa las herramientas MCP de seguridad para verificar este comando.
  
  Comando: $ARGUMENTS.command
  
  1. security-db-check: Verificar si es comando conocido
  2. vulnerability-scan: Analizar código si es necesario
  
  Responder: {"decision": "ALLOW|DENY", "risk_level": "LOW|MEDIUM|HIGH"}
---
```

## Configuración de Agent Hooks

```yaml
# En permissions.yaml
agent_hooks:
  enabled: true
  default_model: "claude-3-opus"
  timeout_ms: 10000
  max_retries: 2
  cache_enabled: true
  cache_ttl_seconds: 300
  
  # Modelos por evento
  models:
    PreToolUse: "claude-3-opus"
    PostToolUse: "claude-3-sonnet"
    SessionStart: "claude-3-haiku"
```

## Mejores Prácticas

1. **Timeouts cortos**: Los hooks bloquean operaciones, usa `timeout_ms: 5000-10000`
2. **Cache respuestas**: Para comandos repetitivos, el cache evita latencia
3. **Fallback to deny**: Ante duda, mejor denegar
4. **Logs detallado**: Siempre logarithmic decisiones para auditoría
5. **No usar en producción crítica**: Los agent hooks add latencia

## Latencia Esperada

| Tipo Hook | Latencia Típica |
|----------|----------------|
| Command | 10-100ms |
| HTTP | 100-500ms |
| Agent | 2000-10000ms |