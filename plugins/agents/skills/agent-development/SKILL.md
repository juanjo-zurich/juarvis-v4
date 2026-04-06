---
name: Desarrollo de Agentes
description: Esta skill debe usarse cuando el usuario pide "crear un agente", "añadir un agente", "escribir un subagente", "frontmatter de agente", "descripción de uso", "ejemplos de agente", "agente autónomo", o necesita orientación sobre estructura de agentes, prompts de sistema, condiciones de activación, o mejores prácticas de desarrollo de agentes.
version: 0.1.0
---

# Desarrollo de Agentes

## Visión General

Los agentes son subprocesos autónomos que manejan tareas complejas multi-paso de forma independiente.

**Conceptos clave:**
- Los agentes son PARA trabajo autónomo, los comandos son PARA acciones iniciadas por el usuario
- Formato de archivo Markdown con frontmatter YAML
- Activación mediante campo description con ejemplos
- El prompt de sistema define el comportamiento del agente
- Personalización de modelo y color

## Estructura del Archivo de Agente

### Formato Completo

```markdown
---
name: identificador-agente
description: Usar este agente cuando [condiciones de activación]. Ejemplos:

<example>
Contexto: [Descripción de situación]
usuario: "[Petición del usuario]"
asistente: "[Cómo debería responder y usar este agente]"
<commentary>
[Razón por la que se activa este agente]
</commentary>
</example>

model: inherit
color: blue
tools: ["Read", "Write", "Grep"]
---

Tú eres [descripción del rol del agente]...

**Responsabilidades Principales:**
1. [Responsabilidad 1]
2. [Responsabilidad 2]

**Proceso de Análisis:**
[Flujo de trabajo paso a paso]

**Formato de Salida:**
[Qué devolver]
```

## Campos de Frontmatter

### name (requerido)

Identificador del agente para namespacing e invocación.

**Formato:** minúsculas, números, solo guiones
**Longitud:** 3-50 caracteres
**Patrón:** Debe empezar y terminar con alfanumérico

**Ejemplos válidos:**
- `code-reviewer`
- `test-generator`
- `api-docs-writer`

### description (requerido)

Define cuándo el agente debería activarse. **Este es el campo más crítico.**

**Debe incluir:**
1. Condiciones de activación ("Usar este agente cuando...")
2. Múltiples bloques `<example>` mostrando uso
3. Contexto, petición del usuario y respuesta del asistente en cada ejemplo
4. `<commentary>` explicando por qué se activa el agente

### model (requerido)

Qué modelo debe usar el agente.

**Opciones:**
- `inherit` - Usar mismo modelo que padre (recomendado)
- `sonnet` - Equilibrado
- `opus` - Más capaz, costoso
- `haiku` - Rápido, económico

### color (requerido)

Identificador visual del agente en UI.

**Opciones:** `blue`, `cyan`, `green`, `yellow`, `magenta`, `red`

**Guías:**
- Blue/cyan: Análisis, revisión
- Green: Tareas orientadas a éxito
- Yellow: Precaución, validación
- Red: Crítico, seguridad
- Magenta: Creativo, generación

### tools (opcional)

Restringir agente a herramientas específicas.

```yaml
tools: ["Read", "Write", "Grep"]
```

**Por defecto:** Si se omite, agente tiene acceso a todas las herramientas.

**Mejor práctica:** Limitar herramientas al mínimo necesario (principio de mínimo privilegio).

## Diseño del Prompt de Sistema

El cuerpo markdown se convierte en el prompt de sistema del agente.

### Estructura Estándar

```markdown
Tú eres [rol] especializado en [dominio].

**Responsabilidades Principales:**
1. [Responsabilidad primaria]
2. [Responsabilidad secundaria]

**Proceso de Análisis:**
1. [Paso uno]
2. [Paso dos]

**Estándares de Calidad:**
- [Estándar 1]
- [Estándar 2]

**Formato de Salida:**
- [Qué incluir]
- [Cómo estructurar]

**Casos Límite:**
- [Caso límite 1]: [Cómo manejar]
- [Caso límite 2]: [Cómo manejar]
```

### Mejores Prácticas

**HACER:**
- Escribir en segunda persona ("Tú eres...", "Tú...")
- Ser específico sobre responsabilidades
- Proporcionar proceso paso a paso
- Definir formato de salida
- Incluir estándares de calidad
- Abordar casos límite
- Mantener bajo 10.000 caracteres

**NO HACER:**
- Escribir en primera persona
- Ser vago o genérico
- Omitir pasos del proceso
- Dejar formato de salida indefinido
- Saltar guía de calidad
- Ignorar casos de error

## Creación de Agentes

### Método 1: Generación Asistida

Usar el patrón de prompt para crear configuración de agente basada en descripción del usuario. Extraer intención central, diseñar persona experta, crear prompt de sistema comprensivo, generar identificador y escribir descripción con condiciones de activación.

### Método 2: Creación Manual

1. Elegir identificador (3-50 chars, minúsculas, guiones)
2. Escribir descripción con ejemplos
3. Seleccionar modelo (usualmente `inherit`)
4. Elegir color para identificación visual
5. Definir herramientas (si restringir acceso)
6. Escribir prompt de sistema con estructura anterior
7. Guardar como `agents/nombre-agente.md`

## Reglas de Validación

### Validación de Identificador

```
✅ Válido: code-reviewer, test-gen, api-analyzer-v2
❌ Inválido: ag (muy corto), -start (empieza con guión), my_agent (guión bajo)
```

### Validación de Descripción

**Longitud:** 10-5.000 caracteres
**Debe incluir:** Condiciones de activación y ejemplos
**Óptimo:** 200-1.000 caracteres con 2-4 ejemplos

### Validación del Prompt de Sistema

**Longitud:** 20-10.000 caracteres
**Óptimo:** 500-3.000 caracteres
**Estructura:** Responsabilidades claras, proceso, formato de salida

## Pruebas de Agentes

### Probar Activación

1. Escribir agente con ejemplos de activación específicos
2. Usar redacción similar a ejemplos en test
3. Verificar que el agente carga correctamente
4. Confirmar que proporciona funcionalidad esperada

### Probar Prompt de Sistema

1. Dar al agente tarea típica
2. Verificar que sigue pasos del proceso
3. Confirmar formato de salida correcto
4. Probar casos límite mencionados en prompt
5. Confirmar estándares de calidad

## Referencia Rápida

### Agente Mínimo

```markdown
---
name: simple-agent
description: Usar este agente cuando... Ejemplos: <example>...</example>
model: inherit
color: blue
---

Tú eres un agente que [hace X].

Proceso:
1. [Paso 1]
2. [Paso 2]

Salida: [Qué proporcionar]
```

### Resumen de Campos

| Campo | Requerido | Formato | Ejemplo |
|-------|-----------|---------|---------|
| name | Sí | minúsculas-guiones | code-reviewer |
| description | Sí | Texto + ejemplos | Usar cuando... <example>... |
| model | Sí | inherit/sonnet/opus/haiku | inherit |
| color | Sí | Nombre de color | blue |
| tools | No | Array de nombres | ["Read", "Grep"] |

## Flujo de Implementación

1. Definir propósito del agente y condiciones de activación
2. Elegir método de creación (asistido o manual)
3. Crear archivo `agents/nombre-agente.md`
4. Escribir frontmatter con todos los campos requeridos
5. Escribir prompt de sistema siguiendo mejores prácticas
6. Incluir 2-4 ejemplos de activación en descripción
7. Validar estructura del agente
8. Probar activación con escenarios reales
9. Documentar agente en README del plugin
