package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"juarvis/pkg/config"
	"juarvis/pkg/output"
	"juarvis/pkg/root"
)

var doctorFix bool
var doctorVerbose bool

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnóstico del ecosistema Juarvis",
	Long:  `Ejecuta un diagnóstico completo del entorno: PATH, Go, Git, ecosistema, memoria, permisos y servicios. Similar a 'brew doctor' o 'flutter doctor'.`,
	Run: func(cmd *cobra.Command, args []string) {
		output.Banner("JUARVIS DOCTOR")
		output.Info("Diagnóstico del ecosistema")
		fmt.Println()

		type checkResult struct {
			name    string
			pass   bool
			hint   string
			details string
		}

		results := []checkResult{}

		check1 := checkResult{name: "juarvis en PATH"}
		if path, err := exec.LookPath("juarvis"); err != nil {
			check1.pass = false
			check1.hint = "Ejecuta: make install"
		} else {
			check1.pass = true
			if doctorVerbose {
				check1.details = fmt.Sprintf("%s (%s)", path, Version)
			}
		}
		results = append(results, check1)

		check2 := checkResult{name: "Go >= 1.23"}
		out, err := exec.Command("go", "version").Output()
		if err != nil {
			check2.pass = false
			check2.hint = "Instala Go 1.23+ en https://go.dev"
		} else {
			versionStr := string(out)
			re := regexp.MustCompile(`go(\d+\.\d+)`)
			matches := re.FindStringSubmatch(versionStr)
			if len(matches) < 2 {
				check2.pass = false
				check2.hint = "No se pudo parsear versión"
			} else {
				minor, _ := strconv.Atoi(strings.Split(matches[1], ".")[1])
				check2.pass = minor >= 23
				check2.details = strings.TrimSpace(versionStr)
				if !check2.pass {
					check2.hint = "Instala Go 1.23+"
				}
			}
		}
		results = append(results, check2)

		check3 := checkResult{name: "Git disponible"}
		if _, err := exec.LookPath("git"); err != nil {
			check3.pass = false
			check3.hint = "Instala Git"
		} else {
			check3.pass = true
		}
		results = append(results, check3)

		check4 := checkResult{name: "Ecosistema inicializado"}
		r, err := root.GetRoot()
		if err != nil {
			check4.pass = false
			check4.hint = "Ejecuta: juarvis init"
		} else {
			check4.pass = true
			check4.details = r
		}
		results = append(results, check4)

		check5 := checkResult{name: "MCP memory accesible"}
		if r != "" {
			memFile := filepath.Join(r, config.JuarDir, config.SkillRegistryFile)
			if _, err := os.Stat(memFile); err != nil {
				check5.pass = false
				check5.hint = "Ejecuta: juarvis up"
			} else {
				check5.pass = true
				check5.details = memFile
			}
		} else {
			check5.pass = false
			check5.hint = "Primero inicializa el ecosistema"
		}
		results = append(results, check5)

		check6 := checkResult{name: "Permisos escritura"}
		configDir, err := os.UserConfigDir()
		if err != nil {
			check6.pass = false
			check6.hint = "No se pudo obtener directorio de configuración"
		} else {
			juarConfigDir := filepath.Join(configDir, "juarvis")
			if err := os.MkdirAll(juarConfigDir, 0755); err != nil {
				check6.pass = false
				check6.hint = fmt.Sprintf("Crear directorio: mkdir -p %s", juarConfigDir)
			} else {
				check6.pass = true
				check6.details = juarConfigDir
			}
		}
		results = append(results, check6)

		check7 := checkResult{name: "Watcher corriendo"}
		if r != "" {
			pidFile := filepath.Join(r, ".juar", "watcher.pid")
			if _, err := os.Stat(pidFile); err != nil {
				check7.pass = false
				check7.hint = "Ejecuta: juarvis watch --daemon"
			} else {
				pidBytes, readErr := os.ReadFile(pidFile)
				if readErr != nil {
					check7.pass = false
					check7.hint = "PID file corrupto"
				} else {
					pid, parseErr := strconv.Atoi(string(pidBytes))
					if parseErr != nil {
						check7.pass = false
						check7.hint = "PID inválido"
					} else {
						proc, procErr := os.FindProcess(pid)
						if procErr != nil {
							check7.pass = false
							check7.hint = fmt.Sprintf("Proceso %d no existe", pid)
						} else if err := proc.Signal(syscall.Signal(0)); err != nil {
							check7.pass = false
							check7.hint = fmt.Sprintf("Proceso %d no existe", pid)
						} else {
							check7.pass = true
							check7.details = fmt.Sprintf("PID: %d", pid)
						}
					}
				}
			}
		} else {
			check7.pass = false
			check7.hint = "Primero inicializa el ecosistema"
		}
		results = append(results, check7)

		check8 := checkResult{name: "Cache válido"}
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			check8.pass = false
			check8.hint = "No se pudo obtener directorio de cache"
		} else {
			juarCacheDir := filepath.Join(cacheDir, "juarvis")
			if _, err := os.Stat(juarCacheDir); err != nil {
				check8.pass = true
				check8.details = "Cache vacío (ok)"
			} else {
				check8.pass = true
				check8.details = juarCacheDir
			}
		}
		results = append(results, check8)

		passing, failing := 0, 0
		for _, c := range results {
			if c.pass {
				passing++
				output.Success(c.name)
				if doctorVerbose && c.details != "" {
					fmt.Printf("    %s\n", c.details)
				}
			} else {
				failing++
				output.Warning("%s → %s", c.name, c.hint)
			}
		}

		fmt.Println()
		fmt.Printf("%d checks, %d passing, %d warnings\n", len(results), passing, failing)
		fmt.Println()

		if doctorFix && failing > 0 {
			output.Info("Ejecutando auto-remedios...")
			fmt.Println()

			for _, c := range results {
				if !c.pass {
					switch c.name {
					case "Permisos escritura":
						configDir, _ := os.UserConfigDir()
						juarConfigDir := filepath.Join(configDir, "juarvis")
						if err := os.MkdirAll(juarConfigDir, 0755); err != nil {
							output.Error("No se pudo crear %s: %v", juarConfigDir, err)
						} else {
							output.Success("Directorio de config creado: %s", juarConfigDir)
						}
					}
				}
			}
			output.Success("Auto-remedios completados")
		} else if failing > 0 {
			output.Info("💡 Para remediar ejecuta: juarvis doctor --fix")
		} else {
			output.Success("¡Todo en orden! ¡A vibrar! 🚀")
		}
	},
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorFix, "fix", false, "Auto-remediar problemas solveables")
	doctorCmd.Flags().BoolVar(&doctorVerbose, "verbose", false, "Mostrar detalles de cada check")
	rootCmd.AddCommand(doctorCmd)
}