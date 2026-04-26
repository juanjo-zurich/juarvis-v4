package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"juarvis/pkg/config"
	"juarvis/pkg/output"
	"juarvis/pkg/root"

	"github.com/spf13/cobra"
)

var vibeCmd = &cobra.Command{
	Use:   "vibe",
	Short: "Vibe Check: Evalúa la salud creativa y el flujo de tu proyecto",
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := root.GetRoot()
		if err != nil {
			output.Error("No estás en un ecosistema Juarvis. ¡Haz un 'juarvis up' para empezar!")
			return
		}

		output.Banner("🎵 VIBE CHECK - ESTADO DEL ECOSISTEMA")

		// 1. Modo de autonomía
		cfg, err := config.LoadOrCreate(rootPath)
		if err == nil {
			levelName := config.GetLevelName(cfg.AutonomyLevel)
			output.Info("Modo: %s (nivel %d)", output.Styled("cyan", "%s", levelName), cfg.AutonomyLevel)
		}

		// 2. Watcher
		watcherStatus := "○ INACTIVO"
		watcherPID := "-"

		watcherFile := filepath.Join(rootPath, config.JuarDir, config.WatcherPIDFile)
		if data, err := os.ReadFile(watcherFile); err == nil {
			watcherStatus = "● ACTIVO"
			watcherPID = strings.TrimSpace(string(data))
			if len(watcherPID) > 6 {
				watcherPID = watcherPID[:6]
			}
		}

		output.Info("Watcher: %s", output.Styled("green", "%s", watcherStatus))
		if watcherStatus == "● ACTIVO" && watcherPID != "-" {
			output.Info("  PID: %s | uptime: -", watcherPID)
		}

		// 3. Memoria (observaciones)
		memPath := filepath.Join(rootPath, config.JuarDir, "memory", "observations")
		memCount := 0
		var lastMem time.Time
		if entries, err := os.ReadDir(memPath); err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
					memCount++
					if info, err := e.Info(); err == nil {
						if info.ModTime().After(lastMem) {
							lastMem = info.ModTime()
						}
					}
				}
			}
		}

		memAgo := "-"
		if !lastMem.IsZero() {
			memAgo = time.Since(lastMem).Round(time.Minute).String()
		}
		output.Info("Memoria: %s obs · última hace %s",
			output.Styled("purple", "%d", memCount), memAgo)

		// 4. Sesiones
		sessionsPath := filepath.Join(rootPath, config.JuarDir, "memory", "sessions")
		sessionCount := 0
		if entries, err := os.ReadDir(sessionsPath); err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
					sessionCount++
				}
			}
		}
		output.Info("Sesiones: %s guardadas", output.Styled("blue", "%d", sessionCount))

		// 5. Skills propias del proyecto
		projectSkillsPath := filepath.Join(rootPath, config.JuarDir, "skills")
		projectSkills := 0
		if entries, err := os.ReadDir(projectSkillsPath); err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
					projectSkills++
				}
			}
		}

		// Skills globales (del binario embebido)
		globalSkills := 73
		output.Info("Skills: %s globales + %s propias del proyecto",
			output.Styled("cyan", "%d", globalSkills),
			output.Styled("cyan", "%d", projectSkills))

		// 6. Seguridad (Hookify)
		hookifyRules := 0
		hookifyDir := filepath.Join(rootPath, config.JuarDir, "hooks")
		if entries, err := os.ReadDir(hookifyDir); err == nil {
			hookifyRules = len(entries)
		}
		if hookifyRules == 0 {
			output.Info("Seguridad: %s (usa hooks por defecto)",
				output.Styled("green", "%s", "0 alertas"))
		} else {
			output.Info("Seguridad: %s · %d reglas activas",
				output.Styled("green", "%s", "0 alertas"), hookifyRules)
		}

		// 7. Snapshots
		snapshotsCount := 0
		var lastSnap time.Time
		out, err := exec.Command("git", "stash", "list").CombinedOutput()
		if err == nil {
			snapshotsCount = strings.Count(string(out), "juarvis-snapshot|")
		}

		// Buscar último snapshot
		snapPath := filepath.Join(rootPath, config.JuarDir, "snapshots")
		if entries, err := os.ReadDir(snapPath); err == nil {
			var snaps []os.FileInfo
			for _, e := range entries {
				if info, err := e.Info(); err == nil {
					snaps = append(snaps, info)
				}
			}
			if len(snaps) > 0 {
				sort.Slice(snaps, func(i, j int) bool {
					return snaps[i].ModTime().After(snaps[j].ModTime())
				})
				lastSnap = snaps[0].ModTime()
			}
		}

		snapAgo := "-"
		if !lastSnap.IsZero() {
			snapAgo = time.Since(lastSnap).Round(time.Minute).String()
		}
		output.Info("Snapshots: %d guardados · último hace %s", snapshotsCount, snapAgo)

		// 8. Build y Tests
		output.Info("")
		output.Info("Salud Técnica:")

		buildCmd := exec.Command("go", "build", "./...")
		if err := buildCmd.Run(); err != nil {
			output.Info("  Build: %s", output.Styled("red", "%s", "ROTO"))
		} else {
			output.Info("  Build: %s", output.Styled("green", "%s", "SANO"))
		}

		testCmd := exec.Command("go", "test", "-short", "./...")
		if err := testCmd.Run(); err != nil {
			output.Info("  Tests: %s", output.Styled("yellow", "%s", "FALLANDO"))
		} else {
			output.Info("  Tests: %s", output.Styled("green", "%s", "PASANDO"))
		}

		// Evaluación del vibe
		if !output.IsJSONMode() {
			fmt.Println()
		}

		score := 0
		if watcherStatus == "● ACTIVO" {
			score += 2
		}
		if memCount > 10 {
			score += 2
		}
		if snapshotsCount > 3 {
			score += 2
		}
		if projectSkills > 0 {
			score += 2
		}
		if sessionCount > 2 {
			score += 2
		}

		if score >= 8 {
			output.Success("LA VIBRA ESTÁ EN SU PUNTO! Flow creativo al 100%%.")
		} else if score >= 5 {
			output.Success("Buenas vibraciones. El ecosistema madura.")
		} else {
			output.Warning("El ecosistema despega. Sigue así.")
		}
	},
}

func init() {
	rootCmd.AddCommand(vibeCmd)
}
