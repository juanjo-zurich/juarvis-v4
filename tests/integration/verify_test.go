//go:build integration

package integration

import (
	"strings"
	"testing"
)

func TestVerify_Passes(t *testing.T) {
	output, err := runJuarvis(t, "verify")
	if err != nil {
		t.Fatalf("juarvis verify failed: %v\n%s", err, output)
	}

	if !strings.Contains(output, "go build") {
		t.Error("expected verify to check go build")
	}
	if !strings.Contains(output, "go test") {
		t.Error("expected verify to check go test")
	}
}
