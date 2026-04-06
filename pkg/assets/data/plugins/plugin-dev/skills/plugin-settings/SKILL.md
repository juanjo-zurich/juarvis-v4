---
name: Configuración de Plugin
description: Esta skill debe usarse cuando el usuario pregunta sobre "configuración de plugin", "guardar configuración de plugin", "archivos .local.md", "archivos de estado de plugin", "leer frontmatter YAML", "configuración por proyecto", o quiere hacer el comportamiento del plugin configurable. Documenta el patrón `.opencode/plugin-name.local.md`.
version: 0.1.0
---

# Patrón de Configuración de Plugin

## Visión General

Los plugins pueden guardar configuración y estado en archivos `.opencode/plugin-name.local.md` dentro del directorio del proyecto. Este patrón usa frontmatter YAML para configuración estructurada y contenido markdown para prompts o contexto adicional.

**Características clave:**
- Ubicación del archivo: `.opencode/nombre-plugin.local.md` en raíz del proyecto
- Estructura: Frontmatter YAML + cuerpo markdown
- Propósito: Configuración y estado del plugin por proyecto
- Uso: Leído desde hooks, comandos y agentes
- Ciclo de vida: Gestionado por el usuario (no en git, debe estar en `.gitignore`)

## Estructura del Archivo

### Plantilla Básica

```markdown
---
enabled: true
setting1: valor1
setting2: valor2
numeric_setting: 42
list_setting: ["item1", "item2"]
---

# Contexto Adicional

Este cuerpo markdown puede contener:
- Descripciones de tareas
- Instrucciones adicionales
- Prompts para devolver al agente
- Documentación o notas
```

## Lectura de Archivos de Configuración

### Desde Hooks (Scripts Bash)

**Patrón: Comprobar existencia y parsear frontmatter**

```bash
#!/bin/bash
set -euo pipefail

STATE_FILE=".opencode/mi-plugin.local.md"

if [[ ! -f "$STATE_FILE" ]]; then
  exit 0  # Plugin no configurado, salir
fi

# Parsear frontmatter YAML (entre marcadores ---)
FRONTMATTER=$(sed -n '/^---$/,/^---$/{ /^---$/d; p; }' "$STATE_FILE")

# Extraer campos individuales
ENABLED=$(echo "$FRONTMATTER" | grep '^enabled:' | sed 's/enabled: *//' | sed 's/^"\(.*\)"$/\1/')

if [[ "$ENABLED" != "true" ]]; then
  exit 0  # Deshabilitado
fi
```

### Desde Comandos

Los comandos pueden leer archivos de configuración para personalizar comportamiento. Leer el archivo con Read tool y parsear el frontmatter YAML.

### Desde Agentes

Los agentes pueden referenciar configuración en sus instrucciones, adaptando comportamiento según los campos del frontmatter.

## Técnicas de Parsing

### Extraer Frontmatter

```bash
FRONTMATTER=$(sed -n '/^---$/,/^---$/{ /^---$/d; p; }' "$FILE")
```

### Leer Campos Individuales

**Campos de texto:**
```bash
VALUE=$(echo "$FRONTMATTER" | grep '^field_name:' | sed 's/field_name: *//' | sed 's/^"\(.*\)"$/\1/')
```

**Campos booleanos:**
```bash
ENABLED=$(echo "$FRONTMATTER" | grep '^enabled:' | sed 's/enabled: *//')
```

**Campos numéricos:**
```bash
MAX=$(echo "$FRONTMATTER" | grep '^max_value:' | sed 's/max_value: *//')
```

### Leer Cuerpo Markdown

Extraer contenido después del segundo `---`:

```bash
BODY=$(awk '/^---$/{i++; next} i>=2' "$FILE")
```

## Patrones Comunes

### Patrón 1: Hooks Activos Temporalmente

Usar archivo de configuración para controlar activación de hooks:

