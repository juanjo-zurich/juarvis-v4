---
name: ci-cd-pipeline
description: >
  Usar cuando se configura integración o entrega continua: GitHub Actions, GitLab CI,
  estrategias de despliegue, secrets, matrix builds, cacheo, artefactos, rollback.
  Activador: "GitHub Actions", "CI/CD", "pipeline", "workflow", "despliegue automático",
  "GitLab CI", "secrets", "Docker build", "test automático", "deploy", ".github/workflows".
version: "1.0"
---

# CI/CD — Pipelines

## GitHub Actions — Estructura Base

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true  # cancela runs anteriores del mismo PR

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: npm          # cachea node_modules automáticamente

      - run: npm ci           # ci en lugar de install (más reproducible)
      - run: npm run typecheck
      - run: npm test -- --coverage
      - run: npm run lint

      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: coverage
          path: coverage/
```

## Pipeline Completo (Node.js)

```yaml
# .github/workflows/main.yml
name: CI/CD

on:
  push:
    branches: [main]
  pull_request:

jobs:
  # ── 1. Quality gates ─────────────────────────────────────────────────────────
  quality:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: 20, cache: npm }
      - run: npm ci
      - run: npm run typecheck
      - run: npm run lint
      - run: npm test -- --coverage --reporter=json

      - name: Check coverage threshold
        run: |
          COVERAGE=$(cat coverage/coverage-summary.json | jq '.total.lines.pct')
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "Coverage $COVERAGE% below 80% threshold"
            exit 1
          fi

  # ── 2. Build ─────────────────────────────────────────────────────────────────
  build:
    needs: quality
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Build Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: false
          tags: myapp:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  # ── 3. Deploy staging (solo en PR) ───────────────────────────────────────────
  deploy-staging:
    needs: build
    if: github.event_name == 'pull_request'
    runs-on: ubuntu-latest
    environment: staging
    steps:
      - name: Deploy to staging
        run: echo "Deploy to staging..."
        env:
          DEPLOY_TOKEN: ${{ secrets.STAGING_DEPLOY_TOKEN }}

  # ── 4. Deploy producción (solo en main) ──────────────────────────────────────
  deploy-production:
    needs: build
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment:
      name: production
      url: https://myapp.com
    steps:
      - name: Deploy to production
        run: echo "Deploy to production..."
        env:
          DEPLOY_TOKEN: ${{ secrets.PROD_DEPLOY_TOKEN }}
```

## Matrix Builds — Tests Multi-versión

```yaml
strategy:
  matrix:
    node: [18, 20, 22]
    os: [ubuntu-latest, macos-latest]
  fail-fast: false   # continuar aunque falle una combinación

steps:
  - uses: actions/setup-node@v4
    with:
      node-version: ${{ matrix.node }}
```

## Secrets y Variables de Entorno

```yaml
# ✅ Secrets para valores sensibles (en Settings > Secrets)
env:
  DATABASE_URL: ${{ secrets.DATABASE_URL }}
  API_KEY: ${{ secrets.API_KEY }}

# ✅ Variables de entorno para configuración no sensible
env:
  NODE_ENV: production
  LOG_LEVEL: info

# ✅ Outputs entre jobs
jobs:
  prepare:
    outputs:
      version: ${{ steps.version.outputs.version }}
    steps:
      - id: version
        run: echo "version=$(cat package.json | jq -r .version)" >> $GITHUB_OUTPUT

  deploy:
    needs: prepare
    steps:
      - run: echo "Deploying version ${{ needs.prepare.outputs.version }}"
```

## Seguridad en Actions

```yaml
# ✅ Pinear actions a commit hash, no a tag
- uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683  # v4.2.2

# ✅ Permisos mínimos
permissions:
  contents: read
  pull-requests: write  # solo si necesario

# ✅ Evitar inyección en run: con variables de entorno
- name: Safe echo
  env:
    TITLE: ${{ github.event.pull_request.title }}
  run: echo "$TITLE"   # ✅ safe via env var

# ❌ Nunca directamente:
- run: echo "${{ github.event.pull_request.title }}"  # inyectable
```

## Cacheo Efectivo

```yaml
# Node.js — automático con setup-node cache: npm
- uses: actions/setup-node@v4
  with:
    node-version: 20
    cache: npm

# Python — automático con setup-python cache: pip
- uses: actions/setup-python@v5
  with:
    python-version: 3.11
    cache: pip

# Cache personalizado (Docker layers, compilaciones)
- uses: actions/cache@v4
  with:
    path: ~/.cache/custom
    key: ${{ runner.os }}-${{ hashFiles('**/lockfile') }}
    restore-keys: ${{ runner.os }}-
```

## Estrategias de Despliegue

| Estrategia | Cuándo | Cómo |
|------------|--------|------|
| **Blue/Green** | Zero-downtime crítico | Dos entornos, switch de tráfico |
| **Rolling** | Default en Kubernetes | Reemplaza pods uno a uno |
| **Canary** | Validar en producción real | 5% → 25% → 100% |
| **Feature flags** | Desacoplar deploy de release | LaunchDarkly, Unleash |

## Rollback

```yaml
# ✅ Siempre tener un paso de rollback explícito
- name: Deploy
  id: deploy
  run: ./deploy.sh ${{ github.sha }}

- name: Rollback on failure
  if: failure() && steps.deploy.conclusion == 'failure'
  run: ./rollback.sh ${{ env.PREVIOUS_SHA }}
```
