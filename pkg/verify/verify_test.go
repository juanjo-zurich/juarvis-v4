package verify

import (
	"testing"
)

func TestRunVerify(t *testing.T) {
	// Se crea una configuración que salta los comandos de CLI y Build 
	// para que el test sea más rápido y no dependa del binario
	opts := VerifyOptions{
		SkipBuild: true,
		SkipCLI:   true,
	}

	results, err := RunVerify(opts)
	if err != nil {
		t.Fatalf("RunVerify falló: %v", err)
	}

	if len(results) == 0 {
		t.Error("Se esperaban resultados de verificación, se obtuvo 0")
	}

	for _, res := range results {
		t.Logf("Check: %s, Passed: %v, Message: %s", res.Name, res.Passed, res.Message)
	}
}

func TestRunVerify(t *testing.T) {
	results, err := RunVerify(VerifyOptions{})
	if err != nil {
		t.Fatalf("RunVerify() error = %v", err)
	}
	// Should have returned results for all checks
	if len(results) == 0 {
		t.Error("RunVerify() should return at least one result")
	}
	// Verify all results have required fields
	for i, r := range results {
		if r.Name == "" {
			t.Errorf("Result[%d].Name is empty", i)
		}
		if r.Message == "" {
			t.Errorf("Result[%d].Message is empty", i)
		}
	}
}
