package evaluator

import (
	"bufio"
	"fmt"
	"github.com/mark07x/TLang/ast"
	"github.com/mark07x/TLang/lexer"
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
#include <memory.h>
#include <stdlib.h>

void * TStringObj(void * ptr) {
	return ptr;
}

void * CStringPtr(void * ptr) {
	void * str = malloc(sizeof(char *) * strlen(ptr));
    memcpy(str, ptr, strlen(ptr));
	return str;
}
*/
import "C"

func PrintParserErrors(out io.Writer, errors []string) {
	_, _ = io.WriteString(out, "PARSER ERRORS:\n")
	for _, msg := range errors {
		_, _ = io.WriteString(out, "    "+msg+"\n")
	}
}

func makeObjectPointer(obj Object) *Object {
	return &obj
}

func init() {
	SharedEnv = NewEnvironment(&map[string]*Object{
		"cdlOpen": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 1 {
				return newError("native function cdlOpen: len(args) should be 1")
			}
			if str, ok := UnwrapReferenceValue(args[0]).(*String); ok {
				return &Integer{Value: int64(uintptr(C.dlopen(C.CString(string(str.Value)), 1)))}
			}
			return newError("native function cdlOpen: arg should be String")
		}}),
		"cdlSym": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 2 {
				return newError("native function cdlSym: len(args) should be 2")
			}
			if i, ok := UnwrapReferenceValue(args[0]).(*Integer); ok {
				if str, ok := UnwrapReferenceValue(args[1]).(*String); ok {
					s := strconv.FormatInt(int64(uintptr(
						C.dlsym(unsafe.Pointer(uintptr(i.Value)), C.CString(string(str.Value))),
					)), 10)
					c := code(`
						#.CFunction(`+s+`, "void");
					`, env)
					return c
				}
				return newError("native function cdlSym: args[1] should be String")
			}
			return newError("native function cdlSym: args[0] should be Int")
		}}),
		"cdlCall": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 4 {
				return newError("native function cdlCall: len(args) should be 3")
			}

			if i, ok := UnwrapReferenceValue(args[0]).(*Integer); ok {
				if arrType, ok := UnwrapReferenceValue(args[1]).(*Array); ok {
					if arrValue, ok := UnwrapReferenceValue(args[2]).(*Array); ok {
						if retType, ok := UnwrapReferenceValue(args[3]).(*String); ok {
							return applyCdlCall(i.Value, arrType.Elements, arrValue.Elements, TypeC(retType.Value), env)
						}
						return newError("native function cdlCall: args[3] should be String")
					}
					return newError("native function cdlCall: args[2] should be Array")
				}
				return newError("native function cdlCall: args[1] should be Array")
			}
			return newError("native function cdlSym: args[0] should be Int")
		}}),
		"super": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 2 {
				return newError("native function super: len(args) should be 2")
			}
			return applyIndex(args[0], []Object{args[1]}, Super, env)
		}}),
		"current": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 2 {
				return newError("native function current: len(args) should be 2")
			}
			return applyIndex(args[0], []Object{args[1]}, Current, env)
			//TODO: here is a bug on Current
			//TODO: Maybe Not, because of @class is not defined
		}}),
		"classType": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 1 {
				return newError("native function classType: len(args) should be 1")
			}
			if h, ok := UnwrapReferenceValue(args[0]).(*Hash); ok {
				return &String{Value: []rune(classType(h))}
			}
			return newError("native function classType: arg should be Hash")
		}}),
		"call": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 2 {
				return newError("native function call: len(args) should be 2")
			}
			v := UnwrapReferenceValue(args[1])
			if arr, ok := v.(*Array); ok {
				return applyCall(args[0], arr.Elements, env)
			}
			return newError("native function call: args[1] should be Array")
		}}),
		"subscript": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 2 {
				return newError("native function subscript: len(args) should be 2")
			}
			v := UnwrapReferenceValue(args[1])
			if arr, ok := v.(*Array); ok {
				return applyIndex(args[0], arr.Elements, Default, env)
			}
			return newError("native function subscript: args[1] should be Array")
		}}),
		"len": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 1 {
				return newError("native function len: len(args) should be 1")
			}
			return getLen(UnwrapReferenceValue(args[0]), env)
		}}),
		"print": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			for _, arg := range args {
				fmt.Print(string(toString(UnwrapReferenceValue(arg), env).(*String).Value))
			}
			return VoidObj
		}}),
		"input": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 0 {
				return newError("native function input: len(args) should be 0")
			}
			var input string
			_, _ = fmt.Scanf("%s", &input)

			return &String{Value: []rune(input)}
		}}),
		"printLine": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) == 0 {
				fmt.Println()
				return VoidObj
			}
			for _, arg := range args {
				fmt.Println(string(toString(UnwrapReferenceValue(arg), env).(*String).Value))
			}
			return VoidObj
		}}),
		"inputLine": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 0 {
				return newError("native function inputLine: len(args) should be 0")
			}
			data, _, _ := bufio.NewReader(os.Stdin).ReadLine()

			return &String{Value: []rune(string(data))}
		}}),
		"string": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 1 {
				return newError("native function string: len(args) should be 1")
			}
			un := UnwrapReferenceValue(args[0])
			return toString(un, env)
		}}),
		"inspect": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 1 {
				return newError("native function inspect: len(args) should be 1")
			}
			un := UnwrapReferenceValue(args[0])
			return &String{Value: []rune(un.Inspect(16, env))}
		}}),
		"exit": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 1 && len(args) != 0 {
				return newError("native function exit: len(args) should be 1 or 0")
			}

			if len(args) == 1 {
				if val, ok := UnwrapReferenceValue(args[0]).(*Integer); ok {
					os.Exit(int(val.Value))
				}
				return newError("native function exit: arg should be Integer")
			}
			os.Exit(0)
			return VoidObj
		}}),
		"eval": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 1 {
				return newError("native function eval: len(args) should be 1")
			}

			if str, ok := UnwrapReferenceValue(args[0]).(*String); ok {
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
		"integer": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 1 {
				return newError("native function int: len(args) should be 1")
			}
			switch arg := UnwrapReferenceValue(args[0]).(type) {
			case *String:
				val, err := strconv.ParseInt(string(arg.Value), 10, 64)
				if err != nil {
					return newError("could not parse %s as integer", string(arg.Value))
				}
				return &Integer{Value: val}
			case *Character:
				val, err := strconv.ParseInt(string(arg.Value), 10, 64)
				if err != nil {
					return newError("could not parse %s as integer", string(arg.Value))
				}
				return &Integer{Value: val}
			case *Boolean:
				if arg.Value {
					return &Integer{Value: 1}
				} else {
					return &Integer{Value: 0}
				}
			case *Float:
				return &Integer{Value: int64(arg.Value)}
			case *Integer:
				return arg
			case *Void:
				return &Integer{Value: 0}
			default:
				return newError("native function integer: arg should be String, Boolean, Number or object.VoidObj")
			}
		}}),

		"float": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 1 {
				return newError("native function float: len(args) should be 1")
			}
			switch arg := UnwrapReferenceValue(args[0]).(type) {
			case *String:
				val, err := strconv.ParseFloat(string(arg.Value), 64)
				if err != nil {
					return newError("could not parse %s as float", string(arg.Value))
				}
				return &Float{Value: val}
			case *Character:
				val, err := strconv.ParseFloat(string(arg.Value), 64)
				if err != nil {
					return newError("could not parse %s as float", string(arg.Value))
				}
				return &Float{Value: val}
			case *Boolean:
				if arg.Value {
					return &Float{Value: 1.}
				} else {
					return &Float{Value: 0.}
				}
			case *Integer:
				return &Float{Value: float64(arg.Value)}
			case *Float:
				return arg
			case *Void:
				return &Float{Value: 0}
			default:
				return newError("native function int: arg should be String, Boolean, Number or object.VoidObj")
			}
		}}),

		"boolean": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 1 {
				return newError("native function boolean: len(args) should be 1")
			}
			return toBoolean(UnwrapReferenceValue(args[0]))
		}}),

		"fetch": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 1 {
				return newError("native function fetch: len(args) should be 1")
			}
			if err, ok := args[0].(*Err); ok {
				return &String{Value: []rune(err.Inspect(16, env))}
			}
			return args[0]
		}}),

		"append": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 2 {
				return newError("native function append: len(args) should be 2")
			}
			if array, ok := UnwrapReferenceValue(args[0]).(*Array); ok {
				return &Array{Elements: append(array.Elements, UnwrapReferenceValue(args[1])), Xvalue: true}
			}
			return newError("native function append: args[0] should be Array")
		}}),

		"first": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 1 {
				return newError("native function first: len(args) should be 1")
			}
			constObj := true
			if refer, ok := args[0].(*Reference); ok {
				constObj = refer.Const
				args[0] = UnwrapReferenceValue(args[0])
			}
			if array, ok := args[0].(*Array); ok {
				if len(array.Elements) == 0 {
					return VoidObj
				}
				return &Reference{Value: &array.Elements[0], Const: constObj}
			}
			return newError("native function first: arg should be Array")
		}}),

		"last": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 1 {
				return newError("native function fetch: len(args) should be 1")
			}
			constObj := true
			if refer, ok := args[0].(*Reference); ok {
				constObj = refer.Const
				args[0] = UnwrapReferenceValue(args[0])
			}
			if array, ok := UnwrapReferenceValue(args[0]).(*Array); ok {
				if len(array.Elements) == 0 {
					return VoidObj
				}
				return &Reference{Value: &array.Elements[len(array.Elements)-1], Const: constObj}
			}
			return newError("native function append: arg should be Array")
		}}),

		"type&": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 1 {
				return newError("native function type&: len(args) should be 1")
			}
			if refer, ok := args[0].(*Reference); ok {
				isConst := ""
				if refer.Const {
					isConst = "Const "
				}
				rawType := "Not Alloc"
				if refer.Value != nil {
					rawType = string((*refer.Value).Type())
				}
				return &String{Value: []rune(isConst + "Reference (" + rawType + ")")}
			}
			return &String{Value: []rune(args[0].Type())}
		}}),

		"type": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 1 {
				return newError("native function type: len(args) should be 1")
			}
			return &String{Value: []rune(UnwrapReferenceValue(args[0]).Type())}
		}}),

		"typeC": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 1 {
				return newError("native function type: len(args) should be 1")
			}
			return &String{Value: []rune(UnwrapReferenceValue(args[0]).TypeC())}
		}}),

		"array": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) == 1 {
				if length, ok := UnwrapReferenceValue(args[0]).(*Integer); ok {
					var elem []Object
					for i := int64(0); i < length.Value; i++ {
						elem = append(elem, VoidObj)
					}

					return &Array{
						Elements: elem,
						Xvalue:   true,
					}
				}
				return newError("native function array: args[0] should be Integer")
			} else if len(args) == 2 {
				if length, ok := UnwrapReferenceValue(args[0]).(*Integer); ok {
					var elem []Object
					for i := int64(0); i < length.Value; i++ {
						elem = append(elem, UnwrapReferenceValue(args[1]))
					}

					return &Array{
						Elements: elem,
						Xvalue:   true,
					}
				}
				return newError("native function array: args[0] should be Integer")
			} else if len(args) == 3 {
				if length, ok := UnwrapReferenceValue(args[0]).(*Integer); ok {
					if function, ok := UnwrapReferenceValue(args[2]).(Functor); ok {
						var elem []Object
						e := UnwrapReferenceValue(args[1])
						for i := int64(0); i < length.Value; i++ {
							e = UnwrapReferenceValue(applyCall(function, []Object{&Integer{Value: i}, e}, env))
							if isError(e) {
								return e
							}
							elem = append(elem, e)
						}

						return &Array{
							Elements: elem,
							Xvalue:   true,
						}
					}
					return newError("native function array: args[2] should be Functor")
				}
				return newError("native function array: args[0] should be Integer")
			}
			return newError("native function array: len(args) should be 1, 2 or 3")
		}}),

		"value": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 1 {
				return newError("native function value: len(args) should be 1")
			}
			return UnwrapReferenceValue(args[0])
		}}),

		"echo": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 1 {
				return newError("native function echo: len(args) should be 1")
			}
			return args[0]
		}}),

		"error": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 1 {
				return newError("native function error: len(args) should be 1")
			}
			if str, ok := UnwrapReferenceValue(args[0]).(*String); ok {
				return newError(string(str.Value))
			} else {
				return newError("native function error: arg should be String")
			}
		}}),

		"import": makeObjectPointer(&Native{Fn: func(env *Environment, args []Object) Object {
			if len(args) != 1 {
				return newError("native function import: len(args) should be 1")
			}
			if str, ok := UnwrapReferenceValue(args[0]).(*String); ok {
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
				return VoidObj
				//return newError("native function import: export obj not found")
			}
			return newError("native function import: arg should be String")
		}}),
	})
	SharedEnv.SetCurrent("#", code(`
				{
					"switch": func(v) {
						let rv = {
							"case": func(c) {
								ret func(f) {
									if (v == c) {
										f();
									};
									ret rv;
								};
							}
						};
						ret rv;
					},
					"printf": func(str, fmt) {
						print(#f(str)(fmt));
					},
					"f": func(str) {
						ret {
							"@()": func(args) {
								let s = "";
								let i = 0;
								let agi = 0;
								loop (i < len(str)) {
									let now = str[i];
									let nxt = if (i + 1 < len(str)) {
										str[i + 1];
									} else {
										' ';
									};
									if (now == '$') {
										if (nxt == '.') {
											s += string(args[agi]);
											agi += 1;
										} else if (nxt == '*') {
											s += inspect(args[agi]);
											agi += 1;
										} else if (nxt == '$') {
											s += "$";
										} else {
											error "bad format";
										};
										i += 2;
									} else {
										s += str[i];
										i += 1;
									};
								};
								ret s;
							},
							"@inspect": func(self) {
								ret #f"#f\"$.\""(str);
							}
						};
					},
					"array": func(n, v, ic) {
						if (type ic == "Void") {
							ret array(n, v);
						};
						ret array(n, v - ic, func(i, v) { ret v + ic; });
					},
					"range": func(n, v, ic) {
						if (type v == "Void") {
							v = 0;
						};
						if (type ic == "Void") {
							ic = 1;
						};
						ret #Range(n, eval(#f"func(x) { ret $. + x * $.; };"(v, ic)));
					},
					"Range": {
						"@class": "Range",
						"@()": func(args, self) {
							if (classType self == "Proto") {
								ret { "@template": self, "@len": func() { ret args[0]; }, "relation": args[1] };
							};
						},
						"@[]": func(args, self) {
							if (classType self == "Instance") {
								ret self.relation(args[0]);
							};
						},
						"@inspect": func(self) {
							if (classType self == "Proto") {
								ret "Range Creator(len, relation)";
							} else if (classType self == "Instance") {
								ret #f"Range(len: $., relation: $."(self.@len(), self.relation);
							};
						}
					},
					"commonRetType": {
						"TStringObj": "string",
						"CStringPtr": "pointer",
						"malloc": "pointer",
						"fopen": "pointer",
						"printf": "int",
						"scanf": "int",
						"fprintf": "int",
						"fscanf": "int",
						"abs": "int",
						"fabs": "double",
						"sqrt": "double",
						"@[]": _ { ret "void"; },
						"@inspect": func(self) {
							ret "commonRetType";
						}
					},
					"C": {
						"@[]": func(args) {
							let f = cdlSym(-2, args[0]);
							f.retType = #commonRetType[args[0]];
							ret f;
						}
					},
					"CType": {
						"@class": "CType",
						"@()": func(args, self) {
							if (classType self == "Proto") {
								if (len args == 1) {
									ret { "@template": self, "cType": typeC(args[0]), "raw": args[0] };
								} else if (len args == 2) {
									ret { "@template": self, "cType": args[1], "raw": args[0] };
								};
							};
						}
					},
					"CFunction": {
						"@class": "CFunction",
						"@()": func(args, self) {
							if (classType self == "Proto") {
								ret { "@template": self, "id": args[0], "retType": args[1] };
							} else if (classType self == "Instance") {
								let tps = [];
								loop &v in (args) {
									if (type &v == "Hash") {
										tps = append(tps, &v.cType);
										&v = &v.raw;
									} else {
										tps = append(tps, typeC &v);
									};
								};
								ret cdlCall(self.id, tps, args, self.retType);
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
							loop x in (args[0]) {
								maximum = if (x > maximum) { x; } else { maximum; };
							};
							ret maximum;
						} else {
							ret #max(args);
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
							loop x in (args[0]) {
								minimum = if (x < minimum) { x; } else { minimum; };
							};
							ret minimum;
						} else {
							ret #min(args);
						};
					},
				
					"abs": _ {
						if (len args != 1) {
							ret void;
						};
						ret if (args[0] < 0) { -args[0]; } else { args[0]; };
					},
				
					"sqrt": _ {
						if (len args != 1) {
							ret void;
						};
						let L = 0;
						let R = #max(1, args[0]);
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
						printLine "TLang by mark07x";
						printLine "T Language v0.1";
						printLine "TLang Standard Library v0.1";
						printLine();
						printLine "Hello World, Mark!";
						printLine();
					}
				};`, SharedEnv))
}

