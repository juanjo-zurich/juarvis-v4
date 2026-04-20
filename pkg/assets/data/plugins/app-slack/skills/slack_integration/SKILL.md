---
name: slack_integration
description: >
  Integración con Slack para mensajes, canales y automatización.
  Trigger: /slack, "envía a Slack", "notifica en canal"
license: MIT
metadata:
  author: juarvis-org
  version: "1.0"
---

# Slack Integration Skill

## Propósito
Conecta Juarvis con Slack para:
- Enviar mensajes a canales y usuarios
- Leer historial de canales
- Crear canales y threads
- Automatizar notificaciones

##Herramientas MCP

| Herramienta | Descripción |
|-----------|-----------|
| `send_message` | Enviar mensaje a canal/usuario |
| `list_channels` | Listar canales disponibles |
| `history` | Obtener historial de canal |
| `create_channel` | Crear nuevo canal |
| `reply` | Responder en thread |

## Uso

### Enviar Notificación
```
1. Identificar canal destino (#general, #dev, etc.)
2. Craft mensaje con contexto
3. send_message al canal
```

### Reporte Automático
```
1. Recopilar información relevante
2. Formatear como mensaje Slack (blocks)
3. Enviar a canal de reportes
```

### Alerta de Errores
```
1. Detectar error en build/test
2. Enviar mensaje a #alerts con contexto
3. Incluir link a logs
```

## Workflows

### Workflow 1: Build Notification
```
1. Ejecutar build
2. Si pasa: enviar "✅ Build exitoso" a #builds
3. Si falla: enviar "❌ Build falló" con errores
```

### Workflow 2: PR Review Alert
```
1. Recibir notificación de nuevo PR
2. Revisar cambios
3. Enviar summary a #pr-reviews
```

### Workflow 3: Daily Summary
```
1. Recopilar métricas del día
2. Formatear como Slack blocks
3. Enviar a #standup
```

## Configuración
```
SLACK_BOT_TOKEN=xoxb-...
SLACK_TEAM_ID=T01234...
```

## Notas
- Usar Slack Blocks para mensajes enriquecidos
- Mentionar usuarios con <@USERNAME>
- Usar canales apropiados por tema
- No spammear - solo notificaciones relevantes