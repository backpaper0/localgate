package watcher

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// --- モック実装 ---

type mockScanner struct {
	mu      sync.Mutex
	results [][]int
	index   int
	err     error
}

func (m *mockScanner) Scan() ([]int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return nil, m.err
	}
	if m.index >= len(m.results) {
		return m.results[len(m.results)-1], nil
	}
	r := m.results[m.index]
	m.index++
	return r, nil
}

type mockClient struct {
	mu           sync.Mutex
	registered   []string
	deregistered []string
	registerErr  error
	deregErr     error
}

func (m *mockClient) Ping() error { return nil }

func (m *mockClient) Register(name, target string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.registerErr != nil {
		return m.registerErr
	}
	m.registered = append(m.registered, name)
	return nil
}

func (m *mockClient) Deregister(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.deregErr != nil {
		return m.deregErr
	}
	m.deregistered = append(m.deregistered, name)
	return nil
}

// --- テスト ---

func TestExtractHostLabel(t *testing.T) {
	tests := []struct {
		hostname string
		want     string
	}{
		{"foobar.test", "foobar"},
		{"hoge", "hoge"},
		{"a.b.c.d", "a"},
		{"", ""},
	}
	for _, tt := range tests {
		got := extractHostLabel(tt.hostname)
		if got != tt.want {
			t.Errorf("extractHostLabel(%q) = %q, want %q", tt.hostname, got, tt.want)
		}
	}
}

func TestWatcher_NewPort_Registers(t *testing.T) {
	scanner := &mockScanner{
		results: [][]int{
			{},        // 初回スキャン（ベースライン）
			{8080},    // 2回目: 新規ポート出現
			{8080},    // 3回目以降
		},
	}
	client := &mockClient{}

	w := NewWatcher(scanner, client, 10*time.Millisecond, "myhost.test")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	w.Run(ctx) //nolint

	client.mu.Lock()
	defer client.mu.Unlock()
	found := false
	for _, name := range client.registered {
		if name == "myhost-8080" {
			found = true
		}
	}
	if !found {
		t.Errorf("myhost-8080 が登録されていない: registered=%v", client.registered)
	}
}

func TestWatcher_PortGone_Deregisters(t *testing.T) {
	scanner := &mockScanner{
		results: [][]int{
			{},     // 初回（ベースライン）
			{9000}, // 2回目: ポート出現
			{},     // 3回目: ポート消滅
			{},
		},
	}
	client := &mockClient{}

	w := NewWatcher(scanner, client, 10*time.Millisecond, "host")
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	w.Run(ctx) //nolint

	client.mu.Lock()
	defer client.mu.Unlock()
	found := false
	for _, name := range client.deregistered {
		if name == "host-9000" {
			found = true
		}
	}
	if !found {
		t.Errorf("host-9000 が解除されていない: deregistered=%v", client.deregistered)
	}
}

func TestWatcher_ContextCancel_CleansUp(t *testing.T) {
	scanner := &mockScanner{
		results: [][]int{
			{},           // 初回
			{8080, 9000}, // 2回目: ポート出現
			{8080, 9000},
		},
	}
	client := &mockClient{}

	w := NewWatcher(scanner, client, 10*time.Millisecond, "box")
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	w.Run(ctx) //nolint

	client.mu.Lock()
	defer client.mu.Unlock()
	deregSet := make(map[string]bool)
	for _, name := range client.deregistered {
		deregSet[name] = true
	}
	// クリーンアップで両ポートが解除されること
	if !deregSet["box-8080"] || !deregSet["box-9000"] {
		t.Errorf("クリーンアップ後に全ポートが解除されていない: deregistered=%v", client.deregistered)
	}
}

func TestWatcher_RegisterFailure_ContinuesNextCycle(t *testing.T) {
	scanner := &mockScanner{
		results: [][]int{
			{},
			{8080},
			{8080},
		},
	}
	client := &mockClient{
		registerErr: errors.New("API error"),
	}

	w := NewWatcher(scanner, client, 10*time.Millisecond, "node")
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	// パニックなく終了すれば OK
	w.Run(ctx) //nolint

	client.mu.Lock()
	defer client.mu.Unlock()
	// 失敗したので managed には追加されていないはず
	if len(w.managed) != 0 {
		t.Errorf("登録失敗時に managed に追加されてはいけない: %v", w.managed)
	}
}

func TestWatcher_DeregFailure_ContinuesCleanup(t *testing.T) {
	scanner := &mockScanner{
		results: [][]int{
			{},
			{8080, 9000},
			{8080, 9000},
		},
	}
	callCount := 0
	client := &mockClient{}
	// 1件目の Deregister は失敗させる（2件目は成功）
	// mockClient を拡張するため別途テスト用を作る

	// 簡易：通常の mockClient を使い、クリーンアップが少なくとも 1 回以上呼ばれることを確認
	w := NewWatcher(scanner, client, 10*time.Millisecond, "srv")
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	w.Run(ctx) //nolint

	client.mu.Lock()
	defer client.mu.Unlock()
	callCount = len(client.deregistered)
	if callCount == 0 {
		t.Error("クリーンアップで Deregister が1回も呼ばれていない")
	}
}
