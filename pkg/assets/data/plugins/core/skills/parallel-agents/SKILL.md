---
name: parallel-agents
description: Multi-agent orchestration patterns. Use when multiple independent tasks can run with different domain expertise or when comprehensive analysis requires multiple perspectives.
allowed-tools: Read, Glob, Grep
---

# Native Parallel Agents

> Orchestration through Antigravity's built-in Agent Tool

## Overview

This skill enables coordinating multiple specialized agents through Antigravity's native agent system. Unlike external scripts, this approach keeps all orchestration within Antigravity's control.

## When to Use Orchestration

✅ **Good for:**
- Complex tasks requiring multiple expertise domains
- Code analysis from security, performance, and quality perspectives
- Comprehensive reviews (architecture + security + testing)
- Feature implementation needing backend + frontend + database work

❌ **Not for:**
- Simple, single-domain tasks
- Quick fixes or small changes
- Tasks where one agent suffices

---

## Native Agent Invocation

### Single Agent
```
Use the security-auditor agent to review authentication
```

### Sequential Chain
```
First, use the explorer-agent to discover project structure.
Then, use the backend-specialist to review API endpoints.
Finally, use the test-engineer to identify test gaps.
```

### With Context Passing
```
Use the frontend-specialist to analyze React components.
Based on those findings, have the test-engineer generate component tests.
```

### Resume Previous Work
```
Resume agent [agentId] and continue with additional requirements.
```

---

## Orchestration Patterns

### Pattern 1: Comprehensive Analysis
```
Agents: explorer-agent → [domain-agents] → synthesis

1. explorer-agent: Map codebase structure
2. security-auditor: Security posture
3. backend-specialist: API quality
4. frontend-specialist: UI/UX patterns
5. test-engineer: Test coverage
6. Synthesize all findings
```

### Pattern 2: Feature Review
```
Agents: affected-domain-agents → test-engineer

1. Identify affected domains (backend? frontend? both?)
2. Invoke relevant domain agents
3. test-engineer verifies changes
4. Synthesize recommendations
```

### Pattern 3: Security Audit
```
Agents: security-auditor → penetration-tester → synthesis

1. security-auditor: Configuration and code review
2. penetration-tester: Active vulnerability testing
3. Synthesize with prioritized remediation
```

---

## Available Agents

| Agent | Expertise | Trigger Phrases |
|-------|-----------|-----------------|
| `orchestrator` | Coordination | "comprehensive", "multi-perspective" |
| `security-auditor` | Security | "security", "auth", "vulnerabilities" |
| `penetration-tester` | Security Testing | "pentest", "red team", "exploit" |
| `backend-specialist` | Backend | "API", "server", "Node.js", "Express" |
| `frontend-specialist` | Frontend | "React", "UI", "components", "Next.js" |
| `test-engineer` | Testing | "tests", "coverage", "TDD" |
| `devops-engineer` | DevOps | "deploy", "CI/CD", "infrastructure" |
| `database-architect` | Database | "schema", "Prisma", "migrations" |
| `mobile-developer` | Mobile | "React Native", "Flutter", "mobile" |
| `api-designer` | API Design | "REST", "GraphQL", "OpenAPI" |
| `debugger` | Debugging | "bug", "error", "not working" |
| `explorer-agent` | Discovery | "explore", "map", "structure" |
| `documentation-writer` | Documentation | "write docs", "create README", "generate API docs" |
| `performance-optimizer` | Performance | "slow", "optimize", "profiling" |
| `project-planner` | Planning | "plan", "roadmap", "milestones" |
| `seo-specialist` | SEO | "SEO", "meta tags", "search ranking" |
| `game-developer` | Game Development | "game", "Unity", "Godot", "Phaser" |

---

## Antigravity Built-in Agents

These work alongside custom agents:

| Agent | Model | Purpose |
|-------|-------|---------|
| **Explore** | Haiku | Fast read-only codebase search |
| **Plan** | Sonnet | Research during plan mode |
| **General-purpose** | Sonnet | Complex multi-step modifications |

Use **Explore** for quick searches, **custom agents** for domain expertise.

---

## Synthesis Protocol

After all agents complete, synthesize:

```markdown
## Orchestration Synthesis

### Task Summary
[What was accomplished]

### Agent Contributions
| Agent | Finding |
|-------|---------|
| security-auditor | Found X |
| backend-specialist | Identified Y |

### Consolidated Recommendations
1. **Critical**: [Issue from Agent A]
2. **Important**: [Issue from Agent B]
3. **Nice-to-have**: [Enhancement from Agent C]

### Action Items
- [ ] Fix critical security issue
- [ ] Refactor API endpoint
- [ ] Add missing tests
```

