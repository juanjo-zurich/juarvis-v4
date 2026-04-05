package pm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"juarvis/pkg/output"
	"juarvis/pkg/root"
)

// CreateSkill genera el andamiaje (scaffolding) para una nueva skill dentro del proyecto
func CreateSkill(name string) error {
	name = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	if name == "" {
		return fmt.Errorf("el nombre de la skill no puede estar vacío")
	}

	rootPath, err := root.GetRoot()
	if err != nil {
		return fmt.Errorf("error obteniendo raíz del proyecto: %w", err)
	}

	// Juarvis local skills always live in .agent/skills/
	skillDir := filepath.Join(rootPath, ".agent", "skills", name)
	if _, err := os.Stat(skillDir); err == nil {
		return fmt.Errorf("la skill '%s' ya existe en %s", name, skillDir)
	}

	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio de la skill: %w", err)
	}

	// Generar SKILL.md con el formato estándar exigido por Agent Skills (skills.sh compatible)
	skillMDPath := filepath.Join(skillDir, "SKILL.md")
	content := fmt.Sprintf(`---
name: %s
description: "PUNTO DE INTERVENCIÓN: Escribe aquí la descripción de lo que hace esta skill."
metadata:
  internal: false
---

# %s

Instrucciones para que el agente siga cuando esta skill se active.

## Cuándo usarla

Describe los escenarios en los que el agente debe activar esta skill.

## Dependencias

- Nombra aquí otras skills o herramientas necesarias.

## Pasos

1. Primer paso...
2. Segundo paso...

// TODO: Agente, rellena aquí con la guía y arquitectura necesaria para el proyecto.
`, name, strings.ToUpper(name))

	if err := os.WriteFile(skillMDPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error escribiendo SKILL.md: %w", err)
	}

	output.Success("Andamiaje de la skill '%s' creado exitosamente.", name)
	output.Info("Ruta: %s", skillMDPath)
	output.Info("Agente: Ahora puedes leer este archivo y rellenar las instrucciones de forma metódica.")

	return nil
}
