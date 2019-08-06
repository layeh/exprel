// Package exprel provides a Spreadsheet-like expression evaluator.
//
//   // Quick start
//
//   import (
//     "layeh.com/exprel"
//   )
//
//   data := map[string]interface{}{
//     "name": "Tim",
//   }
//   expression := `=LOWER(name) & ".jpg"`
//   filename, err := exprel.String(exprel.Evaluate(expression, exprel.SourceMap(data)))
//   if err != nil {
//     panic(err)
//   }
//   // filename = "tim.jpg"
//
// Introduction
//
// All expressions return a single value. Here are a few examples of some valid
// expressions and their return values:
//  Expression                        Return value
//  ----------------------------------------------
//  Hey there                         "Hey there"
//  1234                              "1234"
//  =5+5*2                            15
//  ="A" & " " & "B"                  "A B"
//  =IF(AND(NOT(FALSE());1=1);1+2;2)  3
//
// Expressions with logic must start with an equals sign (=). Otherwise, the
// evaluated value is simply the source string.
//
// Values
//
// The following values are can be returned by and used in an expression:
//  string
//  float64 (number)
//  bool (boolean)
//
// Sources may also return the following type, which defines a function that
// can be called from an expression:
//  func(c *Call) (value interface{}, err error)
//
// Operators
//
// The following operators and built-ins are defined:
//                    Usage              Types
//  ------------------------------------------
//  Addition          a + b              number
//  Subtraction       a - b              number
//  Multiplication    a * b              number
//  Division          a / b              number
//  Exponentiation    a ^ b              number
//  Modulo            a % b              number
//  Concatenation     a & b              string
//
//  Equality          a = b              string, number, boolean
//  Inequality        a <> b             string, number, boolean
//  Greater than      a > b              string, number
//  Greater or equal  a >= b             string, number
//  Less than         a < b              string, number
//  Less or equal     a <= b             string, number
//
//  Logical AND       AND(bool...)
//  Logical OR        OR(bool...)
//  Logical NOT       NOT(bool)
//
//  Condition         IF(bool;ANY;ANY)
//
//  Boolean true      TRUE()
//  Boolean false     FALSE()
//
//
// The following functions are defined as part of Base:
//  CHOOSE(number index; ANY...) ANY
//    Returns the index item of the remaining arguments
//  TYPE(ANY a) number
//    Identifies the type of a. Types are mapped in the following way:
//      Number  = 1
//      String  = 2
//      Boolean = 4
//
//  ABS(number a) number
//    Returns the absolute value of a.
//  EXP(number a) number
//    Returns e^a.
//  LN(number a) number
//    Returns the natural logarithm of a.
//  LOG10(number a) number
//    Returns the base-10 logarithm of a.
//  PI() number
//    Returns π.
//  RAND() number
//    Returns a random number in the range [0, 1).
//  SIGN(number a) number
//    Returns the sign of a.
//
//  CHAR(number...) string
//    Returns a string whose code points are given as arguments.
//  JOIN(string sep; string...) string
//    Returns the trailing string arguments concatenated together with sep.
//  LEFT(string a; number count = 1) string
//    Returns the count left-most characters of a.
//  LEN(string a) number
//    Returns the length of a.
//  LOWER(string a) string
//    Returns a with all uppercase characters transformed to lowercase.
//  MID(string a; number start; number length = 1) string
//    Returns length characters of a, starting from start.
//  REPT(string a; number count) string
//    Returns the string a, repeated count times.
//  RIGHT(string a; number count = 1) string
//    Returns the count right-most characters of a.
//  SEARCH(string needle; string haystack; number start = 1) number
//    Returns the position of needle in haystack, starting from start. -1 is
//    returned if needle was not found.
//  TRIM(string a) string
//    Returns a with whitespace removed from the beginning and end.
//  UPPER(string a) string
//    Returns a with all lowercase characters transformed to uppercase.
package exprel // import "layeh.com/exprel"
