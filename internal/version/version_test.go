package version

import (
	"runtime"
	"strings"
	"testing"
)

func TestFormatOutput_DefaultValues(t *testing.T) {
	// デフォルト値でのフォーマット確認
	Version = "dev"
	Commit = "unknown"
	BuildDate = "unknown"

	out := FormatOutput()

	if !strings.Contains(out, "Version:    dev") {
		t.Errorf("expected 'Version:    dev' in output, got: %q", out)
	}
	if !strings.Contains(out, "Commit:     unknown") {
		t.Errorf("expected 'Commit:     unknown' in output, got: %q", out)
	}
	if !strings.Contains(out, "Build Date: unknown") {
		t.Errorf("expected 'Build Date: unknown' in output, got: %q", out)
	}
}

func TestFormatOutput_GoVersion(t *testing.T) {
	// Go バージョンの出力確認
	out := FormatOutput()

	goVer := runtime.Version()
	if !strings.Contains(out, "Go Version: "+goVer) {
		t.Errorf("expected 'Go Version: %s' in output, got: %q", goVer, out)
	}
}

func TestFormatOutput_CustomValues(t *testing.T) {
	// 任意の値でのフォーマット確認
	Version = "v1.2.3"
	Commit = "abc1234"
	BuildDate = "2026-03-08T00:00:00Z"
	defer func() {
		Version = "dev"
		Commit = "unknown"
		BuildDate = "unknown"
	}()

	out := FormatOutput()

	if !strings.Contains(out, "Version:    v1.2.3") {
		t.Errorf("expected 'Version:    v1.2.3' in output, got: %q", out)
	}
	if !strings.Contains(out, "Commit:     abc1234") {
		t.Errorf("expected 'Commit:     abc1234' in output, got: %q", out)
	}
	if !strings.Contains(out, "Build Date: 2026-03-08T00:00:00Z") {
		t.Errorf("expected 'Build Date: 2026-03-08T00:00:00Z' in output, got: %q", out)
	}
}

func TestFormatOutput_NoTrailingNewline(t *testing.T) {
	// 末尾に余分な改行がないことを確認
	out := FormatOutput()

	if strings.HasSuffix(out, "\n") {
		t.Errorf("output must not end with newline, got: %q", out)
	}
}

func TestFormatOutput_FourLines(t *testing.T) {
	// 出力が 4 行であることを確認
	out := FormatOutput()

	lines := strings.Split(out, "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 lines, got %d: %q", len(lines), out)
	}
}
