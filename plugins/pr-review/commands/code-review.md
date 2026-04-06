---
description: Revisión de código de un pull request
argument-hint: --comment
---

Proporciona una revisión de código para el pull request indicado.

**Suposiciones del agente (aplica a todos los agentes y sub-agentes):**
- Todas las herramientas funcionan y funcionarán sin errores. No pruebes herramientas ni hagas llamadas exploratorias. Asegúrate de que esto quede claro para cada sub-agente que se lance.
- Solo llama a una herramienta si es necesaria para completar la tarea. Cada llamada de herramienta debe tener un propósito claro.

Para hacer esto, sigue estos pasos con precisión:

1. Lanzar un sub-agente rápido para comprobar si alguna de las siguientes condiciones es verdadera:
   - El pull request está cerrado
   - El pull request es un draft
   - El pull request no necesita revisión de código (ej. PR automatizado, cambio trivial que claramente es correcto)
   - Ya has comentado antes en este PR (comprueba `gh pr view <PR> --comments` para comentarios tuyos)

   Si cualquier condición es verdadera, detente y no continues.

Nota: Revisa igual los PRs generados automáticamente.

2. Lanzar un sub-agente rápido para devolver una lista de rutas de archivos (no su contenido) para todos los archivos de AGENTS.md o archivo de contexto del proyecto relevantes, incluyendo:
   - El archivo AGENTS.md del root, si existe
   - Cualquier archivo AGENTS.md en directorios que contengan archivos modificados por el pull request

3. Lanzar un sub-agente para ver el pull request y devolver un resumen de los cambios

4. Lanzar 4 agentes en paralelo para revisar los cambios independiente. Cada agente debe devolver la lista de problemas, donde cada problema incluye una descripción y la razón por la que se marcó (ej. "cumplimiento de AGENTS.md", "bug"). Los agentes deben hacer lo siguiente:

   Agentes 1 + 2: Sub-agentes de cumplimiento de AGENTS.md
   Audita los cambios para cumplimiento de AGENTS.md en paralelo. Nota: Al evaluar el cumplimiento de AGENTS.md para un archivo, solo debes considerar archivos AGENTS.md que compartan una ruta de archivo con el archivo o sus directorios padre.

   Agente 3: Sub-agente especializado de bugs (subagente paralelo con agente 4)
   Escanea bugs obvios. Enfócate solo en el diff mismo sin leer contexto extra. Marca solo bugs significativos; ignora detalles menores y falsos positivos probables. No marques problemas que no puedas validar sin mirar contexto fuera del git diff.

   Agente 4: Sub-agente especializado de bugs (subagente paralelo con agente 3)
   Busca problemas que existen en el código introducido. Esto podría ser problemas de seguridad, lógica incorrecta, etc. Solo busca problemas que estén dentro del código cambiado.

   **CRÍTICO: Solo queremos problemas de ALTA SEÑAL.** Marca problemas donde:
   - El código fallará al compilar o parsear (errores de sintaxis, errores de tipo, imports faltantes, referencias no resueltas)
   - El código producirá definitely wrong results independientemente de las entradas (errores de lógica claros)
   - Cumplimiento claro e inequívoco de AGENTS.md donde puedas citar la regla exacta que se está violando

   NO marques:
   - Preocupaciones de estilo o calidad de código
   - Problemas potenciales que dependen de entradas o estado específicos
   - Sugerencias o mejoras subjetivas

   Si no estás seguro de que un problema sea real, no lo marques. Los falsos positivos erosionan la confianza y desperdician tiempo del revisor.

   Además de lo anterior, cada sub-agente debe recibir el título y descripción del PR. Esto ayudará a proporcionar contexto sobre la intención del autor.

