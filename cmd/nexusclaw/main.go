package main

import (
	"os"

	"github.com/kapella-hub/NexusClaw/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
