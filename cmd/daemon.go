package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"sc/internal/config"
	"sc/internal/daemon"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run the sc daemon in the foreground",
	RunE:  runDaemon,
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}

func runDaemon(cmd *cobra.Command, args []string) error {
	cfgPath := config.ConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().Timestamp().Logger()

	d := daemon.New(cfg, cfgPath, logger)

	sockPath := config.SocketPath()
	srv := daemon.NewServer(d, sockPath, logger)
	if err := srv.Start(); err != nil {
		return err
	}
	defer srv.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info().Msg("received shutdown signal")
		cancel()
	}()

	return d.Run(ctx)
}
