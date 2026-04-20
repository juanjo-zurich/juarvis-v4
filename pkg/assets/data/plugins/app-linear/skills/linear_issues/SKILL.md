---
name: linear_issues
description: >
  Integración con Linear para gestión de issues y proyectos.
  Trigger: /linear, "crea issue", "actualiza estado"
license: MIT
metadata:
  author: juarvis-org
  version: "1.0"
---

# Linear Integration Skill

## Propósito
Conecta Juarvis con Linear para:
- Crear y gestionar issues
- Actualizar estados
- Sincronizar con desarrollo

## Herramientas MCP
| Herramienta | Descripción |
|-------------|-----------|
| `create_issue` | Crear nuevo issue |
| `list_issues` | Listar issues |
| `update_issue` | Actualizar estado |
| `get_issue` | Obtener issue específico |

## Uso
```
1. Identificar tarea del usuario
2. Create issue en Linear
3. Asignar a proyecto correcto
4. Añadir labels y prioridad
```

## Workflow: SDD a Issue
```
1. Fase SDD completada
2. create_issue condescription de tarea
3. Link en SPEC.md
4. Mover a "In Progress"
```

## Config
LINEAR_API_KEY=...