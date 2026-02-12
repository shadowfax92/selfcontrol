package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sc/internal/config"
	"sc/internal/ipc"

	"github.com/spf13/cobra"
)

var unblockCmd = &cobra.Command{
	Use:   "unblock [domain...] [duration]",
	Short: "Temporarily unblock domains (all if none specified)",
	Long:  "Temporarily unblock one or more domains. No args unblocks all. Last argument is parsed as duration (e.g. 15m, 1h). If omitted, uses default_duration from config.",
	RunE:  runUnblock,
}

func init() {
	rootCmd.AddCommand(unblockCmd)
}

func runUnblock(cmd *cobra.Command, args []string) error {
	var domains []string
	var duration string

	// Try parsing last arg as duration
	if len(args) > 0 {
		if _, err := time.ParseDuration(args[len(args)-1]); err == nil {
			duration = args[len(args)-1]
			domains = args[:len(args)-1]
		} else {
			domains = args
		}
	}

	if duration == "" {
		cfg, err := config.Load(config.ConfigPath())
		if err == nil {
			duration = cfg.Settings.DefaultDuration.Duration.String()
		} else {
			duration = "15m"
		}
	}

	client := newClient()
	resp, err := client.Send(ipc.Request{
		Command: ipc.CmdUnblock,
		Args: map[string]string{
			"domains":  strings.Join(domains, ","),
			"duration": duration,
		},
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

	for _, d := range data.Domains {
		fmt.Printf("Unblocked %s for %s\n", d, data.Duration)
	}
	return nil
}
