package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"juarvis/pkg/agents"
	"juarvis/pkg/output"
)

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Gestión de agentes definidos en AGENTS.md",
	Long:  `Comandos para listar, validar y usar agentes definidos en AGENTS.md.

El archivo AGENTS.md usa formato YAML-frontmatter y es compatible con:
- Cursor
- Windsurf
- Claude Code
- AntiGravity

Precedencia de configuración:
  AGENTS.md → juarvis.yaml → flags → defaults`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var agentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "Listar agentes definidos en AGENTS.md",
	Long:  `Lista todos los agentes definidos en AGENTS.md con sus detalles.

Muestra: nombre, descripción, skills, herramientas y modo de cada agente.`,
	RunE: runAgentsList,
}

var agentsValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validar estructura de AGENTS.md",
	Long:  `Valida que AGENTS.md tenga la estructura correcta.

Verifica:
- Versión del formato
- Agentes definidos
- Nombres únicos
- Modos válidos
- Herramientas y skills`,
	RunE: runAgentsValidate,
}

var agentsUseCmd = &cobra.Command{
	Use:   "use <agent>",
	Short: "Usar un agente específico",
	Long:  `Muestra la configuración de un agente específico para usar con herramientas externas.

Ejemplo:
  juarvis agents use developer
  juarvis agents use reviewer --json`,
	Args:      cobra.ExactArgs(1),
	RunE:      runAgentsUse,
	ValidArgs: listAgentNames(),
}

var agentsInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Crear AGENTS.md de ejemplo",
	Long:  `Crea un archivo AGENTS.md de ejemplo en la raíz del proyecto si no existe.`,
	RunE:   runAgentsInit,
}

var agentsShowCmd = &cobra.Command{
	Use:   "show [agent]",
	Short: "Mostrar configuración de agentes",
	Long:  `Muestra la configuración de AGENTS.md. Sin argumentos muestra todos los agentes.

Ejemplo:
  juarvis agents show
  juarvis agents show developer`,
	RunE: runAgentsShow,
}

func init() {
	agentsCmd.AddCommand(agentsListCmd)
	agentsCmd.AddCommand(agentsValidateCmd)
	agentsCmd.AddCommand(agentsUseCmd)
	agentsCmd.AddCommand(agentsInitCmd)
	agentsCmd.AddCommand(agentsShowCmd)

	rootCmd.AddCommand(agentsCmd)
}

func runAgentsList(cmd *cobra.Command, args []string) error {
	loader := agents.NewLoader()
	config, err := loader.LoadFromProject()
	if err != nil {
		if err == agents.ErrNoAgentsFile {
			output.Warning("No se encontró AGENTS.md")
			output.Info("Usa 'juarvis agents init' para crear uno de ejemplo")
			return nil
		}
		// Verificar si es error de formato inválido
		if strings.Contains(err.Error(), "YAML-frontmatter") || strings.Contains(err.Error(), "frontmatter") {
			output.Warning("AGENTS.md existe pero no tiene formato válido")
			output.Info("El archivo debe tener formato YAML-frontmatter (--- delimitadores)")
			output.Info("Usa 'juarvis agents init' para crear uno de ejemplo")
			return nil
		}
		return fmt.Errorf("error cargando AGENTS.md: %w", err)
	}

	if GlobalJSON {
		return printAgentsJSON(config)
	}

	return printAgentsTable(config)
}

func runAgentsValidate(cmd *cobra.Command, args []string) error {
	loader := agents.NewLoader()
	config, err := loader.LoadFromProject()
	if err != nil {
		if err == agents.ErrNoAgentsFile {
			output.Error("No se encontró AGENTS.md")
			return fmt.Errorf("archivo no encontrado")
		}
		return fmt.Errorf("error parseando AGENTS.md: %w", err)
	}

	result := config.Validate()

	if GlobalJSON {
		return printValidationResultJSON(result)
	}

	if result.Valid {
		output.Success("AGENTS.md es válido")
		output.Info("Versión: %s", config.Version)
		output.Info("Agentes definidos: %d", len(config.Agents))
	} else {
		output.Error("AGENTS.md tiene errores de validación:")
		for _, e := range result.Errors {
			output.Error("  - %s: %s", e.Field, e.Message)
		}
		return fmt.Errorf("validación fallida")
	}

	return nil
}

func runAgentsUse(cmd *cobra.Command, args []string) error {
	agentName := args[0]

	loader := agents.NewLoader()
	config, err := loader.LoadFromProject()
	if err != nil {
		if err == agents.ErrNoAgentsFile {
			return fmt.Errorf("no se encontró AGENTS.md")
		}
		return fmt.Errorf("error cargando AGENTS.md: %w", err)
	}

	agent := config.GetAgent(agentName)
	if agent == nil {
		output.Error("Agente '%s' no encontrado", agentName)
		output.Info("Usa 'juarvis agents list' para ver los agentes disponibles")
		return fmt.Errorf("agente no encontrado")
	}

	if GlobalJSON {
		return printAgentJSON(agent)
	}

	printAgentDetails(agent)
	return nil
}

