package scheduler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/adhocore/gronx"
	"juarvis/pkg/output"
	"juarvis/pkg/root"

	"gopkg.in/yaml.v3"
)

// jobsDir es el directorio donde se almacenan los jobs
const jobsDir = ".juar/jobs"

// Cron regex para validación básica (5 campos: min hour day month weekday)
var cronRegex = regexp.MustCompile(`^(\*|[0-9,\-\*]+)\s+(\*|[0-9,\-\*]+)\s+(\*|[0-9,\-\*]+)\s+(\*|[0-9,\-\*]+)\s+(\*|[0-9,\-\*]+)$`)

// LoadAllJobs carga todos los jobs desde .juar/jobs/
func LoadAllJobs() ([]Job, error) {
	juarRoot, err := root.GetRoot()
	if err != nil {
		return nil, fmt.Errorf("no se encontro ecosistema: %w", err)
	}

	jobsPath := filepath.Join(juarRoot, jobsDir)
	if err := os.MkdirAll(jobsPath, 0755); err != nil {
		return nil, fmt.Errorf("creando directorio de jobs: %w", err)
	}

	entries, err := os.ReadDir(jobsPath)
	if err != nil {
		return nil, fmt.Errorf("leyendo directorio de jobs: %w", err)
	}

	var jobs []Job
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(jobsPath, entry.Name()))
		if err != nil {
			output.Warning("saltando job corrupto: %s", entry.Name())
			continue
		}

		var job Job
		if err := yaml.Unmarshal(data, &job); err != nil {
			output.Warning("saltando job corrupto: %s", entry.Name())
			continue
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// LoadJob carga un job específico por nombre
func LoadJob(name string) (*Job, error) {
	jobs, err := LoadAllJobs()
	if err != nil {
		return nil, err
	}

	for _, job := range jobs {
		if job.Name == name {
			return &job, nil
		}
	}

	return nil, fmt.Errorf("job no encontrado: %s", name)
}

// SaveJob guarda un job en el sistema de archivos
func SaveJob(job *Job) error {
	juarRoot, err := root.GetRoot()
	if err != nil {
		return fmt.Errorf("no se encontro ecosistema: %w", err)
	}

	jobsPath := filepath.Join(juarRoot, jobsDir)
	if err := os.MkdirAll(jobsPath, 0755); err != nil {
		return fmt.Errorf("creando directorio de jobs: %w", err)
	}

	filename := sanitizeFilename(job.Name) + ".yaml"
	filePath := filepath.Join(jobsPath, filename)

	data, err := yaml.Marshal(job)
	if err != nil {
		return fmt.Errorf("serializando job: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("escribiendo job: %w", err)
	}

	return nil
}

// DeleteJob elimina un job por nombre
func DeleteJob(name string) error {
	juarRoot, err := root.GetRoot()
	if err != nil {
		return fmt.Errorf("no se encontro ecosistema: %w", err)
	}

	filename := sanitizeFilename(name) + ".yaml"
	filePath := filepath.Join(juarRoot, jobsDir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("job no encontrado: %s", name)
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("eliminando job: %w", err)
	}

	return nil
}

// ValidateCronExpression valida que una expresión cron sea válida
func ValidateCronExpression(expr string) error {
	if !cronRegex.MatchString(expr) {
		return fmt.Errorf("formato cron invalido: %s", expr)
	}
	return nil
}

// IsDue verifica si un job debe ejecutarse según su schedule cron
func IsDue(job *Job) (bool, error) {
	if !job.Enabled {
		return false, nil
	}

	if err := ValidateCronExpression(job.Schedule); err != nil {
		return false, err
	}

	// Si nunca se ha ejecutado, está debido
	if job.LastRun.IsZero() {
		return true, nil
	}

	// Usar gronx paraparsing cron real (sin argumento de tiempo usa time.Now())
	gron := gronx.New()
	due, _ := gron.IsDue(job.Schedule)
	return due, nil
}

// RunJob ejecuta un job inmediatamente
func RunJob(name string) error {
	job, err := LoadJob(name)
	if err != nil {
		return err
	}

	output.Info("Ejecutando job: %s", job.Name)

	// Determinar workdir
	workdir := job.Workdir
	if workdir == "" || workdir == "." {
		workdir, err = root.GetRoot()
		if err != nil {
			workdir, _ = os.Getwd()
		}
	}

	// Construir comando según el agente
	var cmd *exec.Cmd
	switch strings.ToLower(job.Agent) {
	case "opencode":
		cmd = exec.Command("opencode", "--prompt", job.Prompt)
	case "claude":
		cmd = exec.Command("claude", "--print", job.Prompt)
	case "juarvis", "":
		cmd = exec.Command("juarvis", "ask", job.Prompt)
	default:
		return fmt.Errorf("agente no soportado: %s", job.Agent)
	}

	cmd.Dir = workdir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Timeout
	timeout := time.Duration(job.Timeout) * time.Second
	if timeout == 0 {
		timeout = time.Hour
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err = <-done:
		// completado
	case <-time.After(timeout):
		cmd.Process.Kill()
		err = fmt.Errorf("timeout excedido (%d segundos)", job.Timeout)
	}

	// Actualizar LastRun incluso si hay error
	job.LastRun = time.Now()
	if saveErr := SaveJob(job); saveErr != nil {
		output.Warning("no se pudo actualizar last_run: %v", saveErr)
	}

	if err != nil {
		return err
	}

	output.Success("Job %s completado", job.Name)
	return nil
}

// sanitizeFilename limpia un nombre para usarlo como filename
func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, "..", "")
	name = strings.TrimSpace(name)
	if name == "" {
		name = "unnamed"
	}
	return name
}

// ListJobsFormatted retorna jobs formateados para mostrar
func ListJobsFormatted() ([]string, [][]string, error) {
	jobs, err := LoadAllJobs()
	if err != nil {
		return nil, nil, err
	}

	headers := []string{"NOMBRE", "AGENTE", "SCHEDULE", "ULTIMA EJECUCION", "ESTADO"}
	var rows [][]string

	for _, job := range jobs {
		status := "activo"
		if !job.Enabled {
			status = "pausado"
		}

		rows = append(rows, []string{
			job.Name,
			job.Agent,
			FormatSchedule(job.Schedule),
			job.FormatLastRun(),
			status,
		})
	}

	return headers, rows, nil
}
