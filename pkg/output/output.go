package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

var jsonMode = false

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
		fmt.Printf("✅ %s\n", formatted)
	}
}

func Error(msg string, args ...interface{}) {
	formatted := fmt.Sprintf(msg, args...)
	if jsonMode {
		printJSONError(map[string]interface{}{"status": "error", "message": formatted})
	} else {
		fmt.Fprintf(os.Stderr, "❌ %s\n", formatted)
	}
}

func Warning(msg string, args ...interface{}) {
	formatted := fmt.Sprintf(msg, args...)
	if jsonMode {
		printJSON(map[string]interface{}{"status": "warning", "message": formatted})
	} else {
		fmt.Printf("⚠️  %s\n", formatted)
	}
}

func Info(msg string, args ...interface{}) {
	formatted := fmt.Sprintf(msg, args...)
	if jsonMode {
		printJSON(map[string]interface{}{"status": "info", "message": formatted})
	} else {
		fmt.Printf("ℹ️  %s\n", formatted)
	}
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
		os.Exit(1)
	}
}

func printJSONError(data interface{}) {
	enc := json.NewEncoder(os.Stderr)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		fmt.Fprintf(os.Stderr, "Error codificando JSON: %v\n", err)
		os.Exit(1)
	}
}
