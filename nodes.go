package exprel

import (
	"bytes"
	"context"
	"math"
	"strconv"
)

type node interface {
	Evaluate(ctx context.Context, s Source) interface{}
	Encode(b *bytes.Buffer)
}

type stringNode string

func (n stringNode) Evaluate(ctx context.Context, s Source) interface{} {
	return string(n)
}

func (n stringNode) Encode(b *bytes.Buffer) {
	b.WriteString(strconv.Quote(string(n)))
}

type boolNode bool

func (n boolNode) Evaluate(ctx context.Context, s Source) interface{} {
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

func (n numberNode) Evaluate(ctx context.Context, s Source) interface{} {
	return float64(n)
}

func (n numberNode) Encode(b *bytes.Buffer) {
	b.WriteString(strconv.FormatFloat(float64(n), 'f', -1, 64))
}

type notNode struct {
	node
}

func (n *notNode) Evaluate(ctx context.Context, s Source) interface{} {
	val, ok := n.node.Evaluate(ctx, s).(bool)
	if !ok {
		panic(&RuntimeError{Message: "NOT expects bool value"})
	}
	return !val
}

func (n *notNode) Encode(b *bytes.Buffer) {
	b.WriteString("NOT(")
	n.node.Encode(b)
	b.WriteByte(')')
}

type lookupNode string

func (n lookupNode) Evaluate(ctx context.Context, s Source) interface{} {
	id := string(n)
	ret, ok := s.Get(id)
	if !ok {
		panic(&RuntimeError{Message: "unknown identifier " + id})
	}
	switch ret.(type) {
	case string, bool, float64:
	default:
		panic(&RuntimeError{Message: "identifier '" + id + "' has invalid type"})
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

func (n *callNode) Evaluate(ctx context.Context, s Source) interface{} {
	name := n.Name
	fnValue, ok := s.Get(name)
	if !ok {
		panic(&RuntimeError{Message: "unknown function " + name})
	}
	fn, ok := fnValue.(Func)
	if !ok {
		panic(&RuntimeError{Message: "cannot call non-function " + name})
	}
	call := Call{
		Name:   name,
		Values: make([]interface{}, len(n.Args)),

		ctx: ctx,
	}
	for i, arg := range n.Args {
		call.Values[i] = arg.Evaluate(ctx, s)
	}
	ret, err := fn(&call)
	if err != nil {
		panic(&RuntimeError{Err: err})
	}
	switch ret.(type) {
	case string, bool, float64:
		return ret
	default:
		panic(&RuntimeError{Message: "invalid function return type"})
	}
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

func (n concatNode) Evaluate(ctx context.Context, s Source) interface{} {
	lhs, lhsOk := n[0].Evaluate(ctx, s).(string)
	if !lhsOk {
		panic(&RuntimeError{Message: "LHS of & must be string"})
	}
	rhs, rhsOk := n[1].Evaluate(ctx, s).(string)
	if !rhsOk {
		panic(&RuntimeError{Message: "RHS of & must be string"})
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

func (n *mathNode) Evaluate(ctx context.Context, s Source) interface{} {
	lhs, lhsOK := n.LHS.Evaluate(ctx, s).(float64)
	rhs, rhsOK := n.RHS.Evaluate(ctx, s).(float64)
	if !lhsOK || !rhsOK {
		panic(&RuntimeError{Message: "invalid " + string(n.Op) + " operands"})
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
			panic(&RuntimeError{Message: "attempted division by zero"})
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

func (n *eqNode) Evaluate(ctx context.Context, s Source) interface{} {
	lhs := n.LHS.Evaluate(ctx, s)
	rhs := n.RHS.Evaluate(ctx, s)
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
	panic(&RuntimeError{Message: "mismatched comparison operand types"})
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

func (n *cmpNode) Evaluate(ctx context.Context, s Source) interface{} {
	lhs := n.LHS.Evaluate(ctx, s)
	rhs := n.RHS.Evaluate(ctx, s)
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
	panic(&RuntimeError{Message: "mismatched comparison operand types"})
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

func (n andNode) Evaluate(ctx context.Context, s Source) interface{} {
	for _, current := range n {
		value, ok := current.Evaluate(ctx, s).(bool)
		if !ok {
			panic(&RuntimeError{Message: "AND must have boolean arguments"})
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

func (n orNode) Evaluate(ctx context.Context, s Source) interface{} {
	for _, current := range n {
		value, ok := current.Evaluate(ctx, s).(bool)
		if !ok {
			panic(&RuntimeError{Message: "OR must have boolean arguments"})
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

func (n *ifNode) Evaluate(ctx context.Context, s Source) interface{} {
	cond, ok := n.Cond.Evaluate(ctx, s).(bool)
	if !ok {
		panic(&RuntimeError{Message: "IF condition must be boolean"})
	}
	if cond {
		return n.True.Evaluate(ctx, s)
	}
	return n.False.Evaluate(ctx, s)
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
