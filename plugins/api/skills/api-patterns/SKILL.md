---
name: api-patterns
description: >
  Usar cuando se diseña o revisa una API: REST vs GraphQL vs tRPC, versionado,
  paginación, manejo de errores HTTP, autenticación, rate limiting, documentación
  OpenAPI. Activador: "diseñar API", "endpoints REST", "GraphQL", "tRPC",
  "paginación", "versionado API", "OpenAPI", "swagger", "autenticación JWT",
  "rate limiting", "HTTP status codes".
version: "1.0"
---

# Patrones de API

## REST vs GraphQL vs tRPC

| Criterio | REST | GraphQL | tRPC |
|----------|------|---------|------|
| Cuando usarlo | API pública / terceros | Datos complejos / móvil | Monorepo TypeScript full-stack |
| Versionado | Explícito (v1, v2) | Deprecación de campos | Automático por tipos |
| Over/under-fetching | Problema real | Resuelto | Resuelto |
| Caché | Fácil (HTTP) | Complejo | Moderado |
| Curva aprendizaje | Baja | Alta | Media |
| Type safety | Manual | Generado | Automático |

## REST — Convenciones

### URLs y Métodos

```
# ✅ Recursos en plural, sustantivos (no verbos)
GET    /users              # listar
GET    /users/:id          # obtener uno
POST   /users              # crear
PATCH  /users/:id          # actualizar parcial
PUT    /users/:id          # reemplazar completo
DELETE /users/:id          # eliminar

# ✅ Recursos anidados para relaciones
GET    /users/:id/orders   # orders de un usuario
POST   /users/:id/orders   # crear order para usuario

# ❌ Verbos en la URL
GET    /getUsers
POST   /createUser
POST   /users/delete/:id
```

### Códigos HTTP

```
200 OK               — GET exitoso, PATCH exitoso
201 Created          — POST exitoso (incluir Location header)
204 No Content       — DELETE exitoso, PATCH sin cuerpo de respuesta
400 Bad Request      — Validación fallida (incluir detalles del error)
401 Unauthorized     — No autenticado
403 Forbidden        — Autenticado pero sin permisos
404 Not Found        — Recurso no existe
409 Conflict         — Conflicto de estado (email duplicado, etc.)
422 Unprocessable    — Entidad semánticamente inválida
429 Too Many Req.    — Rate limit alcanzado
500 Internal Error   — Error del servidor (no exponer detalles)
```

### Formato de Error Estándar

```json
// RFC 7807 — Problem Details for HTTP APIs
{
  "type": "https://api.example.com/errors/validation",
  "title": "Validation Error",
  "status": 400,
  "detail": "The request body contains invalid fields",
  "instance": "/users/create",
  "errors": {
    "email": "Must be a valid email address",
    "age": "Must be a positive integer"
  }
}
```

## Paginación

```json
// ✅ Cursor-based (mejor para datos que cambian)
{
  "data": [...],
  "pagination": {
    "next_cursor": "eyJpZCI6MTAwfQ",
    "has_more": true,
    "limit": 20
  }
}

// Uso: GET /users?cursor=eyJpZCI6MTAwfQ&limit=20

// ✅ Offset-based (más simple, suficiente para tablas de admin)
{
  "data": [...],
  "pagination": {
    "page": 2,
    "per_page": 20,
    "total": 450,
    "total_pages": 23
  }
}
```

## Versionado

```
# ✅ En la URL — más explícito, más fácil de deprecar
GET /v1/users
GET /v2/users

# ✅ En header Accept — más REST-puro
Accept: application/vnd.api+json; version=2

# ❌ En query param — no estándar
GET /users?version=2
```

**Política de deprecación:**
1. Anunciar deprecación con `Sunset` header: `Sunset: Sat, 31 Dec 2025 23:59:59 GMT`
2. Mantener versión antigua ≥6 meses después del anuncio
3. Documentar cambios en changelog

## Autenticación

```
# ✅ JWT Bearer en Authorization header
Authorization: Bearer eyJhbGc...

# ✅ Estructura JWT
{
  "sub": "user_123",       # subject (user ID)
  "iat": 1700000000,       # issued at
  "exp": 1700003600,       # expiration (1h)
  "jti": "uuid",           # JWT ID (para revocación)
  "roles": ["admin"]
}

# ✅ Access token corto + Refresh token largo
# Access token: 15min - 1h
# Refresh token: 7-30 días, rotación en cada uso
```

## Rate Limiting

```
# Headers de respuesta estándar
X-RateLimit-Limit: 100        # requests por ventana
X-RateLimit-Remaining: 42     # requests restantes
X-RateLimit-Reset: 1700003600 # timestamp de reset (Unix)
Retry-After: 60               # segundos hasta poder reintentar (429)
```

## OpenAPI / Documentación

```yaml
# openapi.yaml mínimo
openapi: "3.1.0"
info:
  title: Mi API
  version: "1.0"

paths:
  /users/{id}:
    get:
      summary: Obtener usuario por ID
      parameters:
        - name: id
          in: path
          required: true
          schema: { type: integer, format: int64 }
      responses:
        "200":
          description: Usuario encontrado
          content:
            application/json:
              schema: { $ref: '#/components/schemas/User' }
        "404":
          $ref: '#/components/responses/NotFound'
```

**Herramientas recomendadas:**
- **FastAPI** — genera OpenAPI automáticamente
- **Zod + zod-openapi** — TypeScript schema → OpenAPI
- **Stoplight** — editor visual de specs
