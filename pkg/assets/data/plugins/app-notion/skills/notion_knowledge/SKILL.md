---
name: notion_knowledge
description: >
  Integración con Notion para gestión de conocimiento y documentación.
  Trigger: /notion, "busca en Notion", "guarda en docs"
license: MIT
metadata:
  author: juarvis-org
  version: "1.0"
---

# Notion Integration Skill

## Propósito
Conecta Juarvis con Notion para:
- Buscar en bases de conocimiento
- Crear y actualizar páginas
- Gestionar databases
- Sincronizar documentación

## Herramientas MCP

| Herramienta | Descripción |
|-----------|-----------|
| `search` | Buscar en Notion |
| `get_page` | Obtener contenido de página |
| `create_page` | Crear nueva página |
| `update_page` | Actualizar página |
| `query_database` | Consultar database |

## Usage

### Buscar Información
```
1. Término de búsqueda
2. search en Notion
3. Mostrar resultados relevantes
4. Permitir profundizar
```

### Documentar Decisiones
```
1. Decisión tomada en SDD
2. Crear página en docs database
3. Incluir contexto y rationale
4. Link desde SPEC.md
```

### Actualizar Estado
```
1. Cambios en proyecto
2. Update página de estado
3. Incluir métricas
4. Notificar en Slack
```

## Workflows

### Workflow 1: Decision Log
```
1. al final de sesión SDD
2. create_page en "Decisions" database
3. Documentar: decisión, alternativas, rationale
```

### Workflow 2: Project Docs
```
1. Nuevo proyecto
2. Create página en "Projects"
3. Añadir: objetivo, stack, timeline
4. Link desde README.md
```

### Workflow 3: Knowledge Base
```
1. Nuevo aprendizaje
2. Buscar en database existente
3. Crear o updating según necesidad
4. Taggear apropiadamente
```

## Config
```
NOTION_API_KEY=secret_...
NOTION_DATABASE_IDS={"decisions":"...","projects":"..."}
```

## Mejores Prácticas

1. **Usar databases** para organizatión estructurada
2. **Tags** para categorización
3. **Templates** para consistencia
4. **Links** bidireccionales
5. **Sincronizar** con proyecto