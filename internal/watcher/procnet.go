package watcher

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// PortScanner は LISTEN 状態の TCP ポート一覧を返すインターフェース。
type PortScanner interface {
	// Scan は現在 LISTEN 中の TCP ポート番号のスライスを返す。
	// 重複なし、順序不定。
	Scan() ([]int, error)
}

// ProcNetScanner は /proc/net/tcp[6] を使用する PortScanner 実装。
// Linux 専用。
type ProcNetScanner struct{}

// NewProcNetScanner は ProcNetScanner を生成する。
func NewProcNetScanner() *ProcNetScanner {
	return &ProcNetScanner{}
}

// Scan は /proc/net/tcp および /proc/net/tcp6 を読み取り、
// LISTEN 状態のポート番号リストを返す。
func (s *ProcNetScanner) Scan() ([]int, error) {
	tcp4Ports, err := scanFile("/proc/net/tcp")
	if err != nil {
		return nil, fmt.Errorf("/proc/net/tcp の読み取りに失敗: %w", err)
	}

	tcp6Ports, err := scanFile("/proc/net/tcp6")
	if err != nil {
		// IPv6 が無効な環境では tcp6 が存在しないことがある
		fmt.Fprintf(os.Stderr, "watch: /proc/net/tcp6 の読み取りをスキップ: %v\n", err)
		return tcp4Ports, nil
	}

	return mergePorts(tcp4Ports, tcp6Ports), nil
}

// scanFile はファイルパスを開いて parseProcNetTCP を呼び出す。
func scanFile(path string) ([]int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseProcNetTCP(f)
}

// parseProcNetTCP は /proc/net/tcp[6] 形式のリーダーを読み取り、
// LISTEN 状態（st=0A）のポート番号スライスを返す。
func parseProcNetTCP(r io.Reader) ([]int, error) {
	scanner := bufio.NewScanner(r)
	portSet := make(map[int]struct{})

	// 先頭行はヘッダなのでスキップ
	if !scanner.Scan() {
		return nil, nil
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		// フォーマット: sl local_address rem_address st ...
		// インデックス: 0  1             2            3
		if len(fields) < 4 {
			continue
		}
		st := fields[3]
		if st != "0A" {
			continue
		}
		// local_address は "HEXIP:HEXPORT"
		localAddr := fields[1]
		colonIdx := strings.LastIndex(localAddr, ":")
		if colonIdx < 0 {
			continue
		}
		hexPort := localAddr[colonIdx+1:]
		port, err := strconv.ParseUint(hexPort, 16, 32)
		if err != nil {
			continue
		}
		portSet[int(port)] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	ports := make([]int, 0, len(portSet))
	for p := range portSet {
		ports = append(ports, p)
	}
	return ports, nil
}

// mergePorts は2つのポートスライスをマージして重複を除去する。
func mergePorts(a, b []int) []int {
	set := make(map[int]struct{}, len(a)+len(b))
	for _, p := range a {
		set[p] = struct{}{}
	}
	for _, p := range b {
		set[p] = struct{}{}
	}
	result := make([]int, 0, len(set))
	for p := range set {
		result = append(result, p)
	}
	return result
}