```bash
#!/bin/bash
STATE_FILE=".opencode/security-scan.local.md"

if [[ ! -f "$STATE_FILE" ]]; then
  exit 0
fi

FRONTMATTER=$(sed -n '/^---$/,/^---$/{ /^---$/d; p; }' "$STATE_FILE")
ENABLED=$(echo "$FRONTMATTER" | grep '^enabled:' | sed 's/enabled: *//')

if [[ "$ENABLED" != "true" ]]; then
  exit 0
fi
# Ejecutar lógica del hook
```

### Patrón 2: Gestión de Estado de Agentes

Almacenar estado específico del agente y configuración (nombre de agente, número de tarea, sesión coordinador, etc.)

### Patrón 3: Comportamiento Dirigido por Configuración

Usar campos de configuración para controlar nivel de validación, tamaño máximo de archivos, extensiones permitidas, etc.

## Creación de Archivos de Configuración

### Desde Comandos

Los comandos pueden crear archivos de configuración preguntando preferencias al usuario y guardando en formato YAML + markdown.

### Generación de Plantilla

Proporcionar plantilla en README del plugin:

```markdown
## Configuración

Crear `.opencode/mi-plugin.local.md` en tu proyecto:

\`\`\`markdown
---
enabled: true
mode: standard
max_retries: 3
---

# Configuración del Plugin

Tu configuración está activa.
\`\`\`

Tras crear o editar, reiniciar para que los cambios surtan efecto.
```

## Mejores Prácticas

### Nombres de Archivo

- Usar formato `.opencode/plugin-name.local.md`
- Coincidir nombre del plugin exactamente
- Usar sufijo `.local.md` para archivos locales del usuario

### Gitignore

Siempre añadir a `.gitignore`:

```gitignore
.opencode/*.local.md
.opencode/*.local.json
```

### Valores por Defecto

Proporcionar valores por defecto sensatos cuando el archivo de configuración no existe:

```bash
if [[ ! -f "$STATE_FILE" ]]; then
  ENABLED=true
  MODE=standard
else
  # Leer del archivo
fi
```

### Validación

Validar valores de configuración:

```bash
MAX=$(echo "$FRONTMATTER" | grep '^max_value:' | sed 's/max_value: *//')

if ! [[ "$MAX" =~ ^[0-9]+$ ]] || [[ $MAX -lt 1 ]] || [[ $MAX -gt 100 ]]; then
  echo "⚠️  max_value inválido en configuración (debe ser 1-100)" >&2
  MAX=10  # Usar valor por defecto
fi
```

## Consideraciones de Seguridad

### Sanitizar Entrada del Usuario

Al escribir archivos de configuración desde entrada del usuario, escapar comillas y validar rutas de archivo.

### Permisos

Los archivos de configuración deben ser:
- Legibles solo por el usuario (`chmod 600`)
- No commiteados a git
- No compartidos entre usuarios

## Referencia Rápida

### Parsing de Frontmatter

```bash
FRONTMATTER=$(sed -n '/^---$/,/^---$/{ /^---$/d; p; }' "$FILE")
VALUE=$(echo "$FRONTMATTER" | grep '^field:' | sed 's/field: *//' | sed 's/^"\(.*\)"$/\1/')
```

### Parsing del Cuerpo

```bash
BODY=$(awk '/^---$/{i++; next} i>=2' "$FILE")
```

### Patrón de Salida Rápida

```bash
if [[ ! -f ".opencode/mi-plugin.local.md" ]]; then
  exit 0  # No configurado
fi
```

## Flujo de Implementación

1. Diseñar esquema de configuración (qué campos, tipos, valores por defecto)
2. Crear plantilla en documentación del plugin
3. Añadir entrada gitignore para `.opencode/*.local.md`
4. Implementar parsing de configuración en hooks/comandos
5. Usar patrón de salida rápida (comprobar existencia, campo enabled)
6. Documentar configuración en README con plantilla
7. Recordar que los cambios requieren reinicio
