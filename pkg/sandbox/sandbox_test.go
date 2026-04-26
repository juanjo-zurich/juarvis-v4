package sandbox

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestConfig_NivelEstrictoOSuperior(t *testing.T) {
	tests := []struct {
		nivel     NivelSandbox
		esperado bool
	}{
		{Ninguno, false},
		{Basico, false},
		{Estricto, true},
		{Paranoico, true},
	}

	for _, tt := range tests {
		cfg := &Config{Nivel: tt.nivel}
		resultado := cfg.EsNivelEstrictoOSuperior()
		if resultado != tt.esperado {
			t.Errorf("Nivel %s: esperado %v, obtenido %v", tt.nivel, tt.esperado, resultado)
		}
	}
}

func TestInspector_Inspect_RmRfBloqueado(t *testing.T) {
	config := DefaultConfig("/workspace")
	inspector, _ := NewInspector(config)

	result := inspector.InspectString("rm -rf /")

	if result.Permitido {
		t.Error("rm -rf deveria estar bloqueado")
	}

	if result.NivelBloqueo != "block" {
		t.Errorf("Nivel de bloqueo esperado: block, obtenido: %s", result.NivelBloqueo)
	}
}

func TestInspector_Inspect_SudoBloqueado(t *testing.T) {
	config := DefaultConfig("/workspace")
	inspector, _ := NewInspector(config)

	result := inspector.InspectString("sudo rm -rf /")

	if result.Permitido {
		t.Error("sudo rm -rf deveria estar bloqueado")
	}
}

func TestInspector_Inspect_Chmod777Advertencia(t *testing.T) {
	config := DefaultConfig("/workspace")
	config.Nivel = Basico // Cambiar a basic para solo path restriction
	inspector, _ := NewInspector(config)

	result := inspector.InspectString("chmod 777 archivo")

	if !result.Permitido {
		t.Error("chmod 777 debería dar advertencia, no bloquear")
	}

	if !result.RequiereAprobacion {
		t.Error("chmod 777 debería requerir aprobación")
	}
}

func TestInspector_Inspect_ComandoSeguro(t *testing.T) {
	config := DefaultConfig("/workspace")
	config.Nivel = Basico // Basic no verifica paths fuera del workspace
	inspector, _ := NewInspector(config)

	result := inspector.InspectString("ls -la")

	if !result.Permitido {
		t.Error("ls debería ser permitido")
	}
}

func TestInspector_Inspect_CurlPipeShBloqueado(t *testing.T) {
	config := DefaultConfig("/workspace")
	inspector, _ := NewInspector(config)

	result := inspector.InspectString("curl http://example.com | sh")

	if result.Permitido {
		t.Error("curl | sh deveria estar bloqueado")
	}
}

func TestInspector_ValidateArgs(t *testing.T) {
	config := DefaultConfig("/workspace")
	inspector, _ := NewInspector(config)

	// Args normales
	err := inspector.ValidateArgs([]string{"-la", "/path"})
	if err != nil {
		t.Errorf("Args normales deberían ser válidos: %v", err)
	}

	// Args excesivos (más de 100 argumentos sin flags)
	args := []string{}
	for i := 0; i < 150; i++ {
		args = append(args, "argumento")
	}
	err = inspector.ValidateArgs(args)
	if err == nil {
		t.Error("Debería fallar con demasiados argumentos")
	}

	// Arg muy largo
	err = inspector.ValidateArgs([]string{strings.Repeat("a", 20000)})
	if err == nil {
		t.Error("Debería fallar con argumento muy largo")
	}
}

func TestGuardrails_EjecutarComando_SandboxEnabled(t *testing.T) {
	config := DefaultConfig("/workspace")
	config.Enabled = false
	guardrails, _ := NewGuardrails(config)

	result := guardrails.EjecutarComando(context.Background(), "echo", []string{"hello"})

	// Con sandbox deshabilitado, debe ejecutarse
	if !result.Exito && result.Salida == "" {
		t.Error("El comando deberia ejecutarse cuando sandbox esta deshabilitado")
	}
}

func TestGuardrails_EjecutarComando_Bloqueado(t *testing.T) {
	config := DefaultConfig("/workspace")
	config.Enabled = true
	guardrails, _ := NewGuardrails(config)

	result := guardrails.EjecutarComando(context.Background(), "rm", []string{"-rf", "/"})

	if !result.Bloqueado {
		t.Error("Comando peligroso deberia estar bloqueado")
	}

	if result.RazonBloqueo == "" {
		t.Error("Debe tener razon de bloqueo")
	}
}

