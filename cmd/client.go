package cmd

import "os"

const defaultServerURL = "http://localhost:9000"

// registerServiceRequest は POST /services のリクエストボディ
type registerServiceRequest struct {
	Name   string `json:"name"`
	Target string `json:"target"`
}

// serviceEntry は登録済みサービスの1エントリ
type serviceEntry struct {
	Name   string `json:"name"`
	Target string `json:"target"`
}

// listServicesResponse は GET /services のレスポンスボディ
type listServicesResponse struct {
	Services []serviceEntry `json:"services"`
}

// apiError はエラーレスポンスのボディ
type apiError struct {
	Error string `json:"error"`
}

// resolveServerURL はサーバーURLを優先順位に従って解決する。
// 優先順位: flagValue > LOCALGATE_SERVER 環境変数 > "http://localhost:9000"
func resolveServerURL(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if v := os.Getenv("LOCALGATE_SERVER"); v != "" {
		return v
	}
	return defaultServerURL
}
