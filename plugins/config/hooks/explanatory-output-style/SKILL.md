# Explanatory Output Style

Skill que activa el modo pedagógico de salida explicativa.

## Descripción

Este skill recrea el modo de salida explicativa como un hook de SessionStart.
Cuando está activo, inyecta instrucciones al inicio de cada sesión para que el
asistente proporcione contexto educativo mientras completa las tareas.

## Comportamiento

Cuando este skill está activo, el asistente:

1. Ofrece contexto educativo sobre las decisiones de implementación
2. Explica los patrones y convenciones del código base
3. Equilibra la completación de tareas con oportunidades de aprendizaje

## Formato de salida

Los insights educativos se presentan con el siguiente formato:

```
★ Insight ─────────────────────────────────────
[2-3 puntos educativos clave]
───────────────────────────────────────────────
```

## Enfoque

Los insights se centran en:

- Decisiones de implementación específicas del código base
- Patrones y convenciones del código
- Compromisos y decisiones de diseño
- Detalles específicos del proyecto, no conceptos generales de programación

## Activación

El hook SessionStart se registra en `hooks/explanatory-output-style/hooks.json`
y ejecuta el script `hooks/explanatory-output-style/hooks-handlers/session-start.sh`
al inicio de cada sesión.

## Coste en tokens

**Aviso:** Este modo incrementa el consumo de tokens por sesión debido al
contexto adicional y las explicaciones generadas. Úsalo cuando el aprendizaje
sea prioritario sobre la eficiencia.
