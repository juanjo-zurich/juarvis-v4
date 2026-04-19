package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Exit codes semánticos.
// Permiten que scripts externos distingan el tipo de fallo sin parsear mensajes.
//
//	juarvis init; echo $?   # → 2 si no hay ecosistema, 6 si sin permisos
const (
	ExitOK           = 0 // todo correcto
	ExitGeneric      = 1 // error genérico no clasificado
	ExitNoEcosystem  = 2 // .juar/ ausente — ecosistema no inicializado
	ExitConfigError  = 3 // configuración corrupta o inválida (JSON, YAML)
	ExitBuildFailed  = 4 // compilación o verificación fallida
	ExitTestFailed   = 5 // tests fallidos
	ExitPermission   = 6 // sin permisos de escritura/lectura
	ExitPluginError  = 7 // error en plugin: instalación, carga o eliminación
	ExitWatcherError = 8 // error en el watcher daemon
)

// exitFunc es la función de salida del proceso. En producción es os.Exit;
// en tests se sustituye por una función de captura para evitar llamadas reales.
var exitFunc = os.Exit

// Fatal imprime el error en stderr, una pista accionable opcional,
// y termina el proceso con el código semántico dado.
//
// Ejemplo:
//
//	output.Fatal(output.ExitNoEcosystem,
//	    "Ejecuta 'juarvis init' en este directorio primero",
//	    "Ecosistema no encontrado: %v", err)
func Fatal(code int, hint, msg string, args ...interface{}) {
	Error(msg, args...)
	if hint != "" {
		fmt.Fprintf(os.Stderr, "   → %s\n", hint)
	}
	exitFunc(code)
}

var jsonMode = false

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

func SetJSONMode(enabled bool) {
	jsonMode = enabled
}

func IsJSONMode() bool {
	return jsonMode
}

func Success(msg string, args ...interface{}) {
	formatted := fmt.Sprintf(msg, args...)
	if jsonMode {
		printJSON(map[string]interface{}{"status": "success", "message": formatted})
	} else {
		fmt.Printf("%s✅ %s%s\n", colorGreen, formatted, colorReset)
	}
}

func Error(msg string, args ...interface{}) {
	formatted := fmt.Sprintf(msg, args...)
	if jsonMode {
		printJSONError(map[string]interface{}{"status": "error", "message": formatted})
	} else {
		fmt.Fprintf(os.Stderr, "%s❌ %s%s\n", colorRed, formatted, colorReset)
	}
}

func Warning(msg string, args ...interface{}) {
	formatted := fmt.Sprintf(msg, args...)
	if jsonMode {
		printJSON(map[string]interface{}{"status": "warning", "message": formatted})
	} else {
		fmt.Printf("%s⚠️  %s%s\n", colorYellow, formatted, colorReset)
	}
}

func Info(msg string, args ...interface{}) {
	formatted := fmt.Sprintf(msg, args...)
	if jsonMode {
		printJSON(map[string]interface{}{"status": "info", "message": formatted})
	} else {
		fmt.Printf("%sℹ️  %s%s\n", colorBlue, formatted, colorReset)
	}
}

func Banner(msg string) {
	if jsonMode {
		return
	}
	fmt.Println()
	fmt.Printf("%s%s%s%s%s\n", colorBold, colorPurple, "== ", msg, " ==")
	fmt.Println(colorReset)
}

func Styled(color string, msg string, args ...interface{}) string {
	formatted := fmt.Sprintf(msg, args...)
	var c string
	switch strings.ToLower(color) {
	case "red":
		c = colorRed
	case "green":
		c = colorGreen
	case "yellow":
		c = colorYellow
	case "blue":
		c = colorBlue
	case "purple":
		c = colorPurple
	case "cyan":
		c = colorCyan
	case "bold":
		c = colorBold
	default:
		return formatted
	}
	return c + formatted + colorReset
}

func PrintJSON(data interface{}) {
	printJSON(data)
}

func PrintTable(headers []string, rows [][]string) {
	if jsonMode {
		items := make([]map[string]string, len(rows))
		for i, row := range rows {
			item := make(map[string]string)
			for j, header := range headers {
				if j < len(row) {
					item[header] = row[j]
				}
			}
			items[i] = item
		}
		printJSON(items)
	} else {
		widths := make([]int, len(headers))
		for i, h := range headers {
			widths[i] = len(h)
		}
		for _, row := range rows {
			for i, cell := range row {
				if i < len(widths) && len(cell) > widths[i] {
					widths[i] = len(cell)
				}
			}
		}

		for i, h := range headers {
			fmt.Printf("%-*s ", widths[i], h)
		}
		fmt.Println()
		for _, w := range widths {
			fmt.Print(strings.Repeat("-", w+1))
		}
		fmt.Println()
		for _, row := range rows {
			for i, cell := range row {
				if i < len(widths) {
					fmt.Printf("%-*s ", widths[i], cell)
				}
			}
			fmt.Println()
		}
	}
}

func printJSON(data interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error codificando JSON: %v\n", err)
		exitFunc(ExitGeneric)
	}
}

func printJSONError(data interface{}) {
	enc := json.NewEncoder(os.Stderr)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		fmt.Fprintf(os.Stderr, "Error codificando JSON: %v\n", err)
		exitFunc(ExitGeneric)
	}
}
