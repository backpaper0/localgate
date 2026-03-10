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
		_, _ = w.Write([]byte("backend response"))
	}))
	defer backend.Close()

	ts, reg := newTestServer()
	defer ts.Close()

	backendAddr := strings.TrimPrefix(backend.URL, "http://")
	if err := reg.Register("foo", backendAddr, false); err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/path", nil)
	req.Host = "foo.localhost"

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

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
	if err := reg.Register("foo", backendAddr, false); err != nil {
		t.Fatal(err)
	}
	if err := reg.Deregister("foo"); err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/", nil)
	req.Host = "foo.localhost"

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestProxyReturns502WhenBackendDown(t *testing.T) {
	ts, reg := newTestServer()
	defer ts.Close()

	if err := reg.Register("foo", "localhost:19998", false); err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/", nil)
	req.Host = "foo.localhost"

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

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
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}

	// GET /services で一覧確認
	req2, _ := http.NewRequest(http.MethodGet, ts.URL+"/services", nil)
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("GET /services failed: %v", err)
	}
	defer func() { _ = resp2.Body.Close() }()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp2.StatusCode)
	}

	var listResp management.ListServicesResponse
	if err := json.NewDecoder(resp2.Body).Decode(&listResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
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

	if err := reg.Register("foo", "localhost:3000", false); err != nil {
		t.Fatal(err)
	}

	// DELETE /services/foo
	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/services/foo", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /services/foo failed: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}

	// GET /services で空確認
	req2, _ := http.NewRequest(http.MethodGet, ts.URL+"/services", nil)
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("GET /services failed: %v", err)
	}
	defer func() { _ = resp2.Body.Close() }()

	var listResp management.ListServicesResponse
	if err := json.NewDecoder(resp2.Body).Decode(&listResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
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
	_ = resp.Body.Close()

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
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// --- タスク 3.2: コンテナ環境を想定した統合テスト ---

func TestContainerEnv_LocalgatetestRoutesToManagementAPI_POST(t *testing.T) {
	ts, _ := newTestServerWithHostname("localgate.test")
	defer ts.Close()

	body := `{"name":"svc","target":"localhost:3001"}`
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/services", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Host = "localgate.test:9000"

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}

func TestContainerEnv_LocalgatetestRoutesToManagementAPI_GET(t *testing.T) {
	ts, _ := newTestServerWithHostname("localgate.test")
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/services", nil)
	req.Host = "localgate.test:9000"

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestContainerEnv_SubdomainOfLocalgateProxied(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	ts, reg := newTestServerWithHostname("localgate.test")
	defer ts.Close()

	backendAddr := strings.TrimPrefix(backend.URL, "http://")
	if err := reg.Register("foo", backendAddr, false); err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/", nil)
	req.Host = "foo.localgate.test:9000"

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 (proxied), got %d", resp.StatusCode)
	}
}

func TestContainerEnv_LocalhostStillRoutesToManagementAPI(t *testing.T) {
	ts, _ := newTestServerWithHostname("localgate.test")
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/services", nil)
	req.Host = "localhost:9000"

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestContainerEnv_WithoutHostnameFlagLocalgateTestIsNotManagementAPI(t *testing.T) {
	ts, _ := newTestServerWithHostname("")
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/services", nil)
	req.Host = "localgate.test:9000"

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	_ = resp.Body.Close()

	// --hostname 未指定時は localgate.test は管理APIへルーティングされない
	if resp.StatusCode == http.StatusOK {
		t.Errorf("without --hostname, localgate.test should not route to management API")
	}
}
