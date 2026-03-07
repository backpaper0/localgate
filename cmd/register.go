package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

func newRegisterCmd() *cobra.Command {
	var serverFlag string

	cmd := &cobra.Command{
		Use:   "register <name> <target|port>",
		Short: "サービスを登録する",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			target := args[1]
			if isPortOnly(target) {
				host, err := os.Hostname()
				if err != nil {
					return fmt.Errorf("ホスト名の取得に失敗しました: %v", err)
				}
				target = host + ":" + target
			}

			serverURL := resolveServerURL(serverFlag)
			reqBody := registerServiceRequest{Name: args[0], Target: target}
			data, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("リクエストの作成に失敗しました: %v", err)
			}

			resp, err := http.Post(serverURL+"/services", "application/json", bytes.NewReader(data))
			if err != nil {
				return fmt.Errorf("接続エラー: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusCreated {
				fmt.Fprintf(cmd.OutOrStdout(), "サービス '%s' を登録しました\n", args[0])
				return nil
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

func isPortOnly(s string) bool {
	n, err := strconv.Atoi(s)
	return err == nil && n >= 1 && n <= 65535
}

func init() {
	rootCmd.AddCommand(newRegisterCmd())
}
