package object

import (
	"TProject/ast"
	"bytes"
	"fmt"
	"strings"
)

type Type string

const (
	INTEGER   = "INTEGER"
	FLOAT     = "FLOAT"
	BOOLEAN   = "BOOLEAN"
	STRING    = "STRING"
	VOID      = "VOID"
	RET       = "RET"
	ERR       = "ERR"
	FUNCTION  = "FUNCTION"
	NATIVE    = "NATIVE"
	ARRAY     = "ARRAY"
	REFERENCE = "REFERENCE"
)

type Object interface {
	Type() Type
	Inspect() string
}

type IndexType interface {
	Object
	Index(index Object) Object
}

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() Type      { return INTEGER }

type Float struct {
	Value float64
}

func (i *Float) Inspect() string { return fmt.Sprintf("%f", i.Value) }
func (i *Float) Type() Type      { return FLOAT }

type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() Type      { return BOOLEAN }

type Void struct{}

func (n *Void) Inspect() string { return "void" }
func (n *Void) Type() Type      { return VOID }

type RetValue struct {
	Value Object
}

func (rv *RetValue) Type() Type      { return RET }
func (rv *RetValue) Inspect() string { return rv.Value.Inspect() }

type Err struct {
	Message string
}

func (err *Err) Type() Type      { return ERR }
func (err *Err) Inspect() string { return "ERROR: " + err.Message }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() Type { return FUNCTION }
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
func (s *String) Inspect() string { return "\"" + s.Value + "\"" }

type Native struct {
	Fn func(env *Environment, args []Object) Object
}

func (n *Native) Type() Type      { return NATIVE }
func (n *Native) Inspect() string { return "func [Native]" }

type Array struct {
	Elements []Object
}

func (a *Array) Type() Type { return ARRAY }
func (a *Array) Inspect() string {
	var out bytes.Buffer

	var elements []string
	for _, e := range a.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type Reference struct {
	Value *Object
}

func (r *Reference) Type() Type { return REFERENCE }
func (r *Reference) Inspect() string {
	var out bytes.Buffer

	out.WriteString("Reference: ")
	out.WriteString((*r.Value).Inspect())

	return out.String()
}
