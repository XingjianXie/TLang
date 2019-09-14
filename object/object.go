package object

import (
	"TProject/ast"
	"bytes"
	"fmt"
	"strings"
)

type Type string

const (
	INTEGER  = "INTEGER"
	FLOAT    = "FLOAT"
	BOOLEAN  = "BOOLEAN"
	STRING   = "STRING"
	VOID     = "VOID"
	RET      = "RET"
	ERR      = "ERR"
	FUNCTION = "FUNCTION"
	NATIVE   = "NATIVE"
)

type Object interface {
	Type() Type
	Inspect() string
	IsNumber() bool
}

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() Type      { return INTEGER }
func (i *Integer) IsNumber() bool  { return true }

type Float struct {
	Value float64
}

func (i *Float) Inspect() string { return fmt.Sprintf("%f", i.Value) }
func (i *Float) Type() Type      { return FLOAT }
func (i *Float) IsNumber() bool  { return true }

type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() Type      { return BOOLEAN }
func (b *Boolean) IsNumber() bool  { return false }

type Void struct{}

func (n *Void) Inspect() string { return "void" }
func (n *Void) Type() Type      { return VOID }
func (n *Void) IsNumber() bool  { return false }

type RetValue struct {
	Value Object
}

func (rv *RetValue) Type() Type      { return RET }
func (rv *RetValue) Inspect() string { return rv.Value.Inspect() }
func (rv *RetValue) IsNumber() bool  { return false }

type Err struct {
	Message string
}

func (err *Err) Type() Type      { return ERR }
func (err *Err) Inspect() string { return "ERROR: " + err.Message }
func (err *Err) IsNumber() bool  { return false }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() Type     { return FUNCTION }
func (f *Function) IsNumber() bool { return false }
func (f *Function) Inspect() string {
	var out bytes.Buffer

	var params []string
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("func")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(f.Body.String())

	return out.String()
}

type String struct {
	Value string
}

func (s *String) Type() Type      { return STRING }
func (s *String) IsNumber() bool  { return false }
func (s *String) Inspect() string { return "\"" + s.Value + "\"" }

type NativeFunction func(env *Environment, args []Object) Object
type Native struct {
	Fn NativeFunction
}

func (n *Native) Type() Type      { return NATIVE }
func (n *Native) IsNumber() bool  { return false }
func (n *Native) Inspect() string { return "func [Native]" }
