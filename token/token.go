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
	Illegal = "Illegal"
	Eof     = "Eof"

	// Identifiers + literals
	Ident     = "Ident"     // add, foobar, x, y, ...
	Number    = "Number"    // 1343456
	String    = "String"    // "Hello World"
	Character = "Character" // '1'

	// Operators
	Assign     = "="
	Plus       = "+"
	Minus      = "-"
	Asterisk   = "*"
	Slash      = "/"
	Percentage = "%"
	Bang       = "!"
	Dot        = "."

	Lt = "<"
	Gt = ">"

	Eq    = "=="
	NotEq = "!="

	PlusEq       = "+="
	MinusEq      = "-="
	AsteriskEq   = "*="
	SlashEq      = "/="
	PercentageEq = "%="

	LtEq = "<="
	GtEq = ">="

	// Delimiters
	Comma     = ","
	Semicolon = ";"
	Colon     = ":"

	Lparen   = "("
	Rparen   = ")"
	Lbrace   = "{"
	Rbrace   = "}"
	Lbracket = "["
	Rbracket = "]"

	// Keywords
	Function  = "Function"
	Let       = "Let"
	Ref       = "Ref"
	True      = "True"
	False     = "False"
	Void      = "Void"
	If        = "If"
	Else      = "Else"
	Loop      = "Loop"
	Out       = "Out"
	Jump      = "Jump"
	Ret       = "Ret"
	Underline = "Underline"
	In        = "In"
	And       = "And"
	Or        = "Or"
	Del       = "Del"
)
