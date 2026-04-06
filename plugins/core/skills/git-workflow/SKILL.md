---
name: git-workflow
description: Flujo de trabajo de git: conventional commits, branching strategy, PRs
trigger: git commit, git branch, git merge, pull request, conventional commit
---

# Git Workflow

## Conventional Commits

Formato: `type(scope): descripción`

| Type | Cuándo usar |
|------|-------------|
| `feat` | Nueva funcionalidad |
| `fix` | Corrección de bug |
| `docs` | Solo documentación |
| `refactor` | Refactorización sin cambio de comportamiento |
| `perf` | Mejora de rendimiento |
| `test` | Añadir o corregir tests |
| `chore` | Cambios de build, CI, dependencias |

Ejemplos:
- `feat: add HTTP rate limiting and retry`
- `fix: race condition in pluginCache`
- `docs: update README with new commands`
- `refactor: deduplicate frontmatter parsing`
- `chore: remove duplicate skill-create command`

## Cuándo hacer commit

- **Haz commits pequeños y frecuentes** — cada commit debe ser una unidad lógica
- **Un commit = un cambio conceptual** — no mezcles fixes con features
- **Commitea tras cada fase completada** — no esperes a terminar todo

## Escribir mensajes de commit útiles

- **Primera línea**: qué cambió y por qué (máx 72 chars)
- **Cuerpo** (opcional): detalles técnicos, contexto, decisiones

Buen ejemplo:
```
fix: race condition in pluginCache — RLock → Lock for delete operation

delete(pluginCache, name) se ejecutaba bajo RLock (read lock),
lo cual es una operación de escritura. Cambiado a Lock/Unlock.
```

Mal ejemplo:
```
fix stuff
```

## Estrategia de ramas

- **main** — siempre estable, protegida por CI
- **feature/** — nuevas funcionalidades
- **fix/** — correcciones de bugs urgentes
- **refactor/** — refactorizaciones grandes

## Pull Requests

- Crea PR para cambios sustanciales (>50 líneas o cambios de arquitectura)
- El CI debe pasar antes de merge
- Describe qué cambió y por qué en la descripción del PR
