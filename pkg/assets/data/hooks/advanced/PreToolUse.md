---
name: "PreToolUse - Security Guard"
description: >
  Evalúa comandos antes de ejecutarse. Detiene operaciones peligrosas.
event: PreToolUse
matcher: "Bash(rm -rf|Bash(mkfs|Bash(dd if=|Bash(:()|Bash(shutdown|Bash(reboot|Bash(halt"
action: deny
reason: "Comando destructivo detectado"
---
# PreToolUse Hook: Security Guard

## Descripción
Este hook se ejecuta **ANTES** de que cualquier herramienta sea ejecutada por el agente.
Intercepta comandos peligrosos y los bloquea.

## Eventos Soportados

| Evento | Descripción |
|--------|-----------|
| `PreToolUse` | Antes de ejecutar herramienta |
| `PostToolUse` | Después de ejecutar herramienta |
| `PostToolUseFailure` | Cuando falla una herramienta |
| `SessionStart` | Al iniciar sesión |
| `SessionEnd` | Al terminar sesión |

## Matcher Patterns

```bash
# Coincidir comandos bash específicos
Bash(rm -rf
Bash(sudo
Bash(mkfs

# Coincidir herramientas específicas
Write(*.env
Edit(*.env
Read(/etc/passwd)

# Coincidir archivos específicos
Write(**/.env
Edit(**/secrets/**
```

## Tipos de Hook

### 1. Command Hook (Shell)
```yaml
action: command
command: ["/path/to/script.sh", "$ARGUMENTS"]
```

### 2. HTTP Hook
```yaml
action: http
url: "https://api.example.com/validate"
method: POST
```

### 3. Agent Hook (LLM Evaluation)
```yaml
action: agent
prompt: "Evaluar si este comando es seguro: $ARGUMENTS. Responder YES o NO"
model: "claude-3-opus"
```

## Ejemplos de Reglas

### Bloquear comandos destructivos
```yaml
---
name: "Block dangerous commands"
event: PreToolUse
matcher: "Bash(rm -rf|Bash(sudo"
action: deny
reason: "Comando potencialmente destructivo"
---
```

### Validar escritura de secrets
```yaml
---
name: "Prevent secrets in code"
event: PreToolUse
matcher: "Write|Edit"
conditions:
  - field: "new_text"
    operator: "contains"
    pattern: "password|api_key|secret|token"
action: deny
reason: "No escribir secrets en código"
---
```

### Log de todas las operaciones
```yaml
---
name: "Log all operations"
event: PostToolUse
matcher: ".*"
action: command
command: ["juarvis", "audit", "log", "$TOOL_NAME", "$ARGUMENTS"]
---
```

### Snapshot automático antes de cambios
```yaml
---
name: "Auto-snapshot before changes"
event: PreToolUse
matcher: "Write|Edit"
action: command
command: ["juarvis", "snapshot", "auto", "$FILE_PATH"]
---
```

## Variables Disponibles

| Variable | Descripción |
|----------|-----------|
| `$TOOL_NAME` | Nombre de la herramienta |
| `$ARGUMENTS` | JSON con argumentos |
| `$FILE_PATH` | Path del archivo (si aplica) |
| `$SESSION_ID` | ID de sesión |
| `$PROJECT_ROOT` | Raíz del proyecto |

## Configuración Global

Los hooks se configuran en `permissions.yaml` o en archivos separados en `.juarvis/hooks/`.

```yaml
hooks:
  enabled: true
  config_dir: ".juarvis/hooks"
  log_file: ".juarvis/logs/hooks.log"
```
