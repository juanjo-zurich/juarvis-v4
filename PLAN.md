# Plan de Mejoras — Juarvis V4

> Generado: 2026-04-04
> Estado: Propuesto

---

## P0 — Criticos (bloqueantes para cualquier release)

### 1. Tests unitarios minimos

**Problema:** Cero cobertura de tests. Cualquier refactor es un salto al vacio.

**Archivos afectados:** `pkg/validate/validate_test.go`, `pkg/snapshot/snapshot_test.go`, `pkg/loader/loader_test.go`, `pkg/pm/pm_test.go`, `pkg/root/root_test.go`

**Acciones:**
- Tests de happy path para cada paquete de `pkg/`
- Tests de error cases (directorio inexistente, JSON corrupto, git no disponible)
- Tests para `pkg/root/root.go` con diferentes escenarios de JUARVIS_ROOT

**Criterios de exito:**
- [ ] `go test ./...` pasa con al menos 1 test por paquete
- [ ] Coverage minimo del 40% en `pkg/`

**Riesgo:** Mockear `exec.Command` y `os.*` requiere interfaces o `t.Setenv`. Usar `t.Setenv` para JUARVIS_ROOT y subdirectorios temporales con `t.TempDir()`.

---

### 2. Propagar errores silenciosos

**Problema:** Multiples operaciones de I/O ignoran errores, dejando el ecosistema en estado inconsistente sin notificar al usuario.

**Archivos afectados:**
- `pkg/loader/loader.go:56` — `json.Unmarshal` sin check
- `pkg/loader/loader.go:69` — `os.Symlink` sin check
- `pkg/setup/setup.go:85-106` — `copyFile` sin verificar retorno
- `pkg/root/root.go:13` — `os.Getwd()` error ignorado
- `cmd/loader.go:47-62` — `os.MkdirAll` y `os.WriteFile` sin check

**Acciones:**
- Anadir check de error a cada `os.MkdirAll`, `os.WriteFile`, `os.Symlink`, `json.Unmarshal`
- Propagar errores con `fmt.Errorf` descriptivo
- En `setup.go`, acumular errores no criticos y mostrar resumen al final

**Criterios de exito:**
- [ ] Cero llamadas a funciones que retornan error sin verificar
- [ ] `go vet ./...` sin warnings

---

### 3. Fix git stash leak en snapshots

**Problema:** `CreateSnapshot` hace `stash push` + `stash apply` pero nunca `stash drop`. Cada snapshot deja un stash permanente que se acumula indefinidamente.

**Archivo afectado:** `pkg/snapshot/snapshot.go:27`

**Acciones:**
- Opcion A: Cambiar `stash apply` por `stash pop` (aplica y elimina en un paso)
- Opcion B: Mantener `apply` pero anadir comando `snapshot prune` que elimine stashes antiguos de juarvis

**Criterios de exito:**
- [ ] Tras 10 `snapshot create`, `git stash list` no tiene 10 entradas de juarvis
- [ ] El working tree queda intacto tras el snapshot

**Riesgo:** `stash pop` puede fallar con conflictos igual que `apply`. La diferencia es que si hay conflicto, `pop` no elimina el stash (comportamiento seguro).

---

## P1 — Importantes (deben ir antes de la primera release publica)

### 4. `GetRoot()` con validacion estricta

**Problema:** Si no encuentra `marketplace.json` subiendo el arbol, devuelve el directorio padre sin verificar. Todo fallara despues silenciosamente.

**Archivo afectado:** `pkg/root/root.go`

**Acciones:**
- Cambiar firma a `func GetRoot() (string, error)`
- Si no encuentra `marketplace.json` tras subir 3 niveles maximo, devolver error: `"no se encontro un ecosistema Juarvis. Usa --root o ejecuta 'juarvis init'"`
- Actualizar todos los callers para manejar el error

**Criterios de exito:**
- [ ] `juarvis check` desde `/tmp` muestra error claro, no fallos silenciosos
- [ ] Todos los comandos que usan `GetRoot()` manejan el error

---

### 5. Comando `juarvis init`

**Problema:** No hay forma de crear un ecosistema desde cero con la CLI.

**Archivos afectados:** Nuevo `cmd/init.go` + `pkg/init/init.go`

**Acciones:**
- `juarvis init [path]` crea estructura base:
  - `marketplace.json` (plantilla vacia o con plugins por defecto)
  - `plugins/` (con al menos `juarvis-core`)
  - `.juar/` (directorio vacio)
  - `AGENTS.md` (desde assets embebidos)
  - `permissions.yaml` (desde assets embebidos)
  - `opencode.json` (desde assets embebidos)
