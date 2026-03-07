package cmd

import (
	"os"
	"testing"
)

func TestResolveServerURL(t *testing.T) {
	tests := []struct {
		name      string
		flagValue string
		envValue  string
		want      string
	}{
		{
			name:      "フラグ値が指定されている場合はフラグ値を返す",
			flagValue: "http://example.com:9000",
			envValue:  "",
			want:      "http://example.com:9000",
		},
		{
			name:      "フラグ値が空でLOCALGATE_SERVERが設定されている場合は環境変数の値を返す",
			flagValue: "",
			envValue:  "http://remote:9000",
			want:      "http://remote:9000",
		},
		{
			name:      "どちらも未設定の場合はデフォルト値を返す",
			flagValue: "",
			envValue:  "",
			want:      "http://localhost:9000",
		},
		{
			name:      "フラグ値が環境変数より優先される",
			flagValue: "http://flag.example.com",
			envValue:  "http://env.example.com",
			want:      "http://flag.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("LOCALGATE_SERVER")
			if tt.envValue != "" {
				os.Setenv("LOCALGATE_SERVER", tt.envValue)
				defer os.Unsetenv("LOCALGATE_SERVER")
			}

			got := resolveServerURL(tt.flagValue)
			if got != tt.want {
				t.Errorf("resolveServerURL(%q) = %q, want %q", tt.flagValue, got, tt.want)
			}
		})
	}
}
