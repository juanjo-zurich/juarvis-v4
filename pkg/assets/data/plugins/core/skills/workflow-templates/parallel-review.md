---
name: workflow-templates/parallel-review
description: Review paralelo por lenguaje. 3× code-reviewer ejecutándose simultáneamente
allowed-tools: Read, Glob, Grep
---

# Template: parallel-review

> Review paralelo por lenguaje de programación

## Overview

Este template ejecuta 3 revisiones en paralelo, una por cada lenguaje principal:
1. **Go**: Archivos `*.go`
2. **Python**: Archivos `*.py`
3. **JavaScript/TypeScript**: Archivos `*.{js,ts,tsx}`

Finalmente, un agente orquestador sintetiza los resultados.

## Configuración

```yaml
template: parallel-review
input: pr/files
description: Review paralelo por lenguaje

flow:
  - id: 1
    agent: code-reviewer
    params:
      target: Go
      files: "*.go"
    description: Review archivos Go
    
  - id: 2
    agent: code-reviewer
    params:
      target: Python
      files: "*.py"
    description: Review archivos Python
    
  - id: 3
    agent: code-reviewer
    params:
      target: JavaScript
      files: "*.{js,ts,tsx}"
    description: Review archivos JS/TS
    
  - id: 4
    agent: orchestrator
    description: Síntesis de resultados
    dependsOn: [1, 2, 3]
    action: merge_results

output:
  format: markdown
  sections:
    - Resumen por lenguaje
    - Problemas críticos consolidados
    - Score general
```

## Uso

```
Ejecutar parallel-review en los archivos del PR
```

## Estado Compartido

El `sharedContext` mantiene:
- `artifacts.review-go`: Resultados del review de Go
- `artifacts.review-python`: Resultados del review de Python
- `artifacts.review-js`: Resultados del review de JS/TS
- `artifacts.synthesized`: Resultado final consolidado

## Resultados por Lenguaje

| Lenguaje | Archivos | Problemas | Score |
|----------|----------|-----------|-------|
| Go | N | M | X/10 |
| Python | N | M | X/10 |
| JS/TS | N | M | X/10 |

## Fallback

- Si un code-reviewer falla → intentar refactor-assistant para ese lenguaje