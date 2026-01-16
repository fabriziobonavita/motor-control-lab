package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "mcl",
		Short: "Motor Control Lab - simulation and analysis tools",
		Long:  "mcl is a command-line tool for running motor control simulations and analyzing results.",
	}

	rootCmd.AddCommand(newSimCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
