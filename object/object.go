package object

import (
	"bytes"
	"fmt"
	"github.com/mark07x/TLang/ast"
	"strconv"
	"strings"
)

type Type string
type TypeC string

var (
	TrueObj  Object = &Boolean{Value: true}
	FalseObj Object = &Boolean{Value: false}
	VoidObj  Object = &Void{}
	JumpObj  Object = &Jump{}
)

const (
	INT64 TypeC = "long long"
	INT TypeC = "int"
	FLOAT64 TypeC = "double"
	POINTER TypeC = "pointer"
	INVALID TypeC = "invalid"
)

const (
	INTEGER Type     = "Integer"
	FLOAT   Type    = "Float"
	BOOLEAN  Type   = "Boolean"
	STRING    Type  = "String"
	CHARACTER Type  = "Character"
	VOID      Type  = "Void"
	RET      Type   = "Ret"
	OUT      Type   = "Out"
	JUMP     Type   = "Jump"
	ERR      Type   = "Err"
	FUNC      Type  = "Func"
	UNDERLINE  Type = "Underline"
	NATIVE    Type  = "Native"
	ARRAY     Type  = "Array"
	REFERENCE Type  = "Reference"
	HASH     Type   = "Hash"
	ENVIRONMENT Type = "Environment"
)

func UnwrapRetValue(obj Object) Object {
	if retValue, ok := obj.(*RetValue); ok {
		return retValue.Value
	}

	return obj
}

func UnwrapOutValue(obj Object) Object {
	if retValue, ok := obj.(*OutValue); ok {
		return retValue.Value
	}

	return obj
}

func UnwrapReferenceValue(obj Object) Object {
	if referenceVal, ok := obj.(*Reference); ok {
		if referenceVal.Value == nil {
			return VoidObj
		}
		if fun, ok := (*referenceVal.Value).(*Function); ok {
			fun.Self = referenceVal.Origin
		}
		return *referenceVal.Value
	}
	return obj
}

type Object interface {
	Type() Type
	TypeC() TypeC
	Inspect(num int) string
	Copy() Object
}

type Number interface {
	Object
	NumberObj()
}

type Letter interface {
	Object
	LetterObj() string
}

type Functor interface {
	Object
	FunctorObj()
}

type HashAble interface {
	Object
	HashKey() HashKey
}

type Allocable interface {
	Object
	DoAlloc(Index Object) (*Object, bool)
	DeAlloc(Index Object) bool
}

type HashKey struct {
	Type  Type
	Value interface{}
}

type Integer struct {
	Value int64
}

func (i *Integer) Inspect(num int) string { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() Type      { return INTEGER }
func (i *Integer) TypeC() TypeC      { return INT64 }
func (i *Integer) Copy() Object    { return i }
func (i *Integer) NumberObj()      {}
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: i.Value}
}

type Float struct {
	Value float64
}

func (f *Float) Inspect(num int) string { return fmt.Sprintf("%g", f.Value) }
func (f *Float) Type() Type      { return FLOAT }
func (f *Float) TypeC() TypeC      { return FLOAT64 }
func (f *Float) Copy() Object    { return f }
func (f *Float) NumberObj()      {}

type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect(num int) string { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() Type      { return BOOLEAN }
func (b *Boolean) TypeC() TypeC      { return INVALID }
func (b *Boolean) Copy() Object    { return b }
func (b *Boolean) HashKey() HashKey {
	return HashKey{Type: b.Type(), Value: b.Value}
}

type Void struct{}

func (v *Void) Inspect(num int) string { return "void" }
func (v *Void) Type() Type      { return VOID }
func (v *Void) TypeC() TypeC      { return INT }
func (v *Void) Copy() Object    { return v }
func (v *Void) HashAble() HashKey {
	return HashKey{Type: v.Type(), Value: 0}
}

type RetValue struct {
	Value Object
}

func (rv *RetValue) Inspect(num int) string { return rv.Value.Inspect(num) }
func (rv *RetValue) Type() Type      { return RET }
func (rv *RetValue) TypeC() TypeC      { return INVALID }
func (rv *RetValue) Copy() Object {
	println("WARNING: COPY RET VALUE")
	return &RetValue{Value: rv.Copy()}
}

type OutValue struct {
	Value Object
}

func (ov *OutValue) Inspect(num int) string { return ov.Value.Inspect(num) }
func (ov *OutValue) Type() Type      { return OUT }
func (ov *OutValue) TypeC() TypeC      { return INVALID }
func (ov *OutValue) Copy() Object {
	println("WARNING: COPY OUT VALUE")
	return &OutValue{Value: ov.Copy()}
}

type Jump struct{}

func (j *Jump) Inspect(num int) string { return "jump" }
func (j *Jump) Type() Type      { return JUMP }
func (j *Jump) TypeC() TypeC      { return INVALID }
func (j *Jump) Copy() Object {
	println("WARNING: COPY JUMP")
	return j
}

type Err struct {
	Message string
}

func (err *Err) Inspect(num int) string { return "ERROR: " + err.Message }
func (err *Err) Type() Type      { return ERR }
func (err *Err) TypeC() TypeC     { return INVALID }
func (err *Err) Copy() Object {
	println("WARNING: COPY ERR")
	return err
}

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
	Self       Object
}

