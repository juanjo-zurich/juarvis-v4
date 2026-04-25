package cmd

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"juarvis/pkg/config"
	"juarvis/pkg/output"
	"juarvis/pkg/root"
	"juarvis/pkg/watcher"

	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Vigila cambios en el proyecto y ejecuta reglas de seguridad automaticamente",
	Long: `Inicia un daemon que monitoriza el sistema de archivos del proyecto.
Cuando detecta cambios, evalua reglas de hookify automaticamente y crea
snapshots de seguridad si se detectan modificaciones masivas.

Esto proporciona un "cinturon de seguridad" automatico que no depende de
la voluntad del agente de IA.

Modos:
  juarvis watch              Modo interactivo (foreground)
  juarvis watch --daemon     Ejecuta en segundo plano (background)
  juarvis watch --stop       Detiene el watcher en segundo plano`,
	Run: func(cmd *cobra.Command, args []string) {
		stop, _ := cmd.Flags().GetBool("stop")
		if stop {
			stopWatcher()
			return
		}

		daemon, _ := cmd.Flags().GetBool("daemon")
		if daemon {
			startDaemon()
			return
		}

		runWatcherForeground(cmd)
	},
}

func runWatcherForeground(cmd *cobra.Command) {
	rootPath, err := root.GetRoot()
	if err != nil {
		output.Error("%v", err)
		os.Exit(1)
	}

	debounceMs, _ := cmd.Flags().GetInt("debounce-ms")
	threshold, _ := cmd.Flags().GetInt("auto-snapshot-threshold")
	noAutoSnapshot, _ := cmd.Flags().GetBool("no-auto-snapshot")
	verbose, _ := cmd.Flags().GetBool("verbose")

	cfg := watcher.DefaultWatcherConfig(rootPath)
	if debounceMs > 0 {
		cfg.DebounceMs = debounceMs
	}
	if threshold > 0 {
		cfg.AutoSnapshotThreshold = threshold
	}
	if noAutoSnapshot {
		cfg.NoAutoSnapshot = true
	}
	cfg.Verbose = verbose
	cfg.QuietMode = !verbose

	w, err := watcher.NewWatcher(cfg)
	if err != nil {
		output.Error("Error iniciando watcher: %v", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	if err := w.Start(ctx); err != nil {
		output.Error("Error en watcher: %v", err)
		os.Exit(1)
	}
}

func startDaemon() {
	rootPath, err := root.GetRoot()
	if err != nil {
		output.Error("%v", err)
		os.Exit(1)
	}

	exe, err := os.Executable()
	if err != nil {
		output.Error("No se pudo obtener el ejecutable: %v", err)
		os.Exit(1)
	}

	juarDir := filepath.Join(rootPath, ".juar")
	if err := os.MkdirAll(juarDir, 0755); err != nil {
		output.Error("No se pudo crear directorio .juar: %v", err)
		os.Exit(1)
	}
	pidFile := filepath.Join(juarDir, config.WatcherPIDFile)

	if _, err := os.Stat(pidFile); err == nil {
		pidBytes, readErr := os.ReadFile(pidFile)
		if readErr == nil {
			oldPID, parseErr := strconv.Atoi(string(pidBytes))
			if parseErr == nil {
				if proc, sigErr := os.FindProcess(oldPID); sigErr == nil {
					if killErr := proc.Signal(syscall.Signal(0)); killErr == nil {
						output.Info("Watcher ya esta corriendo en segundo plano (PID: %d)", oldPID)
						output.Info("Ejecuta 'juarvis watch --stop' para detenerlo")
						return
					}
				}
			}
		}
	}

	cmd := exec.Command(exe, "watch", "--foreground-child")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		output.Error("Error iniciando daemon: %v", err)
		os.Exit(1)
	}

	pid := cmd.Process.Pid
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		output.Warning("No se pudo escribir PID file: %v", err)
	}

	output.Success("Watcher started in background (PID: %d). Run 'juarvis watch --stop' to stop.", pid)
}

func stopWatcher() {
	rootPath, err := root.GetRoot()
	if err != nil {
		output.Error("%v", err)
		os.Exit(1)
	}

	pidFile := filepath.Join(rootPath, config.JuarDir, config.WatcherPIDFile)

	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		output.Info("No hay watcher en segundo plano (no se encontro PID file)")
		return
	}

	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		output.Error("No se pudo leer PID file: %v", err)
		os.Exit(1)
	}

	pid, err := strconv.Atoi(string(pidBytes))
	if err != nil {
		output.Error("PID file corrupto: %v", err)
		_ = os.Remove(pidFile)
		os.Exit(1)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		output.Info("Proceso %d no encontrado, limpiando PID file", pid)
		_ = os.Remove(pidFile)
		return
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		output.Info("Proceso %d ya no existe, limpiando PID file", pid)
		_ = os.Remove(pidFile)
		return
	}

	_ = os.Remove(pidFile)
	output.Success("Watcher detenido (PID: %d)", pid)
}

func init() {
	watchCmd.Flags().Int("debounce-ms", 500, "Ventana de debounce en milisegundos")
	watchCmd.Flags().Int("auto-snapshot-threshold", 5, "Numero de cambios para auto-snapshot")
	watchCmd.Flags().Bool("no-auto-snapshot", false, "Desactivar auto-snapshots")
	watchCmd.Flags().Bool("verbose", false, "Mostrar detalles de puntuación y filtrado de archivos")
	watchCmd.Flags().Bool("daemon", false, "Ejecutar watcher en segundo plano (background)")
	watchCmd.Flags().Bool("stop", false, "Detener el watcher en segundo plano")
	watchCmd.Flags().Bool("foreground-child", false, "Flag interno para proceso hijo del daemon")
	_ = watchCmd.Flags().MarkHidden("foreground-child")
	rootCmd.AddCommand(watchCmd)
}
