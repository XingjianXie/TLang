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
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	VOID  = &object.Void{}
	JUMP  = &object.Jump{}
)

func PrintParserErrors(out io.Writer, errors []string) {
	_, _ = io.WriteString(out, "Parser Errors:\n")
	for _, msg := range errors {
		_, _ = io.WriteString(out, "\t"+msg+"\n")
	}
}

func init() {
	LEN = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 1 {
			return newError("native function len: len(args) should be 1")
		}
		switch arg := unwrapReferenceValue(args[0]).(type) {
		case *object.String:
			return &object.Integer{Value: int64(len(arg.Value))}
		case *object.Array:
			return &object.Integer{Value: int64(len(arg.Elements))}
		default:
			return newError("native function len: arg should be String or Array")
		}
	}
	PRINT = func(env *object.Environment, args []object.Object) object.Object {
		for _, arg := range args {
			fmt.Print(STRING(env, []object.Object{unwrapReferenceValue(arg)}).(*object.String).Value)
		}
		return VOID
	}
	INPUT = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 0 {
			return newError("native function len: len(args) should be 0")
		}
		var input string
		_, _ = fmt.Scanf("%s", &input)

		return &object.String{Value: input}
	}
	PRINT_LINE = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) == 0 {
			fmt.Println()
			return VOID
		}
		for _, arg := range args {
			fmt.Println(STRING(env, []object.Object{unwrapReferenceValue(arg)}).(*object.String).Value)
		}
		return VOID
	}
	INPUT_LINE = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 0 {
			return newError("native function len: len(args) should be 0")
		}
		var input string
		_, _ = fmt.Scanln(&input)

		return &object.String{Value: input}
	}
	STRING = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 1 {
			return newError("native function string: len(args) should be 1")
		}
		if str, ok := unwrapReferenceValue(args[0]).(*object.String); ok {
			return str
		}
		return &object.String{Value: args[0].Inspect()}
	}
	EXIT = func(env *object.Environment, args []object.Object) object.Object {
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
		return VOID
	}
	EVAL = func(env *object.Environment, args []object.Object) object.Object {
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
	INT = func(env *object.Environment, args []object.Object) object.Object {
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

	FLOAT = func(env *object.Environment, args []object.Object) object.Object {
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

	BOOLEAN = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 1 {
			return newError("native function float: len(args) should be 1")
		}
		return booleanify(unwrapReferenceValue(args[0]))
	}

	FETCH = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 1 {
			return newError("native function fetch: len(args) should be 1")
		}
		if err, ok := args[0].(*object.Err); ok {
			return &object.String{Value: err.Inspect()}
		}
		return args[0]
	}

	APPEND = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 2 {
			return newError("native function append: len(args) should be 1")
		}
		if array, ok := unwrapReferenceValue(args[0]).(*object.Array); ok {
			return &object.Array{Elements: append(array.Elements, unwrapReferenceValue(args[0]))}
		}
		return newError("native function append: args[0] should be Array")
	}

	FIRST = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 1 {
			return newError("native function first: len(args) should be 1")
		}
		if array, ok := args[0].(*object.Array); ok {
			if len(array.Elements) == 0 {
				return VOID
			}
			return &object.Reference{Value: &array.Elements[0]}
		}
		return newError("native function first: arg should be Array")
	}

	LAST = func(env *object.Environment, args []object.Object) object.Object {
		if len(args) != 1 {
			return newError("native function fetch: len(args) should be 1")
		}
		if array, ok := args[0].(*object.Array); ok {
			if len(array.Elements) == 0 {
				return VOID
			}
			return &object.Reference{Value: &array.Elements[len(array.Elements)-1]}
		}
		return newError("native function append: arg should be Array")
	}

	natives = map[string]*object.Native{
		"len":       {LEN},
		"print":     {PRINT},
		"input":     {INPUT},
		"printLine": {PRINT_LINE},
		"inputLine": {INPUT_LINE},
		"string":    {STRING},
		"exit":      {EXIT},
		"eval":      {EVAL},
		"int":       {INT},
		"float":     {FLOAT},
		"boolean":   {BOOLEAN},
		"fetch":     {FETCH},
		"append":    {APPEND},
		"first":     {FIRST},
		"last":      {LAST},
	}
}

var (
	LEN        func(env *object.Environment, args []object.Object) object.Object
	PRINT      func(env *object.Environment, args []object.Object) object.Object
	PRINT_LINE func(env *object.Environment, args []object.Object) object.Object
	INPUT      func(env *object.Environment, args []object.Object) object.Object
	INPUT_LINE func(env *object.Environment, args []object.Object) object.Object
	STRING     func(env *object.Environment, args []object.Object) object.Object
	EXIT       func(env *object.Environment, args []object.Object) object.Object
	EVAL       func(env *object.Environment, args []object.Object) object.Object
	INT        func(env *object.Environment, args []object.Object) object.Object
	FLOAT      func(env *object.Environment, args []object.Object) object.Object
	BOOLEAN    func(env *object.Environment, args []object.Object) object.Object
	FETCH      func(env *object.Environment, args []object.Object) object.Object
	APPEND     func(env *object.Environment, args []object.Object) object.Object
	FIRST      func(env *object.Environment, args []object.Object) object.Object
	LAST       func(env *object.Environment, args []object.Object) object.Object
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

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
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
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env, false)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}

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
			if function == natives["fetch"] {
				return applyFunction(function, args, env)
			}
			return args[0]
		}

		return applyFunction(function, args, env)
	case *ast.IndexExpression:
		ident := unwrapReferenceValue(Eval(node.Left, env))
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
		return JUMP
	case *ast.LetStatement:
		val := unwrapReferenceValue(Eval(node.Value, env))
		if isError(val) {
			return val
		}
		if _, ok := env.Get(node.Name.Value); ok {
			return newError("identifier %s already set", node.Name.Value)
		} else {
			env.Set(node.Name.Value, val.Copy())
		}
	case *ast.RefStatement:
		return evalRefStatement(node, env)
	case *ast.DelStatement:
		if ident, ok := node.DelIdent.(*ast.Identifier); ok {
			if _, ok := env.Get(ident.Value); ok {
				env.Del(ident.Value)
			} else {
				return newError("identifier not found: " + ident.Value)
			}
		} else {
			return newError("left value not a identifier: " + node.DelIdent.String())
		}
	}

	return VOID
}

