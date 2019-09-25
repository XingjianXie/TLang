package evaluator

import (
	"TLang/ast"
	"TLang/lexer"
	"TLang/object"
	"TLang/parser"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
)

var (
	True  object.Object = &object.Boolean{Value: true}
	False object.Object = &object.Boolean{Value: false}
	Void  object.Object = &object.Void{}
	Jump  object.Object = &object.Jump{}
)

func PrintParserErrors(out io.Writer, errors []string) {
	_, _ = io.WriteString(out, "PARSER ERRORS:\n")
	for _, msg := range errors {
		_, _ = io.WriteString(out, "    "+msg+"\n")
	}
}

func init() {
	NativeLen = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 1 {
			return newError("native function len: len(args) should be 1")
		}
		switch arg := unwrapReferenceValue(args[0]).(type) {
		case *object.String:
			return &object.Integer{Value: int64(len([]rune(arg.Value)))}
		case *object.Array:
			return &object.Integer{Value: int64(len(arg.Elements))}
		default:
			return newError("native function len: arg should be String or Array")
		}
	}
	NativePrint = func(env *object.Environment, args []object.Object) object.Object {
		for _, arg := range args {
			fmt.Print(NativeString(env, []object.Object{unwrapReferenceValue(arg)}).(*object.String).Value)
		}
		return Void
	}
	NativeInput = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 0 {
			return newError("native function len: len(args) should be 0")
		}
		var input string
		_, _ = fmt.Scanf("%s", &input)

		return &object.String{Value: input}
	}
	NativePrintLine = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) == 0 {
			fmt.Println()
			return Void
		}
		for _, arg := range args {
			fmt.Println(NativeString(env, []object.Object{unwrapReferenceValue(arg)}).(*object.String).Value)
		}
		return Void
	}
	NativeInputLine = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 0 {
			return newError("native function len: len(args) should be 0")
		}
		var input string
		_, _ = fmt.Scanln(&input)

		return &object.String{Value: input}
	}
	NativeString = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 1 {
			return newError("native function string: len(args) should be 1")
		}
		un := unwrapReferenceValue(args[0])
		if str, ok := un.(*object.String); ok {
			return str
		}
		if ch, ok := un.(*object.Character); ok {
			return &object.String{Value: string(ch.Value)}
		}
		return &object.String{Value: un.Inspect()}
	}
	NativeExit = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 1 && len(args) != 0 {
			return newError("native function exit: len(args) should be 1 or 0")
		}

		if len(args) == 1 {
			if val, ok := unwrapReferenceValue(args[0]).(*object.Integer); ok {
				os.Exit(int(val.Value))
			}
			return newError("native function len: arg should be Integer")
		}
		os.Exit(0)
		return Void
	}
	NativeEval = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 1 {
			return newError("native function eval: len(args) should be 1")
		}

		if str, ok := unwrapReferenceValue(args[0]).(*object.String); ok {
			l := lexer.New(str.Value)
			p := parser.New(l)

			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				PrintParserErrors(os.Stdout, p.Errors())
				return newError("error inner eval")
			}

			return Eval(program, env)
		}

		return newError("native function eval: args should be String")
	}
	NativeInt = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 1 {
			return newError("native function int: len(args) should be 1")
		}
		switch arg := unwrapReferenceValue(args[0]).(type) {
		case *object.String:
			val, err := strconv.ParseInt(arg.Value, 10, 64)
			if err != nil {
				return newError("could not parse %s as integer", arg.Value)
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
			return newError("native function int: arg should be String, Boolean, Number or Void")
		}
	}

	NativeFloat = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 1 {
			return newError("native function float: len(args) should be 1")
		}
		switch arg := unwrapReferenceValue(args[0]).(type) {
		case *object.String:
			val, err := strconv.ParseFloat(arg.Value, 64)
			if err != nil {
				return newError("could not parse %s as float", arg.Value)
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
			return newError("native function int: arg should be String, Boolean, Number or Void")
		}
	}

	NativeBoolean = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 1 {
			return newError("native function float: len(args) should be 1")
		}
		return toBoolean(unwrapReferenceValue(args[0]))
	}

	NativeFetch = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 1 {
			return newError("native function fetch: len(args) should be 1")
		}
		if err, ok := args[0].(*object.Err); ok {
			return &object.String{Value: err.Inspect()}
		}
		return args[0]
	}

	NativeAppend = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 2 {
			return newError("native function append: len(args) should be 2")
		}
		if array, ok := unwrapReferenceValue(args[0]).(*object.Array); ok {
			return &object.Array{Elements: append(array.Elements, unwrapReferenceValue(args[1])), Copyable: true}
		}
		return newError("native function append: args[0] should be Array")
	}

	NativeFirst = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 1 {
			return newError("native function first: len(args) should be 1")
		}
		if array, ok := unwrapReferenceValue(args[0]).(*object.Array); ok {
			if len(array.Elements) == 0 {
				return Void
			}
			return &object.Reference{Value: &array.Elements[0], Const: false}
		}
		return newError("native function first: arg should be Array")
	}

	NativeLast = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 1 {
			return newError("native function fetch: len(args) should be 1")
		}
		if array, ok := unwrapReferenceValue(args[0]).(*object.Array); ok {
			if len(array.Elements) == 0 {
				return Void
			}
			return &object.Reference{Value: &array.Elements[len(array.Elements)-1], Const: false}
		}
		return newError("native function append: arg should be Array")
	}

	NativeType = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 1 {
			return newError("native function type: len(args) should be 1")
		}
		if refer, ok := args[0].(*object.Reference); ok {
			isConst := ""
			if refer.Const {
				isConst = "CONST "
			}
			rawType := "[NOT ALLOC]"
			if refer.Value != nil {
				rawType = string((*refer.Value).Type())
			}
			return &object.String{Value: isConst + "REFERENCE: " + rawType}
		}
		return &object.String{Value: string(args[0].Type())}
	}

	NativeArray = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) == 1 {
			if length, ok := unwrapReferenceValue(args[0]).(*object.Integer); ok {
				var elem []object.Object
				for i := int64(0); i < length.Value; i++ {
					elem = append(elem, Void)
				}

				return &object.Array{
					Elements: elem,
					Copyable: true,
				}
			}
			return newError("native function array: args[0] should be Integer")
		} else if len(args) == 2 {
			if length, ok := unwrapReferenceValue(args[0]).(*object.Integer); ok {
				var elem []object.Object
				for i := int64(0); i < length.Value; i++ {
					elem = append(elem, unwrapReferenceValue(args[1]))
				}

				return &object.Array{
					Elements: elem,
					Copyable: true,
				}
			}
			return newError("native function array: args[0] should be Integer")
		} else if len(args) == 3 {
			if length, ok := unwrapReferenceValue(args[0]).(*object.Integer); ok {
				if function, ok := unwrapReferenceValue(args[2]).(object.LikeFunction); ok {
					var elem []object.Object
					e := unwrapReferenceValue(args[1])
					for i := int64(0); i < length.Value; i++ {
						e = unwrapReferenceValue(applyFunction(function, []object.Object{&object.Integer{Value: i}, e}, env))
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
				return newError("native function array: args[2] should be Function")
			}
			return newError("native function array: args[0] should be Integer")
		}
		return newError("native function array: len(args) should be 1, 2 or 3")
	}

	NativeValue = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 1 {
			return newError("native function array: len(args) should be 1")
		}
		return unwrapReferenceValue(args[0])
	}

	NativeEcho = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 1 {
			return newError("native function array: len(args) should be 1")
		}
		return args[0]
	}

	natives = map[string]*object.Native{
		"len":       {NativeLen},
		"print":     {NativePrint},
		"input":     {NativeInput},
		"printLine": {NativePrintLine},
		"inputLine": {NativeInputLine},
		"string":    {NativeString},
		"exit":      {NativeExit},
		"eval":      {NativeEval},
		"int":       {NativeInt},
		"float":     {NativeFloat},
		"boolean":   {NativeBoolean},
		"fetch":     {NativeFetch},
		"append":    {NativeAppend},
		"first":     {NativeFirst},
		"last":      {NativeLast},
		"type":      {NativeType},
		"array":     {NativeArray},
		"value":     {NativeValue},
		"echo":      {NativeEcho},
	}
}

