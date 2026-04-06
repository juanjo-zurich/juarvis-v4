---
name: Desarrollo de Plugins
description: Esta skill debe usarse cuando el usuario pide "crear un plugin", "desarrollar un plugin", "estructura de plugin", "publicar plugin", "hooks de plugin", "comandos de plugin", "agentes de plugin", o necesita orientación sobre desarrollo completo de plugins con hooks, integración MCP, estructura, y publicación en marketplace.
version: 0.1.0
---

# Kit de Desarrollo de Plugins

## Visión General

Un conjunto de siete skills especializadas para construir plugins de alta calidad:

1. **Desarrollo de Hooks** - API avanzada de hooks y automatización basada en eventos
2. **Integración MCP** - Integración con servidores Model Context Protocol
3. **Estructura de Plugin** - Organización de plugins y configuración del manifiesto
4. **Configuración de Plugin** - Patrones de configuración usando archivos `.local.md`
5. **Desarrollo de Comandos** - Creación de comandos slash con frontmatter y argumentos
6. **Desarrollo de Agentes** - Creación de agentes autónomos con generación asistida
7. **Desarrollo de Skills** - Creación de skills con divulgación progresiva y disparadores fuertes

## Flujo Guiado de Creación

### Proceso de 8 Fases

1. **Descubrimiento** - Entender propósito y requisitos del plugin
2. **Planificación de Componentes** - Determinar skills, comandos, agentes, hooks, MCP necesarios
3. **Diseño Detallado** - Especificar cada componente y resolver ambigüedades
4. **Creación de Estructura** - Configurar directorios y manifiesto
5. **Implementación de Componentes** - Crear cada componente
6. **Validación** - Ejecutar validación y comprobaciones específicas
7. **Pruebas** - Verificar funcionamiento
8. **Documentación** - Finalizar README y preparar distribución

## Flujo de Desarrollo

```
Diseñar Estructura  →  skill de estructura de plugin
    ↓
Añadir Componentes  →  Todas las skills proporcionan orientación
    ↓
Integrar Servicios  →  skill de integración MCP
    ↓
Añadir Automatización →  skill de desarrollo de hooks
    ↓
Probar y Validar    →  utilidades de validación
```

## Divulgación Progresiva

Cada skill utiliza un sistema de tres niveles:
1. **Metadatos** (siempre cargados): Descripciones concisas con disparadores fuertes
2. **SKILL.md principal** (al activarse): Referencia esencial (~1.500-2.000 palabras)
3. **Referencias/Ejemplos** (según necesidad): Guías detalladas y código funcional

## Uso del Plugin Root

Usar `${PLUGIN_ROOT}` para todas las rutas internas del plugin. Nunca usar rutas absolutas codificadas.

**Dónde usar:**
- Rutas de comandos en hooks
- Argumentos de comandos del servidor MCP
- Referencias a scripts
- Rutas de archivos de recursos

## Skills Disponibles

Consultar cada skill individual para documentación detallada:
- `hook-development` - Desarrollo de hooks basados en eventos
- `mcp-integration` - Integración con servidores MCP
- `plugin-structure` - Organización y manifiestos de plugins
- `plugin-settings` - Patrones de configuración
- `command-development` - Comandos slash
- `agent-development` - Agentes autónomos
- `skill-development` - Creación de skills

## Mejores Prácticas

- **Seguridad primero**: Validación de entradas en hooks, HTTPS para servidores MCP, variables de entorno para credenciales
- **Portabilidad**: Usar `${PLUGIN_ROOT}` en todas partes, solo rutas relativas
- **Pruebas**: Validar configuraciones, probar hooks con entradas de ejemplo
- **Documentación**: README claros, variables de entorno documentadas, ejemplos de uso
