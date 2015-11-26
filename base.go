package exprel

import (
	"math"
	"strings"
)

// Base contains the base functions, as described in the package documentation.
var Base Source

func init() {
	Base = SourceMap{
		// Math
		"EXP": func(values ...interface{}) (interface{}, error) {
			number := argN("EXP", values...)
			return math.Exp(number), nil
		},
		"PI": func(values ...interface{}) (interface{}, error) {
			arg("PI", values...)
			return float64(math.Pi), nil
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
