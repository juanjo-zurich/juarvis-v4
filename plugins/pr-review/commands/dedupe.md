---
description: Encontrar issues duplicados de GitHub
argument-hint: <número de issue>
---

Encontrar hasta 3 issues probablemente duplicados para un issue de GitHub dado.

Para hacer esto, sigue estos pasos con precisión:

1. Usar un agente para comprobar si el issue de GitHub (a) está cerrado, (b) no necesita ser dedupeado (ej. porque es feedback de producto amplio sin una solución específica, o feedback positivo), o (c) ya tiene un comentario de duplicados que hiciste anteriormente. Si es así, no proceder.
2. Usar un agente para ver un issue de GitHub, y pedir al agente que devuelva un resumen del issue
3. Luego, lanzar 5 agentes en paralelo para buscar en GitHub duplicados de este issue, usando keywords diversas y enfoques de búsqueda diversos, usando el resumen del #1
4. Luego, alimentar los resultados de #1 y #2 en otro agente, para que pueda filtrar falsos positivos, que probablemente no son realmente duplicados del issue original. Si no quedan duplicados, no proceder.
5. Finalmente, usar el script de comentario para postear duplicados:
   ```
   ./scripts/comment-on-duplicates.sh --base-issue <número-de-issue> --potential-duplicates <dup1> <dup2> <dup3>
   ```

Notas (asegúrate de decir esto a tus agentes también):

- Usar `./scripts/gh.sh` para interactuar con GitHub, en lugar de web fetch o `gh` directo. Ejemplos:
  - `./scripts/gh.sh issue view 123` — ver un issue
  - `./scripts/gh.sh issue view 123 --comments` — ver con comentarios
  - `./scripts/gh.sh issue list --state open --limit 20` — listar issues
  - `./scripts/gh.sh search issues "query" --limit 10` — buscar issues
- No usar otras herramientas, más allá de `./scripts/gh.sh` y el script de comentario (ej. no usar otros servidores MCP, edición de archivos, etc.)
- Hacer una lista de tareas primero

(Fin del archivo)