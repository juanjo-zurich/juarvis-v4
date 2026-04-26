package analyze

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"juarvis/pkg/config"
	"juarvis/pkg/output"
)

type ProjectInfo struct {
	Stack         []string       `json:"stack"`
	Conventions   []string       `json:"conventions"`
	Patterns      []string       `json:"patterns"`
	AntiPatterns  []string       `json:"anti_patterns"`
	Architecture  string         `json:"architecture"`
	FileCount     int            `json:"file_count"`
	LanguageStats map[string]int `json:"language_stats"`
}

var (
	stackDetectors = map[string][]string{
		"nextjs":     {"package.json:next", "next.config", "pages/", "app/"},
		"react":      {"package.json:react", "src/App.tsx", "src/index.jsx"},
		"vue":        {"package.json:vue", "src/main.vue", "vue.config"},
		"astro":      {"astro.config", "src/pages/index.astro"},
		"express":    {"express", "app.js", "server.js"},
		"fastapi":    {"fastapi", "main.py", "requirements.txt:fastapi"},
		"django":     {"django", "manage.py", "settings.py"},
		"rails":      {"config/application.rb", "Gemfile:rails"},
		"laravel":    {"artisan", "composer.json:laravel"},
		"go":         {"go.mod", "main.go", "cmd/"},
		"rust":       {"Cargo.toml", "src/main.rs"},
		"python":     {"requirements.txt", "pyproject.toml", "setup.py"},
		"typescript": {"tsconfig.json", "*.ts"},
		"prisma":     {"schema.prisma", "prisma/"},
		"postgresql": {"schema.prisma:postgresql", "postgres://"},
		"mysql":      {"schema.prisma:mysql", "mysql://"},
		"mongodb":    {"schema.prisma:mongodb", "mongodb://"},
		"tailwind":   {"tailwind.config", "postcss.config"},
		"docker":     {"Dockerfile", "docker-compose"},
		"graphql":    {"schema.graphql", "resolvers/"},
	}

	conventionDetectors = map[string][]string{
		"eslint":              {".eslintrc", ".eslintrc.js", ".eslintrc.json"},
		"prettier":            {".prettierrc", "prettier.config"},
		"vitest":              {"vitest.config", "package.json:vitest"},
		"jest":                {"jest.config", "package.json:jest"},
		"commitlint":          {".commitlintrc", "commitlint.config"},
		"husky":               {".husky/", "husky.config"},
		"conventionalcommits": {"package.json:commitizen"},
	}
)

// RunAnalyze analiza el proyecto actual (donde se ejecuta)
func RunAnalyze(update bool, verbose bool) error {
	rootPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("no se pudo obtener directorio actual: %w", err)
	}
	return runAnalyze(rootPath, update, verbose)
}

// RunAnalyzeIn analiza un proyecto específico (para usar desde init)
func RunAnalyzeIn(rootPath string, update bool, verbose bool) error {
	return runAnalyze(rootPath, update, verbose)
}

// GetProjectInfo returns project analysis without generating skills (for display purposes)
func GetProjectInfo(rootPath string) (ProjectInfo, error) {
	info := ProjectInfo{
		Stack:        detectStack(rootPath),
		Conventions:  detectConventions(rootPath),
		Patterns:     detectPatterns(rootPath),
		AntiPatterns: detectAntiPatterns(rootPath),
		Architecture: analyzeArchitecture(rootPath),
	}

	fileCount, langStats := countFiles(rootPath)
	info.FileCount = fileCount
	info.LanguageStats = langStats

	return info, nil
}

func runAnalyze(rootPath string, update bool, verbose bool) error {
	output.Info("🔍 Analizando codebase...")

	// Detectar stack
	stack := detectStack(rootPath)
	output.Info("✅ Stack detectado: %s", strings.Join(stack, ", "))

	// Detectar convenciones
	conventions := detectConventions(rootPath)
	output.Info("✅ %d convenciones extraídas", len(conventions))

	// Detectar patrones
	patterns := detectPatterns(rootPath)
	output.Info("✅ %d patrones detectados", len(patterns))

	// Extraer antipatrones del historial git
	antiPatterns := detectAntiPatterns(rootPath)
	output.Info("✅ %d antipatrones detectados del historial", len(antiPatterns))

	// Analizar arquitectura
	architecture := analyzeArchitecture(rootPath)

	// Contar archivos
	fileCount, langStats := countFiles(rootPath)

	info := ProjectInfo{
		Stack:         stack,
		Conventions:   conventions,
		Patterns:      patterns,
		AntiPatterns:  antiPatterns,
		Architecture:  architecture,
		FileCount:     fileCount,
		LanguageStats: langStats,
	}

	// Generar skills del proyecto
	if err := generateProjectSkills(rootPath, info, update); err != nil {
		return fmt.Errorf("error generando skills: %w", err)
	}

	output.Success("✅ Skills de proyecto generadas en .juar/skills/")
	output.Info("📊 Archivos analizados: %d", fileCount)

	return nil
}

