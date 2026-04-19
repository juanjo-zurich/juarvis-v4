package root

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetRoot obtiene el directorio raíz del ecosistema Juarvis
func GetRoot() (string, error) {
	if envRoot := os.Getenv("JUARVIS_ROOT"); envRoot != "" {
		if _, err := os.Stat(filepath.Join(envRoot, "marketplace.json")); err == nil {
			return envRoot, nil
		}
		return "", fmt.Errorf("JUARVIS_ROOT apunta a un directorio invalido (sin marketplace.json): %s", envRoot)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("no se pudo obtener directorio actual: %w", err)
	}

	current := cwd
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(current, "marketplace.json")); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", fmt.Errorf("no se encontro un ecosistema Juarvis. Usa --root o ejecuta 'juarvis init'")
}
