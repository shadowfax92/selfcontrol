package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"sc/internal/config"
	"sc/internal/ipc"

	"github.com/spf13/cobra"
)

var skipConfirm bool

var unblockCmd = &cobra.Command{
	Use:   "unblock [domain...] [duration]",
	Short: "Temporarily unblock domains (all if none specified)",
	Long:  "Temporarily unblock one or more domains. No args unblocks all. Last argument is parsed as duration (e.g. 15m, 1h). If omitted, uses default_duration from config.",
	RunE:  runUnblock,
}

func init() {
	unblockCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt")
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

	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		cfg = config.Default()
	}

	if duration == "" {
		duration = cfg.Settings.DefaultDuration.Duration.String()
	}

	// Enforce max unblock duration
	if cfg.Settings.MaxUnblockDuration.Duration > 0 {
		dur, _ := time.ParseDuration(duration)
		if dur > cfg.Settings.MaxUnblockDuration.Duration {
			fmt.Printf("Requested duration %s exceeds max allowed %s, capping.\n",
				duration, cfg.Settings.MaxUnblockDuration.Duration)
			duration = cfg.Settings.MaxUnblockDuration.Duration.String()
		}
	}

	// Step through each warning, require confirmation for each
	if !skipConfirm && len(cfg.Settings.UnblockWarnings) > 0 {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println()
		for _, w := range cfg.Settings.UnblockWarnings {
			fmt.Printf("  %s [y/N] ", w)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		target := "all domains"
		if len(domains) > 0 {
			target = strings.Join(domains, ", ")
		}
		fmt.Printf("\n  Unblock %s for %s? [y/N] ", target, duration)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Cancelled.")
			return nil
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
