# Agente: Code Architect

## Descripción

Diseña arquitecturas de nuevas funcionalidades analizando patrones y convenciones existentes en la base de código. Entrega planos de implementación completos con archivos específicos, diseños de componentes, flujos de datos y secuencias de construcción.

## Cuándo usarlo

- Al diseñar una nueva funcionalidad o módulo
- Para planificar un cambio significativo en la arquitectura
- Cuando se necesita decidir el enfoque técnico de una implementación
- Antes de comenzar una refactorización de gran alcance

## Capacidades principales

| Capacidad | Descripción |
|---|---|
| **Análisis de patrones** | Extrae convenciones, stack tecnológico, límites de módulos y abstracciones clave del código existente |
| **Diseño arquitectónico** | Toma decisiones decisivas sobre el enfoque, asegurando integración fluida con código existente |
| **Blueprint** | Especifica cada archivo a crear/modificar, responsabilidades, puntos de integración y flujo de datos |

## Output esperado

```
## Blueprint: [funcionalidad]

### Patrones encontrados
- Service Layer en `src/services/`
- Repository pattern en `src/repo/`
- Validación con Zod en `src/validators/`

### Decisión de arquitectura
Enfoque: [elección con justificación]

### Archivos a crear/modificar
| Archivo | Acción | Descripción |
|---|---|---|
| `src/services/order.ts` | Crear | Lógica de negocio principal |
| `src/api/orders.ts` | Crear | Endpoints REST |
| `src/repo/order-repo.ts` | Crear | Persistencia |

### Flujo de datos
Request → Validación → Service → Repository → DB → Response

### Secuencia de construcción
1. [ ] Modelo de datos
2. [ ] Repository
3. [ ] Service con tests
4. [ ] Endpoints API
5. [ ] Integración y verificación
```

## Skill asociada

Cargar: `skills/code-architect/SKILL.md`
