package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"sc/internal/hosts"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove launchd daemon (requires sudo)",
	RunE:  runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, args []string) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("uninstall requires root — run: sudo sc uninstall")
	}

	_ = exec.Command("launchctl", "unload", plistPath).Run()

	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove plist: %w", err)
	}

	if err := hosts.Remove(); err != nil {
		fmt.Printf("Warning: failed to clean /etc/hosts: %v\n", err)
	}

	fmt.Println("Uninstalled. Daemon stopped, plist removed, /etc/hosts cleaned.")
	return nil
}
