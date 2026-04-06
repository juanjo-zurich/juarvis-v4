---
name: feature-dev
description: Desarrollo guiado de funcionalidades con comprensión profunda de la base de código y enfoque arquitectónico. Incluye 7 fases: descubrimiento, exploración, preguntas aclaratorias, diseño de arquitectura, implementación, revisión de calidad y resumen.
trigger: Cuando el usuario quiere implementar una nueva funcionalidad o cambio significativo siguiendo un flujo estructurado. Activador típico: crear funcionalidad, implementar feature, desarrollo guiado, construir nueva feature.
---

# Desarrollo de Funcionalidades

Estás ayudando a un desarrollador a implementar una nueva funcionalidad. Sigue un enfoque sistemático: comprende la base de código profundamente, identifica y pregunta por todos los detalles sin especificar, diseña arquitecturas elegantes, luego implementa.

## Principios Fundamentales

- **Pregunta para aclarar**: Identifica todas las ambigüedades, casos límite y comportamientos sin especificar. Haz preguntas específicas y concretas en lugar de asumir. Espera las respuestas del usuario antes de proceder con la implementación. Pregunta pronto (después de comprender la base de código, antes de diseñar la arquitectura).
- **Comprende antes de actuar**: Lee y comprende los patrones de código existentes primero
- **Lee los archivos identificados por sub-agentes**: Al lanzar sub-agentes, pídeles que devuelvan listas de los archivos más importantes para leer. Una vez que los sub-agentes terminen, lee esos archivos para construir un contexto detallado antes de proceder.
- **Simple y elegante**: Prioriza código legible, mantenible y arquitectónicamente sólido
- **Usa seguimiento de progreso**: Registra todo el progreso durante la ejecución

---

## Fase 1: Descubrimiento

**Objetivo**: Comprender qué se debe construir

Solicitud inicial: proporcionada por el usuario al iniciar el flujo.

**Acciones**:
1. Crear lista de seguimiento con todas las fases
2. Si la funcionalidad no está clara, preguntar al usuario:
   - ¿Qué problema están resolviendo?
   - ¿Qué debe hacer la funcionalidad?
   - ¿Alguna restricción o requisito?
3. Resumir la comprensión y confirmar con el usuario

---

## Fase 2: Exploración de la Base de Código

**Objetivo**: Comprender el código y patrones existentes relevantes a alto y bajo nivel

**Acciones**:
1. Lanzar 2-3 sub-agentes de exploración en paralelo (skill `code-explorer`). Cada sub-agente debe:
   - Rastrear el código de forma exhaustiva y centrarse en obtener una comprensión completa de abstracciones, arquitectura y flujo de control
   - Objetivo: un aspecto diferente de la base de código (ej: funcionalidades similares, comprensión a alto nivel, comprensión arquitectónica, experiencia de usuario, etc.)
   - Incluir una lista de 5-10 archivos clave para leer

   **Ejemplos de prompts para sub-agentes**:
   - «Encuentra funcionalidades similares a [funcionalidad] y rastrea su implementación de forma exhaustiva»
   - «Mapea la arquitectura y abstracciones para [área de funcionalidad], rastreando el código de forma exhaustiva»
   - «Analiza la implementación actual de [funcionalidad/área existente], rastreando el código de forma exhaustiva»
   - «Identifica patrones de UI, enfoques de testing o puntos de extensión relevantes para [funcionalidad]»

2. Una vez que los sub-agentes devuelvan resultados, leer todos los archivos identificados por ellos para construir una comprensión profunda
3. Presentar un resumen exhaustivo de los hallazgos y patrones descubiertos

---

## Fase 3: Preguntas Aclaratorias

**Objetivo**: Rellenar huecos y resolver todas las ambigüedades antes de diseñar

**CRÍTICO**: Esta es una de las fases más importantes. NO SALTAR.

**Acciones**:
1. Revisar los hallazgos de la base de código y la solicitud original de funcionalidad
2. Identificar aspectos sin especificar: casos límite, gestión de errores, puntos de integración, límites de alcance, preferencias de diseño, compatibilidad hacia atrás, necesidades de rendimiento
3. **Presentar todas las preguntas al usuario en una lista clara y organizada**
4. **Esperar las respuestas antes de proceder al diseño de arquitectura**

Si el usuario dice «lo que tú creas mejor», proporciona tu recomendación y obtén confirmación explícita.

---

## Fase 4: Diseño de Arquitectura

**Objetivo**: Diseñar múltiples enfoques de implementación con diferentes compromisos

**Acciones**:
1. Lanzar 2-3 sub-agentes de arquitectura en paralelo (skill `code-architect`) con diferentes enfoques: cambios mínimos (menor cambio, máxima reutilización), arquitectura limpia (mantenibilidad, abstracciones elegantes), o equilibrio pragmático (velocidad + calidad)
2. Revisar todos los enfoques y formar tu opinión sobre cuál encaja mejor para esta tarea específica (considerar: corrección pequeña vs funcionalidad grande, urgencia, complejidad, contexto del equipo)
3. Presentar al usuario: resumen breve de cada enfoque, comparación de compromisos, **tu recomendación con razonamiento**, diferencias concretas de implementación
4. **Preguntar al usuario qué enfoque prefiere**

---

## Fase 5: Implementación

**Objetivo**: Construir la funcionalidad

**NO EMPEZAR SIN APROBACIÓN DEL USUARIO**

**Acciones**:
1. Esperar aprobación explícita del usuario
2. Leer todos los archivos relevantes identificados en fases anteriores
3. Implementar siguiendo la arquitectura elegida
4. Seguir estrictamente las convenciones de la base de código
5. Escribir código limpio y bien documentado
6. Actualizar el seguimiento de progreso conforme avanzas

---

## Fase 6: Revisión de Calidad

**Objetivo**: Asegurar que el código es simple, DRY, elegante, fácil de leer y funcionalmente correcto

**Acciones**:
1. Lanzar 3 sub-agentes de revisión en paralelo (skill `code-reviewer`) con diferentes enfoques: simplicidad/DRY/elegancia, errores/corrección funcional, convenciones del proyecto/abstracciones
2. Consolidar hallazgos e identificar los problemas de mayor severidad que recomiendas corregir
3. **Presentar hallazgos al usuario y preguntar qué quiere hacer** (corregir ahora, corregir después, o proceder tal cual)
4. Abordar problemas según la decisión del usuario

---

## Fase 7: Resumen

**Objetivo**: Documentar lo que se ha logrado

**Acciones**:
1. Marcar todas las tareas como completadas
2. Resumir:
   - Qué se construyó
   - Decisiones clave tomadas
   - Archivos modificados
   - Siguientes pasos sugeridos

---
