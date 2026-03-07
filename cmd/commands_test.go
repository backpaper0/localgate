package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// --- registerコマンド ---

func TestRegisterCmd_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/services" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var req registerServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Name != "myapp" || req.Target != "http://localhost:3000" {
			t.Errorf("unexpected request body: %+v", req)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(serviceEntry{Name: req.Name, Target: req.Target})
	}))
	defer srv.Close()

	cmd := newRegisterCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--server", srv.URL, "myapp", "http://localhost:3000"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "myapp") {
		t.Errorf("expected output to contain 'myapp', got: %q", out.String())
	}
}

func TestRegisterCmd_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{Error: "名前が既に登録されています"})
	}))
	defer srv.Close()

	cmd := newRegisterCmd()
	cmd.SetArgs([]string{"--server", srv.URL, "myapp", "http://localhost:3000"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "名前が既に登録されています") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRegisterCmd_ConnectionError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close() // すぐに閉じて接続エラーを発生させる

	cmd := newRegisterCmd()
	cmd.SetArgs([]string{"--server", srv.URL, "myapp", "http://localhost:3000"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected connection error, got nil")
	}
	if !strings.Contains(err.Error(), "接続エラー") {
		t.Errorf("expected connection error message, got: %v", err)
	}
}

func TestRegisterCmd_MissingArgs(t *testing.T) {
	cmd := newRegisterCmd()
	cmd.SetArgs([]string{"myapp"}) // targetが欠落
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing args, got nil")
	}
}

func TestRegisterCmd_ConflictWithConfirmYes(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// 1回目: 409 Conflict を返す
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(apiError{Error: "service already exists", ExistingTarget: "localhost:3000"})
			return
		}
		// 2回目: X-Force-Overwrite ヘッダーを確認して 201 Created を返す
		if r.Header.Get("X-Force-Overwrite") != "true" {
			t.Errorf("expected X-Force-Overwrite: true header on retry")
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(serviceEntry{Name: "myapp", Target: "localhost:4000"})
	}))
	defer srv.Close()

	cmd := newRegisterCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetIn(strings.NewReader("y\n"))
	cmd.SetArgs([]string{"--server", srv.URL, "myapp", "localhost:4000"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 HTTP calls, got %d", callCount)
	}
	if !strings.Contains(out.String(), "登録しました") {
		t.Errorf("expected success message, got: %q", out.String())
	}
}

func TestRegisterCmd_ConflictWithConfirmNo(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(apiError{Error: "service already exists", ExistingTarget: "localhost:3000"})
	}))
	defer srv.Close()

	cmd := newRegisterCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetIn(strings.NewReader("n\n"))
	cmd.SetArgs([]string{"--server", srv.URL, "myapp", "localhost:4000"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected only 1 HTTP call, got %d", callCount)
	}
	if !strings.Contains(out.String(), "キャンセル") {
		t.Errorf("expected cancel message, got: %q", out.String())
	}
}

func TestRegisterCmd_ConflictWithForceFlag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Force-Overwrite") != "true" {
			t.Errorf("expected X-Force-Overwrite: true header with --force flag")
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(serviceEntry{Name: "myapp", Target: "localhost:4000"})
	}))
	defer srv.Close()

	cmd := newRegisterCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--server", srv.URL, "--force", "myapp", "localhost:4000"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "登録しました") {
		t.Errorf("expected success message, got: %q", out.String())
	}
}

func TestRegisterCmd_PortOnly(t *testing.T) {
	hostname, err := os.Hostname()
	if err != nil {
		t.Fatalf("failed to get hostname: %v", err)
	}
	expectedTarget := hostname + ":3000"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req registerServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Target != expectedTarget {
			t.Errorf("expected target %q, got %q", expectedTarget, req.Target)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(serviceEntry{Name: req.Name, Target: req.Target})
	}))
	defer srv.Close()

	cmd := newRegisterCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--server", srv.URL, "myapp", "3000"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "myapp") {
		t.Errorf("expected output to contain 'myapp', got: %q", out.String())
	}
}

// --- unregisterコマンド ---

func TestUnregisterCmd_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/services/myapp" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	cmd := newUnregisterCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--server", srv.URL, "myapp"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "myapp") {
		t.Errorf("expected output to contain 'myapp', got: %q", out.String())
	}
}

func TestUnregisterCmd_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(apiError{Error: "service not found"})
	}))
	defer srv.Close()

	cmd := newUnregisterCmd()
	cmd.SetArgs([]string{"--server", srv.URL, "unknown"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "見つかりません") {
		t.Errorf("expected 'not found' message, got: %v", err)
	}
}

func TestUnregisterCmd_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(apiError{Error: "内部エラーが発生しました"})
	}))
	defer srv.Close()

	cmd := newUnregisterCmd()
	cmd.SetArgs([]string{"--server", srv.URL, "myapp"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "内部エラーが発生しました") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestUnregisterCmd_ConnectionError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	cmd := newUnregisterCmd()
	cmd.SetArgs([]string{"--server", srv.URL, "myapp"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected connection error, got nil")
	}
	if !strings.Contains(err.Error(), "接続エラー") {
		t.Errorf("expected connection error message, got: %v", err)
	}
}

func TestUnregisterCmd_MissingArgs(t *testing.T) {
	cmd := newUnregisterCmd()
	cmd.SetArgs([]string{}) // nameが欠落
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing args, got nil")
	}
}

// --- listコマンド ---

func TestListCmd_WithServices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/services" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(listServicesResponse{
			Services: []serviceEntry{
				{Name: "app1", Target: "http://localhost:3001"},
				{Name: "app2", Target: "http://localhost:3002"},
			},
		})
	}))
	defer srv.Close()

	cmd := newListCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--server", srv.URL})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "app1") || !strings.Contains(output, "app2") {
		t.Errorf("expected service names in output, got: %q", output)
	}
	if !strings.Contains(output, "http://localhost:3001") {
		t.Errorf("expected target URLs in output, got: %q", output)
	}
}

func TestListCmd_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(listServicesResponse{Services: []serviceEntry{}})
	}))
	defer srv.Close()

	cmd := newListCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--server", srv.URL})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "登録済みサービスはありません") {
		t.Errorf("expected empty message, got: %q", out.String())
	}
}

func TestListCmd_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(apiError{Error: "サーバーエラー"})
	}))
	defer srv.Close()

	cmd := newListCmd()
	cmd.SetArgs([]string{"--server", srv.URL})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "サーバーエラー") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestListCmd_ConnectionError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	cmd := newListCmd()
	cmd.SetArgs([]string{"--server", srv.URL})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected connection error, got nil")
	}
	if !strings.Contains(err.Error(), "接続エラー") {
		t.Errorf("expected connection error message, got: %v", err)
	}
}

func TestListCmd_ExtraArgs(t *testing.T) {
	cmd := newListCmd()
	cmd.SetArgs([]string{"unexpected-arg"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for extra args, got nil")
	}
}
