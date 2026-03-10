package watcher

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestManagementHTTPClient_Ping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/services" {
			t.Errorf("予期しないリクエスト: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewManagementHTTPClient(srv.URL)
	if err := c.Ping(); err != nil {
		t.Fatalf("Ping() error: %v", err)
	}
}

func TestManagementHTTPClient_Ping_Failure(t *testing.T) {
	c := NewManagementHTTPClient("http://127.0.0.1:1") // 接続できないポート
	if err := c.Ping(); err == nil {
		t.Fatal("接続失敗時にエラーが期待されたが nil が返った")
	}
}

func TestManagementHTTPClient_Register(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c := NewManagementHTTPClient(srv.URL)
	if err := c.Register("test-8080", "myhost:8080"); err != nil {
		t.Fatalf("Register() error: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("メソッド: got %s, want POST", gotMethod)
	}
	if gotPath != "/services" {
		t.Errorf("パス: got %s, want /services", gotPath)
	}
	if gotBody["name"] != "test-8080" || gotBody["target"] != "myhost:8080" {
		t.Errorf("ボディ: got %v", gotBody)
	}
}

func TestManagementHTTPClient_Register_Conflict(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}))
	defer srv.Close()

	c := NewManagementHTTPClient(srv.URL)
	if err := c.Register("test-8080", "myhost:8080"); err == nil {
		t.Fatal("409 Conflict 時にエラーが期待されたが nil が返った")
	}
}

func TestManagementHTTPClient_Deregister(t *testing.T) {
	var gotMethod, gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewManagementHTTPClient(srv.URL)
	if err := c.Deregister("test-8080"); err != nil {
		t.Fatalf("Deregister() error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("メソッド: got %s, want DELETE", gotMethod)
	}
	if gotPath != "/services/test-8080" {
		t.Errorf("パス: got %s, want /services/test-8080", gotPath)
	}
}

func TestManagementHTTPClient_Deregister_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewManagementHTTPClient(srv.URL)
	// 404 は正常扱い（冪等性）
	if err := c.Deregister("nonexistent"); err != nil {
		t.Fatalf("404 時にエラーは不要だが got: %v", err)
	}
}
