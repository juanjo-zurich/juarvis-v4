package verify

import (
	"testing"
)

func TestRunVerify(t *testing.T) {
	opts := VerifyOptions{
		SkipBuild:   true,
		SkipTest:    true, // IMPORTANTE: skip test check para evitar loop infinito
		SkipCLI:     true,
		SkipJSON:    true, // Los assets ya se testean en otros tests
		SkipPlugins: true,
	}

	results, err := RunVerify(opts)
	if err != nil {
		t.Fatalf("RunVerify falló: %v", err)
	}

	if len(results) == 0 {
		t.Error("Se esperaba al menos un resultado de verificación")
	}
}
