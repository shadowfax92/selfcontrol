package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"sc/internal/config"
	"sc/internal/logs"

	"github.com/spf13/cobra"
)

var (
	logDomain string
	logPeriod string
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show unblock/reblock history and stats",
	RunE:  runLogs,
}

func init() {
	logsCmd.Flags().StringVar(&logDomain, "domain", "", "filter by domain")
	logsCmd.Flags().StringVar(&logPeriod, "period", "all", "time period: today, week, month, all")
	rootCmd.AddCommand(logsCmd)
}

func runLogs(cmd *cobra.Command, args []string) error {
	entries, err := logs.Query(config.LogsPath(), logs.QueryOpts{
		Domain: logDomain,
		Period: logPeriod,
	})
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Println("No log entries found")
		return nil
	}

	stats := logs.Stats(entries)
	if len(stats) > 0 {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "DOMAIN\tUNBLOCKS\tTOTAL TIME\tLAST UNBLOCK")
		for _, s := range stats {
			lastUnblock := s.LastUnblock.Format("2006-01-02 15:04")
			fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", s.Domain, s.Unblocks, logs.FormatDuration(s.TotalTime), lastUnblock)
		}
		w.Flush()
		fmt.Println()
	}

	fmt.Printf("Recent events (%d total):\n", len(entries))
	// Show last 20 entries
	start := 0
	if len(entries) > 20 {
		start = len(entries) - 20
	}
	for _, e := range entries[start:] {
		ts := e.Timestamp.Format("Jan 02 15:04")
		switch e.Event {
		case "unblock":
			fmt.Printf("  %s  unblock  %-20s  for %s\n", ts, e.Domain, e.Duration)
		case "reblock":
			reason := e.Reason
			if reason == "" {
				reason = "manual"
			}
			fmt.Printf("  %s  reblock  %-20s  (%s)\n", ts, e.Domain, reason)
		}
	}

	return nil
}
