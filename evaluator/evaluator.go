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
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

/*
#cgo LDFLAGS: -lffi
#include <dlfcn.h>
#include <ffi.h>
#include <string.h>
 */
import "C"

func PrintParserErrors(out io.Writer, errors []string) {
	_, _ = io.WriteString(out, "PARSER ERRORS:\n")
	for _, msg := range errors {
		_, _ = io.WriteString(out, "    "+msg+"\n")
	}
}

func makeObjectPointer(obj object.Object) *object.Object {
	return &obj
}

func init() {
	SharedEnv = object.NewEnvironment(&map[string]*object.Object{
		"cdlOpen": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function cdlOpen: len(args) should be 1")
			}
			if str, ok := object.UnwrapReferenceValue(args[0]).(*object.String); ok {
				return &object.Integer{Value: int64(uintptr(C.dlopen(C.CString(string(str.Value)), 1)))}
			}
			return newError("native function cdlOpen: arg should be String")
		}}),
		"cdlSym": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 2 {
				return newError("native function cdlSym: len(args) should be 2")
			}
			if i, ok := object.UnwrapReferenceValue(args[0]).(*object.Integer); ok {
				if str, ok := object.UnwrapReferenceValue(args[1]).(*object.String); ok {
					s := strconv.FormatInt(int64(uintptr(
						C.dlsym(unsafe.Pointer(uintptr(i.Value)), C.CString(string(str.Value))),
					)), 10)
					c := code(`
						#.CFunctionP(` + s + `);
					`, env)
					return c
				}
				return newError("native function cdlSym: args[1] should be String")
			}
			return newError("native function cdlSym: args[0] should be Int")
		}}),
		"cdlCall": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 4 {
				return newError("native function cdlCall: len(args) should be 3")
			}

			if i, ok := object.UnwrapReferenceValue(args[0]).(*object.Integer); ok {
				if arrType, ok := object.UnwrapReferenceValue(args[1]).(*object.Array); ok {
					if arrValue, ok := object.UnwrapReferenceValue(args[2]).(*object.Array); ok {
						if retType, ok := object.UnwrapReferenceValue(args[3]).(*object.String); ok {
							return applyCdlCall(i.Value, arrType.Elements, arrValue.Elements, object.TypeC(retType.Value), env)
						}
						return newError("native function cdlCall: args[3] should be String")
					}
					return newError("native function cdlCall: args[2] should be Array")
				}
				return newError("native function cdlCall: args[1] should be Array")
			}
			return newError("native function cdlSym: args[0] should be Int")
		}}),
		"super": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 2 {
				return newError("native function super: len(args) should be 2")
			}
			return applyIndex(args[0], []object.Object{args[1]}, Super, env)
		}}),
		"current": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 2 {
				return newError("native function current: len(args) should be 2")
			}
			return applyIndex(args[0], []object.Object{args[1]}, Current, env)
			//TODO: here is a bug on Current
			//TODO: Maybe Not, because of @class is not defined
		}}),
		"classType": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function classType: len(args) should be 1")
			}
			if h, ok := object.UnwrapReferenceValue(args[0]).(*object.Hash); ok {
				return &object.String{Value: []rune(classType(h))}
			}
			return newError("native function classType: arg should be Hash")
		}}),
		"call": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 2 {
				return newError("native function call: len(args) should be 2")
			}
			v := object.UnwrapReferenceValue(args[1])
			if arr, ok := v.(*object.Array); ok {
				return applyCall(args[0], arr.Elements, env)
			}
			return newError("native function call: args[1] should be Array")
		}}),
		"subscript": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 2 {
				return newError("native function subscript: len(args) should be 2")
			}
			v := object.UnwrapReferenceValue(args[1])
			if arr, ok := v.(*object.Array); ok {
				return applyIndex(args[0], arr.Elements, Default, env)
			}
			return newError("native function subscript: args[1] should be Array")
		}}),
		"len": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function len: len(args) should be 1")
			}
			return getLen(object.UnwrapReferenceValue(args[0]), env)
		}}),
		"print": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			for _, arg := range args {
				fmt.Print(string(toString(object.UnwrapReferenceValue(arg), env).(*object.String).Value))
			}
			return object.VoidObj
		}}),
		"input": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 0 {
				return newError("native function input: len(args) should be 0")
			}
			var input string
			_, _ = fmt.Scanf("%s", &input)

			return &object.String{Value: []rune(input)}
		}}),
		"printLine": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) == 0 {
				fmt.Println()
				return object.VoidObj
			}
			for _, arg := range args {
				fmt.Println(string(toString(object.UnwrapReferenceValue(arg), env).(*object.String).Value))
			}
			return object.VoidObj
		}}),
		"inputLine": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 0 {
				return newError("native function inputLine: len(args) should be 0")
			}
			data, _, _ := bufio.NewReader(os.Stdin).ReadLine()

			return &object.String{Value: []rune(string(data))}
		}}),
		"string": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function string: len(args) should be 1")
			}
			un := object.UnwrapReferenceValue(args[0])
			return toString(un, env)
		}}),
		"exit": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
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
		}}),
		"eval": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
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
		}}),
		"integer": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
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
		}}),

		"float": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
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
		}}),

		"boolean": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function boolean: len(args) should be 1")
			}
			return toBoolean(object.UnwrapReferenceValue(args[0]))
		}}),

		"fetch": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function fetch: len(args) should be 1")
			}
			if err, ok := args[0].(*object.Err); ok {
				return &object.String{Value: []rune(err.Inspect(16))}
			}
			return args[0]
		}}),

		"append": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 2 {
				return newError("native function append: len(args) should be 2")
			}
			if array, ok := object.UnwrapReferenceValue(args[0]).(*object.Array); ok {
				return &object.Array{Elements: append(array.Elements, object.UnwrapReferenceValue(args[1])), Copyable: true}
			}
			return newError("native function append: args[0] should be Array")
		}}),

		"first": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function first: len(args) should be 1")
			}
			constObj := true
			if refer, ok := args[0].(*object.Reference); ok {
				constObj = refer.Const
				args[0] = object.UnwrapReferenceValue(args[0])
			}
			if array, ok := args[0].(*object.Array); ok {
				if len(array.Elements) == 0 {
					return object.VoidObj
				}
				return &object.Reference{Value: &array.Elements[0], Const: constObj}
			}
			return newError("native function first: arg should be Array")
		}}),

		"last": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function fetch: len(args) should be 1")
			}
			constObj := true
			if refer, ok := args[0].(*object.Reference); ok {
				constObj = refer.Const
				args[0] = object.UnwrapReferenceValue(args[0])
			}
			if array, ok := object.UnwrapReferenceValue(args[0]).(*object.Array); ok {
				if len(array.Elements) == 0 {
					return object.VoidObj
				}
				return &object.Reference{Value: &array.Elements[len(array.Elements)-1], Const: constObj}
			}
			return newError("native function append: arg should be Array")
		}}),

		"typeFull": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function typeFull: len(args) should be 1")
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
		}}),

		"type": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function type: len(args) should be 1")
			}
			return &object.String{Value: []rune(object.UnwrapReferenceValue(args[0]).Type())}
		}}),

		"typeC": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function type: len(args) should be 1")
			}
			return &object.String{Value: []rune(object.UnwrapReferenceValue(args[0]).TypeC())}
		}}),

		"array": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
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
							e = object.UnwrapReferenceValue(applyCall(function, []object.Object{&object.Integer{Value: i}, e}, env))
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
		}}),

		"value": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function value: len(args) should be 1")
			}
			return object.UnwrapReferenceValue(args[0])
		}}),

		"echo": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function echo: len(args) should be 1")
			}
			return args[0]
		}}),

		"error": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
			if len(args) != 1 {
				return newError("native function error: len(args) should be 1")
			}
			if str, ok := object.UnwrapReferenceValue(args[0]).(*object.String); ok {
				return newError(string(str.Value))
			} else {
				return newError("native function error: arg should be String")
			}
		}}),

		"import": makeObjectPointer(&object.Native{Fn: func(env *object.Environment, args []object.Object) object.Object {
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

				importEnv := SharedEnv.NewEnclosedEnvironment()
				result := Eval(program, importEnv)
				if isError(result) {
					return result
				}
				if export, ok := importEnv.Get("export"); ok {
					return *export
				}
				return object.VoidObj
				//return newError("native function import: export obj not found")
			}
			return newError("native function import: arg should be String")
		}}),
	})
	SharedEnv.SetCurrent("#", code(`
				 {
					"C": {
						"@[]": func(args) {
							ret cdlSym(-2, args[0]);
						}
					},
					"CType": {
						"@class": "CType",
						"@()": func(args, self) {
							if (classType self == "Proto") {
								if (len args == 1) {
									ret { "@template": self, "cType": typeC(args[0]), "raw": args[0] };
								} else if (len args == 2) {
									ret { "@template": self, "cType": value(args[1]), "raw": args[0] };
								};
							};
						}
					},
					"CFunctionP": {
						"@class": "CFunctionP",
						"@()": func(args, self) {
							if (classType self == "Proto") {
								ret { "@template": value(self), "id": value(args[0]) };
							} else if (classType self == "Instance") {
								ret call(#.CFunction(self.id, "void"), args);
							};
						},
						"@[]": func(args, self) {
							ret #.CFunction(self.id, args[0]);
						}
					},
					"CFunction": {
						"@class": "CFunction",
						"@()": func(args, self) {
							if (classType self == "Proto") {
								ret { "@template": value(self), "id": value(args[0]), "retType": value(args[1]) };
							} else if (classType self == "Instance") {
								let tps = [];
								let ags = args;
								loop v in ags {
									tps = append(tps, 
										if (type v == "Hash") {
											let r = v.cType;
											v = v.raw;
											r;
										} else {
											typeC v;
										}
									);
								};
								ret cdlCall(self.id, tps, ags, self.retType);
							};
						}
					},
					"max": _ {
						if (len(args) == 0) {
							ret void;
						};
						if (len(args) == 1 and type(args[0]) == "Array") {
							if (len(args[0]) == 0) {
								ret void;
							};
							let maximum = args[0][0];
							loop x in args[0] {
								maximum = if (x > maximum) { x; } else { maximum; };
							};
							ret maximum;
						} else {
							ret #.max(args);
						};
					},
				
					"min": _ {
						if (len(args) == 0) {
							ret void;
						};
						if (len(args) == 1 and type(args[0]) == "Array") {
							if (len(args[0]) == 0) {
								ret void;
							};
							let minimum = args[0][0];
							loop x in args[0] {
								minimum = if (x < minimum) { x; } else { minimum; };
							};
							ret minimum;
						} else {
							ret #.min(args);
						};
					},
				
					"abs": _ {
						if (len(args) != 1) {
							ret void;
						};
						ret if (args[0] < 0) { -args[0]; } else { args[0]; };
					},
				
					"sqrt": _ {
						if (len(args) != 1) {
							ret void;
						};
						let L = 0;
						let R = #.max(1, args[0]);
						ret integer((loop (R - L >= 1e-12) {
							let M = (L + R) / 2;
							let K = M * M;
							if (#.abs(K - args[0]) <= 1e-12) {
								out M;
							};
							if (K > args[0]) {
								R;
							} else if (K < args[0]) {
								L;
							} = M;
						} * 1e11 + 5) / 10) / 1e10;
					},
				
					"about": _ {
						printLine();
						printLine("TLang by mark07x");
						printLine("T Language v0.1");
						printLine("TLang Standard Library v0.1");
						printLine();
						printLine("Hello World, Mark!");
						printLine();
					}
				};`, SharedEnv))
}