- Usar assets embebidos de `pkg/assets/data/` como fuentes
- Si `path` no se especifica, usar cwd

**Criterios de exito:**
- [ ] `juarvis init` desde directorio vacio crea estructura completa
- [ ] `juarvis init` en directorio con contenido no sobreescribe archivos existentes
- [ ] `juarvis check` pasa inmediatamente despues de `juarvis init`

---

### 6. Flag `--version`

**Problema:** Imposible saber que version de Juarvis se esta ejecutando.

**Archivos afectados:** `cmd/root.go`, `main.go`

**Acciones:**
- Anadir variable `var version = "dev"` en `cmd/root.go`
- Inyectar en build time con `ldflags`: `-X juarvis/cmd.version=1.0.0`
- `juarvis --version` muestra version + commit hash + fecha de build

**Criterios de exito:**
- [ ] `juarvis --version` muestra `juarvis version 1.0.0 (commit: abc123, built: 2026-04-04)`
- [ ] Build sin ldflags muestra `juarvis version dev`

---

### 7. Fix `skill-create` con validacion de directorios

**Problema:** `skill-create` no verifica que `plugins/` existe ni verifica errores de `MkdirAll`/`WriteFile`.

**Archivo afectado:** `cmd/loader.go:34-71`

**Acciones:**
- Verificar que `plugins/` existe, crearlo si no
- Verificar retorno de cada `os.MkdirAll` y `os.WriteFile`
- Mostrar error claro si falla la creacion

**Criterios de exito:**
- [ ] `juarvis skill-create test` desde ecosistema vacio funciona
- [ ] Errores de I/O se muestran al usuario

---

## P2 — Mejoras funcionales y de DX

### 8. Makefile con targets estandar

**Problema:** No hay instruccion de como compilar, testear o linkear.

**Archivo afectado:** Nuevo `Makefile`

**Acciones:**
```makefile
VERSION ?= dev
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -X juarvis/cmd.version=$(VERSION) -X juarvis/cmd.commit=$(COMMIT)

.PHONY: build test lint install clean

build:
	go build -ldflags "$(LDFLAGS)" -o juarvis .

test:
	go test -v -cover ./...

lint:
	go vet ./...
	golangci-lint run ./...

install: build
	cp juarvis /usr/local/bin/

clean:
	rm -f juarvis
```

**Criterios de exito:**
- [ ] `make build` genera binario `juarvis`
- [ ] `make test` ejecuta tests con coverage
- [ ] `make lint` pasa sin errores

---

### 9. Comando `pm install`

**Problema:** El package manager no tiene forma de instalar plugins desde el marketplace.

**Archivos afectados:** `cmd/pm.go`, `pkg/pm/pm.go`

**Acciones:**
- `juarvis pm install <plugin>` busca plugin en `marketplace.json`
- Clona/descarga desde `source` a `plugins/<nombre>`
- Crea estructura `.juarvis-plugin/plugin.json` si no existe
- Ejecuta `loader.RunLoader()` para indexar
- Soportar fuentes: rutas relativas (`./plugins/...`), URLs git (`https://...`)

**Criterios de exito:**
- [ ] `juarvis pm install juarvis-core` instala el plugin
- [ ] Plugin aparece en `juarvis pm list` como instalado
- [ ] Error claro si el plugin no existe en el marketplace

---

### 10. Output `--json` para consumo programatico

**Problema:** Toda la CLI usa `fmt.Println` con emojis. Los agentes IA no pueden parsear la salida facilmente.

**Archivos afectados:** Todos los comandos en `cmd/` y `pkg/`

**Acciones:**
- Anadir flag global `--json` en `root.go`
- Crear `pkg/output/output.go` con funciones `PrintJSON`, `PrintTable`, `PrintError`
- Cuando `--json` esta activo, toda salida es JSON estructurado
- Comandos afectados prioritarios: `check`, `pm list`, `snapshot restore`

**Ejemplo de output JSON para `check`:**
```json
{
  "status": "ok",
  "checks": [
    {"name": "git", "status": "pass", "message": "Git detectado"},
    {"name": "marketplace", "status": "pass", "message": "Catalogo enlazado"}
  ]
}
```

**Criterios de exito:**
- [ ] `juarvis check --json` devuelve JSON valido
- [ ] `juarvis pm list --json` devuelve array de plugins
- [ ] Exit code 0 = ok, 1 = error (consistente con y sin `--json`)

---

### 11. Unificar formato de errores y output

**Problema:** Mensajes inconsistentes: algunos usan `fmt.Println("❌ Error:", err)`, otros `fmt.Printf`, otros propagan.

**Archivos afectados:** Todos los archivos en `cmd/` y `pkg/`

