package proxy_test

import (
	"testing"

	"localgate/internal/proxy"
)

func TestExtractSubdomain(t *testing.T) {
	tests := []struct {
		host     string
		expected string
	}{
		{"foo.localhost:9000", "foo"},
		{"bar.localhost:9000", "bar"},
		{"foo.localhost", "foo"},
		{"localhost:9000", ""},
		{"localhost", ""},
		{"", ""},
		{"foo.bar.localhost:9000", "foo"},
		{"127.0.0.1:9000", ""},
		{"127.0.0.1", ""},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			got := proxy.ExtractSubdomain(tt.host)
			if got != tt.expected {
				t.Errorf("ExtractSubdomain(%q) = %q, want %q", tt.host, got, tt.expected)
			}
		})
	}
}
