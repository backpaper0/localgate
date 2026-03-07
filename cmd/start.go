package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"localgate/internal/registry"
	"localgate/internal/server"
)

var port int

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "プロキシサーバを起動する",
	RunE:  runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().IntVar(&port, "port", 9000, "待ち受けポート番号 (1-65535)")
}

func runStart(cmd *cobra.Command, args []string) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("ポート番号は1〜65535の範囲で指定してください: %d", port)
	}

	reg := registry.NewServiceRegistry()
	srv := server.NewProxyServer(server.ServerConfig{Port: port}, reg)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Fprintf(os.Stdout, "LocalGate starting on port %d\n", port)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	select {
	case <-ctx.Done():
		fmt.Fprintln(os.Stdout, "Shutting down...")
		return srv.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}
