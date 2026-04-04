---
name: security-scanner
description: Escaneo OWASP 2025, detección de secretos y análisis de superficie de ataque
trigger: Cuando se solicita una auditoría de seguridad, antes de despliegues, o al detectar código sensible
license: MIT
metadata:
  version: 1.0.0
  author: Juarvis
  language: es-ES
---

# Security Scanner — OWASP 2025

## Objetivo

Identificar vulnerabilidades de seguridad basadas en OWASP Top 10 2025, detectar secretos expuestos, y mapear la superficie de ataque del proyecto.

## OWASP Top 10 2025 — Checklist

| # | Categoría | Qué buscar |
|---|-----------|------------|
| A01 | Pérdida de control de acceso | Endpoints sin verificación de permisos, IDOR, escalación de privilegios |
| A02 | Fallas criptográficas | Algoritmos débiles (MD5, SHA1 para passwords), claves hardcodeadas, HTTP sin TLS |
| A03 | Inyección | SQL injection, XSS, command injection, LDAP injection, template injection |
| A04 | Diseño inseguro | Falta de rate limiting, ausencia de threat modeling, trust boundaries rotas |
| A05 | Configuración de seguridad incorrecta | Debug activo en producción, headers de seguridad faltantes, CORS abierto |
| A06 | Componentes vulnerables | Dependencias con CVEs, versiones desactualizadas |
| A07 | Fallos de autenticación | Credenciales débiles, falta de MFA, sesiones sin expiración |
| A08 | Fallos de integridad de datos | Deserialización insegura, CI/CD sin verificación, supply chain |
| A09 | Fallos de logging y monitorización | Ausencia de logs de auditoría, logs con datos sensibles |
| A10 | Server-Side Request Forgery | SSRF en fetch/http interno, validación de URLs |

## Detección de secretos

Patrones a buscar:

```regex
# API keys
/api[_-]?key\s*[:=]\s*['"][^'"]+['"]

# Tokens
/token\s*[:=]\s*['"][^'"]+['"]

# Passwords
/password\s*[:=]\s*['"][^'"]+['"]

# AWS
/AKIA[0-9A-Z]{16}/

# Private keys
/-----BEGIN (RSA |EC )?PRIVATE KEY-----/

# URLs con credenciales
/https?:\/\/[^:]+:[^@]+@/
```

## Superficie de ataque

Mapear:
1. **Endpoints públicos**: rutas expuestas sin autenticación.
2. **Inputs de usuario**: formularios, APIs, websockets, archivos subidos.
3. **Integraciones externas**: APIs de terceros, webhooks, OAuth.
4. **Infraestructura**: bases de datos accesibles, colas de mensajes, almacenamiento.

## Formato del informe

```markdown
## Auditoría de Seguridad — [fecha]

### Resumen
- **Vulnerabilidades críticas**: N
- **Vulnerabilidades altas**: N
- **Secretos expuestos**: N
- **Superficie de ataque**: N endpoints públicos

### Hallazgos críticos

#### [A03] SQL Injection en `api/users.py:45`
- **Riesgo**: Acceso completo a base de datos
- **Prueba**: `GET /users?id=1 OR 1=1`
- **Corrección**: Usar parámetros preparados
  ```python
  # ANTES
  query = f"SELECT * FROM users WHERE id = {user_id}"
  # DESPUÉS
  query = "SELECT * FROM users WHERE id = %s"
  cursor.execute(query, (user_id,))
  ```

#### Secretos expuestos
| Archivo | Tipo | Línea | Acción |
|---------|------|-------|--------|
| `config.py` | API Key | 12 | Mover a variable de entorno |
| `.env` | DB Password | 3 | Añadir a `.gitignore` |

### Recomendaciones inmediatas
1. Rotar credenciales expuestas — **urgente**
2. Implementar prepared statements en endpoints vulnerables
3. Añadir headers de seguridad (CSP, HSTS, X-Frame-Options)
```

## Reglas

- **Nunca** incluir secretos reales en el informe, solo la ubicación.
- **Priorizar** por explotabilidad real, no solo por severidad teórica.
- **Incluir** prueba de concepto para vulnerabilidades críticas (sin explotar en producción).
- **Recomendar** corrección con código concreto, no solo descripción del problema.
