---
name: Desarrollo de Skills
description: Esta skill debe usarse cuando el usuario quiere "crear una skill", "añadir una skill a un plugin", "escribir una nueva skill", "mejorar descripción de skill", "organizar contenido de skill", o necesita orientación sobre estructura de skills, divulgación progresiva, o mejores prácticas de desarrollo de skills.
version: 0.1.0
---

# Desarrollo de Skills

## Visión General

Las skills son paquetes modulares y autocontenidos que extienden las capacidades del agente proporcionando conocimiento especializado, flujos de trabajo y herramientas.

### Qué Proporcionan las Skills

1. Flujos de trabajo especializados - Procedimientos multi-paso para dominios específicos
2. Integraciones de herramientas - Instrucciones para trabajar con formatos o APIs específicas
3. Experiencia de dominio - Conocimiento específico de la empresa, esquemas, lógica de negocio
4. Recursos empaquetados - Scripts, referencias y activos para tareas complejas y repetitivas

### Anatomía de una Skill

Cada skill consiste en un archivo SKILL.md requerido y recursos empaquetados opcionales:

```
nombre-skill/
├── SKILL.md (requerido)
│   ├── Metadatos YAML frontmatter (requerido)
│   │   ├── name: (requerido)
│   │   └── description: (requerido)
│   └── Instrucciones Markdown (requerido)
└── Recursos Empaquetados (opcional)
    ├── scripts/          - Código ejecutable (Python/Bash/etc.)
    ├── references/       - Documentación para cargar en contexto según necesidad
    └── assets/           - Archivos usados en salida (plantillas, iconos, fuentes, etc.)
```

## Principio de Divulgación Progresiva

Las skills usan un sistema de carga de tres niveles para gestionar contexto eficientemente:

1. **Metadatos (name + description)** - Siempre en contexto (~100 palabras)
2. **Cuerpo SKILL.md** - Cuando skill se activa (<5k palabras)
3. **Recursos empaquetados** - Según necesidad del agente (Ilimitado*)

## Proceso de Creación de Skills

### Paso 1: Entender la Skill con Ejemplos Concretos

Clarificar ejemplos concretos de cómo se usará la skill. Preguntar al usuario sobre:
- Qué funcionalidad debe soportar
- Ejemplos de uso
- Qué diría el usuario para activar la skill

### Paso 2: Planificar Contenido Reutilizable

Analizar cada ejemplo considerando:
1. Cómo ejecutar el ejemplo desde cero
2. Qué scripts, referencias y activos serían útiles

**Ejemplo:** Para una skill de edición de imágenes, identificar que rotar un PDF requiere reescribir el mismo código cada vez → guardar `scripts/rotate_pdf.py`.

### Paso 3: Crear Estructura de la Skill

```bash
mkdir -p plugin-name/skills/nombre-skill/{references,examples,scripts}
touch plugin-name/skills/nombre-skill/SKILL.md
```

### Paso 4: Editar la Skill

#### Empezar con Contenido Reutilizable

Crear los archivos `scripts/`, `references/` y `assets/` identificados. Eliminar archivos de ejemplo no necesarios. Crear solo los directorios realmente necesarios.

#### Actualizar SKILL.md

**Estilo de Escritura:** Usar **forma imperativa/infinitiva** (instrucciones verbo-primero), no segunda persona. Ejemplo: "Para hacer X, hacer Y" en lugar de "Deberías hacer X".

**Descripción (Frontmatter):** Usar formato tercera persona con frases de activación específicas:

```yaml
---
name: Nombre de Skill
description: Esta skill debe usarse cuando el usuario pide "frase específica 1", "frase específica 2". Incluir frases exactas que los usuarios dirían para activar esta skill. Ser concreto y específico.
version: 0.1.0
---
```

**Ejemplo bueno de descripción:**
```yaml
description: Esta skill debe usarse cuando el usuario pide "crear un hook", "añadir un hook PreToolUse", "validar uso de herramientas".
```

**Ejemplo malo de descripción:**
```yaml
description: Guía para trabajar con hooks.  # Vago, sin frases de activación
```

