package cmd

import (
	"bytes"
	"testing"
)

func TestRootCommand(t *testing.T) {
	// Capturamos el output
	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetErr(b)

	// Invocamos el comando sin args
	rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()

	if err != nil {
		t.Errorf("El comando root fallo: %v", err)
	}
}

func TestRootCommand_VersionFlag(t *testing.T) {
	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetErr(b)

	rootCmd.SetArgs([]string{"--version"})
	err := rootCmd.Execute()

	if err != nil {
		t.Errorf("El comando --version fallo: %v", err)
	}
}
