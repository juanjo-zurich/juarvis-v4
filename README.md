# Juarvis V4 — Sistema Operativo para Agentes IA

Juarvis V4 es una herramienta de sistema que convierte cualquier carpeta en un entorno de desarrollo profesional gobernado por agentes de IA autónomos. 

No es solo un script; es un **motor global** que permite que IAs (Claude, Cursor, Windsurf, etc.) sigan protocolos de ingeniería rigurosos, autogestionen sus herramientas y aprendan de sus errores sin que tú tengas que configurar nada.

---

## 🚀 Guía Rápida para Humanos (Sin conocimientos técnicos)

Si solo quieres que tu IA empiece a trabajar de forma profesional, sigue estos 3 pasos:

### 1. Instalación (Solo una vez)
Descarga el proyecto y ejecútalo en tu terminal:
```bash
git clone https://github.com/juanjo-zurich/juarvis-v4.git
cd juarvis-v4
sudo make install
```
*Esto coloca el comando `juarvis` en tu ordenador como si fuera `git` o `npm`.*

### 2. Preparar tu Proyecto
Ve a la carpeta donde quieras crear tu aplicación y ejecuta:
```bash
juarvis up
```
*Esto hace todo por ti: crea las reglas, configura tu editor (Cursor, VSCode, etc.) y activa la vigilancia de seguridad.*

### 3. ¡Habla con tu IA!
Abre tu proyecto en tu editor favorito. Verás que ahora tu IA es mucho más inteligente. Dile algo como:
> "Crea una página de aterrizaje para mi nuevo negocio utilizando Astro."

**¡Y listo!** La IA detectará que tiene Juarvis instalado y empezará a crear snapshots y seguir protocolos de ingeniería automáticamente.

---

## 🧠 ¿Cómo funciona para los Agentes? (Modo Autónomo)

A diferencia de otras herramientas, Juarvis trata a los agentes de IA como **ingenieros senior autónomos**:

1.  **Protocolo de Misión**: Cada proyecto tiene un archivo `AGENTS.md` (generado por `juarvis init`) que actúa como "La Constitución". Los agentes están obligados a leerlo y seguir sus normas.
2.  **Caja Negra**: Los agentes usan el comando global `juarvis` del sistema. No necesitan entender el código fuente de Juarvis; simplemente usan sus capacidades (Snapshots, Memoria, SDD).
3.  **Bucle de Ingeniería (SDD)**: Para tareas grandes, los agentes inician el `sdd` (Spec-Driven Development) para asegurar que el diseño es perfecto antes de escribir código.
4.  **Auto-Aprovisionamiento**: Si a la IA le falta una habilidad, ella misma ejecuta `juarvis pm install` para adquirirla.

### Comandos Esenciales para Agentes (Tú también puedes usarlos):
- `juarvis snapshot create "titulo"`: Crea un punto de restauración del código.
- `juarvis sdd init`: Inicia una nueva misión guiada por especificaciones.
- `juarvis verify`: Comprueba que todo el proyecto esté "sano" y los tests pasen.
- `juarvis pm list`: Muestra todas las "habilidades" (skills) disponibles.

---

## 🛠 Arquitectura Simplificada

- **Nivel Global**: El binario `juarvis` en tu sistema. Es el cerebro que sabe ejecutar.
- **Nivel Local**: Los archivos `.juar/`, `agent-settings.json` y `AGENTS.md` en tu proyecto. Son los mapas que le dicen a la IA qué hacer en **ese** proyecto.

---

## ⚠️ Seguridad y Confianza
Juarvis vigila cada cambio. Si una IA intenta hacer algo peligroso (como borrar archivos clave o ignorar tests), el **Watcher** de Juarvis lo detectará y te avisará o detendrá la operación.

---

