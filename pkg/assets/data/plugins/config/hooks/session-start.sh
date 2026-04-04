#!/usr/bin/env bash
# Session Start Hook — Juarvis_V3
# Inyecta contexto pedagógico al inicio de cada sesión.
# Activado automáticamente si se habilita en hooks/config.yaml (enabled: true).
#
# Comportamiento:
#   - Carga el resumen de sesión anterior si existe (Engram degradado)
#   - Recuerda al agente las reglas de delegación más importantes
#   - Muestra el estado del proyecto si hay cambios pendientes en git

set -euo pipefail

LOG_DIR="${HOME}/.config/opencode/hooks/logs"
mkdir -p "$LOG_DIR"
TIMESTAMP="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
LOG_FILE="${LOG_DIR}/session-start-$(date +%Y-%m-%d).log"

echo "[${TIMESTAMP}] SESSION-START | Iniciando sesión Juarvis_V3" >> "$LOG_FILE"

# ── Estado del proyecto ────────────────────────────────────────────────────────
GIT_STATUS=""
if git rev-parse --git-dir &>/dev/null 2>&1; then
    BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "desconocida")
    CHANGED=$(git status --porcelain 2>/dev/null | wc -l | tr -d ' ')
    STAGED=$(git diff --cached --name-only 2>/dev/null | wc -l | tr -d ' ')

    if [[ "$CHANGED" -gt 0 ]] || [[ "$STAGED" -gt 0 ]]; then
        GIT_STATUS="📂 Proyecto: rama \`${BRANCH}\` — ${CHANGED} archivo(s) modificado(s), ${STAGED} en staging"
    else
        GIT_STATUS="✅ Proyecto: rama \`${BRANCH}\` — directorio limpio"
    fi
fi

# ── Ralph activo ───────────────────────────────────────────────────────────────
RALPH_MSG=""
if [[ -f ".opencode/ralph-loop.local.md" ]]; then
    ITERATION=$(sed -n '/^---$/,/^---$/{ /^---$/d; p; }' ".opencode/ralph-loop.local.md" \
        | grep '^iteration:' | sed 's/iteration: *//' || echo "?")
    RALPH_MSG="🔄 Ralph activo — iteración ${ITERATION}"
fi

# ── Mensaje de contexto ────────────────────────────────────────────────────────
MSG=""
[[ -n "$GIT_STATUS" ]] && MSG+="${GIT_STATUS}\n"
[[ -n "$RALPH_MSG" ]] && MSG+="${RALPH_MSG}\n"

if [[ -n "$MSG" ]]; then
    # Imprimir a stderr para que el agente lo vea como contexto del sistema
    printf "%b" "$MSG" >&2
fi

echo "[${TIMESTAMP}] SESSION-START | Completado" >> "$LOG_FILE"
exit 0
