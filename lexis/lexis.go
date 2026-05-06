package lexis

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"json-parser/common/ascii"
	"json-parser/common/cnsmr"
	"json-parser/common/runes"
)

type Lexer struct {
	c *cnsmr.RuneConsumer
	b interface {
		String() string
		Reset()
	}
}

func New(r io.RuneReader) (*Lexer, error) {
	b := &strings.Builder{}
	c, err := cnsmr.NewRuneConsumer(r, b)
	if err != nil {
		return nil, err
	}
	return &Lexer{c, b}, nil
}

func (l *Lexer) Scan() (res []string, err error) {
	res = make([]string, 0, 50)

	defer l.recover(false, &err)
	for char := l.c.Char(); err == nil; char = l.c.Char() {
		switch {
		case ascii.IsDigit(char):
			err = l.handleNumber()
		case char == '"':
			err = l.handleString()
		case char == ' ':
			err = l.c.AtMin(1, runes.Is(' '))
		default:
			panic(fmt.Sprintf("unexpected char '%c'", char))
		}
		res = append(res, l.b.String())
		l.b.Reset()
	}

	if err != io.EOF {
		panic("must not panic here. " + err.Error())
	}
	return res, nil
}

func (l *Lexer) handleNumber() error {
	err := l.c.AtMin(1, ascii.IsDigit)
	if err != nil {
		panic(err)
	}
	err = l.c.Exact(1, runes.Is('.'))
	if err != nil {
		panic(err)
	}
	err = l.c.AtMin(1, ascii.IsDigit)
	if err != nil {
		if err == io.EOF {
			return io.EOF
		}
		panic(err)
	}
	return nil
}

func (l *Lexer) handleString() (err error) {
	err = l.c.Exact(1, runes.Is('"'))
	if err != nil {
		if errors.Is(err, cnsmr.ErrCondition) {
			l.c.Reset()
		} else {
			panic(err)
		}
	}
	err = l.c.AtMin(1, runes.IsNot('"'))
	if err != nil {
		if errors.Is(err, cnsmr.ErrCondition) {
			l.c.Reset()
		} else {
			panic(err)
		}
	}
	err = l.c.Exact(1, runes.Is('"'))
	if err != nil {
		panic(err)
	}
	return nil
}

func (l *Lexer) recover(can bool, err *error) {
	if !can {
		return
	}
	if e := recover().(error); e != nil {
		*err = e
	}
}
