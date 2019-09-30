package parser

import (
	"TLang/ast"
	"TLang/lexer"
	"TLang/token"
	"fmt"
	"strconv"
)

const (
	_ int = iota
	Lowest
	Assign      // =
	Or          // or
	And         // and
	Equals      // ==
	LessGreater // > or <
	Sum         // +
	Product     // *
	Prefix      // -X or !X
	Call        // myFunction(X)
	Index       // values[2]
)

var precedences = map[token.Type]int{
	token.Eq:           Equals,
	token.NotEq:        Equals,
	token.Lt:           LessGreater,
	token.Gt:           LessGreater,
	token.LtEq:         LessGreater,
	token.GtEq:         LessGreater,
	token.Plus:         Sum,
	token.Minus:        Sum,
	token.Asterisk:     Product,
	token.Slash:        Product,
	token.Percentage:   Product,
	token.Lparen:       Call,
	token.Lbracket:     Index,
	token.Dot:          Index,
	token.And:          And,
	token.Or:           Or,
	token.Assign:       Assign,
	token.PlusEq:       Assign,
	token.MinusEq:      Assign,
	token.AsteriskEq:   Assign,
	token.SlashEq:      Assign,
	token.PercentageEq: Assign,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.Type]prefixParseFn
	infixParseFns  map[token.Type]infixParseFn
}

func (p *Parser) registerPrefix(tokenType token.Type, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.Type, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return Lowest
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return Lowest
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) noPrefixParseFnError(t token.Type) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.Type) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.Type) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.Type) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) peekError(t token.Type) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.Type]prefixParseFn)

	p.registerPrefix(token.Bang, p.parsePrefixExpression)
	p.registerPrefix(token.Plus, p.parsePrefixExpression)
	p.registerPrefix(token.Minus, p.parsePrefixExpression)

	p.registerPrefix(token.Ident, p.parseIdentifier)
	p.registerPrefix(token.Number, p.parseNumberLiteral)
	p.registerPrefix(token.String, p.parseStringLiteral)
	p.registerPrefix(token.Character, p.parseCharacterLiteral)
	p.registerPrefix(token.True, p.parseBooleanLiteral)
	p.registerPrefix(token.False, p.parseBooleanLiteral)
	p.registerPrefix(token.Void, p.parseVoidLiteral)
	p.registerPrefix(token.Lparen, p.parseGroupedExpression)
	p.registerPrefix(token.Lbracket, p.parseArrayLiteral)
	p.registerPrefix(token.Lbrace, p.parseHashLiteral)
	p.registerPrefix(token.If, p.parseIfExpression)
	p.registerPrefix(token.Loop, p.parseLoopExpression)
	p.registerPrefix(token.Function, p.parseFunctionLiteral)
	p.registerPrefix(token.Underline, p.parseUnderLineLiteral)

	p.infixParseFns = make(map[token.Type]infixParseFn)

	p.registerInfix(token.Plus, p.parseInfixExpression)
	p.registerInfix(token.Minus, p.parseInfixExpression)
	p.registerInfix(token.Asterisk, p.parseInfixExpression)
	p.registerInfix(token.Slash, p.parseInfixExpression)
	p.registerInfix(token.Percentage, p.parseInfixExpression)
	p.registerInfix(token.Eq, p.parseInfixExpression)
	p.registerInfix(token.NotEq, p.parseInfixExpression)
	p.registerInfix(token.Lt, p.parseInfixExpression)
	p.registerInfix(token.Gt, p.parseInfixExpression)
	p.registerInfix(token.LtEq, p.parseInfixExpression)
	p.registerInfix(token.GtEq, p.parseInfixExpression)
	p.registerInfix(token.And, p.parseInfixExpression)
	p.registerInfix(token.Or, p.parseInfixExpression)
	p.registerInfix(token.Assign, p.parseInfixExpression)

	p.registerInfix(token.Lparen, p.parseCallExpression)
	p.registerInfix(token.Dot, p.parseDotExpression)
	p.registerInfix(token.Lbracket, p.parseIndexExpression)
	p.registerInfix(token.Assign, p.parseAssignExpression)
	p.registerInfix(token.PlusEq, p.parseAssignExpression)
	p.registerInfix(token.MinusEq, p.parseAssignExpression)
	p.registerInfix(token.AsteriskEq, p.parseAssignExpression)
	p.registerInfix(token.SlashEq, p.parseAssignExpression)
	p.registerInfix(token.PercentageEq, p.parseAssignExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.Eof {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.Let:
		return p.parseLetStatement()
	case token.Ref:
		return p.parseRefStatement()
	case token.Ret:
		return p.parseRetStatement()
	case token.Out:
		return p.parseOutStatement()
	case token.Jump:
		return p.parseJumpStatement()
	case token.Del:
		return p.parseDelStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.Ident) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekTokenIs(token.Semicolon) {
		p.nextToken()
		return stmt
	}

	if !p.expectPeek(token.Assign) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(Lowest)

	if !p.expectPeek(token.Semicolon) {
		return nil
	}

	return stmt
}

