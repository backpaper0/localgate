package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"localgate/internal/management"
	"localgate/internal/proxy"
	"localgate/internal/registry"
)

// ServerConfig はサーバの設定を保持する
type ServerConfig struct {
	Port int
}

// ProxyServer はHTTPリクエストを受信しルーティングを行うサーバ
type ProxyServer struct {
	config     ServerConfig
	registry   registry.ServiceRegistry
	proxy      proxy.Handler
	management *management.API
	httpServer *http.Server
}

// NewProxyServer は新しい ProxyServer を返す
func NewProxyServer(config ServerConfig, reg registry.ServiceRegistry) *ProxyServer {
	s := &ProxyServer{
		config:     config,
		registry:   reg,
		proxy:      proxy.NewHandler(),
		management: management.NewAPI(reg),
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

// ServeHTTP はリクエストをサブドメインの有無でプロキシまたは管理APIへルーティングする
func (s *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if host == "" {
		writeError(w, http.StatusBadRequest, "invalid host header")
		return
	}

	subdomain := proxy.ExtractSubdomain(host)
	if subdomain == "" {
		s.management.ServeHTTP(w, r)
		return
	}

	target, found := s.registry.Lookup(subdomain)
	if !found {
		writeError(w, http.StatusNotFound, "service not found")
		return
	}

	s.proxy.ServeHTTP(w, r, target)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
