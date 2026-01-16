package main

import (
	"github.com/spf13/cobra"
)

func newSimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sim",
		Short: "Run simulations",
		Long:  "Run motor control simulations with various configurations.",
	}

	cmd.AddCommand(newSimStepCmd())

	return cmd
}