var SharedEnv *Environment

func newError(format string, a ...interface{}) *Err {
	return &Err{Message: fmt.Sprintf(format, a...)}
}

func isError(obj Object) bool {
	return obj.Type() == ERR
}

func isSkip(obj Object) bool {
	return obj.Type() == RET || obj.Type() == OUT || obj.Type() == JUMP
}

func nativeBoolToBooleanObject(input bool) Object {
	if input {
		return TrueObj
	}
	return FalseObj
}

func Eval(node ast.Node, env *Environment) Object {
	switch node := node.(type) {
	case *ast.Program:
		return UnwrapReferenceValue(evalProgram(node, env))

	case *ast.IntegerLiteral:
		return &Integer{Value: node.Value}
	case *ast.FloatLiteral:
		return &Float{Value: node.Value}
	case *ast.StringLiteral:
		return &String{Value: []rune(node.Value)}
	case *ast.CharacterLiteral:
		return &Character{Value: node.Value}
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &Function{Parameters: params, Env: env, Body: body, Self: VoidObj}
	case *ast.UnderLineLiteral:
		body := node.Body
		return &UnderLine{Env: env, Body: body}
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env, true)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &Array{Elements: elements, Xvalue: false}
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)

	case *ast.PrefixExpression:
		right := UnwrapReferenceValue(Eval(node.Right, env))
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := UnwrapReferenceValue(Eval(node.Left, env))
		if isError(left) {
			return left
		}

		right := UnwrapReferenceValue(Eval(node.Right, env))
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
			if UnwrapReferenceValue(function) != *f {
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
			return applyIndex(left, []Object{&String{Value: []rune(str.Value)}}, Default, env)
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
		return &RetValue{Value: val}
	case *ast.OutStatement:
		val := Eval(node.OutValue, env)
		if isError(val) {
			return val
		}
		return &OutValue{Value: val}
	case *ast.JumpStatement:
		return JumpObj
	case *ast.LetStatement:
		val := VoidObj
		if node.Value != nil {
			val = UnwrapReferenceValue(Eval(node.Value, env))
		}

		if isError(val) {
			return val
		}
		if node.Name.Value[0] == '&' {
			return evalRefStatement(node, env)
		}
		if _, ok := env.SetCurrent(node.Name.Value, val.Copy()); !ok {
			return newError("identifier %s already set", node.Name.Value)
		}
	case *ast.DelStatement:
		if ident, ok := node.DelIdent.(*ast.Identifier); ok {
			if _, ok := env.Get(ident.Value); ok {
				if !env.Free(&String{Value: []rune(ident.Value)}) {
					return newError("unable to dealloc: %s", node.DelIdent.String())
				}
			} else {
				return newError("identifier not found: %s", ident.Value)
			}
		} else {
			if refer, ok := Eval(node.DelIdent, env).(*Reference); ok {
				if refer.Const {
					return newError("delete a constant reference: %s", refer.Inspect(16, env))
				}
				if refer.Origin != nil {
					if !refer.Origin.Free(refer.Index) {
						return newError("unable to dealloc: %s", refer.Inspect(16, env))
					}
					return VoidObj
				}
			}
			return newError("left value not Identifier or Allocable: %s", node.DelIdent.String())
		}
	}

	return VoidObj
}

