package lexer

import (
	"TLang/token"
)

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.Eq, Literal: literal}
		} else {
			tok = newToken(token.Assign, l.ch)
		}

	case '+':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.PlusEq, Literal: literal}
		} else {
			tok = newToken(token.Plus, l.ch)
		}

	case '-':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.MinusEq, Literal: literal}
		} else {
			tok = newToken(token.Minus, l.ch)
		}

	case '*':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.AsteriskEq, Literal: literal}
		} else {
			tok = newToken(token.Asterisk, l.ch)
		}

	case '/':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.SlashEq, Literal: literal}
		} else {
			tok = newToken(token.Slash, l.ch)
		}

	case '%':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.PercentageEq, Literal: literal}
		} else {
			tok = newToken(token.Percentage, l.ch)
		}

	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.NotEq, Literal: literal}
		} else {
			tok = newToken(token.Bang, l.ch)
		}

	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.LtEq, Literal: literal}
		} else {
			tok = newToken(token.Lt, l.ch)
		}

	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.GtEq, Literal: literal}
		} else {
			tok = newToken(token.Gt, l.ch)
		}

	case ';':
		tok = newToken(token.Semicolon, l.ch)
	case ',':
		tok = newToken(token.Comma, l.ch)
	case '(':
		tok = newToken(token.Lparen, l.ch)
	case ')':
		tok = newToken(token.Rparen, l.ch)
	case '{':
		tok = newToken(token.Lbrace, l.ch)
	case '}':
		tok = newToken(token.Rbrace, l.ch)
	case '[':
		tok = newToken(token.Lbracket, l.ch)
	case ']':
		tok = newToken(token.Rbracket, l.ch)
	case '.':
		if isDigit(l.peekChar()) {
			tok.Literal = l.readNumber()
			tok.Type = token.Number
			return tok
		}
		tok = newToken(token.Dot, l.ch)
	case '"':
		tok.Type = token.String
		tok.Literal = l.readString()
	case '\'':
		tok.Type = token.Character
		tok.Literal = l.readCharacter()
	case 0:
		tok.Literal = ""
		tok.Type = token.Eof
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = token.Number
			return tok
		} else {
			tok = newToken(token.Illegal, l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) readCharacter() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '\'' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	eExist := false
	for isDigitEx(l.ch, &eExist) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '$' || ch == '@'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isDigitEx(ch byte, eExist *bool) bool {
	if !*eExist && ch == 'e' {
		*eExist = true
		return true
	}
	return '0' <= ch && ch <= '9' || (!*eExist && ch == '.') || (*eExist && (ch == '+' || ch == '-'))
}

func newToken(tokenType token.Type, ch byte) token.Token {
	return token.Token{
		Type:    tokenType,
		Literal: string(ch),
	}
}