func (f *Function) Inspect(num int) string {
	if num <= 1 {
		return "func(...) {...}"
	}
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
func (f *Function) Type() Type       { return FUNC }
func (f *Function) TypeC() TypeC       { return INVALID }
func (f *Function) Copy() Object { return f }
func (f *Function) FunctorObj()  {}

type UnderLine struct {
	Body *ast.BlockStatement
	Env  *Environment
}

func (u *UnderLine) Inspect(num int) string {
	if num <= 1 {
		return "_ {...}"
	}
	var out bytes.Buffer

	out.WriteString("_ ")
	out.WriteString(u.Body.String())

	return out.String()
}
func (u *UnderLine) Type() Type       { return UNDERLINE }
func (u *UnderLine) TypeC() TypeC       { return INVALID }
func (u *UnderLine) Copy() Object { return u }
func (u *UnderLine) FunctorObj()  {}

type String struct {
	Value []rune
}

func (s *String) Inspect(num int) string   { return strconv.Quote(string(s.Value)) }
func (s *String) Type() Type        { return STRING }
func (s *String) TypeC() TypeC        { return POINTER }
func (s *String) Copy() Object      { return s }
func (s *String) LetterObj() string { return string(s.Value) }
func (s *String) HashKey() HashKey {
	return HashKey{Type: s.Type(), Value: string(s.Value)}
}

type Character struct {
	Value rune
}

func (c *Character) Inspect(num int) string   { return "'" + string(c.Value) + "'" }
func (c *Character) Type() Type        { return CHARACTER }
func (c *Character) TypeC() TypeC        { return INVALID }
func (c *Character) Copy() Object      { return c }
func (c *Character) LetterObj() string { return string(c.Value) }
func (c *Character) HashKey() HashKey {
	return HashKey{Type: c.Type(), Value: c.Value}
}

type Native struct {
	Fn func(env *Environment, args []Object) Object
}

func (n *Native) Inspect(num int) string  { return "func [Native]" }
func (n *Native) Type() Type       { return NATIVE }
func (n *Native) TypeC() TypeC       { return INVALID }
func (n *Native) Copy() Object { return n }
func (n *Native) FunctorObj()  {}

type Array struct {
	Elements []Object
	Copyable bool
}

func (a *Array) Inspect(num int) string {
	if num <= 0 {
		return "[...]"
	}
	var out bytes.Buffer

	var elements []string
	for _, e := range a.Elements {
		elements = append(elements, e.Inspect(num - 1))
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}
func (a *Array) Type() Type { return ARRAY }
func (a *Array) TypeC() TypeC { return POINTER }
func (a *Array) Copy() Object {
	if a.Copyable {
		a.Copyable = false
		return a
	}
	var elements []Object
	for _, e := range a.Elements {
		elements = append(elements, e.Copy())
	}
	return &Array{Elements: elements}
}

type Reference struct {
	Value  *Object
	Origin Allocable
	Index  Object
	Const  bool
}

func (r *Reference) Inspect(num int) string {
	var out bytes.Buffer

	if r.Const {
		out.WriteString("Const ")
	}
	out.WriteString("Reference ")
	if r.Value != nil {
		out.WriteString("(" + (*r.Value).Inspect(num) + ")")
	} else {
		out.WriteString("(Not Alloc)")
	}


	return out.String()
}
func (r *Reference) Type() Type { return REFERENCE }
func (r *Reference) TypeC() TypeC { return INVALID }
func (r *Reference) Copy() Object {
	return r
}

type HashPair struct {
	Key   Object
	Value *Object
}

type Hash struct {
	Pairs    map[HashKey]HashPair
	Copyable bool
}

func (h *Hash) FunctorObj()  {}
func (h *Hash) DoAlloc(Index Object) (*Object, bool) {
	if hashIndex, ok := Index.(HashAble); ok {
		key := hashIndex.HashKey()
		if _, ok := h.Pairs[key]; !ok {
			var obj Object = nil
			h.Pairs[key] = HashPair{
				Key:   hashIndex,
				Value: &obj,
			}
			return &obj, true
		}
	}
	return nil, false
}
func (h *Hash) DeAlloc(Index Object) bool {
	if hashIndex, ok := Index.(HashAble); ok {
		key := hashIndex.HashKey()
		if _, ok := h.Pairs[key]; ok {
			delete(h.Pairs, key)
			return true
		}
	}
	return false
}

func (h *Hash) Inspect(num int) string {
	if num <= 0 {
		return "{...}"
	}
	var out bytes.Buffer

	var pairs []string
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			pair.Key.Inspect(num - 1), (*pair.Value).Inspect(num - 1)))
	}

	out.WriteString("{ ")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString(" }")

	return out.String()
}
func (h *Hash) Type() Type { return HASH }
func (h *Hash) TypeC() TypeC { return INVALID }
func (h *Hash) Copy() Object {
	if h.Copyable {
		h.Copyable = false
		return h
	}
	pairs := make(map[HashKey]HashPair)
	for index, pair := range h.Pairs {
		newVal := (*pair.Value).Copy()
		pairs[index] = HashPair{
			Key:   pair.Key,
			Value: &newVal,
		}
	}

	return &Hash{
		Pairs:    pairs,
		Copyable: false,
	}
}
