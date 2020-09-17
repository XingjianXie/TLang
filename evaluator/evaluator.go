package evaluator

import (
	"bufio"
	"fmt"
	"github.com/mark07x/TLang/ast"
	"github.com/mark07x/TLang/lexer"
	"github.com/mark07x/TLang/object"
	"github.com/mark07x/TLang/parser"
	"io"
	"io/ioutil"
	"math"
	"os"
	"strconv"
)

func PrintParserErrors(out io.Writer, errors []string) {
	_, _ = io.WriteString(out, "PARSER ERRORS:\n")
	for _, msg := range errors {
		_, _ = io.WriteString(out, "    "+msg+"\n")
	}
}

func init() {
	Bases = map[string]object.Object{
		"super": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 2 {
				return newError("native function super: len(args) should be 2")
			}
			return applyIndex(args[0], []object.Object{args[1]}, Super)
		}},
		"current": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 2 {
				return newError("native function current: len(args) should be 2")
			}
			return applyIndex(args[0], []object.Object{args[1]}, Current)
		}},
		"classType": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function instance: len(args) should be 1")
			}
			if h, ok := object.UnwrapReferenceValue(args[0]).(*object.Hash); ok {
				return &object.String{Value: []rune(classType(h))}
			}
			return newError("native function instance: arg should be Hash")
		}},
		"call": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 2 {
				return newError("native function call: len(args) should be 2")
			}
			v := object.UnwrapReferenceValue(args[1])
			if arr, ok := v.(*object.Array); ok {
				return applyFunction(object.UnwrapReferenceValue(args[0]), arr.Elements, env)
			}
			return newError("native function call: args[1] should be Array")
		}},
		"subscript": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 2 {
				return newError("native function subscript: len(args) should be 2")
			}
			v := object.UnwrapReferenceValue(args[1])
			if arr, ok := v.(*object.Array); ok {
				return applyIndex(args[0], arr.Elements, Default)
			}
			return newError("native function subscript: args[1] should be Array")
		}},
		"len": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function len: len(args) should be 1")
			}
			switch arg := object.UnwrapReferenceValue(args[0]).(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			default:
				return newError("native function len: arg should be String or Array")
			}
		}},
		"print": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			for _, arg := range args {
				if f, ok := env.Get("string"); ok {
					v := applyFunction(*f, []object.Object{arg}, env)
					fmt.Print(string(v.(*object.String).Value))
				} else {
					return newError("string lost")
				}
			}
			return object.VoidObj
		}},
		"input": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 0 {
				return newError("native function input: len(args) should be 0")
			}
			var input string
			_, _ = fmt.Scanf("%s", &input)

			return &object.String{Value: []rune(input)}
		}},
		"printLine": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) == 0 {
				fmt.Println()
				return object.VoidObj
			}
			for _, arg := range args {
				if f, ok := env.Get("string"); ok {
					v := applyFunction(*f, []object.Object{arg}, env)
					fmt.Println(string(v.(*object.String).Value))
				} else {
					return newError("string lost")
				}
			}
			return object.VoidObj
		}},
		"inputLine": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 0 {
				return newError("native function inputLine: len(args) should be 0")
			}
			data, _, _ := bufio.NewReader(os.Stdin).ReadLine()

			return &object.String{Value: []rune(string(data))}
		}},
		"string": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function string: len(args) should be 1")
			}
			un := object.UnwrapReferenceValue(args[0])
			if str, ok := un.(*object.String); ok {
				return str
			}
			if ch, ok := un.(*object.Character); ok {
				return &object.String{Value: []rune{ch.Value}}
			}
			return &object.String{Value: []rune(un.Inspect())}
		}},
		"exit": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 && len(args) != 0 {
				return newError("native function exit: len(args) should be 1 or 0")
			}

			if len(args) == 1 {
				if val, ok := object.UnwrapReferenceValue(args[0]).(*object.Integer); ok {
					os.Exit(int(val.Value))
				}
				return newError("native function exit: arg should be Integer")
			}
			os.Exit(0)
			return object.VoidObj
		}},
		"eval": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function eval: len(args) should be 1")
			}

			if str, ok := object.UnwrapReferenceValue(args[0]).(*object.String); ok {
				l := lexer.New(string(str.Value))
				p := parser.New(l)

				program := p.ParseProgram()
				if len(p.Errors()) != 0 {
					PrintParserErrors(os.Stdout, p.Errors())
					return newError("error inner eval")
				}

				return Eval(program, env)
			}

			return newError("native function eval: arg should be String")
		}},
		"integer": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function int: len(args) should be 1")
			}
			switch arg := object.UnwrapReferenceValue(args[0]).(type) {
			case *object.String:
				val, err := strconv.ParseInt(string(arg.Value), 10, 64)
				if err != nil {
					return newError("could not parse %s as integer", string(arg.Value))
				}
				return &object.Integer{Value: val}
			case *object.Character:
				val, err := strconv.ParseInt(string(arg.Value), 10, 64)
				if err != nil {
					return newError("could not parse %s as integer", string(arg.Value))
				}
				return &object.Integer{Value: val}
			case *object.Boolean:
				if arg.Value {
					return &object.Integer{Value: 1}
				} else {
					return &object.Integer{Value: 0}
				}
			case *object.Float:
				return &object.Integer{Value: int64(arg.Value)}
			case *object.Integer:
				return arg
			case *object.Void:
				return &object.Integer{Value: 0}
			default:
				return newError("native function integer: arg should be String, Boolean, Number or object.VoidObj")
			}
		}},

		"float": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function float: len(args) should be 1")
			}
			switch arg := object.UnwrapReferenceValue(args[0]).(type) {
			case *object.String:
				val, err := strconv.ParseFloat(string(arg.Value), 64)
				if err != nil {
					return newError("could not parse %s as float", string(arg.Value))
				}
				return &object.Float{Value: val}
			case *object.Character:
				val, err := strconv.ParseFloat(string(arg.Value), 64)
				if err != nil {
					return newError("could not parse %s as float", string(arg.Value))
				}
				return &object.Float{Value: val}
			case *object.Boolean:
				if arg.Value {
					return &object.Float{Value: 1.}
				} else {
					return &object.Float{Value: 0.}
				}
			case *object.Integer:
				return &object.Float{Value: float64(arg.Value)}
			case *object.Float:
				return arg
			case *object.Void:
				return &object.Float{Value: 0}
			default:
				return newError("native function int: arg should be String, Boolean, Number or object.VoidObj")
			}
		}},

		"boolean": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function boolean: len(args) should be 1")
			}
			return toBoolean(object.UnwrapReferenceValue(args[0]))
		}},

		"fetch": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function fetch: len(args) should be 1")
			}
			if err, ok := args[0].(*object.Err); ok {
				return &object.String{Value: []rune(err.Inspect())}
			}
			return args[0]
		}},

		"append": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 2 {
				return newError("native function append: len(args) should be 2")
			}
			if array, ok := object.UnwrapReferenceValue(args[0]).(*object.Array); ok {
				return &object.Array{Elements: append(array.Elements, object.UnwrapReferenceValue(args[1])), Copyable: true}
			}
			return newError("native function append: args[0] should be Array")
		}},

		"first": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function first: len(args) should be 1")
			}
			if array, ok := object.UnwrapReferenceValue(args[0]).(*object.Array); ok {
				if len(array.Elements) == 0 {
					return object.VoidObj
				}
				return &object.Reference{Value: &array.Elements[0], Const: false}
			}
			return newError("native function first: arg should be Array")
		}},

		"last": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function fetch: len(args) should be 1")
			}
			if array, ok := object.UnwrapReferenceValue(args[0]).(*object.Array); ok {
				if len(array.Elements) == 0 {
					return object.VoidObj
				}
				return &object.Reference{Value: &array.Elements[len(array.Elements)-1], Const: false}
			}
			return newError("native function append: arg should be Array")
		}},

		"type": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function type: len(args) should be 1")
			}
			if refer, ok := args[0].(*object.Reference); ok {
				isConst := ""
				if refer.Const {
					isConst = "Const "
				}
				rawType := "Not Alloc"
				if refer.Value != nil {
					rawType = string((*refer.Value).Type())
				}
				return &object.String{Value: []rune(isConst + "Reference (" + rawType + ")")}
			}
			return &object.String{Value: []rune(args[0].Type())}
		}},

		"array": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) == 1 {
				if length, ok := object.UnwrapReferenceValue(args[0]).(*object.Integer); ok {
					var elem []object.Object
					for i := int64(0); i < length.Value; i++ {
						elem = append(elem, object.VoidObj)
					}

					return &object.Array{
						Elements: elem,
						Copyable: true,
					}
				}
				return newError("native function array: args[0] should be Integer")
			} else if len(args) == 2 {
				if length, ok := object.UnwrapReferenceValue(args[0]).(*object.Integer); ok {
					var elem []object.Object
					for i := int64(0); i < length.Value; i++ {
						elem = append(elem, object.UnwrapReferenceValue(args[1]))
					}

					return &object.Array{
						Elements: elem,
						Copyable: true,
					}
				}
				return newError("native function array: args[0] should be Integer")
			} else if len(args) == 3 {
				if length, ok := object.UnwrapReferenceValue(args[0]).(*object.Integer); ok {
					if function, ok := object.UnwrapReferenceValue(args[2]).(object.Functor); ok {
						var elem []object.Object
						e := object.UnwrapReferenceValue(args[1])
						for i := int64(0); i < length.Value; i++ {
							e = object.UnwrapReferenceValue(applyFunction(function, []object.Object{&object.Integer{Value: i}, e}, env))
							if isError(e) {
								return e
							}
							elem = append(elem, e)
						}

						return &object.Array{
							Elements: elem,
							Copyable: true,
						}
					}
					return newError("native function array: args[2] should be Functor")
				}
				return newError("native function array: args[0] should be Integer")
			}
			return newError("native function array: len(args) should be 1, 2 or 3")
		}},

		"value": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function value: len(args) should be 1")
			}
			return object.UnwrapReferenceValue(args[0])
		}},

		"echo": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function echo: len(args) should be 1")
			}
			return args[0]
		}},

		"error": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function error: len(args) should be 1")
			}
			if str, ok := object.UnwrapReferenceValue(args[0]).(*object.String); ok {
				return newError(string(str.Value))
			} else {
				return newError("native function error: arg should be String")
			}
		}},

		"import": &object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function import: len(args) should be 1")
			}
			if str, ok := object.UnwrapReferenceValue(args[0]).(*object.String); ok {
				data, err := ioutil.ReadFile(string(str.Value))
				if err != nil {
					return newError("unable to read file %s: %s", string(str.Value), err.Error())
				}
				l := lexer.New(string(data))
				p := parser.New(l)

				program := p.ParseProgram()
				if len(p.Errors()) != 0 {
					PrintParserErrors(os.Stdout, p.Errors())
					return newError("error inner import")
				}

				importEnv := object.NewEnvironment(Bases)
				result := Eval(program, importEnv)
				if isError(result) {
					return result
				}
				if export, ok := importEnv.Get("export"); ok {
					return &object.Reference{Value: export, Const: true}
				}
				return object.VoidObj
				//return newError("native function import: export obj not found")
			}
			return newError("native function import: arg should be String")
		}},
	}
}

