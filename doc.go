// Package exprel provides a Spreadsheet-like expression evaluator.
//
// Introduction
//
// All expressions return a single value.
//
// Values
//
// The following values are can be returned and used by an expression:
//  string
//  float64
//  bool
//
// Sources may also return the following type, which defines a function that
// can be called from an expression:
//  func(arguments ...interface{}) (value interface{}, ok bool)
//
// Operators
//
// The following operators and built-ins are defined:
//                  Usage         Operand type   Notes
//  --------------------------------------------------
//  Addition        a + b         float6
//  Subtraction     a - b         float64
//  Multiplication  a * b         float64
//  Divition        a / b         float64
//  Exponentiation  a ^ b         float64
//  Concatenation   a & b         string
//
//  Logical AND     AND(a;b;...)  bool           Operands lazily evaluated
//  Logical OR      OR(a;b;...)   bool           Operands lazily evaluated
//  Logical NOT     NOT(a)        bool           Operands lazily evaluated
//
//  Condition       IF(cond;a;b)  bool;ANY;ANY   Lazily evaluated
//
//  Boolean true    TRUE()        N/A
//  Boolean false   FALSE()       N/A
//
//
// The following functions are defined as part of Base:
//             Argument type    Return type
//  ---------------------------------------
//  EXP(a)     float64          float64
//  PI()                        float64
//
//  LEN(a)     string           float64
//  LOWER(a)   string           string
//  REPT(a;b)  string, float64  string
//  UPPER(a)   string           string
//
package exprel
