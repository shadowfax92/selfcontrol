package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"sc/internal/ipc"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <domain...>",
	Short: "Remove domains from the block list",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	client := newClient()
	resp, err := client.Send(ipc.Request{
		Command: ipc.CmdRemove,
		Args:    map[string]string{"domains": strings.Join(args, ",")},
	})
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("daemon: %s", resp.Error)
	}

	raw, _ := json.Marshal(resp.Data)
	var data ipc.MutateData
	json.Unmarshal(raw, &data)

	if len(data.Removed) == 0 {
		fmt.Println("No matching domains found")
	} else {
		for _, d := range data.Removed {
			fmt.Printf("Removed %s\n", d)
		}
	}
	return nil
}
