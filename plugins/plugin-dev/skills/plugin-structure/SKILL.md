---
name: Estructura de Plugin
description: Esta skill debe usarse cuando el usuario pide "crear un plugin", "montar un plugin", "entender estructura de plugin", "organizar componentes de plugin", "configurar manifiesto", "auto-descubrimiento", o necesita orientación sobre diseño de directorios de plugins, configuración de manifiestos, organización de componentes, convenciones de nombres o arquitectura de plugins.
version: 0.1.0
---

# Estructura de Plugins

## Visión General

Los plugins siguen una estructura de directorios estandarizada con descubrimiento automático de componentes.

**Conceptos clave:**
- Diseño convencional de directorios para descubrimiento automático
- Configuración dirigida por manifiesto en `.plugin/plugin.json`
- Organización basada en componentes (commands, agents, skills, hooks)
- Referencias de rutas portables usando `${PLUGIN_ROOT}`
- Carga de componentes explícita vs auto-descubierta

## Estructura de Directorios

```
nombre-plugin/
├── .plugin/
│   └── plugin.json          # Requerido: Manifiesto del plugin
├── commands/                 # Comandos slash (archivos .md)
├── agents/                   # Definiciones de subagentes (archivos .md)
├── skills/                   # Skills del agente (subdirectorios)
│   └── nombre-skill/
│       └── SKILL.md         # Requerido para cada skill
├── hooks/
│   └── hooks.json           # Configuración de manejadores de eventos
├── .mcp.json                # Definiciones de servidor MCP
└── scripts/                 # Scripts auxiliares y utilidades
```

**Reglas críticas:**

1. **Ubicación del manifiesto**: El `plugin.json` DEBE estar en `.plugin/`
2. **Ubicación de componentes**: Todos los directorios DEBE estar en raíz del plugin
3. **Componentes opcionales**: Solo crear directorios para componentes usados
4. **Convención de nombres**: Usar kebab-case para todos los nombres

## Manifiesto del Plugin (plugin.json)

### Campos Requeridos

```json
{
  "name": "nombre-plugin"
}
```

**Requisitos del nombre:**
- Formato kebab-case (minúsculas con guiones)
- Único entre plugins instalados
- Sin espacios ni caracteres especiales

### Metadatos Recomendados

```json
{
  "name": "nombre-plugin",
  "version": "1.0.0",
  "description": "Breve explicación del propósito",
  "author": {
    "name": "Nombre Autor",
    "email": "autor@ejemplo.com"
  },
  "license": "MIT",
  "keywords": ["testing", "automatización"]
}
```

### Configuración de Rutas de Componentes

```json
{
  "name": "nombre-plugin",
  "commands": "./custom-commands",
  "agents": ["./agents", "./specialized-agents"],
  "hooks": "./config/hooks.json",
  "mcpServers": "./.mcp.json"
}
```

## Organización de Componentes

### Commands

**Ubicación:** `commands/`
**Formato:** Archivos Markdown con frontmatter YAML
**Descubrimiento:** Todos los `.md` en `commands/` se cargan automáticamente

### Agents

**Ubicación:** `agents/`
**Formato:** Archivos Markdown con frontmatter YAML
**Descubrimiento:** Todos los `.md` en `agents/` se cargan automáticamente

### Skills

**Ubicación:** `skills/` con subdirectorios por skill
**Formato:** Cada skill en su directorio con `SKILL.md`
**Descubrimiento:** Todos los `SKILL.md` en subdirectorios se cargan automáticamente

### Hooks

**Ubicación:** `hooks/hooks.json` o inline en manifiesto
**Formato:** Configuración JSON definiendo manejadores de eventos

### Servidores MCP

**Ubicación:** `.mcp.json` en raíz del plugin o inline en manifiesto
**Formato:** Configuración JSON para definiciones de servidores MCP

## Convenciones de Nombres

**Commands:** kebab-case `.md` → `code-review.md` → `/code-review`
**Agents:** kebab-case `.md` descriptivo → `test-generator.md`
**Skills:** kebab-case para nombres de directorio → `api-testing/`
**Scripts:** kebab-case con extensiones apropiadas → `validate-input.sh`

## Mecanismo de Auto-Descubrimiento

1. **Manifiesto**: Lee `.plugin/plugin.json` al habilitar plugin
2. **Commands**: Escanea `commands/` para archivos `.md`
3. **Agents**: Escanea `agents/` para archivos `.md`
4. **Skills**: Escanea `skills/` para subdirectorios con `SKILL.md`
5. **Hooks**: Carga configuración de `hooks/hooks.json`
6. **Servidores MCP**: Carga configuración de `.mcp.json`

## Mejores Prácticas

### Organización
- Agrupación lógica de componentes relacionados
- Mantener manifiesto mínimo
- Incluir archivos README

### Portabilidad
- Siempre usar `${PLUGIN_ROOT}`
- Probar en múltiples sistemas
- Documentar dependencias
- Evitar características específicas del sistema

### Mantenimiento
- Versionar consistentemente
- Deprecar con gracia
- Documentar cambios importantes
- Probar exhaustivamente tras cambios

## Patrones Comunes

### Plugin Mínimo
```
mi-plugin/
├── .plugin/
│   └── plugin.json
└── commands/
    └── hello.md
```

### Plugin Completo
```
mi-plugin/
├── .plugin/
│   └── plugin.json
├── commands/
├── agents/
├── skills/
├── hooks/
│   └── hooks.json
├── .mcp.json
└── scripts/
```

## Solución de Problemas

**Componente no carga:**
- Verificar ubicación correcta con extensión correcta
- Comprobar sintaxis del frontmatter YAML
- Asegurar que skill tiene `SKILL.md`
- Confirmar que plugin está habilitado

**Errores de resolución de rutas:**
- Reemplazar todas las rutas hardcodeadas con `${PLUGIN_ROOT}`
- Verificar que rutas son relativas
- Comprobar que archivos referenciados existen

**Auto-descubrimiento no funciona:**
- Confirmar directorios en raíz del plugin
- Verificar convenciones de nombres
- Comprobar rutas personalizadas en manifiesto