var Bases map[string]object.Object

func newError(format string, a ...interface{}) *object.Err {
	return &object.Err{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	return obj.Type() == object.ERR
}

func isSkip(obj object.Object) bool {
	return obj.Type() == object.RET || obj.Type() == object.OUT || obj.Type() == object.JUMP
}

func nativeBoolToBooleanObject(input bool) object.Object {
	if input {
		return object.TrueObj
	}
	return object.FalseObj
}

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return object.UnwrapReferenceValue(evalProgram(node, env))

	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: []rune(node.Value)}
	case *ast.CharacterLiteral:
		return &object.Character{Value: node.Value}
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}
	case *ast.UnderLineLiteral:
		body := node.Body
		return &object.UnderLine{Env: env, Body: body}
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env, true)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements, Copyable: false}
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)

	case *ast.PrefixExpression:
		right := object.UnwrapReferenceValue(Eval(node.Right, env))
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := object.UnwrapReferenceValue(Eval(node.Left, env))
		if isError(left) {
			return left
		}

		right := object.UnwrapReferenceValue(Eval(node.Right, env))
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)

	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.LoopExpression:
		return evalLoopExpression(node, env)
	case *ast.LoopInExpression:
		return evalLoopInExpression(node, env)
	case *ast.AssignExpression:
		return evalAssignExpression(node, env)
	case *ast.CallExpression:
		function := object.UnwrapReferenceValue(Eval(node.Function, env))
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env, false)
		if len(args) == 1 && isError(args[0]) {
			if function != Bases["fetch"] {
				return args[0]
			}
		}

		return applyFunction(function, args, env)
	case *ast.IndexExpression:
		ident := Eval(node.Left, env)
		if isError(ident) {
			return ident
		}
		indexes := evalExpressions(node.Indexes, env, true)
		if len(indexes) == 1 && isError(indexes[0]) {
			return indexes[0]
		}
		if f, ok := env.Get("subscript"); ok {
			return applyFunction(*f, []object.Object{ident, &object.Array{Elements: indexes}}, env)
		}
		return newError("subscript lost")
	case *ast.DotExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		if str, ok := node.Right.(*ast.Identifier); ok {
			if f, ok := env.Get("subscript"); ok {
				return applyFunction(*f, []object.Object{left, &object.Array{Elements: []object.Object{&object.String{Value: []rune(str.Value)}}}}, env)
			}
			return newError("subscript lost")
		}
		return newError("Not a key: %s", node.Right.String())

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.RetStatement:
		val := Eval(node.RetValue, env)
		if isError(val) {
			return val
		}
		return &object.RetValue{Value: val}
	case *ast.OutStatement:
		val := Eval(node.OutValue, env)
		if isError(val) {
			return val
		}
		return &object.OutValue{Value: val}
	case *ast.JumpStatement:
		return object.JumpObj
	case *ast.LetStatement:
		val := object.VoidObj
		if node.Value != nil {
			val = object.UnwrapReferenceValue(Eval(node.Value, env))
		}

		if isError(val) {
			return val
		}
		if _, ok := env.SetCurrent(node.Name.Value, val.Copy()); !ok {
			return newError("identifier %s already set", node.Name.Value)
		}
	case *ast.RefStatement:
		return evalRefStatement(node, env)
	case *ast.DelStatement:
		if ident, ok := node.DelIdent.(*ast.Identifier); ok {
			if _, ok := env.Get(ident.Value); ok {
				if !env.DeAlloc(&object.String{Value: []rune(ident.Value)}) {
					return newError("unable to dealloc: %s", node.DelIdent.String())
				}
			} else {
				return newError("identifier not found: %s", ident.Value)
			}
		} else {
			if refer, ok := Eval(node.DelIdent, env).(*object.Reference); ok {
				if refer.Const {
					return newError("delete a constant reference: %s", refer.Inspect())
				}
				if refer.Origin != nil {
					if !refer.Origin.DeAlloc(refer.Index) {
						return newError("unable to dealloc: %s", refer.Inspect())
					}
					return object.VoidObj
				}
			}
			return newError("left value not Identifier or AllocRequired: %s", node.DelIdent.String())
		}
	}

	return object.VoidObj
}