type classFlag int

const (
	_ classFlag = iota
	Default
	Current
	Super
)

func classType(hash *Hash) string {
	c, hasClass := hash.Pairs[HashKey{
		Type:  "String",
		Value: "@class",
	}]
	if hasClass {
		_, hasClass = (*c.Value).(*String)
	}
	t, hasTemplate := hash.Pairs[HashKey{
		Type:  "String",
		Value: "@template",
	}]
	if hasTemplate {
		_, hasTemplate = (UnwrapReferenceValue(*t.Value)).(*Hash)
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

func template(hash *Hash) (*Hash, bool) {
	if t, ok := hash.Pairs[HashKey{Type: "String", Value: "@template"}]; ok {
		if h, ok := (UnwrapReferenceValue(*t.Value)).(*Hash); ok {
			return h, true
		}
	}
	return nil, false
}

func applyIndex(obj Object, indexes []Object, flag classFlag, env *Environment) Object {
	constObj := true
	if refer, ok := obj.(*Reference); ok {
		constObj = refer.Const
		obj = UnwrapReferenceValue(obj)
	}

	if arr, ok := obj.(*Array); ok {
		if len(indexes) != 1 {
			return newError("array: len(indexes) should be 1")
		}
		if indexes[0].Type() != INTEGER {
			return newError("array: index should be Integer")
		}
		index := indexes[0].(*Integer).Value
		length := int64(len(arr.Elements))
		if index >= length || index < 0 {
			return newError("array: out of range")
		}
		refObj := &arr.Elements[index]
		if refer, ok := (*refObj).(*Reference); ok {
			return refer
		}
		return &Reference{Value: refObj, Const: constObj}
	}
	if str, ok := obj.(*String); ok {
		//runeStr := []rune(str.Value)
		if len(indexes) != 1 {
			return newError("string: len(indexes) should be 1")
		}
		if indexes[0].Type() != INTEGER {
			return newError("string: index should be Integer")
		}
		index := indexes[0].(*Integer).Value
		length := int64(len(str.Value))
		if index >= length || index < 0 {
			return newError("string: out of range")
		}
		var c Object = &Character{Value: str.Value[index]}
		return &Reference{Value: &c, Const: true}
	}
	if hash, ok := obj.(*Hash); ok {
		if len(indexes) != 1 {
			return newError("string: len(indexes) should be 1")
		}

		key, ok := indexes[0].(HashAble)
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
			if refer, ok := (*refObj).(*Reference); ok {
				return &Reference{Value: refer.Value, Const: preserveConst || constObj, Origin: hashOld, Index: key}
			}
			return &Reference{Value: pair.Value, Const: preserveConst || constObj, Origin: hashOld, Index: key}
		} else {
			if s, ok := key.HashKey().Value.(string); !ok || !strings.HasPrefix(s, "@") {
				ref := applyIndex(obj, []Object{&String{Value: []rune("@[]")}}, Default, env).(*Reference)
				if ref.Value != nil {
					return applyCall(ref, []Object{&Array{Elements: indexes}}, env)
				}
			}
			return &Reference{Value: nil, Const: constObj, Origin: hashOld, Index: key}
		}
	}
	return newError("not Array, String or Hash: %s", obj.Type())
}

func applyCdlCall(id int64, argsType []Object, argsValue []Object, retType TypeC, env *Environment) Object {
	if len(argsType) != len(argsValue) {
		return newError("len(argsType) != len(argsValue)")
	}
	l := len(argsType)

	var cif = (*C.ffi_cif)(C.malloc(C.sizeof_ffi_cif))
	var argsTypeFFIRaw = C.malloc(C.ulong(C.sizeof_size_t * l))
	var argsValueFFIRaw = C.malloc(C.ulong(C.sizeof_size_t * l))

	defer C.free(unsafe.Pointer(cif))
	defer C.free(argsTypeFFIRaw)
	defer C.free(argsValueFFIRaw)

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
		if str, ok := argsType[i].(*String); ok {
			typeName := TypeC(str.Value)
			var cMem unsafe.Pointer
			defer C.free(cMem)
			switch typeName {
			case "long long":
				argsTypeFFI[i] = &C.ffi_type_sint64
				cMem = C.malloc(C.sizeof_longlong)
				*(*int64)(cMem) = UnwrapReferenceValue(argsValue[i]).(*Integer).Value
				argsValueFFI[i] = cMem
			case "int":
				sint := C.ffi_type_sint
				argsTypeFFI[i] = &sint
				cMem = C.malloc(C.sizeof_int)
				*(*int)(cMem) = int(UnwrapReferenceValue(argsValue[i]).(*Integer).Value)
				argsValueFFI[i] = cMem
			case "double":
				argsTypeFFI[i] = &C.ffi_type_double
				cMem = C.malloc(C.sizeof_double)
				*(*float64)(cMem) = UnwrapReferenceValue(argsValue[i]).(*Float).Value
				argsValueFFI[i] = cMem
			case "pointer", "string":
				argsTypeFFI[i] = &C.ffi_type_pointer
				cMem = C.malloc(C.sizeof_size_t)
				switch v := UnwrapReferenceValue(argsValue[i]).(type) {
				case *Integer:
					*(*unsafe.Pointer)(cMem) = unsafe.Pointer(uintptr(v.Value))
				case *String:
					p := unsafe.Pointer(C.CString(string(v.Value)))
					*(*unsafe.Pointer)(cMem) = p
					defer C.free(p)
				default:
					*(*unsafe.Pointer)(cMem) = unsafe.Pointer(uintptr(0))
				}
				argsValueFFI[i] = cMem
			default:
				sint := C.ffi_type_sint
				argsTypeFFI[i] = &sint
				cMem = C.malloc(C.sizeof_int)
				*(*int)(cMem) = 0
				argsValueFFI[i] = cMem
			}
		} else {
			return newError("Function args type not string")
		}
	}
	var rc unsafe.Pointer
	defer C.free(rc)
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
	case "pointer", "string":
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
			return &Integer{Value: *(*int64)(rc)}
		case "int":
			return code("#.CType("+strconv.Itoa(*(*int)(rc))+", \"int\");", env)
		case "double":
			return &Float{Value: *(*float64)(rc)}
		case "pointer":
			return code("#.CType("+strconv.FormatInt(int64(uintptr(*(*unsafe.Pointer)(rc))), 10)+", \"pointer\");", env)
		case "string":
			return &String{Value: []rune(C.GoString(*(**C.char)(rc)))}
		default:
			return VoidObj
		}
	}
	return newError("C Function Produce Failed")
}