func detectStack(rootPath string) []string {
	var stack []string

	// Leer package.json si existe
	pkgJSON := filepath.Join(rootPath, "package.json")
	if data, err := os.ReadFile(pkgJSON); err == nil {
		content := string(data)
		for name, patterns := range stackDetectors {
			for _, pattern := range patterns {
				if strings.Contains(content, pattern) {
					if !contains(stack, name) {
						stack = append(stack, name)
					}
					break
				}
			}
		}
	}

	// Verificar archivos específicos
	filesToCheck := []struct {
		filename string
		name     string
	}{
		{"go.mod", "go"},
		{"Cargo.toml", "rust"},
		{"requirements.txt", "python"},
		{"pyproject.toml", "python"},
		{"Gemfile", "ruby"},
		{"pom.xml", "java"},
		{"build.gradle", "kotlin"},
		{"schema.prisma", "prisma"},
		{"Dockerfile", "docker"},
		{"docker-compose.yml", "docker"},
	}

	for _, fc := range filesToCheck {
		if _, err := os.Stat(filepath.Join(rootPath, fc.filename)); err == nil {
			if !contains(stack, fc.name) {
				stack = append(stack, fc.name)
			}
		}
	}

	// Detectar frontend frameworks
	dirs := []string{"src", "lib", "app", "pages", "components"}
	for _, dir := range dirs {
		if _, err := os.Stat(filepath.Join(rootPath, dir)); err == nil {
			// Check for specific patterns
			for name, patterns := range stackDetectors {
				for _, pattern := range patterns {
					if strings.HasPrefix(pattern, dir+"/") || strings.HasPrefix(pattern, dir+"\\") {
						if !contains(stack, name) {
							stack = append(stack, name)
						}
					}
				}
			}
		}
	}

	// Defaults
	if len(stack) == 0 {
		stack = append(stack, "vanilla")
	}

	return stack
}

func detectConventions(rootPath string) []string {
	var conventions []string

	// Linters y formatters
	for name, files := range conventionDetectors {
		for _, file := range files {
			if _, err := os.Stat(filepath.Join(rootPath, file)); err == nil {
				conventions = append(conventions, name)
				break
			}
		}
	}

	// Estructura de directorios
	commonDirs := []string{"src/components", "src/hooks", "src/utils", "src/services", "lib/", "pkg/", "internal/"}
	for _, dir := range commonDirs {
		if _, err := os.Stat(filepath.Join(rootPath, dir)); err == nil {
			conventions = append(conventions, "modular:"+dir)
		}
	}

	// Testing
	if _, err := os.Stat(filepath.Join(rootPath, "__tests__")); err == nil {
		conventions = append(conventions, "unit-tests:__tests__")
	}
	if _, err := os.Stat(filepath.Join(rootPath, "tests")); err == nil {
		conventions = append(conventions, "integration-tests:tests")
	}

	// TypeScript check
	if _, err := os.Stat(filepath.Join(rootPath, "tsconfig.json")); err == nil {
		conventions = append(conventions, "typescript-strict")
	}

	return conventions
}

func detectPatterns(rootPath string) []string {
	var patterns []string

	// Detectar patrones comunes por archivos
	patternFiles := map[string][]string{
		"hooks":      {"src/hooks", "lib/hooks"},
		"context":    {"src/context", "lib/context"},
		"components": {"src/components", "components/"},
		"services":   {"src/services", "lib/services"},
		"utils":      {"src/utils", "lib/utils"},
		"api":        {"src/api", "api/", "routes/"},
		"models":     {"src/models", "models/"},
		"middleware": {"src/middleware", "middleware/"},
	}

	for name, paths := range patternFiles {
		for _, path := range paths {
			if _, err := os.Stat(filepath.Join(rootPath, path)); err == nil {
				patterns = append(patterns, name)
				break
			}
		}
	}

	// State management - buscar USO REAL, no solo imports
	// Por ejemplo: "use zustand", "createStore", no solo la palabra "zustand"
	stateUsagePatterns := []struct {
		pattern string
		name    string
	}{
		{"use zustand", "zustand"},
		{"createStore", "zustand"},
		{"createSlice", "zustand"},
		{"configureStore", "redux"},
		{"createReducer", "redux"},
		{"useAtom", "jotai"},
		{"useRecoilValue", "recoil"},
		{"makeAutoObservable", "mobx"},
	}

	for _, sp := range stateUsagePatterns {
		if found := searchInProjectCode(rootPath, sp.pattern); found {
			patterns = append(patterns, "state:"+sp.name)
		}
	}

	return patterns
}

