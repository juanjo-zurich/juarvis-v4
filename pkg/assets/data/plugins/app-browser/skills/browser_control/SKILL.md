---
name: browser_control
description: >
  Automatización de navegador para testing y web scraping.
  Trigger: /browser, "abre navegador", "toma screenshot"
license: MIT
metadata:
  author: juarvis-org
  version: "1.0"
---

# Browser Automation Skill

## Propósito
Conecta Juarvis con navegador para:
- Automatización de testing
- Web scraping
- Screenshots y visual regression
- Fill forms

## Herramientas MCP
| Herramienta | Descripción |
|-------------|-----------|
| `browser_navigate` | Navegar a URL |
| `browser_screenshot` | Tomar screenshot |
| `browser_click` | Click en elemento |
| `browser_fill` | Rellenar formulario |
| `browser_evaluate` | Ejecutar JS |

## Uso
```
1. URL objetivo
2. browser_navigate
3. browser_screenshot
4. Analizar resultado
```

## Workflow: Visual Testing
```
1. browser_navigate a página
2. browser_screenshot
3. Comparar con baseline
4. Reportar diferencias
```

## Config
# No requiere API key para local