package watcher

import (
	"strings"
	"testing"
)

// テスト用の /proc/net/tcp 形式コンテンツ（ヘッダ行 + データ行）
const procNetTCPContent = `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 0100007F:0035 00000000:0000 0A 00000000:00000000 00:00000000 00000000   101        0 12345 1 0000000000000000 100 0 0 10 0
   1: 0F02000A:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 23456 1 0000000000000000 100 0 0 10 0
   2: 0100007F:0050 12345678:9999 06 00000000:00000000 00:00000000 00000000     0        0 34567 1 0000000000000000 100 0 0 10 0
   3: 0000007F:1388 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 45678 1 0000000000000000 100 0 0 10 0
`

// 0A=LISTEN: port 0x0035=53, 0x1F90=8080, 0x1388=5000
// 0x0050=80 は st=06(ESTABLISHED) なのでスキップ

const procNetTCP6Content = `  sl  local_address                         remote_address                        st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 00000000000000000000000001000000:1F90 00000000000000000000000000000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 56789 1 0000000000000000 100 0 0 10 0
   1: 00000000000000000000000001000000:0050 00000000000000000000000000000000:0000 06 00000000:00000000 00:00000000 00000000     0        0 67890 1 0000000000000000 100 0 0 10 0
`

// tcp6 の LISTEN: port 0x1F90=8080 (tcp4 と重複)

func TestParseProcNetTCP_ListenPorts(t *testing.T) {
	ports, err := parseProcNetTCP(strings.NewReader(procNetTCPContent))
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	want := map[int]bool{53: true, 8080: true, 5000: true}
	if len(ports) != len(want) {
		t.Errorf("ポート数: got %d, want %d (ports=%v)", len(ports), len(want), ports)
	}
	for _, p := range ports {
		if !want[p] {
			t.Errorf("予期しないポート: %d", p)
		}
	}
}

func TestParseProcNetTCP_EmptyContent(t *testing.T) {
	// ヘッダ行のみ
	content := "  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode\n"
	ports, err := parseProcNetTCP(strings.NewReader(content))
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if len(ports) != 0 {
		t.Errorf("ポートなしを期待したが got %v", ports)
	}
}

func TestMergePorts_Dedup(t *testing.T) {
	a := []int{53, 8080, 5000}
	b := []int{8080, 9000} // 8080 は重複
	result := mergePorts(a, b)
	want := map[int]bool{53: true, 8080: true, 5000: true, 9000: true}
	if len(result) != len(want) {
		t.Errorf("マージ後ポート数: got %d, want %d (ports=%v)", len(result), len(want), result)
	}
	for _, p := range result {
		if !want[p] {
			t.Errorf("予期しないポート: %d", p)
		}
	}
}

func TestParseProcNetTCP_IPv6(t *testing.T) {
	ports, err := parseProcNetTCP(strings.NewReader(procNetTCP6Content))
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	// LISTENは 0x1F90=8080 のみ (0x0050=80はESTABLISHED)
	if len(ports) != 1 || ports[0] != 8080 {
		t.Errorf("got %v, want [8080]", ports)
	}
}
