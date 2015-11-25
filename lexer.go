package exprel

import (
	"errors"
	"io"
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

// type:
//  identifier -> Identifier
//  string     -> String
//  float64    -> Number
//  rune       -> Token
func (l *lexer) Next() (interface{}, error) {
	if l.R.Len() == 0 {
		// empty reader
		return nil, io.EOF
	}
	defer l.skipWhitespace()

	r, _, _ := l.R.ReadRune()
	switch {
	case r == tknAdd, r == tknSubtract, r == tknMultiply, r == tknDivide, r == tknPower, r == tknEquals, r == tknConcat, r == tknSep, r == tknOpen, r == tknClose:
		// simple operators
		return r, nil
	case r == tknGreater:
		// greater than, greater than or equal
		peek, _, err := l.R.ReadRune()
		if err != nil || peek != '=' {
			l.R.UnreadRune()
			return r, nil
		}
		return tknGreaterEqual, nil
	case r == tknLess:
		// less than, less than or equal, not equal
		peek, _, err := l.R.ReadRune()
		if err != nil || (peek != tknEquals && peek != tknGreater) {
			l.R.UnreadRune()
			return r, nil
		}
		if peek == tknEquals {
			return tknLessEqual, nil
		}
		return tknInequal, nil
	case unicode.IsDigit(r):
		// number
		l.R.UnreadRune()
		num, err := l.nextNumber()
		if err != nil {
			return nil, err
		}
		return num, nil
	case unicode.IsLetter(r):
		// identifier
		l.R.UnreadRune()
		id, err := l.nextIdentifier()
		if err != nil {
			return nil, err
		}
		return id, nil
	case r == '"':
		// string
		l.R.UnreadRune()
		str, err := l.nextString()
		if err != nil {
			return nil, err
		}
		return str, nil
	default:
		return nil, errors.New("unknown character")
	}
}

func (l *lexer) nextIdentifier() (interface{}, error) {
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
	return identifier(chars), nil
}

func (l *lexer) nextString() (interface{}, error) {
	r, _, _ := l.R.ReadRune()
	chars := []rune{r}
	for {
		r, _, err := l.R.ReadRune()
		if err != nil {
			return nil, err
		}
		if r == '\\' {
			peek, _, err := l.R.ReadRune()
			if err != nil {
				return nil, err
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
	return strconv.Unquote(string(chars))
}

func (l *lexer) nextNumber() (interface{}, error) {
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
	return strconv.ParseFloat(string(chars), 64)
}
