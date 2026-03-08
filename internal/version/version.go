package version

import (
	"fmt"
	"runtime"
)

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// FormatOutput はバージョン情報をラベル付きテキスト形式で返す（末尾改行なし）。
func FormatOutput() string {
	return fmt.Sprintf(
		"Version:    %s\nCommit:     %s\nBuild Date: %s\nGo Version: %s",
		Version, Commit, BuildDate, runtime.Version(),
	)
}
