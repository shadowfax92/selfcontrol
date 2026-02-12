package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"sc/internal/ipc"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <domain...>",
	Short: "Add domains to the block list",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	client := newClient()
	resp, err := client.Send(ipc.Request{
		Command: ipc.CmdAdd,
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

	if len(data.Added) == 0 {
		fmt.Println("All domains already in block list")
	} else {
		for _, d := range data.Added {
			fmt.Printf("Added %s\n", d)
		}
	}
	return nil
}