func applyCall(fn Object, args []Object, env *Environment) Object {
	if _, ok := fn.(*Reference); ok {
		fn = UnwrapReferenceValue(fn)
	}
	if function, ok := fn.(*Function); ok {
		extendedEnv := extendFunctionEnv(function, args)
		evaluated := Eval(function.Body, extendedEnv)
		return UnwrapRetValue(evaluated)
	}

	if function, ok := fn.(*UnderLine); ok {
		inner := function.Env.NewEnclosedEnvironment()
		var argsRef []Object
		for _, arg := range args {
			if refer, ok := arg.(*Reference); ok {
				argsRef = append(argsRef, refer)
			} else {
				arr := arg
				argsRef = append(argsRef, &Reference{Value: &arr, Const: true})
			}
		}
		in := &Array{Elements: argsRef, Xvalue: false}
		inner.SetCurrent("&args", in)
		inner.SetCurrent("args", UnwrapArrayReferenceValue(in))
		evaluated := Eval(function.Body, inner)
		return UnwrapRetValue(evaluated)
	}

	if native, ok := fn.(*Native); ok {
		return native.Fn(env, args)
	}

	if hash, ok := fn.(*Hash); ok {
		ref := applyIndex(hash, []Object{&String{Value: []rune("@()")}}, Default, env).(*Reference)
		if ref.Value != nil {
			return applyCall(ref, []Object{&Array{Elements: args}}, env)
		}
	}

	return newError("not a function, underline function or a native function: %s", fn.Type())
}

