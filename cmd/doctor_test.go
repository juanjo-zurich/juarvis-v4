package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"juarvis/pkg/config"
)

func TestDoctorCommand(t *testing.T) {
	// Verificar que el comando está registrado
	if rootCmd == nil {
		t.Fatal("rootCmd es nil")
	}

	// Buscar doctor en subcomandos
	var found bool
	for _, c := range rootCmd.Commands() {
		if c.Use == "doctor" {
			found = true
			break
		}
	}
	if !found {
		t.Error("comando doctor no registrado en rootCmd")
	}
}

func TestDoctorCheckJuarvisPath(t *testing.T) {
	// Crear directorio temporal con juarvis
	tmpDir := t.TempDir()
	juarDir := filepath.Join(tmpDir, ".juarvis")
	os.MkdirAll(juarDir, 0755)

	// Verificar que .juarvis existe
	info, err := os.Stat(juarDir)
	if err != nil {
		t.Fatalf("error stating .juarvis: %v", err)
	}
	if !info.IsDir() {
		t.Error(".juarvis no es un directorio")
	}
}

func TestDoctorCheckConfig(t *testing.T) {
	tmpDir := t.TempDir()
	juarDir := filepath.Join(tmpDir, ".juarvis")
	os.MkdirAll(juarDir, 0755)

	// Verificar que LoadOrCreate funciona
	cfg, err := config.LoadOrCreate(tmpDir)
	if err != nil {
		t.Fatalf("LoadOrCreate falló: %v", err)
	}
	if cfg == nil {
		t.Fatal("cfg nil")
	}
}
