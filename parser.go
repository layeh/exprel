package exprel

import (
	"errors"
)

type parser struct {
	Node node

	l       *lexer
	lastTkn interface{}
}

func parseString(s string) (n node, err error) {
	p := &parser{
		l: newLexer(s),
	}
	// TODO: fix recover
	defer func() {
		if rec := recover(); rec != nil {
			err = errors.New(rec.(string))
		}
	}()
	n = p.parseProgram()
	return
}

func (p *parser) next() interface{} {
	if p.lastTkn != nil {
		tkn := p.lastTkn
		p.lastTkn = nil
		return tkn
	}

	tkn, err := p.l.Next()
	if err != nil {
		panic(err)
	}
	return tkn
}

func (p *parser) nextRune(expecting rune) {
	r, ok := p.next().(rune)
	if !ok || r != expecting {
		panic("expecting " + string(expecting))
	}
}

func (p *parser) peek() interface{} {
	if p.lastTkn != nil {
		return p.lastTkn
	}
	tkn, _ := p.l.Next()
	p.lastTkn = tkn
	return tkn
}

func (p *parser) peekRune(expecting rune) bool {
	if r, ok := p.peek().(rune); ok && r == expecting {
		return true
	}
	return false
}

/*
 * PROGRAM     EXPRESSION
 */
func (p *parser) parseProgram() node {
	expr := p.parseExpression()
	if p.l.HasNext() {
		panic("expecting EOF")
	}
	return expr
}

/*
 * EXPRESSION: BIN1 ["=" | ">" | ">=" | "<" | "<=" | "<>" ] EXPRESSION
 *             BIN1
 */
func (p *parser) parseExpression() node {
	lhs := p.parseBin1()
	if r, ok := p.peek().(rune); ok {
		switch r {
		case tknEquals, tknInequal:
			p.next()
			rhs := p.parseExpression()
			return &eqNode{r, lhs, rhs}
		case tknGreater, tknGreaterEqual, tknLess, tknLessEqual:
			p.next()
			rhs := p.parseExpression()
			return &cmpNode{r, lhs, rhs}
		}
	}
	return lhs
}

/*
 * BIN1        BIN2
 *             BIN2 ["+" | "-" | "&" ] EXPRESSION
 */
func (p *parser) parseBin1() node {
	lhs := p.parseBin2()
	if r, ok := p.peek().(rune); ok {
		switch r {
		case tknAdd, tknSubtract:
			p.next()
			rhs := p.parseExpression()
			return &mathNode{r, lhs, rhs}
		case tknConcat:
			p.next()
			rhs := p.parseExpression()
			return concatNode{lhs, rhs}
		}
	}
	return lhs
}

/*
 * BIN2        TERM
 *             TERM ["*" | "/" | "^" ] EXPRESSION
 */
func (p *parser) parseBin2() node {
	lhs := p.parseTerm()
	if r, ok := p.peek().(rune); ok {
		switch r {
		case tknMultiply, tknDivide, tknPower:
			p.next()
			rhs := p.parseExpression()
			return &mathNode{r, lhs, rhs}
		}
	}
	return lhs
}

/*
 * TERM        "(" EXPRESSION ")"
 *             "IF" "(" EXPRESSION ";" EXPRESSION ";" EXPRESSION" ")"
 *             STRING
 *             NUMBER
 *             "TRUE" "(" ")"
 *             "FALSE" "(" ")"
 *             "AND" "(" EXPRESSION ( ";" EXPRESSION )* ")"
 *             "OR" "(" EXPRESSION ( ";" EXPRESSION )* ")"
 *             "NOT" "(" EXPRESSION ")"
 *             "-" NUMBER
 *             IDENTIFIER "(" (EXPRESSION ( ";" EXPRESSION )*)? ")"
 *             IDENTIFIER
 */
func (p *parser) parseTerm() node {
	tkn := p.next()

	switch v := tkn.(type) {
	case rune:
		switch v {
		case tknOpen:
			expr := p.parseExpression()
			p.nextRune(tknClose)
			return expr
		case tknSubtract:
			num, ok := p.next().(float64)
			if !ok {
				panic("expecting number")
			}
			return numberNode(-num)
		default:
			panic("unexpected " + string(v))
		}
	case identifier:
		if p.peekRune(tknOpen) {
			switch string(v) {
			case "IF":
				p.next()
				ifCond := p.parseExpression()
				p.nextRune(tknSep)
				ifTrue := p.parseExpression()
				p.nextRune(tknSep)
				ifFalse := p.parseExpression()
				p.nextRune(tknClose)
				return &ifNode{ifCond, ifTrue, ifFalse}
			case "TRUE":
				p.next()
				p.nextRune(tknClose)
				return trueNode{}
			case "FALSE":
				p.next()
				p.nextRune(tknClose)
				return falseNode{}
			case "NOT":
				p.next()
				expr := p.parseExpression()
				p.nextRune(tknClose)
				return &notNode{expr}
			case "AND":
				p.next()
				var n andNode
				for {
					expr := p.parseExpression()
					n = append(n, expr)
					if !p.peekRune(tknSep) {
						break
					}
					p.nextRune(tknSep)
				}
				p.nextRune(tknClose)
				return n
			case "OR":
				p.next()
				var n orNode
				for {
					expr := p.parseExpression()
					n = append(n, expr)
					if !p.peekRune(tknSep) {
						break
					}
					p.nextRune(tknSep)
				}
				p.nextRune(tknClose)
				return n
			default:
				p.next()
				call := &callNode{
					Name: string(v),
				}
				if !p.peekRune(tknClose) {
					for {
						expr := p.parseExpression()
						call.Args = append(call.Args, expr)
						if !p.peekRune(tknSep) {
							break
						}
						p.next()
					}
				}
				p.nextRune(tknClose)
				return call
			}
		}
		return lookupNode(string(v))
	case bool:
		return boolNode(v)
	case string:
		return stringNode(v)
	case float64:
		return numberNode(v)
	default:
		panic("unexpected token")
	}
}
