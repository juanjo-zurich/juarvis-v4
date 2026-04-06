---
name: Motor de Reglas Hookify
description: Esta skill debe usarse cuando el usuario quiere "entender el motor de reglas", "evaluar reglas de hook", "lógica de coincidencia de patrones", "evaluación de condiciones de hook", o necesita referencia técnica sobre cómo el motor de reglas evalúa reglas contra datos de entrada.
version: 0.1.0
---

# Motor de Reglas - Referencia Técnica

## Visión General

El motor de reglas es el componente que evalúa reglas configuradas contra datos de entrada de hooks. Gestiona la coincidencia de patrones, evaluación de condiciones y generación de respuestas de hook.

## Arquitectura

### Clase RuleEngine

Responsable de evaluar reglas y devolver resultados combinados.

**Método principal:** `evaluate_rules(rules, input_data)`

Comprueba todas las reglas y acumula coincidencias. Las reglas de bloqueo tienen prioridad sobre las de advertencia. Todas las reglas coincidentes se combinan.

### Proceso de Evaluación

1. Obtener `hook_event_name` de los datos de entrada
2. Para cada regla, comprobar si coincide con `_rule_matches()`
3. Separar reglas coincidentes en blocking y warning
4. Si hay reglas de bloqueo, devolver respuesta de bloqueo con mensajes combinados
5. Si solo hay advertencias, devolver mensajes como `systemMessage`
6. Si no hay coincidencias, devolver `{}` (permitir operación)

### Coincidencia de Herramientas

El método `_matches_tool()` comprueba si el nombre de la herramienta coincide con el patrón matcher:
- `*` coincide con todas las herramientas
- Patrones separados por `|` se comprueba como lista (OR)

### Evaluación de Condiciones

El método `_check_condition()` evalúa una condición individual:

1. Extraer valor del campo con `_extract_field()`
2. Aplicar operador:
   - `regex_match`: Coincidencia regex (case-insensitive)
   - `contains`: Contiene substring
   - `equals`: Coincidencia exacta
   - `not_contains`: No contiene substring
   - `starts_with`: Empieza con prefijo
   - `ends_with`: Termina con sufijo

### Extracción de Campos

El método `_extract_field()` obtiene valores de diferentes fuentes según el tipo de herramienta:

**Bash:** `command` de `tool_input`
**Edit/Write:** `file_path`, `new_string`, `old_string`, `content`
**MultiEdit:** `file_path`, concatenación de `new_string` de todas las ediciones
**Stop events:** `reason`, `transcript` (lee archivo de transcripción)
**UserPromptSubmit:** `user_prompt`

## Formato de Respuesta

### Respuesta de Bloqueo (Stop)

```json
{
  "decision": "block",
  "reason": "Mensaje combinado",
  "systemMessage": "Mensaje combinado"
}
```

### Respuesta de Bloqueo (PreToolUse/PostToolUse)

```json
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "deny"
  },
  "systemMessage": "Mensaje combinado"
}
```

### Respuesta de Advertencia

```json
{
  "systemMessage": "Mensaje combinado de reglas coincidentes"
}
```

### Sin Coincidencia

```json
{}
```

## Optimización

### Cache de Regex

Usa `@lru_cache(maxsize=128)` para cachear regex compilados, mejorando rendimiento en evaluaciones repetidas.

### Carga de Reglas

El `config_loader.py` gestiona la carga de reglas desde archivos `.opencode/hookify.*.local.md`:

1. Busca archivos que coincidan con el patrón
2. Parsea frontmatter YAML y cuerpo del mensaje
3. Crea objetos `Rule` con condiciones
4. Filtra por evento si se especifica
5. Solo incluye reglas habilitadas

## Formato de Datos de Entrada

Los hooks reciben JSON por stdin con campos comunes:

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.txt",
  "cwd": "/current/working/dir",
  "hook_event_name": "PreToolUse",
  "tool_name": "Write",
  "tool_input": {
    "file_path": "/test/file.txt",
    "new_string": "contenido"
  }
}
```

## Manejo de Errores

- **Errores de importación**: Permitir operación y registrar error
- **Regex inválido**: Registrar advertencia, devolver False para esa condición
- **Archivo de transcripción no encontrado**: Registrar advertencia, devolver string vacío
- **Errores de parsing JSON**: Capturar y devolver mensaje de error como systemMessage
- **Errores inesperados**: Capturar excepción general, permitir operación, registrar error

## Referencia Rápida de Clases

### Rule
- `name`: Identificador único
- `enabled`: Booleano de activación
- `event`: Tipo de evento (bash, file, stop, prompt, all)
- `pattern`: Patrón simple (legacy)
- `conditions`: Lista de objetos Condition
- `action`: "warn" o "block"
- `tool_matcher`: Override de coincidencia de herramientas
- `message`: Cuerpo del mensaje markdown

### Condition
- `field`: Campo a comprobar (command, new_text, file_path, etc.)
- `operator`: Cómo coincidir (regex_match, contains, equals, etc.)
- `pattern`: Patrón o string a coincidir
