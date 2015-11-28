package exprel

import (
	"math"
)

type node interface {
	Evaluate(s Source) interface{}
}

type stringNode string

func (n stringNode) Evaluate(s Source) interface{} {
	return string(n)
}

type boolNode bool

func (n boolNode) Evaluate(s Source) interface{} {
	return bool(n)
}

type numberNode float64

func (n numberNode) Evaluate(s Source) interface{} {
	return float64(n)
}

type notNode struct {
	node
}

func (n *notNode) Evaluate(s Source) interface{} {
	val, ok := n.node.Evaluate(s).(bool)
	if !ok {
		re("NOT expects bool value")
	}
	return !val
}

type lookupNode string

func (n lookupNode) Evaluate(s Source) interface{} {
	id := string(n)
	ret, ok := s.Get(id)
	if !ok {
		re("unknown identifier " + id)
	}
	switch ret.(type) {
	case string, bool, float64:
	default:
		re("identifier '" + id + "' has invalid type")
	}
	return ret
}

type callNode struct {
	Name string
	Args []node
}

func (n *callNode) Evaluate(s Source) interface{} {
	name := n.Name
	fnValue, ok := s.Get(name)
	if !ok {
		re("unknown function " + name)
	}
	fn, ok := fnValue.(Func)
	if !ok {
		fn2, ok2 := fnValue.(func(*Call) (interface{}, error))
		if !ok2 {
			re("cannot call non-function " + name)
		}
		fn = fn2
	}
	call := Call{
		Name:   name,
		Values: make([]interface{}, len(n.Args)),
	}
	for i, arg := range n.Args {
		call.Values[i] = arg.Evaluate(s)
	}
	ret, err := fn(&call)
	if err != nil {
		panic(&RuntimeError{Err: err})
	}
	switch ret.(type) {
	case string, bool, float64:
		return ret
	default:
		re("invalid function return type")
	}
	panic("never called")
}

type concatNode [2]node

func (n concatNode) Evaluate(s Source) interface{} {
	lhs, lhsOk := n[0].Evaluate(s).(string)
	if !lhsOk {
		re("LHS of & must be string")
	}
	rhs, rhsOk := n[1].Evaluate(s).(string)
	if !rhsOk {
		re("RHS of & must be string")
	}
	return lhs + rhs
}

type mathNode struct {
	Op  rune
	LHS node
	RHS node
}

func (n *mathNode) Evaluate(s Source) interface{} {
	lhs, lhsOK := n.LHS.Evaluate(s).(float64)
	rhs, rhsOK := n.RHS.Evaluate(s).(float64)
	if !lhsOK || !rhsOK {
		re("invalid " + string(n.Op) + " operands")
	}
	switch n.Op {
	case tknAdd:
		return lhs + rhs
	case tknSubtract:
		return lhs - rhs
	case tknMultiply:
		return lhs * rhs
	case tknDivide:
		if rhs == 0 {
			re("attempted division by zero")
		}
		return lhs / rhs
	case tknPower:
		return math.Pow(lhs, rhs)
	default:
		panic("never triggered")
	}
}

type eqNode struct {
	Op  rune
	LHS node
	RHS node
}

func (n *eqNode) Evaluate(s Source) interface{} {
	lhs := n.LHS.Evaluate(s)
	rhs := n.RHS.Evaluate(s)
	{
		a, aOK := lhs.(string)
		b, bOK := rhs.(string)
		if aOK && bOK {
			if n.Op == tknEquals {
				return a == b
			}
			return a != b
		}
	}
	{
		a, aOK := lhs.(bool)
		b, bOK := rhs.(bool)
		if aOK && bOK {
			if n.Op == tknEquals {
				return a == b
			}
			return a != b
		}
	}
	{
		a, aOK := lhs.(float64)
		b, bOK := rhs.(float64)
		if aOK && bOK {
			if n.Op == tknEquals {
				return a == b
			}
			return a != b
		}
	}
	re("mismatched comparison operand types")
	panic("never called")
}

type cmpNode struct {
	Op  rune
	LHS node
	RHS node
}

func (n *cmpNode) Evaluate(s Source) interface{} {
	lhs := n.LHS.Evaluate(s)
	rhs := n.RHS.Evaluate(s)
	{
		a, aOK := lhs.(string)
		b, bOK := rhs.(string)
		if aOK && bOK {
			switch n.Op {
			case tknGreater:
				return a > b
			case tknGreaterEqual:
				return a >= b
			case tknLess:
				return a < b
			case tknLessEqual:
				return a <= b
			}
		}
	}
	{
		a, aOK := lhs.(float64)
		b, bOK := rhs.(float64)
		if aOK && bOK {
			switch n.Op {
			case tknGreater:
				return a > b
			case tknGreaterEqual:
				return a >= b
			case tknLess:
				return a < b
			case tknLessEqual:
				return a <= b
			}
		}
	}
	re("mismatched comparison operand types")
	panic("never called")
}

type andNode []node

func (n andNode) Evaluate(s Source) interface{} {
	for _, current := range n {
		value, ok := current.Evaluate(s).(bool)
		if !ok {
			re("AND must have boolean arguments")
		}
		if !value {
			return false
		}
	}
	return true
}

type orNode []node

func (n orNode) Evaluate(s Source) interface{} {
	for _, current := range n {
		value, ok := current.Evaluate(s).(bool)
		if !ok {
			re("OR must have boolean arguments")
		}
		if value {
			return true
		}
	}
	return false
}

type ifNode struct {
	Cond  node
	True  node
	False node
}

func (n *ifNode) Evaluate(s Source) interface{} {
	cond, ok := n.Cond.Evaluate(s).(bool)
	if !ok {
		re("IF condition must be boolean")
	}
	if cond {
		return n.True.Evaluate(s)
	}
	return n.False.Evaluate(s)
}
