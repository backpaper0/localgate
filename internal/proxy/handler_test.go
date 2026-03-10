package proxy_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"localgate/internal/proxy"
)

func TestProxyForwardsRequest(t *testing.T) {
	// バックエンドサーバをセットアップ
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello from backend"))
	}))
	defer backend.Close()

	// バックエンドのアドレス (scheme なし)
	target := strings.TrimPrefix(backend.URL, "http://")

	handler := proxy.NewHandler()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)

	handler.ServeHTTP(w, r, target)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "hello from backend" {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProxyPreservesMethod(t *testing.T) {
	var receivedMethod string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	target := strings.TrimPrefix(backend.URL, "http://")
	handler := proxy.NewHandler()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("body"))
	r.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(w, r, target)

	if receivedMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", receivedMethod)
	}
}

func TestProxyPreservesHeaders(t *testing.T) {
	var receivedHeader string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-Custom-Header")
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	target := strings.TrimPrefix(backend.URL, "http://")
	handler := proxy.NewHandler()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Custom-Header", "test-value")

	handler.ServeHTTP(w, r, target)

	if receivedHeader != "test-value" {
		t.Errorf("expected header value 'test-value', got '%s'", receivedHeader)
	}
}

func TestProxyReturns502OnBackendDown(t *testing.T) {
	handler := proxy.NewHandler()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	// 存在しないバックエンド
	handler.ServeHTTP(w, r, "localhost:19999")

	if w.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", w.Code)
	}
}
