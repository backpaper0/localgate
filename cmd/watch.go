package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"localgate/internal/watcher"
)

func newWatchCmd() *cobra.Command {
	var serverFlag string
	var intervalFlag int

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "LISTENポートを監視して自動登録・解除する",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			if intervalFlag < 1 {
				return fmt.Errorf("--interval は 1 以上の値を指定してください")
			}

			serverURL := resolveServerURL(serverFlag)
			interval := time.Duration(intervalFlag) * time.Second

			// 起動時の接続確認
			client := watcher.NewManagementHTTPClient(serverURL)
			if err := client.Ping(); err != nil {
				fmt.Fprintf(os.Stderr, "watch: localgate サーバへの接続に失敗しました: %v\n", err)
				return err
			}

			// シグナルハンドリング（SIGINT/SIGTERM でコンテキストをキャンセル）
			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			hostname, err := os.Hostname()
			if err != nil {
				return fmt.Errorf("ホスト名の取得に失敗: %w", err)
			}

			scanner := watcher.NewProcNetScanner()
			w := watcher.NewWatcher(scanner, client, interval, hostname)
			return w.Run(ctx)
		},
	}

	cmd.Flags().StringVar(&serverFlag, "server", "", "サーバーURL (デフォルト: $LOCALGATE_SERVER または http://localhost:9000)")
	cmd.Flags().IntVar(&intervalFlag, "interval", 1, "ポーリング間隔（秒）")
	return cmd
}

func init() {
	rootCmd.AddCommand(newWatchCmd())
}
