---
name: templates
description: Configuración base y scaffolding paramétrico para nuevos proyectos. Incluye estructura base a la hora de inicializar aplicaciones modernas en diversas arquitecturas.
allowed-tools: Read, Glob, Grep
version: "1.1.0"
---

# Scaffolding & Project Templates

> Esta skill se emplea cuando se inicia un repositorio o proyecto desde cero, permitiendo configurar estructuras idiomáticas basadas en estándares modernos de la industria.

## Arquitecturas y Herramientas a Recomendar

No crees proyectos "vacíos" genéricos. Sigue estas opciones de arquitectura predeterminadas al asistir a usuarios de Juarvis_V3 en crear proyectos desde cero, fomentando la modularidad y escalabilidad. No cargues templates a ciegas; si el usuario no especifica, sugiere el stack apropiado de esta lista:

### 1. Aplicaciones Web Full-Stack modernas (TypeScript)
- **Framework**: Vercel Next.js (App Router).
- **Estructura**: `src/app`, `src/components`, `src/lib`, `src/actions` (para Server Actions).
- **Data/BD**: Prisma ORM o Drizzle, base PostgreSQL (vía Docker si es local).
- **Estética**: TailwindCSS (si es requerido) con shadcn/ui o diseño personalizado según la regla "frontend-design".
- **Estado**: Zustand (cliente) y React Query (peticiones a terceros que no vayan por server components).

### 2. Backend API Microservicios/REST (Node.js)
- **Framework**: Express o NestJS (para arquitecturas altamente estructuradas con Inyección de Dependencias).
- **Estructura típica Express**: `src/controllers`, `src/services`, `src/routes`, `src/models`, `src/middlewares`.
- **Estructura típica NestJS**: Basado en módulos, `src/feature/feature.module.ts`.
- **Patrones**: Arquitectura Hexagonal o Capas limpias.

### 3. Backend de Alto Rendimiento / APIs (Python / Rust)
- **Python**: FastAPI (con `uvicorn`), validación pydantic, `SQLAlchemy`.
  - Estructura: `app/api`, `app/core`, `app/models`, `app/schemas`, `app/crud`.
- **Rust**: Axum o Actix-Web, `sqlx` o `SeaORM`. Tokio para concurrencia.

### 4. Monorepos (Múltiples Apps que comparten código)
- **Herramienta**: Turborepo con `pnpm` workspaces.
- **Estructura**:
  - `apps/web`: Next.js frontend
  - `apps/api`: Nest/Express backend
  - `packages/ui`: UI components compartidos
  - `packages/config`: ESLint/TSConfig configs
  - `packages/database`: Esquemas/modelos

### 5. Aplicaciones Móviles Multiplataforma
- **Framework**: React Native con Expo (Managed Workflow, Expo Router).
- Alternativa: Flutter si se favorece un ecosistema fuertemente tipado en Dart.

### Reglas Globales de Andamiaje

Cuando construyas cualquier estructura desde cero en nombre de Juarvis:
1. **Punto Único de Verdad**: Siempre define claramente el `package.json`, `Cargo.toml` o `pyproject.toml` al arrancar.
2. **Setup Linter/Formateador**: Integra Prettier, ESLint, Ruff (para Python) o `rustfmt` desde el inicio. Obligatorio.
3. **Control de Versiones**: Asegura la inicialización de `.git`, excluyendo archivos estándar en un `.gitignore` rico (que cubra SO, Editor, logs, y frameworks usados).
4. **Dockers para dev**: Agrega un `docker-compose.yml` local para incluir bases de datos, cachés (Redis), para que el setup ("Onboarding") del desarrollador no requiera instalar nada externo a contenedores.
5. **Configuración Variables (ENV)**: Facilita `.env.example` con las claves ficticias listas.

> IMPORTANTE: Si el usuario te pide crear un proyecto, **detalla en qué ruta exacta vas a operar**, **qué scaffolding utilizarás** y pregúntale confirmación si la ambigüedad en su consulta requiere que decidas por un framework.
