---
name: pr-review-toolkit
description: Revisión integral de Pull Requests con agentes especializados. Analiza comentarios, tests, manejo de errores, diseño de tipos, calidad de código y simplificación. Trigger: Cuando el usuario solicita revisar un PR o quiere análisis de calidad antes de hacer commit/merge.
type: orchestrator
version: "1.0"
---

# PR Review Toolkit

Colección completa de agentes especializados para revisión exhaustiva de Pull Requests, cubriendo comentarios, cobertura de tests, manejo de errores, diseño de tipos, calidad de código y simplificación.

## Visión general

Este skill orquesta 6 agentes expertos, cada uno centrado en un aspecto específico de la calidad del código. Se pueden usar individualmente para revisiones dirigidas o en conjunto para un análisis completo del PR.

## Agentes disponibles

### 1. pr-comment-analyzer
**Enfoque**: Precisión y mantenibilidad de los comentarios del código

**Analiza:**
- Precisión de los comentarios frente al código real
- Completitud de la documentación
- Putrefacción de comentarios y deuda técnica
- Comentarios engaños u obsoletos

**Cuándo usarlo:**
- Después de añadir documentación
- Antes de finalizar PRs con cambios en comentarios
- Al revisar comentarios existentes

**Activadores:**
```
"Comprueba si los comentarios son precisos"
"Revisa la documentación que he añadido"
"Analiza los comentarios en busca de deuda técnica"
```

### 2. pr-test-analyzer
**Enfoque**: Calidad y completitud de la cobertura de tests

**Analiza:**
- Cobertura comportamental frente a cobertura de líneas
- Huecos críticos en la cobertura de tests
- Calidad y resiliencia de los tests
- Casos límite y condiciones de error

**Cuándo usarlo:**
- Después de crear un PR
- Al añadir nueva funcionalidad
- Para verificar la exhaustividad de los tests

**Activadores:**
```
"Comprueba si los tests son exhaustivos"
"Revisa la cobertura de tests de este PR"
"¿Hay huecos críticos en los tests?"
```

### 3. pr-silent-failure-hunter
**Enfoque**: Manejo de errores y fallos silenciosos

**Analiza:**
- Fallos silenciosos en bloques catch
- Manejo de errores inadecuado
- Comportamiento de respaldo inapropiado
- Falta de registro de errores (logging)

**Cuándo usarlo:**
- Después de implementar manejo de errores
- Al revisar bloques try/catch
- Antes de finalizar PRs con manejo de errores

**Activadores:**
```
"Revisa el manejo de errores"
"Comprueba si hay fallos silenciosos"
"Analiza los bloques catch de este PR"
```

### 4. pr-type-design-analyzer
**Enfoque**: Calidad del diseño de tipos e invariantes

**Analiza:**
- Encapsulación de tipos (valoración 1-10)
- Expresión de invariantes (valoración 1-10)
- Utilidad del tipo (valoración 1-10)
- Cumplimiento de invariantes (valoración 1-10)

**Cuándo usarlo:**
- Al introducir nuevos tipos
- Durante la creación de PRs con modelos de datos
- Al refactorizar diseños de tipos

**Activadores:**
```
"Revisa el diseño del tipo UserAccount"
"Analiza el diseño de tipos de este PR"
"Comprueba si este tipo tiene invariantes fuertes"
```

### 5. pr-code-reviewer
**Enfoque**: Revisión general de código según las directrices del proyecto

**Analiza:**
- Cumplimiento de las normas del proyecto
- Violaciones de estilo
- Detección de errores
- Problemas de calidad de código

**Cuándo usarlo:**
- Después de escribir o modificar código
- Antes de hacer commit
- Antes de crear Pull Requests

**Activadores:**
```
"Revisa mis cambios recientes"
"Comprueba si todo está correcto"
"Revisa este código antes de hacer commit"
```

### 6. pr-code-simplifier
**Enfoque**: Simplificación y refactorización de código

