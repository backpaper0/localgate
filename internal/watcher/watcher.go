package watcher

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// Watcher はポート監視と自動登録・解除を行う。
type Watcher struct {
	scanner   PortScanner
	client    ManagementClient
	interval  time.Duration
	hostname  string
	hostLabel string
	managed   map[int]struct{} // このwatcherが登録したポートのセット
}

// NewWatcher は Watcher を生成する。hostname が空の場合は os.Hostname() を使用する。
func NewWatcher(scanner PortScanner, client ManagementClient, interval time.Duration, hostname string) *Watcher {
	if hostname == "" {
		h, err := os.Hostname()
		if err != nil {
			fmt.Fprintf(os.Stderr, "watch: ホスト名の取得に失敗: %v\n", err)
		}
		hostname = h
	}
	return &Watcher{
		scanner:   scanner,
		client:    client,
		interval:  interval,
		hostname:  hostname,
		hostLabel: extractHostLabel(hostname),
		managed:   make(map[int]struct{}),
	}
}

// extractHostLabel はホスト名から先頭ラベルを抽出する。
// "foobar.test" → "foobar"、"hoge" → "hoge"
func extractHostLabel(hostname string) string {
	if idx := strings.Index(hostname, "."); idx >= 0 {
		return hostname[:idx]
	}
	return hostname
}

// serviceName はポート番号からサービス名を生成する。
func (w *Watcher) serviceName(port int) string {
	return fmt.Sprintf("%s-%d", w.hostLabel, port)
}

// target はポート番号からターゲット文字列を生成する。
func (w *Watcher) target(port int) string {
	return fmt.Sprintf("%s:%d", w.hostname, port)
}

// Run はポーリングループを開始し、ctx がキャンセルされると
// クリーンアップ後に return する。
func (w *Watcher) Run(ctx context.Context) error {
	// 初回スキャンでベースラインを取得（起動前のポートは登録しない）
	prev, err := w.scanner.Scan()
	if err != nil {
		fmt.Fprintf(os.Stderr, "watch: 初回スキャンに失敗: %v\n", err)
		prev = []int{}
	}
	prevSet := toSet(prev)

	fmt.Fprintf(os.Stdout, "watch: ポート監視を開始しました (ホスト: %s, 間隔: %v)\n", w.hostname, w.interval)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.cleanup()
			return nil
		case <-ticker.C:
			current, err := w.scanner.Scan()
			if err != nil {
				fmt.Fprintf(os.Stderr, "watch: スキャンに失敗: %v\n", err)
				continue
			}
			currentSet := toSet(current)

			// 新規ポートを登録
			for port := range currentSet {
				if _, existed := prevSet[port]; !existed {
					name := w.serviceName(port)
					target := w.target(port)
					if err := w.client.Register(name, target); err != nil {
						fmt.Fprintf(os.Stderr, "watch: サービス登録に失敗 (port=%d): %v\n", port, err)
						continue
					}
					w.managed[port] = struct{}{}
					fmt.Fprintf(os.Stdout, "watch: サービス登録 %s → %s\n", name, target)
				}
			}

			// 消滅ポートを解除（このwatcherが管理するもののみ）
			for port := range prevSet {
				if _, exists := currentSet[port]; !exists {
					if _, managed := w.managed[port]; managed {
						name := w.serviceName(port)
						if err := w.client.Deregister(name); err != nil {
							fmt.Fprintf(os.Stderr, "watch: サービス解除に失敗 (port=%d): %v\n", port, err)
							continue
						}
						delete(w.managed, port)
						fmt.Fprintf(os.Stdout, "watch: サービス解除 %s\n", name)
					}
				}
			}

			prevSet = currentSet
		}
	}
}

// cleanup は管理済みポートをすべて解除する。
func (w *Watcher) cleanup() {
	fmt.Fprintf(os.Stdout, "watch: クリーンアップを開始しています...\n")
	for port := range w.managed {
		name := w.serviceName(port)
		if err := w.client.Deregister(name); err != nil {
			fmt.Fprintf(os.Stderr, "watch: クリーンアップ中にサービス解除失敗 (port=%d): %v\n", port, err)
			continue
		}
		delete(w.managed, port)
		fmt.Fprintf(os.Stdout, "watch: クリーンアップ: サービス解除 %s\n", name)
	}
	fmt.Fprintf(os.Stdout, "watch: クリーンアップが完了しました\n")
}

// toSet はポートスライスをセット (map) に変換する。
func toSet(ports []int) map[int]struct{} {
	m := make(map[int]struct{}, len(ports))
	for _, p := range ports {
		m[p] = struct{}{}
	}
	return m
}
