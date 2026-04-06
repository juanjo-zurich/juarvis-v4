---
name: Desarrollo de Comandos
description: Esta skill debe usarse cuando el usuario pide "crear un comando slash", "añadir un comando", "escribir un comando personalizado", "definir argumentos de comando", "usar frontmatter en comandos", "organizar comandos", "comando con referencias a archivos", o necesita orientación sobre estructura de comandos slash, campos YAML frontmatter, argumentos dinámicos, ejecución bash, o mejores prácticas de desarrollo de comandos.
version: 0.2.0
---

# Desarrollo de Comandos

## Visión General

Los comandos slash son prompts usados frecuentemente definidos como archivos Markdown que el agente ejecuta durante sesiones interactivas.

**Conceptos clave:**
- Formato de archivo Markdown para comandos
- Frontmatter YAML para configuración
- Argumentos dinámicos y referencias a archivos
- Ejecución bash para contexto
- Organización de comandos y namespacing

## Concepto Crítico: Comandos son Instrucciones PARA el Agente

**Los comandos se escriben para consumo del agente, no del usuario.**

**Correcto (instrucciones para el agente):**
```markdown
Revisar este código por vulnerabilidades de seguridad incluyendo:
- Inyección SQL
- XSS
- Problemas de autenticación

Proporcionar números de línea específicos y niveles de severidad.
```

**Incorrecto (mensajes al usuario):**
```markdown
Este comando revisará tu código por problemas de seguridad.
Recibirás un informe con detalles de vulnerabilidades.
```

## Ubicaciones de Comandos

**Comandos de proyecto** (compartidos con equipo):
- Ubicación: `.juarvis/commands/`
- Alcance: Disponible en proyecto específico

**Comandos personales** (disponibles en todas partes):
- Ubicación: `~/.juarvis/commands/`
- Alcance: Disponible en todos los proyectos

**Comandos de plugin** (empaquetados con plugins):
- Ubicación: `plugin-name/commands/`
- Alcance: Disponible cuando plugin está instalado

## Formato de Archivo

### Estructura Básica

```markdown
---
description: Descripción del comando
allowed-tools: Read, Grep, Bash(git:*)
model: sonnet
---

Instrucciones del comando...
```

## Campos de Frontmatter YAML

### description

**Propósito:** Descripción breve mostrada en `/help`
**Tipo:** String

```yaml
---
description: Revisar PR por calidad de código
---
```

### allowed-tools

**Propósito:** Especificar qué herramientas puede usar el comando
**Tipo:** String o Array

```yaml
---
allowed-tools: Read, Write, Edit, Bash(git:*)
---
```

**Patrones:**
- `Read, Write, Edit` - Herramientas específicas
- `Bash(git:*)` - Solo comandos git en Bash
- `*` - Todas las herramientas

### model

**Propósito:** Especificar modelo para ejecución del comando
**Tipo:** String (sonnet, opus, haiku)

```yaml
---
model: haiku
---
```

### argument-hint

**Propósito:** Documentar argumentos esperados para autocompletado
**Tipo:** String

```yaml
---
argument-hint: [pr-number] [priority] [assignee]
---
```

## Argumentos Dinámicos

### Usando $ARGUMENTS

Capturar todos los argumentos como string único:

```markdown
---
argument-hint: [issue-number]
---

Corregir issue #$ARGUMENTS siguiendo nuestros estándares de código.
```

### Argumentos Posicionales

Capturar argumentos individuales con `$1`, `$2`, `$3`:

```markdown
---
argument-hint: [pr-number] [priority] [assignee]
---

Revisar pull request #$1 con nivel de prioridad $2.
Tras revisión, asignar a $3 para seguimiento.
```

## Referencias a Archivos

### Usando Sintaxis @

Incluir contenido de archivos en comando:

```markdown
---
argument-hint: [file-path]
---

Revisar @$1 por:
- Calidad de código
- Mejores prácticas
- Bugs potenciales
```

### Referencias Estáticas

Referenciar archivos conocidos sin argumentos:

```markdown
Revisar @package.json y @tsconfig.json por consistencia.
```

## Ejecución Bash en Comandos

Los comandos pueden ejecutar bash inline para recopilar contexto dinámico:

**Usos:**
- Incluir contexto dinámico (git status, variables de entorno)
- Recopilar estado del proyecto/repositorio
- Construir flujos de trabajo conscientes del contexto

## Organización de Comandos

### Estructura Plana

```
.juarvis/commands/
├── build.md
├── test.md
├── deploy.md
└── review.md
```

### Estructura con Namespace

```
.juarvis/commands/
├── ci/
│   ├── build.md
│   └── test.md
├── git/
│   ├── commit.md
│   └── pr.md
└── docs/
    └── generate.md
```

## Características Específicas de Plugin

### Variable PLUGIN_ROOT

Los comandos de plugin tienen acceso a `${PLUGIN_ROOT}`:

```markdown
---
allowed-tools: Bash(node:*)
---

Ejecutar análisis: !`node ${PLUGIN_ROOT}/scripts/analyze.js $1`
```

**Patrones comunes:**
```markdown
# Ejecutar script del plugin
!`bash ${PLUGIN_ROOT}/scripts/script.sh`

# Cargar configuración del plugin
@${PLUGIN_ROOT}/config/settings.json

# Usar plantilla del plugin
@${PLUGIN_ROOT}/templates/report.md
```

## Patrones Comunes

### Patrón de Revisión

```markdown
---
description: Revisar cambios de código
allowed-tools: Read, Bash(git:*)
---

Archivos cambiados: !`git diff --name-only`

Revisar cada archivo por:
1. Calidad de código y estilo
2. Bugs potenciales o problemas
3. Cobertura de tests
4. Necesidades de documentación
```

### Patrón de Testing

```markdown
---
argument-hint: [test-file]
allowed-tools: Bash(npm:*)
---

Ejecutar tests: !`npm test $1`

Analizar resultados y sugerir correcciones para fallos.
```

## Validación de Comandos

### Validación de Argumentos

```markdown
---
argument-hint: [environment]
---

Validar entorno: !`echo "$1" | grep -E "^(dev|staging|prod)$" || echo "INVÁLIDO"`

Si $1 es entorno válido:
  Desplegar a $1
Si no:
  Explicar entornos válidos: dev, staging, prod
```

## Mejores Prácticas

### Diseño de Comandos
1. **Responsabilidad única:** Un comando, una tarea
2. **Descripciones claras:** Auto-explicativo en `/help`
3. **Dependencias explícitas:** Usar `allowed-tools` cuando necesario
4. **Documentar argumentos:** Siempre proporcionar `argument-hint`
5. **Nomenclatura consistente:** Patrón verbo-sustantivo (review-pr, fix-issue)

### Argumentos
1. Validar argumentos requeridos en prompt
2. Proporcionar valores por defecto
3. Documentar formato esperado
4. Manejar casos límite

## Solución de Problemas

**Comando no aparece:**
- Verificar ubicación correcta con extensión `.md`
- Asegurar formato Markdown válido
- Reiniciar sesión

**Argumentos no funcionan:**
- Verificar sintaxis `$1`, `$2`
- Comprobar que `argument-hint` coincide con uso
