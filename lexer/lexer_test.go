package lexer

import (
	"testing"

	"TLang/token"
)

func TestNextToken(t *testing.T) {
	input := `
1.0==1.
.2==-2e-04
3!=3
4!=4
5>=5
6>6
7>!7
8<8/
9/=9
10*=10
"Hello World"
"123"
""
[][]
`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.NUMBER, "1.0"},
		{token.EQ, "=="},
		{token.NUMBER, "1."},

		{token.NUMBER, ".2"},
		{token.EQ, "=="},
		{token.MINUS, "-"},
		{token.NUMBER, "2e-04"},

		{token.NUMBER, "3"},
		{token.NOT_EQ, "!="},
		{token.NUMBER, "3"},

		{token.NUMBER, "4"},
		{token.NOT_EQ, "!="},
		{token.NUMBER, "4"},

		{token.NUMBER, "5"},
		{token.GT_EQ, ">="},
		{token.NUMBER, "5"},

		{token.NUMBER, "6"},
		{token.GT, ">"},
		{token.NUMBER, "6"},

		{token.NUMBER, "7"},
		{token.GT, ">"},
		{token.BANG, "!"},
		{token.NUMBER, "7"},

		{token.NUMBER, "8"},
		{token.LT, "<"},
		{token.NUMBER, "8"},
		{token.SLASH, "/"},

		{token.NUMBER, "9"},
		{token.SLASH_EQ, "/="},
		{token.NUMBER, "9"},

		{token.NUMBER, "10"},
		{token.ASTERISK_EQ, "*="},
		{token.NUMBER, "10"},

		{token.STRING, "Hello World"},
		{token.STRING, "123"},
		{token.STRING, ""},

		{token.LBRACKET, "["},
		{token.RBRACKET, "]"},
		{token.LBRACKET, "["},
		{token.RBRACKET, "]"},

		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
