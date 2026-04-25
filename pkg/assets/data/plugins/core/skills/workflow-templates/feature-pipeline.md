---
name: workflow-templates/feature-pipeline
description: Pipeline completo para implementar una feature. Flow: explore → apply → test-runner → verify
allowed-tools: Read, Glob, Grep
---

# Template: feature-pipeline

> Pipeline completo para implementar una feature

## Overview

Este template ejecuta el ciclo completo de desarrollo de una feature:
1. **sdd-explore**: Investigar código existente y requisitos
2. **sdd-apply**: Implementar código
3. **test-runner**: Ejecutar tests
4. **verification-before-completion**: Verificar implementación

## Configuración

```yaml
template: feature-pipeline
input: feature_request
description: Pipeline completo para implementar feature

flow:
  - id: 1
    agent: sdd-explore
    description: Investigar código existente y requisitos
    timeout: 60000
    
  - id: 2
    agent: sdd-apply
    description: Implementar código
    dependsOn: [1]
    timeout: 120000
    
  - id: 3
    agent: test-runner
    description: Ejecutar tests
    dependsOn: [2]
    timeout: 60000
    
  - id: 4
    agent: verification-before-completion
    description: Verificar implementación
    dependsOn: [3]
    timeout: 30000

output:
  format: markdown
  sections:
    - Código implementado
    - Tests ejecutados (resultados)
    - Verificación passing/failing
```

## Uso

```
Ejecutar feature-pipeline para "nombre-de-feature"
```

## Estado Compartido

El `sharedContext` mantiene:
- `artifacts.explore-results`: Hallazgos de la investigación
- `artifacts.implementation`: Código implementado
- `artifacts.test-results`: Resultados de tests (passing/failing)
- `artifacts.verification`: Resultado de la verificación

## Estados del Pipeline

| Paso | Estado | Descripción |
|------|--------|-------------|
| 1 | explore | Investigando código |
| 2 | implementing | Implementando código |
| 3 | testing | Ejecutando tests |
| 4 | verifying | Verificando implementación |

## Retry y Fallback

- Retry en cada paso (max 3 intentos, backoff exponencial)
- Fallback: sdd-explore → sdd-patch (si la investigación es trivial)
- Si sdd-apply falla: revisar errores y reintentar

## Resultados Finales

```markdown
## Feature Pipeline Result

### Implementation
- [Status: Complete/Failed]
- Files modified: [...]

### Test Results
- Passing: N
- Failing: M
- Coverage: X%

### Verification
- Status: PASS/FAIL
- Issues: [...]
```