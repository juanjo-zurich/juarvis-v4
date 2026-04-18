package verify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"juarvis/pkg/assets"
)

type CheckResult struct {
	Name    string
	Passed  bool
	Message string
}

type VerifyOptions struct {
	SkipBuild   bool
	SkipVet     bool
	SkipTest    bool
	SkipJSON    bool
	SkipPlugins bool
	SkipCLI     bool
}

func RunVerify(opts VerifyOptions) ([]CheckResult, error) {
	var results []CheckResult

	if !opts.SkipBuild {
		results = append(results, checkGoBuild())
	}
	if !opts.SkipVet {
		results = append(results, checkGoVet())
	}
	if !opts.SkipTest {
		results = append(results, checkGoTest())
	}
	if !opts.SkipJSON {
		results = append(results, checkEmbeddedJSONs())
	}
	if !opts.SkipPlugins {
		results = append(results, checkPluginManifests())
	}
	if !opts.SkipCLI {
		results = append(results, checkCLICommands())
	}

	return results, nil
}

func checkGoBuild() CheckResult {
	cmd := exec.Command("go", "build", "./...")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return CheckResult{Name: "go build", Passed: false, Message: strings.TrimSpace(string(output))}
	}
	return CheckResult{Name: "go build", Passed: true, Message: "Compilación exitosa"}
}

func checkGoVet() CheckResult {
	cmd := exec.Command("go", "vet", "./...")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return CheckResult{Name: "go vet", Passed: false, Message: strings.TrimSpace(string(output))}
	}
	return CheckResult{Name: "go vet", Passed: true, Message: "Sin warnings"}
}

func checkGoTest() CheckResult {
	// Updated cmd to include a timeout of 5 minutes
	cmd := exec.Command("go", "test", "./...", "-cover", "-timeout", "5m")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return CheckResult{Name: "go test", Passed: false, Message: strings.TrimSpace(string(output))}
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ok") || strings.HasPrefix(line, "coverage:") {
			return CheckResult{Name: "go test", Passed: true, Message: line}
		}
	}
	return CheckResult{Name: "go test", Passed: true, Message: "Todos los tests pasan"}
}

func checkEmbeddedJSONs() CheckResult {
	efs, err := assets.GetEmbeddedFS()
	if err != nil {
		return CheckResult{Name: "embedded JSON", Passed: false, Message: fmt.Sprintf("no se pudo acceder a assets: %v", err)}
	}

	jsonFiles := []string{"marketplace.json", "opencode.json", "permissions.yaml"}
	for _, f := range jsonFiles {
		data, err := fs.ReadFile(efs, f)
		if err != nil {
			return CheckResult{Name: "embedded JSON", Passed: false, Message: fmt.Sprintf("%s no encontrado: %v", f, err)}
		}
		if strings.HasSuffix(f, ".json") {
			var v interface{}
			if err := json.Unmarshal(data, &v); err != nil {
				return CheckResult{Name: "embedded JSON", Passed: false, Message: fmt.Sprintf("%s JSON inválido: %v", f, err)}
			}
		}
	}
	return CheckResult{Name: "embedded JSON", Passed: true, Message: "Todos los JSON embebidos son válidos"}
}

func checkPluginManifests() CheckResult {
	efs, err := assets.GetEmbeddedFS()
	if err != nil {
		return CheckResult{Name: "plugin manifests", Passed: false, Message: fmt.Sprintf("no se pudo acceder a assets: %v", err)}
	}

	entries, err := fs.ReadDir(efs, "plugins")
	if err != nil {
		return CheckResult{Name: "plugin manifests", Passed: false, Message: fmt.Sprintf("no se pudo leer plugins: %v", err)}
	}

	pluginCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		manifestPath := filepath.Join("plugins", entry.Name(), ".juarvis-plugin", "plugin.json")
		data, err := fs.ReadFile(efs, manifestPath)
		if err != nil {
			continue
		}
		var manifest map[string]interface{}
		if err := json.Unmarshal(data, &manifest); err != nil {
			return CheckResult{Name: "plugin manifests", Passed: false, Message: fmt.Sprintf("%s/plugin.json inválido: %v", entry.Name(), err)}
		}
		for _, field := range []string{"name", "version", "description", "category"} {
			if _, ok := manifest[field]; !ok {
				return CheckResult{Name: "plugin manifests", Passed: false, Message: fmt.Sprintf("%s/plugin.json falta campo: %s", entry.Name(), field)}
			}
		}
		pluginCount++
	}

	if pluginCount == 0 {
		return CheckResult{Name: "plugin manifests", Passed: true, Message: "Plugins OK (manifests se crean en init)"}
	}
	return CheckResult{Name: "plugin manifests", Passed: true, Message: fmt.Sprintf("%d plugins verificados", pluginCount)}
}

func checkCLICommands() CheckResult {
	binPath, err := findBinary()
	if err != nil {
		return CheckResult{Name: "CLI commands", Passed: false, Message: fmt.Sprintf("binario no encontrado: %v", err)}
	}

	commands := []string{"--help", "--version", "init --help", "check --help"}
	for _, cmd := range commands {
		parts := strings.Fields(cmd)
		c := exec.Command(binPath, parts...)
		var stderr bytes.Buffer
		c.Stderr = &stderr
		if err := c.Run(); err != nil {
			return CheckResult{Name: "CLI commands", Passed: false, Message: fmt.Sprintf("juarvis %s falló: %s", cmd, stderr.String())}
		}
	}
	return CheckResult{Name: "CLI commands", Passed: true, Message: "Todos los comandos responden correctamente"}
}

func findBinary() (string, error) {
	if _, err := os.Stat("./juarvis"); err == nil {
		abs, _ := filepath.Abs("./juarvis")
		return abs, nil
	}
	path, err := exec.LookPath("juarvis")
	if err == nil {
		return path, nil
	}
	return "", fmt.Errorf("juarvis binary not found")
}
