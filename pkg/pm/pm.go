package pm

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"juarvis/pkg/assets"
	"juarvis/pkg/config"
	"juarvis/pkg/output"
	"juarvis/pkg/root"
	"juarvis/pkg/utils"
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

// httpClient con timeout para evitar bloqueos infinitos
var httpClient = &http.Client{Timeout: 10 * time.Second}

var httpGetFunc = func(url string) (*http.Response, error) {
	// Skip network en CI para evitar timeouts
	if os.Getenv("JUARVIS_SKIP_NETWORK") == "true" {
		return nil, fmt.Errorf("network disabled (JUARVIS_SKIP_NETWORK=true)")
	}
	return httpClient.Get(url)
}

var pluginCache map[string]string
var pluginCacheMu sync.RWMutex

var (
	lastRequestTime map[string]time.Time
	requestMu       sync.Mutex
)

// httpGetWithRetry realiza una petición HTTP con throttle y retry con backoff exponencial.
func httpGetWithRetry(url string, maxRetries int) (*http.Response, error) {
	endpoint := url
	if idx := strings.Index(url, "?"); idx > 0 {
		endpoint = url[:idx]
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Throttle: esperar 1s mínimo entre requests al mismo endpoint
		requestMu.Lock()
		if lastRequestTime == nil {
			lastRequestTime = make(map[string]time.Time)
		}
		if last, ok := lastRequestTime[endpoint]; ok {
			wait := time.Second - time.Since(last)
			if wait > 0 {
				requestMu.Unlock()
				time.Sleep(wait)
				requestMu.Lock()
			}
		}
		lastRequestTime[endpoint] = time.Now()
		requestMu.Unlock()

		resp, err := httpGetFunc(url)
		if err != nil {
			if attempt == maxRetries {
				return nil, err
			}
			time.Sleep(time.Duration(1<<attempt) * time.Second)
			continue
		}

		// Retry en 429 o 5xx
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			retryAfter := time.Duration(1<<attempt) * time.Second
			if resp.StatusCode == 429 {
				if ra := resp.Header.Get("Retry-After"); ra != "" {
					if secs, err := strconv.Atoi(ra); err == nil {
						retryAfter = time.Duration(secs) * time.Second
					}
				}
			}
			resp.Body.Close()
			if attempt == maxRetries {
				return nil, fmt.Errorf("rate limit o error del servidor (HTTP %d)", resp.StatusCode)
			}
			time.Sleep(retryAfter)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("agotados %d intentos para %s", maxRetries+1, url)
}

func loadMarketplace() (*Marketplace, error) {
	output.Info("Sincronizando con ecosistema global remoto (Vercel Agent Skills / skills.sh)...")

	// Prioridad 1: Obtener Skills oficiales de Vercel Labs vía GitHub API
	resp, reqErr := httpGetWithRetry("https://api.github.com/repos/vercel-labs/agent-skills/contents/skills", 2)
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
	if err == nil {
		file, err := os.ReadFile(filepath.Join(rootPath, "marketplace.json"))
		if err == nil {
			var market Marketplace
			if err := json.Unmarshal(file, &market); err != nil {
				return nil, fmt.Errorf("JSON corrupto: %w", err)
			}
			return &market, nil
		}
	}

	// Prioridad 3: Fallback al marketplace embebido en el binario
	embeddedFS, embErr := assets.GetEmbeddedFS()
	if embErr == nil {
		file, err := fs.ReadFile(embeddedFS, "marketplace.json")
		if err == nil {
			var market Marketplace
			if err := json.Unmarshal(file, &market); err != nil {
				return nil, fmt.Errorf("marketplace embebido corrupto: %w", err)
			}
			return &market, nil
		}
	}

	return nil, fmt.Errorf("no se encontro marketplace.json en ningun sitio")
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

// getOfficialProviders obtiene la whitelist de proveedores.
// Carga defaults y los fusiona con .juar/providers.json si existe.
func getOfficialProviders() map[string]bool {
	providers := map[string]bool{
		"vercel-labs":      true,
		"github":           true,
		"google-labs-code": true,
		"vercel":           true,
		"sveltejs":         true,
		"google-gemini":    true,
		"resend":           true,
	}

	rootPath, err := root.GetRoot()
	if err == nil {
		customFile := filepath.Join(rootPath, ".juar", "providers.json")
		if data, err := os.ReadFile(customFile); err == nil {
			var custom []string
			if json.Unmarshal(data, &custom) == nil {
				for _, p := range custom {
					providers[p] = true
				}
			}
		}
	}
	return providers
}

func isOfficialProvider(source string) bool {
	parts := strings.Split(source, "/")
	if len(parts) > 0 {
		providers := getOfficialProviders()
		if providers[parts[0]] {
			return true
		}
	}
	return false
}

func SearchPlugins(query string) {
	if len(query) < 2 {
		output.Error("La búsqueda requiere al menos 2 caracteres")
		return
	}
	output.Info("Buscando '%s' en el directorio global (skills.sh - Múltiples proveedores)...", query)
	searchURL := fmt.Sprintf("https://skills.sh/api/search?q=%s&limit=100", query)
	resp, err := httpGetWithRetry(searchURL, 2)
	if err != nil {
		output.Error("Error contactando el directorio global de skills: %v", err)
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
	pluginCacheMu.Lock()
	if path, ok := pluginCache[name]; ok {
		if _, err := os.Stat(path); err == nil {
			pluginCacheMu.Unlock()
			return path, nil
		}
		delete(pluginCache, name)
	}
	pluginCacheMu.Unlock()

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
		manifestFile := filepath.Join(path, config.JuarvisPluginDir, "plugin.json")
		var plug Plugin
		if data, err := os.ReadFile(manifestFile); err == nil {
			_ = json.Unmarshal(data, &plug)
			if plug.Name == name {
				pluginCacheMu.Lock()
				if pluginCache == nil {
					pluginCache = make(map[string]string)
				}
				pluginCache[name] = path
				pluginCacheMu.Unlock()
				return path, nil
			}
		} else if e.Name() == name || "juarvis-"+e.Name() == name {
			pluginCacheMu.Lock()
			if pluginCache == nil {
				pluginCache = make(map[string]string)
			}
			pluginCache[name] = path
			pluginCacheMu.Unlock()
			return path, nil
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

	targetFile := filepath.Join(pluginDir, config.JuarvisPluginDir, "enabled")

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

	err = os.RemoveAll(pluginDir)
	if err == nil {
		pluginCacheMu.Lock()
		delete(pluginCache, name)
		pluginCacheMu.Unlock()
	}
	return err
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
	switch {
	case strings.HasPrefix(targetPlugin.Source, "ext:"):
		sParts := strings.Split(strings.TrimPrefix(targetPlugin.Source, "ext:"), "|")
		err = installExternalSkill(sParts[0], sParts[1], pluginDir)
	case strings.HasPrefix(targetPlugin.Source, "vercel:"):
		skillName := strings.TrimPrefix(targetPlugin.Source, "vercel:")
		err = installVercelSkill(skillName, pluginDir)
	case strings.HasPrefix(targetPlugin.Source, "http"):
		err = installFromGit(targetPlugin.Source, pluginDir, targetPlugin.Name)
	default:
		err = installFromLocal(targetPlugin.Source, pluginDir, rootPath)
	}
	if err == nil {
		pluginCacheMu.Lock()
		delete(pluginCache, targetPlugin.Name)
		pluginCacheMu.Unlock()
	}
	return err
}

func getGlobalCacheDir() string {
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		// Fallback manual si falla la detección del sistema
		home, _ := os.UserHomeDir()
		cacheRoot = filepath.Join(home, ".cache")
	}
	cacheDir := filepath.Join(cacheRoot, "juarvis", "repos")
	_ = os.MkdirAll(cacheDir, 0755)
	return cacheDir
}

func syncRepoToCache(repoUrl string) (string, error) {
	cacheDir := getGlobalCacheDir()
	// Crear un nombre de directorio seguro basado en la URL
	repoName := strings.ReplaceAll(strings.ReplaceAll(repoUrl, "https://", ""), "/", "_")
	repoPath := filepath.Join(cacheDir, repoName)

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		output.Info("Clonando repositorio en caché global: %s", repoUrl)
		cmd := exec.Command("git", "clone", "--depth", "1", repoUrl, repoPath)
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("error clonando repositorio: %s", string(out))
		}
	} else {
		output.Info("Actualizando repositorio en caché: %s", repoUrl)
		cmd := exec.Command("git", "-C", repoPath, "pull")
		if _, err := cmd.CombinedOutput(); err != nil {
			// Si falla el pull (ej. rama cambiada), borramos y re-clonamos
			_ = os.RemoveAll(repoPath)
			return syncRepoToCache(repoUrl)
		}
	}
	return repoPath, nil
}

func installExternalSkill(repoUrl, skillName, destDir string) error {
	parsed, err := url.Parse(repoUrl)
	if err != nil {
		return fmt.Errorf("URL invalida: %w", err)
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("solo se permiten URLs https, rechazado: %s", repoUrl)
	}

	repoPath, err := syncRepoToCache(repoUrl)
	if err != nil {
		return err
	}

	// Identificar la ruta correcta de la skill
	skillDir := filepath.Join(repoPath, "skills", skillName)
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		skillDir = repoPath // Asumir que la skill está en la raíz
	}

	if err := copyDir(skillDir, destDir); err != nil {
		return fmt.Errorf("error copiando la skill externa: %w", err)
	}

	if err := utils.CreatePluginManifest(destDir, skillName, "1.0.0", "External Provider Skill", "external-skills"); err != nil {
		return err
	}
	return nil
}

