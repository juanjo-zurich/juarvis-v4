---
name: Cargador de Configuración Hookify
description: Esta skill debe usarse cuando el usuario quiere "entender cómo se cargan las reglas", "parsear frontmatter YAML", "estructura de archivos de regla", o necesita referencia técnica sobre el cargador de configuración que lee y parsea archivos .local.md.
version: 0.1.0
---

# Cargador de Configuración - Referencia Técnica

## Visión General

El cargador de configuración gestiona la lectura y parsing de archivos `.juarvis/hookify.*.local.md`. Extrae frontmatter YAML y cuerpo de mensaje para crear objetos `Rule` evaluables.

## Estructura de Datos

### Condition (Condición)

```python
@dataclass
class Condition:
    field: str       # "command", "new_text", "old_text", "file_path"
    operator: str    # "regex_match", "contains", "equals", etc.
    pattern: str     # Patrón a coincidir
```

### Rule (Regla)

```python
@dataclass
class Rule:
    name: str
    enabled: bool
    event: str                    # "bash", "file", "stop", "all"
    pattern: Optional[str]        # Patrón simple (legacy)
    conditions: List[Condition]   # Lista de condiciones
    action: str = "warn"          # "warn" o "block"
    tool_matcher: Optional[str]   # Override de coincidencia de herramientas
    message: str = ""             # Cuerpo del mensaje markdown
```

## Parsing de Frontmatter

### Extracción

La función `extract_frontmatter()` divide el contenido en marcadores `---`:

1. Verificar que el contenido empieza con `---`
2. Dividir en 3 partes por `---`
3. Parte 1: frontmatter YAML (parsear línea a línea)
4. Parte 2: cuerpo del mensaje (después del segundo `---`)

### Parser YAML Simple

El parser maneja:
- Pares clave-valor simples: `key: value`
- Booleanos: `true`/`false` (case-insensitive)
- Listas simples: `- item`
- Listas con diccionarios inline: `- field: x, operator: y`
- Diccionarios multi-línea en listas:
  ```yaml
  conditions:
    - field: command
      operator: regex_match
      pattern: rm\s+-rf
  ```

### Conversión Simple Pattern a Conditions

Si solo hay un campo `pattern` (sin `conditions`), se convierte automáticamente:

- `event: bash` → `field: command`
- `event: file` → `field: new_text`
- Otros → `field: content`

Se crea una sola Condition con `operator: regex_match`.

## Carga de Reglas

### load_rules(event=None)

1. Buscar todos los archivos `.juarvis/hookify.*.local.md` con `glob`
2. Para cada archivo, llamar `load_rule_file()`
3. Filtrar por evento si se especifica
4. Solo incluir reglas habilitadas
5. Manejar errores de I/O, parsing y codificación

### load_rule_file(file_path)

1. Leer contenido del archivo
2. Extraer frontmatter y mensaje
3. Validar que frontmatter existe
4. Crear objeto `Rule` con `Rule.from_dict()`
5. Manejar errores de I/O, parsing y codificación de forma robusta

## Manejo de Errores

El cargador es resiliente a errores:
- **Errores de I/O**: Registrar advertencia, continuar con siguiente archivo
- **Errores de parsing**: Registrar advertencia, continuar
- **Errores de codificación**: Registrar advertencia, continuar
- **Archivo sin frontmatter**: Registrar advertencia, devolver None
- **Archivo vacío o corrupto**: Registrar advertencia, devolver None

## Ejemplo de Archivo de Regla

```markdown
---
name: warn-dangerous-rm
enabled: true
event: bash
pattern: rm\s+-rf
action: block
---

⚠️ **¡Comando rm peligroso detectado!**

Este comando podría borrar archivos importantes.
```

### Parsing Resultante

```
Rule(
    name="warn-dangerous-rm",
    enabled=True,
    event="bash",
    pattern="rm\\s+-rf",
    conditions=[
        Condition(field="command", operator="regex_match", pattern="rm\\s+-rf")
    ],
    action="block",
    message="⚠️ **¡Comando rm peligroso detectado!**..."
)
```

## Referencia Rápida

### Extraer Frontmatter
```python
frontmatter, message = extract_frontmatter(content)
```

### Cargar Todas las Reglas
```python
rules = load_rules()            # Todas las reglas habilitadas
rules = load_rules(event='bash')  # Solo reglas de bash
```

### Crear Regla desde Frontmatter
```python
rule = Rule.from_dict(frontmatter, message)
```
