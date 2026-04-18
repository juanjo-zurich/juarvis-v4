package verify

import (
	"testing"
)

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
}
