package cmd

import (
	"bytes"
	"runtime"
	"strings"
	"testing"

	"localgate/internal/version"
)

func TestVersionCmd_DefaultOutput(t *testing.T) {
	// デフォルト値でバージョン情報が出力されることを確認
	version.Version = "dev"
	version.Commit = "unknown"
	version.BuildDate = "unknown"

	cmd := newVersionCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Version:    dev") {
		t.Errorf("expected 'Version:    dev' in output, got: %q", output)
	}
	if !strings.Contains(output, "Commit:     unknown") {
		t.Errorf("expected 'Commit:     unknown' in output, got: %q", output)
	}
	if !strings.Contains(output, "Build Date: unknown") {
		t.Errorf("expected 'Build Date: unknown' in output, got: %q", output)
	}
}

func TestVersionCmd_GoVersion(t *testing.T) {
	// Go バージョンが出力に含まれることを確認
	cmd := newVersionCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "Go Version: "+runtime.Version()) {
		t.Errorf("expected Go version in output, got: %q", out.String())
	}
}

func TestVersionCmd_CustomValues(t *testing.T) {
	// 任意の値が出力に反映されることを確認
	version.Version = "v1.0.0"
	version.Commit = "deadbeef"
	version.BuildDate = "2026-03-08T00:00:00Z"
	defer func() {
		version.Version = "dev"
		version.Commit = "unknown"
		version.BuildDate = "unknown"
	}()

	cmd := newVersionCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "v1.0.0") {
		t.Errorf("expected 'v1.0.0' in output, got: %q", output)
	}
	if !strings.Contains(output, "deadbeef") {
		t.Errorf("expected 'deadbeef' in output, got: %q", output)
	}
	if !strings.Contains(output, "2026-03-08T00:00:00Z") {
		t.Errorf("expected build date in output, got: %q", output)
	}
}

func TestVersionCmd_NoError(t *testing.T) {
	// コマンドが正常終了することを確認
	cmd := newVersionCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}
