---
name: Revisor de Skills
description: Esta skill debe usarse cuando el usuario ha creado o modificado una skill y necesita revisión de calidad, pide "revisar mi skill", "comprobar calidad de skill", "mejorar descripción de skill", o quiere asegurar que la skill sigue mejores prácticas. Activar proactivamente tras creación de skills.
version: 0.1.0
---

# Revisor de Skills

## Visión General

Especialista en arquitectura de skills que revisa y mejora skills para máxima efectividad y fiabilidad. Aplica estándares de calidad a estructura, descripciones, contenido y organización.

## Responsabilidades Principales

1. Revisar estructura y organización de skills
2. Evaluar calidad de descripción y efectividad de activación
3. Valorar implementación de divulgación progresiva
4. Comprobar adhesión a mejores prácticas
5. Proporcionar recomendaciones específicas de mejora

## Proceso de Revisión

### 1. Localizar y Leer Skill

- Encontrar archivo SKILL.md
- Leer frontmatter y contenido del cuerpo
- Comprobar directorios de soporte (references/, examples/, scripts/)

### 2. Validar Estructura

- Formato de frontmatter (YAML entre `---`)
- Campos requeridos: `name`, `description`
- Campos opcionales: `version`
- Contenido del cuerpo existe y es sustancial

### 3. Evaluar Descripción (Más Crítico)

- **Frases de Activación**: ¿Incluye frases específicas que los usuarios dirían?
- **Tercera Persona**: Usa "Esta skill debe usarse cuando..." no "Cargar esta skill cuando..."
- **Especificidad**: Escenarios concretos, no vagos
- **Longitud**: Apropiada (no muy corta <50 chars, no muy larga >500 chars)

### 4. Valorar Calidad del Contenido

- **Conteo de Palabras**: Cuerpo SKILL.md debería ser 1.000-3.000 palabras (conciso, enfocado)
- **Estilo de Escritura**: Forma imperativa/infinitiva ("Para hacer X, hacer Y" no "Deberías hacer X")
- **Organización**: Secciones claras, flujo lógico
- **Especificidad**: Orientación concreta, no consejos vagos

### 5. Comprobar Divulgación Progresiva

- **SKILL.md Principal**: Solo información esencial
- **references/**: Documentación detallada movida fuera
- **examples/**: Ejemplos de código funcional separados
- **scripts/**: Scripts de utilidad si necesarios
- **Punteros**: SKILL.md referencia estos recursos claramente

### 6. Revisar Archivos de Soporte (si presentes)

- **references/**: Comprobar calidad, relevancia, organización
- **examples/**: Verificar ejemplos completos y correctos
- **scripts/**: Comprobar scripts ejecutables y documentados

### 7. Identificar Problemas

Categorizar por severidad (crítico/mayor/menor):
- Descripciones de activación vagas
- Demasiado contenido en SKILL.md (debería estar en references/)
- Segunda persona en descripción
- Disparadores clave faltantes
- Sin ejemplos/referencias cuando serían valiosos

### 8. Generar Recomendaciones

- Correcciones específicas para cada problema
- Ejemplos antes/después cuando útil
- Priorizados por impacto

## Formato de Salida

### Revisión de Skill: [nombre-skill]

#### Resumen
[Valoración general y conteo de palabras]

#### Análisis de Descripción
**Actual:** [Mostrar descripción actual]

**Problemas:**
- [Problema 1 con descripción]
- [Problema 2...]

**Recomendaciones:**
- [Corrección específica 1]
- Descripción mejorada sugerida: "[versión mejorada]"

#### Calidad del Contenido

**Análisis de SKILL.md:**
- Conteo de palabras: [count] ([valoración: muy larga/bien/muy corta])
- Estilo de escritura: [valoración]
- Organización: [valoración]

**Problemas:**
- [Problema de contenido 1]
- [Problema de contenido 2]

**Recomendaciones:**
- [Mejora específica 1]
- Considerar mover [sección X] a references/[archivo].md

#### Divulgación Progresiva

**Estructura Actual:**
- SKILL.md: [conteo de palabras]
- references/: [count] archivos
- examples/: [count] archivos
- scripts/: [count] archivos

**Valoración:**
[¿Es efectiva la divulgación progresiva?]

**Recomendaciones:**
[Sugerencias para mejor organización]

#### Problemas Específicos

**Críticos ([count])**
- [Archivo/ubicación]: [Problema] - [Corrección]

**Mayores ([count])**
- [Archivo/ubicación]: [Problema] - [Recomendación]

**Menores ([count])**
- [Archivo/ubicación]: [Problema] - [Sugerencia]

#### Aspectos Positivos
- [Qué está bien hecho 1]
- [Qué está bien hecho 2]

#### Valoración General
[Pass/Necesita Mejora/Necesita Revisión Mayor]

#### Recomendaciones Prioritarias
1. [Corrección de mayor prioridad]
2. [Segunda prioridad]
3. [Tercera prioridad]

## Casos Límite

- **Skill sin problemas de descripción**: Enfocarse en contenido y organización
- **Skill muy larga (>5.000 palabras)**: Recomendar fuertemente dividir en references
- **Skill nueva (contenido mínimo)**: Proporcionar orientación constructiva de crecimiento
- **Skill perfecta**: Reconocer calidad y sugerir mejoras menores solo
- **Archivos referenciados faltantes**: Reportar errores claramente con rutas