func detectAntiPatterns(rootPath string) []string {
	var antiPatterns []string

	// Buscar antipatrones comunes en el historial git
	cmd := exec.Command("git", "log", "--all", "--oneline", "-100")
	cmd.Dir = rootPath
	output, err := cmd.Output()
	if err != nil {
		return antiPatterns
	}

	history := string(output)

	// Detectar mensajes de fix comunes
	antiPatternKeywords := map[string][]string{
		"memory-leak":    {"memory leak", "leak", "cleanup"},
		"race-condition": {"race condition", "concurrent", "thread"},
		"security-fix":   {"security", "vulnerability", "xss", "injection"},
		"performance":    {"performance", "slow", "bottleneck", "optimize"},
		"bugfix":         {"fix", "fix:", "fixed"},
		"naming":         {"naming", "rename", "typo"},
	}

	for name, keywords := range antiPatternKeywords {
		count := 0
		for _, kw := range keywords {
			count += strings.Count(history, kw)
		}
		if count > 5 { // Threshold más alto para evitar ruido
			antiPatterns = append(antiPatterns, fmt.Sprintf("%s (%d mentions in commits)", name, count))
		}
	}

	//También buscar antipatrones reales en código
	codeAntiPatterns := detectCodeAntiPatterns(rootPath)
	antiPatterns = append(antiPatterns, codeAntiPatterns...)

	return antiPatterns
}

// detectCodeAntiPatterns busca antipatrones conocidos en el código real
func detectCodeAntiPatterns(rootPath string) []string {
	var antipatterns []string

	// Patrones de código que son anti-patterns conocidos
	codeSearches := []struct {
		pattern string
		name    string
	}{
		{"TODO", "TODO comments"},
		{"FIXME", "FIXME comments"},
		{"XXX", "XXX comments"},
		{"hack", "HACK comments"},
		{"console.log", "console.log statements"},
		{"print(", "print statements"},
		{"fmt.Print", "fmt.Print statements"},
		{"log.Fatal", "log.Fatal calls"},
		{"panic(", "panic calls"},
		{"TODO(", "TODO() function calls"},
		{"any", "any type usage (Go)"},
		{"interface{}", "empty interface"},
		{"// nolint", "nolint suppressions"},
		{"eslint-disable", "ESLint disables"},
		{"// @ts-ignore", "TS ignore comments"},
		{"// @ts-expect-error", "TS expect errors"},
		{"password =", "hardcoded password"},
		{"api_key =", "hardcoded api_key"},
		{"secret =", "hardcoded secret"},
	}

	counts := make(map[string]int)

	filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if isIgnoredDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".go" && ext != ".ts" && ext != ".js" && ext != ".py" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		strContent := string(content)
		for _, search := range codeSearches {
			counts[search.name] += strings.Count(strContent, search.pattern)
		}
		return nil
	})

	for _, search := range codeSearches {
		if counts[search.name] > 0 {
			antipatterns = append(antipatterns, fmt.Sprintf("%s (%d occurrences)", search.name, counts[search.name]))
		}
	}

	return antipatterns
}

func analyzeArchitecture(rootPath string) string {
	var parts []string

	// Estructura principal
	structures := []struct {
		dir  string
		desc string
	}{
		{"cmd/", "CLI application"},
		{"internal/", "Internal packages"},
		{"pkg/", "Public packages"},
		{"api/", "API layer"},
		{"services/", "Business logic"},
		{"models/", "Data models"},
		{"db/", "Database"},
		{"migrations/", "Database migrations"},
		{"scripts/", "Build scripts"},
		{"config/", "Configuration"},
		{"docs/", "Documentation"},
		{"test/", "Test utilities"},
	}

	for _, s := range structures {
		if _, err := os.Stat(filepath.Join(rootPath, s.dir)); err == nil {
			parts = append(parts, s.desc)
		}
	}

	if len(parts) == 0 {
		return "Monolithic or simple project structure"
	}

	return strings.Join(parts, " → ")
}

