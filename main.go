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
	setLogger()

	l := lexis.New(strings.NewReader("0"))
	fmt.Println(l.Scan())

	l = lexis.New(strings.NewReader("1"))
	fmt.Println(l.Scan())

	l = lexis.New(strings.NewReader("02"))
	fmt.Println(l.Scan())

	// l = lexis.New(strings.NewReader("0."))
	// fmt.Println(l.Scan())

	// l = lexis.New(strings.NewReader("1."))
	// fmt.Println(l.Scan())

	// l = lexis.New(strings.NewReader("02."))
	// fmt.Println(l.Scan())

	l = lexis.New(strings.NewReader("102"))
	fmt.Println(l.Scan())

	l = lexis.New(strings.NewReader("102.2"))
	fmt.Println(l.Scan())

	l = lexis.New(strings.NewReader("102.22"))
	fmt.Println(l.Scan())

	l = lexis.New(strings.NewReader("102.220"))
	fmt.Println(l.Scan())

	l = lexis.New(strings.NewReader("102.220 3"))
	fmt.Println(l.Scan())

	l = lexis.New(strings.NewReader("102.220 33"))
	fmt.Println(l.Scan())

	l = lexis.New(strings.NewReader("102.220 333"))
	fmt.Println(l.Scan())

	l = lexis.New(strings.NewReader("102.220 333 3.3"))
	fmt.Println(l.Scan())

	l = lexis.New(strings.NewReader("102.220 333 3.3 "))
	fmt.Println(l.Scan())

	l = lexis.New(strings.NewReader(" 102.220   333  3.3  "))
	fmt.Println(l.Scan())

	l = lexis.New(strings.NewReader("  102.220   333  3.3"))
	fmt.Println(l.Scan())
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
