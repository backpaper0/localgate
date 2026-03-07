package proxy

import (
	"net"
	"strings"
)

// ExtractSubdomain はHostヘッダからサブドメイン部分を抽出する。
// "foo.localhost:9000" → "foo"
// "localhost:9000"     → ""
func ExtractSubdomain(host string) string {
	if host == "" {
		return ""
	}

	// ポート番号を除去
	hostname, _, err := net.SplitHostPort(host)
	if err != nil {
		// ポートなしの場合はそのまま
		hostname = host
	}

	// IPアドレスの場合はサブドメインなし
	if net.ParseIP(hostname) != nil {
		return ""
	}

	parts := strings.Split(hostname, ".")
	if len(parts) < 2 {
		return ""
	}

	// サブドメインは最初のラベル（ドットより前）
	return parts[0]
}
