---
name: Hookify - Hooks Configurables
description: Esta skill debe usarse cuando el usuario pide "crear hooks personalizados", "prevenir comportamientos no deseados", "configurar reglas de hook", "hookify", "reglas de validación", o quiere crear hooks sin editar archivos hooks.json complejos. Crea reglas markdown con patrones regex y mensajes de advertencia.
version: 0.1.0
---

# Hookify - Framework de Hooks Configurables

## Visión General

Hookify permite crear hooks fácilmente sin editar archivos `hooks.json` complejos. En su lugar, se crean archivos de configuración markdown ligeros que definen patrones a vigilar y mensajes a mostrar cuando esos patrones coinciden.

**Características clave:**
- Analizar conversaciones para encontrar comportamientos no deseados automáticamente
- Archivos de configuración markdown simples con frontmatter YAML
- Coincidencia por patrones regex para reglas potentes
- Sin código requerido - solo describir el comportamiento
- Activar/desactivar fácilmente sin reiniciar

## Inicio Rápido

### 1. Crear Primera Regla

```
/hookify Avisarme cuando use comandos rm -rf
```

Esto analiza la petición y crea `.claude/hookify.warn-rm.local.md`.

### 2. Probar Inmediatamente

**¡Sin reiniciar!** Las reglas toman efecto en la siguiente herramienta.

## Formato de Configuración de Reglas

### Regla Simple (Patrón Único)

`.claude/hookify.dangerous-rm.local.md`:
```markdown
---
name: block-dangerous-rm
enabled: true
event: bash
pattern: rm\s+-rf
action: block
---

⚠️ **¡Comando rm peligroso detectado!**

Este comando podría borrar archivos importantes. Por favor:
- Verificar que la ruta es correcta
- Considerar un enfoque más seguro
- Asegurar que hay copias de seguridad
```

**Campo action:**
- `warn`: Muestra advertencia pero permite operación (por defecto)
- `block`: Previene operación (PreToolUse) o detiene sesión (Stop)

### Regla Avanzada (Múltiples Condiciones)

`.claude/hookify.sensitive-files.local.md`:
```markdown
---
name: warn-sensitive-files
enabled: true
event: file
action: warn
conditions:
  - field: file_path
    operator: regex_match
    pattern: \.env$|credentials|secrets
  - field: new_text
    operator: contains
    pattern: KEY
---

🔐 **Edición de archivo sensible detectada!**

Asegurar que las credenciales no están hardcodeadas y el archivo está en .gitignore.
```

**Todas las condiciones deben coincidir** para que la regla se active.

## Tipos de Eventos

- **`bash`**: Se activa en comandos Bash
- **`file`**: Se activa en herramientas Edit, Write, MultiEdit
- **`stop`**: Se activa cuando el agente quiere detenerse
- **`prompt`**: Se activa al enviar un prompt del usuario
- **`all`**: Se activa en todos los eventos

## Sintaxis de Patrones

Usar sintaxis regex de Python:

| Patrón | Coincide con | Ejemplo |
|--------|-------------|---------|
| `rm\s+-rf` | rm -rf | rm -rf /tmp |
| `console\.log\(` | console.log( | console.log("test") |
| `(eval\|exec)\(` | eval( o exec( | eval("code") |
| `\.env$` | archivos terminados en .env | .env, .env.local |
| `chmod\s+777` | chmod 777 | chmod 777 file.txt |

**Consejos:**
- Usar `\s` para espacio en blanco
- Escapar caracteres especiales: `\.` para punto literal
- Usar `|` para OR: `(foo|bar)`
- Usar `.*` para coincidir cualquier cosa
- Poner `action: block` para operaciones peligrosas
- Poner `action: warn` (o omitir) para advertencias informativas

## Operadores

- `regex_match`: El patrón debe coincidir (más común)
- `contains`: El string debe contener el patrón
- `equals`: Coincidencia exacta de string
- `not_contains`: El string NO debe contener el patrón
- `starts_with`: El string empieza con el patrón
- `ends_with`: El string termina con el patrón

## Campos por Tipo de Evento

**Para eventos bash:**
- `command`: El string del comando bash

**Para eventos file:**
- `file_path`: Ruta del archivo siendo editado
- `new_text`: Nuevo contenido siendo añadido
- `old_text`: Viejo contenido siendo reemplazado
- `content`: Contenido del archivo (solo Write)

**Para eventos prompt:**
- `user_prompt`: El texto del prompt del usuario

## Gestión de Reglas

### Activar/Desactivar

**Desactivar temporalmente:** Editar el archivo `.local.md` y poner `enabled: false`
**Reactivar:** Poner `enabled: true`

### Eliminar Reglas

Simplemente borrar el archivo `.local.md`:
```bash
rm .claude/hookify.mi-regla.local.md
```

## Organización de Archivos

**Ubicación:** Todos los reglamentos en directorio `.claude/`
**Nomenclatura:** `.claude/hookify.{nombre-descriptivo}.local.md`
**Gitignore:** Añadir `.claude/*.local.md` a `.gitignore`

**Buenos nombres:**
- `hookify.dangerous-rm.local.md`
- `hookify.console-log.local.md`
- `hookify.require-tests.local.md`

## Flujo de Trabajo

### Crear una Regla

1. Identificar comportamiento no deseado
2. Determinar qué herramienta está involucrada (Bash, Edit, etc.)
3. Elegir tipo de evento (bash, file, stop, etc.)
4. Escribir patrón regex
5. Crear archivo `.claude/hookify.{nombre}.local.md`
6. Probar inmediatamente - las reglas se leen dinámicamente

### Refinar una Regla

1. Editar el archivo `.local.md`
2. Ajustar patrón o mensaje
3. Probar inmediatamente - cambios toman efecto en siguiente uso

### Desactivar una Regla

**Temporal:** Poner `enabled: false` en frontmatter
**Permanente:** Borrar el archivo `.local.md`

## Consejos de Escritura de Regex

**Caracteres especiales necesitan escape:**
- `.` (cualquier char) → `\.` (punto literal)
- `(` `)` → `\(` `\)` (paréntesis literales)

**Meta-caracteres comunes:**
- `\s` - espacio en blanco
- `\d` - dígito (0-9)
- `.` - cualquier carácter
- `+` - uno o más
- `*` - cero o más
- `|` - OR

**Consejo:** Usar patrones sin comillas en YAML para evitar problemas de escape.

## Referencia Rápida

Regla mínima viable:
```markdown
---
name: mi-regla
enabled: true
event: bash
pattern: dangerous_command
---

Mensaje de advertencia aquí
```
