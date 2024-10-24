package main

import (
	"fmt"
	"os"

	_ "go.uber.org/automaxprocs"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
