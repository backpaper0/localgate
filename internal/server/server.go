package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"localgate/internal/logger"
	"localgate/internal/management"
	"localgate/internal/proxy"
	"localgate/internal/registry"
)

// ServerConfig はサーバの設定を保持する
type ServerConfig struct {
	Port                  int
	Hostname              string // 追加の自己ホスト名（省略時は空文字列 ""）
	PortalRefreshInterval int    // ポータル画面のポーリング間隔（秒）、デフォルト 2
}

// ProxyServer はHTTPリクエストを受信しルーティングを行うサーバ
type ProxyServer struct {
	config        ServerConfig
	registry      registry.ServiceRegistry
	proxy         proxy.Handler
	management    *management.API
	httpServer    *http.Server
	selfHostnames map[string]struct{}
}

// NewProxyServer は新しい ProxyServer を返す。
// "localhost" は常に自己ホスト名として扱われる。
// config.Hostname が空でない場合は小文字正規化して追加する。
func NewProxyServer(config ServerConfig, reg registry.ServiceRegistry) *ProxyServer {
	selfHostnames := map[string]struct{}{
		"localhost": {},
	}
	if config.Hostname != "" {
		selfHostnames[strings.ToLower(config.Hostname)] = struct{}{}
	}

	s := &ProxyServer{
		config:        config,
		registry:      reg,
		proxy:         proxy.NewHandler(),
		management:    management.NewAPI(reg, config.PortalRefreshInterval),
		selfHostnames: selfHostnames,
	}
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: s,
	}
	return s
}

// Start はHTTPサーバを起動する
func (s *ProxyServer) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown はグレースフルシャットダウンを行う
func (s *ProxyServer) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// ServeHTTP はリクエストをルーティングする。
// 判定順序:
//  1. Hostヘッダからポートを除去し小文字に正規化
//  2. selfHostnames に一致する → management.API へ
//  3. proxy.ExtractSubdomain でサブドメイン抽出 → "" なら management.API へ
//  4. サブドメインあり → registry.Lookup → proxy.Handler または 404
func (s *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if host == "" {
		logger.Debug("リクエスト拒否: Hostヘッダなし", "method", r.Method, "path", r.URL.Path)
		writeError(w, http.StatusBadRequest, "invalid host header")
		return
	}

	logger.Debug("リクエスト受信", "method", r.Method, "host", host, "path", r.URL.Path)

	// ポートを除去して小文字に正規化
	hostname := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = h
	}
	hostname = strings.ToLower(hostname)

	// 自己ホスト名チェック
	if _, isSelf := s.selfHostnames[hostname]; isSelf {
		logger.Debug("管理APIへルーティング (自己ホスト名)", "hostname", hostname)
		s.management.ServeHTTP(w, r)
		return
	}

	subdomain := proxy.ExtractSubdomain(host)
	if subdomain == "" {
		logger.Debug("管理APIへルーティング (サブドメインなし)", "hostname", hostname)
		s.management.ServeHTTP(w, r)
		return
	}

	logger.Debug("サービス検索", "subdomain", subdomain)
	target, found := s.registry.Lookup(subdomain)
	if !found {
		logger.Debug("サービス未登録", "subdomain", subdomain)
		writeError(w, http.StatusNotFound, "service not found")
		return
	}

	logger.Debug("プロキシ転送", "subdomain", subdomain, "target", target)
	s.proxy.ServeHTTP(w, r, target)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
