---
name: Escritura de Reglas Hookify
description: Esta skill debe usarse cuando el usuario pide "crear una regla hookify", "escribir una regla de hook", "configurar hookify", "añadir una regla hookify", o necesita orientación sobre sintaxis y patrones de reglas hookify.
version: 0.1.0
---

# Escritura de Reglas Hookify

## Visión General

Las reglas hookify son archivos markdown con frontmatter YAML que definen patrones a vigilar y mensajes a mostrar cuando esos patrones coinciden. Las reglas se guardan en archivos `.juarvis/hookify.{nombre-regla}.local.md`.

## Formato de Archivo de Regla

### Estructura Básica

```markdown
---
name: identificador-regla
enabled: true
event: bash|file|stop|prompt|all
pattern: regex-patrones-aquí
---

Mensaje a mostrar al agente cuando la regla se active.
Puede incluir formato markdown, advertencias, sugerencias, etc.
```

### Campos de Frontmatter

**name** (requerido): Identificador único para la regla
- Usar kebab-case: `warn-dangerous-rm`, `block-console-log`
- Ser descriptivo y orientado a acción
- Empezar con verbo: warn, prevent, block, require, check

**enabled** (requerido): Booleano para activar/desactivar
- `true`: Regla activa
- `false`: Regla deshabilitada (no se activará)

**event** (requerido): En qué evento de hook se activa
- `bash`: Comandos de herramienta Bash
- `file`: Herramientas Edit, Write, MultiEdit
- `stop`: Cuando el agente quiere detenerse
- `prompt`: Cuando el usuario envía un prompt
- `all`: Todos los eventos

**action** (opcional): Qué hacer cuando la regla coincide
- `warn`: Mostrar advertencia pero permitir operación (por defecto)
- `block`: Prevenir operación o detener sesión

**pattern** (formato simple): Patrón regex para coincidir
- Para reglas simples de condición única
- Coincide contra comando (bash) o new_text (file)

### Formato Avanzado (Múltiples Condiciones)

Para reglas complejas con múltiples condiciones:

```markdown
---
name: warn-env-file-edits
enabled: true
event: file
conditions:
  - field: file_path
    operator: regex_match
    pattern: \.env$
  - field: new_text
    operator: contains
    pattern: API_KEY
---

Estás añadiendo una API key a un archivo .env. ¡Asegúrate de que este archivo está en .gitignore!
```

**Campos de condición:**
- `field`: Qué campo comprobar
  - Para bash: `command`
  - Para file: `file_path`, `new_text`, `old_text`, `content`
- `operator`: Cómo coincidir
  - `regex_match`: Coincidencia de patrón regex
  - `contains`: Comprobación de substring
  - `equals`: Coincidencia exacta
  - `not_contains`: El substring NO debe estar presente
  - `starts_with`: Comprobación de prefijo
  - `ends_with`: Comprobación de sufijo
- `pattern`: Patrón o string a coincidir

**Todas las condiciones deben coincidir para que la regla se active.**

## Cuerpo del Mensaje

El contenido markdown después del frontmatter se muestra al agente cuando la regla se activa.

**Buenos mensajes:**
- Explicar qué se detectó
- Explicar por qué es problemático
- Sugerir alternativas o mejores prácticas
- Usar formato para claridad (negrita, listas, etc.)

**Ejemplo:**
```markdown
⚠️ **¡Console.log detectado!**

Estás añadiendo console.log a código de producción.

**Por qué importa:**
- Los logs de debug no deben ir a producción
- Console.log puede exponer datos sensibles
- Impacta rendimiento del navegador

**Alternativas:**
- Usar una librería de logging apropiada
- Eliminar antes de commitear
- Usar builds de debug condicionales
```

## Guía de Tipos de Evento

### Eventos bash

Coincidir patrones de comandos Bash:

```markdown
---
event: bash
pattern: sudo\s+|rm\s+-rf|chmod\s+777
---

¡Comando peligroso detectado!
```

**Patrones comunes:**
- Comandos peligrosos: `rm\s+-rf`, `dd\s+if=`, `mkfs`
- Escalación de privilegios: `sudo\s+`, `su\s+`
- Problemas de permisos: `chmod\s+777`, `chown\s+root`

### Eventos file

Coincidir operaciones Edit/Write/MultiEdit:

```markdown
---
event: file
pattern: console\.log\(|eval\(|innerHTML\s*=
---

¡Patrón de código potencialmente problemático detectado!
```

**Patrones comunes:**
- Código debug: `console\.log\(`, `debugger`, `print\(`
- Riesgos de seguridad: `eval\(`, `innerHTML\s*=`
- Archivos sensibles: `\.env$`, `credentials`
- Archivos generados: `node_modules/`, `dist/`

### Eventos stop

Coincidir cuando el agente quiere detenerse (comprobaciones de completitud):

```markdown
---
event: stop
pattern: .*
---

Antes de parar, verificar:
- [ ] Tests ejecutados
- [ ] Build exitoso
- [ ] Documentación actualizada
```

## Consejos de Patrones Regex

**Caracteres literales:** La mayoría de caracteres se coinciden a sí mismos
- `rm` coincide con "rm"

**Caracteres especiales necesitan escape:**
- `.` → `\.` (punto literal)
- `(` `)` → `\(` `\)` (paréntesis literales)

**Meta-caracteres comunes:**
- `\s` - espacio en blanco
- `\d` - dígito (0-9)
- `.` - cualquier carácter
- `+` - uno o más
- `*` - cero o más
- `|` - OR

**Probar patrones antes de usar:**
```bash
python3 -c "import re; print(re.search(r'tu_patron', 'texto de prueba'))"
```

## Errores Comunes

**Demasiado amplio:**
```yaml
pattern: log    # Coincide con "log", "login", "dialog", "catalog"
```
Mejor: `console\.log\(|logger\.`

**Demasiado específico:**
```yaml
pattern: rm -rf /tmp  # Solo coincide con ruta exacta
```
Mejor: `rm\s+-rf`

**Problemas de escape:**
- Strings YAML con comillas: `"pattern"` requiere doble escape `\\s`
- YAML sin comillas: `pattern: \s` funciona tal cual
- **Recomendación**: Usar patrones sin comillas en YAML

## Organización de Archivos

**Ubicación:** Todas las reglas en directorio `.juarvis/`
**Nomenclatura:** `.juarvis/hookify.{nombre-descriptivo}.local.md`
**Gitignore:** Añadir `.juarvis/*.local.md` a `.gitignore`

## Referencia Rápida

### Tipos de evento
- `bash` - Comandos Bash
- `file` - Ediciones de archivos
- `stop` - Comprobaciones de completitud
- `prompt` - Entrada del usuario
- `all` - Todos los eventos

### Opciones de campo
- Bash: `command`
- File: `file_path`, `new_text`, `old_text`, `content`
- Prompt: `user_prompt`

### Operadores
- `regex_match`, `contains`, `equals`, `not_contains`, `starts_with`, `ends_with`