5. Para cada problema encontrado en el paso anterior por los agentes 3 y 4, lanzar sub-agentes paralelos para validar el problema. Estos sub-agentes deben recibir el título y descripción del PR junto con una descripción del problema. El trabajo del agente es revisar el problema para validar que el problema declarado es realmente un problema con alta confianza. Por ejemplo, si se marcó un problema como "variable no definida", el trabajo del sub-agente sería validar que eso es realmente cierto en el código. Otro ejemplo sería problemas de AGENTS.md. El agente debe validar que la regla de AGENTS.md que se violó aplica para este archivo y realmente está violada. Usa sub-agentes especializados para bugs y problemas de lógica, y sub-agentes para violaciones de AGENTS.md.

6. Filtrar cualquier problema que no fue validado en el paso 5. Este paso nos dará nuestra lista de problemas de alta señal para nuestra revisión.

7. Imprimir un resumen de los hallazgos de la revisión en la terminal:
   - Si se encontraron problemas, listar cada problema con una breve descripción.
   - Si no se encontraron problemas, stating: "No se encontraron problemas. Verificado cumplimiento de AGENTS.md y bugs."

   Si el argumento `--comment` NO fue proporcionado, detente aquí. No postear ningún comentario en GitHub.

   Si el argumento `--comment` ES proporcionado y NO se encontraron problemas, postear un comentario resumen usando `gh pr comment` y detenerse.

   Si el argumento `--comment` ES PROVIDENCIADO y SE encontraron problemas, continuar al paso 8.

8. Crear una lista de todos los comentarios que planeas dejar. Esto es solo para que estés seguro de que te sientes cómodo con los comentarios. No postear esta lista en ningún lugar.

9. Postear comentarios en línea para cada problema usando `gh pr comment` con `confirmed: true`. Para cada comentario:
   - Proporcionar una breve descripción del problema
   - Para fixes pequeños y autocontenidos, incluir un bloque de sugerencia committeable
   - Para fixes más grandes (6+ líneas, cambios estructurales, o cambios que abarcan múltiples ubicaciones), describir el problema y fix sugerido sin un bloque de sugerencia
   - Nunca postear una sugerencia committeable A MENOS QUE comprometer la sugerencia resuelva el problema enteramente. Si se requieren follow up steps, no dejar una sugerencia committeable.

   **IMPORTANTE: Solo postear UN comentario por problema único. No postear comentarios duplicados.**

Usa esta lista al evaluar problemas en los Pasos 4 y 5 (estos son falsos positivos, NO marcar):

- Problemas preexistentes
- Algo que parece un bug pero en realidad es correcto
- Detalles menores que un ingeniero senior no marcaría
- Problemas que un linter detectará (no ejecutar el linter para verificar)
- Preocupaciones generales de calidad de código (ej. falta de cobertura de tests, problemas de seguridad generales) a menos que explícitamente se requiera en AGENTS.md
- Problemas mencionados en AGENTS.md pero explícitamente silenciados en el código (ej. via un comentario lint ignore)

Notas:

- Usar gh CLI para interactuar con GitHub (ej. fetch pull requests, crear comentarios). No usar web fetch.
- Crear una lista de tareas antes de empezar.
- Debes citar y enlazar cada problema en comentarios en línea (ej. si te refieres a un AGENTS.md, incluir un enlace a este).
- Si no se encuentran problemas y se proporciona el argumento `--comment`, postear un comentario con el siguiente formato:

---

## Revisión de código

No se encontraron problemas. Verificado cumplimiento de AGENTS.md y bugs.

---

- Cuando enlaces a código en comentarios en línea, seguir el siguiente formato con precisión, de lo contrario el preview de Markdown no renderizará correctamente: https://github.com/owner/repo/blob/COMMIT_SHA/package.json#L10-L15
  - Requiere el sha completo de git
  - Debes proporcionar el sha completo. Comandos como `https://github.com/owner/repo/blob/$(git rev-parse HEAD)/foo/bar` no funcionarán, ya que tu comentario se renderizará directamente en Markdown.
  - El nombre del repo debe coincidir con el repo que estás revisando
  - Signo # después del nombre del archivo
  - El formato del rango de líneas es L[inicio]-L[fin]
  - Proporcionar al menos 1 línea de contexto antes y después, centrado en la línea sobre la que estás comentando (ej. si estás comentando sobre líneas 5-6, debes enlazar a `L4-7`)