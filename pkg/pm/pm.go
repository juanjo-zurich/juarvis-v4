package pm

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"juarvis/pkg/assets"
	"juarvis/pkg/output"
	"juarvis/pkg/root"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Marketplace representa la estructura del JSON del catálogo
type Marketplace struct {
	Name    string   `json:"name"`
	Plugins []Plugin `json:"plugins"`
}

type Plugin struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Source      string `json:"source"`
	Category    string `json:"category"`
}

func loadMarketplace() (*Marketplace, error) {
	output.Info("Sincronizando con ecosistema global remoto (Vercel Agent Skills / skills.sh)...")
	
	// Prioridad 1: Obtener Skills oficiales de Vercel Labs vía GitHub API
	resp, reqErr := http.Get("https://api.github.com/repos/vercel-labs/agent-skills/contents/skills")
	if reqErr == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		var contents []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&contents); err == nil {
			var market Marketplace
			market.Name = "Vercel Agent Skills & Juarvis"
			for _, item := range contents {
				name, ok1 := item["name"].(string)
				itemType, ok2 := item["type"].(string)
				if ok1 && ok2 && itemType == "dir" {
					market.Plugins = append(market.Plugins, Plugin{
						Name:        name,
						Description: "Official Agent Skill: " + name,
						Version:     "1.0.0",
						Source:      "vercel:" + name,
						Category:    "vercel-skills",
					})
				}
			}
			return &market, nil
		}
	}

	output.Warning("Límite de GitHub API excedido o sin conexión. Usando catálogo offline.")

	// Prioridad 2: Fallback al marketplace local si no hay conexión
	rootPath, err := root.GetRoot()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo root para fallback: %w", err)
	}
	file, err := os.ReadFile(filepath.Join(rootPath, "marketplace.json"))
	if err != nil {
		// Prioridad 3: Fallback al marketplace embebido en el binario

		embeddedFS, embErr := assets.GetEmbeddedFS()
		if embErr == nil {
			file, err = fs.ReadFile(embeddedFS, "marketplace.json")
			if err == nil {
				var market Marketplace
				if err := json.Unmarshal(file, &market); err != nil {
					return nil, fmt.Errorf("marketplace embebido corrupto: %v", err)
				}
				return &market, nil
			}
		}
		return nil, fmt.Errorf("no se encontro marketplace.json en %s", rootPath)
	}

	var market Marketplace
	if err := json.Unmarshal(file, &market); err != nil {
		return nil, fmt.Errorf("JSON corrupto: %v", err)
	}
	return &market, nil
}

func ListPlugins() {
	market, err := loadMarketplace()
	if err != nil {
		output.Error("%v", err)
		return
	}

	output.Info("Catálogo: %s", market.Name)
	headers := []string{"NAME", "CATEGORY", "VERSION", "DESCRIPTION"}
	rows := [][]string{}
	for _, p := range market.Plugins {
		rows = append(rows, []string{p.Name, p.Category, p.Version, p.Description})
	}
	output.PrintTable(headers, rows)
}

type SkillsSearchResult struct {
	Skills []struct {
		ID       string `json:"id"`
		SkillID  string `json:"skillId"`
		Name     string `json:"name"`
		Installs int    `json:"installs"`
		Source   string `json:"source"`
	} `json:"skills"`
}

// Proveedores seguros verificados para mitigar orígenes maliciosos
var officialProviders = map[string]bool{
	"vercel-labs":      true,
	"github":           true,
	"google-labs-code": true,
	"vercel":           true,
	"sveltejs":         true,
	"google-gemini":    true,
	"resend":           true,
}

func isOfficialProvider(source string) bool {
	parts := strings.Split(source, "/")
	if len(parts) > 0 && officialProviders[parts[0]] {
		return true
	}
	return false
}

func SearchPlugins(query string) {
	if len(query) < 2 {
		output.Error("La búsqueda requiere al menos 2 caracteres")
		return
	}
	output.Info("Buscando '%s' en el directorio global (skills.sh - Múltiples proveedores)...", query)
	url := fmt.Sprintf("https://skills.sh/api/search?q=%s&limit=100", query) // Ampliado a 100 para compensar el filtro
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		output.Error("Error contactando el directorio global de skills")
		return
	}
	defer resp.Body.Close()
	
	var res SkillsSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		output.Error("Error procesando los resultados de búsqueda")
		return
	}

	headers := []string{"ID DE INSTALACIÓN (COPIA ESTO)", "NOMBRE", "PROVEEDOR OFICIAL", "DESCARGAS"}
	rows := [][]string{}

	// Filtrar solo los oficiales para proteger al usuario
	count := 0
	for _, s := range res.Skills {
		if isOfficialProvider(s.Source) {
			rows = append(rows, []string{s.ID, s.Name, s.Source + " ✅", fmt.Sprintf("%d", s.Installs)})
			count++
			if count >= 20 { // Cap de visualización
				break
			}
		}
	}

	if len(rows) == 0 {
		output.Warning("No se encontraron skills de proveedores OFICIALES para '%s'.", query)
		output.Info("Agente: Te corresponde a ti crearla y documentarla para el usuario.")
		output.Info("=> Ejecuta 'juarvis skill create %s' para generar el esqueleto local.", strings.ReplaceAll(query, " ", "-"))
		return
	}

	output.PrintTable(headers, rows)
}

