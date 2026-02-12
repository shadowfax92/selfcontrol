package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"sc/internal/ipc"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show blocked/unblocked status of all domains",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	client := newClient()
	resp, err := client.Send(ipc.Request{Command: ipc.CmdStatus})
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("daemon: %s", resp.Error)
	}

	raw, _ := json.Marshal(resp.Data)
	var data ipc.StatusData
	json.Unmarshal(raw, &data)

	fmt.Printf("Uptime: %s\n\n", data.Uptime)

	if len(data.Domains) == 0 {
		fmt.Println("No domains configured. Use: sc add <domain>")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "DOMAIN\tSTATE\tREMAINING")
	for _, d := range data.Domains {
		remaining := "-"
		if d.Remaining != "" {
			remaining = d.Remaining
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", d.Domain, d.State, remaining)
	}
	w.Flush()

	return nil
}
