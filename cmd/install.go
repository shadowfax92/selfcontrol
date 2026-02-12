package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"sc/internal/config"

	"github.com/spf13/cobra"
)

const plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.sc.daemon</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.BinaryPath}}</string>
        <string>daemon</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>{{.LogPath}}</string>
    <key>StandardErrorPath</key>
    <string>{{.LogPath}}</string>
</dict>
</plist>
`

const plistPath = "/Library/LaunchDaemons/com.sc.daemon.plist"

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install launchd daemon (requires sudo)",
	RunE:  runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("install requires root — run: sudo sc install")
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine binary path: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("resolve symlinks: %w", err)
	}

	if err := os.MkdirAll(config.ConfigDir(), 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if err := os.MkdirAll(config.DataDir(), 0755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	// Create default config if missing
	if _, err := os.Stat(config.ConfigPath()); os.IsNotExist(err) {
		if _, err := config.Load(config.ConfigPath()); err != nil {
			return fmt.Errorf("create default config: %w", err)
		}
	}

	tmpl, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return err
	}

	f, err := os.Create(plistPath)
	if err != nil {
		return fmt.Errorf("create plist: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, struct {
		BinaryPath string
		LogPath    string
	}{
		BinaryPath: exe,
		LogPath:    config.DaemonLog(),
	}); err != nil {
		return err
	}

	if err := exec.Command("launchctl", "load", plistPath).Run(); err != nil {
		return fmt.Errorf("launchctl load failed: %w", err)
	}

	fmt.Println("Installed and started.")
	fmt.Printf("  Plist:  %s\n", plistPath)
	fmt.Printf("  Log:    %s\n", config.DaemonLog())
	fmt.Printf("  Config: %s\n", config.ConfigPath())
	fmt.Printf("  Socket: %s\n", config.SocketPath())
	return nil
}
