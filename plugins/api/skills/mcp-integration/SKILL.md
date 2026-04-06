---
name: Integración MCP
description: Esta skill debe usarse cuando el usuario pide "añadir servidor MCP", "integrar MCP", "configurar MCP en plugin", "usar .mcp.json", "Model Context Protocol", "conectar servicio externo", o discute tipos de servidor MCP (SSE, stdio, HTTP, WebSocket). Proporciona orientación para integrar servidores MCP en plugins.
version: 0.1.0
---

# Integración MCP para Plugins

## Visión General

Model Context Protocol (MCP) permite que los plugins se integren con servicios y APIs externos proporcionando acceso estructurado a herramientas.

**Capacidades clave:**
- Conectar a servicios externos (bases de datos, APIs, sistemas de archivos)
- Proporcionar 10+ herramientas relacionadas desde un solo servicio
- Manejar OAuth y flujos de autenticación complejos
- Empaquetar servidores MCP con plugins para configuración automática

## Métodos de Configuración del Servidor MCP

### Método 1: .mcp.json Dedicado (Recomendado)

Crear `.mcp.json` en la raíz del plugin:

```json
{
  "database-tools": {
    "command": "${PLUGIN_ROOT}/servers/db-server",
    "args": ["--config", "${PLUGIN_ROOT}/config.json"],
    "env": {
      "DB_URL": "${DB_URL}"
    }
  }
}
```

### Método 2: Inline en plugin.json

Añadir campo `mcpServers` al manifiesto:

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "mcpServers": {
    "plugin-api": {
      "command": "${PLUGIN_ROOT}/servers/api-server",
      "args": ["--port", "8080"]
    }
  }
}
```

## Tipos de Servidor MCP

### stdio (Proceso Local)

Ejecutar servidores MCP locales como procesos hijo. Mejor para herramientas locales y servidores personalizados.

**Configuración:**
```json
{
  "filesystem": {
    "command": "npx",
    "args": ["-y", "@modelcontextprotocol/server-filesystem", "/ruta/permitida"],
    "env": {
      "LOG_LEVEL": "debug"
    }
  }
}
```

### SSE (Server-Sent Events)

Conectar a servidores MCP alojados con soporte OAuth. Mejor para servicios en la nube.

**Configuración:**
```json
{
  "asana": {
    "type": "sse",
    "url": "https://mcp.asana.com/sse"
  }
}
```

**Autenticación:** Flujos OAuth manejados automáticamente, usuario autenticado en navegador en primer uso.

### HTTP (API REST)

Conectar a servidores MCP RESTful con autenticación por token.

**Configuración:**
```json
{
  "api-service": {
    "type": "http",
    "url": "https://api.example.com/mcp",
    "headers": {
      "Authorization": "Bearer ${API_TOKEN}"
    }
  }
}
```

### WebSocket (Tiempo Real)

Conectar a servidores MCP WebSocket para comunicación bidireccional en tiempo real.

**Configuración:**
```json
{
  "realtime-service": {
    "type": "ws",
    "url": "wss://mcp.example.com/ws",
    "headers": {
      "Authorization": "Bearer ${TOKEN}"
    }
  }
}
```

## Expansión de Variables de Entorno

**${PLUGIN_ROOT}** - Directorio del plugin (siempre usar para portabilidad):
```json
{
  "command": "${PLUGIN_ROOT}/servers/my-server"
}
```

**Variables de entorno del usuario** - Del shell del usuario:
```json
{
  "env": {
    "API_KEY": "${MY_API_KEY}",
    "DATABASE_URL": "${DB_URL}"
  }
}
```

## Nomenclatura de Herramientas MCP

Las herramientas MCP se prefijan automáticamente:

**Formato:** `mcp__plugin_<nombre-plugin>_<nombre-servidor>__<nombre-herramienta>`

**Uso en comandos:**
```markdown
---
allowed-tools: [
  "mcp__plugin_asana_asana__asana_create_task",
  "mcp__plugin_asana_asana__asana_search_tasks"
]
---
```

## Patrones de Autenticación

### OAuth (SSE/HTTP)

OAuth manejado automáticamente. El usuario se autentica en navegador en primer uso.

### Basado en Token (Headers)

Tokens estáticos o de variables de entorno:

```json
{
  "type": "http",
  "url": "https://api.example.com",
  "headers": {
    "Authorization": "Bearer ${API_TOKEN}"
  }
}
```

### Variables de Entorno (stdio)

Pasar configuración al servidor MCP:

```json
{
  "command": "python",
  "args": ["-m", "my_mcp_server"],
  "env": {
    "DATABASE_URL": "${DB_URL}",
    "API_KEY": "${API_KEY}",
    "LOG_LEVEL": "info"
  }
}
```

## Patrones de Integración

### Patrón 1: Envoltura Simple de Herramienta

Los comandos usan herramientas MCP con interacción del usuario.

### Patrón 2: Agente Autónomo

Los agentes usan herramientas MCP autónomamente para flujos de trabajo multi-paso.

### Patrón 3: Plugin Multi-Servidor

Integrar múltiples servidores MCP para flujos de trabajo que abarcan varios servicios.

## Mejores Prácticas de Seguridad

- **Usar HTTPS/WSS**: Nunca conexiones inseguras
- **Gestión de Tokens**: Usar variables de entorno, nunca hardcodear
- **Alcance de Permisos**: Pre-permitir solo herramientas MCP necesarias, no comodines

## Referencia Rápida

| Tipo | Transporte | Mejor para | Auth |
|------|-----------|-----------|------|
| stdio | Proceso | Herramientas locales | Variables env |
| SSE | HTTP | Servicios alojados | OAuth |
| HTTP | REST | Backends API | Tokens |
| ws | WebSocket | Tiempo real | Tokens |

## Flujo de Implementación

1. Elegir tipo de servidor MCP (stdio, SSE, HTTP, ws)
2. Crear `.mcp.json` en la raíz del plugin con configuración
3. Usar `${PLUGIN_ROOT}` para todas las referencias de archivos
4. Documentar variables de entorno requeridas en README
5. Probar localmente
6. Pre-permitir herramientas MCP en comandos relevantes
7. Manejar autenticación (OAuth o tokens)
8. Probar casos de error (fallos de conexión, errores de auth)
9. Documentar integración MCP en README del plugin
