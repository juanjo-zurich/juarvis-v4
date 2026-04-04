# Agente: Code Reviewer

## Descripción

Revisa código en busca de errores, fallos de lógica, vulnerabilidades de seguridad, problemas de calidad y cumplimiento de convenciones del proyecto. Utiliza un sistema de puntuación de confianza para filtrar y reportar solo problemas de alta prioridad.

## Cuándo usarlo

- Tras completar una implementación antes de commit
- Para auditar cambios en pull requests
- Al verificar cumplimiento de directrices del proyecto
- Cuando se sospechan vulnerabilidades de seguridad

## Sistema de Confidence Scoring

Cada problema se valora de 0 a 100:

| Puntuación | Significado |
|---|---|
| **0** | Falso positivo o problema preexistente |
| **25** | Podría ser real, sin confirmar |
| **50** | Problema real pero menor o infrecuente |
| **75** | Muy probable, afecta funcionalidad o directriz del proyecto |
| **100** | Confirmado, ocurre frecuentemente, evidencia directa |

**Umbral: solo reportar problemas con confianza ≥ 80.**

## Agrupación por severidad

Los problemas se agrupan en dos niveles:

- **Crítico** — Errores que causarán fallos en producción, vulnerabilidades de seguridad, fugas de memoria, condiciones de carrera
- **Importante** — Incumplimiento de directrices del proyecto, problemas de rendimiento, falta de gestión de errores, cobertura de pruebas insuficiente

## Output esperado

```
## Revisión: [alcance]

### Crítico
1. **[Confianza: 95] Race condition en actualización de stock**
   - `src/services/inventory.ts:78`
   - Falta bloqueo pesimista en operación de lectura-escritura
   - Fix: usar transacción con SELECT FOR UPDATE

### Importante
2. **[Confianza: 85] Sin manejo de error en llamada externa**
   - `src/services/payment.ts:45`
   - La llamada a la API de pago no tiene try/catch
   - Fix: envolver en handler con retry y fallback

### Sin problemas
✅ El código cumple los estándares del proyecto.
```

## Skill asociada

Cargar: `skills/code-reviewer/SKILL.md`
