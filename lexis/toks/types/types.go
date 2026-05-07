package types

type Type string

const (
	String  Type = "string"
	Number  Type = "number"
	Boolean Type = "boolean"
)

const (
	LeftBrace  Type = "opening brace"
	RightBrace Type = "closing brace"

	LeftBracket  Type = "opening bracket"
	RightBracket Type = "closing bracket"
)

const Comma Type = "comma"
