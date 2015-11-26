package exprel_test

import (
	"errors"
	"regexp"
	"testing"

	"github.com/layeh/exprel"
)

func TestSimple(t *testing.T) {
	expr := `Hello World!`
	testString(t, expr, "Hello World!", nil)
}

func TestErrSyntax(t *testing.T) {
	expr := `=5 + $`
	testSyntaxError(t, expr, "expected character", nil)
}

func TestErrRuntime(t *testing.T) {
	expr := `=5 + "hello"`
	testRuntimeError(t, expr, "invalid.*operand", nil)
}

func TestConcat(t *testing.T) {
	expr := `="Hello" & "\n" & "World"`
	testString(t, expr, "Hello\nWorld", nil)
}

func TestSingleNumber(t *testing.T) {
	expr := `=12345`
	testNumber(t, expr, 12345, nil)
}

func TestNumberExpression(t *testing.T) {
	expr := `=5+5*2/0.5`
	testNumber(t, expr, 5+5*2/0.5, nil)
}

func TestSimpleIf(t *testing.T) {
	expr := `=IF(TRUE();FALSE();2)`
	testBool(t, expr, false, nil)
}

func TestIf(t *testing.T) {
	expr := `=IF( 5 >= 3; 3; 5)`
	testNumber(t, expr, 3, nil)
}

func TestAndShort(t *testing.T) {
	fns := exprel.SourceMap{
		"FUNC": func(...interface{}) (interface{}, error) {
			return nil, errors.New("FUNC should not be called")
		},
	}
	expr := `=AND(TRUE(); 5 = 2; FUNC())`
	testBool(t, expr, false, fns)
}

func TestAnd(t *testing.T) {
	expr := `=AND(TRUE(); 5 > 2; "hey" = "hey")`
	testBool(t, expr, true, nil)
}

func TestOrShort(t *testing.T) {
	fns := exprel.SourceMap{
		"FUNC": func(...interface{}) (interface{}, error) {
			return nil, errors.New("FUNC should not be called")
		},
	}
	expr := `=OR(FALSE(); 5 >= 2; FUNC())`
	testBool(t, expr, true, fns)
}

func TestOr(t *testing.T) {
	fns := exprel.SourceMap{
		"FUNC": func(...interface{}) (interface{}, error) {
			return false, nil
		},
	}
	expr := `=OR(FALSE(); 5 = 2; FUNC())`
	testBool(t, expr, false, fns)
}

func TestUnaryMinus(t *testing.T) {
	expr := `=1 - -4`
	testNumber(t, expr, 1 - -4, nil)
}

func TestNoArgFunction(t *testing.T) {
	fns := exprel.SourceMap{
		"SAYHELLO": func(...interface{}) (interface{}, error) {
			return "Hello World!", nil
		},
	}
	expr := `=SAYHELLO()`
	testString(t, expr, "Hello World!", fns)
}

func TestMapSource(t *testing.T) {
	m := map[string]interface{}{
		"A": float64(123),
		"B": float64(456),
	}
	expr := `=A + B`
	testNumber(t, expr, 123+456, exprel.SourceMap(m))
}

func TestBuiltinNOT(t *testing.T) {
	expr := `=NOT(TRUE())`
	testBool(t, expr, false, nil)
}
func TestBaseLEN(t *testing.T) {
	expr := `=LEN("hélloworld")`
	testNumber(t, expr, float64(len("hélloworld")), exprel.Base)
}

func TestBaseLOWER(t *testing.T) {
	expr := `=LOWER("hey" & "THERE")`
	testString(t, expr, "heythere", exprel.Base)
}

func TestBaseREPT(t *testing.T) {
	expr := `=REPT("1"; 5)`
	testString(t, expr, "11111", exprel.Base)
}

func TestBaseTRIM(t *testing.T) {
	expr := `=TRIM(" hello  world   ")`
	testString(t, expr, "hello  world", exprel.Base)
}

func TestBaseUPPER(t *testing.T) {
	expr := `=UPPER("hey" & "THERE")`
	testString(t, expr, "HEYTHERE", exprel.Base)
}

// testing helpers

func testSyntaxError(t *testing.T, expr, messageRegex string, source exprel.Source) {
	_, err := exprel.Parse(expr)
	if err == nil {
		t.Fatalf("expecting parsing error\n")
	}
	syntaxErr, ok := err.(*exprel.SyntaxError)
	if !ok {
		t.Fatalf("expecting syntax error\n")
	}
	matched, err := regexp.MatchString(messageRegex, syntaxErr.Message)
	if err != nil {
		panic(err)
	}
	if !matched {
		t.Fatalf("error message does not match (regex `%s`, got `%s`)", messageRegex, syntaxErr.Message)
	}
}

func testRuntimeError(t *testing.T, expr, messageRegex string, source exprel.Source) {
	e, err := exprel.Parse(expr)
	if err != nil {
		t.Fatalf("could not parse expression: %s\n", err)
	}
	_, err = e.Evaluate(source)
	if err == nil {
		t.Fatalf("expecting runtime error\n")
	}
	runtimeErr, ok := err.(*exprel.RuntimeError)
	if !ok {
		t.Fatalf("expecting runtime error\n")
	}
	matched, err := regexp.MatchString(messageRegex, runtimeErr.Message)
	if err != nil {
		panic(err)
	}
	if !matched {
		t.Fatalf("error message does not match (regex `%s`, got `%s`)", messageRegex, runtimeErr.Message)
	}
}

func testString(t *testing.T, expr, expected string, source exprel.Source) {
	e, err := exprel.Parse(expr)
	if err != nil {
		t.Fatalf("could not parse expression: %s\n", err)
	}
	ret, err := e.Evaluate(source)
	if err != nil {
		t.Fatalf("could not evaluate expression: %s\n", err)
	}
	val, ok := ret.(string)
	if !ok {
		t.Fatalf("expression result should be string\n")
	}
	if val != expected {
		t.Fatalf("incorrect value (expecting %s, got %s)\n", expected, val)
	}
}

func testNumber(t *testing.T, expr string, expected float64, source exprel.Source) {
	e, err := exprel.Parse(expr)
	if err != nil {
		t.Fatalf("could not parse expression: %s\n", err)
	}
	ret, err := e.Evaluate(source)
	if err != nil {
		t.Fatalf("could not evaluate expression: %s\n", err)
	}
	val, ok := ret.(float64)
	if !ok {
		t.Fatalf("expression result should be float64\n")
	}
	if val != expected {
		t.Fatalf("incorrect value (expecting %f, got %f)\n", expected, val)
	}
}

func testBool(t *testing.T, expr string, expected bool, source exprel.Source) {
	e, err := exprel.Parse(expr)
	if err != nil {
		t.Fatalf("could not parse expression: %s\n", err)
	}
	ret, err := e.Evaluate(source)
	if err != nil {
		t.Fatalf("could not evaluate expression: %s\n", err)
	}
	val, ok := ret.(bool)
	if !ok {
		t.Fatalf("expression result should be bool\n")
	}
	if val != expected {
		t.Fatalf("incorrect value (expecting %v, got %v)\n", expected, val)
	}
}