type classFlag int
const (
	_ classFlag = iota
	Default
	Current
	Super
)

func classType(hash *object.Hash) string {
	c, hasClass := hash.Pairs[object.HashKey{
		Type:  "String",
		Value: "@class",
	}]
	if hasClass {
		_, hasClass = (*c.Value).(*object.String)
	}
	t, hasTemplate := hash.Pairs[object.HashKey{
		Type:  "String",
		Value: "@template",
	}]
	if hasTemplate {
		_, hasTemplate = (object.UnwrapReferenceValue(*t.Value)).(*object.Hash)
	}
	switch {
	case hasClass:
		return "Proto"
	case !hasClass && hasTemplate:
		return "Instance"
	default:
		return ""
	}
}

func template(hash *object.Hash) (*object.Hash, bool) {
	if t, ok := hash.Pairs[object.HashKey{Type: "String", Value: "@template"}]; ok {
		if h, ok := (object.UnwrapReferenceValue(*t.Value)).(*object.Hash); ok {
			return h, true
		}
	}
	return nil, false
}

func applyIndex(ident object.Object, indexes []object.Object, flag classFlag) object.Object {
	constObj := true
	if refer, ok := ident.(*object.Reference); ok {
		constObj = refer.Const
		ident = object.UnwrapReferenceValue(ident)
	}

	if arr, ok := ident.(*object.Array); ok {
		if len(indexes) != 1 {
			return newError("array: len(indexes) should be 1")
		}
		if indexes[0].Type() != object.INTEGER {
			return newError("array: index should be Integer")
		}
		index := indexes[0].(*object.Integer).Value
		length := int64(len(arr.Elements))
		if index >= length || index < 0 {
			return newError("array: out of range")
		}
		refObj := &arr.Elements[index]
		if refer, ok := (*refObj).(*object.Reference); ok {
			return refer
		}
		return &object.Reference{Value: refObj, Const: constObj}
	}
	if str, ok := ident.(*object.String); ok {
		//runeStr := []rune(str.Value)
		if len(indexes) != 1 {
			return newError("string: len(indexes) should be 1")
		}
		if indexes[0].Type() != object.INTEGER {
			return newError("string: index should be Integer")
		}
		index := indexes[0].(*object.Integer).Value
		length := int64(len(str.Value))
		if index >= length || index < 0 {
			return newError("string: out of range")
		}
		var c object.Object = &object.Character{Value: str.Value[index]}
		return &object.Reference{Value: &c, Const: true}
	}
	if hash, ok := ident.(*object.Hash); ok {
		if len(indexes) != 1 {
			return newError("string: len(indexes) should be 1")
		}

		key, ok := indexes[0].(object.HashAble)
		if !ok {
			return newError("unusable as hash key: %s", indexes[0].Type())
		}
		pair, ok := hash.Pairs[key.HashKey()]
		hashOld := hash
		preserveConst := false

		if !ok && classType(hash) == "Instance" {
			preserveConst = true
			hash, _ = template(hash)
			pair, ok = hash.Pairs[key.HashKey()]
		}

		if flag == Super {
			ok = false
		}

		if !ok {
			preserveConst = true
			for flag != Current {
				if hash, ok = template(hash); !ok {
					break
				}
				if pair, ok = hash.Pairs[key.HashKey()]; ok {
					break
				}
			}
		}

		if ok {
			refObj := pair.Value
			if refer, ok := (*refObj).(*object.Reference); ok {
				return &object.Reference{Value: refer.Value, Const: preserveConst || constObj, Origin: hashOld, Index: key}
				//return refer
			}
			return &object.Reference{Value: pair.Value, Const: preserveConst || constObj, Origin: hashOld, Index: key}
		} else {
			return &object.Reference{Value: nil, Const: constObj, Origin: hashOld, Index: key}
		}
	}
	return newError("not Array, String or Hash: %s", ident.Type())
}

