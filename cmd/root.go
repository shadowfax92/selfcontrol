package cmd

import (
	"fmt"
	"os"

	"sc/internal/config"
	"sc/internal/ipc"

	"github.com/spf13/cobra"
)

var version = "dev"

func SetVersion(v string) {
	version = v
}

var rootCmd = &cobra.Command{
	Use:   "sc",
	Short: "Block distracting websites by default, temporarily unblock with timers",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sc", version)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func newClient() *ipc.Client {
	return ipc.NewClient(config.SocketPath())
}
