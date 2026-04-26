package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"juarvis/pkg/artifacts"
	"juarvis/pkg/output"
	"juarvis/pkg/root"
)

var (
	artifactsFilter string
	artifactsTag   string
	artifactsJSON  bool
	artifactsTags []string
)

var artifactsCmd = &cobra.Command{
	Use:   "artifacts",
	Short: "Gestión de artifacts verificables del ecosistema",
	Long: `Sistema de artifacts para generar verificables tangibles que construyen 
confianza y transparencia. Cada artifact es un registro verificable de una 
operación o resultado del sistema.`,
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := root.GetRoot()
		if err != nil {
			output.Fatal(output.ExitNoEcosystem,
				"Ejecuta 'juarvis init' primero",
				"No se detectó un ecosistema Juarvis en este directorio")
		}
		manager := artifacts.NewManager(rootPath)
		listArtifacts(rootPath, manager, false)
	},
}

func init() {
	rootCmd.AddCommand(artifactsCmd)
}

func listArtifacts(rootPath string, manager *artifacts.Manager, showJSON bool) {
	var filterType artifacts.ArtifactType
	if artifactsFilter != "" {
		var err error
		filterType, err = artifacts.ParseArtifactType(artifactsFilter)
		if err != nil {
			output.Warning("Filtro ignorado: %v", err)
		}
	}
	artifactList, err := manager.List(filterType)
	if err != nil {
		output.Error("Error listando artifacts: %v", err)
		return
	}
	if len(artifactList) == 0 {
		output.Info("No se encontraron artifacts")
		if output.IsJSONMode() {
			output.PrintJSON([]interface{}{})
		}
		return
	}
	if output.IsJSONMode() || showJSON || artifactsJSON {
		output.PrintJSON(artifactList)
		return
	}
	output.Info("%d artifacts encontrados:", len(artifactList))
	fmt.Println()
	for _, a := range artifactList {
		tags := ""
		if len(a.Tags) > 0 {
			tags = fmt.Sprintf(" [%s]", strings.Join(a.Tags, ", "))
		}
		status := output.Styled("cyan", "%s", a.Type.String())
		fmt.Printf("  %s %s%s\n", status, a.ID[:8], tags)
		fmt.Printf("     → %s\n", a.Summary)
		fmt.Printf("     %s\n", a.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}
}

var artifactsListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista todos los artifacts",
	Long: `Lista artifacts con filtros opcionales por tipo o tag.
Usa --type para filtrar por tipo de artifact.
Usa --tag para filtrar por tag.`,
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := root.GetRoot()
		if err != nil {
			output.Fatal(output.ExitNoEcosystem,
				"Ejecuta 'juarvis init' primero",
				"No se detectó un ecosistema Juarvis")
		}
		manager := artifacts.NewManager(rootPath)

		if artifactsTag != "" {
			listByTag(rootPath, manager, artifactsTag)
			return
		}
		listArtifacts(rootPath, manager, false)
	},
}

func init() {
	artifactsListCmd.Flags().StringVar(&artifactsFilter, "type", "", "Filtrar por tipo (task_list, implementation_plan, screenshot, test_result, verification_report, code_diff)")
	artifactsListCmd.Flags().StringVar(&artifactsTag, "tag", "", "Filtrar por tag")
	artifactsListCmd.Flags().BoolVarP(&artifactsJSON, "json", "j", false, "Salida en JSON")
	artifactsCmd.AddCommand(artifactsListCmd)
}

func listByTag(rootPath string, manager *artifacts.Manager, tag string) {
	results, err := manager.FindByTag(tag)
	if err != nil {
		output.Error("Error buscando por tag: %v", err)
		return
	}
	if len(results) == 0 {
		output.Info("No se encontraron artifacts con tag: %s", tag)
		return
	}
	output.Info("%d artifacts encontrados con tag '%s':", len(results), tag)
	fmt.Println()
	for _, a := range results {
		fmt.Printf("  %s %s\n", a.ID[:8], a.Type)
		fmt.Printf("     → %s\n", a.Summary)
	}
}

var artifactsGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Obtiene los detalles de un artifact",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := root.GetRoot()
		if err != nil {
			output.Fatal(output.ExitNoEcosystem,
				"Ejecuta 'juarvis init' primero",
				"No se detectó un ecosistema Juarvis")
		}
		manager := artifacts.NewManager(rootPath)
		id := args[0]
		artifact, err := manager.Get(id)
		if err != nil {
			output.Error("Artifact no encontrado: %s", id)
			return
		}
		if output.IsJSONMode() || artifactsJSON {
			output.PrintJSON(artifact)
		} else {
			printArtifact(artifact)
		}
	},
}

func init() {
	artifactsGetCmd.Flags().BoolVarP(&artifactsJSON, "json", "j", false, "Salida en JSON")
	artifactsCmd.AddCommand(artifactsGetCmd)
}

