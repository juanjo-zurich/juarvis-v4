---
name: Ralph Wiggum - Bucles Auto-Referenciales
description: Esta skill debe usarse cuando el usuario pide "bucle de desarrollo iterativo", "ralph loop", "bucle auto-referencial", "desarrollo iterativo con IA", "bucle while infinito", o quiere implementar la técnica Ralph Wiggum para desarrollo iterativo continuo donde el agente mejora su propio trabajo en cada iteración.
version: 0.1.0
---

# Ralph Wiggum - Técnica de Bucles Auto-Referenciales

## Visión General

Ralph es una metodología de desarrollo basada en bucles continuos de agente IA. Como Geoffrey Huntley lo describe: **"Ralph es un bucle Bash"** - un simple `while true` que alimenta repetidamente un agente IA con un archivo de prompt, permitiéndole mejorar iterativamente su trabajo hasta completarse.

La técnica lleva el nombre de Ralph Wiggum de Los Simpson, encarnando la filosofía de iteración persistente a pesar de los contratiempos.

### Concepto Central

Este plugin implementa Ralph usando un **hook Stop** que intercepta los intentos de salida del agente:

```
El usuario ejecuta UNA VEZ: /ralph-loop "Descripción de la tarea"

Luego el sistema automáticamente:
1. Trabaja en la tarea
2. Intenta salir
3. El hook Stop bloquea la salida
4. El hook Stop alimenta el MISMO prompt de vuelta
5. Repetir hasta completarse
```

El bucle ocurre **dentro de la sesión actual** - no se necesitan bucles bash externos.

## Conceptos Clave

### Auto-Referencia

El "bucle" no significa que el agente hable consigo mismo. Significa:
- Mismo prompt repetido
- El trabajo del agente persiste en archivos
- Cada iteración ve intentos anteriores
- Construye incrementalmente hacia el objetivo

### Promesas de Completitud

Para señalizar completitud, el agente debe emitir una etiqueta `<promise>`:

```
<promise>TAREA COMPLETA</promise>
```

El hook Stop busca esta etiqueta específica. Sin ella (ni `--max-iterations`), Ralph corre infinitamente.

### Archivo de Estado

El estado del bucle se guarda en `.opencode/ralph-loop.local.md` con frontmatter YAML:

```markdown
---
active: true
iteration: 1
max_iterations: 10
completion_promise: "TAREA COMPLETA"
started_at: "2024-01-15T10:30:00Z"
---

Descripción de la tarea a realizar...
```

## Comandos Disponibles

### /ralph-loop

Iniciar un bucle Ralph en la sesión actual.

**Uso:**
```
/ralph-loop "<prompt>" --max-iterations <n> --completion-promise "<texto>"
```

**Opciones:**
- `--max-iterations <n>` - Parar tras N iteraciones (por defecto: ilimitado)
- `--completion-promise <texto>` - Frase que señaliza completitud

**Ejemplo:**
```
/ralph-loop "Construir una API REST para todos. Requisitos: operaciones CRUD, validación de entrada, tests. Emitir <promise>COMPLETO</promise> cuando terminado." --completion-promise "COMPLETO" --max-iterations 50
```

### /cancel-ralph

Cancelar el bucle Ralph activo.

**Uso:**
```
/cancel-ralph
```

Elimina el archivo de estado `.opencode/ralph-loop.local.md`.

## Mejores Prácticas de Prompts

### 1. Criterios de Completitud Claros

**Mal:** "Construir una API de todos y hacerla buena."

**Bien:**
```
Construir una API REST para todos.

Cuando complete:
- Todos los endpoints CRUD funcionando
- Validación de entrada en su lugar
- Tests pasando (cobertura > 80%)
- README con docs de API
- Emitir: <promise>COMPLETO</promise>
```

### 2. Objetivos Incrementales

**Mal:** "Crear una plataforma de e-commerce completa."

**Bien:**
```
Fase 1: Autenticación de usuario (JWT, tests)
Fase 2: Catálogo de productos (listar/buscar, tests)
Fase 3: Carrito de compra (añadir/eliminar, tests)

Emitir <promise>COMPLETO</promise> cuando todas las fases terminadas.
```

### 3. Auto-Corrección

**Mal:** "Escribir código para la feature X."

**Bien:**
```
Implementar feature X siguiendo TDD:
1. Escribir tests fallando
2. Implementar feature
3. Ejecutar tests
4. Si alguno falla, debuggear y arreglar
5. Refactorizar si necesario
6. Repetir hasta todos en verde
7. Emitir: <promise>COMPLETO</promise>
```

### 4. Escapes de Emergencia

Siempre usar `--max-iterations` como red de seguridad:

```bash
# Recomendado: Siempre establecer un límite razonable
/ralph-loop "Intentar implementar feature X" --max-iterations 20

# En el prompt, incluir qué hacer si se atasca:
# "Tras 15 iteraciones, si no completo:
#  - Documentar qué bloquea el progreso
#  - Listar lo que se intentó
#  - Sugerir enfoques alternativos"
```

## Filosofía

### Iteración > Perfección

No apuntar a perfección en el primer intento. Dejar que el bucle refine el trabajo.

### Los Fallos son Datos

"Mal de forma determinista" significa que los fallos son predecibles e informativos. Usarlos para ajustar prompts.

### La Habilidad del Operador Importa

El éxito depende de escribir buenos prompts, no solo de tener un buen modelo.

### La Persistencia Gana

Seguir intentando hasta el éxito. El bucle maneja la lógica de reintentos automáticamente.

## Cuándo Usar Ralph

**Bueno para:**
- Tareas bien definidas con criterios de éxito claros
- Tareas que requieren iteración y refinamiento
- Proyectos desde cero donde puedes dejar funcionando
- Tareas con verificación automática (tests, linters)

**No bueno para:**
- Tareas que requieren juicio humano o decisiones de diseño
- Operaciones de un solo disparo
- Tareas con criterios de éxito poco claros
- Depuración en producción (usar depuración dirigida)

## Resultados del Mundo Real

- Generación exitosa de 6 repositorios durante la noche en pruebas de hackathon
- Un contrato de $50k completado por $297 en costes de API
- Lenguaje de programación completo ("cursed") creado en 3 meses con este enfoque

## Hook Stop

El hook stop (`hooks/stop-hook.sh`) es el corazón de Ralph:

1. Comprueba si existe `.opencode/ralph-loop.local.md`
2. Si no existe, permitir salida normalmente
3. Si existe, parsear frontmatter para iteración, máximo y promesa
4. Comprobar si máximo de iteraciones alcanzado
5. Comprobar si promesa de completitud detectada en último mensaje
6. Si no completo, incrementar iteración, alimentar prompt de vuelta, bloquear salida

## Seguridad del Bucle

- Validación de campos numéricos antes de operaciones aritméticas
- Comprobación de existencia de archivo de transcripción
- Detección de archivos de estado corruptos con mensajes de error claros
- Limpieza automática de estado corrupto (eliminación del archivo)
- Comparación literal de strings (no pattern matching) para promesas
