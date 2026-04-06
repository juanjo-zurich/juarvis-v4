#!/usr/bin/env bash
# Security Audit Hook — PreToolUse
# Detecta patrones de seguridad peligrosos antes de cada ejecución de herramienta.
# Adaptado de security_reminder_hook.py de Claude Code.
#
# Patrones detectados:
#   - Inyección de comandos (child_process.exec, os.system)
#   - XSS (innerHTML, dangerouslySetInnerHTML, document.write)
#   - eval(), new Function()
#   - pickle (deserialización insegura)
#   - Acceso a .env, credentials, secrets
#
# Exit codes:
#   0 = aprobado, continuar ejecución
#   2 = bloqueado, peligro de seguridad detectado

set -euo pipefail

# --- Configuración ---
LOG_DIR="${HOME}/.config/opencode/hooks/logs"
STATE_DIR="${HOME}/.config/opencode/hooks/state"
MAX_CONTENT_LENGTH=1048576  # 1MB máximo para inspeccionar

# --- Patrones de seguridad ---
# Formato: "NOMBRE_REGLA|PATRÓN_REGEX|MENSAJE"

SECURITY_PATTERNS=(
  "child_process_exec|child_process\.exec\b|⚠️ Aviso de seguridad: child_process.exec() puede provocar inyección de comandos. Usa execFile() con argumentos como array en su lugar."
  "child_process_exec_sync|execSync\b|⚠️ Aviso de seguridad: execSync() puede provocar inyección de comandos. Usa execFileSync() con argumentos como array."
  "os_system_injection|os\.system\b|⚠️ Aviso de seguridad: os.system() ejecuta comandos con shell. Nunca lo uses con argumentos que puedan ser controlados por el usuario."
  "eval_injection|eval\s*\(|⚠️ Aviso de seguridad: eval() ejecuta código arbitrario y es un riesgo grave de seguridad. Usa JSON.parse() para datos o patrones alternativos."
  "new_function_injection|new\s+Function\s*\(|⚠️ Aviso de seguridad: new Function() con cadenas dinámicas puede provocar inyección de código. Considera alternativas que no evalúen código arbitrario."
  "dangerously_set_inner_html|dangerouslySetInnerHTML|⚠️ Aviso de seguridad: dangerouslySetInnerHTML puede provocar XSS si se usa con contenido no confiable. Sanea el contenido con DOMPurify o usa alternativas seguras."
  "document_write_xss|document\.write\s*\(|⚠️ Aviso de seguridad: document.write() puede ser explotado para ataques XSS. Usa createElement() y appendChild() en su lugar."
  "innerHTML_xss|\.innerHTML\s*=|⚠️ Aviso de seguridad: Asignar innerHTML con contenido no confiable puede provocar XSS. Usa textContent para texto plano o sanea con DOMPurify."
  "pickle_deserialization|import\s+pickle|pickle\.loads?\s*\(|⚠️ Aviso de seguridad: pickle con contenido no confiable puede ejecutar código arbitrario. Usa JSON u otros formatos de serialización seguros."
  "env_file_access|\.env\b|credentials\.json|secrets/|⚠️ Aviso de seguridad: Acceso detectado a archivos sensibles (.env, credentials, secrets). Nunca incluyas secretos en el código fuente."
  "github_actions_injection|\$\{\{\s*github\.event\.|⚠️ Aviso de seguridad: Inyección potencial en GitHub Actions. Usa variables de entorno en lugar de interpolación directa en run:."
  "curl_pipe_sh|curl.*\|\s*sh|wget.*\|\s*sh|⚠️ Aviso de seguridad: Ejecución remota de scripts (curl|sh) es un riesgo grave. Descarga, verifica y luego ejecuta."
  "sql_injection|SELECT.*\+.*\$|INSERT.*\+.*\$|UPDATE.*\+.*\$|⚠️ Aviso de seguridad: Posible inyección SQL detectada. Usa consultas parametrizadas o un ORM."
)

# --- Funciones ---