**Acciones:**
- Crear `pkg/output/output.go` como paquete central de presentacion
- Funciones: `Success(msg)`, `Error(msg)`, `Warning(msg)`, `Info(msg)`
- Usar en todos los comandos en vez de `fmt.Println` directo
- Mantener compatibilidad con `--json` (mismo paquete, diferente backend)

**Criterios de exito:**
- [ ] Todos los mensajes de exito usan mismo prefijo/formato
- [ ] Todos los errores incluyen contexto descriptivo
- [ ] Cero `fmt.Println` directo en comandos (solo via `pkg/output`)

---

## P3 — Robustez y mantenimiento

### 12. Atomicidad en `loader.go`

**Problema:** `os.RemoveAll(skillsDir)` + `os.MkdirAll(skillsDir)` no es atomico. Si se interrumpe, el directorio desaparece.

**Archivo afectado:** `pkg/loader/loader.go:25-26`

**Acciones:**
- Crear directorio temporal: `tmpDir, _ := os.MkdirTemp("", "juarvis-loader-*")`
- Construir symlinks en `tmpDir`
- `os.RemoveAll(skillsDir)` + `os.Rename(tmpDir, skillsDir)` (atomico en mismo filesystem)

**Criterios de exito:**
- [ ] Si el proceso muere durante el rebuild, `skills/` no desaparece
- [ ] El loader funciona correctamente tras la migracion a temporal

---

### 13. Sync automatico de assets embebidos

**Problema:** `pkg/assets/data/` contiene copias de `AGENTS.md`, `marketplace.json`, etc. Si se modifican en el repositorio y no se actualizan en `data/`, el binario distribuye versiones obsoletas.

**Archivos afectados:** `Makefile` (nuevo target), `pkg/assets/data/`

**Acciones:**
- Anadir target `make sync-assets` que copie desde la raiz del proyecto padre a `pkg/assets/data/`
- O anadir script `scripts/sync-assets.sh` ejecutado pre-build
- Documentar en README que los assets se generan, no se editan a mano

**Criterios de exito:**
- [ ] `make sync-assets` actualiza `pkg/assets/data/` desde fuentes
- [ ] Documentacion clara sobre el flujo de actualizacion

---

### 14. Autocompletado de shell

**Problema:** Cobra soporta autocompletado pero no esta configurado.

**Archivos afectados:** `cmd/completion.go` (nuevo)

**Acciones:**
- Anadir subcomando `juarvis completion [bash|zsh|fish|powershell]`
- Usar `rootCmd.GenBashCompletion(os.Stdout)` y equivalentes

**Criterios de exito:**
- [ ] `juarvis completion zsh` genera script valido
- [ ] Tras `source <(juarvis completion zsh)`, tab-complete funciona

---

### 15. Comando `snapshot prune`

**Problema:** Los stashes de juarvis se acumulan. No hay forma de limpiarlos.

**Archivos afectados:** `cmd/snapshot.go`, `pkg/snapshot/snapshot.go`

**Acciones:**
- `juarvis snapshot prune [--older-than 7d]` elimina stashes de juarvis antiguos
- `juarvis snapshot prune --all` elimina todos los stashes de juarvis
- Mostrar cuantos stashes se eliminaron

**Criterios de exito:**
- [ ] `snapshot prune --all` elimina solo stashes con prefijo `juarvis-snapshot|`
- [ ] Stashes del usuario (no juarvis) no se tocan
- [ ] Mensaje informativo: "3 snapshots eliminados"

---

## Dependencias entre tareas

```
P0-2 (errores) ──► P1-4 (GetRoot) ──► P1-5 (init)
P0-3 (stash) ──► P3-15 (prune)
P0-1 (tests) ──► (todo lo demas, recomendado hacer primero)
P2-11 (output) ──► P2-10 (--json)
P2-8 (Makefile) ──► P3-13 (sync-assets)
P2-9 (pm install) ──► (independiente)
P1-6 (--version) ──► (independiente, pero necesita Makefile para ldflags)
```

## Orden de ejecucion recomendado

1. **P0-2** — Propagar errores (afecta a todo, bajo riesgo)
2. **P0-3** — Fix stash leak (1 linea, alto impacto)
3. **P0-1** — Tests unitarios (bloquea confianza para lo demas)
4. **P1-4** — GetRoot con validacion
5. **P1-5** — Comando init
6. **P1-6** — Flag --version
7. **P1-7** — Fix skill-create
8. **P2-8** — Makefile
9. **P2-11** — Unificar output
10. **P2-10** -- Output --json
11. **P2-9** — pm install
12. **P3-12** — Atomicidad loader
13. **P3-13** — Sync assets
14. **P3-14** — Autocompletado
15. **P3-15** — Snapshot prune
