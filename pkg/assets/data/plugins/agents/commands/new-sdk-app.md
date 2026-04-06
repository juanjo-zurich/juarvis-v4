---
description: Crear y configurar una nueva aplicación con Agent SDK
argument-hint: [nombre-del-proyecto]
---

Tu tarea es ayudar al usuario a crear una nueva aplicación con Agent SDK. Sigue estos pasos cuidadosamente:

## Documentación de Referencia

Antes de empezar, revisa la documentación oficial para asegurar que proporcionas orientación precisa y actualizada. Usa WebFetch para leer estas páginas:

1. **Empieza con el resumen**: https://docs.claude.com/en/api/agent-sdk/overview
2. **Según la elección de lenguaje del usuario, lee la referencia del SDK apropiada**:
   - TypeScript: https://docs.claude.com/en/api/agent-sdk/typescript
   - Python: https://docs.claude.com/en/api/agent-sdk/python
3. **Lee las guías relevantes mencionadas en el resumen** como:
   - Streaming vs Single Mode
   - Permisos
   - Herramientas personalizadas
   - Integración MCP
   - Subagentes
   - Sesiones
   - Cualquier otra guía relevante según las necesidades del usuario

**IMPORTANTE**: Comprueba y usa siempre las últimas versiones de los paquetes. Usa WebSearch o WebFetch para verificar las versiones actuales antes de la instalación.

## Recopilar Requisitos

IMPORTANTE: Haz estas preguntas una a la vez. Espera la respuesta del usuario antes de hacer la siguiente pregunta. Esto facilita que el usuario responda.

Haz las preguntas en este orden (salta cualquiera que el usuario ya haya proporcionado mediante argumentos):

1. **Lenguaje** (preguntar primero): «¿Te gustaría usar TypeScript o Python?»

   - Espera la respuesta antes de continuar

2. **Nombre del proyecto** (preguntar segundo): «¿Qué nombre quieres darle a tu proyecto?»

   - Si se proporciona $ARGUMENTS, úsalo como nombre del proyecto y salta esta pregunta
   - Espera la respuesta antes de continuar

