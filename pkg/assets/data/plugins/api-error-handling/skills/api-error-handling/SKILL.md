---
name: api-error-handling
description: Patrones avanzados de manejo de errores en APIs: retry, circuit breaker, error classification, graceful degradation
trigger: error handling, retry, backoff, circuit breaker, API errors, error classification, graceful degradation
---

# API Error Handling & Recovery

Patrones robustos para manejar errores en APIs de forma profesional y predecible.

## Cuando Activar

- Disenando el sistema de errores de una API
- Implementando reintentos para llamadas externas
- Necesitando circuit breaker para servicios dependientes
- Clasificando tipos de errores para diferentes respuestas
- Implementando graceful degradation

## Clasificacion de Errores

### Errores Operacionales vs Programaticos

```
Operacionales (esperados, manejables):
- Base de datos no disponible temporalmente
- API externa con timeout
- Rate limit excedido
- Recurso no encontrado
- Validacion fallida

Programaticos (bugs, no recuperables):
- TypeError: undefined no tiene propiedad X
- SyntaxError en JSON.parse
- Memory leak
- Deadlock
```

**Regla**: Los errores operacionales se manejan y recuperan. Los programaticos se loguean y el proceso se reinicia.

## Patrones de Retry

### Exponential Backoff con Jitter

```typescript
async function fetchWithRetry<T>(
  fn: () => Promise<T>,
  options: { maxRetries?: number, baseDelay?: number, maxDelay?: number } = {}
): Promise<T> {
  const { maxRetries = 3, baseDelay = 1000, maxDelay = 30000 } = options
  let lastError: Error

  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn()
    } catch (error) {
      lastError = error as Error

      // No reintentar errores del cliente
      if (isClientError(error)) throw error

      if (i < maxRetries - 1) {
        // Exponential backoff + jitter
        const delay = Math.min(
          baseDelay * Math.pow(2, i) + Math.random() * baseDelay,
          maxDelay
        )
        await sleep(delay)
      }
    }
  }

  throw lastError!
}
```

**Cuando reintentar**:
- ✅ Timeouts (504)
- ✅ Service unavailable (503)
- ✅ Too many requests (429) — respetar header Retry-After
- ✅ Connection refused temporal

**Cuando NO reintentar**:
- ❌ Bad request (400)
- ❌ Unauthorized (401)
- ❌ Forbidden (403)
- ❌ Not found (404)
- ❌ Validation errors
- ❌ Errores de logica de negocio

### Retry Idempotente

Solo reintentar operaciones idempotentes:
- GET, PUT, DELETE → seguros para reintentar
- POST → solo si el servidor soporta idempotency keys
- PATCH → depende de la implementacion

## Circuit Breaker Pattern

```
Estados del circuit breaker:

CLOSED (normal) → Las requests pasan normalmente
  ↓ (fallos consecutivos >= threshold)
OPEN (circuito abierto) → Las requests fallan inmediatamente
  ↓ (despues de timeout)
HALF-OPEN (prueba) → Una request de prueba
  ↓ (exito)
CLOSED (recuperado)

  ↓ (fallo)
OPEN (vuelve a abrirse)
```

**Configuracion tipica**:
- `failureThreshold`: 5 fallos consecutivos
- `resetTimeout`: 30 segundos
- `halfOpenMaxRequests`: 1

**Implementacion conceptual**:

```typescript
class CircuitBreaker {
  private state: 'CLOSED' | 'OPEN' | 'HALF_OPEN' = 'CLOSED'
  private failures = 0
  private lastFailureTime = 0

  async execute<T>(fn: () => Promise<T>): Promise<T> {
    if (this.state === 'OPEN') {
      if (Date.now() - this.lastFailureTime > this.resetTimeout) {
        this.state = 'HALF_OPEN'
      } else {
        throw new Error('Circuit breaker abierto')
      }
    }

    try {
      const result = await fn()
      this.onSuccess()
      return result
    } catch (error) {
      this.onFailure()
      throw error
    }
  }

  private onSuccess() {
    this.failures = 0
    this.state = 'CLOSED'
  }

  private onFailure() {
    this.failures++
    this.lastFailureTime = Date.now()
    if (this.failures >= this.failureThreshold) {
      this.state = 'OPEN'
    }
  }
}
```

## Error Response Estandar

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "El campo 'email' es obligatorio",
    "details": [
      { "field": "email", "message": "es obligatorio" },
      { "field": "password", "message": "minimo 8 caracteres" }
    ],
    "requestId": "req-abc-123",
    "timestamp": "2026-04-04T12:00:00Z"
  }
}
```

**Codigos de error comunes**:
- `VALIDATION_ERROR` — Input invalido (400)
- `NOT_FOUND` — Recurso no existe (404)
- `UNAUTHORIZED` — Sin autenticacion (401)
- `FORBIDDEN` — Sin permisos (403)
- `CONFLICT` — Conflicto de estado (409)
- `RATE_LIMITED` — Demasiadas requests (429)
- `INTERNAL_ERROR` — Error del servidor (500)
- `SERVICE_UNAVAILABLE` — Servicio dependiente caido (503)

## Graceful Degradation

Cuando un servicio dependiente falla, ofrecer funcionalidad reducida:

```
Servicio principal: Busqueda con IA
  ↓ (servicio de IA caido)
Degradacion: Busqueda por texto simple
  ↓ (BD tambien caida)
Degradacion: Cache local de resultados recientes
  ↓ (todo caido)
Degradacion: Pagina de mantenimiento con ETA
```

**Reglas**:
- Definir niveles de degradacion ANTES de que ocurra el fallo
- Cada nivel debe ser funcional por si solo
- Loguear cuando se activa la degradacion
- Alertar al equipo cuando se alcanza el nivel mas bajo

## Timeout Patterns

```typescript
async function withTimeout<T>(
  fn: () => Promise<T>,
  timeoutMs: number
): Promise<T> {
  const timeout = new Promise<never>((_, reject) => {
    setTimeout(() => reject(new Error(`Timeout despues de ${timeoutMs}ms`)), timeoutMs)
  })

  return Promise.race([fn(), timeout])
}
```

**Timeouts recomendados**:
- API externa: 5-10 segundos
- Base de datos: 2-5 segundos
- Cache (Redis): 500ms - 1 segundo
- Operacion interna: depende de la complejidad

**Recuerda**: Un buen sistema de errores hace la diferencia entre una API profesional y una que frustra a los consumidores. Cada error debe ser informativo, accionable y consistente.
