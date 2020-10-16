package token

type Type string

type Token struct {
	Type    Type
	Literal string
}

var keywords = map[string]Type{
	"func":  Function,
	"let":   Let,
	"ref":   Ref,
	"true":  True,
	"false": False,
	"void":  Void,
	"if":    If,
	"else":  Else,
	"loop":  Loop,
	"out":   Out,
	"jump":  Jump,
	"ret":   Ret,
	"del":   Del,
	"and":   And,
	"or":    Or,
	"in":    In,
	"_":     Underline,
}

func LookupIdent(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return Ident
}

const (
	Illegal Type = "Illegal"
	Eof     Type = "Eof"

	// Identifiers + literals
	Ident     Type = "Ident"     // add, foobar, x, y, ...
	Number    Type = "Number"    // 1343456
	String    Type = "String"    // "Hello World"
	Character Type = "Character" // '1'

	// Operators
	Assign     Type = "="
	Plus       Type = "+"
	Minus      Type = "-"
	Asterisk   Type = "*"
	Slash      Type = "/"
	Percentage Type = "%"
	Bang       Type = "!"
	Dot        Type = "."

	Lt Type = "<"
	Gt Type = ">"

	Eq    Type = "=="
	NotEq Type = "!="

	PlusEq       Type = "+="
	MinusEq      Type = "-="
	AsteriskEq   Type = "*="
	SlashEq      Type = "/="
	PercentageEq Type = "%="

	LtEq Type = "<="
	GtEq Type = ">="

	// Delimiters
	Comma     Type = ","
	Semicolon Type = ";"
	Colon     Type = ":"

	Lparen   Type = "("
	Rparen   Type = ")"
	Lbrace   Type = "{"
	Rbrace   Type = "}"
	Lbracket Type = "["
	Rbracket Type = "]"

	// Keywords
	Function  Type = "Function"
	Let       Type = "Let"
	Ref       Type = "Ref"
	True      Type = "True"
	False     Type = "False"
	Void      Type = "Void"
	If        Type = "If"
	Else      Type = "Else"
	Loop      Type = "Loop"
	Out       Type = "Out"
	Jump      Type = "Jump"
	Ret       Type = "Ret"
	Underline Type = "Underline"
	In        Type = "In"
	And       Type = "And"
	Or        Type = "Or"
	Del       Type = "Del"
)
