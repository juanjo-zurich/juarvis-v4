package cmd

import (
	"fmt"
	"os"

	"juarvis/pkg/kb"
	"juarvis/pkg/output"
	"juarvis/pkg/root"

	"github.com/spf13/cobra"
)

var kbCmd = &cobra.Command{
	Use:   "kb",
	Short: "Knowledge Base - gestionarbase de conocimiento",
	Long:  `Comandos para buscar, añadir y gestionar conocimiento aprendido.`,
}

// kbSearchCmd: juarvis kb search <query>
var kbSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Buscar en la base de conocimiento",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, _ := root.GetRoot()

		kbPath := rootPath + "/.juar/kb"
		os.MkdirAll(kbPath, 0755)

		kbBase, err := kb.NewKnowledgeBase(kb.DefaultConfig())
		if err != nil {
			output.Error("Error inicializando KB: %v", err)
			return
		}

		query := args[0]
		results, err := kbBase.Search(query, nil)
		if err != nil {
			output.Error("Error buscando: %v", err)
			return
		}

		if len(results) == 0 {
			output.Info("No se encontraron resultados para: %s", query)
			return
		}

		output.Info("Resultados (%d):", len(results))
		for _, r := range results {
			fmt.Printf("  [%s] %s\n", r.Type, r.Title)
		}
	},
}

// kbListCmd: juarvis kb list
var kbListCmd = &cobra.Command{
	Use:   "list",
	Short: "Listar todas las entradas",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		kbBase, err := kb.NewKnowledgeBase(kb.DefaultConfig())
		if err != nil {
			output.Error("Error inicializando KB: %v", err)
			return
		}

		entries := kbBase.GetAll()
		if len(entries) == 0 {
			output.Info("La base de conocimiento está vacía")
			return
		}

		output.Info("Entradas (%d):", len(entries))
		for _, e := range entries {
			fmt.Printf("  [%s] %s\n", e.Type, e.Title)
		}
	},
}

// kbStatsCmd: juarvis kb stats
var kbStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Mostrar estadísticas de la KB",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		kbBase, err := kb.NewKnowledgeBase(kb.DefaultConfig())
		if err != nil {
			output.Error("Error inicializando KB: %v", err)
			return
		}

		stats := kbBase.GetStats()
		output.Info("Estadísticas:")
		fmt.Printf("  Total entradas: %d\n", stats.TotalEntries)
		fmt.Printf("  Por tipo:\n")
		for t, count := range stats.ByType {
			fmt.Printf("    %s: %d\n", t, count)
		}
	},
}

// kbAddCmd: juarvis kb add <type> <title> [content]
var kbAddCmd = &cobra.Command{
	Use:   "add [type] [title]",
	Short: "Añadir entrada a la KB",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		kbBase, err := kb.NewKnowledgeBase(kb.DefaultConfig())
		if err != nil {
			output.Error("Error inicializando KB: %v", err)
			return
		}

		entry := kb.NewKnowledgeEntry(kb.KnowledgeType(args[0]), args[1], "")
		if len(args) > 2 {
			entry.Content = args[2]
		}

		if err := kbBase.Add(entry); err != nil {
			output.Error("Error añadiendo entrada: %v", err)
			return
		}

		output.Success("Entrada añadida: %s", entry.ID)
	},
}

func init() {
	// Registrar subcomandos
	kbCmd.AddCommand(kbSearchCmd)
	kbCmd.AddCommand(kbListCmd)
	kbCmd.AddCommand(kbStatsCmd)
	kbCmd.AddCommand(kbAddCmd)
	rootCmd.AddCommand(kbCmd)
}