**Analiza:**
- Claridad y legibilidad del código
- Complejidad innecesaria y anidamiento
- Código y abstracciones redundantes
- Coherencia con los estándares del proyecto
- Código excesivamente compacto o ingenioso

**Cuándo usarlo:**
- Después de escribir o modificar código
- Después de pasar la revisión de código
- Cuando el código funciona pero parece complejo

**Activadores:**
```
"Simplifica este código"
"Haz esto más claro"
"Refina esta implementación"
```

## Patrones de uso

### Uso individual de agentes

Simplemente solicita una revisión que coincida con el área de enfoque de un agente:

```
"¿Puedes comprobar si los tests cubren todos los casos límite?"
→ Activa pr-test-analyzer

"Revisa el manejo de errores del cliente API"
→ Activa pr-silent-failure-hunter

"He añadido documentación, ¿es precisa?"
→ Activa pr-comment-analyzer
```

### Revisión completa de PR

Para una revisión exhaustiva, solicita múltiples aspectos:

```
"Estoy listo para crear este PR. Por favor:
1. Revisa la cobertura de tests
2. Comprueba si hay fallos silenciosos
3. Verifica que los comentarios del código son precisos
4. Revisa los nuevos tipos
5. Revisión general de código"
```

Esto activará todos los agentes relevantes para analizar diferentes aspectos del PR.

### Revisión proactiva

El orquestador puede activar proactivamente estos agentes según el contexto:

- **Después de escribir código** → pr-code-reviewer
- **Después de añadir documentación** → pr-comment-analyzer
- **Antes de crear un PR** → Múltiples agentes según corresponda
- **Después de añadir tipos** → pr-type-design-analyzer

## Puntuación de confianza

Los agentes proporcionan puntuaciones de confianza para sus hallazgos:

- **pr-comment-analyzer**: Identifica problemas con alta confianza en verificaciones de precisión
- **pr-test-analyzer**: Valora los huecos de tests del 1 al 10 (10 = crítico, imprescindible añadir)
- **pr-silent-failure-hunter**: Marca la gravedad de los problemas de manejo de errores
- **pr-type-design-analyzer**: Valora 4 dimensiones en escala 1-10
- **pr-code-reviewer**: Puntuación de 0 a 100 (91-100 = crítico)
- **pr-code-simplifier**: Identifica complejidad y sugiere simplificaciones

## Formato de salida

Todos los agentes proporcionan salida estructurada y accionable:
- Identificación clara del problema
- Referencias específicas a archivo y línea
- Explicación de por qué es un problema
- Sugerencias de mejora
- Priorizadas por gravedad

## Mejores prácticas

### Cuándo usar cada agente

**Antes de hacer commit:**
- pr-code-reviewer (calidad general)
- pr-silent-failure-hunter (si se ha cambiado el manejo de errores)

**Antes de crear PR:**
- pr-test-analyzer (comprobación de cobertura de tests)
- pr-comment-analyzer (si se han añadido/modificado comentarios)
- pr-type-design-analyzer (si se han añadido/modificado tipos)
- pr-code-reviewer (revisión final)

**Después de pasar la revisión:**
- pr-code-simplifier (mejorar claridad y mantenibilidad)

**Durante la revisión del PR:**
- Cualquier agente para preocupaciones específicas planteadas
- Re-revisión dirigida después de correcciones

### Ejecución de múltiples agentes

Se pueden solicitar múltiples agentes en paralelo o secuencialmente:

**En paralelo** (más rápido):
```
"Ejecuta pr-test-analyzer y pr-comment-analyzer en paralelo"
```

**Secuencial** (cuando uno informa al otro):
```
"Primero revisa la cobertura de tests, después comprueba la calidad del código"
```

## Flujo de trabajo recomendado

1. Escribir código → **pr-code-reviewer**
2. Corregir problemas → **pr-silent-failure-hunter** (si hay manejo de errores)
3. Añadir tests → **pr-test-analyzer**
4. Documentar → **pr-comment-analyzer**
5. La revisión pasa → **pr-code-simplifier** (pulir)
6. Crear PR
