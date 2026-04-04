---
name: backend-patterns
description: Patrones de arquitectura backend, diseño de APIs, optimizacion de bases de datos y mejores practicas server-side
trigger: backend, API, REST, GraphQL, repository, service layer, middleware, caching, auth, rate limiting, colas
---

# Backend Development Patterns

Patrones de arquitectura backend para aplicaciones server-side escalables y mantenibles.

## Cuando Activar

- Disenando endpoints REST o GraphQL
- Implementando capas de repositorio, servicio o controlador
- Optimizando consultas de base de datos (N+1, indexacion, connection pooling)
- Anadiendo caching (Redis, memoria, HTTP cache headers)
- Configurando jobs en segundo plano o procesamiento async
- Estructurando manejo de errores y validacion para APIs
- Construyendo middleware (auth, logging, rate limiting)

## Patrones de Diseno de API

### Estructura RESTful

```
// PASS: URLs basadas en recursos
GET    /api/recursos              # Listar recursos
GET    /api/recursos/:id          # Obtener recurso individual
POST   /api/recursos              # Crear recurso
PUT    /api/recursos/:id          # Reemplazar recurso
PATCH  /api/recursos/:id          # Actualizar parcialmente
DELETE /api/recursos/:id          # Eliminar recurso

// PASS: Parametros de query para filtrado, ordenacion, paginacion
GET /api/recursos?estado=activo&orden=volumen&limite=20&offset=0
```

### Repository Pattern

Abstrae la logica de acceso a datos:

```typescript
interface Repositorio<T> {
  findAll(filtros?: Filtros): Promise<T[]>
  findById(id: string): Promise<T | null>
  create(data: CreateDto): Promise<T>
  update(id: string, data: UpdateDto): Promise<T>
  delete(id: string): Promise<void>
}
```

**Reglas**:
- La capa de repositorio SOLO accede a datos — sin logica de negocio
- Cada entidad tiene su propio repositorio
- Los filtros se pasan como objetos tipados, no como strings sueltos

### Service Layer Pattern

Separa la logica de negocio del acceso a datos:

```typescript
class Servicio {
  constructor(private repo: Repositorio) {}

  async buscar(query: string, limite: number = 10): Promise<Resultado[]> {
    // 1. Logica de negocio aqui
    // 2. Delegar acceso a datos al repositorio
    // 3. Transformar y retornar resultado
  }
}
```

**Reglas**:
- El servicio NUNCA accede directamente a la base de datos
- El servicio ORQUESTA repositorios, no los reemplaza
- Un servicio puede usar multiples repositorios

### Middleware Pattern

Pipeline de procesamiento de requests:

```typescript
function withAuth(handler: Handler): Handler {
  return async (req, res) => {
    const token = req.headers.authorization?.replace('Bearer ', '')
    if (!token) return res.status(401).json({ error: 'Unauthorized' })
    try {
      const user = await verifyToken(token)
      req.user = user
      return handler(req, res)
    } catch {
      return res.status(401).json({ error: 'Invalid token' })
    }
  }
}
```

**Orden tipico de middleware**:
1. Logging/Request ID
2. CORS
3. Rate limiting
4. Auth
5. Validacion de input
6. Handler

## Patrones de Base de Datos

### Optimizacion de Consultas

```
// PASS: Seleccionar solo columnas necesarias
SELECT id, nombre, estado FROM recursos WHERE estado = 'activo' LIMIT 10

// FAIL: Seleccionar todo
SELECT * FROM recursos
```

### Prevencion de N+1

```
// FAIL: N+1 queries
for recurso in recursos:
    recurso.creador = obtener_usuario(recurso.creador_id)  // N queries

// PASS: Batch fetch
creador_ids = [r.creador_id for r in recursos]
creadores = obtener_usuarios(creador_ids)  // 1 query
creador_map = {c.id: c for c in creadores}
for recurso in recursos:
    recurso.creador = creador_map.get(recurso.creador_id)
```

### Patron de Transaccion

- Usar transacciones de base de datos para operaciones que modifican multiples tablas
- Si una operacion falla, TODO se revierte automaticamente
- Mantener transacciones lo mas cortas posible

## Estrategias de Caching

### Cache-Aside Pattern

```
1. Intentar leer de cache
2. Si hay cache hit → retornar
3. Si hay cache miss → leer de BD → guardar en cache → retornar
4. Invalidar cache al actualizar/eliminar
```

**TTL recomendados**:
- Datos estaticos: 1 hora+
- Datos de usuario: 5-15 minutos
- Datos en tiempo real: sin cache o TTL muy corto (segundos)

### Cuando NO cachear

- Datos que cambian frecuentemente
- Datos sensibles sin encriptar en cache
- Resultados de consultas complejas con muchos parametros

## Manejo de Errores

### Error Centralizado

```typescript
class ApiError extends Error {
  constructor(
    public statusCode: number,
    public message: string,
    public isOperational = true
  ) {
    super(message)
  }
}

// Todos los errores operacionales extienden ApiError
// Los errores no operacionales (bugs) se loguean y retornan 500 generico
```

### Retry con Exponential Backoff

```
Intento 1: esperar 1s
Intento 2: esperar 2s
Intento 3: esperar 4s
Maximo: 3 intentos
```

**Solo reintentar errores transitorios**: timeouts, 503, connection refused.
**Nunca reintentar**: 400, 401, 403, 404, errores de validacion.

## Auth y Autorizacion

### JWT Token Validation

- Verificar firma, expiracion, issuer en cada request
- Usar refresh tokens para sesiones largas
- Nunca almacenar datos sensibles en el payload del JWT

### Role-Based Access Control (RBAC)

```
admin: [read, write, delete, admin]
moderator: [read, write, delete]
user: [read, write]
```

- Verificar permisos ANTES de ejecutar la operacion
- Fallar con 403, no con 401 si el usuario esta autenticado pero no tiene permiso

## Rate Limiting

- Implementar por IP y/o por usuario autenticado
- Limites tipicos: 100 req/min para usuarios, 20 req/min para anonimos
- Retornar 429 con header `Retry-After`
- Usar Redis para rate limiting distribuido

## Background Jobs y Colas

- Operaciones lentas (enviar email, procesar imagen, generar PDF) → cola
- El endpoint retorna inmediatamente con "job queued"
- Worker procesa la cola en segundo plano
- Implementar retry con backoff para jobs fallidos

## Logging Estructurado

```json
{
  "timestamp": "2026-04-04T12:00:00Z",
  "level": "info",
  "message": "Request procesado",
  "requestId": "abc-123",
  "method": "GET",
  "path": "/api/recursos",
  "duration": 45,
  "userId": "user-456"
}
```

**Reglas**:
- Loguear cada request con request ID unico
- Loguear errores con stack trace
- NUNCA loguear passwords, tokens, datos sensibles
- Usar niveles: debug, info, warn, error

**Recuerda**: Los patrones backend habilitan aplicaciones escalables y mantenibles. Elige patrones que se ajusten a tu nivel de complejidad — no sobre-ingenieries un CRUD simple.
