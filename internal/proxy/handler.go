package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Handler はバックエンドへのリバースプロキシ転送を担う
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request, target string)
}

type reverseProxyHandler struct{}

// NewHandler は新しい ProxyHandler を返す
func NewHandler() Handler {
	return &reverseProxyHandler{}
}

func (h *reverseProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, target string) {
	targetURL := &url.URL{
		Scheme: "http",
		Host:   target,
	}

	rp := httputil.NewSingleHostReverseProxy(targetURL)
	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{"error": "backend unavailable"})
	}

	rp.ServeHTTP(w, r)
}
