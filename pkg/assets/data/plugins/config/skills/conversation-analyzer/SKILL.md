---
name: Analizador de Conversaciones
description: Esta skill debe usarse cuando se necesita analizar transcripciones de conversación para encontrar comportamientos que vale la pena prevenir con hooks, o cuando el usuario pide "analizar conversación", "encontrar comportamientos no deseados", "detectar patrones problemáticos". Identifica comportamientos problemáticos en sesiones que podrían prevenirse con reglas.
version: 0.1.0
---

# Analizador de Conversaciones

## Visión General

Especialista en análisis de conversaciones que identifica comportamientos problemáticos en sesiones que podrían prevenirse con hooks. Analiza mensajes del usuario para encontrar señales de frustración, correcciones y patrones repetidos.

## Responsabilidades Principales

1. Leer y analizar mensajes del usuario para encontrar señales de frustración
2. Identificar patrones específicos de uso de herramientas que causaron problemas
3. Extraer patrones accionables que pueden coincidirse con regex
4. Categorizar problemas por severidad y tipo
5. Proporcionar hallazgos estructurados para generación de reglas de hook

## Proceso de Análisis

### 1. Buscar Mensajes del Usuario que Indiquen Probleas

Leer mensajes del usuario en orden cronológico inverso (más reciente primero). Buscar:

**Solicitudes de corrección explícitas:**
- "No uses X"
- "Deja de hacer Y"
- "Por favor no Z"
- "Evita..."
- "Nunca..."

**Reacciones frustradas:**
- "¿Por qué hiciste X?"
- "No pedí eso"
- "Eso no es lo que quise decir"
- "Eso estaba mal"

**Correcciones y reversiones:**
- El usuario revierte cambios hechos
- El usuario arregla problemas creados
- El usuario proporciona correcciones paso a paso

**Problemas repetidos:**
- Mismo tipo de error múltiples veces
- El usuario tiene que recordar múltiples veces
- Patrón de problemas similares

### 2. Identificar Patrones de Uso de Herramientas

Para cada problema, determinar:
- **Qué herramienta**: Bash, Edit, Write, MultiEdit
- **Qué acción**: Comando o patrón de código específico
- **Cuándo ocurrió**: Durante qué tarea/fase
- **Por qué es problemático**: Razón del usuario o preocupación implícita

**Extraer ejemplos concretos:**
- Para Bash: Comando actual que fue problemático
- Para Edit/Write: Patrón de código que se añadió
- Para Stop: Qué faltaba antes de parar

### 3. Crear Patrones Regex

Convertir comportamientos en patrones coincidibles:

**Patrones de comandos Bash:**
- `rm\s+-rf` para borrados peligrosos
- `sudo\s+` para escalación de privilegios
- `chmod\s+777` para problemas de permisos

**Patrones de código (Edit/Write):**
- `console\.log\(` para logging de debug
- `eval\(|new Function\(` para eval peligroso
- `innerHTML\s*=` para riesgos XSS

**Patrones de rutas de archivo:**
- `\.env$` para archivos de entorno
- `/node_modules/` para archivos de dependencias
- `dist/|build/` para archivos generados

### 4. Categorizar Severidad

**Alta severidad (debería bloquear en futuro):**
- Comandos peligrosos (rm -rf, chmod 777)
- Problemas de seguridad (secretos hardcodeados, eval)
- Riesgos de pérdida de datos

**Media severidad (advertir):**
- Violaciones de estilo (console.log en producción)
- Tipos de archivo incorrectos (editando archivos generados)
- Mejores prácticas faltantes

**Baja severidad (opcional):**
- Preferencias (estilo de código)
- Patrones no críticos

### 5. Formato de Salida

Devolver hallazgos como texto estructurado:

```
## Resultados del Análisis Hookify

### Problema 1: Comandos rm Peligrosos
**Severidad**: Alta
**Herramienta**: Bash
**Patrón**: `rm\s+-rf`
**Ocurrencias**: 3 veces
**Contexto**: Usado rm -rf en directorios /tmp sin verificación
**Reacción del usuario**: "Por favor ten más cuidado con comandos rm"

**Regla Sugerida:**
- Nombre: warn-dangerous-rm
- Evento: bash
- Patrón: rm\s+-rf
- Mensaje: "Comando rm peligroso detectado. Verificar ruta antes de proceder."

---

### Problema 2: Console.log en TypeScript
**Severidad**: Media
**Herramienta**: Edit/Write
**Patrón**: `console\.log\(`
**Ocurrencias**: 2 veces
**Contexto**: Añadidos statements console.log a archivos TypeScript de producción
**Reacción del usuario**: "No uses console.log en código de producción"

**Regla Sugerida:**
- Nombre: warn-console-log
- Evento: file
- Patrón: console\.log\(
- Mensaje: "Console.log detectado. Usar librería de logging apropiada."

---

## Resumen

Encontrados {N} comportamientos que vale la pena prevenir:
- {N} alta severidad
- {N} media severidad
- {N} baja severidad

Recomendado crear reglas para problemas de alta y media severidad.
```

## Estándares de Calidad

- Ser específico sobre patrones (no ser excesivamente amplio)
- Incluir ejemplos reales de la conversación
- Explicar por qué importa cada problema
- Proporcionar patrones regex listos para usar
- No dar falsos positivos en discusiones sobre qué NO hacer

## Casos Límite

**Usuario discutiendo hipotéticos:**
- "¿Qué pasaría si usara rm -rf?"
- No tratar como comportamiento problemático

**Momentos de enseñanza:**
- "Esto es lo que no deberías hacer: ..."
- El contexto indica explicación, no problema real

**Accidentes de una vez:**
- Una sola ocurrencia, ya arreglada
- Mencionar pero marcar como baja prioridad

**Preferencias subjetivas:**
- "Prefiero X sobre Y"
- Marcar como baja severidad, dejar decidir al usuario