func (p *Parser) parseRefStatement() *ast.RefStatement {
	stmt := &ast.RefStatement{Token: p.curToken}

	if !p.expectPeek(token.Ident) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.Assign) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(Lowest)

	if !p.expectPeek(token.Semicolon) {
		return nil
	}

	return stmt
}

func (p *Parser) parseAssignExpression(left ast.Expression) ast.Expression {
	stmt := &ast.AssignExpression{Token: p.curToken}

	stmt.Left = left
	stmt.Operator = p.curToken.Literal

	p.nextToken()
	stmt.Value = p.parseExpression(Lowest)

	return stmt
}

func (p *Parser) parseRetStatement() *ast.RetStatement {
	stmt := &ast.RetStatement{Token: p.curToken}

	if p.peekTokenIs(token.Semicolon) {
		p.nextToken()
		stmt.RetValue = p.parseVoidLiteral()
		return stmt
	}

	p.nextToken()

	stmt.RetValue = p.parseExpression(Lowest)

	if !p.expectPeek(token.Semicolon) {
		return nil
	}

	return stmt
}

func (p *Parser) parseOutStatement() *ast.OutStatement {
	stmt := &ast.OutStatement{Token: p.curToken}

	if p.peekTokenIs(token.Semicolon) {
		p.nextToken()
		stmt.OutValue = p.parseVoidLiteral()
		return stmt
	}

	p.nextToken()

	stmt.OutValue = p.parseExpression(Lowest)

	if !p.expectPeek(token.Semicolon) {
		return nil
	}

	return stmt
}

func (p *Parser) parseJumpStatement() *ast.JumpStatement {
	stmt := &ast.JumpStatement{Token: p.curToken}

	if !p.expectPeek(token.Semicolon) {
		return nil
	}

	return stmt
}

