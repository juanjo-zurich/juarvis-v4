---
description: "Cancelar bucle Ralph Wiggum activo"
allowed-tools: ["Bash(test *)", "Bash(rm *)", "Read"]
hide-from-slash-command-tool: "true"
---

# Cancelar Ralph

Para cancelar el bucle Ralph:

1. Comprobar si `.juarvis/ralph-loop.local.md` existe usando Bash: `test -f .juarvis/ralph-loop.local.md && echo "EXISTS" || echo "NOT_FOUND"`

2. **Si NOT_FOUND**: Decir "No se encontró un bucle Ralph activo."

3. **Si EXISTS**:
   - Leer `.juarvis/ralph-loop.local.md` para obtener el número de iteración actual del campo `iteration:`
   - Eliminar el archivo usando Bash: `rm .juarvis/ralph-loop.local.md`
   - Reportar: "Bucle Ralph cancelado (estaba en iteración N)" donde N es el valor de iteración