var (
	NativeLen       func(env *object.Environment, args []object.Object) object.Object
	NativePrint     func(env *object.Environment, args []object.Object) object.Object
	NativePrintLine func(env *object.Environment, args []object.Object) object.Object
	NativeInput     func(env *object.Environment, args []object.Object) object.Object
	NativeInputLine func(env *object.Environment, args []object.Object) object.Object
	NativeString    func(env *object.Environment, args []object.Object) object.Object
	NativeExit      func(env *object.Environment, args []object.Object) object.Object
	NativeEval      func(env *object.Environment, args []object.Object) object.Object
	NativeInt       func(env *object.Environment, args []object.Object) object.Object
	NativeFloat     func(env *object.Environment, args []object.Object) object.Object
	NativeBoolean   func(env *object.Environment, args []object.Object) object.Object
	NativeFetch     func(env *object.Environment, args []object.Object) object.Object
	NativeAppend    func(env *object.Environment, args []object.Object) object.Object
	NativeFirst     func(env *object.Environment, args []object.Object) object.Object
	NativeLast      func(env *object.Environment, args []object.Object) object.Object
	NativeType      func(env *object.Environment, args []object.Object) object.Object
	NativeArray     func(env *object.Environment, args []object.Object) object.Object
	NativeValue     func(env *object.Environment, args []object.Object) object.Object
	NativeEcho      func(env *object.Environment, args []object.Object) object.Object
)

