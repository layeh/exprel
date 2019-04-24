package exprel

const maximumDepth = 1024

type parser struct {
	Node node

	depth   int
	l       *lexer
	lastTkn interface{}
}

func parseString(s string) (n node, err error) {
	p := &parser{
		l: newLexer(s),
	}
	defer func() {
		if rec := recover(); rec != nil {
			if syntaxErr, ok := rec.(*SyntaxError); ok {
				err = syntaxErr
				return
			}
			panic(rec)
		}
	}()
	n = p.do(p.parseProgram)
	return
}

func (p *parser) do(fn func() node) node {
	p.depth++
	if p.depth >= maximumDepth {
		panic(&SyntaxError{
			Message:  "maximum depth reached",
			Position: p.l.pos(),
		})
	}

	n := fn()

	p.depth--

	return n
}

func (p *parser) next() interface{} {
	if p.lastTkn != nil {
		tkn := p.lastTkn
		p.lastTkn = nil
		return tkn
	}
	return p.l.Next()
}

func (p *parser) nextRune(expecting rune) {
	r, ok := p.next().(rune)
	if !ok || r != expecting {
		panic(&SyntaxError{
			Message:  "expecting '" + string(expecting) + "'",
			Position: p.l.pos(),
		})
	}
}

func (p *parser) peek() interface{} {
	if p.lastTkn != nil {
		return p.lastTkn
	}
	defer func() {
		recover()
	}()
	p.lastTkn = p.l.Next()
	return p.lastTkn
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
	expr := p.do(p.parseExpression)
	if p.l.HasNext() {
		panic(&SyntaxError{
			Message:  "expecting EOF",
			Position: p.l.pos(),
		})
	}
	return expr
}

/*
 * EXPRESSION: BIN1 ["=" | ">" | ">=" | "<" | "<=" | "<>" ] EXPRESSION
 *             BIN1
 */
func (p *parser) parseExpression() node {
	lhs := p.do(p.parseBin1)
	if r, ok := p.peek().(rune); ok {
		switch r {
		case tknEquals, tknInequal:
			p.next()
			rhs := p.do(p.parseExpression)
			return &eqNode{r, lhs, rhs}
		case tknGreater, tknGreaterEqual, tknLess, tknLessEqual:
			p.next()
			rhs := p.do(p.parseExpression)
			return &cmpNode{r, lhs, rhs}
		}
	}
	return lhs
}

/*
 * BIN1        BIN2
 *             BIN2 ["+" | "-" | "&" ] BIN1
 */
func (p *parser) parseBin1() node {
	lhs := p.do(p.parseBin2)
	if r, ok := p.peek().(rune); ok {
		switch r {
		case tknAdd, tknSubtract:
			p.next()
			rhs := p.do(p.parseBin1)
			return &mathNode{r, lhs, rhs}
		case tknConcat:
			p.next()
			rhs := p.do(p.parseBin1)
			return concatNode{lhs, rhs}
		}
	}
	return lhs
}

/*
 * BIN2        TERM
 *             TERM ["*" | "/" | "^" | "%" ] BIN2
 */
func (p *parser) parseBin2() node {
	lhs := p.do(p.parseTerm)
	if r, ok := p.peek().(rune); ok {
		switch r {
		case tknMultiply, tknDivide, tknPower, tknModulo:
			p.next()
			rhs := p.do(p.parseBin2)
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
			expr := p.do(p.parseExpression)
			p.nextRune(tknClose)
			return expr
		case tknSubtract:
			num, ok := p.next().(float64)
			if !ok {
				panic(&SyntaxError{
					Message:  "expecting number",
					Position: p.l.pos(),
				})
			}
			return numberNode(-num)
		default:
			panic(&SyntaxError{
				Message:  "unexpected '" + string(v) + "'",
				Position: p.l.pos(),
			})
		}
	case identifier:
		if p.peekRune(tknOpen) {
			switch string(v) {
			case "IF":
				p.next()
				ifCond := p.do(p.parseExpression)
				p.nextRune(tknSep)
				ifTrue := p.do(p.parseExpression)
				p.nextRune(tknSep)
				ifFalse := p.do(p.parseExpression)
				p.nextRune(tknClose)
				return &ifNode{ifCond, ifTrue, ifFalse}
			case "TRUE":
				p.next()
				p.nextRune(tknClose)
				return boolNode(true)
			case "FALSE":
				p.next()
				p.nextRune(tknClose)
				return boolNode(false)
			case "NOT":
				p.next()
				expr := p.do(p.parseExpression)
				p.nextRune(tknClose)
				return &notNode{expr}
			case "AND":
				p.next()
				var n andNode
				for {
					expr := p.do(p.parseExpression)
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
					expr := p.do(p.parseExpression)
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
						expr := p.do(p.parseExpression)
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
		panic(&SyntaxError{
			Message:  "expecting token",
			Position: p.l.pos(),
		})
	}
}
