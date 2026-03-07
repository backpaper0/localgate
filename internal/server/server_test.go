package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"localgate/internal/management"
	"localgate/internal/registry"
	"localgate/internal/server"
)

func newTestServer() (*httptest.Server, registry.ServiceRegistry) {
	reg := registry.NewServiceRegistry()
	srv := server.NewProxyServer(server.ServerConfig{Port: 0}, reg)
	return httptest.NewServer(srv), reg
}

// --- タスク 6.1: エンドツーエンドプロキシ動作テスト ---

func TestProxyForwardsToRegisteredBackend(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend response"))
	}))
	defer backend.Close()

	ts, reg := newTestServer()
	defer ts.Close()

	backendAddr := strings.TrimPrefix(backend.URL, "http://")
	reg.Register("foo", backendAddr)

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/path", nil)
	req.Host = "foo.localhost"

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestProxyReturns404AfterDeregister(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	ts, reg := newTestServer()
	defer ts.Close()

	backendAddr := strings.TrimPrefix(backend.URL, "http://")
	reg.Register("foo", backendAddr)
	reg.Deregister("foo")

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/", nil)
	req.Host = "foo.localhost"

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestProxyReturns404ForUnregisteredSubdomain(t *testing.T) {
	ts, _ := newTestServer()
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/", nil)
	req.Host = "unknown.localhost"

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestProxyReturns502WhenBackendDown(t *testing.T) {
	ts, reg := newTestServer()
	defer ts.Close()

	reg.Register("foo", "localhost:19998")

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/", nil)
	req.Host = "foo.localhost"

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", resp.StatusCode)
	}
}

// --- タスク 6.2: 管理API統合テスト ---

func TestManagementAPIRegisterAndList(t *testing.T) {
	ts, _ := newTestServer()
	defer ts.Close()

	// POST /services でサービス登録
	body := `{"name":"foo","target":"localhost:3000"}`
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/services", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /services failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}

	// GET /services で一覧確認
	req2, _ := http.NewRequest(http.MethodGet, ts.URL+"/services", nil)
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("GET /services failed: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp2.StatusCode)
	}

	var listResp management.ListServicesResponse
	json.NewDecoder(resp2.Body).Decode(&listResp)
	if len(listResp.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(listResp.Services))
	}
	if listResp.Services[0].Name != "foo" {
		t.Errorf("expected service name 'foo', got '%s'", listResp.Services[0].Name)
	}
}

func TestManagementAPIDeleteAndList(t *testing.T) {
	ts, reg := newTestServer()
	defer ts.Close()

	reg.Register("foo", "localhost:3000")

	// DELETE /services/foo
	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/services/foo", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /services/foo failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}

	// GET /services で空確認
	req2, _ := http.NewRequest(http.MethodGet, ts.URL+"/services", nil)
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("GET /services failed: %v", err)
	}
	defer resp2.Body.Close()

	var listResp management.ListServicesResponse
	json.NewDecoder(resp2.Body).Decode(&listResp)
	if len(listResp.Services) != 0 {
		t.Errorf("expected 0 services after delete, got %d", len(listResp.Services))
	}
}

func TestManagementAPIDeleteNotFound(t *testing.T) {
	ts, _ := newTestServer()
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/services/nonexistent", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestManagementAPIPostMissingFields(t *testing.T) {
	ts, _ := newTestServer()
	defer ts.Close()

	body := `{"name":"foo"}`
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/services", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}
