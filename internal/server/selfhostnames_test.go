package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"localgate/internal/registry"
	"localgate/internal/server"
)

func newTestServerWithHostname(hostname string) (*httptest.Server, registry.ServiceRegistry) {
	reg := registry.NewServiceRegistry()
	srv := server.NewProxyServer(server.ServerConfig{Port: 0, Hostname: hostname}, reg)
	return httptest.NewServer(srv), reg
}

// TestSelfHostnames_DefaultLocalhostOnly: ホスト名未指定時に localhost のみが管理APIホスト名として扱われる
func TestSelfHostnames_DefaultLocalhostOnly(t *testing.T) {
	ts, _ := newTestServerWithHostname("")
	defer ts.Close()

	// localhost は管理APIへ
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/services", nil)
	req.Host = "localhost"
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("localhost should route to management API, got %d", resp.StatusCode)
	}

	// --hostname 未指定時に localgate.test は管理APIへルーティングされない（プロキシ判定 → 404）
	req2, _ := http.NewRequest(http.MethodGet, ts.URL+"/services", nil)
	req2.Host = "localgate.test:9000"
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	_ = resp2.Body.Close()
	if resp2.StatusCode != http.StatusNotFound {
		t.Errorf("localgate.test without --hostname should not route to management API, got %d", resp2.StatusCode)
	}
}

// TestSelfHostnames_WithCustomHostname: --hostname=localgate.test 指定時に localhost と localgate.test の両方が管理APIホスト名として扱われる
func TestSelfHostnames_WithCustomHostname(t *testing.T) {
	ts, _ := newTestServerWithHostname("localgate.test")
	defer ts.Close()

	// localhost は管理APIへ
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/services", nil)
	req.Host = "localhost"
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("localhost should route to management API, got %d", resp.StatusCode)
	}

	// localgate.test も管理APIへ
	req2, _ := http.NewRequest(http.MethodGet, ts.URL+"/services", nil)
	req2.Host = "localgate.test:9000"
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	_ = resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("localgate.test with --hostname should route to management API, got %d", resp2.StatusCode)
	}
}

// TestSelfHostnames_CaseInsensitive: 大文字を含むホスト名が正規化されてセットに格納される
func TestSelfHostnames_CaseInsensitive(t *testing.T) {
	ts, _ := newTestServerWithHostname("Localgate.TEST")
	defer ts.Close()

	// localgate.test（小文字）でアクセスしても管理APIへ
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/services", nil)
	req.Host = "localgate.test:9000"
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("case-normalized hostname should route to management API, got %d", resp.StatusCode)
	}
}