func findPluginDir(name string) (string, error) {
	rootPath, err := root.GetRoot()
	if err != nil {
		return "", fmt.Errorf("error obteniendo root: %w", err)
	}
	pluginDir := filepath.Join(rootPath, "plugins")

	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return "", err
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(pluginDir, e.Name())
		manifestFile := filepath.Join(path, ".juarvis-plugin", "plugin.json")
		var plug Plugin
		if data, err := os.ReadFile(manifestFile); err == nil {
			json.Unmarshal(data, &plug)
			if plug.Name == name {
				return path, nil
			}
		} else {
			// Fallback si no tiene manifest pero la carpeta se llama igual
			if e.Name() == name || "juarvis-"+e.Name() == name {
				return path, nil
			}
		}
	}
	return "", fmt.Errorf("plugin '%s' no encontrado en el sistema de archivos", name)
}

// SetPluginStatus cambia el estado (habilitado/deshabilitado) creando el fichero 'enabled'
func SetPluginStatus(name string, enabled bool) error {
	pluginDir, err := findPluginDir(name)
	if err != nil {
		return err
	}

	targetFile := filepath.Join(pluginDir, ".juarvis-plugin", "enabled")

	val := "false"
	if enabled {
		val = "true"
	}

	errWrite := os.WriteFile(targetFile, []byte(val), 0644)
	if errWrite != nil {
		return fmt.Errorf("error al escribir estado: %v", errWrite)
	}
	return nil
}

// RemovePlugin borra recursivamente un plugin instalado
func RemovePlugin(name string) error {
	pluginDir, err := findPluginDir(name)
	if err != nil {
		return err
	}

	return os.RemoveAll(pluginDir)
}

// InstallPlugin instala un plugin desde el marketplace o directamente desde un proveedor (owner/repo/skill)
func InstallPlugin(pluginName string) error {
	var targetPlugin *Plugin
	parts := strings.Split(pluginName, "/")

	if len(parts) >= 2 { // Instalación dinámica desde proveedor (owner/repo/skillId)
		owner := parts[0]
		
		// 🛡️ BARRERA DE SEGURIDAD (Zero-Trust): Solo permitimos Organizaciones Oficiales
		if !isOfficialProvider(owner) {
			return fmt.Errorf("🛡️  ALERTA DE SEGURIDAD: Instalación bloqueada. El proveedor '%s' no está en la lista blanca oficial (verified). Operación abortada para evitar skills maliciosas", owner)
		}

		repo := parts[1]
		repoUrl := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
		skillFolder := repo
		if len(parts) >= 3 {
			skillFolder = parts[2]
		}

		targetPlugin = &Plugin{
			Name:        skillFolder,
			Description: "Plugin externo: " + pluginName,
			Version:     "1.0.0",
			Source:      "ext:" + repoUrl + "|" + skillFolder,
			Category:    "external-provider",
		}
	} else {
		market, err := loadMarketplace()
		if err != nil {
			return fmt.Errorf("error cargando marketplace: %w", err)
		}

		for _, p := range market.Plugins {
			if p.Name == pluginName {
				targetPlugin = &p
				break
			}
		}

		if targetPlugin == nil {
			return fmt.Errorf("plugin '%s' no encontrado. Usa 'juarvis pm search <query>' para buscar en la red", pluginName)
		}
	}

	rootPath, err := root.GetRoot()
	if err != nil {
		return fmt.Errorf("error obteniendo root: %w", err)
	}

	pluginDir := filepath.Join(rootPath, "plugins", targetPlugin.Name)

	if _, err := os.Stat(pluginDir); err == nil {
		return fmt.Errorf("plugin '%s' ya instalado. Usa 'juarvis pm remove %s' primero", targetPlugin.Name, targetPlugin.Name)
	}

	// Determinar tipo de fuente
	if strings.HasPrefix(targetPlugin.Source, "ext:") {
		sParts := strings.Split(strings.TrimPrefix(targetPlugin.Source, "ext:"), "|")
		return installExternalSkill(sParts[0], sParts[1], pluginDir)
	}
	if strings.HasPrefix(targetPlugin.Source, "vercel:") {
		skillName := strings.TrimPrefix(targetPlugin.Source, "vercel:")
		return installVercelSkill(skillName, pluginDir)
	}
	if strings.HasPrefix(targetPlugin.Source, "http") {
		return installFromGit(targetPlugin.Source, pluginDir, targetPlugin.Name)
	}
	return installFromLocal(targetPlugin.Source, pluginDir, rootPath)
}

