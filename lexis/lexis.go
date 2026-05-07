package lexis

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"json-parser/common/ascii"
	"json-parser/common/cnsmr"
	"json-parser/common/runes"
	"json-parser/lexis/toks"
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

func (l *Lexer) Scan() (res []toks.Token, err error) {
	res = make([]toks.Token, 0, 64)
	var tok *toks.Token

	defer l.recover(false, &err)
	for char := l.c.Char(); err == nil; char = l.c.Char() {
		switch {
		case ascii.IsDigit(char):
			tok, err = l.handleNumber()
		// case char == 't' || char == 'f':
		// 	l.c.Equal(1, "true")
		// 	l.c.Equal(1, "false")
		case char == '"':
			tok, err = l.handleString()
		case char == '{':
			tok = toks.LeftBrace
		case char == '[':
			tok = toks.LeftBracket
		case char == ',':
			tok = toks.Comma
		case char == ' ':
			err = l.c.SkipWhile(runes.Is(' '))
			continue
		default:
			panic(fmt.Sprintf("unexpected char '%c'", char))
		}
		res = append(res, *tok)
		l.b.Reset()
	}

	if err != io.EOF {
		panic("must not panic here. " + err.Error())
	}
	return res, nil
}

func (l *Lexer) handleNumber() (tok *toks.Token, err error) {
	err = l.c.AtMin(1, ascii.IsDigit)
	if err != nil {
		panic(err)
	}
	err = l.c.Exact(1, runes.Is('.'))
	if err != nil {
		panic(err)
	}
	err = l.c.AtMin(1, ascii.IsDigit)
	if err != nil && err != io.EOF {
		panic(err)
	}
	return toks.NewNumber(l.b.String()), err
}

func (l *Lexer) handleString() (tok *toks.Token, err error) {
	err = l.c.Exact(1, runes.Is('"'))
	if err != nil {
		if errors.Is(err, cnsmr.ErrCondition) {
			err = nil
			l.c.Reset()
		} else {
			panic(err)
		}
	}
	err = l.c.AtMin(1, runes.IsNot('"'))
	if err != nil {
		if errors.Is(err, cnsmr.ErrCondition) {
			err = nil
			l.c.Reset()
		} else {
			panic(err)
		}
	}
	err = l.c.Exact(1, runes.Is('"'))
	if err != nil && err != io.EOF {
		panic(err)
	}
	return toks.NewString(l.b.String()), err
}

func (l *Lexer) recover(can bool, err *error) {
	if !can {
		return
	}
	if e := recover().(error); e != nil {
		*err = e
	}
}
