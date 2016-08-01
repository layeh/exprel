package exprel

import (
	"bytes"
	"errors"
)

// Expression is an user-defined expression that can be evaluated.
type Expression struct {
	node node
}

// Parse returned a new, executable expression from s. The syntax of s is
// outlined in the package documentation.
//
// Upon success, expression and nil are returned.  Upon failure, nil and error
// are returned.
func Parse(s string) (*Expression, error) {
	if len(s) == 0 {
		return nil, &SyntaxError{
			Message:  "empty expression",
			Position: 0,
		}
	}
	// simple expression; nothing to parse
	if s[0] != '=' {
		return &Expression{
			node: stringNode(s),
		}, nil
	}

	n, err := parseString(s[1:])
	if err != nil {
		return nil, err
	}

	return &Expression{
		node: n,
	}, nil
}

// Evaluate evaluates the expression with the given source.
//
// Upon success, value and nil are returned. Upon failure, nil and error are
// returned.
func (e *Expression) Evaluate(s Source) (val interface{}, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			if runtimeErr, ok := rec.(*RuntimeError); ok {
				err = runtimeErr
				return
			}
			panic(rec)
		}
	}()
	return e.node.Evaluate(s), nil
}

// MarshalText implements encoding.TextMarshaler.
func (e *Expression) MarshalText() ([]byte, error) {
	if e.node == nil {
		return nil, errors.New("empty expression")
	}

	var b bytes.Buffer
	b.WriteByte('=')
	e.node.Encode(&b)
	return b.Bytes(), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (e *Expression) UnmarshalText(text []byte) error {
	expr, err := Parse(string(text))
	if err != nil {
		return err
	}
	*e = *expr
	return nil
}

// Evaluate parses the given string and evaluates it with the given sources
// (Base is automatically included).
//
// Upon success, value and nil are returned. Upon failure, nil and error are
// returned.
func Evaluate(s string, source ...Source) (val interface{}, err error) {
	expr, err := Parse(s)
	if err != nil {
		return nil, err
	}
	data := make(Sources, len(source)+1)
	data[0] = Base
	copy(data[1:], source)
	result, err := expr.Evaluate(data)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// String ensures that an evaluated expression's return type is string.
func String(val interface{}, err error) (string, error) {
	if err != nil {
		return "", err
	}
	casted, ok := val.(string)
	if !ok {
		return "", errors.New("exprel: invalid return type (string expected, got " + typename(val) + ")")
	}
	return casted, nil
}

// Number ensures that an evaluated expression's return type is float64.
func Number(val interface{}, err error) (float64, error) {
	if err != nil {
		return 0, err
	}
	casted, ok := val.(float64)
	if !ok {
		return 0, errors.New("exprel: invalid return type (number expected, got " + typename(val) + ")")
	}
	return casted, nil
}

// Boolean ensures that an evaluated expression's return type is bool.
func Boolean(val interface{}, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	casted, ok := val.(bool)
	if !ok {
		return false, errors.New("exprel: invalid return type (boolean expected, got " + typename(val) + ")")
	}
	return casted, nil
}

func typename(val interface{}) string {
	switch val.(type) {
	case string:
		return "string"
	case bool:
		return "boolean"
	case float64:
		return "number"
	}
	return ""
}
