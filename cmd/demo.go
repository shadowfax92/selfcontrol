package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"sc/internal/config"
	"sc/internal/ipc"

	"github.com/spf13/cobra"
)

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Unblock all sites for a short window with no confirmation (for quick demos)",
	Long:  "Demo mode immediately unblocks every configured domain for demo_duration (default 5m) with no warnings or cooldown, then they auto-reblock. Intended for quick demos.",
	Args:  cobra.NoArgs,
	RunE:  runDemo,
}

func init() {
	rootCmd.AddCommand(demoCmd)
}

func runDemo(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		cfg = config.Default()
	}
	dur := resolveDemoDuration(cfg)

	client := newClient()
	resp, err := client.Send(ipc.Request{
		Command: ipc.CmdUnblock,
		Args:    map[string]string{"duration": dur.String()},
	})
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("daemon: %s", resp.Error)
	}

	raw, _ := json.Marshal(resp.Data)
	var data ipc.UnblockData
	json.Unmarshal(raw, &data)

	if len(data.Domains) == 0 {
		fmt.Println("No domains configured — nothing to unblock.")
		return nil
	}
	fmt.Printf("Demo mode: all sites unblocked for %s. Run 'sc reblock' to end early.\n", data.Duration)
	return nil
}

// resolveDemoDuration falls back to the default when demo_duration is unset or
// non-positive, so a hand-edited config can't produce a do-nothing demo.
func resolveDemoDuration(cfg *config.Config) time.Duration {
	if d := cfg.Settings.DemoDuration.Duration; d > 0 {
		return d
	}
	return config.DefaultDemoDuration
}
