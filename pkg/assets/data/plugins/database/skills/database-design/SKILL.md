---
name: database-design
description: >
  Usar cuando se diseña esquema de base de datos, se elige entre SQL/NoSQL, se definen
  índices, se modelan relaciones, se planifican migraciones o se selecciona ORM.
  Activador: "diseño de base de datos", "esquema SQL", "índices", "normalización",
  "PostgreSQL", "MongoDB", "SQLAlchemy", "Prisma", "migraciones", "relaciones",
  "foreign key", "query lento", "N+1".
version: "1.0"
---

# Diseño de Base de Datos

## Selección SQL vs NoSQL

**Usar SQL (PostgreSQL) cuando:**
- Datos relacionales con integridad referencial
- Transacciones ACID necesarias
- Consultas complejas con JOINs
- Esquema relativamente estable
- Reporting / analytics

**Usar NoSQL cuando:**
- Documentos sin esquema fijo (MongoDB)
- Cache / sesiones (Redis)
- Time-series (InfluxDB, TimescaleDB)
- Búsqueda full-text (Elasticsearch)
- Grafos (Neo4j)

**Regla práctica:** Empezar con PostgreSQL. Añadir NoSQL solo para casos de uso específicos.

## Principios de Esquema

### Normalización

```sql
-- ❌ MAL — datos duplicados
CREATE TABLE orders (
    id         BIGSERIAL PRIMARY KEY,
    user_email VARCHAR(255),  -- duplicado si el usuario cambia email
    user_name  VARCHAR(255),  -- idem
    product    VARCHAR(255)   -- ¿qué pasa si cambia el nombre del producto?
);

-- ✅ BIEN — referencias normalizadas
CREATE TABLE users (
    id    BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name  VARCHAR(255) NOT NULL
);

CREATE TABLE products (
    id    BIGSERIAL PRIMARY KEY,
    name  VARCHAR(255) NOT NULL,
    price NUMERIC(10,2) NOT NULL CHECK (price >= 0)
);

CREATE TABLE orders (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE order_items (
    order_id   BIGINT NOT NULL REFERENCES orders(id),
    product_id BIGINT NOT NULL REFERENCES products(id),
    quantity   INT NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(10,2) NOT NULL,  -- precio en momento de compra
    PRIMARY KEY (order_id, product_id)
);
```

### Tipos de Datos

```sql
-- ✅ Usar tipos semánticamente correctos
id          BIGSERIAL / UUID          -- BIGSERIAL para claves internas, UUID para APIs públicas
email       VARCHAR(255)              -- no TEXT para campos con longitud conocida
price       NUMERIC(10,2)            -- nunca FLOAT para dinero (imprecisión)
created_at  TIMESTAMPTZ              -- siempre con zona horaria
is_active   BOOLEAN DEFAULT TRUE     -- no SMALLINT
metadata    JSONB                     -- JSONB (indexable) > JSON
```

### Constraints Siempre

```sql
-- Validaciones en la base de datos, no solo en la aplicación
ALTER TABLE users ADD CONSTRAINT email_format
    CHECK (email ~* '^[^@]+@[^@]+\.[^@]+$');

ALTER TABLE products ADD CONSTRAINT positive_price
    CHECK (price >= 0);

-- NOT NULL en campos obligatorios
-- UNIQUE en campos que deben serlo
-- DEFAULT en campos con valor habitual
```

## Índices

### Cuándo Crear

```sql
-- ✅ Siempre en foreign keys (PostgreSQL no los crea automáticamente)
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);

-- ✅ En columnas de búsqueda frecuente
CREATE INDEX idx_users_email ON users(email);  -- login

-- ✅ Índice parcial para queries con filtro común
CREATE INDEX idx_orders_active ON orders(user_id)
    WHERE status = 'active';  -- solo orders activas

-- ✅ Índice compuesto cuando siempre se filtra por ambas columnas
CREATE INDEX idx_events_user_created ON events(user_id, created_at DESC);

-- ❌ No indexar todo — cada índice ralentiza escrituras
```

### Detección de N+1

```python
# ❌ N+1 — una query por cada orden
orders = session.query(Order).all()
for order in orders:
    print(order.user.name)  # query por cada iteración

# ✅ Eager loading
from sqlalchemy.orm import joinedload

orders = session.query(Order).options(joinedload(Order.user)).all()
for order in orders:
    print(order.user.name)  # sin queries adicionales
```

## Migraciones

### Principios

```sql
-- ✅ Migraciones siempre hacia adelante — nunca editar las existentes
-- ✅ Cada migración debe ser reversible (down migration)
-- ✅ Zero-downtime: añadir columnas con DEFAULT primero, NOT NULL después

-- Paso 1: Añadir columna nullable (zero-downtime)
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- Paso 2: Rellenar datos existentes
UPDATE users SET phone = '' WHERE phone IS NULL;

-- Paso 3: Añadir constraint NOT NULL (después del deploy)
ALTER TABLE users ALTER COLUMN phone SET NOT NULL;
```

### Con Alembic (Python)

```bash
# Generar migración
alembic revision --autogenerate -m "add_phone_to_users"

# Aplicar
alembic upgrade head

# Revertir
alembic downgrade -1
```

### Con Prisma (TypeScript)

```bash
# Generar migración
npx prisma migrate dev --name add_phone_to_users

# Aplicar en producción
npx prisma migrate deploy
```

## Selección de ORM

| ORM | Lenguaje | Cuándo Usarlo |
|-----|----------|---------------|
| **SQLAlchemy 2** | Python | Aplicaciones complejas, control fino |
| **Prisma** | TypeScript | Developer experience, type-safe queries |
| **Drizzle** | TypeScript | Ligero, similar a SQL puro, Edge-compatible |
| **TypeORM** | TypeScript | Proyectos NestJS legacy |
| **Django ORM** | Python | Proyectos Django |

## Anti-patrones Comunes

```sql
-- ❌ SELECT * en producción
SELECT * FROM users;  -- trae columnas innecesarias, rompe cuando añades columnas

-- ✅ Siempre columnas explícitas
SELECT id, email, name FROM users;

-- ❌ DELETE sin WHERE
DELETE FROM sessions;  -- borra todo

-- ✅ Siempre con WHERE y límite
DELETE FROM sessions WHERE expires_at < NOW() LIMIT 1000;

-- ❌ Lógica de negocio en triggers
-- ✅ Triggers solo para auditoría y updated_at
```
