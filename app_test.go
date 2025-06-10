package main

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestBindCobraAndViper_NoConfigFile(t *testing.T) {
	tempDir := t.TempDir()

	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir to temp dir: %v", err)
	}
	defer os.Chdir(prev)

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("foo", false, "")

	if err := bindCobraAndViper(cmd); err != nil {
		t.Fatalf("bindCobraAndViper returned error: %v", err)
	}
}
