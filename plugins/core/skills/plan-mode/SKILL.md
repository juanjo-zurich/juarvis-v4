---
name: plan-mode
description: Planificación estructurada antes de ejecutar cambios complejos
trigger: Cuando la tarea implica múltiples archivos, dependencias entre pasos, o es un refactor/cambio arquitectónico
license: MIT
metadata:
  version: 1.0.0
  author: Juarvis
  language: es-ES
---

# Plan Mode — Planificación antes de Ejecutar

## Objetivo

Generar un plan estructurado y presentarlo al usuario antes de ejecutar cualquier cambio que implique riesgo o complejidad.

## Cuándo activar

- Múltiples archivos necesitan modificación
- Los comandos tienen dependencias entre sí
- Es un refactor o cambio arquitectónico
- La tarea tiene requisitos secuenciales de múltiples pasos
- El usuario solicita explícitamente `/plan`

## Formato del plan

```markdown
## Plan: [título descriptivo]

### Resumen
[1-2 frases describiendo el objetivo global]

### Pasos
**Paso 1: [acción]**
- Archivos: `path/to/file.ext`
- Comando: `{comando}`
- Verificación: {criterio de verificación}

**Paso 2: [acción]**
- Archivos: `path/to/file.ext`
- Comando: `{comando}`
- Verificación: {criterio de verificación}

### Criterios de Verificación
- [ ] {criterio 1}
- [ ] {criterio 2}
- [ ] {criterio 3}

### Riesgos
- {riesgo 1 y mitigación}
- {riesgo 2 y mitigación}

### Rollback
{cómo deshacer los cambios si algo falla}
```

## Flujo de aprobación

1. **Presentar** el plan al usuario.
2. **Esperar** aprobación explícita antes de cualquier ejecución.
3. **Respuestas válidas**:
   - `aprobar`, `sí`, `proceed` → ejecutar
   - Feedback de modificación → actualizar plan, re-presentar
   - `no`, `cancelar` → detener, pedir nuevas instrucciones

## Reglas

- **Nunca** ejecutar sin aprobación del usuario cuando el plan está activo.
- **Siempre** incluir criterios de verificación medibles.
- **Siempre** incluir un plan de rollback.
- Los planes simples (1-2 archivos, sin dependencias) no requieren este protocolo.
- Si durante la ejecución se descubre que el plan es inválido, detener y re-planificar.

## Ejemplo de plan simple

```markdown
## Plan: Añadir validación de email al formulario de registro

### Resumen
Añadir validación de formato de email en el frontend y backend.

### Pasos
**Paso 1: Validación frontend**
- Archivos: `src/components/RegisterForm.tsx`
- Comando: Añadir regex de validación + mensaje de error
- Verificación: Email inválido muestra error sin enviar formulario

**Paso 2: Validación backend**
- Archivos: `api/routes/auth.py`
- Comando: Añadir Pydantic validator para email
- Verificación: Petición con email inválido devuelve 422

### Criterios de Verificación
- [ ] Email sin @ rechazado en frontend
- [ ] Email sin dominio rechazado en frontend
- [ ] Email inválido rechazado en backend con 422
- [ ] Email válido pasa ambas validaciones

### Riesgos
- Regex demasiado restrictivo puede rechazar emails válidos → usar regex RFC 5322 simplificada

### Rollback
Revertir los dos commits de validación frontend y backend.
```
