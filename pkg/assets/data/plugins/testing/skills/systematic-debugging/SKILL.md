---
name: systematic-debugging
description: Metodología sistemática de depuración en 4 fases con análisis de causa raíz y verificación basada en evidencia. Úsalo para depurar problemas complejos.
allowed-tools: Read, Glob, Grep
---

# Depuración Sistemática

## Visión general
Esta habilidad proporciona un enfoque estructurado para la depuración que evita adivinaciones aleatorias y asegura que los problemas se comprendan correctamente antes de resolverlos.

## Proceso de Depuración en 4 Fases

### Fase 1: Reproducir
Antes de solucionar, reproduzca el problema de manera confiable.

```markdown
## Pasos para Reproducir
1. [Paso exacto para reproducir]
2. [Próximo paso]
3. [Resultado esperado vs resultado real]

## Tasa de Reproducción
- [ ] Siempre (100%)
- [ ] A menudo (50-90%)
- [ ] Algunas veces (10-50%)
- [ ] Rara vez (<10%)
```

### Fase 2: Aislar
Reduzca el origen del problema.

```markdown
## Preguntas de Aislamiento
- ¿Cuándo comenzó a ocurrir esto?
- ¿Qué cambió recientemente?
- ¿Ocurre en todos los entornos?
- ¿Podemos reproducirlo con código mínimo?
- ¿Cuál es el cambio más pequeño que lo desencadena?
```

### Fase 3: Comprender
Encuentre la causa raíz, no solo los síntomas.

```markdown
## Análisis de Causa Raíz
### Los 5 Porqués
1. ¿Por qué?: [Primera observación]
2. ¿Por qué?: [Razón más profunda]
3. ¿Por qué?: [Aún más profunda]
4. ¿Por qué?: [Acercándose más]
5. ¿Por qué?: [Causa raíz]
```

### Fase 4: Corregir y Verificar
Corrija el problema y verifique que esté realmente solucionado.

```markdown
## Verificación de Corrección
- [ ] El error ya no se reproduce
- [ ] La funcionalidad relacionada aún funciona
- [ ] No se han introducido nuevos problemas
- [ ] Se agregó una prueba para prevenir regresiones
```

## Lista de Verificación de Depuración

```markdown
## Antes de Comenzar
- [ ] Puedo reproducir consistentemente
- [ ] Tengo un caso mínimo de reproducción
- [ ] Entiendo el comportamiento esperado

## Durante la Investigación
- [ ] Verificar cambios recientes (git log)
- [ ] Revisar logs en busca de errores
- [ ] Agregar registro si es necesario
- [ ] Utilizar depurador/puntos de interrupción

## Después de la Corrección
- [ ] Causa raíz documentada
- [ ] Corrección verificada
- [ ] Prueba de regresión agregada
- [ ] Código similar revisado
```

## Comandos Comunes de Depuración

```bash
# Cambios recientes
git log --oneline -20
git diff HEAD~5

# Buscar patrón
grep -r "errorPattern" --include="*.ts"

# Revisar logs
pm2 logs app-name --err --lines 100
```

## Anti-Patrones

❌ **Cambios aleatorios** - "Tal vez si cambio esto..."
❌ **Ignorar evidencia** - "Eso no puede ser la causa"
❌ **Asumir** - "Debe ser X" sin pruebas
❌ **No reproducir primero** - Corregir ciegamente
❌ **Detenerse en síntomas** - No encontrar la causa raíz