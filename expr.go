package exprel

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