func installExternalSkill(repoUrl, skillName, destDir string) error {
	output.Info("Clonando repositorio de proveedor externo (%s)...", repoUrl)
	tmpDir, err := os.MkdirTemp("", "ext-skill")
	if err != nil {
		return fmt.Errorf("error creando directorio temporal: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("git", "clone", "--depth", "1", repoUrl, tmpDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error descargando repositorio proveedor: %s", string(out))
	}

	// Identificar la ruta correcta de la skill
	skillDir := filepath.Join(tmpDir, "skills", skillName)
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		skillDir = tmpDir // Asumir que la skill está en la raíz
	}

	if err := copyDir(skillDir, destDir); err != nil {
		return fmt.Errorf("error copiando la skill externa: %w", err)
	}

	manifestDir := filepath.Join(destDir, ".juarvis-plugin")
	os.MkdirAll(manifestDir, 0755)
	manifest := fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "External Provider Skill",
  "category": "external-skills"
}`, skillName)
	os.WriteFile(filepath.Join(manifestDir, "plugin.json"), []byte(manifest), 0644)
	return nil
}

func installVercelSkill(skillName, destDir string) error {
	output.Info("Clonando repositorio oficial de Vercel Agent Skills...")
	tmpDir, err := os.MkdirTemp("", "vercel-skill")
	if err != nil {
		return fmt.Errorf("error creando directorio temporal: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("git", "clone", "--depth", "1", "https://github.com/vercel-labs/agent-skills.git", tmpDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error descargando vercel skills: %s", string(out))
	}

	skillDir := filepath.Join(tmpDir, "skills", skillName)
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return fmt.Errorf("la skill '%s' no existe en el repositorio de Vercel", skillName)
	}

	if err := copyDir(skillDir, destDir); err != nil {
		return fmt.Errorf("error copiando la skill '%s': %w", skillName, err)
	}

	// Inyectar el manifiesto plugin.json de Juarvis para compatibilidad nativa
	manifestDir := filepath.Join(destDir, ".juarvis-plugin")
	os.MkdirAll(manifestDir, 0755)
	manifest := fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "Vercel Agent Skill oficial",
  "category": "vercel-skills"
}`, skillName)
	os.WriteFile(filepath.Join(manifestDir, "plugin.json"), []byte(manifest), 0644)
	
	return nil
}

func installFromGit(url, destDir, pluginName string) error {
	cmd := exec.Command("git", "clone", "--depth", "1", url, destDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error clonando repositorio: %s", string(output))
	}

	// Crear estructura .juarvis-plugin si no existe
	manifestDir := filepath.Join(destDir, ".juarvis-plugin")
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio de manifiesto: %w", err)
	}

	// Crear plugin.json básico si no existe
	pluginJSON := filepath.Join(manifestDir, "plugin.json")
	if _, err := os.Stat(pluginJSON); os.IsNotExist(err) {
		manifest := fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "Plugin instalado desde %s",
  "category": "external"
}`, pluginName, url)
		if err := os.WriteFile(pluginJSON, []byte(manifest), 0644); err != nil {
			return fmt.Errorf("error creando plugin.json: %w", err)
		}
	}

	return nil
}

func installFromLocal(source, destDir, rootPath string) error {
	srcPath := filepath.Join(rootPath, source)
	if _, err := os.Stat(srcPath); err == nil {
		return copyDir(srcPath, destDir)
	}

	// Fallback: buscar en assets embebidos
	embeddedFS, embErr := assets.GetEmbeddedFS()
	if embErr != nil {
		return fmt.Errorf("fuente local no encontrada: %s y assets embebidos no disponibles", srcPath)
	}

	// source es algo como "./plugins/core" -> "plugins/core"
	embedPath := strings.TrimPrefix(source, "./")
	if _, err := fs.Stat(embeddedFS, embedPath); err != nil {
		return fmt.Errorf("fuente no encontrada ni en filesystem (%s) ni en assets embebidos (%s)", srcPath, embedPath)
	}

	return copyEmbeddedDir(embeddedFS, embedPath, destDir)
}

// copyDir copia un directorio recursivamente
func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("error leyendo directorio fuente: %w", err)
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("error creando directorio destino: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("error leyendo %s: %w", srcPath, err)
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return fmt.Errorf("error escribiendo %s: %w", dstPath, err)
			}
		}
	}

	return nil
}

// copyEmbeddedDir copia un directorio del embed.FS al filesystem
func copyEmbeddedDir(targetFS fs.FS, srcPath, destPath string) error {
	return fs.WalkDir(targetFS, srcPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath := strings.TrimPrefix(path, srcPath)
		if relPath == "" {
			relPath = "."
		}
		dest := filepath.Join(destPath, strings.TrimPrefix(relPath, string(filepath.Separator)))
		if relPath == "." {
			dest = destPath
		}
		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		content, err := fs.ReadFile(targetFS, path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, content, 0644)
	})
}
