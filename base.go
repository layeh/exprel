package exprel

import (
	"bytes"
	"math"
	"math/rand"
	"strings"
)

// Base contains the base functions, as described in the package documentation.
var Base Source

func init() {
	Base = SourceMap{
		// Etc.
		"CHOOSE": func(c *Call) (interface{}, error) {
			number := c.Number(0)
			rest := c.Values[1:]
			index := int(number)
			if index < 0 || index >= len(rest) {
				re("CHOOSE index out of range")
			}
			return rest[index], nil
		},

		// Math
		"ABS": func(c *Call) (interface{}, error) {
			number := c.Number(0)
			return math.Abs(number), nil
		},
		"EXP": func(c *Call) (interface{}, error) {
			number := c.Number(0)
			return math.Exp(number), nil
		},
		"PI": func(c *Call) (interface{}, error) {
			return float64(math.Pi), nil
		},
		"RAND": func(c *Call) (interface{}, error) {
			return rand.Float64(), nil
		},

		// Strings
		"CHAR": func(c *Call) (interface{}, error) {
			var r []rune
			for _, v := range c.Values {
				code, ok := v.(float64)
				if !ok {
					re("CHAR argument must be float64")
				}
				r = append(r, rune(code))
			}
			return string(r), nil
		},
		"JOIN": func(c *Call) (interface{}, error) {
			sep := c.String(0)
			var buff bytes.Buffer
			for i, v := range c.Values[1:] {
				str, ok := v.(string)
				if !ok {
					re("JOIN arguments must be string")
				}
				if i > 0 {
					buff.WriteString(sep)
				}
				buff.WriteString(str)
			}
			return buff.String(), nil
		},
		"LEFT": func(c *Call) (interface{}, error) {
			str := c.String(0)
			count := int(c.OptNumber(1, 1))
			if count > len(str) {
				return str, nil
			}
			return str[:count], nil
		},
		"LEN": func(c *Call) (interface{}, error) {
			str := c.String(0)
			return float64(len(str)), nil
		},
		"LOWER": func(c *Call) (interface{}, error) {
			str := c.String(0)
			return strings.ToLower(str), nil
		},
		"REPT": func(c *Call) (interface{}, error) {
			str := c.String(0)
			count := c.Number(1)
			if count < 0 {
				re("REPT argument must be positive")
			}
			return strings.Repeat(str, int(count)), nil
		},
		"RIGHT": func(c *Call) (interface{}, error) {
			str := c.String(0)
			count := int(c.OptNumber(1, 1))
			if count > len(str) {
				return str, nil
			}
			return str[len(str)-count:], nil
		},
		"TRIM": func(c *Call) (interface{}, error) {
			str := c.String(0)
			return strings.TrimSpace(str), nil
		},
		"UPPER": func(c *Call) (interface{}, error) {
			str := c.String(0)
			return strings.ToUpper(str), nil
		},
	}
}
