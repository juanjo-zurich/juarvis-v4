---
name: nodejs-best-practices
description: >
  Usar cuando se desarrolla con Node.js o TypeScript: selección de framework,
  patrones async/await, seguridad, arquitectura de módulos, gestión de errores,
  variables de entorno, logging estructurado y optimización de rendimiento.
  Activador: "proyecto Node", "Express", "Fastify", "NestJS", "TypeScript backend",
  "API Node", "async patterns", "npm scripts", "tsconfig".
version: "1.0"
---

# Node.js / TypeScript — Mejores Prácticas

## Selección de Framework

| Caso de Uso | Framework Recomendado | Por Qué |
|-------------|----------------------|---------|
| API REST simple/media | **Fastify** | Más rápido que Express, validación integrada, TypeScript nativo |
| API compleja / enterprise | **NestJS** | DI, módulos, decoradores, muy opinionado |
| Microservicio mínimo | **Hono** | Ultra-ligero, edge-compatible, TSX nativo |
| BFF / full-stack | **Next.js (App Router)** | RSC, Server Actions, despliegue trivial |
| CLI | **Commander + Inquirer** | Estándar de facto |

## Configuración TypeScript

```json
// tsconfig.json mínimo para backend moderno
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "Node16",
    "moduleResolution": "Node16",
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "exactOptionalPropertyTypes": true,
    "outDir": "dist",
    "sourceMap": true
  }
}
```

**Reglas estrictas que activar siempre:**
- `strict: true` — engloba noImplicitAny, strictNullChecks, etc.
- `noUncheckedIndexedAccess` — `arr[0]` devuelve `T | undefined`
- `exactOptionalPropertyTypes` — diferencia `{ a?: string }` de `{ a: string | undefined }`

## Patrones Async

```typescript
// ✅ Promise.all para operaciones paralelas independientes
const [user, orders] = await Promise.all([
  getUser(id),
  getOrders(id)
])

// ✅ Promise.allSettled cuando el fallo parcial es aceptable
const results = await Promise.allSettled(ids.map(fetchItem))
const successful = results
  .filter((r): r is PromiseFulfilledResult<Item> => r.status === 'fulfilled')
  .map(r => r.value)

// ❌ Nunca await en bucle sin razón
for (const id of ids) {
  await fetchItem(id)  // secuencial innecesario — usa Promise.all
}

// ✅ Timeout explícito con AbortController
const controller = new AbortController()
const timeoutId = setTimeout(() => controller.abort(), 5000)
try {
  const res = await fetch(url, { signal: controller.signal })
} finally {
  clearTimeout(timeoutId)
}
```

## Gestión de Errores

```typescript
// ✅ Errores tipados con discriminated union
type AppError =
  | { type: 'NOT_FOUND'; resource: string; id: string }
  | { type: 'UNAUTHORIZED'; reason: string }
  | { type: 'VALIDATION'; fields: Record<string, string> }
  | { type: 'INTERNAL'; cause: unknown }

// ✅ Result type para errores esperados (evita try/catch)
type Result<T, E = AppError> = { ok: true; value: T } | { ok: false; error: E }

function parseUserId(raw: string): Result<UserId> {
  if (!raw.match(/^[0-9a-f-]{36}$/)) {
    return { ok: false, error: { type: 'VALIDATION', fields: { id: 'Invalid UUID' } } }
  }
  return { ok: true, value: raw as UserId }
}

// ✅ Handler global para errores no capturados
process.on('unhandledRejection', (reason) => {
  logger.error('Unhandled rejection', { reason })
  process.exit(1)
})
```

## Variables de Entorno

```typescript
// ✅ Validar entorno al arranque con zod
import { z } from 'zod'

const envSchema = z.object({
  NODE_ENV: z.enum(['development', 'test', 'production']),
  PORT: z.coerce.number().int().min(1).max(65535).default(3000),
  DATABASE_URL: z.string().url(),
  JWT_SECRET: z.string().min(32),
})

export const env = envSchema.parse(process.env)
// Si falla: proceso termina con mensaje claro de qué falta
```

## Estructura de Proyecto

```
src/
├── index.ts              # Entry point — solo arranque
├── app.ts                # Configuración del servidor
├── config/
│   └── env.ts            # Validación de entorno (arriba)
├── modules/
│   └── users/
│       ├── users.router.ts
│       ├── users.service.ts
│       ├── users.repository.ts
│       ├── users.schema.ts   # Validación Zod
│       └── users.types.ts
├── shared/
│   ├── errors.ts
│   ├── logger.ts
│   └── middleware/
└── __tests__/
```

## Logging

```typescript
// ✅ Structured logging con pino (nunca console.log en producción)
import pino from 'pino'

export const logger = pino({
  level: env.NODE_ENV === 'production' ? 'info' : 'debug',
  transport: env.NODE_ENV !== 'production'
    ? { target: 'pino-pretty' }
    : undefined,
})

// ✅ Contexto en cada log
logger.info({ userId, action: 'login', ip }, 'User logged in')
// ❌ Nunca:
console.log('User logged in')
logger.info('User ' + userId + ' logged in')
```

## Seguridad

```typescript
// ✅ Rate limiting (fastify-rate-limit / express-rate-limit)
// ✅ Helmet para cabeceras HTTP seguras
// ✅ Validación de entrada con Zod en cada endpoint
// ✅ Sanitizar antes de queries (usar ORM/query builder)
// ✅ JWT con expiración corta + refresh tokens
// ❌ Nunca almacenar secrets en código — solo env vars
// ❌ Nunca confiar en req.body sin validar schema
```

## Scripts npm Recomendados

```json
{
  "scripts": {
    "dev":       "tsx watch src/index.ts",
    "build":     "tsc --noEmit && tsup src/index.ts",
    "start":     "node dist/index.js",
    "test":      "vitest",
    "test:ci":   "vitest run --coverage",
    "typecheck": "tsc --noEmit",
    "lint":      "eslint src --max-warnings 0",
    "format":    "prettier --write src"
  }
}
```
