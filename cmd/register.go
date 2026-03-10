package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func newRegisterCmd() *cobra.Command {
	var serverFlag string
	var forceFlag bool

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
			resp, err := sendRegisterRequest(serverURL, args[0], target, forceFlag)
			if err != nil {
				return fmt.Errorf("接続エラー: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode == http.StatusCreated {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "サービス '%s' を登録しました\n", args[0])
				return nil
			}

			var apiErr apiError
			if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil || apiErr.Error == "" {
				return fmt.Errorf("エラー (HTTP %d)", resp.StatusCode)
			}

			if resp.StatusCode == http.StatusConflict {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "サービス '%s' は既に '%s' として登録されています。上書きしますか？ [y/N]: ", args[0], apiErr.ExistingTarget)
				scanner := bufio.NewScanner(cmd.InOrStdin())
				scanner.Scan()
				answer := strings.TrimSpace(scanner.Text())
				if answer != "y" && answer != "Y" {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "キャンセルしました\n")
					return nil
				}
				resp2, err := sendRegisterRequest(serverURL, args[0], target, true)
				if err != nil {
					return fmt.Errorf("接続エラー: %v", err)
				}
				defer func() { _ = resp2.Body.Close() }()
				if resp2.StatusCode == http.StatusCreated {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "サービス '%s' を登録しました\n", args[0])
					return nil
				}
				var apiErr2 apiError
				if err := json.NewDecoder(resp2.Body).Decode(&apiErr2); err != nil || apiErr2.Error == "" {
					return fmt.Errorf("エラー (HTTP %d)", resp2.StatusCode)
				}
				return fmt.Errorf("%s", apiErr2.Error)
			}

			return fmt.Errorf("%s", apiErr.Error)
		},
	}

	cmd.Flags().StringVar(&serverFlag, "server", "", "サーバーURL (デフォルト: $LOCALGATE_SERVER または http://localhost:9000)")
	cmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "確認なしで上書き登録する")
	return cmd
}

func sendRegisterRequest(serverURL, name, target string, force bool) (*http.Response, error) {
	reqBody := registerServiceRequest{Name: name, Target: target}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("リクエストの作成に失敗しました: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, serverURL+"/services", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if force {
		req.Header.Set("X-Force-Overwrite", "true")
	}
	return http.DefaultClient.Do(req)
}

func isPortOnly(s string) bool {
	n, err := strconv.Atoi(s)
	return err == nil && n >= 1 && n <= 65535
}

func init() {
	rootCmd.AddCommand(newRegisterCmd())
}
