package watcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"localgate/internal/logger"
)

// ManagementClient は localgate 管理 API のクライアントインターフェース。
type ManagementClient interface {
	// Ping は管理 API への疎通確認を行う。
	Ping() error
	// Register はサービスを登録する。
	Register(name, target string) error
	// Deregister はサービスを解除する。
	// サービスが存在しない場合もエラーとせず nil を返す。
	Deregister(name string) error
}

// registerRequest は POST /services のリクエストボディ。
type registerRequest struct {
	Name   string `json:"name"`
	Target string `json:"target"`
}

// ManagementHTTPClient は ManagementClient の HTTP 実装。
type ManagementHTTPClient struct {
	serverURL  string
	httpClient *http.Client
}

// NewManagementHTTPClient は ManagementHTTPClient を生成する。
func NewManagementHTTPClient(serverURL string) *ManagementHTTPClient {
	return &ManagementHTTPClient{
		serverURL:  serverURL,
		httpClient: &http.Client{},
	}
}

// Ping は GET /services で管理 API への疎通確認を行う。
func (c *ManagementHTTPClient) Ping() error {
	logger.Debug("Ping: 管理APIへ接続確認", "url", c.serverURL+"/services")
	resp, err := c.httpClient.Get(c.serverURL + "/services")
	if err != nil {
		logger.Debug("Ping: 接続失敗", "error", err)
		return fmt.Errorf("管理APIへの接続に失敗: %w", err)
	}
	_ = resp.Body.Close()
	logger.Debug("Ping: 接続成功", "status", resp.StatusCode)
	return nil
}

// Register は POST /services でサービスを登録する。
// 409 Conflict の場合はエラーを返す。
func (c *ManagementHTTPClient) Register(name, target string) error {
	body, err := json.Marshal(registerRequest{Name: name, Target: target})
	if err != nil {
		return fmt.Errorf("リクエストのシリアライズに失敗: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.serverURL+"/services", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	logger.Debug("Register: POSTリクエスト送信", "name", name, "target", target)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Debug("Register: リクエスト失敗", "name", name, "error", err)
		return fmt.Errorf("サービス登録リクエストに失敗: %w", err)
	}
	_ = resp.Body.Close()

	logger.Debug("Register: レスポンス受信", "name", name, "status", resp.StatusCode)
	if resp.StatusCode == http.StatusCreated {
		return nil
	}
	return fmt.Errorf("サービス登録に失敗 (HTTP %d)", resp.StatusCode)
}

// Deregister は DELETE /services/{name} でサービスを解除する。
// 404 Not Found の場合は nil を返す（冪等性）。
func (c *ManagementHTTPClient) Deregister(name string) error {
	req, err := http.NewRequest(http.MethodDelete, c.serverURL+"/services/"+name, nil)
	if err != nil {
		return err
	}

	logger.Debug("Deregister: DELETEリクエスト送信", "name", name)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Debug("Deregister: リクエスト失敗", "name", name, "error", err)
		return fmt.Errorf("サービス解除リクエストに失敗: %w", err)
	}
	_ = resp.Body.Close()

	logger.Debug("Deregister: レスポンス受信", "name", name, "status", resp.StatusCode)
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusNotFound {
		return nil
	}
	return fmt.Errorf("サービス解除に失敗 (HTTP %d)", resp.StatusCode)
}
