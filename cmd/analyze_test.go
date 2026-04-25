package cmd

import (
	"bytes"
	"testing"
)

func TestAnalyzeCommand_HelpFlag(t *testing.T) {
	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetErr(b)

	// Invocamos analyze con --help
	rootCmd.SetArgs([]string{"analyze", "--help"})
	err := rootCmd.Execute()

	if err != nil {
		t.Errorf("analyze --help fallo: %v", err)
	}

	output := b.String()
	if !bytes.Contains(b.Bytes(), []byte("Uso:")) && !bytes.Contains(b.Bytes(), []byte("Usage:")) {
		t.Errorf("No se mostro la ayuda de analyze: %s", output)
	}
}
