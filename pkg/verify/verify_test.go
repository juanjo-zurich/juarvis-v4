package verify

import (
	"testing"
)

// TestRunVerify verifica que el motor de validación funcione.
// Usamos SkipBuild y SkipCLI para evitar que el test falle si el binario no está presente.
func TestRunVerify(t *testing.T) {
	opts := VerifyOptions{
		SkipBuild: true,
		SkipCLI:   true,
	}

	results, err := RunVerify(opts)
	if err != nil {
		t.Fatalf("RunVerify falló: %v", err)
	}

	if len(results) == 0 {
		t.Error("Se esperaba al menos un resultado de verificación")
	}

	for _, res := range results {
		t.Logf("Resultado de %s: %v (%s)", res.Name, res.Passed, res.Message)
	}
}
