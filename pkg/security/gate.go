package security

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"juarvis/pkg/hookify"
	"juarvis/pkg/root"

	"gopkg.in/yaml.v3"
)

// SecurityGateEvalúa las tres capas de seguridad en cascada
type SecurityGate struct {
	sandbox *Sandbox
	permissions *Permissions
	hooks    *hookify.Hookify
	enabledLayers []string // ["sandbox", "permissions", "hookify"]
}

type Result struct {
	Allowed bool
	Layer   string // Capa que bloqueó
	Message string
	AutoFix *AutoFixAction
}

type AutoFixAction struct {
	Command string
	Files   []string
}

// NewSecurityGate crea un gate de seguridad unificado
func NewSecurityGate(rootPath string) (*SecurityGate, error) {
	g := &SecurityGate{
		enabledLayers: []string{"sandbox", "permissions", "hookify"},
	}

	// Capa 1: Sandbox
	g.sandbox = NewSandboxGuard(rootPath)

	// Capa 2: Permissions.yaml
	g.permissions = NewPermissions(rootPath)

	// Capa 3: Hookify
	g.hooks = hookify.New(rootPath)

	return g, nil
}

// Evalúa todas las capas en cascada
func (g *SecurityGate) Eval(ctx context.Context, command string, args []string) *Result {
	for _, layer := range g.enabledLayers {
		switch layer {
		case "sandbox":
			if err := g.sandbox.Check(command); err != nil {
				return &Result{Allowed: false, Layer: "sandbox", Message: err.Error()}
			}

		case "permissions":
			if !g.permissions.Allow(command) {
				reason := g.permissions.Reason(command)
				return &Result{Allowed: false, Layer: "permissions", Message: reason}
			}

		case "hookify":
			fix := g.hooks.CheckAutoFix(command, args)
			if fix != nil {
				return &Result{
					Allowed:  true,
					Layer:    "hookify",
					Message: "Auto-fix disponible",
					AutoFix: &AutoFixAction{
						Command: fix.Command,
						Files:   fix.Files,
					},
				}
			}
		}
	}

	return &Result{Allowed: true, Layer: "all", Message: "allowed"}
}

// DisableLayer desactiva una capa
func (g *SecurityGate) DisableLayer(layer string) {
	var remaining []string
	for _, l := range g.enabledLayers {
		if l != layer {
			remaining = append(remaining, l)
		}
	}
	g.enabledLayers = remaining
}

// EnableLayeractiva una capa
func (g *SecurityGate) EnableLayer(layer string) {
	for _, l := range g.enabledLayers {
		if l == layer {
			return // Ya activo
		}
	}
	g.enabledLayers = append(g.enabledLayers, layer)
}

// =====================
// CAPA 1: SANDBOX
// =====================

type Sandbox struct {
	workspace string
	blacklist []string
}

func NewSandboxGuard(workspace string) *Sandbox {
	return &Sandbox{
		workspace: workspace,
		blacklist: []string{
			"rm -rf /",
			"dd if=",
			">/dev/sda",
			"mkfs",
			":(){:|:&};:", // Fork bomb
		},
	}
}

func (s *Sandbox) Check(cmd string) error {
	full := cmd
	for _, b := range s.blacklist {
		if strings.Contains(full, b) {
			return fmt.Errorf("comando en blacklist: %s", b)
		}
	}
	// Verificar workspace
	abs, _ := filepath.Abs(s.workspace)
	if !strings.HasPrefix(abs, abs) {
		return fmt.Errorf("comando sale del workspace")
	}
	return nil
}

// =====================
// CAPA 2: PERMISSIONS
// =====================

type Permissions struct {
	rules map[string][]Rule
}

type Rule struct {
	Pattern string `yaml:"pattern"`
	Action  string `yaml:"action"` // allow, deny, warn
	Reason  string `yaml:"reason"`
}

func NewPermissions(rootPath string) *Permissions {
	p := &Permissions{rules: make(map[string][]Rule)}

	// Cargar permissions.yaml
	permFile := filepath.Join(rootPath, "permissions.yaml")
	if data, err := os.ReadFile(permFile); err == nil {
		yaml.Unmarshal(data, &p.rules)
	}

	return p
}

func (p *Permissions) Allow(cmd string) bool {
	rules, ok := p.rules["bash"]
	if !ok {
		return true // Sin reglas = permitir
	}

	for _, r := range rules {
		if r.Action == "deny" && strings.Contains(cmd, r.Pattern) {
			return false
		}
	}
	return true
}

func (p *Permissions) Reason(cmd string) string {
	rules := p.rules["bash"]
	for _, r := range rules {
		if r.Action == "deny" && strings.Contains(cmd, r.Pattern) {
			return r.Reason
		}
	}
	return "denied by permissions.yaml"
}

// =====================
// EJEMPLO DE USO
// =====================

/*
// En comandos CLI:
func runWithSecurity(ctx context.Context, cmd string, args []string) error {
	root, _ := root.GetRoot()
	gate, err := security.NewSecurityGate(root)
	if err != nil {
		return err
	}

	result := gate.Eval(ctx, cmd, args)
	if !result.Allowed {
		return fmt.Errorf("[%s] %s", result.Layer, result.Message)
	}

	// Ejecutar comando
	execCmd := exec.Command(cmd, args...)
	return execCmd.Run()
}
*/