var SharedEnv *object.Environment

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
		return &object.Function{Parameters: params, Env: env, Body: body, Self: object.VoidObj}
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
		return evalInfixExpression(node.Operator, left, right, env)

	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.LoopExpression:
		return evalLoopExpression(node, env)
	case *ast.LoopInExpression:
		return evalLoopInExpression(node, env)
	case *ast.AssignExpression:
		return evalAssignExpression(node, env)
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env, false)
		if len(args) == 1 && isError(args[0]) {
			f, _ := SharedEnv.Get("fetch")
			if function != *f {
				return args[0]
			}
		}

		return applyCall(function, args, env)
	case *ast.IndexExpression:
		ident := Eval(node.Left, env)
		if isError(ident) {
			return ident
		}
		indexes := evalExpressions(node.Indexes, env, true)
		if len(indexes) == 1 && isError(indexes[0]) {
			return indexes[0]
		}
		return applyIndex(ident, indexes, Default, env)
	case *ast.DotExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		if str, ok := node.Right.(*ast.Identifier); ok {
			return applyIndex(left, []object.Object{&object.String{Value: []rune(str.Value)}}, Default, env)
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
					return newError("delete a constant reference: %s", refer.Inspect(16))
				}
				if refer.Origin != nil {
					if !refer.Origin.DeAlloc(refer.Index) {
						return newError("unable to dealloc: %s", refer.Inspect(16))
					}
					return object.VoidObj
				}
			}
			return newError("left value not Identifier or Allocable: %s", node.DelIdent.String())
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

