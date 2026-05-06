package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"json-parser/lexis"
)

func main() {
	l, err := lexis.New(strings.NewReader("123.3 \"\""))
	if err != nil {
		slog.Error(err.Error())
		return
	}

	lexemes, err := l.Scan()
	for i, l := range lexemes {
		fmt.Printf("%d\tlexeme: %s\n", i, l)
	}
	if err != nil {
		fmt.Println(err.Error())
	}
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
