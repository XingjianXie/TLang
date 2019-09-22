package evaluator

import (
	"TLang/lexer"
	"TLang/object"
	"TLang/parser"
	"testing"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5;", 5},
		{"10;", 10},
		{"-5;", -5},
		{"-10;", -10},
		{"5 + 5 + 5 + 5 - 10;", 10},
		{"2 * 2 * 2 * 2 * 2;", 32},
		{"-50 + 100 + -50;", 0},
		{"5 * 2 + 10;", 20},
		{"5 + 2 * 10;", 25},
		{"20 + 2 * -10;", 0},
		{"2 * (5 + 10);", 30},
		{"3 * 3 * 3 + 10;", 37},
		{"3 * (3 * 3) + 10;", 37},
		{"5%2;", 1},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	return Eval(program, env)
}

func TestEvalFloatExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"5.0;", 5.0},
		{"10.0;", 10.0},
		{"-5.;", -5.},
		{"-10e2;", -10e2},
		{"5 + 5 + 5 + 5 - 10.0;", 10.0},
		{"2 * 2 * 2 * 2 * 2.0;", 32.0},
		{"-50 + 100 + -50.0;", 0.0},
		{"5 * 2 + 10.0;", 20.0},
		{"5 + 2 * 10.0;", 25.0},
		{"20 + 2. * -10;", 0.},
		{"2 * (5 + 10.);", 30.},
		{"3 * 3 * 3. + 10;", 37.},
		{"3 * (3 * 3) + 10.;", 37.},
		{"1 / 2;", .5},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testFloatObject(t, evaluated, tt.expected)
	}
}

