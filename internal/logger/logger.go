package logger

import (
	"log/slog"
	"os"
)

var l *slog.Logger

func init() {
	level := slog.LevelInfo
	if os.Getenv("LOCALGATE_DEBUG") == "1" {
		level = slog.LevelDebug
	}
	l = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
}

// SetDebug はデバッグログの有効/無効を切り替える。
// --debug フラグから呼び出される。
func SetDebug(enabled bool) {
	level := slog.LevelInfo
	if enabled {
		level = slog.LevelDebug
	}
	l = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
}

func Debug(msg string, args ...any) { l.Debug(msg, args...) }
func Info(msg string, args ...any)  { l.Info(msg, args...) }
func Warn(msg string, args ...any)  { l.Warn(msg, args...) }
func Error(msg string, args ...any) { l.Error(msg, args...) }
