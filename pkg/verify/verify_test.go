package verify

import (
	"testing"
)

// TestRunVerify verifica que la función principal de validación funcione correctamente.
// Se usa SkipBuild y SkipCLI para que el test no dependa del entorno externo.
func TestRunVerify(t *testing.T) {
	opts := VerifyOptions{
		SkipBuild: true,
		SkipCLI:   true,
	}

	results, err := RunVerify(opts)
	if err != nil {
		t.Fatalf("RunVerify falló inesperadamente: %v", err)
	}

	if len(results) == 0 {
		t.Error("Se esperaban resultados de verificación, pero se obtuvo una lista vacía")
	}

	// Verificamos que al menos uno de los checks haya corrido (ej. JSON o Plugins)
	foundCheck := false
	for _, res := range results {
		t.Logf("Check ejecutado: %s | Pasó: %v | Mensaje: %s", res.Name, res.Passed, res.Message)
		if res.Name != "" {
			foundCheck = true
		}
	}

	if !foundCheck {
		t.Error("Los resultados devueltos no contienen nombres de checks válidos")
	}
}
