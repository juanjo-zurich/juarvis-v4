---
name: containerization-patterns
description: Muestra los patrones y buenas prácticas de Docker y Docker Compose para orquestar aplicaciones modernas, separando entornos de dev/prod y gestionando dependencias correctamente.
trigger: "Cuando el sistema deba contenerizar la aplicación, escribir un Dockerfile o configurar Docker Compose"
version: "1.1.0"
---

# Patrones de Contenerización

> Estrategias avanzadas para entornos consistentes usando Docker.

## 1. Patrones de Dockerfile (Best Practices)

- **Construcciones Multi-Etapa (Multi-stage Builds)**: 
  Usa múltiples bloques `FROM` para separar las dependencias de construcción (compiladores, SDKs, paquetes de dev) del entorno de ejecución. Esto reduce drásticamente el tamaño de la imagen final y minimiza la superficie de ataque.
  
- **Usuario No Root**: 
  Nunca ejecutes el proceso principal como root. Crea un usuario `appuser` o similar y utiliza directivas `USER` en el Dockerfile después de las instalaciones requeridas.

- **Optimización de Capas (Layer Caching)**:
  Ordena los comandos desde los menos frecuentes (instalación de SO/paquetes) hasta los más frecuentes (código fuente). Por ejemplo, copia `package.json` -> instala dependencias -> luego copia todo el código.

- **Uso de .dockerignore**:
  Excluye explícitamente `node_modules`, `venv`, `target`, `.git`, `.env` y archivos locales irrelevantes para reducir el tiempo de carga del contexto y evitar fugas de secretos.

- **Etiquetado Específico**:
  Evita `latest`. Fija la imagen base a una versión específica y determinista (ej. `node:20.11.0-alpine3.19`). Usa imágenes mínimas como `alpine` o `distroless` para producción.

## 2. Patrones de Docker Compose

- **Aislamiento de Servicios (Service Isolation)**:
  Separa la aplicación en contenedores lógicos (`backend`, `frontend`, `db`, `cache`). No metas base de datos y aplicación en el mismo contenedor.

- **Montajes para Desarrollo (Bind Mounts)**:
  Usa `volumes: - ./:/app` para sincronizar el código local dentro del contenedor y permitir hot-reloading en desarrollo sin reconstruir la imagen.

- **Gestión de Salud (Healthchecks)**:
  No uses solo `depends_on`. Usa contenedores que esperen activamente a que los servicios requeridos estén listos:
  ```yaml
  depends_on:
    db:
      condition: service_healthy
  ```

- **Redes Privadas**:
  Crea redes overlay o bridge para que los servicios se comuniquen por nombre interno sin exponer puertos al host externo innecesariamente. Solo expón al host el frontend o la API pública.

- **Gestión de Variables de Entorno**:
  Pasa variables dinámicas usando un `.env` configurado en `env_file`. No pongas claves en el archivo compose.

## 3. Ejemplo de Estructura de Producción vs Dev

- **Dev**: Usa `docker-compose.yml`. Orientado a facilidad de uso, volúmenes locales montados, puertos expuestos y herramientas de debug activadas.
- **Prod**: Usa `docker-compose.prod.yml` (o Kubernetes/Helm). Imágenes puras sin volúmenes locales de código, reinicios configurados, límites de recursos (CPU/RAM) definidos explícitamente, y logs estructurados vía drivers.

## 4. Archivos de Inicialización

Utiliza scripts (por ejemplo, `entrypoint.sh`) para tareas en tiempo de arranque, como ejecutar migraciones de base de datos (`npm run migrate`, `alembic upgrade head`) antes de iniciar la aplicación principal, garantizando que el entorno esté listo.
