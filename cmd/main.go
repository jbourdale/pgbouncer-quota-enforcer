package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	verbose bool
	config  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pgbouncer-quota-enforcer",
	Short: "pgbouncer-quota-enforcer",
	Long:  `pgbouncer-quota-enforcer is a tool to enforce quota limits on pgbouncer connections.`,
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			fmt.Println("Verbose mode enabled")
		}
		if config != "" {
			fmt.Printf("Using config file: %s\n", config)
		}
		fmt.Println("Hello from the root command!")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags that will be available to all commands
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&config, "config", "c", "", "config file")
}

func main() {
	Execute()
}