func (p *Parser) parseDelStatement() *ast.DelStatement {
	stmt := &ast.DelStatement{Token: p.curToken}

	p.nextToken()

	stmt.DelIdent = p.parseExpression(Lowest)

	if !p.expectPeek(token.Semicolon) {
		return nil
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(Lowest)

	if !p.expectPeek(token.Semicolon) {
		return nil
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.Semicolon) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseNumberLiteral() ast.Expression {
	valueInt, errInt := strconv.ParseInt(p.curToken.Literal, 0, 64)
	valueFloat, errFloat := strconv.ParseFloat(p.curToken.Literal, 64)
	if errInt != nil && errFloat != nil {
		msg := fmt.Sprintf("could not parse %q as integer or float", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	if errInt == nil {
		lit := &ast.IntegerLiteral{Token: p.curToken}
		lit.Value = valueInt
		return lit
	} else {
		lit := &ast.FloatLiteral{Token: p.curToken}
		lit.Value = valueFloat
		return lit
	}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.True)}
}

func (p *Parser) parseVoidLiteral() ast.Expression {
	return &ast.VoidLiteral{Token: p.curToken}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	str, err := strconv.Unquote("\"" + p.curToken.Literal + "\"")
	if err != nil {
		msg := fmt.Sprintf("escape failed: %s", err.Error())
		p.errors = append(p.errors, msg)
		return nil
	}
	return &ast.StringLiteral{Token: p.curToken, Value: str}
}

func (p *Parser) parseCharacterLiteral() ast.Expression {
	val := []rune(p.curToken.Literal)
	if len(val) != 1 {
		msg := fmt.Sprintf("expected character, got string")
		p.errors = append(p.errors, msg)
		return nil
	}
	return &ast.CharacterLiteral{Token: p.curToken, Value: val[0]}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	exp := &ast.ArrayLiteral{Token: p.curToken}
	exp.Elements = p.parseExpressionList(token.Rbracket)
	return exp
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.peekTokenIs(token.Rbrace) {
		p.nextToken()
		key := p.parseExpression(Lowest)

		if !p.expectPeek(token.Colon) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(Lowest)

		hash.Pairs[key] = value

		if !p.peekTokenIs(token.Rbrace) && !p.expectPeek(token.Comma) {
			return nil
		}
	}

	if !p.expectPeek(token.Rbrace) {
		return nil
	}

	return hash
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(Prefix)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(Lowest)

	if !p.expectPeek(token.Rparen) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	if !p.expectPeek(token.Lparen) {
		return nil
	}

	p.nextToken()
	expression.Condition = p.parseExpression(Lowest)

	if !p.expectPeek(token.Rparen) {
		return nil
	}

	if !p.expectPeek(token.Lbrace) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.Else) {
		p.nextToken()

		if !p.peekTokenIs(token.Lbrace) {
			if !p.expectPeek(token.If) {
				return nil
			}
			cur := p.curToken
			expression.Alternative = &ast.BlockStatement{
				Token: token.Token{
					Type:    token.Lbrace,
					Literal: "",
				},
				Statements: []ast.Statement{
					&ast.ExpressionStatement{
						Token:      cur,
						Expression: p.parseIfExpression(),
					},
				},
			}
			return expression
		}

		p.nextToken()
		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseLoopExpression() ast.Expression {
	Token := p.curToken

	if p.peekTokenIs(token.Ident) {
		p.nextToken()
		expression := &ast.LoopInExpression{
			Token: Token,
			Name:  &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
		}
		if !p.expectPeek(token.In) {
			return nil
		}
		p.nextToken()
		expression.Range = p.parseExpression(Lowest)

		if !p.expectPeek(token.Lbrace) {
			return nil
		}

		expression.Body = p.parseBlockStatement()
		return expression
	}

	if !p.expectPeek(token.Lparen) {
		return nil
	}

	expression := &ast.LoopExpression{Token: Token}

	p.nextToken()
	expression.Condition = p.parseExpression(Lowest)

	if !p.expectPeek(token.Rparen) {
		return nil
	}

	if !p.expectPeek(token.Lbrace) {
		return nil
	}

	expression.Body = p.parseBlockStatement()

	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.Rbrace) && !p.curTokenIs(token.Eof) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.Lparen) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.Lbrace) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseUnderLineLiteral() ast.Expression {
	lit := &ast.UnderLineLiteral{Token: p.curToken}

	if !p.expectPeek(token.Lbrace) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	var identifiers []*ast.Identifier

	if p.peekTokenIs(token.Rparen) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.Comma) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.Rparen) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.Rparen)
	return exp
}

func (p *Parser) parseDotExpression(left ast.Expression) ast.Expression {
	exp := &ast.DotExpression{Token: p.curToken, Left: left}
	p.nextToken()
	exp.Right = p.parseExpression(Index)
	return exp
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}
	exp.Indexes = p.parseExpressionList(token.Rbracket)
	return exp
}

func (p *Parser) parseExpressionList(end token.Type) []ast.Expression {
	var list []ast.Expression

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(Lowest))

	for p.peekTokenIs(token.Comma) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(Lowest))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}