func printArtifact(a interface{}) {
	switch ar := a.(type) {
	case *artifacts.TaskListArtifact:
		fmt.Printf("TaskList: %s\n", ar.Title)
		fmt.Printf("ID: %s\n", ar.ID)
		fmt.Printf("Timestamp: %s\n", ar.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("Descripción: %s\n", ar.Description)
		fmt.Printf("Tareas: %d\n", len(ar.Tasks))
		for i, t := range ar.Tasks {
			status := output.Styled("green", "✓")
			if t.Status != artifacts.TaskCompleted {
				status = output.Styled("yellow", "○")
			}
			fmt.Printf("  %d. %s %s [%s]\n", i+1, status, t.Title, t.Status)
		}
	case *artifacts.ImplementationPlanArtifact:
		fmt.Printf("ImplementationPlan: %s\n", ar.Title)
		fmt.Printf("ID: %s\n", ar.ID)
		fmt.Printf("Timestamp: %s\n", ar.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("Objetivo: %s\n", ar.Goal)
		fmt.Printf("Pasos: %d\n", len(ar.Steps))
		for _, s := range ar.Steps {
			status := output.Styled("green", "✓")
			if !s.Completed {
				status = output.Styled("yellow", "○")
			}
			fmt.Printf("  %d. %s %s\n", s.StepNumber, status, s.Title)
			fmt.Printf("     %s\n", s.Description)
		}
	case *artifacts.ScreenshotArtifact:
		fmt.Printf("Screenshot: %dx%d %s\n", ar.Width, ar.Height, ar.Format)
		fmt.Printf("ID: %s\n", ar.ID)
		fmt.Printf("Timestamp: %s\n", ar.Timestamp.Format("2006-01-02 15:04:05"))
		if ar.URL != "" {
			fmt.Printf("URL: %s\n", ar.URL)
		}
		fmt.Printf("Contenido (base64): %d bytes\n", len(ar.Content))
	case *artifacts.TestResultArtifact:
		fmt.Printf("TestResult\n")
		fmt.Printf("ID: %s\n", ar.ID)
		fmt.Printf("Timestamp: %s\n", ar.Timestamp_.Format("2006-01-02 15:04:05"))
		passedColor := "green"
		if ar.Failed > 0 {
			passedColor = "red"
		}
		status := output.Styled(passedColor, "%d/%d passed", ar.Passed, ar.TotalTests)
		fmt.Printf("Tests: %s\n", status)
		fmt.Printf("Skipped: %d, Failed: %d\n", ar.Skipped, ar.Failed)
		if ar.Coverage > 0 {
			fmt.Printf("Coverage: %.1f%%\n", ar.Coverage)
		}
	case *artifacts.VerificationReportArtifact:
		fmt.Printf("VerificationReport\n")
		fmt.Printf("ID: %s\n", ar.ID)
		fmt.Printf("Timestamp: %s\n", ar.Timestamp.Format("2006-01-02 15:04:05"))
		status := output.Styled("green", "PASSED")
		if !ar.Passed {
			status = output.Styled("red", "FAILED")
		}
		fmt.Printf("Status: %s\n", status)
		fmt.Printf("Checks: %d/%d passed\n", ar.PassedCount, ar.Total)
		for _, c := range ar.Checks {
			checkStatus := output.Styled("green", "✓")
			if !c.Passed {
				checkStatus = output.Styled("red", "✗")
			}
			fmt.Printf("  %s %s\n", checkStatus, c.Name)
			if c.Message != "" {
				fmt.Printf("     %s\n", c.Message)
			}
		}
		if ar.Duration != "" {
			fmt.Printf("Duración: %s\n", ar.Duration)
		}
	case *artifacts.CodeDiffArtifact:
		fmt.Printf("CodeDiff: %s → %s\n", ar.BaseBranch, ar.HeadBranch)
		fmt.Printf("ID: %s\n", ar.ID)
		fmt.Printf("Timestamp: %s\n", ar.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("Archivos: %d\n", len(ar.Files))
		fmt.Printf("Cambios: +%d / -%d\n", ar.Stats.Insertions, ar.Stats.Deletions)
		for _, f := range ar.Files {
			modeColor := "cyan"
			if f.Mode == "added" {
				modeColor = "green"
			} else if f.Mode == "deleted" {
				modeColor = "red"
			}
			mode := output.Styled(modeColor, "%s", f.Mode)
			fmt.Printf("  %s %s\n", mode, f.NewPath)
		}
	default:
		fmt.Printf("Tipo de artifact desconocido\n")
	}
}

var artifactsDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Elimina un artifact",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := root.GetRoot()
		if err != nil {
			output.Fatal(output.ExitNoEcosystem,
				"Ejecuta 'juarvis init' primero",
				"No se detectó un ecosistema Juarvis")
		}
		manager := artifacts.NewManager(rootPath)
		id := args[0]
		if err := manager.Delete(id); err != nil {
			output.Error("Error eliminando artifact: %v", err)
			return
		}
		output.Success("Artifact eliminado: %s", id[:8])
	},
}

func init() {
	artifactsCmd.AddCommand(artifactsDeleteCmd)
}

var artifactsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Crea un nuevo artifact",
}

var artifactsCreateTaskListCmd = &cobra.Command{
	Use:   "task-list [title]",
	Short: "Crea un nuevo TaskList",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		title := args[0]
		tasks := []artifacts.Task{
			{
				ID:          filepath.Base(os.Args[0]) + "-" + fmt.Sprint(len(args)),
				Title:       "Tarea inicial",
				Status:      artifacts.TaskPending,
				CreatedAt:  artifacts.NewBase(artifacts.TaskList).Timestamp,
				UpdatedAt:  artifacts.NewBase(artifacts.TaskList).Timestamp,
			},
		}
		artifact := artifacts.NewTaskListArtifact(title, "Creado desde CLI", tasks)
		rootPath, _ := root.GetRoot()
		manager := artifacts.NewManager(rootPath)
		if err := manager.SaveTaskList(artifact); err != nil {
			output.Error("Error guardando artifact: %v", err)
			return
		}
		output.Success("TaskList creado: %s", artifact.ID[:8])
	},
}

func init() {
	var tags []string
	artifactsCreateTaskListCmd.Flags().StringSliceVarP(&tags, "tag", "t", []string{}, "Tags para el artifact")
	artifactsCreateCmd.AddCommand(artifactsCreateTaskListCmd)
	artifactsCmd.AddCommand(artifactsCreateCmd)
}