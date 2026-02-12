package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"sc/internal/ipc"

	"github.com/spf13/cobra"
)

var reblockCmd = &cobra.Command{
	Use:   "reblock [domain...]",
	Short: "Immediately reblock domains (all if none specified)",
	RunE:  runReblock,
}

func init() {
	rootCmd.AddCommand(reblockCmd)
}

func runReblock(cmd *cobra.Command, args []string) error {
	reqArgs := map[string]string{}
	if len(args) > 0 {
		reqArgs["domains"] = strings.Join(args, ",")
	}

	client := newClient()
	resp, err := client.Send(ipc.Request{
		Command: ipc.CmdReblock,
		Args:    reqArgs,
	})
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("daemon: %s", resp.Error)
	}

	raw, _ := json.Marshal(resp.Data)
	var data ipc.ReblockData
	json.Unmarshal(raw, &data)

	if len(data.Domains) == 0 {
		fmt.Println("No domains were unblocked")
	} else {
		for _, d := range data.Domains {
			fmt.Printf("Reblocked %s\n", d)
		}
	}
	return nil
}
