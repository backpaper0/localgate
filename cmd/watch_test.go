package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWatchCmd_ServerFlagDefault(t *testing.T) {
	cmd := newWatchCmd()
	flag := cmd.Flags().Lookup("server")
	if flag == nil {
		t.Fatal("--server フラグが定義されていない")
	}
	if flag.DefValue != "" {
		t.Errorf("--server デフォルト値: got %q, want %q", flag.DefValue, "")
	}
}

func TestWatchCmd_IntervalFlag(t *testing.T) {
	cmd := newWatchCmd()
	flag := cmd.Flags().Lookup("interval")
	if flag == nil {
		t.Fatal("--interval フラグが定義されていない")
	}
	if flag.DefValue != "1" {
		t.Errorf("--interval デフォルト値: got %q, want %q", flag.DefValue, "1")
	}
}

func TestWatchCmd_IntervalValidation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cmd := newWatchCmd()
	cmd.SetArgs([]string{"--server", srv.URL, "--interval", "0"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("interval=0 ではバリデーションエラーが期待されたが nil が返った")
	}
}

func TestWatchCmd_ServerConnectionFailure(t *testing.T) {
	cmd := newWatchCmd()
	cmd.SetArgs([]string{"--server", "http://127.0.0.1:1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("接続失敗時にエラーが期待されたが nil が返った")
	}
}