func applyIndex(ident object.Object, indexes []object.Object) object.Object {
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
		return &object.Reference{Value: refObj}
	}
	return newError("not a array: %s", ident.Type())
}

func applyFunction(fn object.Object, args []object.Object, env *object.Environment) object.Object {
	if function, ok := fn.(*object.Function); ok {
		extendedEnv := extendFunctionEnv(function, args)
		evaluated := Eval(function.Body, extendedEnv)
		return unwrapRetValue(evaluated)
	}

	if native, ok := fn.(*object.Native); ok {
		return native.Fn(env, args)
	}

	return newError("not a function or a native function: %s", fn.Type())
}

func extendFunctionEnv(
	fn *object.Function,
	args []object.Object,
) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		if paramIdx >= len(args) {
			env.Set(param.Value, VOID)
		} else {
			env.Set(param.Value, args[paramIdx])
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
		return val
	}

	if native, ok := natives[node.Value]; ok {
		return native
	}

	return newError("identifier not found: " + node.Value)
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object = VOID

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
	var result object.Object = VOID

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if isError(result) || isSkip(result) {
			return result
		}
	}

	return result
}

func evalRefStatement(node *ast.RefStatement, env *object.Environment) object.Object {
	val := unwrapReferenceValue(Eval(node.Value, env))
	if isError(val) {
		return val
	}
	if _, ok := env.Get(node.Name.Value); ok {
		return newError("identifier %s already set", node.Name.Value)
	} else {
		left := Eval(node.Value, env)
		if refer, ok := left.(*object.Reference); ok {
			env.Set(node.Name.Value, refer)
			return VOID
		} else {
			if ident, ok := node.Value.(*ast.Identifier); ok {
				//DONE REFERENCE WON'T WORK WHEN TARGET IS A IDENTIFIER

				if obj, ok := env.Get(ident.Value); ok {
					env.Set(node.Name.Value, obj)
					return VOID
				}
				//return newError("not support yet, refer to identifier: " + ident.Value)
			}
		}
		return newError("left value not a identifier or a reference: " + node.Value.String())
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
	if refer, ok := left.(*object.Reference); ok {
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
	} else {
		if ident, ok := node.Left.(*ast.Identifier); ok {
			if _, ok := env.Get(ident.Value); ok {
				var newVal object.Object
				switch node.Operator {
				case "+=":
					newVal = evalInfixExpression("+", evalIdentifier(ident, env), val)
				case "-=":
					newVal = evalInfixExpression("-", evalIdentifier(ident, env), val)
				case "*=":
					newVal = evalInfixExpression("*", evalIdentifier(ident, env), val)
				case "/=":
					newVal = evalInfixExpression("/", evalIdentifier(ident, env), val)
				case "%=":
					newVal = evalInfixExpression("%", evalIdentifier(ident, env), val)
				case "=":
					newVal = val
				}
				if isError(newVal) {
					return newVal
				}
				env.Set(ident.Value, newVal)
				return newVal
			}
		}
	}
	return newError("left value not a identifier or a reference: " + node.Left.String())
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

	case left.Type() == object.STRING:
		if right.Type() == object.STRING {
			return evalStringInfixExpression(operator, left.(*object.String), right.(*object.String))
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
	left, right *object.String,
) object.Object {
	switch operator {
	case "==":
		return nativeBoolToBooleanObject(left.Value == right.Value)
	case "!=":
		return nativeBoolToBooleanObject(left.Value != right.Value)
	case "+":
		return &object.String{Value: left.Value + right.Value}
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
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
	return nativeBoolToBooleanObject(booleanify(right) == FALSE)
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
		return VOID
	}
}

func evalLoopExpression(le *ast.LoopExpression, env *object.Environment) object.Object {
	var result object.Object = VOID

	condition := unwrapReferenceValue(Eval(le.Condition, env))
	if isError(condition) {
		return condition
	}

	for isTruthy(condition) {
		newResult := Eval(le.Body, env)
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
	return booleanify(obj) == TRUE
}

func booleanify(number object.Object) object.Object {
	switch number.Type() {
	case object.INTEGER:
		if number.(*object.Integer).Value != 0 {
			return TRUE
		}
		return FALSE
	case object.FLOAT:
		if number.(*object.Float).Value != 0 && !math.IsNaN(number.(*object.Float).Value) {
			return TRUE
		}
		return FALSE
	case object.BOOLEAN:
		return number
	case object.VOID:
		return FALSE
	}
	return newError("could not parse %s as boolean", number.Inspect())
}