func extendFunctionEnv(
	fn *Function,
	args []Object,
) *Environment {
	env := fn.Env.NewEnclosedEnvironment()

	l := len(fn.Parameters)
	if l != 0 {
		if fn.Parameters[l-1].Value == "self" || fn.Parameters[l-1].Value == "&self" {
			for len(args) < len(fn.Parameters)-1 {
				args = append(args, VoidObj)
			}
			if fn.Parameters[l-1].Value == "&self" {
				args = append(args, &Reference{Value: &fn.Self, Const: false})
			} else {
				args = append(args, fn.Self)
			}
		}
	}

	for paramIdx, param := range fn.Parameters {
		if param.Value[0] == '&' {
			if paramIdx >= len(args) {
				env.SetCurrent(param.Value, &Reference{Value: &VoidObj, Const: true})
			} else {
				if refer, ok := args[paramIdx].(*Reference); ok {
					env.SetCurrent(param.Value, refer)
				} else {
					env.SetCurrent(param.Value, &Reference{Value: &args[paramIdx], Const: true})
				}
			}
		} else {
			if paramIdx >= len(args) {
				env.SetCurrent(param.Value, VoidObj)
			} else {
				env.SetCurrent(param.Value, UnwrapReferenceValue(args[paramIdx]))
			}
		}
	}

	return env
}

