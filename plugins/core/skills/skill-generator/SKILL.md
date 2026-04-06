---
name: skill-generator
description: >
  Detecta la falta de habilidades técnicas o arquitectónicas en el proyecto y propone su creación automática.
  Trigger: always
license: MIT
metadata:
  author: Antigravity-Auto
  version: "1.0.0"
---

## Propósito

Tu misión es actuar como un detector de brechas de conocimiento. Si el usuario pide algo que requiere convenciones específicas (ej. un framework, un patrón de diseño, un flujo de testing) y no existe un skill en el `skill-registry` que lo cubra, debes proponer al usuario crear ese skill usando las herramientas internas.

## Cuándo Ejecutar

Este skill tiene un trigger `always`. Debes evaluar la `user_request` al inicio de cada interacción.

## Qué Hacer

### Paso 1: Mapeo de Intención
Analiza el pedido del usuario buscando conceptos "skillables":
- **Tecnologías**: React, Angular, Golang, Rust, Python, Node, Terraform, etc.
- **Arquitectura**: Hexagonal, Clean, DDD, Microservicios, MVC, etc.
- **Dominios**: Auth, API, Testing, Database, CI/CD, CSS-Modules, etc.

### Paso 2: Verificación de Registro
Busca en el archivo `.juar/skill-registry.md` (o en la memoria de engram para "skill-registry"):
- ¿Existe algún skill que se active con estos conceptos?
- Si NO existe, has identificado un **Skill Gap**.

### Paso 3: Propuesta al Usuario
Si detectas un Gap, debes interrumpir el flujo normal y proponer:
"He detectado que no tengo un skill para **{concept}**. Para asegurar que mi trabajo siga tus estándares, ¿quieres que cree un nuevo skill usando `scripts/skill-create.sh`?"

Explica qué vas a incluir basándote en lo que el usuario pidió (ej: "Voy a incluir reglas para la estructura de carpetas de {concept}").

### Paso 4: Creación (con aprobación)
Si el usuario dice que sí:
1. Identifica un nombre corto y en minúsculas (ej: `auth-jwt`).
2. Ejecuta: `./scripts/skill-create/skill-create.sh <skill-name> --type "custom"`
3. Lee el archivo generado (normalmente en `skills/custom/<skill-name>.md`).
4. **Bootstrap**: Si el usuario dio detalles en su prompt original, incorpóralos como reglas en la sección `## Rules`.

### Paso 5: Validación y Re-indexación
1. Valida el skill: `./scripts/skill-create/templates/validate-skill.sh skills/custom/<skill-name>.md`
2. Dispara el `skill-registry` para que el nuevo skill sea usable de inmediato.

## Reglas

- **NUNCA** crees un skill sin preguntar primero.
- **SIEMPRE** usá el script `skill-create.sh` para mantener consistencia con los templates.
- Sé específico: es mejor tener `react-testing` que un genérico `testing` si el proyecto es solo React.
- Si el usuario rechaza la propuesta, procedé con el trabajo usando tus conocimientos generales, pero advirtiéndole que sería mejor tener un skill.
