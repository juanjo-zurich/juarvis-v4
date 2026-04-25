package cmd

import (
	"juarvis/pkg/output"
	"juarvis/pkg/scheduler"

	"github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Gestión de jobs programados",
	Long:  `Permite crear, listar, ejecutar y eliminar jobs programados usando expresiones cron.`,
}

// addJobCmd: juarvis schedule add --name --schedule --prompt [--agent]
var addJobCmd = &cobra.Command{
	Use:   "add",
	Short: "Crea un nuevo job programado",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		schedule, _ := cmd.Flags().GetString("schedule")
		prompt, _ := cmd.Flags().GetString("prompt")
		agent, _ := cmd.Flags().GetString("agent")
		timeout, _ := cmd.Flags().GetInt("timeout")
		workdir, _ := cmd.Flags().GetString("workdir")

		if name == "" {
			output.Fatal(output.ExitGeneric,
				"Usa --name para especificar el nombre del job",
				"falta parametro --name")
		}
		if schedule == "" {
			output.Fatal(output.ExitGeneric,
				"Usa --schedule para especificar el horario (ej: '0 9 * * *')",
				"falta parametro --schedule")
		}
		if prompt == "" {
			output.Fatal(output.ExitGeneric,
				"Usa --prompt para especificar el prompt a ejecutar",
				"falta parametro --prompt")
		}

		// Validar expresión cron
		if err := scheduler.ValidateCronExpression(schedule); err != nil {
			output.Fatal(output.ExitConfigError,
				"Verifica el formato: '0 9 * * *' = diario a las 9am",
				"cron invalido: %v", err)
		}

		// Crear job
		job := scheduler.NewJob(name, schedule, prompt, agent)
		job.Timeout = timeout
		job.Workdir = workdir

		if err := scheduler.SaveJob(job); err != nil {
			output.Fatal(output.ExitGeneric,
				"verifica que juarvis init fue ejecutado",
				"error guardando job: %v", err)
		}

		output.Success("Job '%s' creado exitosamente", name)
		output.Info("Schedule: %s (%s)", schedule, scheduler.FormatSchedule(schedule))
	},
}

// listJobsCmd: juarvis schedule list
var listJobsCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista todos los jobs programados",
	Run: func(cmd *cobra.Command, args []string) {
		headers, rows, err := scheduler.ListJobsFormatted()
		if err != nil {
			output.Fatal(output.ExitGeneric,
				"verifica que juarvis init fue ejecutado",
				"error listando jobs: %v", err)
		}

		if len(rows) == 0 {
			output.Info("No hay jobs programados. Usa 'juarvis schedule add --help' para crear uno.")
			return
		}

		output.Info("%d job(s) encontrado(s):", len(rows))
		output.PrintTable(headers, rows)
	},
}

// runJobCmd: juarvis schedule run --name
var runJobCmd = &cobra.Command{
	Use:   "run",
	Short: "Ejecuta un job inmediatamente",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			output.Fatal(output.ExitGeneric,
				"Usa --name para especificar el job",
				"falta parametro --name")
		}

		if err := scheduler.RunJob(name); err != nil {
			output.Fatal(output.ExitGeneric,
				"verifica que el job existe con 'juarvis schedule list'",
				"error ejecutando job: %v", err)
		}
	},
}

// deleteJobCmd: juarvis schedule delete --name
var deleteJobCmd = &cobra.Command{
	Use:   "delete",
	Short: "Elimina un job programado",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			output.Fatal(output.ExitGeneric,
				"Usa --name para especificar el job",
				"falta parametro --name")
		}

		if err := scheduler.DeleteJob(name); err != nil {
			output.Fatal(output.ExitGeneric,
				"verifica que el job existe con 'juarvis schedule list'",
				"error eliminando job: %v", err)
		}

		output.Success("Job '%s' eliminado", name)
	},
}

// enableJobCmd: juarvis schedule enable --name
var enableJobCmd = &cobra.Command{
	Use:   "enable",
	Short: "Activa un job pausado",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			output.Fatal(output.ExitGeneric,
				"Usa --name para especificar el job",
				"falta parametro --name")
		}

		job, err := scheduler.LoadJob(name)
		if err != nil {
			output.Fatal(output.ExitGeneric,
				"verifica que el job existe con 'juarvis schedule list'",
				"job no encontrado: %v", err)
		}

		job.Enabled = true
		if err := scheduler.SaveJob(job); err != nil {
			output.Fatal(output.ExitGeneric,
				"error guardando job",
				"%v", err)
		}

		output.Success("Job '%s' activado", name)
	},
}

// disableJobCmd: juarvis schedule disable --name
var disableJobCmd = &cobra.Command{
	Use:   "disable",
	Short: "Pausa un job activo",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			output.Fatal(output.ExitGeneric,
				"Usa --name para especificar el job",
				"falta parametro --name")
		}

		job, err := scheduler.LoadJob(name)
		if err != nil {
			output.Fatal(output.ExitGeneric,
				"verifica que el job existe con 'juarvis schedule list'",
				"job no encontrado: %v", err)
		}

		job.Enabled = false
		if err := scheduler.SaveJob(job); err != nil {
			output.Fatal(output.ExitGeneric,
				"error guardando job",
				"%v", err)
		}

		output.Success("Job '%s' pausado", name)
	},
}

func init() {
	// Flags compartidos
	addJobCmd.Flags().StringP("name", "n", "", "nombre del job")
	addJobCmd.Flags().String("schedule", "", "expresion cron (ej: '0 9 * * *')")
	addJobCmd.Flags().String("prompt", "", "prompt a ejecutar")
	addJobCmd.Flags().StringP("agent", "a", "juarvis", "agente a usar (opencode, claude, juarvis)")
	addJobCmd.Flags().Int("timeout", 3600, "timeout en segundos")
	addJobCmd.Flags().String("workdir", ".", "directorio de trabajo")

	runJobCmd.Flags().StringP("name", "n", "", "nombre del job a ejecutar")
	deleteJobCmd.Flags().StringP("name", "n", "", "nombre del job a eliminar")
	enableJobCmd.Flags().StringP("name", "n", "", "nombre del job a activar")
	disableJobCmd.Flags().StringP("name", "n", "", "nombre del job a pausar")

	// Registrar subcomandos
	scheduleCmd.AddCommand(addJobCmd)
	scheduleCmd.AddCommand(listJobsCmd)
	scheduleCmd.AddCommand(runJobCmd)
	scheduleCmd.AddCommand(deleteJobCmd)
	scheduleCmd.AddCommand(enableJobCmd)
	scheduleCmd.AddCommand(disableJobCmd)
	rootCmd.AddCommand(scheduleCmd)
}
