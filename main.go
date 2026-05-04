package main

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
)

func main() {
	setLogger()
}

func setLogger() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				fn := source.Function
				if i := strings.LastIndex(fn, "/"); i >= 0 {
					fn = fn[i+1:]
				}
				if i := strings.Index(fn, "."); i >= 0 {
					fn = fn[i+1:]
				}
				a.Value = slog.StringValue(fn + ":" + strconv.Itoa(source.Line))
			}
			return a
		},
	}))
	slog.SetDefault(logger)
}
