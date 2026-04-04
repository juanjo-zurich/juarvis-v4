# Revisión completa de PR

Ejecuta una revisión completa de Pull Request usando múltiples agentes especializados, cada uno centrado en un aspecto diferente de la calidad del código.

**Aspectos de revisión (opcional):** "$ARGUMENTS"

## Flujo de revisión:

1. **Determinar el alcance de la revisión**
   - Comprobar el estado de git para identificar archivos modificados
   - Analizar los argumentos para ver si el usuario ha solicitado aspectos específicos
   - Por defecto: ejecutar todas las revisiones aplicables

2. **Aspectos de revisión disponibles:**

   - **comments** — Analizar la precisión y mantenibilidad de los comentarios del código
   - **tests** — Revisar la calidad y completitud de la cobertura de tests
   - **errors** — Comprobar el manejo de errores en busca de fallos silenciosos
   - **types** — Analizar el diseño de tipos e invariantes (si se han añadido nuevos tipos)
   - **code** — Revisión general de código según las directrices del proyecto
   - **simplify** — Simplificar código para claridad y mantenibilidad
   - **all** — Ejecutar todas las revisiones aplicables (por defecto)

3. **Identificar archivos modificados**
   - Ejecutar `git diff --name-only` para ver archivos modificados
   - Comprobar si ya existe un PR: `gh pr view`
   - Identificar tipos de archivo y qué revisiones aplican

4. **Determinar revisiones aplicables**

   Según los cambios:
   - **Siempre aplicable**: pr-code-reviewer (calidad general)
   - **Si se modifican archivos de tests**: pr-test-analyzer
   - **Si se añaden comentarios/documentación**: pr-comment-analyzer
   - **Si se cambia el manejo de errores**: pr-silent-failure-hunter
   - **Si se añaden/modifican tipos**: pr-type-design-analyzer
   - **Después de pasar la revisión**: pr-code-simplifier (pulir y refinar)

5. **Lanzar agentes de revisión**

   **Enfoque secuencial** (uno a uno):
   - Más fácil de entender y abordar
   - Cada informe está completo antes del siguiente
   - Adecuado para revisión interactiva

   **Enfoque en paralelo** (el usuario puede solicitarlo):
   - Lanzar todos los agentes simultáneamente
   - Más rápido para revisión completa
   - Los resultados llegan juntos

6. **Agregar resultados**

   Después de que los agentes completen, resumir:
   - **Problemas críticos** (debe corregir antes de mergear)
   - **Problemas importantes** (debería corregir)
   - **Sugerencias** (sería bueno tener)
   - **Observaciones positivas** (lo que está bien hecho)

7. **Proporcionar plan de acción**

   Organizar los hallazgos:
   ```markdown
   # Resumen de revisión de PR

   ## Problemas críticos (X encontrados)
   - [agente]: Descripción del problema [archivo:línea]

   ## Problemas importantes (X encontrados)
   - [agente]: Descripción del problema [archivo:línea]

   ## Sugerencias (X encontradas)
   - [agente]: Sugerencia [archivo:línea]

   ## Fortalezas
   - Lo que está bien hecho en este PR

   ## Acción recomendada
   1. Corregir problemas críticos primero
   2. Abordar problemas importantes
   3. Considerar sugerencias
   4. Re-ejecutar la revisión después de las correcciones
   ```

## Ejemplos de uso:

**Revisión completa (por defecto):**
```
/review-pr
```

**Aspectos específicos:**
```
/review-pr tests errors
# Revisa solo cobertura de tests y manejo de errores

/review-pr comments
# Revisa solo comentarios del código

/review-pr simplify
# Simplifica el código después de pasar la revisión
```

**Revisión en paralelo:**
```
/review-pr all parallel
# Lanza todos los agentes en paralelo
```

## Descripción de agentes:

**pr-comment-analyzer:**
- Verifica la precisión de los comentarios frente al código
- Identifica putrefacción de comentarios
- Comprueba la completitud de la documentación

**pr-test-analyzer:**
- Revisa la cobertura comportamental de tests
- Identifica huecos críticos
- Evalúa la calidad de los tests

**pr-silent-failure-hunter:**
- Encuentra fallos silenciosos
- Revisa bloques catch
- Comprueba el registro de errores

**pr-type-design-analyzer:**
- Analiza la encapsulación de tipos
- Revisa la expresión de invariantes
- Valora la calidad del diseño de tipos

**pr-code-reviewer:**
- Comprueba el cumplimiento de las normas del proyecto
- Detecta errores y problemas
- Revisa la calidad general del código

**pr-code-simplifier:**
- Simplifica código complejo
- Mejora la claridad y legibilidad
- Aplica estándares del proyecto
- Preserva la funcionalidad

## Consejos:

- **Ejecutar pronto**: Antes de crear el PR, no después
- **Centrarse en cambios**: Los agentes analizan git diff por defecto
- **Abordar lo crítico primero**: Corregir problemas de alta prioridad antes que los de baja
- **Re-ejecutar después de correcciones**: Verificar que los problemas están resueltos
- **Usar revisiones específicas**: Dirigir aspectos concretos cuando se conoce la preocupación

## Integración con el flujo de trabajo:

**Antes de hacer commit:**
```
1. Escribir código
2. Ejecutar: /review-pr code errors
3. Corregir problemas críticos
4. Hacer commit
```

**Antes de crear PR:**
```
1. Preparar todos los cambios
2. Ejecutar: /review-pr all
3. Abordar todos los problemas críticos e importantes
4. Ejecutar revisiones específicas de nuevo para verificar
5. Crear PR
```

**Después del feedback del PR:**
```
1. Realizar los cambios solicitados
2. Ejecutar revisiones dirigidas según el feedback
3. Verificar que los problemas están resueltos
4. Subir actualizaciones
```
