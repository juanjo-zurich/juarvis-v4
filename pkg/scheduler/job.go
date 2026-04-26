package scheduler

import (
	"time"

	"gopkg.in/yaml.v3"
)

// Job representa una tarea programada
type Job struct {
	Name     string    `yaml:"name"`
	Schedule string    `yaml:"schedule"` // formato cron: "0 9 * * *" = daily at 9am
	Prompt   string    `yaml:"prompt"`
	Agent    string    `yaml:"agent"` // "opencode", "claude", etc.
	Enabled  bool      `yaml:"enabled"`
	LastRun  time.Time `yaml:"last_run"`
	Workdir  string    `yaml:"workdir"`
	Timeout  int       `yaml:"timeout"` // segundos
}

// JobFile representa el archivo YAML de almacenamiento de jobs
type JobFile struct {
	Jobs []Job `yaml:"jobs"`
}

// NewJob crea un nuevo job con valores por defecto
func NewJob(name, schedule, prompt, agent string) *Job {
	return &Job{
		Name:     name,
		Schedule: schedule,
		Prompt:   prompt,
		Agent:    agent,
		Enabled:  true,
		Workdir:  ".",
		Timeout:  3600, // 1 hora por defecto
	}
}

// ToYAML serializa el job a YAML
func (j *Job) ToYAML() ([]byte, error) {
	return yaml.Marshal(j)
}

// FromYAML deserializa un job desde YAML
func (j *Job) FromYAML(data []byte) error {
	return yaml.Unmarshal(data, j)
}

// FormatLastRun devuelve una representación legible del último execution
func (j *Job) FormatLastRun() string {
	if j.LastRun.IsZero() {
		return "nunca"
	}
	return j.LastRun.Format("2006-01-02 15:04:05")
}

// FormatSchedule devuelve una descripción legible del schedule
func FormatSchedule(cronExpr string) string {
	// Parser simple para formatos comunes de cron
	switch cronExpr {
	case "0 9 * * *":
		return "diario a las 9:00"
	case "0 * * * *":
		return "cada hora"
	case "0 0 * * *":
		return "a medianoche"
	case "*/5 * * * *":
		return "cada 5 minutos"
	case "*/15 * * * *":
		return "cada 15 minutos"
	case "*/30 * * * *":
		return "cada 30 minutos"
	case "0 6 * * *":
		return "diario a las 6:00"
	case "0 12 * * *":
		return "diario a las 12:00"
	case "0 18 * * *":
		return "diario a las 18:00"
	default:
		return cronExpr
	}
}