var natives map[string]*object.Native

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
		return True
	}
	return False
}

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return unwrapReferenceValue(evalProgram(node, env))

	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
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
		right := unwrapReferenceValue(Eval(node.Right, env))
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := unwrapReferenceValue(Eval(node.Left, env))
		if isError(left) {
			return left
		}

		right := unwrapReferenceValue(Eval(node.Right, env))
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)

	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.LoopExpression:
		return evalLoopExpression(node, env)
	case *ast.AssignExpression:
		return evalAssignExpression(node, env)
	case *ast.CallExpression:
		function := unwrapReferenceValue(Eval(node.Function, env))
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env, false)
		if len(args) == 1 && isError(args[0]) {
			if function != natives["fetch"] {
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

		return applyIndex(ident, indexes)

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
		return Jump
	case *ast.LetStatement:
		val := Void
		if node.Value != nil {
			val = unwrapReferenceValue(Eval(node.Value, env))
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
				if !env.DeAlloc(&object.String{Value: ident.Value}) {
					return newError("unable to dealloc: %s", node.DelIdent.String())
				}
			} else {
				return newError("identifier not found: %s", ident.Value)
			}
		} else {
			if refer, ok := Eval(node.DelIdent, env).(*object.Reference); ok {
				if refer.Const {
					return newError("delete a constant reference: %s", node.DelIdent.String())
				}
				if refer.Origin != nil {
					if !refer.Origin.DeAlloc(refer.Index) {
						return newError("unable to dealloc: %s", node.DelIdent.String())
					}
					return Void
				}
			}
			return newError("left value not Identifier or AllocRequired: %s", node.DelIdent.String())
		}
	}

	return Void
}

func applyIndex(ident object.Object, indexes []object.Object) object.Object {
	constObj := true
	if refer, ok := ident.(*object.Reference); ok {
		constObj = refer.Const
		ident = unwrapReferenceValue(ident)
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
		runeStr := []rune(str.Value)
		if len(indexes) != 1 {
			return newError("string: len(indexes) should be 1")
		}
		if indexes[0].Type() != object.INTEGER {
			return newError("string: index should be Integer")
		}
		index := indexes[0].(*object.Integer).Value
		length := int64(len(runeStr))
		if index >= length || index < 0 {
			return newError("string: out of range")
		}
		return &object.Character{Value: runeStr[index]}
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
		if !ok {
			return &object.Reference{Value: nil, Const: constObj, Origin: hash, Index: key}
			//return Void
		}

		return &object.Reference{Value: pair.Value, Const: constObj, Origin: hash, Index: key}
	}
	return newError("not Array, String or Hash: %s", ident.Type())
}

