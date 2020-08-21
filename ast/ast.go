package ast

import (
	"bytes"
	"github.com/mark07x/TLang/token"
	"strconv"
	"strings"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

type LetStatement struct {
	Token token.Token // the token.Let token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())

	if ls.Value != nil {
		out.WriteString(" = ")
		out.WriteString(ls.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

type RefStatement struct {
	Token token.Token // the token.Ref token
	Name  *Identifier
	Value Expression
}

func (rs *RefStatement) statementNode()       {}
func (rs *RefStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *RefStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")
	out.WriteString(rs.Name.String())

	if rs.Value != nil {
		out.WriteString(" = ")
		out.WriteString(rs.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

type AssignExpression struct {
	Token    token.Token // the token.Let token
	Left     Expression
	Operator string
	Value    Expression
}

func (as *AssignExpression) expressionNode()      {}
func (as *AssignExpression) TokenLiteral() string { return as.Token.Literal }
func (as *AssignExpression) String() string {
	var out bytes.Buffer

	out.WriteString(as.Left.String())
	out.WriteString(" " + as.Operator + " ")
	out.WriteString(as.Value.String())

	return out.String()
}

type RetStatement struct {
	Token    token.Token // the 'ret' token
	RetValue Expression
}

func (rs *RetStatement) statementNode()       {}
func (rs *RetStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *RetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral())

	if rs.RetValue != nil {
		out.WriteString(" " + rs.RetValue.String())
	}

	out.WriteString(";")

	return out.String()
}

type OutStatement struct {
	Token    token.Token // the 'out' token
	OutValue Expression
}

func (o *OutStatement) statementNode()       {}
func (o *OutStatement) TokenLiteral() string { return o.Token.Literal }
func (o *OutStatement) String() string {
	var out bytes.Buffer

	out.WriteString(o.TokenLiteral())

	if o.OutValue != nil {
		out.WriteString(" " + o.OutValue.String())
	}

	out.WriteString(";")

	return out.String()
}

type JumpStatement struct {
	Token token.Token // the 'jump' token
}

func (js *JumpStatement) statementNode()       {}
func (js *JumpStatement) TokenLiteral() string { return js.Token.Literal }
func (js *JumpStatement) String() string {
	var out bytes.Buffer

	out.WriteString(js.TokenLiteral())
	out.WriteString(";")

	return out.String()
}

type DelStatement struct {
	Token    token.Token // the 'ret' token
	DelIdent Expression
}

func (rs *DelStatement) statementNode()       {}
func (rs *DelStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *DelStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")
	out.WriteString(rs.DelIdent.String())
	out.WriteString(";")

	return out.String()
}

type Identifier struct {
	Token token.Token // the token.Ident token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String() + ";"
	}
	return ""
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (il *FloatLiteral) expressionNode()      {}
func (il *FloatLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *FloatLiteral) String() string       { return il.Token.Literal }

type PrefixExpression struct {
	Token    token.Token // The prefix token, e.g. !
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (b *BooleanLiteral) expressionNode()      {}
func (b *BooleanLiteral) TokenLiteral() string { return b.Token.Literal }
func (b *BooleanLiteral) String() string       { return b.Token.Literal }

type VoidLiteral struct {
	Token token.Token
}

func (v *VoidLiteral) expressionNode()      {}
func (v *VoidLiteral) TokenLiteral() string { return v.Token.Literal }
func (v *VoidLiteral) String() string       { return v.Token.Literal }

type IfExpression struct {
	Token       token.Token // The 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if ")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())

	if ie.Alternative != nil {
		out.WriteString(" else ")
		out.WriteString(ie.Alternative.String())
	}

	return out.String()
}

type LoopExpression struct {
	Token     token.Token // The 'loop' token
	Condition Expression
	Body      *BlockStatement
}

func (le *LoopExpression) expressionNode()      {}
func (le *LoopExpression) TokenLiteral() string { return le.Token.Literal }
func (le *LoopExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if ")
	out.WriteString(le.Condition.String())
	out.WriteString(" ")
	out.WriteString(le.Body.String())

	return out.String()
}

type LoopInExpression struct {
	Token token.Token // The 'ref' token
	Name  *Identifier
	Range Expression
	Body  *BlockStatement
}

func (li *LoopInExpression) expressionNode()      {}
func (li *LoopInExpression) TokenLiteral() string { return li.Token.Literal }
func (li *LoopInExpression) String() string {
	var out bytes.Buffer

	out.WriteString("loop ")
	out.WriteString(li.Name.Value)
	out.WriteString(" in ")
	out.WriteString(li.Range.String())
	out.WriteString(" ")
	out.WriteString(li.Body.String())

	return out.String()
}

type BlockStatement struct {
	Token      token.Token // the { token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	out.WriteString("{ ")
	for index, s := range bs.Statements {
		if index != 0 {
			out.WriteString(" ")
		}
		out.WriteString(s.String())
	}
	out.WriteString(" }")

	return out.String()
}

type FunctionLiteral struct {
	Token      token.Token // The 'func' token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	var params []string
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())

	return out.String()
}

type UnderLineLiteral struct {
	Token token.Token // The '_' token
	Body  *BlockStatement
}

func (ul *UnderLineLiteral) expressionNode()      {}
func (ul *UnderLineLiteral) TokenLiteral() string { return ul.Token.Literal }
func (ul *UnderLineLiteral) String() string {
	var out bytes.Buffer

	out.WriteString(ul.TokenLiteral())
	out.WriteString(" ")
	out.WriteString(ul.Body.String())

	return out.String()
}

type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	var args []string
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

type DotExpression struct {
	Token token.Token // The '.' token
	Left  Expression
	Right Expression
}

func (de *DotExpression) expressionNode()      {}
func (de *DotExpression) TokenLiteral() string { return de.Token.Literal }
func (de *DotExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(de.Left.String())
	out.WriteString(".")
	out.WriteString(de.Right.String())
	out.WriteString(")")

	return out.String()
}

type IndexExpression struct {
	Token   token.Token // The '[' token
	Left    Expression
	Indexes []Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer

	var args []string
	for _, a := range ie.Indexes {
		args = append(args, a.String())
	}

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString("]")
	out.WriteString(")")

	return out.String()
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return strconv.Quote(sl.Value) }

type CharacterLiteral struct {
	Token token.Token
	Value rune
}

func (cl *CharacterLiteral) expressionNode()      {}
func (cl *CharacterLiteral) TokenLiteral() string { return cl.Token.Literal }
func (cl *CharacterLiteral) String() string       { return "'" + cl.Token.Literal + "'" }

type ArrayLiteral struct {
	Token    token.Token // the '[' token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer

	var elements []string
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type HashLiteral struct {
	Token token.Token // the '{' token
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }
func (hl *HashLiteral) String() string {
	var out bytes.Buffer

	var pairs []string
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}

	out.WriteString("{ ")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString(" }")

	return out.String()
}
