package watcher

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// integrationMockScanner は外部から結果を制御できるスキャナー。
type integrationMockScanner struct {
	mu      sync.Mutex
	current []int
}

func (s *integrationMockScanner) set(ports []int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.current = ports
}

func (s *integrationMockScanner) Scan() ([]int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]int, len(s.current))
	copy(result, s.current)
	return result, nil
}

// TestIntegration_FullCycle は新規検出→登録→消滅→解除のフルサイクルを検証する。
func TestIntegration_FullCycle(t *testing.T) {
	type req struct {
		method string
		path   string
		name   string
		target string
	}

	var mu sync.Mutex
	var requests []req

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		record := req{method: r.Method, path: r.URL.Path}
		switch r.Method {
		case http.MethodPost:
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			record.name = body["name"]
			record.target = body["target"]
			w.WriteHeader(http.StatusCreated)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusOK)
		}
		requests = append(requests, record)
	}))
	defer srv.Close()

	scanner := &integrationMockScanner{}
	client := NewManagementHTTPClient(srv.URL)
	w := NewWatcher(scanner, client, 20*time.Millisecond, "testhost.local")

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- w.Run(ctx)
	}()

	// ベースライン後にポート出現
	time.Sleep(30 * time.Millisecond)
	scanner.set([]int{8080})

	// 登録が行われるまで待つ
	time.Sleep(60 * time.Millisecond)

	// ポートを消滅させる
	scanner.set([]int{})
	time.Sleep(60 * time.Millisecond)

	// コンテキストキャンセルでクリーンアップ
	cancel()
	<-done

	mu.Lock()
	defer mu.Unlock()

	// 登録リクエストを確認
	registered := false
	deregistered := 0
	for _, r := range requests {
		if r.method == http.MethodPost && r.name == "testhost-8080" && r.target == "testhost.local:8080" {
			registered = true
		}
		if r.method == http.MethodDelete && r.path == "/services/testhost-8080" {
			deregistered++
		}
	}

	if !registered {
		t.Errorf("testhost-8080 の登録リクエストが送られていない: %v", requests)
	}
	if deregistered == 0 {
		t.Errorf("testhost-8080 の解除リクエストが送られていない: %v", requests)
	}
}

// TestIntegration_ContextCancel_CleansUpAll はキャンセル時に全管理サービスが解除されることを検証する。
func TestIntegration_ContextCancel_CleansUpAll(t *testing.T) {
	var mu sync.Mutex
	deleted := make(map[string]int)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
		case http.MethodDelete:
			deleted[r.URL.Path]++
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	scanner := &integrationMockScanner{}
	client := NewManagementHTTPClient(srv.URL)
	w := NewWatcher(scanner, client, 20*time.Millisecond, "node")

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- w.Run(ctx)
	}()

	// 複数ポートを出現させる
	time.Sleep(30 * time.Millisecond)
	scanner.set([]int{3000, 4000})
	time.Sleep(60 * time.Millisecond)

	// キャンセル → クリーンアップ
	cancel()
	<-done

	mu.Lock()
	defer mu.Unlock()
	if deleted["/services/node-3000"] == 0 {
		t.Errorf("node-3000 の解除リクエストが送られていない: %v", deleted)
	}
	if deleted["/services/node-4000"] == 0 {
		t.Errorf("node-4000 の解除リクエストが送られていない: %v", deleted)
	}
}
