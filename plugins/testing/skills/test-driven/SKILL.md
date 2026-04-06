---
name: test-driven
description: Guía de mejores prácticas para el Desarrollo Guiado por Tests (Test-Driven Development), el diagnóstico sistemático y las pruebas automáticas eficientes.
triggers:
  - "tdd"
  - "test-driven"
  - "red-green-refactor"
  - "escribe tests"
  - "quiero más tests"
version: "1.1.0"
---

# Test-Driven Development (TDD) y Estrategias Guiadas por Tests

El enfoque de Juarvis_V3 hacia las pruebas es riguroso. Los tests no son una ocurrencia tardía, sino la especificación ejecutable que dirige el desarrollo. 

## El Ciclo Central: RED → GREEN → REFACTOR

El núcleo del TDD es un ciclo rápido y estricto:

### 1. RED — Escribir un test que falla (La Especificación)
- Escribe *solo un* test (o un paquete pequeño) para la próxima funcionalidad deseada.
- El test DEBE fallar ("Rojo"). Si pasa sin código nuevo, el test no está comprobando lo que crees que está comprobando o la funcionalidad ya existe.
- Enfócate en el comportamiento, no en la implementación. Define claramente cuáles son las entradas y la salida esperada.

### 2. GREEN — Hacerlo pasar (La Implementación Mínima)
- Escribe *solo* el código de producción suficiente para hacer pasar el test.
- ¡No sobre-diseñes! Si el código "feísimo" o un valor hardcodeado hace pasar el test temporalmente, úsalo si es todo lo que requiere el test. La elegancia vendrá después.
- Asegúrate de que tanto el nuevo test como todos los tests anteriores pasen (Verde absoluto).

### 3. REFACTOR — Mejorar el diseño (La Limpieza)
- Ya con el test en verde, revisa tu código.
- Elimina duplicidad.
- Mejora nombres de variables (intención reveladora).
- Extrae métodos, aplica patrones de diseño.
- Modifica la estructura sin miedo: como tienes cobertura total, si rompes algo, sabrás exactamente qué fue (Volverá a Rojo).

## Prácticas Recomendadas

### Mantén la Frecuencia Alta
El ciclo Red-Green-Refactor debería tomar entre un par de minutos y máximo 15-20 minutos por iteración. Si el ciclo se alarga, el paso RED fue demasiado grande. Desglosa ese test en pasos más pequeños.

### "Given - When - Then" (Arrange - Act - Assert)
Estructura tus tests visualmente:
1. **Given (Arrange)**: Configura el estado inicial (mocks, datos, objetos).
2. **When (Act)**: Llama a la función o comportamiento que estás testeando.
3. **Then (Assert)**: Comprueba el resultado post-condición.
- Mantén solo 1-2 Asserts por test, enfocados en un comportamiento específico.

### FIRST: Propiedades de un Buen Test
- **F**ast: Rápidos. Si tardan minutos nadie los correrá. Unitaria != Lenta.
- **I**ndependent: No depender de otro test ni del estado global. Orden aleatorio no debe afectar.
- **R**epeatable: Debe dar igual resultado tu PC local, la de tu amigo, o CI.
- **S**elf-Validating: Automáticos, no requieren inspección manual (un pass/fail claro).
- **T**imely: Escritos justo a tiempo (antes del código).

### El Peligro de los Mocks
- Cuidado con "sobre-mockear". Los mocks excesivos atan el test a implementaciones específicas (cómo hace las cosas) en vez de comportamiento (qué debe devolver de resultado).
- Mockea estrictamente los límites (I/O, APIs, BD), usa stubs/fakes, y prueba la lógica de negocio real.

## Workflow en Juarvis

Cuando el usuario pida aplicar TDD:
1. Crea/edita el archivo de tests PRIMERO.
2. Muestra que falla o explica el diseño del test.
3. Propón la implementación para el archivo de producción.
