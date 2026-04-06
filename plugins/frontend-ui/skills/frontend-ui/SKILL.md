---
name: frontend-ui
description: Estrategias basadas en investigacion para crear UIs atractivas y visualmente distintivas
trigger: UI aesthetics, visual design, beautiful UI, UI polish, design system, componente visual, interfaz atractiva
---

# Frontend UI — Research-Backed Aesthetics

Estrategias de diseno UI basadas en investigacion de Anthropic para crear interfaces visualmente atractivas.

## Cuando Activar

- Creando interfaces que necesitan ser visualmente atractivas
- Mejorando la estetica de componentes existentes
- Disenando dashboards o paginas de presentacion
- Creando sistemas de diseno coherentes
- Pulendo detalles visuales de una aplicacion

## Principios de Estetica UI

### 1. Consistencia Visual

Todos los elementos similares deben verse similares:
- Botones del mismo nivel → mismo estilo
- Cards del mismo tipo → mismo padding, sombra, border-radius
- Inputs → mismo estilo de borde, focus, error

**Regla**: Si dos elementos cumplen la misma funcion, deben verse iguales.

### 2. Jerarquia Visual por Contraste

```
Nivel 1 (mas importante): Texto mas grande, color primario, posicion prominente
Nivel 2 (secundario): Tamano medio, color secundario
Nivel 3 (terciario): Texto pequeno, color gris, posicion discreta
```

**Tecnicas de jerarquia**:
- Tamano (el mas efectivo)
- Color (primario vs neutro)
- Peso (bold vs regular)
- Espacio (mas espacio = mas importancia)
- Posicion (arriba/izquierda = mas importante en LTR)

### 3. Espacio en Blanco como Herramienta

El espacio en blanco no es espacio vacio — es una herramienta activa de diseno:

```
Poco espacio → Sensacion de densidad, urgencia, informacion compacta
Mucho espacio → Sensacion de premium, calma, importancia
```

**Reglas practicas**:
- Hero sections: 96px+ de padding vertical
- Secciones de contenido: 48-64px entre secciones
- Dentro de cards: 24px de padding
- Entre elementos relacionados: 8-16px
- Entre elementos no relacionados: 32px+

### 4. Sombras y Profundidad

Sistema de sombras coherente:

```css
/* Sombra sutil (cards en reposo) */
shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);

/* Sombra media (cards en hover) */
shadow-md: 0 4px 6px rgba(0, 0, 0, 0.07), 0 2px 4px rgba(0, 0, 0, 0.06);

/* Sombra grande (modals, dropdowns) */
shadow-lg: 0 10px 15px rgba(0, 0, 0, 0.1), 0 4px 6px rgba(0, 0, 0, 0.05);

/* Sombra extra grande (popovers, tooltips) */
shadow-xl: 0 20px 25px rgba(0, 0, 0, 0.1), 0 10px 10px rgba(0, 0, 0, 0.04);
```

**Regla**: La elevacion visual debe corresponder con la importancia del elemento.

### 5. Border Radius Consistente

```
0px    — Serio, corporativo, datos
4px    — Sutil, profesional
8px    — Moderno, amigable (default recomendado)
12px   — Soft, approachable
16px+  — Jugueton, casual
```

**Regla**: Elegir UN radio y usarlo consistentemente en toda la aplicacion. Variar solo para casos especiales (avatars = 50%, pills = 9999px).

### 6. Color con Proposito

**Paleta funcional**:
```
Primario: Acciones principales, links, estados activos
Secundario: Acciones alternativas, informacion complementaria
Exito: Confirmaciones, completado, positivo
Advertencia: Precaucion, atencion necesaria
Error: Fallos, eliminacion, negativo
Info: Informacion neutral, ayuda
```

**Regla del color**: Si todo es destacado, nada lo es. Usar color primario con moderacion (maximo 1-2 elementos primarios por viewport).

## Patrones de Componentes Atractivos

### Hero Section

```
┌───────────────────────────────────────────┐
│                                           │
│           [Badge opcional]                │
│                                           │
│        Titulo Grande y Claro              │
│                                           │
│     Descripcion breve y persuasiva        │
│                                           │
│     [CTA Principal]  [CTA Secundario]     │
│                                           │
│        [Imagen/Ilustracion]               │
│                                           │
└───────────────────────────────────────────┘
```

- Centrado para impacto maximo
- Titulo: clamp(2rem, 5vw, 3.5rem)
- Descripcion: max-width 600px, centrado
- CTA primario visible sin scroll

### Dashboard Card

```
┌─────────────────────────────┐
│ Icono  Titulo        [Menu] │
│                              │
│  1,234                      │
│  +12% vs mes anterior       │
│                              │
│  ▁▃▅▇▆▄▂▁▃▅ (sparkline)    │
└─────────────────────────────┘
```

- Metrica grande y prominente
- Contexto comparativo (vs periodo anterior)
- Visualizacion mini si aplica
- Accion accesible desde la card

### Lista de Items

```
┌─────────────────────────────┐
│ [Avatar] Nombre       [→]  │
│          Descripcion       │
├─────────────────────────────┤
│ [Avatar] Nombre       [→]  │
│          Descripcion       │
└─────────────────────────────┘
```

- Separador sutil entre items
- Hover state en toda la fila
- Click area = toda la fila, no solo el texto
- Avatar/imagen para reconocimiento rapido

## Tecnicas de Pulido Visual

### 1. Transiciones Suaves

```css
/* Todo elemento interactivo debe tener transicion */
.interactive {
  transition: all 200ms ease-out;
}

/* Mejor: transiciones especificas por propiedad */
.card {
  transition: transform 200ms ease-out,
              box-shadow 200ms ease-out;
}
```

### 2. Skeleton Loading

Mostrar esqueletos animados en vez de spinners:

```css
.skeleton {
  background: linear-gradient(
    90deg,
    #f0f0f0 25%,
    #e0e0e0 50%,
    #f0f0f0 75%
  );
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}
```

### 3. Empty States

Nunca mostrar una lista/tabla vacia sin contexto:

```
┌─────────────────────────────┐
│                             │
│       [Ilustracion]         │
│                             │
│    Todavia no hay items     │
│    Crea tu primer item      │
│    para empezar             │
│                             │
│    [Crear primer item]      │
│                             │
└─────────────────────────────┘
```

### 4. Error States Visuales

```
┌─────────────────────────────┐
│                             │
│       ⚠️ [Icono]            │
│                             │
│    Algo salio mal           │
│    No pudimos cargar los    │
│    datos. Intenta de nuevo. │
│                             │
│    [Reintentar]             │
│                             │
└─────────────────────────────┘
```

## Checklist de UI Atractiva

- [ ] Espaciado consistente (sistema de 8px)
- [ ] Jerarquia visual clara en cada pantalla
- [ ] Colores con proposito, no decorativos
- [ ] Transiciones en todos los estados interactivos
- [ ] Empty states con contexto y accion
- [ ] Error states utiles y amigables
- [ ] Loading states informativos (skeleton > spinner)
- [ ] Border radius consistente
- [ ] Sombras proporcionales a la elevacion
- [ ] Tipografia con jerarquia clara
- [ ] Contraste suficiente para accesibilidad
- [ ] Responsive en movil y desktop

**Recuerda**: La diferencia entre una UI buena y una excelente esta en los detalles — el espaciado exacto, la transicion suave, el empty state considerado. Estos detalles son lo que hace que una aplicacion se sienta "profesional".
