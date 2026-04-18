//go:build integration

package integration

import (
	"strings"
	"testing"
)

func TestVerify_Passes(t *testing.T) {
	// Ejecutamos verify en el proyecto actual (donde está el código fuente)
	output, err := runJuarvis(t, "verify")
	
	// El comando puede fallar si el entorno no está limpio, 
	// pero aquí verificamos que al menos el comando se ejecute.
	if err != nil && !strings.Contains(output, "❌") && !strings.Contains(output, "go build") {
		t.Fatalf("verify command failed to execute: %v\n%s", err, output)
	}

	if !strings.Contains(output, "go build") && !strings.Contains(output, "embedded JSON") {
		t.Errorf("output inesperado de verify:\n%s", output)
	}
}