**Mantener SKILL.md conciso:** Objetivo de 1.500-2.000 palabras para el cuerpo. Mover contenido detallado a `references/`.

**Referenciar recursos en SKILL.md:**
```markdown
## Recursos Adicionales

### Archivos de Referencia
- **`references/patterns.md`** - Patrones comunes
- **`references/advanced.md`** - Casos avanzados

### Ejemplos
- **`examples/script.sh`** - Ejemplo funcional
```

### Paso 5: Validar y Probar

1. **Comprobar estructura**: Archivo SKILL.md existe con frontmatter válido
2. **Validar SKILL.md**: Tiene frontmatter con name y description
3. **Comprobar frases de activación**: Descripción incluye consultas específicas de usuario
4. **Verificar estilo de escritura**: Cuerpo usa forma imperativa/infinitiva
5. **Probar divulgación progresiva**: SKILL.md conciso, contenido detallado en references/
6. **Comprobar referencias**: Todos los archivos referenciados existen
7. **Validar ejemplos**: Ejemplos completos y correctos
8. **Probar scripts**: Scripts ejecutables y funcionan correctamente

### Paso 6: Iterar

Tras probar la skill, mejorar según uso:
- Fortalecer frases de activación en descripción
- Mover secciones largas de SKILL.md a references/
- Añadir ejemplos o scripts faltantes
- Aclarar instrucciones ambiguas
- Añadir manejo de casos límite

## Consideraciones Específicas

### Estilo de Escritura

**Correcto (imperativo):**
```
Para crear un hook, definir el tipo de evento.
Configurar el servidor MCP con autenticación.
Validar configuración antes de usar.
```

**Incorrecto (segunda persona):**
```
Deberías crear un hook definiendo el tipo de evento.
Necesitas configurar el servidor MCP.
Debes validar la configuración.
```

### Tercera Persona en Descripción

**Correcto:**
```yaml
description: Esta skill debe usarse cuando el usuario pide "crear X", "configurar Y"...
```

**Incorrecto:**
```yaml
description: Usar esta skill cuando quieras crear X...
```

## Errores Comunes a Evitar

### Error 1: Descripción Débil

❌ `description: Proporciona orientación para hooks.`
✅ `description: Esta skill debe usarse cuando el usuario pide "crear un hook", "añadir un hook PreToolUse", "validar uso de herramientas".`

### Error 2: Demasiado Contenido en SKILL.md

❌ SKILL.md de 8.000 palabras
✅ SKILL.md de 1.800 palabras + references/ con 2.500 palabras

### Error 3: Escritura en Segunda Persona

❌ `Deberías empezar leyendo el archivo de configuración.`
✅ `Empezar leyendo el archivo de configuración.`

### Error 4: Referencias a Recursos Faltantes

❌ SKILL.md sin mención de references/
✅ SKILL.md que referencia claramente resources adicionales

## Referencia Rápida

### Skill Mínima
```
nombre-skill/
└── SKILL.md
```

### Skill Estándar (Recomendado)
```
nombre-skill/
├── SKILL.md
├── references/
│   └── detailed-guide.md
└── examples/
    └── working-example.sh
```

### Skill Completa
```
nombre-skill/
├── SKILL.md
├── references/
│   ├── patterns.md
│   └── advanced.md
├── examples/
│   ├── example1.sh
│   └── example2.json
└── scripts/
    └── validate.sh
```

## Flujo de Implementación

1. **Entender casos de uso**: Identificar ejemplos concretos
2. **Planificar recursos**: Determinar scripts/referencias/ejemplos necesarios
3. **Crear estructura**: `mkdir -p skills/nombre-skill/{references,examples,scripts}`
4. **Escribir SKILL.md**: Frontmatter con descripción tercera persona, cuerpo conciso en forma imperativa
5. **Añadir recursos**: Crear references/, examples/, scripts/ según necesidad
6. **Validar**: Comprobar descripción, estilo de escritura, organización
7. **Probar**: Verificar que skill carga en activadores esperados
8. **Iterar**: Mejorar según uso
