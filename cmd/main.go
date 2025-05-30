package main

import (
	"os"
	"pgbouncer-quota-enforcer/internal/app/interfaces"
)

func main() {
	rootCmd := interfaces.NewRootCommand()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
