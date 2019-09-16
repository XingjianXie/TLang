package object

import (
	"TLang/ast"
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
	CHARACTER = "CHARACTER"
	VOID      = "VOID"
	RET       = "RET"
	OUT       = "OUT"
	JUMP      = "JUMP"
	ERR       = "ERR"
	FUNCTION  = "FUNCTION"
	NATIVE    = "NATIVE"
	ARRAY     = "ARRAY"
	REFERENCE = "REFERENCE"
)

type Object interface {
	Type() Type
	Inspect() string
	Copy() Object
}

type Number interface {
	NumberObj()
}

type Letter interface {
	LetterObj() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() Type      { return INTEGER }
func (i *Integer) Copy() Object    { return *&i }
func (i *Integer) NumberObj()      {}

type Float struct {
	Value float64
}

func (f *Float) Inspect() string { return fmt.Sprintf("%f", f.Value) }
func (f *Float) Type() Type      { return FLOAT }
func (f *Float) Copy() Object    { return *&f }
func (f *Float) NumberObj()      {}

type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() Type      { return BOOLEAN }
func (b *Boolean) Copy() Object    { return *&b }

type Void struct{}

func (v *Void) Inspect() string { return "void" }
func (v *Void) Type() Type      { return VOID }
func (v *Void) Copy() Object    { return *&v }

type RetValue struct {
	Value Object
}

func (rv *RetValue) Inspect() string { return rv.Value.Inspect() }
func (rv *RetValue) Type() Type      { return RET }
func (rv *RetValue) Copy() Object    { return &RetValue{Value: rv.Copy()} }

type OutValue struct {
	Value Object
}

func (ov *OutValue) Inspect() string { return ov.Value.Inspect() }
func (ov *OutValue) Type() Type      { return OUT }
func (ov *OutValue) Copy() Object    { return &OutValue{Value: ov.Copy()} }

type Jump struct{}

func (j *Jump) Inspect() string { return "jump" }
func (j *Jump) Type() Type      { return JUMP }
func (j *Jump) Copy() Object    { return *&j }

type Err struct {
	Message string
}

func (err *Err) Inspect() string { return "ERROR: " + err.Message }
func (err *Err) Type() Type      { return ERR }
func (err *Err) Copy() Object    { return *&err }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

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
func (f *Function) Type() Type   { return FUNCTION }
func (f *Function) Copy() Object { return *&f }

type String struct {
	Value string
}

func (s *String) Inspect() string   { return "\"" + s.Value + "\"" }
func (s *String) Type() Type        { return STRING }
func (s *String) Copy() Object      { return *&s }
func (s *String) LetterObj() string { return s.Value }

type Character struct {
	Value rune
}

func (c *Character) Inspect() string   { return "'" + string(c.Value) + "'" }
func (c *Character) Type() Type        { return CHARACTER }
func (c *Character) Copy() Object      { return *&c }
func (c *Character) LetterObj() string { return string(c.Value) }

type Native struct {
	Fn func(env *Environment, args []Object) Object
}

func (n *Native) Inspect() string { return "func [Native]" }
func (n *Native) Type() Type      { return NATIVE }
func (n *Native) Copy() Object    { return *&n }

type Array struct {
	Elements []Object
}

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
func (a *Array) Type() Type { return ARRAY }
func (a *Array) Copy() Object {
	var elements []Object
	for _, e := range a.Elements {
		elements = append(elements, e.Copy())
	}
	return &Array{Elements: elements}
}

type Reference struct {
	Value *Object
}

func (r *Reference) Inspect() string {
	var out bytes.Buffer

	out.WriteString("Reference: ")
	out.WriteString((*r.Value).Inspect())

	return out.String()
}
func (r *Reference) Type() Type   { return REFERENCE }
func (r *Reference) Copy() Object { return (*r.Value).Copy() }
