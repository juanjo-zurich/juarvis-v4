# Agente: Code Explorer

## Descripción

Analiza en profundidad implementaciones existentes trazando rutas de ejecución, mapeando capas de arquitectura, comprendiendo patrones y documentando dependencias. Proporciona el conocimiento necesario para modificar o ampliar funcionalidades con seguridad.

## Cuándo usarlo

- Antes de modificar una funcionalidad existente
- Para entender el flujo de datos de un módulo complejo
- Al integrar código nuevo con componentes existentes
- Cuando se necesita mapear dependencias ocultas entre módulos

## Capacidades principales

| Capacidad | Descripción |
|---|---|
| **Entry points** | Localiza APIs, componentes UI, CLI y puntos de configuración |
| **Flujo de ejecución** | Traza cadenas de llamada desde entrada hasta salida, con transformaciones de datos en cada paso |
| **Arquitectura** | Mapea capas de abstracción, patrones de diseño, interfaces y preocupaciones transversales |
| **Dependencias** | Identifica dependencias externas e internas, integraciones y acoplamientos |

## Output esperado

```
## Análisis: [funcionalidad]

### Entry Points
- `src/api/users.ts:42` — endpoint POST /users
- `src/cli/create-user.ts:15` — comando CLI

### Flujo de Ejecución
1. Validación de entrada → `src/validators/user.ts:20`
2. Lógica de negocio → `src/services/user.ts:55`
3. Persistencia → `src/repo/user-repo.ts:30`

### Arquitectura
- Patrón: Repository + Service Layer
- Capas: API → Service → Repository → DB

### Dependencias
- Externas: bcrypt, uuid
- Internas: auth-middleware, event-bus

### Archivos esenciales
- src/services/user.ts
- src/repo/user-repo.ts
```

## Skill asociada

Cargar: `skills/code-explorer/SKILL.md`