func applyIndex(obj object.Object, indexes []object.Object, flag classFlag, env *object.Environment) object.Object {
	constObj := true
	if refer, ok := obj.(*object.Reference); ok {
		constObj = refer.Const
		obj = object.UnwrapReferenceValue(obj)
	}

	if arr, ok := obj.(*object.Array); ok {
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
	if str, ok := obj.(*object.String); ok {
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
	if hash, ok := obj.(*object.Hash); ok {
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
			}
			return &object.Reference{Value: pair.Value, Const: preserveConst || constObj, Origin: hashOld, Index: key}
		} else {
			if s, ok := key.HashKey().Value.(string); !ok || !strings.HasPrefix(s, "@") {
				ref := applyIndex(obj, []object.Object{&object.String{Value: []rune("@[]")}}, Default, env).(*object.Reference)
				if ref.Value != nil {
					return applyCall(ref, []object.Object{&object.Array{Elements: indexes}}, env)
				}
			}
			return &object.Reference{Value: nil, Const: constObj, Origin: hashOld, Index: key}
		}
	}
	return newError("not Array, String or Hash: %s", obj.Type())
}

func applyCdlCall(id int64, argsType []object.Object, argsValue []object.Object, retType object.TypeC, env *object.Environment) object.Object {
	if len(argsType) != len(argsValue) {
		return newError("len(argsType) != len(argsValue)")
	}
	l := len(argsType)

	var cif = (*C.ffi_cif)(C.malloc(C.sizeof_ffi_cif))
	var argsTypeFFIRaw = C.malloc(C.ulong(C.sizeof_size_t * l))
	var argsValueFFIRaw = C.malloc(C.ulong(C.sizeof_size_t * l))

	var argsTypeFFI = *(*[]*C.ffi_type)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(argsTypeFFIRaw),
		Len:  l,
		Cap:  l,
	}))
	var argsValueFFI = *(*[]unsafe.Pointer)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(argsValueFFIRaw),
		Len:  l,
		Cap:  l,
	}))
	for i := 0; i < l; i++ {
		if str, ok := argsType[i].(*object.String); ok {
			typeName := object.TypeC(str.Value)
			switch typeName {
			case "long long":
				argsTypeFFI[i] = &C.ffi_type_sint64
				cMem := C.malloc(C.sizeof_longlong)
				*(*int64)(unsafe.Pointer(cMem)) = object.UnwrapReferenceValue(argsValue[i]).(*object.Integer).Value
				argsValueFFI[i] = cMem
			case "int":
				sint := C.ffi_type_sint
				argsTypeFFI[i] = &sint
				cMem := C.malloc(C.sizeof_int)
				*(*int)(unsafe.Pointer(cMem)) = int(object.UnwrapReferenceValue(argsValue[i]).(*object.Integer).Value)
				argsValueFFI[i] = cMem
			case "double":
				argsTypeFFI[i] = &C.ffi_type_double
				cMem := C.malloc(C.sizeof_double)
				*(*float64)(unsafe.Pointer(cMem)) = object.UnwrapReferenceValue(argsValue[i]).(*object.Float).Value
				argsValueFFI[i] = cMem
			case "pointer":
				argsTypeFFI[i] = &C.ffi_type_pointer
				cMem := C.malloc(C.sizeof_size_t)
				switch v := object.UnwrapReferenceValue(argsValue[i]).(type) {
				case *object.Integer:
					*(*unsafe.Pointer)(unsafe.Pointer(cMem)) = unsafe.Pointer(uintptr(v.Value))
				case *object.String:
					*(*unsafe.Pointer)(unsafe.Pointer(cMem)) = unsafe.Pointer(C.CString(string(v.Value)))
				default:
					*(*unsafe.Pointer)(unsafe.Pointer(cMem)) = unsafe.Pointer(uintptr(0))
				}
				argsValueFFI[i] = cMem
			default:
				sint := C.ffi_type_sint
				argsTypeFFI[i] = &sint
				cMem := C.malloc(C.sizeof_int)
				*(*int)(unsafe.Pointer(cMem)) = 0
				argsValueFFI[i] = cMem
			}
		} else {
			return newError("Function args type not string")
		}
	}
	var rc unsafe.Pointer
	var rt *C.ffi_type
	switch retType {
	case "long long":
		rc = C.malloc(C.sizeof_longlong)
		rt = &C.ffi_type_sint64
	case "int":
		rc = C.malloc(C.sizeof_int)
		sint := C.ffi_type_sint
		rt = &sint
	case "double":
		rc = C.malloc(C.sizeof_double)
		rt = &C.ffi_type_double
	case "pointer":
		rc = C.malloc(C.sizeof_size_t)
		rt = &C.ffi_type_pointer
	default:
		rc = C.malloc(C.sizeof_void)
		rt = &C.ffi_type_void
	}
	if C.ffi_prep_cif(cif, C.FFI_DEFAULT_ABI, C.uint(l),
			rt, (**C.ffi_type)(argsTypeFFIRaw)) == C.FFI_OK {
		C.ffi_call(cif, (*[0]byte)(unsafe.Pointer(uintptr(id))), rc, (*unsafe.Pointer)(argsValueFFIRaw))
		switch retType {
		case "long long":
			return &object.Integer{Value: *(*int64)(rc)}
		case "int":
			return code("#.CType(" + strconv.Itoa(*(*int)(rc)) + ", \"int\");", env)
		case "double":
			return &object.Float{Value: *(*float64)(rc)}
		case "pointer":
			return code("#.CType(" + strconv.FormatInt(int64(uintptr(*(*unsafe.Pointer)(rc))), 10) + ", \"pointer\");", env)
		default:
			return object.VoidObj
		}
	}
	return newError("C Function Produce Failed")
}

