# Juarvis V4 — Sistema Operativo para Agentes IA

Juarvis V4 es una herramienta de sistema que convierte cualquier carpeta en un entorno de desarrollo profesional gobernado por agentes de IA autónomos.

No es solo un script; es un **motor global** que permite que IAs (Claude, Cursor, OpenCode, etc.) sigan protocolos de ingeniería rigurosos, autogestionen sus herramientas y aprendan de sus errores sin que tú tengas que configurar nada.

---

## 🚀 Guía Rápida

### 1. Instalación (Solo una vez)
```bash
git clone https://github.com/juanjo-zurich/juarvis-v4.git
cd juarvis-v4
make install
```

### 2. Preparar tu Proyecto
```bash
juarvis up
```
*Esto hace todo: init + setup + watch + analiza tu proyecto*

### 3. Ver el Estado
```bash
juarvis dashboard    # Panel visual del ecosistema
juarvis mode        # Ver/cambiar nivel de autonomía (0-4)
```

---

## 🧠 Niveles de Autonomía

| Nivel | Nombre | Comportamiento |
|-------|--------|--------------|
| 0 | Vibe Puro | Ejecución directa, sin planes |
| 1 | Seguro | + snapshot automático antes de cambios grandes |
| 2 | Estructurado | + descomposición de tareas (default) |
| 3 | Semi-SDD | + spec antes de implementar |
| 4 | SDD Completo | Pipeline completo: explore→spec→design→verify |

Cambia el modo con: `juarvis mode 0` a `juarvis mode 4`

---

## ⚡ Comandos Esenciales

```bash
juarvis init           # Inicializar ecosistema
juarvis up            # Onboarding completo (init + setup + watch)
juarvis watch        # Vigilancia activa de archivos
juarvis verify       # Verificar que todo esté sano
juarvis snapshot     # Puntos de restauración
juarvis analyze      # Analizar codebase
juarvis memory      # Servidor MCP de memoria
juarvis mode [n]    # Cambiar nivel de autonomía
juarvis dashboard   # Panel visual del ecosistema
```

---

## 🗃️ Memoria Persistente

Después de `juarvis init`, el agente recuerda lo que aprende:

```
🎯 LO QUE TU AGENTE AHORA SABE DE TU PROYECTO
  📦 Stack: Go, React, PostgreSQL
  📁 Archivos: 143
  📋 Convenciones: conventional commits, ESLint

💡 ANTES: cada sesión → cold context
💡 AHORA: desde el minuto 0 → contexto persistente
```

Usa las herramientas MCP:
- `mem_save(title, content, type)` — guardar observación
- `mem_search(query)` — buscar en memoria
- `mem_context(limit)` — ver sesiones recientes

---

## 🔒 Seguridad

El **Watcher** vigila cada cambio. Si una IA intenta:
- Borrar archivos clave
- Ignorar tests
- Hacer operaciones destructivas

...el watcher lo detectará y te avisará o detendrá la operación.

---

## 🏗️ Arquitectura

- **Binario**: `juarvis` — CLI en Go, zero runtime deps
- ****.juar/*** en proyecto: memoria, skills, config
- **AGENTS.md**: protocolo activo (se actualiza con `juarvis mode`)
- **Servidor MCP**: memoria local (JSON + índice token en RAM)

---

## ✅ installación Verificada

```bash
go build .                    # Build binario
CGO_ENABLED=0 go build .     # Pure Go (sin C)
go test ./...                 # Tests
juarvis verify               # Verificar ecosistema
```

---

## 📦 Características

- CLI completa en Go (~30 comandos)
- 75+ skills embebidas en plugins
- Servidor MCP de memoria local
- Pipeline SDD completo
- Hooks de seguridad (Hookify)
- Ralph — bucle autónomo
- Dashboard visual (Bubble Tea)
- Tests offline deterministas

Más info: `juarvis --help`