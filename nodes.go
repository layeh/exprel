package exprel // import "layeh.com/exprel"

import (
	"bytes"
	"math"
	"strconv"
)

type node interface {
	Evaluate(s Source) interface{}
	Encode(b *bytes.Buffer)
}

type stringNode string

func (n stringNode) Evaluate(s Source) interface{} {
	return string(n)
}

func (n stringNode) Encode(b *bytes.Buffer) {
	b.WriteString(strconv.Quote(string(n)))
}

type boolNode bool

func (n boolNode) Evaluate(s Source) interface{} {
	return bool(n)
}

func (n boolNode) Encode(b *bytes.Buffer) {
	if n {
		b.WriteString("TRUE()")
	} else {
		b.WriteString("FALSE()")
	}
}

type numberNode float64

func (n numberNode) Evaluate(s Source) interface{} {
	return float64(n)
}

func (n numberNode) Encode(b *bytes.Buffer) {
	b.WriteString(strconv.FormatFloat(float64(n), 'f', -1, 64))
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

func (n *notNode) Encode(b *bytes.Buffer) {
	b.WriteString("NOT(")
	n.node.Encode(b)
	b.WriteByte(')')
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

func (n lookupNode) Encode(b *bytes.Buffer) {
	b.WriteString(string(n))
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

func (n *callNode) Encode(b *bytes.Buffer) {
	b.WriteString(n.Name)
	b.WriteByte('(')
	for i, arg := range n.Args {
		if i > 0 {
			b.WriteString("; ")
		}
		arg.Encode(b)
	}
	b.WriteByte(')')
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

func (n concatNode) Encode(b *bytes.Buffer) {
	n[0].Encode(b)
	b.WriteString(" & ")
	n[1].Encode(b)
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
	case tknModulo:
		return math.Mod(lhs, rhs)
	default:
		panic("never triggered")
	}
}

func (n *mathNode) Encode(b *bytes.Buffer) {
	n.LHS.Encode(b)
	b.WriteByte(' ')
	b.WriteRune(n.Op)
	b.WriteByte(' ')
	n.RHS.Encode(b)
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

func (n *eqNode) Encode(b *bytes.Buffer) {
	n.LHS.Encode(b)
	b.WriteByte(' ')
	b.WriteRune(n.Op)
	b.WriteByte(' ')
	n.RHS.Encode(b)
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

func (n *cmpNode) Encode(b *bytes.Buffer) {
	n.LHS.Encode(b)
	b.WriteByte(' ')
	switch n.Op {
	case tknGreaterEqual:
		b.WriteString(">=")
	case tknLessEqual:
		b.WriteString("<=")
	case tknInequal:
		b.WriteString("<>")
	default:
		b.WriteRune(n.Op)
	}
	b.WriteByte(' ')
	n.RHS.Encode(b)
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

func (n andNode) Encode(b *bytes.Buffer) {
	b.WriteString("AND(")
	for i, operand := range n {
		if i > 0 {
			b.WriteString("; ")
		}
		operand.Encode(b)
	}
	b.WriteByte(')')
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

func (n orNode) Encode(b *bytes.Buffer) {
	b.WriteString("OR(")
	for i, operand := range n {
		if i > 0 {
			b.WriteString("; ")
		}
		operand.Encode(b)
	}
	b.WriteByte(')')
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

func (n *ifNode) Encode(b *bytes.Buffer) {
	b.WriteString("IF(")
	n.Cond.Encode(b)
	b.WriteString("; ")
	n.True.Encode(b)
	b.WriteString("; ")
	n.False.Encode(b)
	b.WriteByte(')')
}
