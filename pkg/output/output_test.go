package output

import (
	"testing"
)

func TestSetJSONMode(t *testing.T) {
	SetJSONMode(true)
	if !IsJSONMode() {
		t.Error("expected JSON mode to be true")
	}
	SetJSONMode(false)
	if IsJSONMode() {
		t.Error("expected JSON mode to be false")
	}
}

func TestSuccess_NoPanic(t *testing.T) {
	SetJSONMode(false)
	Success("test %s", "message")
}

func TestError_NoPanic(t *testing.T) {
	SetJSONMode(false)
	Error("test %s", "error")
}

func TestWarning_NoPanic(t *testing.T) {
	SetJSONMode(false)
	Warning("test %s", "warning")
}

func TestInfo_NoPanic(t *testing.T) {
	SetJSONMode(false)
	Info("test %s", "info")
}

func TestPrintTable_NoPanic(t *testing.T) {
	SetJSONMode(false)
	PrintTable([]string{"A", "B"}, [][]string{{"1", "2"}, {"3", "4"}})
}

func TestPrintJSON_NoPanic(t *testing.T) {
	SetJSONMode(true)
	PrintJSON(map[string]string{"key": "value"})
	SetJSONMode(false)
}

func TestSuccess_JSONMode(t *testing.T) {
	SetJSONMode(true)
	Success("test")
	SetJSONMode(false)
}

func TestError_JSONMode(t *testing.T) {
	SetJSONMode(true)
	Error("test error")
	SetJSONMode(false)
}

func TestPrintTable_JSONMode(t *testing.T) {
	SetJSONMode(true)
	PrintTable([]string{"A", "B"}, [][]string{{"1", "2"}})
	SetJSONMode(false)
}

// TestExitCodes verifica que las constantes tienen los valores correctos
// y no colisionan entre sí.
func TestExitCodes_Values(t *testing.T) {
	codes := map[string]int{
		"ExitOK":           ExitOK,
		"ExitGeneric":      ExitGeneric,
		"ExitNoEcosystem":  ExitNoEcosystem,
		"ExitConfigError":  ExitConfigError,
		"ExitBuildFailed":  ExitBuildFailed,
		"ExitTestFailed":   ExitTestFailed,
		"ExitPermission":   ExitPermission,
		"ExitPluginError":  ExitPluginError,
		"ExitWatcherError": ExitWatcherError,
	}

	// Verificar que todos son distintos
	seen := make(map[int]string)
	for name, code := range codes {
		if prev, ok := seen[code]; ok {
			t.Errorf("exit code collision: %s y %s comparten el código %d", name, prev, code)
		}
		seen[code] = name
	}

	// Verificar valores esperados
	if ExitOK != 0 {
		t.Errorf("ExitOK debe ser 0, got %d", ExitOK)
	}
	if ExitGeneric != 1 {
		t.Errorf("ExitGeneric debe ser 1, got %d", ExitGeneric)
	}
}

// TestFatal_CallsExitWithCode verifica que Fatal llama al exitFunc
// con el código correcto sin terminar el proceso de test.
func TestFatal_CallsExitWithCode(t *testing.T) {
	original := exitFunc
	defer func() { exitFunc = original }()

	var capturedCode int
	exitFunc = func(code int) {
		capturedCode = code
	}

	Fatal(ExitNoEcosystem, "Ejecuta: juarvis init", "ecosistema no encontrado")

	if capturedCode != ExitNoEcosystem {
		t.Errorf("Fatal usó código %d, esperado %d (ExitNoEcosystem)", capturedCode, ExitNoEcosystem)
	}
}

// TestFatal_WithoutHint verifica que Fatal funciona sin pista accionable.
func TestFatal_WithoutHint(t *testing.T) {
	original := exitFunc
	defer func() { exitFunc = original }()

	var capturedCode int
	exitFunc = func(code int) { capturedCode = code }

	Fatal(ExitGeneric, "", "error genérico")

	if capturedCode != ExitGeneric {
		t.Errorf("Fatal usó código %d, esperado %d (ExitGeneric)", capturedCode, ExitGeneric)
	}
}

// TestFatal_WithFormat verifica que Fatal formatea el mensaje correctamente.
func TestFatal_WithFormat(t *testing.T) {
	original := exitFunc
	defer func() { exitFunc = original }()

	exitFunc = func(code int) {}

	// No debe entrar en pánico con args de formato
	Fatal(ExitPluginError, "Ejecuta: juarvis pm list", "plugin %q no encontrado", "mi-plugin")
}

// TestFatal_AllCodes verifica que todos los exit codes son aceptados por Fatal.
func TestFatal_AllCodes(t *testing.T) {
	original := exitFunc
	defer func() { exitFunc = original }()

	allCodes := []int{
		ExitOK, ExitGeneric, ExitNoEcosystem, ExitConfigError,
		ExitBuildFailed, ExitTestFailed, ExitPermission,
		ExitPluginError, ExitWatcherError,
	}

	for _, code := range allCodes {
		var got int
		exitFunc = func(c int) { got = c }
		Fatal(code, "", "test")
		if got != code {
			t.Errorf("Fatal(%d) llamó exitFunc con %d", code, got)
		}
	}
}
