package exprel

import (
	"math"
	"math/rand"
	"strings"
)

// Base contains the base functions, as described in the package documentation.
var Base Source

func init() {
	Base = SourceMap{
		// Etc.
		"CHOOSE": func(values ...interface{}) (interface{}, error) {
			number, rest := argNRest("CHOOSE", values...)
			index := int(number)
			if index < 0 || index >= len(rest) {
				re("CHOOSE index out of range")
			}
			return rest[index], nil
		},

		// Math
		"ABS": func(values ...interface{}) (interface{}, error) {
			number := argN("ABS", values...)
			return math.Abs(number), nil
		},
		"EXP": func(values ...interface{}) (interface{}, error) {
			number := argN("EXP", values...)
			return math.Exp(number), nil
		},
		"PI": func(values ...interface{}) (interface{}, error) {
			arg("PI", values...)
			return float64(math.Pi), nil
		},
		"RAND": func(values ...interface{}) (interface{}, error) {
			arg("RAND", values...)
			return rand.Float64(), nil
		},

		// Strings
		"LEN": func(values ...interface{}) (interface{}, error) {
			str := argS("LEN", values...)
			return float64(len(str)), nil
		},
		"LOWER": func(values ...interface{}) (interface{}, error) {
			str := argS("LOWER", values...)
			return strings.ToLower(str), nil
		},
		"REPT": func(values ...interface{}) (interface{}, error) {
			str, count := argSN("REPT", values...)
			if count < 0 {
				re("REPT argument must be positive")
			}
			return strings.Repeat(str, int(count)), nil
		},
		"TRIM": func(values ...interface{}) (interface{}, error) {
			str := argS("TRIM", values...)
			return strings.TrimSpace(str), nil
		},
		"UPPER": func(values ...interface{}) (interface{}, error) {
			str := argS("UPPER", values...)
			return strings.ToUpper(str), nil
		},
	}
}

// argument parser helpers

func argS(name string, values ...interface{}) string {
	if len(values) != 1 {
		re(name + " expects a string argument")
	}
	value, ok := values[0].(string)
	if !ok {
		re(name + " expects a string argument")
	}
	return value
}

func argN(name string, values ...interface{}) float64 {
	if len(values) != 1 {
		re(name + " expects a number argument")
	}
	value, ok := values[0].(float64)
	if !ok {
		re(name + " expects a number argument")
	}
	return value
}

func argNRest(name string, values ...interface{}) (float64, []interface{}) {
	if len(values) < 1 {
		re(name + " expects at least one number argument")
	}
	value, ok := values[0].(float64)
	if !ok {
		re(name + " expects at least one number argument")
	}
	return value, values[1:]
}

func argSN(name string, values ...interface{}) (string, float64) {
	if len(values) != 2 {
		re(name + " expects a string and a number argument")
	}
	str, strOk := values[0].(string)
	number, numberOk := values[1].(float64)
	if !strOk || !numberOk {
		re(name + " expects a string and a number argument")
	}
	return str, number
}

func arg(name string, values ...interface{}) {
	if len(values) != 0 {
		re(name + " expects no arguments")
	}
}
