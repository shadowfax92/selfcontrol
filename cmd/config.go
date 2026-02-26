package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"sc/internal/config"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current configuration",
	RunE:  runConfigShow,
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print config file path",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(config.ConfigPath())
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open config in $EDITOR",
	RunE: func(cmd *cobra.Command, args []string) error {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}
		c := exec.Command(editor, config.ConfigPath())
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func init() {
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configEditCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(config.ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("No config file found at %s\n", config.ConfigPath())
			fmt.Println("Run any sc command to create a default config.")
			return nil
		}
		return err
	}

	fmt.Printf("# %s\n", config.ConfigPath())
	fmt.Print(string(data))
	return nil
}
