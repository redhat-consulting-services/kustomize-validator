package main

import (
	"os"

	"github.com/redhat-consulting-services/kustomize-validator/commands"
)

func main() {
	if err := commands.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
