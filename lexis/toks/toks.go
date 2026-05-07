package toks

import (
	"fmt"

	"json-parser/lexis/toks/types"
)

type Token struct {
	Lexeme string
	Type   types.Type
}

func (t *Token) String() string {
	return fmt.Sprintf(
		"Token{Lexeme: \"%s\", Type: %s}",
		t.Lexeme, t.Type)
}

var (
	Comma = &Token{Lexeme: ",", Type: types.Comma}

	LeftBrace  = &Token{Lexeme: "{", Type: types.LeftBrace}
	RightBrace = &Token{Lexeme: "}", Type: types.RightBrace}

	LeftBracket  = &Token{Lexeme: "[", Type: types.LeftBracket}
	RightBracket = &Token{Lexeme: "]", Type: types.RightBracket}
)

var (
	True  = &Token{Lexeme: "true", Type: types.Boolean}
	False = &Token{Lexeme: "false", Type: types.Boolean}
)

func NewNumber(lexeme string) *Token {
	return &Token{Lexeme: lexeme, Type: types.Number}
}

func NewString(lexeme string) *Token {
	return &Token{Lexeme: lexeme, Type: types.String}
}
