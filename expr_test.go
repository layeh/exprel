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

func TestEvaluate(t *testing.T) {
	data := map[string]interface{}{
		"name": "Tim",
	}
	result, err := exprel.Evaluate(`=LOWER(name) & ".jpg"`, exprel.SourceMap(data))
	if err != nil {
		t.Fatal(err)
	}
	filename := result.(string)
	const expecting = "tim.jpg"
	if filename != expecting {
		t.Fatalf("got %s, expecting %s\n", filename, expecting)
	}
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
		"FUNC": func(*exprel.Call) (interface{}, error) {
			return nil, errors.New("FUNC should not be called")
		},
	}
	expr := `=OR(FALSE(); 5 >= 2; FUNC())`
	testBool(t, expr, true, fns)
}

func TestOr(t *testing.T) {
	fns := exprel.SourceMap{
		"FUNC": func(*exprel.Call) (interface{}, error) {
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

func TestModulo(t *testing.T) {
	expr := `=0 % 5`
	testNumber(t, expr, 0, nil)

	expr = `=5 % 2`
	testNumber(t, expr, 1, nil)

	expr = `=1 % 1`
	testNumber(t, expr, 0, nil)
}

func TestNoArgFunction(t *testing.T) {
	fns := exprel.SourceMap{
		"SAYHELLO": func(*exprel.Call) (interface{}, error) {
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

func TestBaseCHOOSE(t *testing.T) {
	expr := `=CHOOSE(1; 10; 20; 30)`
	testNumber(t, expr, 20, exprel.Base)
}

func TestBaseTYPE(t *testing.T) {
	expr := `=TYPE(123)`
	testNumber(t, expr, 1, exprel.Base)

	expr = `=TYPE("hello")`
	testNumber(t, expr, 2, exprel.Base)

	expr = `=TYPE(TRUE())`
	testNumber(t, expr, 4, exprel.Base)
}

func TestBaseABS(t *testing.T) {
	expr := `=ABS(-342)`
	testNumber(t, expr, 342, exprel.Base)
}

func TestBaseSIGN(t *testing.T) {
	expr := `=SIGN(-34)`
	testNumber(t, expr, -1, exprel.Base)

	expr = `=SIGN(0)`
	testNumber(t, expr, 0, exprel.Base)

	expr = `=SIGN(242342)`
	testNumber(t, expr, 1, exprel.Base)
}

func TestBaseLN(t *testing.T) {
	expr := `=LN(1)`
	testNumber(t, expr, 0, exprel.Base)

	expr = `=LN(EXP(1))`
	testNumber(t, expr, 1, exprel.Base)
}

func TestBaseLOG10(t *testing.T) {
	expr := `=LOG10(10)`
	testNumber(t, expr, 1, exprel.Base)

	expr = `=LOG10(100)`
	testNumber(t, expr, 2, exprel.Base)
}

func TestBaseCHAR(t *testing.T) {
	expr := `=CHAR(72; 101; 108; 108; 111; 33)`
	testString(t, expr, "Hello!", exprel.Base)

	expr = `=CHAR()`
	testString(t, expr, "", exprel.Base)
}

func TestBaseJOIN(t *testing.T) {
	expr := `=JOIN(", "; "a"; "b"; "c")`
	testString(t, expr, "a, b, c", exprel.Base)

	expr = `=JOIN("!!!")`
	testString(t, expr, "", exprel.Base)
}

func TestBaseLEFT(t *testing.T) {
	expr := `=LEFT("hello")`
	testString(t, expr, "h", exprel.Base)

	expr = `=LEFT("hello";10)`
	testString(t, expr, "hello", exprel.Base)
}

func TestBaseLEN(t *testing.T) {
	expr := `=LEN("hélloworld")`
	testNumber(t, expr, float64(len("hélloworld")), exprel.Base)
}

func TestBaseLOWER(t *testing.T) {
	expr := `=LOWER("hey" & "THERE")`
	testString(t, expr, "heythere", exprel.Base)
}

func TestBaseMID(t *testing.T) {
	expr := `=MID("hello world";1;5)`
	testString(t, expr, "hello", exprel.Base)

	expr = `=MID("hello world";-5;3)`
	testString(t, expr, "", exprel.Base)

	expr = `=MID("hello world";20)`
	testString(t, expr, "", exprel.Base)

	expr = `=MID("hello world";7;1)`
	testString(t, expr, "w", exprel.Base)

	expr = `=MID("hello world";7;100)`
	testString(t, expr, "world", exprel.Base)
}

func TestBaseREPT(t *testing.T) {
	expr := `=REPT("1"; 5)`
	testString(t, expr, "11111", exprel.Base)
}

func TestBaseRIGHT(t *testing.T) {
	expr := `=RIGHT("hello")`
	testString(t, expr, "o", exprel.Base)

	expr = `=RIGHT("hello";3)`
	testString(t, expr, "llo", exprel.Base)

	expr = `=RIGHT("hello";10)`
	testString(t, expr, "hello", exprel.Base)
}

func TestBaseSEARCH(t *testing.T) {
	expr := `=SEARCH("e"; "hello world")`
	testNumber(t, expr, 2, exprel.Base)

	expr = `=SEARCH("z"; "hello world")`
	testNumber(t, expr, -1, exprel.Base)

	expr = `=SEARCH("e"; "hello world"; 2)`
	testNumber(t, expr, 2, exprel.Base)

	expr = `=SEARCH("e"; "hello world"; 3)`
	testNumber(t, expr, -1, exprel.Base)

	expr = `=SEARCH("e"; "hello world"; 100)`
	testNumber(t, expr, -1, exprel.Base)
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
	val, err := exprel.String(e.Evaluate(source))
	if err != nil {
		t.Fatalf("could not evaluate expression: %s\n", err)
	}
	if val != expected {
		t.Fatalf("incorrect value (expecting `%s`, got `%s`)\n", expected, val)
	}
}

func testNumber(t *testing.T, expr string, expected float64, source exprel.Source) {
	e, err := exprel.Parse(expr)
	if err != nil {
		t.Fatalf("could not parse expression: %s\n", err)
	}
	val, err := exprel.Number(e.Evaluate(source))
	if err != nil {
		t.Fatalf("could not evaluate expression: %s\n", err)
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
	val, err := exprel.Boolean(e.Evaluate(source))
	if err != nil {
		t.Fatalf("could not evaluate expression: %s\n", err)
	}
	if val != expected {
		t.Fatalf("incorrect value (expecting %v, got %v)\n", expected, val)
	}
}
