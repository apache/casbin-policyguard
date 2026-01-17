package cli

import (
	"fmt"
	"os"

	"github.com/casbin/policywall/internal/server"
	"github.com/spf13/cobra"
)

var (
	port string
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Start the web dashboard",
	Long:  `Start the PolicyWall web dashboard for policy visualization, editing, and testing.`,
	Run: func(cmd *cobra.Command, args []string) {
		srv := server.NewServer(port)
		if err := srv.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start dashboard: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	dashboardCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to run the dashboard on")
}