func evalExpressions(
	exps []ast.Expression,
	env *Environment,
	unwrap bool,
) []Object {
	var result []Object

	for _, e := range exps {
		var evaluated Object
		if unwrap {
			evaluated = UnwrapReferenceValue(Eval(e, env))
		} else {
			evaluated = Eval(e, env)
		}
		if isError(evaluated) {
			return []Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func evalIdentifier(
	node *ast.Identifier,
	env *Environment,
) Object {
	if val, ok := env.Get(node.Value); ok {
		if refer, ok := (*val).(*Reference); ok {
			return refer
		}
		return &Reference{
			Value:  val,
			Origin: env,
			Index:  &String{Value: []rune(node.Value)},
			Const:  false,
		}
	}

	return newError("identifier not found: " + node.Value)
}

func evalHashLiteral(
	node *ast.HashLiteral,
	env *Environment,
) Object {
	pairs := make(map[HashKey]HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := UnwrapReferenceValue(Eval(keyNode, env))
		if isError(key) {
			return key
		}

		hashKey, ok := key.(HashAble)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = HashPair{Key: key, Value: &value}
	}

	return &Hash{Pairs: pairs}
}

func evalProgram(program *ast.Program, env *Environment) Object {
	result := VoidObj

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *RetValue:
			return result.Value
		case *Err:
			return result
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *Environment) Object {
	result := VoidObj

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if isError(result) || isSkip(result) {
			return result
		}
	}

	return result
}

func evalRefStatement(node *ast.LetStatement, env *Environment) Object {
	left := Eval(node.Value, env)
	if isError(left) {
		return left
	}
	if refer, ok := left.(*Reference); ok {
		if refer.Value == nil {
			return newError("refer to [NOT ALLOC]: %s", left.Inspect(16, env))
		}
		if _, ok := env.SetCurrent(node.Name.Value, refer); !ok {
			return newError("identifier %s already set", left.Inspect(16, env))
		}
		return VoidObj
	} else {
		if _, ok := env.SetCurrent(node.Name.Value, &Reference{Value: &left, Const: true}); !ok {
			return newError("identifier %s already set", left.Inspect(16, env))
		}
		return VoidObj
	}
}

func evalPrefixExpression(operator string, right Object) Object {
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

func evalAssignExpression(node *ast.AssignExpression, env *Environment) Object {
	val := UnwrapReferenceValue(Eval(node.Value, env))
	if isError(val) {
		return val
	}

	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}
	if refer, ok := left.(*Reference); ok {
		var newVal Object
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
			if refer.Value, ok = refer.Origin.Alloc(refer.Index); !ok {
				return newError("assign to empty reference with alloc function failed")
			}
		}
		*refer.Value = newVal
		return newVal
	}
	return newError("left value not Reference: %s", left.Inspect(16, env))
}

func evalInfixExpression(
	operator string,
	left, right Object,
	env *Environment,
) Object {
	switch {
	case left.Type() == INTEGER || left.Type() == FLOAT:
		if right.Type() == INTEGER || right.Type() == FLOAT {
			return evalNumberInfixExpression(operator, left, right)
		}
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())

	case left.Type() == BOOLEAN:
		if right.Type() == BOOLEAN {
			return evalBooleanInfixExpression(operator, left.(*Boolean), right.(*Boolean))
		}
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())

	case left.Type() == STRING || left.Type() == CHARACTER:
		if right.Type() == STRING || right.Type() == CHARACTER {
			return evalStringInfixExpression(operator, left.(Letter).LetterObj(), right.(Letter).LetterObj())
		}
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())

	case left.Type() == HASH:
		return evalHashInfixExpression(operator, left.(*Hash), right, env)

	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalHashInfixExpression(
	operator string,
	left *Hash,
	right Object,
	env *Environment,
) Object {
	return applyCall(
		applyIndex(left, []Object{&String{Value: []rune("@" + operator)}}, Default, env),
		[]Object{right, left},
		env,
	)
}