check_security_patterns() {
  local content="$1"
  local file_path="${2:-}"

  # Limitar tamaño del contenido a inspeccionar
  if [[ ${#content} -gt $MAX_CONTENT_LENGTH ]]; then
    content="${content:0:$MAX_CONTENT_LENGTH}"
  fi

  for pattern_entry in "${SECURITY_PATTERNS[@]}"; do
    IFS='|' read -r rule_name regex message <<< "$pattern_entry"
    if echo "$content" | grep -qE "$regex"; then
      echo "$message"
      echo "REGLA: $rule_name | ARCHIVO: $file_path"
      return 2
    fi
  done

  # Verificar acceso a archivos sensibles por ruta
  if [[ -n "$file_path" ]]; then
    if [[ "$file_path" =~ \.env$ ]] || [[ "$file_path" =~ \.env\. ]] || \
       [[ "$file_path" =~ credentials\.json$ ]] || [[ "$file_path" =~ /secrets/ ]]; then
      echo "⚠️ Aviso de seguridad: Intento de acceso a archivo sensible: $file_path"
      echo "REGLA: sensitive_file_access | ARCHIVO: $file_path"
      return 2
    fi
  fi

  return 0
}

extract_content() {
  local tool_name="$1"
  shift
  local args="$*"

  case "$tool_name" in
    Write|Edit)
      # Extraer el contenido del argumento (último parámetro grande)
      echo "$args"
      ;;
    *)
      echo "$args"
      ;;
  esac
}

# --- Inicio del hook ---

# Crear directorios necesarios
mkdir -p "$LOG_DIR" "$STATE_DIR" 2>/dev/null || true

LOG_FILE="${LOG_DIR}/security-audit-$(date +%Y-%m-%d).log"
TIMESTAMP="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# Recibir datos del hook system (stdin JSON o argumentos)
TOOL_NAME="${1:-unknown}"
TOOL_ARGS="${2:-}"

# Si no hay argumentos, intentar leer de stdin
if [[ "$TOOL_NAME" == "unknown" ]] || [[ -z "$TOOL_ARGS" ]]; then
  if [[ -t 0 ]]; then
    # Sin stdin y sin argumentos — permitir
    echo "[${TIMESTAMP}] SECURITY-AUDIT | Tool: ${TOOL_NAME} | Status: skip (no input)" >> "$LOG_FILE"
    exit 0
  fi
  # Leer stdin
  INPUT_DATA=$(cat)
  TOOL_NAME=$(echo "$INPUT_DATA" | grep -oP '"tool_name"\s*:\s*"\K[^"]+' 2>/dev/null || echo "unknown")
  TOOL_ARGS=$(echo "$INPUT_DATA" | grep -oP '"content"\s*:\s*"\K[^"]+' 2>/dev/null || echo "")
  FILE_PATH=$(echo "$INPUT_DATA" | grep -oP '"file_path"\s*:\s*"\K[^"]+' 2>/dev/null || echo "")
else
  FILE_PATH="${3:-}"
fi

# Solo auditar herramientas de escritura
case "$TOOL_NAME" in
  Write|Edit|MultiEdit|Bash)
    ;;
  *)
    echo "[${TIMESTAMP}] SECURITY-AUDIT | Tool: ${TOOL_NAME} | Status: skip (non-write tool)" >> "$LOG_FILE"
    exit 0
    ;;
esac

# Extraer contenido para auditar
CONTENT=$(extract_content "$TOOL_NAME" "$TOOL_ARGS")

# Comprobar patrones de seguridad
WARNING=$(check_security_patterns "$CONTENT" "$FILE_PATH") || {
  # Patrón de seguridad detectado — bloquear ejecución
  echo "[${TIMESTAMP}] SECURITY-AUDIT | Tool: ${TOOL_NAME} | File: ${FILE_PATH} | Status: BLOCKED" >> "$LOG_FILE"
  echo "[${TIMESTAMP}] SECURITY-AUDIT | Warning: ${WARNING}" >> "$LOG_FILE"
  echo "$WARNING" >&2
  exit 2
}

# Sin problemas — permitir ejecución
echo "[${TIMESTAMP}] SECURITY-AUDIT | Tool: ${TOOL_NAME} | File: ${FILE_PATH} | Status: approved" >> "$LOG_FILE"
exit 0