func applyFunction(fn object.Object, args []object.Object, env *object.Environment) object.Object {
	if function, ok := fn.(*object.Function); ok {
		extendedEnv := extendFunctionEnv(function, args)
		evaluated := Eval(function.Body, extendedEnv)
		return object.UnwrapRetValue(evaluated)
	}

	if function, ok := fn.(*object.UnderLine); ok {
		inner := function.Env.NewEnclosedEnvironment()
		var argsRef []object.Object
		for _, arg := range args {
			if refer, ok := arg.(*object.Reference); ok {
				argsRef = append(argsRef, refer)
			} else {
				arr := arg
				argsRef = append(argsRef, &object.Reference{Value: &arr, Const: true})
			}
		}
		inner.SetCurrent("args", &object.Array{Elements: argsRef, Copyable: false})
		evaluated := Eval(function.Body, inner)
		return object.UnwrapRetValue(evaluated)
	}

	if native, ok := fn.(*object.Native); ok {
		return native.Fn(env, args)
	}

	return newError("not a function, underline function or a native function: %s", fn.Type())
}

func extendFunctionEnv(
	fn *object.Function,
	args []object.Object,
) *object.Environment {
	env := fn.Env.NewEnclosedEnvironment()

	l := len(fn.Parameters)
	if l != 0 {
		if fn.Parameters[l - 1].Value == "self" {
			args = append(args, &object.Reference{Value: &fn.Self, Const: true})
		}
	}

	for paramIdx, param := range fn.Parameters {
		if paramIdx >= len(args) {
			env.SetCurrent(param.Value, &object.Reference{Value: &object.VoidObj, Const: true})
		} else {
			if refer, ok := args[paramIdx].(*object.Reference); ok {
				env.SetCurrent(param.Value, refer)
			} else {
				env.SetCurrent(param.Value, &object.Reference{Value: &args[paramIdx], Const: true})
			}
		}
	}

	return env
}