func runAgentsInit(cmd *cobra.Command, args []string) error {
	// Intentar obtener la raíz del proyecto
	rootPath, err := getProjectRoot()
	if err != nil {
		output.Warning("No se detectó un ecosistema Juarvis")
		output.Info("Creando AGENTS.md en el directorio actual")

		// Crear en el directorio actual
		if err := agents.EnsureAgentsFile("."); err != nil {
			return fmt.Errorf("error creando AGENTS.md: %w", err)
		}
		output.Success("AGENTS.md creado en ./AGENTS.md")
		return nil
	}

	if err := agents.EnsureAgentsFile(rootPath); err != nil {
		return fmt.Errorf("error creando AGENTS.md: %w", err)
	}

	output.Success("AGENTS.md creado en %s/AGENTS.md", rootPath)
	return nil
}

func runAgentsShow(cmd *cobra.Command, args []string) error {
	loader := agents.NewLoader()
	config, err := loader.LoadFromProject()
	if err != nil {
		if err == agents.ErrNoAgentsFile {
			output.Warning("No se encontró AGENTS.md")
			output.Info("Usa 'juarvis agents init' para crear uno de ejemplo")
			return nil
		}
		return fmt.Errorf("error cargando AGENTS.md: %w", err)
	}

	if len(args) > 0 {
		agentName := args[0]
		agent := config.GetAgent(agentName)
		if agent == nil {
			return fmt.Errorf("agente '%s' no encontrado", agentName)
		}

		if GlobalJSON {
			return printAgentJSON(agent)
		}
		printAgentDetails(agent)
		return nil
	}

	if GlobalJSON {
		return printAgentsJSON(config)
	}

	return printAgentsTable(config)
}

func printAgentsJSON(config *agents.AgentsConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializando JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func printValidationResultJSON(result agents.ValidationResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializando JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func printAgentJSON(agent *agents.Agent) error {
	data, err := json.MarshalIndent(agent, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializando JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func printAgentsTable(config *agents.AgentsConfig) error {
	fmt.Println()
	fmt.Printf("  %-15s %-40s %-15s %s\n", "NOMBRE", "DESCRIPCIÓN", "MODO", "HERRAMIENTAS")
	fmt.Printf("  %s\n", "────────────────────────────────────────────────────────────────────────────────")

	for _, agent := range config.Agents {
		desc := agent.Description
		if len(desc) > 37 {
			desc = desc[:37] + "..."
		}

		mode := string(agent.Mode)
		if mode == "" {
			mode = "interactive"
		}

		tools := len(agent.Tools)
		if tools == 0 {
			tools = len(agent.Tools)
		}

		fmt.Printf("  %-15s %-40s %-15s %d\n",
			agent.Name,
			desc,
			mode,
			tools,
		)
	}

	fmt.Println()
	fmt.Printf("  Total: %d agente(s)\n", len(config.Agents))
	fmt.Printf("  Versión: %s\n", config.Version)
	return nil
}

func printAgentDetails(agent *agents.Agent) error {
	fmt.Println()
	fmt.Printf("  Nombre:        %s\n", agent.Name)
	fmt.Printf("  Descripción:  %s\n", agent.Description)

	if len(agent.Skills) > 0 {
		fmt.Printf("  Skills:        %s\n", joinStrings(agent.Skills))
	} else {
		fmt.Printf("  Skills:        (ninguno)\n")
	}

	if len(agent.Tools) > 0 {
		fmt.Printf("  Herramientas:  %s\n", joinStrings(agent.Tools))
	} else {
		fmt.Printf("  Herramientas:  (todas)\n")
	}

	mode := string(agent.Mode)
	if mode == "" {
		mode = "interactive"
	}
	fmt.Printf("  Modo:         %s\n", mode)

	if agent.MaxTurns > 0 {
		fmt.Printf("  Max Turns:    %d\n", agent.MaxTurns)
	}

	if len(agent.Permissions) > 0 {
		fmt.Println()
		fmt.Println("  Permisos:")
		for _, p := range agent.Permissions {
			fmt.Printf("    - %s: %s\n", p.Name, p.Permission)
		}
	}

	if len(agent.Env) > 0 {
		fmt.Println()
		fmt.Println("  Variables de entorno:")
		for k, v := range agent.Env {
			fmt.Printf("    - %s=%s\n", k, v)
		}
	}

	fmt.Println()
	return nil
}

func joinStrings(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	result := ss[0]
	for i := 1; i < len(ss); i++ {
		result += ", " + ss[i]
	}
	return result
}

func listAgentNames() []string {
	loader := agents.NewLoader()
	config, err := loader.TryLoadAgents()
	if err != nil {
		return []string{}
	}
	return config.GetAgentNames()
}

func getProjectRoot() (string, error) {
	// Intentar usar root del entorno o detectar automáticamente
	if GlobalRoot != "" {
		return GlobalRoot, nil
	}

	// Intentar detectar desde el directorio actual
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Buscar hacia arriba hasta encontrar indicadores de proyecto
	for {
		// Verificar si es un proyecto Juarvis
		indicators := []string{"AGENTS.md", "juarvis.yaml", ".juar"}
		found := false
		for _, ind := range indicators {
			if _, err := os.Stat(ind); err == nil {
				found = true
				break
			}
		}
		if found {
			return pwd, nil
		}

		// Subir un nivel
		parent := fmt.Sprintf("..%c", os.PathSeparator)
		if parent == pwd+string(os.PathSeparator)+".." {
			break
		}
		pwd = fmt.Sprintf("%s%c..", pwd, os.PathSeparator)
	}

	return "", fmt.Errorf("no se encontró raíz del proyecto")
}