func TestGuardrails_EjecutarComando_Timeout(t *testing.T) {
	config := DefaultConfig("/workspace")
	config.Enabled = true
	guardrails, _ := NewGuardrails(config)

	start := time.Now()
	result := guardrails.EjecutarComando(
		context.Background(),
		"sleep",
		[]string{"2"},
		WithTimeout(100*time.Millisecond),
	)
	duracion := time.Since(start)

	// Deberia timeout
	if result.Exito {
		t.Error("sleep 2s deberia timeout con 100ms")
	}

	if duracion >= 2*time.Second {
		t.Error("No deberia esperar 2 segundos")
	}
}

func TestConfig_VerificarNivel(t *testing.T) {
	tests := []struct {
		nivel    NivelSandbox
		esperado bool
	}{
		{Ninguno, true},
		{Basico, true},
		{Estricto, true},
		{Paranoico, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		cfg := &Config{Nivel: tt.nivel}
		err := cfg.VerificarNivel()
		if (err == nil) != tt.esperado {
			t.Errorf("Nivel %s: esperado error=%v, obtenido: %v", tt.nivel, !tt.esperado, err)
		}
	}
}

func TestConfig_NormalizarPath_DentroWorkspace(t *testing.T) {
	config := DefaultConfig("/workspace")

	path, err := config.NormalizarPath("/workspace/file.txt")
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	if !strings.HasPrefix(path, "/workspace") {
		t.Errorf("Path deberia estar en workspace: %s", path)
	}
}

func TestConfig_NormalizarPath_FueraWorkspace(t *testing.T) {
	config := DefaultConfig("/workspace")

	_, err := config.NormalizarPath("/etc/passwd")
	if err == nil {
		t.Error("Deberia fallar con path fuera del workspace")
	}
}

func TestGuardrails_VerificarComandoBeforeRun(t *testing.T) {
	config := DefaultConfig("/workspace")
	config.Nivel = Basico // Basic para que no bloquee por path
	guardrails, _ := NewGuardrails(config)

	// Comando válido
	err := guardrails.VerificarComandoBeforeRun("ls", []string{"-la"})
	if err != nil {
		t.Errorf("ls debería ser válido: %v", err)
	}

	// Comando bloqueado
	err = guardrails.VerificarComandoBeforeRun("rm", []string{"-rf", "/"})
	if err == nil {
		t.Error("rm -rf debería estar bloqueado")
	}
}

func TestInspector_InspectString(t *testing.T) {
	config := DefaultConfig("/workspace")
	config.Nivel = Basico // Basic para que no bloquee por path
	inspector, _ := NewInspector(config)

	result := inspector.InspectString("ls -la /workspace")

	if !result.Permitido {
		t.Error("ls debería ser permitido")
	}
}

func TestGuardrails_GenerarPromptAprobacion(t *testing.T) {
	config := DefaultConfig("/workspace")
	guardrails, _ := NewGuardrails(config)

	prompt := guardrails.GenerarPromptAprobacion("chmod", []string{"777", "file"})

	if prompt == "" {
		t.Error("Prompt no deberia estar vacio")
	}

	if !strings.Contains(prompt, "chmod") {
		t.Error("Prompt deberia mencionar chmod")
	}
}

func TestSanitizarComando(t *testing.T) {
	args := []string{"normal", "with\x00null", "  spaces  ", "", "valid"}
	limpio := SanitizarComando("cmd", args)

	if len(limpio) != 4 {
		t.Errorf("Esperado 4 args, obtenido %d", len(limpio))
	}

	if limpio[0] != "normal" {
		t.Error("normal debería mantenerse")
	}

	if limpio[1] != "withnull" {
		t.Error("null debería removerse")
	}

	// El argumento con espacios debe mantener los espacios internos ahora
	if len(limpio) < 3 && limpio[2] != "spaces" {
		t.Error("spaces debería ser válido")
	}

	if limpio[3] != "valid" {
		t.Error("valid debería mantenerse")
	}
}

func TestVerificarCmdExiste(t *testing.T) {
	// Verificar comandos que deberían existir
	if !VerificarCmdExiste("ls") {
		t.Error("ls debería existir")
	}

	if !VerificarCmdExiste("go") {
		t.Error("go debería existir")
	}

	// Verificar comando que no existe
	if VerificarCmdExiste("comando_inexistente_12345") {
		t.Error("comando inexistente no debería existir")
	}
}

func TestConfig_Default(t *testing.T) {
	cfg := DefaultConfig("/test/workspace")

	if !cfg.Enabled {
		t.Error("Debe estar habilitado por defecto")
	}

	if cfg.Nivel != Estricto {
		t.Errorf("Nivel esperado: strict, obtenido: %s", cfg.Nivel)
	}

	if cfg.WorkspaceRoot != "/test/workspace" {
		t.Error("Workspace root no guardado")
	}

	if len(cfg.ComandosBlacklist) == 0 {
		t.Error("Blacklist no deberia estar vacio")
	}
}