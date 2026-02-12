package cmd

import (
	"encoding/json"
	"fmt"

	"sc/internal/ipc"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all blocked domains",
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	client := newClient()
	resp, err := client.Send(ipc.Request{Command: ipc.CmdList})
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("daemon: %s", resp.Error)
	}

	raw, _ := json.Marshal(resp.Data)
	var data ipc.ListData
	json.Unmarshal(raw, &data)

	if len(data.Domains) == 0 {
		fmt.Println("No domains configured")
	} else {
		for _, d := range data.Domains {
			fmt.Println(d)
		}
	}
	return nil
}
