---
name: use_figma
description: >
  Integración con Figma para diseño, componentes y Code Connect.
  Trigger: /figma, "abre Figma", "diseña en Figma", "convierte Figma a código"
license: MIT
metadata:
  author: juarvis-org
  version: "1.0"
---

# Figma Integration Skill

## Propósito
Conecta Juarvis con Figma para permitir:
- Navegar y explorar archivos Figma
- Extraer componentes y estilos
- Generar código a partir de diseños
- Sincronizar cambios de diseño

## Herramientas MCP Disponibles

| Herramienta | Descripción |
|------------|-------------|
| `get_file` | Obtener un archivo Figma por URL |
| `get_components` | Listar todos los componentes |
| `get_styles` | Extraer estilos (colores, tipografía) |
| `get_image` | Exportar como imagen |
| `get_code` | Generar código (React, Flutter, etc.) |

## Uso

### Explorando Figma
```
1. Pedir al usuario la URL del archivo Figma
2. Usar get_file para cargar el archivo
3. Navegar por páginas y frames
4. Identificar componentes relevantes
```

### Extrayendo Componentes
```
1. get_components para listar componentes
2. Seleccionar componentes objetivo
3. get_styles para extraer tokens de diseño
4. Generar código con get_code
```

### Código Generado
- **React**: Componentes funcionals con styled-components o CSS modules
- **Flutter**: Widgets con Material Design
- **HTML/CSS**: HTML semántico + CSS

## Workflows

### Workflow 1: Design Review
```
1. Cargar archivo Figma
2. Identificar componentes главную
3. Extraer estilos relevantes
4. Documentar en SPEC.md
```

### Workflow 2: Prototipo a Código
```
1. Cargar archivo Figma
2. Seleccionar frame principal
3. get_code para convertir
4. Implementar en proyecto
```

### Workflow 3: Sincronizar Cambios
```
1. Comparar diseño actual con código
2. Identificar diferencias
3. Actualizar diseño o código según necesidad
```

## Configuración

### Variables de Entorno
```
FIGMA_ACCESS_TOKEN=...
FIGMA_FILE_KEY=...
```

### Permisos Requeridos
- `read`: Archivos y componentes
- `write`: Crear comentarios
- `admin`: Archivos propios

## Mejores Prácticas

1. **Siempre pedir la URL** del archivo Figma al usuario
2. **Documentar decisiones de diseño** en SPEC.md
3. **Mantener sincronía** entre Figma y código
4. **Usar componentes** en lugar de duplicar código
5. **Versionar** cambios de diseño

## Notas de Seguridad

- No exponer tokens de Figma en código
- Usar variables de entorno
- No hacer commit de tokens
- Rotar tokens periódicamente