package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var serverFlag string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "登録済みサービスを一覧表示する",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			serverURL := resolveServerURL(serverFlag)
			resp, err := http.Get(serverURL + "/services")
			if err != nil {
				return fmt.Errorf("接続エラー: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				var apiErr apiError
				if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil || apiErr.Error == "" {
					return fmt.Errorf("エラー (HTTP %d)", resp.StatusCode)
				}
				return fmt.Errorf("%s", apiErr.Error)
			}

			var listResp listServicesResponse
			if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
				return fmt.Errorf("レスポンスの解析に失敗しました: %v", err)
			}

			if len(listResp.Services) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "登録済みサービスはありません")
				return nil
			}

			for _, svc := range listResp.Services {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", svc.Name, svc.Target)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&serverFlag, "server", "", "サーバーURL (デフォルト: $LOCALGATE_SERVER または http://localhost:9000)")
	return cmd
}

func init() {
	rootCmd.AddCommand(newListCmd())
}
