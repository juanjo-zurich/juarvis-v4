---
name: Validador de Plugins
description: Esta skill debe usarse cuando el usuario pide "validar mi plugin", "comprobar estructura de plugin", "verificar plugin", "validar manifiesto", "comprobar archivos de plugin", o menciona validación de plugins. También activar proactivamente tras crear o modificar componentes de plugin.
version: 0.1.0
---

# Validador de Plugins

## Visión General

Especialista en validación comprensiva de estructura, configuración y componentes de plugins. Comprueba la organización, el manifiesto, los componentes y las mejores prácticas.

## Responsabilidades Principales

1. Validar estructura y organización del plugin
2. Comprobar manifiesto plugin.json por corrección
3. Validar todos los archivos de componentes (commands, agents, skills, hooks)
4. Verificar convenciones de nombres y organización de archivos
5. Comprobar problemas comunes y anti-patrones
6. Proporcionar recomendaciones específicas y accionables

## Proceso de Validación

### 1. Localizar Raíz del Plugin

- Buscar `.plugin/plugin.json`
- Verificar estructura de directorios del plugin
- Notar ubicación del plugin

### 2. Validar Manifiesto (`.plugin/plugin.json`)

- Comprobar sintaxis JSON
- Verificar campo requerido: `name`
- Comprobar formato del nombre (kebab-case, sin espacios)
- Validar campos opcionales si presentes:
  - `version`: Formato de versionado semántico (X.Y.Z)
  - `description`: String no vacío
  - `author`: Estructura válida
  - `mcpServers`: Configuraciones de servidor válidas
- Comprobar campos desconocidos (advertir pero no fallar)

### 3. Validar Estructura de Directorios

- Buscar directorios de componentes estándar:
  - `commands/` para comandos slash
  - `agents/` para definiciones de agentes
  - `skills/` para directorios de skills
  - `hooks/hooks.json` para hooks
- Verificar auto-descubrimiento funciona

### 4. Validar Commands

- Buscar `commands/**/*.md`
- Para cada archivo de comando:
  - Comprobar frontmatter YAML presente (empieza con `---`)
  - Verificar campo `description` existe
  - Comprobar formato `argument-hint` si presente
  - Validar `allowed-tools` es array si presente
  - Asegurar contenido markdown existe

### 5. Validar Agents

- Buscar `agents/**/*.md`
- Para cada archivo de agente:
  - Frontmatter con `name`, `description`, `model`, `color`
  - Formato del nombre (minúsculas, guiones, 3-50 chars)
  - Descripción incluye bloques `<example>`
  - Modelo válido (inherit/sonnet/opus/haiku)
  - Color válido (blue/cyan/green/yellow/magenta/red)
  - Prompt de sistema existe y es sustancial

### 6. Validar Skills

- Buscar `skills/*/SKILL.md`
- Para cada directorio de skill:
  - Verificar archivo `SKILL.md` existe
  - Comprobar frontmatter YAML con `name` y `description`
  - Verificar descripción concisa y clara
  - Comprobar subdirectorios references/, examples/, scripts/
  - Validar que archivos referenciados existen

### 7. Validar Hooks

- Si `hooks/hooks.json` existe:
  - Sintaxis JSON válida
  - Nombres de eventos válidos
  - Cada hook tiene `matcher` y array `hooks`
  - Tipo de hook es `command` o `prompt`
  - Comandos referencian scripts existentes con `${PLUGIN_ROOT}`

### 8. Validar Configuración MCP

- Comprobar sintaxis JSON
- Verificar configuraciones de servidor:
  - stdio: tiene campo `command`
  - sse/http/ws: tiene campo `url`
- Comprobar uso de `${PLUGIN_ROOT}` para portabilidad

### 9. Comprobar Organización de Archivos

- README.md existe y es comprensivo
- Sin archivos innecesarios
- .gitignore presente si necesario
- Archivo LICENSE presente

### 10. Comprobaciones de Seguridad

- Sin credenciales hardcodeadas en archivos
- Servidores MCP usan HTTPS/WSS no HTTP/WS
- Hooks no tienen problemas de seguridad obvios
- Sin secretos en archivos de ejemplo

## Formato de Salida

### Informe de Validación de Plugin

#### Plugin: [nombre]
Ubicación: [ruta]

#### Resumen
[Valoración general - pass/fail con estadísticas clave]

#### Problemas Críticos ([count])
- `ruta/archivo` - [Problema] - [Corrección]

#### Advertencias ([count])
- `ruta/archivo` - [Problema] - [Recomendación]

#### Resumen de Componentes
- Commands: [count] encontrados, [count] válidos
- Agents: [count] encontrados, [count] válidos
- Skills: [count] encontrados, [count] válidos
- Hooks: [presente/no presente], [válido/inválido]
- Servidores MCP: [count] configurados

#### Hallazgos Positivos
- [Qué está bien hecho]

#### Recomendaciones
1. [Recomendación prioritaria]
2. [Recomendación adicional]

#### Valoración General
[PASS/FAIL] - [Razonamiento]

## Casos Límite

- **Plugin mínimo** (solo plugin.json): Válido si manifiesto correcto
- **Directorios vacíos**: Advertir pero no fallar
- **Campos desconocidos en manifiesto**: Advertir pero no fallar
- **Múltiples errores de validación**: Agrupar por archivo, priorizar críticos
- **Plugin no encontrado**: Mensaje de error claro con orientación
- **Archivos corruptos**: Saltar y reportar, continuar validación
