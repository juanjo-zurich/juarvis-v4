---
name: workflow-templates/full-audit
description: Auditoría completa de seguridad y código. Flow: security-scanner → test-driven → code-reviewer
allowed-tools: Read, Glob, Grep
---

# Template: full-audit

> Auditoría completa de seguridad y código

## Overview

Este template ejecuta un flujo secuencial de auditoría:
1. **security-scanner**: Análisis de vulnerabilidades
2. **test-driven**: Generación de tests de seguridad
3. **code-reviewer**: Review de código y patrones

## Configuración

```yaml
template: full-audit
input: ruta_del_proyecto
description: Auditoría completa de seguridad y código

flow:
  - id: 1
    agent: security-scanner
    description: Análisis de vulnerabilidades en dependencias
    timeout: 120000
    
  - id: 2
    agent: test-driven
    description: Generación de tests de seguridad
    dependsOn: [1]
    timeout: 60000
    
  - id: 3
    agent: code-reviewer
    description: Review de código y patrones de seguridad
    dependsOn: [2]
    timeout: 90000

output:
  format: markdown
  sections:
    - Vulnerabilidades encontradas
    - Tests de seguridad generados
    - Problemas de código identificados
    - Recomendaciones
```

## Uso

```
Ejecutar full-audit en /path/to/proyecto
```

## Estado Compartido

El `sharedContext` mantiene:
- `artifacts.security-scan`: Resultados del scanner
- `artifacts.test-files`: Tests generados
- `artifacts.review-results`: Resultados del review
- `state.currentStep`: Paso actual (1-3)
- `state.completedSteps`: Pasos completados

##Errores y Fallback

- Si security-scanner falla → intentar dependency-audit
- Si test-driven falla → intentar test-engineer
- Si code-reviewer falla → intentar refactor-assistant