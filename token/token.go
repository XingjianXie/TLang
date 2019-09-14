package token

type Type string

type Token struct {
	Type    Type
	Literal string
}

var keywords = map[string]Type{
	"func":  FUNCTION,
	"let":   LET,
	"ref":   REF,
	"true":  TRUE,
	"false": FALSE,
	"if":    IF,
	"else":  ELSE,
	"ret":   RET,
	"del":   DEL,
	"and":   AND,
	"or":    OR,
	"_":     UNDERLINE,
}

func LookupIdent(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	IDENT  = "IDENT"  // add, foobar, x, y, ...
	NUMBER = "NUMBER" // 1343456
	STRING = "STRING" // "Hello World"

	// Operators
	ASSIGN     = "="
	PLUS       = "+"
	MINUS      = "-"
	ASTERISK   = "*"
	SLASH      = "/"
	PERCENTAGE = "%"
	BANG       = "!"
	DOT        = "."

	LT = "<"
	GT = ">"

	EQ     = "=="
	NOT_EQ = "!="

	PLUS_EQ       = "+="
	MINUS_EQ      = "-="
	ASTERISK_EQ   = "*="
	SLASH_EQ      = "/="
	PERCENTAGE_EQ = "%="

	LT_EQ = "<="
	GT_EQ = ">="

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// Keywords
	FUNCTION  = "FUNCTION"
	LET       = "LET"
	REF       = "REF"
	TRUE      = "TRUE"
	FALSE     = "FALSE"
	IF        = "IF"
	ELSE      = "ELSE"
	RET       = "RET"
	UNDERLINE = "UNDERLINE"
	AND       = "AND"
	OR        = "OR"
	DEL       = "DEL"

	// Invisible
	INVISIBLE = ""
)

func MakeInvisible() Token {
	return Token{
		Type:    INVISIBLE,
		Literal: "",
	}
}
