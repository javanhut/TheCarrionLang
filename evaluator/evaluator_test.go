package evaluator

import (
	"testing"

	"thecarrionlanguage/lexer"
	"thecarrionlanguage/object"
	"thecarrionlanguage/parser"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"1++", 2},
		{"5++", 6},
		{"10++", 11},
		{"1--", 0},
		{"0--", -1},
		{"10--", 9},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
		{"10 % 3", 1},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func testEval(input string) object.Object {
	// fmt.Printf("Evaluating input: %s\n", input)
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()
	// fmt.Printf("Parsed program: %+v\n", program)
	result := Eval(program, env)
	// fmt.Printf("Evaluated result: %v\n", result)
	return result
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, wanted=%d", result.Value, expected)
		return false
	}
	return true
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"True", true},
		{"False", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"10 >= 10", true},
		{"10 <= 9", false},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t",
			result.Value, expected)
		return false
	}
	return true
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!True", false},
		{"!False", true},
		{"!5", false},
		{"!!True", true},
		{"!!False", false},
		{"!!5", true},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if(True): 10", 10},
		{"if(False): 10", nil},
		{"if(1): 10", 10},
		{"if(1 < 2): 10", 10},
		{"if(1 > 10): 10", nil},
		{"if(1 < 2): 10 else: 20", 10},
		{"if(1 > 2): 10 else: 20", 20},
		{`
      if (1<0): 
        return 0
      otherwise (1 > 0):
        return 1 
      else:
        return -1`, 1},
		{`if 10 > 1:
        if 10 > 1:
              return 10
        return 1`, 10},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNoneObject(t, evaluated)
		}
	}
}

func testNoneObject(t *testing.T, obj object.Object) bool {
	if obj != NONE {
		t.Errorf("object is not NONE. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"5 + True",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + True 5",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-True",
			"unknown operator: -BOOLEAN",
		},
		{
			"True + False",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5 True + False 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (10 > 1): True + False ",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
      if (10 > 1):
          if (10 > 1):
              return True + False
      return 1
      `,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{"foobar", "identifier not found: foobar"},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)",
				evaluated, evaluated)
			continue
		}
		if errObj.Message != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q",
				tt.expectedMessage, errObj.Message)
		}
	}
}

func TestAssignmentStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"x = 5 x", 5},
		{"x = 5 * 5 x", 25},
		{"a = 5 b= 5 a b ", 5},
		{"a = 5 b = a c = a + b + 5 c", 15},
	}
	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionDefinitionAndCall(t *testing.T) {
	input := `
spell add(x, y):
    return x + y

result = add(2, 3)
result
`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 5)
}

func TestFunctionDefinitionInline(t *testing.T) {
	// demonstrates a single-line function body
	input := `
spell identity(x): return x
identity(42)
`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 42)
}
