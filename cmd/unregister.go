package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func newUnregisterCmd() *cobra.Command {
	var serverFlag string

	cmd := &cobra.Command{
		Use:   "unregister <name>",
		Short: "サービスを解除する",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			serverURL := resolveServerURL(serverFlag)
			req, err := http.NewRequest(http.MethodDelete, serverURL+"/services/"+args[0], nil)
			if err != nil {
				return fmt.Errorf("リクエストの作成に失敗しました: %v", err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("接続エラー: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode == http.StatusNoContent {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "サービス '%s' を解除しました\n", args[0])
				return nil
			}

			if resp.StatusCode == http.StatusNotFound {
				return fmt.Errorf("サービスが見つかりません: %s", args[0])
			}

			var apiErr apiError
			if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil || apiErr.Error == "" {
				return fmt.Errorf("エラー (HTTP %d)", resp.StatusCode)
			}
			return fmt.Errorf("%s", apiErr.Error)
		},
	}

	cmd.Flags().StringVar(&serverFlag, "server", "", "サーバーURL (デフォルト: $LOCALGATE_SERVER または http://localhost:9000)")
	return cmd
}

func init() {
	rootCmd.AddCommand(newUnregisterCmd())
}
