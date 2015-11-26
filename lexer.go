package exprel

import (
	"strconv"
	"strings"
	"unicode"
)

type identifier string

const (
	tknAdd      rune = '+'
	tknSubtract rune = '-'
	tknMultiply rune = '*'
	tknDivide   rune = '/'
	tknPower    rune = '^'
	tknConcat   rune = '&'

	tknEquals       rune = '='
	tknGreater      rune = '>'
	tknGreaterEqual rune = '≥'
	tknLess         rune = '<'
	tknLessEqual    rune = '≤'
	tknInequal      rune = '≠'

	tknSep rune = ';'

	tknOpen  rune = '('
	tknClose rune = ')'
)

type lexer struct {
	R *strings.Reader
}

func newLexer(s string) *lexer {
	l := &lexer{
		R: strings.NewReader(s),
	}
	l.skipWhitespace()
	return l
}

func (l *lexer) skipWhitespace() {
	for l.R.Len() > 0 {
		r, _, err := l.R.ReadRune()
		if err != nil {
			return
		}
		if !unicode.IsSpace(r) {
			l.R.UnreadRune()
			return
		}
	}
}

func (l *lexer) HasNext() bool {
	return l.R.Len() > 0
}

func (l *lexer) pos() int {
	return int(l.R.Size()) - l.R.Len()
}

// type:
//  identifier -> Identifier
//  string     -> String
//  float64    -> Number
//  rune       -> Token
func (l *lexer) Next() interface{} {
	if l.R.Len() == 0 {
		// empty reader
		panic(&SyntaxError{
			Message:  "unexpected EOF",
			Position: l.pos(),
		})
	}
	defer l.skipWhitespace()

	r, _, _ := l.R.ReadRune()
	switch {
	case r == tknAdd, r == tknSubtract, r == tknMultiply, r == tknDivide, r == tknPower, r == tknEquals, r == tknConcat, r == tknSep, r == tknOpen, r == tknClose:
		// simple operators
		return r
	case r == tknGreater:
		// greater than, greater than or equal
		peek, _, err := l.R.ReadRune()
		if err != nil || peek != '=' {
			l.R.UnreadRune()
			return r
		}
		return tknGreaterEqual
	case r == tknLess:
		// less than, less than or equal, not equal
		peek, _, err := l.R.ReadRune()
		if err != nil || (peek != tknEquals && peek != tknGreater) {
			l.R.UnreadRune()
			return r
		}
		if peek == tknEquals {
			return tknLessEqual
		}
		return tknInequal
	case unicode.IsDigit(r):
		// number
		l.R.UnreadRune()
		return l.nextNumber()
	case unicode.IsLetter(r):
		// identifier
		l.R.UnreadRune()
		return l.nextIdentifier()
	case r == '"':
		// string
		l.R.UnreadRune()
		return l.nextString()
	default:
		panic(&SyntaxError{
			Message:  "unexpected character '" + string(r) + "'",
			Position: l.pos(),
		})
	}
}

func (l *lexer) nextIdentifier() interface{} {
	r, _, _ := l.R.ReadRune()
	chars := []rune{r}
	for {
		r, _, err := l.R.ReadRune()
		if err != nil || (!unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_') {
			l.R.UnreadRune()
			break
		}
		chars = append(chars, r)
	}
	return identifier(chars)
}

func (l *lexer) nextString() interface{} {
	r, _, _ := l.R.ReadRune()
	chars := []rune{r}
	for {
		r, _, err := l.R.ReadRune()
		if err != nil {
			panic(&SyntaxError{
				Message:  "unexpected EOF",
				Position: l.pos(),
			})
		}
		if r == '\\' {
			peek, _, err := l.R.ReadRune()
			if err != nil {
				panic(&SyntaxError{
					Message:  "unexpected EOF",
					Position: l.pos(),
				})
			}
			if peek == '"' {
				chars = append(chars, r, peek)
				continue
			}
			l.R.UnreadRune()
		}
		chars = append(chars, r)
		if r == '"' {
			break
		}
	}
	str, err := strconv.Unquote(string(chars))
	if err != nil {
		panic(&SyntaxError{
			Message:  err.Error(),
			Position: l.pos(),
		})
	}
	return str
}

func (l *lexer) nextNumber() interface{} {
	r, _, _ := l.R.ReadRune()
	chars := []rune{r}
	hasDecimal := false
	for {
		r, _, err := l.R.ReadRune()
		if err != nil {
			l.R.UnreadRune()
			break
		}
		if unicode.IsDigit(r) || r == '.' && !hasDecimal {
			chars = append(chars, r)
		} else {
			l.R.UnreadRune()
			break
		}
	}
	number, err := strconv.ParseFloat(string(chars), 64)
	if err != nil {
		panic(&SyntaxError{
			Message:  err.Error(),
			Position: l.pos(),
		})
	}
	return number
}
