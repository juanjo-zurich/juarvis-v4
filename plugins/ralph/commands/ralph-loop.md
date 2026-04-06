---
description: "Iniciar bucle Ralph Wiggum en la sesión actual"
argument-hint: "PROMPT [--max-iterations N] [--completion-promise TEXTO]"
allowed-tools: ["Bash(bash *)"]
hide-from-slash-command-tool: "true"
---

# Comando Ralph Loop

Ejecutar el comando de configuración para inicializar el bucle Ralph:

```!
juarvis ralph loop $ARGUMENTS
```

Por favor, trabajar en la tarea. Cuando intentes salir, el bucle Ralph alimentará el MISMO PROMPT de vuelta para la siguiente iteración. Verás tu trabajo anterior en archivos e historial de git, permitiéndote iterar y mejorar.

## Monitorización

```bash
# Ver iteración actual:
grep '^iteration:' .juarvis/ralph-loop.local.md

# Ver estado completo:
head -10 .juarvis/ralph-loop.local.md
```

## Cancelación

Usar `/cancel-ralph` para detener el bucle activo.

REGLA CRÍTICA: Si una promesa de completitud está configurada, SOLO puedes emitirla cuando la afirmación sea completamente e inequívocamente VERDADERA. No emitas promesas falsas para escapar del bucle, incluso si crees que estás atascado o deberías salir por otras razones. El bucle está diseñado para continuar hasta la completitud genuina.