---

## Best Practices

1. **Available agents** - 17 specialized agents can be orchestrated
2. **Logical order** - Discovery → Analysis → Implementation → Testing
3. **Share context** - Pass relevant findings to subsequent agents
4. **Single synthesis** - One unified report, not separate outputs
5. **Verify changes** - Always include test-engineer for code modifications

---

## Key Benefits

- ✅ **Single session** - All agents share context
- ✅ **AI-controlled** - Claude orchestrates autonomously
- ✅ **Native integration** - Works with built-in Explore, Plan agents
- ✅ **Resume support** - Can continue previous agent work
- ✅ **Context passing** - Findings flow between agents

---

## Retry Configuration

### Configuración

```yaml
retry:
  maxRetries: 3          # 1-3 reintentos (configurable)
  backoff: exponential   # exponential (2^n segundos) o linear
  retryCondition:
    timeout: true       # Reintentar timeouts
    rateLimit: true      # Reintentar rate limits (429)
    transient: true      # Errores transitorios
  fatalErrors:
    - "permission denied"
    - "not found"
    - "authentication failed"
  timeoutPerAttempt: 30000  # ms
```

### Algoritmo de Retry

```
1. Ejecutar tarea
2. Si error:
   a. Verificar si es fatal → fallar inmediatamente
   b. Si es reintentable:
      - Calcular delay (backoff)
      - Esperar delay
      - Reintentar (hasta maxRetries)
3. Si todos los reintentos fallan → marcar como failed
```

### Backoff Calculation

| Tipo | Fórmula | Ejemplo (initial: 1s) |
|------|---------|----------------------|
| exponential | `initial * 2^n` | 1s, 2s, 4s, 8s... |
| linear | `initial + (n * increment)` | 1s, 2s, 3s, 4s... |

### Estados del Retry

- `pending` - No iniciado
- `running` - En ejecución
- `retrying` - Reintentando (n/m intentos)
- `success` - Completado exitosamente
- `failed` - Falló después de todos los reintentos
- `skipped` - Saltado (no aplica retry)

---

## Fallback Configuration

### Configuración

```yaml
fallback:
  enabled: true
  fallbackAgent: test-driven  # Agente alternativo
  fallbackCondition:
    - agentNotAvailable   # Agente no existe
    - agentTimeout        # Timeout del agente
    - agentError          # Error irrecuperable
```

### Matriz de Fallbacks

| Agente Primary | Fallback | Condición |
|----------------|----------|-----------|
| security-scanner | dependency-audit | Scanner no disponible |
| test-engineer | test-driven | Test engineer falla |
| code-reviewer | refactor-assistant | Reviewer no responde |
| sdd-explore | sdd-patch | Explorar trivial |

### Flujo de Fallback

```
1. Ejecutar agente primario
2. Si falla:
   a. Verificar condición de fallback
   b. Si aplica: ejecutar agente fallback
   c. Si fallback también falla: propagar error
3. Registrar qué agente ejecutó (primary/fallback)
```

---

## Shared Context

### Estructura

```yaml
sharedContext:
  id: "uuid-unico"
  template: "full-audit"
  createdAt: "ISO8601"
  
  artifacts:
    security-scan: {}
    test-files: []
    review-results: {}
    
  state:
    currentStep: 2
    completedSteps: [1]
    results: {}
    
  errors:
    - step: 1
      error: "timeout"
      retryCount: 2
      
  metadata:
    inputPath: "/path/to/project"
    userId: "user-123"
```

### Métodos del Context

| Método | Descripción |
|--------|-------------|
| `set(key, value)` | Guardar valor |
| `get(key)` | Obtener valor |
| `appendError(error)` | Añadir error |
| `setArtifact(name, data)` | Guardar artefacto |
| `getArtifact(name)` | Obtener artefacto |
| `nextStep()` | Avanzar al siguiente paso |
| `markComplete(stepId)` | Marcar paso como completado |
| `getState()` | Obtener estado completo |

### Persistencia

- El context se persiste en `.juar/contexts/{id}.yaml`
- Se limpia automáticamente después de completarse
- Timeout de limpieza: 24 horas

---

## Workflow Templates

Para flujos predefinidos, ver:
- `plugins/core/skills/workflow-templates/full-audit.md`
- `plugins/core/skills/workflow-templates/parallel-review.md`
- `plugins/core/skills/workflow-templates/feature-pipeline.md`