func installVercelSkill(skillName, destDir string) error {
	repoUrl := "https://github.com/vercel-labs/agent-skills.git"
	repoPath, err := syncRepoToCache(repoUrl)
	if err != nil {
		return err
	}

	skillDir := filepath.Join(repoPath, "skills", skillName)
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return fmt.Errorf("la skill '%s' no existe en el repositorio de Vercel", skillName)
	}

	if err := copyDir(skillDir, destDir); err != nil {
		return fmt.Errorf("error copiando la skill '%s': %w", skillName, err)
	}

	// Inyectar el manifiesto plugin.json de Juarvis para compatibilidad nativa
	return utils.CreatePluginManifest(destDir, skillName, "1.0.0", "Vercel Agent Skill oficial", "vercel-skills")
}

func installFromGit(gitURL, destDir, pluginName string) error {
	parsed, err := url.Parse(gitURL)
	if err != nil {
		return fmt.Errorf("URL invalida: %w", err)
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("solo se permiten URLs https, rechazado: %s", gitURL)
	}
	cmd := exec.Command("git", "clone", "--depth", "1", gitURL, destDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error clonando repositorio: %s", string(output))
	}

	// Crear estructura .juarvis-plugin si no existe
	manifestDir := filepath.Join(destDir, config.JuarvisPluginDir)
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
}`, pluginName, gitURL)
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

	return utils.CopyEmbeddedDir(embeddedFS, embedPath, destDir)
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

var pluginVersions = make(map[string][]string)

func loadPluginVersions() {
	rootPath, _ := root.GetRoot()
	if rootPath == "" {
		return
	}

	pluginsDir := filepath.Join(rootPath, "plugins")
	entries, _ := os.ReadDir(pluginsDir)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pluginDir := filepath.Join(pluginsDir, e.Name(), config.JuarvisPluginDir)
		manifestFile := filepath.Join(pluginDir, "plugin.json")
		data, err := os.ReadFile(manifestFile)
		if err != nil {
			continue
		}
		var plug Plugin
		if err := json.Unmarshal(data, &plug); err != nil {
			continue
		}
		versions := []string{plug.Version}
		backupFile := filepath.Join(pluginDir, "version.history")
		if backupData, err := os.ReadFile(backupFile); err == nil {
			versions = append(versions, strings.Split(string(backupData), "\n")...)
		}
		pluginVersions[e.Name()] = versions
	}
}

func CheckUpdates() error {
	loadPluginVersions()
	market, err := loadMarketplace()
	if err != nil {
		return err
	}
	output.Info("Verificando actualizaciones...")
	hasUpdates := false

	matchedPlugins := make(map[string]bool)
	for _, p := range market.Plugins {
		if localVersions, ok := pluginVersions[p.Name]; ok {
			matchedPlugins[p.Name] = true
			if localVersions[0] != p.Version {
				output.Info("  %s: %s → %s", p.Name, localVersions[0], p.Version)
				hasUpdates = true
			}
		}
	}

	if !hasUpdates {
		output.Success("Todos los plugins están actualizados")
	}
	return nil
}

func UpdateAllPlugins(force bool) (int, error) {
	loadPluginVersions()
	market, err := loadMarketplace()
	if err != nil {
		return 0, err
	}

	// Get installed plugins
	rootPath, err := root.GetRoot()
	if err != nil {
		return 0, fmt.Errorf("error obteniendo root: %w", err)
	}
	pluginDir := filepath.Join(rootPath, "plugins")
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return 0, err
	}

	updated := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pluginName := e.Name()
		manifestPath := filepath.Join(pluginDir, pluginName, ".juarvis-plugin", "plugin.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue // Skip if no manifest
		}

		var p Plugin
		if err := json.Unmarshal(data, &p); err != nil {
			continue
		}

		// Find in marketplace
		var latest *Plugin
		for _, mp := range market.Plugins {
			if mp.Name == p.Name {
				latest = &mp
				break
			}
		}

		if latest == nil {
			continue
		}

		// Check if update needed
		if latest.Version != p.Version || force {
			if err := UpdatePlugin(p.Name); err != nil {
				output.Warning("Error actualizando %s: %v", p.Name, err)
				continue
			}
			updated++
		}
	}

	return updated, nil
}

func UpdatePlugin(name string) error {
	pluginDir, err := findPluginDir(name)
	if err != nil {
		return err
	}

	manifestFile := filepath.Join(pluginDir, config.JuarvisPluginDir, "plugin.json")
	data, err := os.ReadFile(manifestFile)
	if err != nil {
		return fmt.Errorf("error leyendo manifest: %w", err)
	}

	var current Plugin
	if err := json.Unmarshal(data, &current); err != nil {
		return fmt.Errorf("error parseando manifest: %w", err)
	}

	market, err := loadMarketplace()
	if err != nil {
		return err
	}

	var latest *Plugin
	for _, p := range market.Plugins {
		if p.Name == name {
			latest = &p
			break
		}
	}

	if latest == nil || latest.Version == current.Version {
		return fmt.Errorf("no hay actualización disponible para %s", name)
	}

	backupManifest := filepath.Join(pluginDir, config.JuarvisPluginDir, "version.history")
	oldVersion := current.Version
	if existing, err := os.ReadFile(backupManifest); err == nil {
		oldVersion = string(existing) + "\n" + current.Version
	}
	if err := os.WriteFile(backupManifest, []byte(oldVersion), 0644); err != nil {
		return fmt.Errorf("error creando backup: %w", err)
	}

	current.Version = latest.Version
	data, err = json.MarshalIndent(current, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializando: %w", err)
	}
	if err := os.WriteFile(manifestFile, data, 0644); err != nil {
		return fmt.Errorf("error escribiendo manifest: %w", err)
	}

	output.Success("Plugin '%s' actualizado: %s → %s", name, latest.Version, current.Version)
	return nil
}

func RollbackPlugin(name string) error {
	pluginDir, err := findPluginDir(name)
	if err != nil {
		return err
	}

	backupManifest := filepath.Join(pluginDir, config.JuarvisPluginDir, "version.history")
	data, err := os.ReadFile(backupManifest)
	if err != nil {
		return fmt.Errorf("no hay historial para %s", name)
	}

	versions := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(versions) == 0 {
		return fmt.Errorf("historial vacío para %s", name)
	}

	currentVersion := versions[len(versions)-1]
	versions = versions[:len(versions)-1]

	manifestFile := filepath.Join(pluginDir, config.JuarvisPluginDir, "plugin.json")
	manifestData, err := os.ReadFile(manifestFile)
	if err != nil {
		return err
	}

	var current Plugin
	if err := json.Unmarshal(manifestData, &current); err != nil {
		return err
	}

	current.Version = currentVersion
	newData, _ := json.MarshalIndent(current, "", "  ")
	if err := os.WriteFile(manifestFile, newData, 0644); err != nil {
		return err
	}

	if len(versions) > 0 {
		os.WriteFile(backupManifest, []byte(strings.Join(versions, "\n")), 0644)
	} else {
		os.Remove(backupManifest)
	}

	output.Success("Plugin '%s' revertido a %s", name, current.Version)
	return nil
}