func countFiles(rootPath string) (int, map[string]int) {
	stats := make(map[string]int)
	count := 0

	extensions := map[string]string{
		".go":     "Go",
		".ts":     "TypeScript",
		".tsx":    "TypeScript React",
		".js":     "JavaScript",
		".jsx":    "React",
		".py":     "Python",
		".rb":     "Ruby",
		".java":   "Java",
		".kt":     "Kotlin",
		".rs":     "Rust",
		".c":      "C",
		".cpp":    "C++",
		".cs":     "C#",
		".php":    "PHP",
		".swift":  "Swift",
		".vue":    "Vue",
		".svelte": "Svelte",
		".yaml":   "YAML",
		".yml":    "YAML",
		".json":   "JSON",
		".md":     "Markdown",
	}

	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip common ignored dirs - INCLUYE plugins y skills
		if info.IsDir() && isIgnoredDir(info.Name()) {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			count++
			ext := filepath.Ext(info.Name())
			if lang, ok := extensions[ext]; ok {
				stats[lang]++
			}
		}
		return nil
	})

	return count, stats
}

func isIgnoredDir(dirName string) bool {
	skipDirs := []string{".git", "node_modules", "dist", "build", ".next", "coverage", ".cache", "vendor", ".opencode", "plugins", "skills", ".juar", ".agent"}
	for _, skip := range skipDirs {
		if dirName == skip {
			return true
		}
	}
	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func generateProjectSkills(rootPath string, info ProjectInfo, update bool) error {
	skillsDir := filepath.Join(rootPath, config.JuarDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio skills: %w", err)
	}

	// Generate project-context.md
	contextContent := fmt.Sprintf(`---
name: project-context
description: >
  Contexto específico de este proyecto: stack, arquitectura y decisiones técnicas.
  Se genera automáticamente con 'juarvis analyze'.
metadata:
  author: juarvis-auto
  version: "1.0"
  generated: true
---

# Contexto del Proyecto

## Stack Tecnológico
%s

## Arquitectura
%s

## Estadísticas
- Archivos: %d
- Lenguajes: %s
`, formatList(info.Stack), info.Architecture, info.FileCount, formatMap(info.LanguageStats))

	if err := os.WriteFile(filepath.Join(skillsDir, "project-context.md"), []byte(contextContent), 0644); err != nil {
		return err
	}

	// Generate conventions.md
	conventionsContent := fmt.Sprintf(`---
name: project-conventions
description: >
  Convenciones específicas de este proyecto.
  Se genera automáticamente con 'juarvis analyze'.
metadata:
  author: juarvis-auto
  version: "1.0"
  generated: true
---

# Convenciones del Proyecto

%s

## Patrones Utilizados
%s

## Antipatrones a Evitar
%s

> ⚠️ Estas convenciones se detectaron automáticamente. Verifica que son correctas.
`, formatList(info.Conventions), formatList(info.Patterns), formatList(info.AntiPatterns))

	if err := os.WriteFile(filepath.Join(skillsDir, "conventions.md"), []byte(conventionsContent), 0644); err != nil {
		return err
	}

	return nil
}

func formatList(items []string) string {
	if len(items) == 0 {
		return "- (ninguno detectado)"
	}
	var result string
	for _, item := range items {
		result += fmt.Sprintf("- %s\n", item)
	}
	return result
}

func formatMap(m map[string]int) string {
	if len(m) == 0 {
		return "N/A"
	}
	var result string
	for k, v := range m {
		result += fmt.Sprintf("%s: %d, ", k, v)
	}
	return strings.TrimSuffix(result, ", ")
}

// searchInProjectCode busca solo en código del proyecto (src/, cmd/, no en plugins/skills)
func searchInProjectCode(rootPath, pattern string) bool {
	found := false
	filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if found || err != nil {
			return err
		}
		if d.IsDir() {
			if isIgnoredDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		validExts := map[string]bool{".go": true, ".ts": true, ".tsx": true, ".js": true, ".jsx": true, ".py": true, ".rs": true}
		if !validExts[ext] {
			return nil
		}

		content, err := os.ReadFile(path)
		if err == nil && strings.Contains(string(content), pattern) {
			found = true
			return filepath.SkipDir
		}
		return nil
	})
	return found
}
