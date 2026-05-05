package lexis

import (
	"bufio"
	"io"
	"strings"
)

type Lexer struct {
	r bufio.Reader
}

func New(r io.Reader) *Lexer {
	return &Lexer{*bufio.NewReader(r)}
}

func (l *Lexer) Scan() ([]string, error) {
	res := make([]string, 0, 50)
	b := strings.Builder{}
	for {
		char, _, err := l.r.ReadRune()
		if err == io.EOF {
			return res, err
		} else if err != nil {
			panic("!!!")
			return res, err
		}
		switch {
		case char >= '0' && char <= '9':
			for char >= '0' && char <= '9' {
				_, err = b.WriteRune(char)
				if err != nil {
					panic("!!!")
					return res, err
				}
				char, _, err = l.r.ReadRune()
				if err == io.EOF {
					res = append(res, b.String())
					b.Reset()
					return res, err
				} else if err != nil {
					panic("!!!")
				}
			}
			if char != '.' {
				break
			}
			_, err = b.WriteRune(char)
			if err != nil {
				panic("!!!")
				return res, err
			}
			char, _, err = l.r.ReadRune()
			if err != nil {
				panic("!!!")
				return res, err
			}
			isAtLeastOnce := false
			for char >= '0' && char <= '9' {
				_, err = b.WriteRune(char)
				if err != nil {
					panic("!!!")
					return res, err
				}
				isAtLeastOnce = true
				char, _, err = l.r.ReadRune()
				if err == io.EOF {
					res = append(res, b.String())
					b.Reset()
					return res, err
				} else if err != nil {
					panic("!!!")
				}
			}
			if !isAtLeastOnce {
				panic("!!!")
			}
		case char == ' ':
			continue
		default:
			panic("!!!")
		}
		res = append(res, b.String())
		b.Reset()
		if err == io.EOF {
			return res, nil
		}
	}
}
