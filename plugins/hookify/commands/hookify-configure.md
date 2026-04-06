---
description: Configurar reglas hookify existentes
argument-hint: <nombre-regla> [--enable|--disable|--priority alta|media|baja]
---

## /hookify-configure

Configura una regla hookify existente: habilitar, deshabilitar o cambiar prioridad.

### Uso
/hookify-configure <nombre> --enable
/hookify-configure <nombre> --disable
/hookify-configure <nombre> --priority alta|media|baja

### Comportamiento
1. Buscar la regla en `~/.juarvis/hookify.*.local.md` y `plugins/*/hooks/`
2. Si no se encuentra, mostrar error con sugerencia de crearla con `/hookify`
3. Validar la opción proporcionada
4. Actualizar el frontmatter YAML de la regla
5. Ejecutar `scripts/plugin-loader.sh --reload-hooks` para aplicar cambios
6. Mostrar confirmación con el nuevo estado

### Opciones disponibles
- `--enable`: Establecer `enabled: true` en el frontmatter
- `--disable`: Establecer `enabled: false` en el frontmatter
- `--priority <nivel>`: Cambiar prioridad (baja, media, alta)

### Flujo de configuración
1. Validar que el nombre de la regla no está vacío
2. Buscar el archivo de la regla:
   - Local: `~/.juarvis/hookify.<nombre>.local.md`
   - Plugin: `plugins/*/hooks/<nombre>.md`
3. Si no se encuentra:
   - Mostrar: "Regla '<nombre>' no encontrada"
   - Sugerir: "Usa `/hookify` para crear una nueva regla"
4. Leer el frontmatter actual
5. Aplicar los cambios solicitados
6. Escribir el archivo actualizado
7. Recargar hooks si es necesario
8. Mostrar confirmación

### Ejemplos
```
/hookify-configure no-console-log-ts --disable
/hookify-configure block-test-deletion --enable
/hookify-configure auto-format-on-save --priority alta
```

### Notas
- Las reglas de plugins solo pueden modificarse si el plugin lo permite
- Las reglas locales siempre pueden modificarse
- Los cambios se aplican inmediatamente tras la configuración