3. **Tipo de agente** (preguntar tercero, pero salta si #2 fue suficientemente detallado): «¿Qué tipo de agente estás construyendo? Algunos ejemplos:

   - Agente de código (SRE, revisión de seguridad, revisión de código)
   - Agente de negocio (soporte al cliente, creación de contenido)
   - Agente personalizado (describe tu caso de uso)»
   - Espera la respuesta antes de continuar

4. **Punto de partida** (preguntar cuarto): «¿Te gustaría:

   - Un ejemplo mínimo de «Hola Mundo» para empezar
   - Un agente básico con funcionalidades comunes
   - Un ejemplo específico basado en tu caso de uso»
   - Espera la respuesta antes de continuar

5. **Elección de herramientas** (preguntar quinto): Informa al usuario qué herramientas usarás y confirma con él que son las que quiere usar (por ejemplo, puede preferir pnpm o bun en lugar de npm). Respeta las preferencias del usuario al ejecutar los requisitos.

Después de responder todas las preguntas, procede a crear el plan de configuración.

## Plan de Configuración

Basado en las respuestas del usuario, crea un plan que incluya:

1. **Inicialización del proyecto**:

   - Crear directorio del proyecto (si no existe)
   - Inicializar gestor de paquetes:
     - TypeScript: `npm init -y` y configurar `package.json` con type: "module" y scripts (incluir un script "typecheck")
     - Python: Crear `requirements.txt` o usar `poetry init`
   - Añadir archivos de configuración necesarios:
     - TypeScript: Crear `tsconfig.json` con ajustes apropiados para el SDK
     - Python: Opcionalmente crear archivos de configuración si es necesario

2. **Comprobar Últimas Versiones**:

   - ANTES de instalar, usa WebSearch o comprueba npm/PyPI para encontrar la última versión
   - Para TypeScript: Comprobar https://www.npmjs.com/package/@anthropic-ai/claude-agent-sdk
   - Para Python: Comprobar https://pypi.org/project/claude-agent-sdk/
   - Informa al usuario qué versión estás instalando

3. **Instalación del SDK**:

   - TypeScript: `npm install @anthropic-ai/claude-agent-sdk@latest` (o especificar última versión)
   - Python: `pip install claude-agent-sdk` (pip instala la última por defecto)
   - Tras la instalación, verifica la versión instalada:
     - TypeScript: Comprobar package.json o ejecutar `npm list @anthropic-ai/claude-agent-sdk`
     - Python: Ejecutar `pip show claude-agent-sdk`

4. **Crear archivos iniciales**:

   - TypeScript: Crear un `index.ts` o `src/index.ts` con un ejemplo básico de consulta
   - Python: Crear un `main.py` con un ejemplo básico de consulta
   - Incluir imports correctos y manejo básico de errores
   - Usar sintaxis moderna y actualizada de la última versión del SDK

5. **Configuración del entorno**:

   - Crear un archivo `.env.example` con `ANTHROPIC_API_KEY=your_api_key_here`
   - Añadir `.env` a `.gitignore`
   - Explicar cómo obtener una API key desde https://console.anthropic.com/

6. **Opcional: Crear estructura de directorios**:
   - Ofrecer crear directorio `.opencode/` para agentes, comandos y ajustes
   - Preguntar si quiere algún subagente o comando de ejemplo

## Implementación

Tras recopilar requisitos y obtener confirmación del usuario sobre el plan:

1. Comprueba las últimas versiones de paquetes usando WebSearch o WebFetch
2. Ejecuta los pasos de configuración
3. Crea todos los archivos necesarios
4. Instala las dependencias (usa siempre las últimas versiones estables)
5. Verifica las versiones instaladas e informa al usuario
6. Crea un ejemplo funcional basado en su tipo de agente
7. Añade comentarios útiles en el código explicando qué hace cada parte
8. **VERIFICA QUE EL CÓDIGO FUNCIONA ANTES DE TERMINAR**:
   - Para TypeScript:
     - Ejecuta `npx tsc --noEmit` para comprobar errores de tipos
     - Corrige TODOS los errores de tipos hasta que la comprobación pase completamente
     - Asegúrate de que los imports y tipos son correctos
     - Solo continúa cuando la comprobación de tipos pase sin errores
   - Para Python:
     - Verifica que los imports son correctos
     - Comprueba errores básicos de sintaxis
   - **NO consideres la configuración completa hasta que el código se verifique correctamente**

## Verificación

Tras crear todos los archivos e instalar las dependencias, usa el verificador apropiado para validar que la aplicación con Agent SDK está correctamente configurada y lista para usar:

1. **Para proyectos TypeScript**: Lanza el agente **agent-sdk-verifier-ts** para validar la configuración
2. **Para proyectos Python**: Lanza el agente **agent-sdk-verifier-py** para validar la configuración
3. El agente comprobará el uso del SDK, la configuración, la funcionalidad y la adherencia a la documentación oficial
4. Revisa el informe de verificación y corrige cualquier problema

## Guía de Inicio

Una vez completada la configuración y verificada, proporciona al usuario:

1. **Siguientes pasos**:

   - Cómo configurar su API key
   - Cómo ejecutar su agente:
     - TypeScript: `npm start` o `node --loader ts-node/esm index.ts`
     - Python: `python main.py`

2. **Recursos útiles**:

   - Enlace a la referencia del SDK TypeScript: https://docs.claude.com/en/api/agent-sdk/typescript
   - Enlace a la referencia del SDK Python: https://docs.claude.com/en/api/agent-sdk/python
   - Explica conceptos clave: system prompts, permisos, herramientas, servidores MCP

3. **Siguientes pasos comunes**:
   - Cómo personalizar el system prompt
   - Cómo añadir herramientas personalizadas mediante MCP
   - Cómo configurar permisos
   - Cómo crear subagentes

## Notas Importantes

- **USA SIEMPRE LAS ÚLTIMAS VERSIONES**: Antes de instalar cualquier paquete, comprueba las últimas versiones usando WebSearch o consultando npm/PyPI directamente
- **VERIFICA QUE EL CÓDIGO SE EJECUTA CORRECTAMENTE**:
  - Para TypeScript: Ejecuta `npx tsc --noEmit` y corrige TODOS los errores de tipos antes de terminar
  - Para Python: Verifica la sintaxis y los imports son correctos
  - NO consideres la tarea completa hasta que el código pase la verificación
- Verifica la versión instalada tras la instalación e informa al usuario
- Comprueba la documentación oficial para cualquier requisito específico de versión (versión de Node.js, versión de Python, etc.)
- Comprueba siempre si los directorios/archivos ya existen antes de crearlos
- Usa el gestor de paquetes preferido del usuario (npm, yarn, pnpm para TypeScript; pip, poetry para Python)
- Asegúrate de que todos los ejemplos de código son funcionales e incluyen un manejo adecuado de errores
- Usa sintaxis y patrones modernos compatibles con la última versión del SDK
- Haz la experiencia interactiva y educativa
- **HAZ LAS PREGUNTAS UNA A LA VEZ** — No hagas múltiples preguntas en una única respuesta

Empieza haciendo SOLAMENTE la primera pregunta de requisitos. Espera la respuesta del usuario antes de continuar con la siguiente pregunta.
