package exprel

import (
	"fmt"
)

// SyntaxError represents an error that is triggered when parsing an
// expression.
type SyntaxError struct {
	Message  string
	Position int
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("exprel: syntax error near index %d: %s", e.Position, e.Message)
}

// RuntimeError represents an error that is triggered when evaluating an
// expression.
type RuntimeError struct {
	Message string
	Err     error
}

func (e *RuntimeError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("exprel: runtime error: %s", e.Err.Error())
	}
	return fmt.Sprintf("exprel: runtime error: %s", e.Message)
}