func evalExpressions(
	exps []ast.Expression,
	env *object.Environment,
	unwrap bool,
) []object.Object {
	var result []object.Object

	for _, e := range exps {
		var evaluated object.Object
		if unwrap {
			evaluated = object.UnwrapReferenceValue(Eval(e, env))
		} else {
			evaluated = Eval(e, env)
		}
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func evalIdentifier(
	node *ast.Identifier,
	env *object.Environment,
) object.Object {
	if val, ok := env.Get(node.Value); ok {
		if refer, ok := (*val).(*object.Reference); ok {
			return refer
		}
		return &object.Reference{
			Value:  val,
			Origin: env,
			Index:  &object.String{Value: []rune(node.Value)},
			Const:  false,
		}
	}

	return newError("identifier not found: " + node.Value)
}

func evalHashLiteral(
	node *ast.HashLiteral,
	env *object.Environment,
) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := object.UnwrapReferenceValue(Eval(keyNode, env))
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.HashAble)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: &value}
	}

	return &object.Hash{Pairs: pairs}
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	result := object.VoidObj

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.RetValue:
			return result.Value
		case *object.Err:
			return result
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	result := object.VoidObj

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if isError(result) || isSkip(result) {
			return result
		}
	}

	return result
}

func evalRefStatement(node *ast.RefStatement, env *object.Environment) object.Object {
	left := Eval(node.Value, env)
	if isError(left) {
		return left
	}
	if refer, ok := left.(*object.Reference); ok {
		if refer.Value == nil {
			return newError("refer to [NOT ALLOC]: %s", left.Inspect())
		}
		if _, ok := env.SetCurrent(node.Name.Value, refer); !ok {
			return newError("identifier %s already set", left.Inspect())
		}
		return object.VoidObj
	} else {
		if _, ok := env.SetCurrent(node.Name.Value, &object.Reference{Value: &left, Const: true}); !ok {
			return newError("identifier %s already set", left.Inspect())
		}
		return object.VoidObj
	}
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "+":
		return evalPlusPrefixOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalAssignExpression(node *ast.AssignExpression, env *object.Environment) object.Object {
	val := object.UnwrapReferenceValue(Eval(node.Value, env))
	if isError(val) {
		return val
	}

	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}
	if refer, ok := left.(*object.Reference); ok {
		if refer.Const {
			return newError("assign to const reference")
		}
		if refer.Value == nil {
			if refer.Origin == nil {
				return newError("assign to empty reference with no alloc function")
			}
			if refer.Value, ok = refer.Origin.DoAlloc(refer.Index); !ok {
				return newError("assign to empty reference with alloc function failed")
			}
		}
		var newVal object.Object
		switch node.Operator {
		case "+=":
			newVal = evalInfixExpression("+", *refer.Value, val)
		case "-=":
			newVal = evalInfixExpression("-", *refer.Value, val)
		case "*=":
			newVal = evalInfixExpression("*", *refer.Value, val)
		case "/=":
			newVal = evalInfixExpression("/", *refer.Value, val)
		case "%=":
			newVal = evalInfixExpression("%", *refer.Value, val)
		case "=":
			newVal = val.Copy()
		}
		if isError(newVal) {
			return newVal
		}
		*refer.Value = newVal
		return newVal
	}
	return newError("left value not Reference: %s", left.Inspect())
}

func evalInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	switch {
	case left.Type() == object.INTEGER || left.Type() == object.FLOAT:
		if right.Type() == object.INTEGER || right.Type() == object.FLOAT {
			return evalNumberInfixExpression(operator, left, right)
		}
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())

	case left.Type() == object.BOOLEAN:
		if right.Type() == object.BOOLEAN {
			return evalBooleanInfixExpression(operator, left.(*object.Boolean), right.(*object.Boolean))
		}
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())

	case left.Type() == object.STRING || left.Type() == object.CHARACTER:
		if right.Type() == object.STRING || right.Type() == object.CHARACTER {
			return evalStringInfixExpression(operator, left.(object.Letter).LetterObj(), right.(object.Letter).LetterObj())
		}
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())

	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalBooleanInfixExpression(
	operator string,
	left, right *object.Boolean,
) object.Object {
	switch operator {
	case "==":
		return nativeBoolToBooleanObject(left == right)
	case "!=":
		return nativeBoolToBooleanObject(left != right)
	case "and":
		return nativeBoolToBooleanObject(left.Value && right.Value)
	case "or":
		return nativeBoolToBooleanObject(left.Value || right.Value)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(
	operator string,
	left, right string,
) object.Object {
	switch operator {
	case "==":
		return nativeBoolToBooleanObject(left == right)
	case "!=":
		return nativeBoolToBooleanObject(left != right)
	case "+":
		return &object.String{Value: []rune(left + right)}
	default:
		return newError("unknown operator: %s %s %s",
			object.STRING, operator, object.STRING)
	}
}

func evalNumberInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	switch left.Type() {
	case object.INTEGER:
		return evalIntegerInfixExpression(operator, left, right)
	case object.FLOAT:
		return evalFloatInfixExpression(operator, left, right)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	leftVal := left.(*object.Integer).Value
	switch right.Type() {
	case object.INTEGER:
		rightVal := right.(*object.Integer).Value
		switch operator {
		case "+":
			return &object.Integer{Value: leftVal + rightVal}
		case "-":
			return &object.Integer{Value: leftVal - rightVal}
		case "*":
			return &object.Integer{Value: leftVal * rightVal}
		case "/":
			return &object.Float{Value: float64(leftVal) / float64(rightVal)}
		case "%":
			return &object.Integer{Value: leftVal % rightVal}

		case "<":
			return nativeBoolToBooleanObject(leftVal < rightVal)
		case ">":
			return nativeBoolToBooleanObject(leftVal > rightVal)
		case "<=":
			return nativeBoolToBooleanObject(leftVal <= rightVal)
		case ">=":
			return nativeBoolToBooleanObject(leftVal >= rightVal)
		case "==":
			return nativeBoolToBooleanObject(leftVal == rightVal)
		case "!=":
			return nativeBoolToBooleanObject(leftVal != rightVal)

		default:
			return newError("unknown operator: %s %s %s",
				left.Type(), operator, right.Type())
		}
	case object.FLOAT:
		rightVal := right.(*object.Float).Value
		switch operator {
		case "+":
			return &object.Float{Value: float64(leftVal) + rightVal}
		case "-":
			return &object.Float{Value: float64(leftVal) - rightVal}
		case "*":
			return &object.Float{Value: float64(leftVal) * rightVal}
		case "/":
			return &object.Float{Value: float64(leftVal) / rightVal}

		case "<":
			return nativeBoolToBooleanObject(float64(leftVal) < rightVal)
		case ">":
			return nativeBoolToBooleanObject(float64(leftVal) > rightVal)
		case "<=":
			return nativeBoolToBooleanObject(float64(leftVal) <= rightVal)
		case ">=":
			return nativeBoolToBooleanObject(float64(leftVal) >= rightVal)
		case "==":
			return nativeBoolToBooleanObject(float64(leftVal) == rightVal)
		case "!=":
			return nativeBoolToBooleanObject(float64(leftVal) != rightVal)

		default:
			return newError("unknown operator: %s %s %s",
				left.Type(), operator, right.Type())
		}
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalFloatInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	leftVal := left.(*object.Float).Value
	switch right.Type() {
	case object.INTEGER:
		rightVal := right.(*object.Integer).Value
		switch operator {
		case "+":
			return &object.Float{Value: leftVal + float64(rightVal)}
		case "-":
			return &object.Float{Value: leftVal - float64(rightVal)}
		case "*":
			return &object.Float{Value: leftVal * float64(rightVal)}
		case "/":
			return &object.Float{Value: leftVal / float64(rightVal)}

		case "<":
			return nativeBoolToBooleanObject(leftVal < float64(rightVal))
		case ">":
			return nativeBoolToBooleanObject(leftVal > float64(rightVal))
		case "<=":
			return nativeBoolToBooleanObject(leftVal <= float64(rightVal))
		case ">=":
			return nativeBoolToBooleanObject(leftVal >= float64(rightVal))
		case "==":
			return nativeBoolToBooleanObject(leftVal == float64(rightVal))
		case "!=":
			return nativeBoolToBooleanObject(leftVal != float64(rightVal))

		default:
			return newError("unknown operator: %s %s %s",
				left.Type(), operator, right.Type())
		}
	case object.FLOAT:
		rightVal := right.(*object.Float).Value
		switch operator {
		case "+":
			return &object.Float{Value: leftVal + rightVal}
		case "-":
			return &object.Float{Value: leftVal - rightVal}
		case "*":
			return &object.Float{Value: leftVal * rightVal}
		case "/":
			return &object.Float{Value: leftVal / rightVal}

		case "<":
			return nativeBoolToBooleanObject(leftVal < rightVal)
		case ">":
			return nativeBoolToBooleanObject(leftVal > rightVal)
		case "<=":
			return nativeBoolToBooleanObject(leftVal <= rightVal)
		case ">=":
			return nativeBoolToBooleanObject(leftVal >= rightVal)
		case "==":
			return nativeBoolToBooleanObject(leftVal == rightVal)
		case "!=":
			return nativeBoolToBooleanObject(leftVal != rightVal)

		default:
			return newError("unknown operator: %s %s %s",
				left.Type(), operator, right.Type())
		}
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	return nativeBoolToBooleanObject(toBoolean(right) == object.FalseObj)
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	switch right.Type() {
	case object.INTEGER:
		value := right.(*object.Integer).Value
		return &object.Integer{Value: -value}
	case object.FLOAT:
		value := right.(*object.Float).Value
		return &object.Float{Value: -value}
	}
	return newError("unknown operator: -%s", right.Type())
}

func evalPlusPrefixOperatorExpression(right object.Object) object.Object {
	switch right.Type() {
	case object.INTEGER:
		return right
	case object.FLOAT:
		return right
	}
	return newError("unknown operator: +%s", right.Type())
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := object.UnwrapReferenceValue(Eval(ie.Condition, env))
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return object.VoidObj
	}
}

func evalLoopExpression(le *ast.LoopExpression, env *object.Environment) object.Object {
	result := object.VoidObj

	condition := object.UnwrapReferenceValue(Eval(le.Condition, env))
	if isError(condition) {
		return condition
	}

	for isTruthy(condition) {
		newResult := Eval(le.Body, env.NewEnclosedEnvironment())
		if isError(newResult) || newResult.Type() == object.RET {
			return newResult
		}

		if newResult.Type() == object.OUT {
			return object.UnwrapOutValue(newResult)
		}

		if newResult.Type() != object.JUMP {
			result = newResult
		}

		condition = object.UnwrapReferenceValue(Eval(le.Condition, env))
		if isError(condition) {
			return condition
		}
	}
	return result
}

func evalLoopInExpression(le *ast.LoopInExpression, env *object.Environment) object.Object {
	result := object.VoidObj

	loopRange := Eval(le.Range, env)
	if isError(loopRange) {
		return loopRange
	}

	if f, ok := env.Get("len"); ok {
		length := applyFunction(*f, []object.Object{loopRange}, env)
		if isError(length) {
			return length
		}

		if f, ok := env.Get("subscript"); ok {
			for i := int64(0); i < length.(*object.Integer).Value; i++ {
				newEnv := env.NewEnclosedEnvironment()
				v := applyFunction(*f, []object.Object{loopRange, &object.Array{Elements: []object.Object{&object.Integer{Value: i}}}}, env)
				if isError(v) {
					return v
				}
				newEnv.SetCurrent(le.Name.Value, v)
				newResult := Eval(le.Body, newEnv)
				if isError(newResult) || newResult.Type() == object.RET {
					return newResult
				}

				if newResult.Type() == object.OUT {
					return object.UnwrapOutValue(newResult)
				}

				if newResult.Type() != object.JUMP {
					result = newResult
				}
			}
			return result
		}
		return newError("subscript lost")
	}
	return newError("len")
}

func isTruthy(obj object.Object) bool {
	return toBoolean(obj) == object.TrueObj
}

func toBoolean(number object.Object) object.Object {
	switch number.Type() {
	case object.INTEGER:
		if number.(*object.Integer).Value != 0 {
			return object.TrueObj
		}
		return object.FalseObj
	case object.FLOAT:
		if number.(*object.Float).Value != 0 && !math.IsNaN(number.(*object.Float).Value) {
			return object.TrueObj
		}
		return object.FalseObj
	case object.BOOLEAN:
		return number
	case object.VOID:
		return object.FalseObj
	}
	return newError("could not parse %s as boolean", number.Inspect())
}