func applyFunction(fn object.Object, args []object.Object, env *object.Environment) object.Object {
	if function, ok := fn.(*object.Function); ok {
		extendedEnv := extendFunctionEnv(function, args)
		evaluated := Eval(function.Body, extendedEnv)
		return unwrapRetValue(evaluated)
	}

	if function, ok := fn.(*object.UnderLine); ok {
		inner := env.NewEnclosedEnvironment()
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
		return unwrapRetValue(evaluated)
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

	for paramIdx, param := range fn.Parameters {
		if paramIdx >= len(args) {
			env.SetCurrent(param.Value, &object.Reference{Value: &Void, Const: true})
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

func unwrapRetValue(obj object.Object) object.Object {
	if retValue, ok := obj.(*object.RetValue); ok {
		return retValue.Value
	}

	return obj
}

func unwrapOutValue(obj object.Object) object.Object {
	if retValue, ok := obj.(*object.OutValue); ok {
		return retValue.Value
	}

	return obj
}

func unwrapReferenceValue(obj object.Object) object.Object {
	if referenceVal, ok := obj.(*object.Reference); ok {
		if referenceVal.Value == nil {
			return Void
		}
		return *referenceVal.Value
	}

	return obj
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
			evaluated = unwrapReferenceValue(Eval(e, env))
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
			Index:  &object.String{Value: node.Value},
			Const:  false,
		}
	}

	if native, ok := natives[node.Value]; ok {
		return native
	}

	return newError("identifier not found: " + node.Value)
}

func evalHashLiteral(
	node *ast.HashLiteral,
	env *object.Environment,
) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := unwrapReferenceValue(Eval(keyNode, env))
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
	result := Void

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
	result := Void

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
		return Void
	} else {
		if _, ok := env.SetCurrent(node.Name.Value, &object.Reference{Value: &left, Const: true}); !ok {
			return newError("identifier %s already set", left.Inspect())
		}
		return Void
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
	val := unwrapReferenceValue(Eval(node.Value, env))
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
		return &object.String{Value: left + right}
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
	return nativeBoolToBooleanObject(toBoolean(right) == False)
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
	condition := unwrapReferenceValue(Eval(ie.Condition, env))
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return Void
	}
}

func evalLoopExpression(le *ast.LoopExpression, env *object.Environment) object.Object {
	result := Void

	condition := unwrapReferenceValue(Eval(le.Condition, env))
	if isError(condition) {
		return condition
	}

	for isTruthy(condition) {
		newResult := Eval(le.Body, env.NewEnclosedEnvironment())
		if isError(newResult) || newResult.Type() == object.RET {
			return newResult
		}

		if newResult.Type() == object.OUT {
			return unwrapOutValue(newResult)
		}

		if newResult.Type() != object.JUMP {
			result = newResult
		}

		condition = unwrapReferenceValue(Eval(le.Condition, env))
		if isError(condition) {
			return condition
		}
	}
	return result
}

func isTruthy(obj object.Object) bool {
	return toBoolean(obj) == True
}

func toBoolean(number object.Object) object.Object {
	switch number.Type() {
	case object.INTEGER:
		if number.(*object.Integer).Value != 0 {
			return True
		}
		return False
	case object.FLOAT:
		if number.(*object.Float).Value != 0 && !math.IsNaN(number.(*object.Float).Value) {
			return True
		}
		return False
	case object.BOOLEAN:
		return number
	case object.VOID:
		return False
	}
	return newError("could not parse %s as boolean", number.Inspect())
}