func applyCall(fn object.Object, args []object.Object, env *object.Environment) object.Object {
	if _, ok := fn.(*object.Reference); ok {
		fn = object.UnwrapReferenceValue(fn)
	}
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

	if hash, ok := fn.(*object.Hash); ok {
		ref := applyIndex(hash, []object.Object{&object.String{Value: []rune("@()")}}, Default, env).(*object.Reference)
		if ref.Value != nil {
			return applyCall(ref, []object.Object{&object.Array{Elements: args}}, env)
		}
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
		if fn.Parameters[l - 1].Value == "self" || fn.Parameters[l - 1].Value == "selfChangeable" {
			for len(args) < len(fn.Parameters) - 1{
				args = append(args, object.VoidObj)
			}
			args = append(args, &object.Reference{Value: &fn.Self, Const: fn.Parameters[l - 1].Value == "self"})
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
			return newError("refer to [NOT ALLOC]: %s", left.Inspect(16))
		}
		if _, ok := env.SetCurrent(node.Name.Value, refer); !ok {
			return newError("identifier %s already set", left.Inspect(16))
		}
		return object.VoidObj
	} else {
		if _, ok := env.SetCurrent(node.Name.Value, &object.Reference{Value: &left, Const: true}); !ok {
			return newError("identifier %s already set", left.Inspect(16))
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
		var newVal object.Object
		switch node.Operator {
		case "+=":
			newVal = evalInfixExpression("+", *refer.Value, val, env)
		case "-=":
			newVal = evalInfixExpression("-", *refer.Value, val, env)
		case "*=":
			newVal = evalInfixExpression("*", *refer.Value, val, env)
		case "/=":
			newVal = evalInfixExpression("/", *refer.Value, val, env)
		case "%=":
			newVal = evalInfixExpression("%", *refer.Value, val, env)
		case "=":
			newVal = val.Copy()
		}
		if isError(newVal) {
			return newVal
		}
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
		*refer.Value = newVal
		return newVal
	}
	return newError("left value not Reference: %s", left.Inspect(16))
}

func evalInfixExpression(
	operator string,
	left, right object.Object,
	env *object.Environment,
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

	case left.Type() == object.HASH:
		return evalHashInfixExpression(operator, left.(*object.Hash), right, env)

	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalHashInfixExpression(
	operator string,
	left *object.Hash,
	right object.Object,
	env *object.Environment,
) object.Object {
	return applyCall(
		applyIndex(left, []object.Object{&object.String{Value: []rune("@" + operator)}}, Default, env),
		[]object.Object{right, left},
		env,
		)
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
		length := applyCall(*f, []object.Object{loopRange}, env)
		if isError(length) {
			return length
		}

		for i := int64(0); i < length.(*object.Integer).Value; i++ {
			newEnv := env.NewEnclosedEnvironment()
			v := applyIndex(loopRange, []object.Object{&object.Integer{Value: i}}, Default, env)
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
	return newError("len")
}

func isTruthy(obj object.Object) bool {
	return toBoolean(obj) == object.TrueObj
}

func toBoolean(obj object.Object) object.Object {
	switch obj.Type() {
	case object.INTEGER:
		if obj.(*object.Integer).Value != 0 {
			return object.TrueObj
		}
		return object.FalseObj
	case object.FLOAT:
		if obj.(*object.Float).Value != 0 && !math.IsNaN(obj.(*object.Float).Value) {
			return object.TrueObj
		}
		return object.FalseObj
	case object.BOOLEAN:
		return obj
	case object.VOID:
		return object.FalseObj
	}
	return newError("could not parse %s as boolean", obj.Inspect(16))
}

func toString(obj object.Object, env *object.Environment) object.Object {
	if str, ok := obj.(*object.String); ok {
		return str
	}
	if ch, ok := obj.(*object.Character); ok {
		return &object.String{Value: []rune{ch.Value}}
	}
	if hash, ok := obj.(*object.Hash); ok {
		ref := applyIndex(hash, []object.Object{&object.String{Value: []rune("@string")}}, Default, env).(*object.Reference)
		if ref.Value != nil {
			return object.UnwrapReferenceValue(applyCall(ref, []object.Object{}, env))
		}
	}
	return &object.String{Value: []rune(obj.Inspect(16))}
}

func getLen(obj object.Object, env *object.Environment) object.Object {
	switch arg := object.UnwrapReferenceValue(obj).(type) {
	case *object.String:
		return &object.Integer{Value: int64(len(arg.Value))}
	case *object.Array:
		return &object.Integer{Value: int64(len(arg.Elements))}
	case *object.Hash:
		ref := applyIndex(arg, []object.Object{&object.String{Value: []rune("@len")}}, Default, env).(*object.Reference)
		if ref.Value != nil {
			return object.UnwrapReferenceValue(applyCall(ref, []object.Object{}, env))
		}
		return newError("native function len: arg should be String or Array")
	default:
		return newError("native function len: arg should be String or Array")
	}
}

func code(str string, env *object.Environment) object.Object {
	return Eval(parser.New(lexer.New(str)).ParseProgram(), env)
}