func testFloatObject(t *testing.T, obj object.Object, expected float64) bool {
	result, ok := obj.(*object.Float)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%g, want=%g",
			result.Value, expected)
		return false
	}

	return true
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d",
			result.Value, expected)
		return false
	}

	return true
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true;", true},
		{"false;", false},
		{"1 < 2;", true},
		{"1 > 2;", false},
		{"1 < 1;", false},
		{"1 > 1;", false},
		{"1 == 1;", true},
		{"1 != 1;", false},
		{"1 == 2;", false},
		{"1 != 2;", true},
		{"1>=1;", true},
		{"1<=2;", true},
		{"1.3>=1;", true},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected, i)
	}
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool, i int) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t, at=%d",
			result.Value, expected, i)
		return false
	}
	return true
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true;", false},
		{"!false;", true},
		{"!5;", false},
		{"!!true;", true},
		{"!!false;", false},
		{"!!5;", true},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected, i)
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10; };", 10},
		{"if (false) { 10; };", nil},
		{"if (1) { 10; };", 10},
		{"if (1 < 2) { 10; };", 10},
		{"if (1 > 2) { 10; };", nil},
		{"if (1 > 2) { 10; } else { 20; };", 20},
		{"if (1 < 2) { 10; } else { 20; };", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != Void {
		t.Errorf("object is not Void. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}

func testErrObject(t *testing.T, obj object.Object, message string) bool {
	if obj.Type() != object.ERR {
		t.Errorf("object is not ERR. got=%T (%+v)", obj, obj)
		return false
	}
	if obj.(*object.Err).Message != message {
		t.Errorf("message is not %s. got=%s", message, obj.(*object.Err).Message)
		return false
	}
	return true
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"ret 10;", 10},
		{"ret 10; 9;", 10},
		{"ret 2 * 5; 9;", 10},
		{"9; ret 2 * 5; 9;", 10},
		{
			`
if (10 > 1) {
  if (10 > 1) {
    ret 10;
  };

  ret 1;
};
`,
			10,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true;",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5; true + false; 5;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (10 > 1) { true + false; };",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar;",
			"identifier not found: foobar",
		},
		{
			`
if (10 > 1) {
  if (10 > 1) {
    ret true + false;
  };

  ret 1;
};
`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Err)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v), at=%d",
				evaluated, evaluated, i)
			continue
		}

		if errObj.Message != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q, at=%d",
				tt.expectedMessage, errObj.Message, i)
		}
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestAssignStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a = 3;", 3},
		{"let a = 5 * 5; a = 2;", 2},
		{"let a = 5; let b = a; b = 1;", 1},
		{"let a = 5; let b = a; let c = a + b + 5; c += 1; c += 1;", 17},
		{"let a = 5; a *= 2;", 10},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	input := "func(x) { x + 2; };"

	evaluated := testEval(input)
	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", evaluated, evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Parameters=%+v",
			fn.Parameters)
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}

	expectedBody := "{ (x + 2); }"

	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let identity = func(x) { x; }; identity(5);", 5},
		{"let identity = func(x) { ret x; }; identity(5);", 5},
		{"let double = func(x) { x * 2; }; double(5);", 10},
		{"let add = func(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = func(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"func(x) { x; }(5);", 5},
		{"_ { args[0]; }(5);", 5},
		{"let t = func(x) { x + 1; }; t(t(t(1)));", 4},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestStringLiteral(t *testing.T) {
	input := `"Hello World!";`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestCharacterLiteral(t *testing.T) {
	input := `'1';`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.Character)
	if !ok {
		t.Fatalf("object is not Character. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != '1' {
		t.Errorf("Character has wrong value. got=%q", string(str.Value))
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "World!";`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestConvertFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"int(\"123\");", 123},
		{"int(float(\"123.3\"));", 123},
		{"int(float(\"123.222\"));", 123},
		{"int(float(\"122.9\"));", 122},
		{"int(string(int(\"123\") + 4) + \"2\");", 1272},
		{"int(boolean(1));", 1},
		{"int(boolean(float(\"Nan\")));", 0},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3];"

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is not Array. got=%T (%+v)", evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong num of elements. got=%d",
			len(result.Elements))
	}

	testIntegerObject(t, result.Elements[0], 1)
	testIntegerObject(t, result.Elements[1], 4)
	testIntegerObject(t, result.Elements[2], 6)
}

func TestArrayIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			"[1, 2, 3][0];",
			1,
		},
		{
			"[1, 2, 3][1];",
			2,
		},
		{
			"[1, 2, 3][2];",
			3,
		},
		{
			"let i = 0; [1][i];",
			1,
		},
		{
			"[1, 2, 3][1 + 1];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[2];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];",
			6,
		},
		{
			"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i];",
			2,
		},
		{
			"[1, 2, 3][3];",
			"array: out of range",
		},
		{
			"[1, 2, 3][-1];",
			"array: out of range",
		},
		{
			"let a = [1, 2]; del a; a;",
			"identifier not found: a",
		},
		{
			"let a = [1, 2]; del a[0];",
			"left value not a identifier: (a[0])",
		},
		{
			"let a = [1, 2]; a[0] = 2; a[0];",
			2,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testErrObject(t, evaluated, tt.expected.(string))
		}
	}
}

func TestArrayFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"len([1,2,3]);", 3},
		{"len([]);", 0},
		{"len(append([1,2,3], 4));", 4},
		{"last([1,2]);", 2},
		{"first([1,2]);", 1},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestReference(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 1; let b = a; a = 2; b;", 1},
		{"let a = 0; let b = 0; a = 1; b = a; a = 2; b;", 1},
		{"let b = 0; let a = 1; b = a; a = 2; b;", 1},
		{"let a = 0; a = 1; let b = a; a = 2; b;", 1},

		{"let a = 1; let b = a; b = 2; a;", 1},
		{"let a = 0; let b = 0; a = 1; b = a; b = 2; a;", 1},
		{"let b = 0; let a = 1; b = a; b = 2; a;", 1},
		{"let a = 0; a = 1; let b = a; b = 2; a;", 1},

		{"let a = 1; ref b = a; a = 2; b;", 2},
		{"let a = 0; a = 1; ref b = a; a = 2; b;", 2},

		{"let a = 1; ref b = a; b = 2; a;", 2},
		{"let a = 0; a = 1; ref b = a; b = 2; a;", 2},

		{"let a = [1,2,3,4,[1,2,3]]; let b = a[4]; b[0] = 2; a[4][0];", 1},

		{"let a = [1,2,3]; ref b = a[0]; b = 2; a[0];", 2},
		{"let a = 0; a = [1,2,3,4]; ref b = a[1+2-1]; a[2] = 5; b;", 5},

		{"let a = [1,2,3,4,[1,2,3]]; ref b = a[4]; b[0] = 2; a[4][0];", 2},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}
