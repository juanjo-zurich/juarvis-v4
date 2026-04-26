# Arquitectura de Seguridad - Juarvis

## Resumen

Juarvis implementa **tres capas de seguridad** que trabajan en cascada:

```
┌─────────────────────────────────────────────────────────────────┐
│                    CAPA 3: Hookify                            │
│  (patterns complejos, fixers automáticos, acciones post-tool)  │
├─────────────────────────────────────────────────────────────────┤
│                    CAPA 2: Permissions.yaml                   │
│     (reglas de proyecto, audit logs, rate limits)          │
├─────────────────────────────────────────────────────────────────┤
│                    CAPA 1: Sandbox/Guard                     │
│     (workspace boundary, blacklist de comandos)         │
└─────────────────────────────────────────────────────────────────┘
```

## Capa 1: Sandbox/Guard (Más Baja)

**Propósito:** Seguridad fundamental - delimitar dónde se puede ejecutar commands.

**Archivos:**
- `pkg/sandbox/guard.go` - Lógica principal
- `cmd/guard.go` - Command CLI (`juarvis guard`)

**Qué hace:**
- Verifica que commands no salgan del workspace
- Mantiene blacklist de commands peligrosos (`rm -rf /`, `dd`, etc.)
- Limita acceso a paths sensibles (`~/.ssh`, `/etc`, etc.)
- stats de ejecuciones

**Cuándo se usa:**
- Por defecto en todas las ejecuciones CLI
- `juarvis sandbox run <cmd>` - modo explícito

## Capa 2: Permissions.yaml (Proyecto)

**Propósito:** Reglas específicas del proyecto.

**Archivos:**
- `permissions.yaml` - en raíz del proyecto
- `cmd/guard.go` - evalúa las reglas

**Qué hace:**
- Permite/bloquea commands por pattern
- Incluye reasons para transparencia
- Implementa rate limits
- Audit logs de todas las ejecuciones

**Ejemplo `permissions.yaml`:**
```yaml
version: "1.0"
rules:
  bash:
    - pattern: "git push.*--force"
      action: deny
      reason: "Force push destruye historia"
    - pattern: "rm -rf"
      action: warn
      reason: "Comando destructivo"
limits:
  bash_per_minute: 30
  api_per_hour: 100
audit:
  enabled: true
  log_file: .juar/audit.log
```

## Capa 3: Hookify (Más Alta)

**Propósito:** Automatización avanzada basada en patterns.

**Archivos:**
- `skills/hookify/**/*.md` - reglas hookify
- `pkg/hookify/` - motor de evaluación
- Hooks en `plugins/*/hooks/`

**Qué hace:**
- Pattern matching complejo (regex, wildcards)
- Fixers automáticos (auto-format, auto-import)
- Acciones post-ejecución
- Pre-ejecución (prompt de confirmación)

**Ejemplo:**
```markdown
---
name: auto-format
trigger: "**.go"
on: post-tooluse
action: gofmt -w {{file}}
---
```

## Orden de Evaluación

```
1. Sandbox (Capa 1)
   ↓ Si permitido
2. Permissions.yaml (Capa 2)
   ↓ Si permitido
3. Hookify (Capa 3)
```

**Nota:** En el futuro, estas capas podrían unificarse en un solo flujo:
```go
func EvaluateSecurity(cmd string, args []string) (bool, error) {
    // Capa 1: Sandbox check
    if !sandbox.IsAllowed(cmd) {
        return false, fmt.Errorf("blocked by sandbox: %s", cmd)
    }

    // Capa 2: Permissions.yaml
    if !permissions.Allowed(cmd) {
        return false, fmt.Errorf("blocked by permissions.yaml: %s", cmd)
    }

    // Capa 3: Hookify (acciones automático)
    hooks.Eval(cmd)

    return true, nil
}
```

## Comandos CLI

```bash
# Sandbox
juarvis sandbox run "npm install express"
juarvis sandbox check "git push"
juarvis sandbox stats

# Permissions
juarvis guard run "git push --force"
juarvis guard allow "npm *"

# Hookify (plugins)
juarvis hookify list
```

## Configuración Recomendada

| Nivel de Seguridad | Configuración |
|-------------------|---------------|
| Estricto | sandbox + permissions + hookify |
| Medio | sandbox + permissions |
| Básico | sandbox solo |