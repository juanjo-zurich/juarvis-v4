# GitHub Branch Protection

Para proteger la rama `main` y asegurar que ningún merge se realice sin tests pasando:

## Configuración

1. Ve a **Settings** > **Branches** > **Branch protection rules**
2. Click **Add rule**
3. Branch name pattern: `main`
4. Marca **Require status checks to pass before merging**
5. Busca y selecciona los siguientes status checks:
   - `unit-tests`
   - `integration-tests`
   - `regression-tests`
   - `e2e-tests`
   - `verify`
6. Marca **Require branches to be up to date before merging**
7. Click **Create**

## Resultado

- Ningún PR puede hacer merge a `main` sin que los 5 jobs del CI pasen
- Ni humanos ni agentes pueden saltarse esta protección
- Los commits directos a `main` están bloqueados (requieren PR)
