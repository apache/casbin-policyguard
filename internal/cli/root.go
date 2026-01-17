package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "policywall",
	Short: "PolicyWall - Kubernetes admission controller with Casbin",
	Long: `PolicyWall is a Kubernetes admission controller that uses Casbin for policy enforcement.
It provides a web dashboard for policy visualization, editing, and testing.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}