func evalBooleanInfixExpression(
	operator string,
	left, right *Boolean,
) Object {
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
) Object {
	switch operator {
	case "==":
		return nativeBoolToBooleanObject(left == right)
	case "!=":
		return nativeBoolToBooleanObject(left != right)
	case "+":
		return &String{Value: []rune(left + right)}
	default:
		return newError("unknown operator: %s %s %s",
			STRING, operator, STRING)
	}
}

func evalNumberInfixExpression(
	operator string,
	left, right Object,
) Object {
	switch left.Type() {
	case INTEGER:
		return evalIntegerInfixExpression(operator, left, right)
	case FLOAT:
		return evalFloatInfixExpression(operator, left, right)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(
	operator string,
	left, right Object,
) Object {
	leftVal := left.(*Integer).Value
	switch right.Type() {
	case INTEGER:
		rightVal := right.(*Integer).Value
		switch operator {
		case "+":
			return &Integer{Value: leftVal + rightVal}
		case "-":
			return &Integer{Value: leftVal - rightVal}
		case "*":
			return &Integer{Value: leftVal * rightVal}
		case "/":
			return &Float{Value: float64(leftVal) / float64(rightVal)}
		case "%":
			return &Integer{Value: leftVal % rightVal}

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
	case FLOAT:
		rightVal := right.(*Float).Value
		switch operator {
		case "+":
			return &Float{Value: float64(leftVal) + rightVal}
		case "-":
			return &Float{Value: float64(leftVal) - rightVal}
		case "*":
			return &Float{Value: float64(leftVal) * rightVal}
		case "/":
			return &Float{Value: float64(leftVal) / rightVal}

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
	left, right Object,
) Object {
	leftVal := left.(*Float).Value
	switch right.Type() {
	case INTEGER:
		rightVal := right.(*Integer).Value
		switch operator {
		case "+":
			return &Float{Value: leftVal + float64(rightVal)}
		case "-":
			return &Float{Value: leftVal - float64(rightVal)}
		case "*":
			return &Float{Value: leftVal * float64(rightVal)}
		case "/":
			return &Float{Value: leftVal / float64(rightVal)}

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
	case FLOAT:
		rightVal := right.(*Float).Value
		switch operator {
		case "+":
			return &Float{Value: leftVal + rightVal}
		case "-":
			return &Float{Value: leftVal - rightVal}
		case "*":
			return &Float{Value: leftVal * rightVal}
		case "/":
			return &Float{Value: leftVal / rightVal}

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

func evalBangOperatorExpression(right Object) Object {
	return nativeBoolToBooleanObject(toBoolean(right) == FalseObj)
}

func evalMinusPrefixOperatorExpression(right Object) Object {
	switch right.Type() {
	case INTEGER:
		value := right.(*Integer).Value
		return &Integer{Value: -value}
	case FLOAT:
		value := right.(*Float).Value
		return &Float{Value: -value}
	}
	return newError("unknown operator: -%s", right.Type())
}

func evalPlusPrefixOperatorExpression(right Object) Object {
	switch right.Type() {
	case INTEGER:
		return right
	case FLOAT:
		return right
	}
	return newError("unknown operator: +%s", right.Type())
}

func evalIfExpression(ie *ast.IfExpression, env *Environment) Object {
	condition := UnwrapReferenceValue(Eval(ie.Condition, env))
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return VoidObj
	}
}

func evalLoopExpression(le *ast.LoopExpression, env *Environment) Object {
	result := VoidObj

	condition := UnwrapReferenceValue(Eval(le.Condition, env))
	if isError(condition) {
		return condition
	}

	for isTruthy(condition) {
		newResult := Eval(le.Body, env.NewEnclosedEnvironment())
		if isError(newResult) || newResult.Type() == RET {
			return newResult
		}

		if newResult.Type() == OUT {
			return UnwrapOutValue(newResult)
		}

		if newResult.Type() != JUMP {
			result = newResult
		}

		condition = UnwrapReferenceValue(Eval(le.Condition, env))
		if isError(condition) {
			return condition
		}
	}
	return result
}

func evalLoopInExpression(le *ast.LoopInExpression, env *Environment) Object {
	result := VoidObj

	loopRange := Eval(le.Range, env)
	if isError(loopRange) {
		return loopRange
	}
	if f, ok := env.Get("len"); ok {
		length := applyCall(*f, []Object{loopRange}, env)
		if isError(length) {
			return length
		}

		for i := int64(0); i < length.(*Integer).Value; i++ {
			newEnv := env.NewEnclosedEnvironment()
			v := applyIndex(loopRange, []Object{&Integer{Value: i}}, Default, env)
			if isError(v) {
				return v
			}
			if le.Name.Value[0] == '&' {
				newEnv.SetCurrent(le.Name.Value, v)
			} else {
				newEnv.SetCurrent(le.Name.Value, UnwrapReferenceValue(v))
			}

			newResult := Eval(le.Body, newEnv)
			if isError(newResult) || newResult.Type() == RET {
				return newResult
			}

			if newResult.Type() == OUT {
				return UnwrapOutValue(newResult)
			}

			if newResult.Type() != JUMP {
				result = newResult
			}
		}
		return result
	}
	return newError("len")
}

func isTruthy(obj Object) bool {
	return toBoolean(obj) == TrueObj
}

func toBoolean(obj Object) Object {
	switch obj.Type() {
	case INTEGER:
		if obj.(*Integer).Value != 0 {
			return TrueObj
		}
		return FalseObj
	case FLOAT:
		if obj.(*Float).Value != 0 && !math.IsNaN(obj.(*Float).Value) {
			return TrueObj
		}
		return FalseObj
	case BOOLEAN:
		return obj
	default:
		return FalseObj
	}
}

func toString(obj Object, env *Environment) Object {
	if str, ok := obj.(*String); ok {
		return str
	}
	if ch, ok := obj.(*Character); ok {
		return &String{Value: []rune{ch.Value}}
	}
	if hash, ok := obj.(*Hash); ok {
		ref := applyIndex(hash, []Object{&String{Value: []rune("@string")}}, Default, env).(*Reference)
		if ref.Value != nil {
			return UnwrapReferenceValue(applyCall(ref, []Object{}, env))
		}
	}
	return &String{Value: []rune(obj.Inspect(16, env))}
}

func getLen(obj Object, env *Environment) Object {
	switch arg := UnwrapReferenceValue(obj).(type) {
	case *String:
		return &Integer{Value: int64(len(arg.Value))}
	case *Array:
		return &Integer{Value: int64(len(arg.Elements))}
	case *Hash:
		ref := applyIndex(arg, []Object{&String{Value: []rune("@len")}}, Default, env).(*Reference)
		if ref.Value != nil {
			return UnwrapReferenceValue(applyCall(ref, []Object{}, env))
		}
		return newError("native function len: arg should be String or Array")
	default:
		return newError("native function len: arg should be String or Array")
	}
}

func code(str string, env *Environment) Object {
	return Eval(parser.New(lexer.New(str)).ParseProgram(), env)
}
