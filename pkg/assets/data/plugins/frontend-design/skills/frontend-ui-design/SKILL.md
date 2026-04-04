---
name: frontend-design
description: Principios de diseno frontend: tipografia, color, espaciado, layouts modernos, componentes esteticos
trigger: frontend design, UI design, typography, color scheme, layout, spacing, responsive design, CSS, componentes visuales
---

# Frontend Design Principles

Principios de diseno para crear interfaces web atractivas, funcionales y accesibles.

## Cuando Activar

- Creando componentes UI desde cero
- Disenando layouts de paginas o secciones
- Eligiendo esquemas de color o tipografia
- Mejorando la estetica visual de una interfaz
- Implementando diseno responsive
- Creando sistemas de diseno o design tokens

## Tipografia

### Jerarquia Tipografica

```
H1: 2.5rem (40px) — Titulo principal, una vez por pagina
H2: 2rem (32px) — Secciones principales
H3: 1.5rem (24px) — Subsecciones
H4: 1.25rem (20px) — Sub-subsecciones
Body: 1rem (16px) — Texto principal
Small: 0.875rem (14px) — Texto secundario, captions
```

### Reglas de Tipografia

- Maximo 2 fuentes por proyecto (una para headings, otra para body)
- Line-height: 1.5 para body text, 1.2 para headings
- Maximo 75 caracteres por linea para legibilidad
- Usar `clamp()` para tipografia fluida: `font-size: clamp(1rem, 2vw, 1.5rem)`
- Contraste minimo 4.5:1 para texto normal, 3:1 para texto grande

## Color

### Sistema de Color

```css
:root {
  /* Primario */
  --color-primary-50: #eff6ff;
  --color-primary-100: #dbeafe;
  --color-primary-500: #3b82f6;
  --color-primary-600: #2563eb;
  --color-primary-700: #1d4ed8;

  /* Neutros */
  --color-gray-50: #f9fafb;
  --color-gray-100: #f3f4f6;
  --color-gray-500: #6b7280;
  --color-gray-900: #111827;

  /* Semanticos */
  --color-success: #10b981;
  --color-warning: #f59e0b;
  --color-error: #ef4444;
  --color-info: #3b82f6;
}
```

### Reglas de Color

- Regla 60-30-10: 60% color dominante, 30% secundario, 10% acento
- Usar colores semanticos consistentes (verde = exito, rojo = error)
- Probar accesibilidad con herramientas como WebAIM Contrast Checker
- Los estados interactivos deben tener diferencia visual clara (hover, focus, active)
- Dark mode: invertir escala de grises, ajustar saturacion de colores

## Espaciado

### Sistema de Espaciado (base 4px)

```
4px   — Separacion minima (iconos pequenos)
8px   — Espaciado compacto (items de lista)
12px  — Espaciado medio-small
16px  — Espaciado estandar (padding de cards)
24px  — Espaciado medio (entre secciones)
32px  — Espaciado grande (entre componentes)
48px  — Espaciado XL (entre secciones de pagina)
64px  — Espaciado 2XL (hero sections)
96px  — Espaciado 3XL (separacion mayor)
```

### Reglas de Espaciado

- Usar sistema de 8px como base (4px para ajustes finos)
- Espaciado vertical > espaciado horizontal para secciones
- Padding interno consistente en componentes similares
- El espacio en blanco es una herramienta de diseno, no espacio desperdiciado
- Agrupar elementos relacionados (Ley de proximidad de Gestalt)

## Layout

### Patrones de Layout Modernos

#### 1. Container Centrado

```css
.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 24px;
}
```

#### 2. Grid de Cards

```css
.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 24px;
}
```

#### 3. Holy Grail Layout

```css
.page {
  display: grid;
  grid-template:
    "header header" auto
    "sidebar main" 1fr
    "footer footer" auto
    / 280px 1fr;
  min-height: 100vh;
}
```

#### 4. Stack Layout

```css
.stack > * + * {
  margin-top: 16px;
}
```

### Reglas de Layout

- Mobile-first: disenar para movil primero, luego escalar
- Usar `grid` para layouts 2D, `flexbox` para layouts 1D
- Breakpoints tipicos: 640px (sm), 768px (md), 1024px (lg), 1280px (xl)
- Contenido debe ser legible sin zoom en movil
- Touch targets minimo 44x44px

## Componentes

### Card

```
┌─────────────────────────────┐
│ [Imagen opcional]           │
│                             │
│ Titulo de la Card           │
│ Descripcion breve del       │
│ contenido o accion.         │
│                             │
│ [Accion principal]          │
└─────────────────────────────┘
```

- Padding: 24px
- Border-radius: 8-12px
- Sombra sutil: `0 1px 3px rgba(0,0,0,0.1)`
- Hover: sombra mas pronunciada + ligero translateY(-2px)

### Botones

```
Primario: Fondo solido, texto blanco, padding 12px 24px
Secundario: Borde, fondo transparente, mismo padding
Ghost: Sin borde, sin fondo, solo texto
Danger: Rojo para acciones destructivas
```

- Estados: default, hover, focus, active, disabled
- Icono + texto es mejor que solo icono
- Texto de accion claro: "Guardar cambios" no "Enviar"

### Formularios

- Labels siempre visibles (no solo placeholders)
- Mensajes de error inline debajo del campo
- Validacion en tiempo real despues del primer blur
- Agrupar campos relacionados con fieldset/legend
- Indicador de progreso en formularios largos

## Animaciones

### Principios

- Duracion: 150-300ms para interacciones UI
- Easing: `ease-out` para entradas, `ease-in` para salidas
- Propiedades animables: opacity, transform (GPU accelerated)
- NUNCA animar: width, height, top, left (causan reflow)
- Respetar `prefers-reduced-motion`

### Micro-interacciones

```css
/* Hover suave */
.button {
  transition: transform 150ms ease-out, box-shadow 150ms ease-out;
}
.button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
}

/* Fade in */
.fade-in {
  animation: fadeIn 300ms ease-out;
}
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}
```

## Accesibilidad

- Contraste minimo WCAG AA: 4.5:1 texto normal, 3:1 texto grande
- Navegacion por teclado: todos los elementos interactivos alcanzables con Tab
- Focus visible: outline claro en elementos focuseados
- Alt text descriptivo en imagenes
- ARIA labels donde el texto visual no es suficiente
- Formularios: labels asociados con `for`/`id`

## Checklist de Diseno Visual

- [ ] Jerarquia visual clara (que es lo mas importante?)
- [ ] Espaciado consistente (sistema de 8px)
- [ ] Colores accesibles (contraste suficiente)
- [ ] Tipografia legible (tamano, line-height, longitud de linea)
- [ ] Estados interactivos definidos (hover, focus, active, disabled)
- [ ] Responsive en 3 breakpoints minimo
- [ ] Animaciones sutiles y con proposito
- [ ] Navegacion por teclado funcional

**Recuerda**: El buen diseno es invisible. El usuario no nota cuando el espaciado es correcto, pero si nota cuando no lo es. Prioriza consistencia sobre creatividad en los fundamentos.
