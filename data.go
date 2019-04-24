package exprel

import (
	"context"
	"strconv"
)

// Func is a function that can be executed from an Expression.
type Func = func(call *Call) (interface{}, error)

// Call contains information about an expression function call.
type Call struct {
	// The name used to invoke the function.
	Name string
	// The arguments passed to the function.
	Values []interface{}

	ctx context.Context
}

// Context returns the context for the current function call.
func (c *Call) Context() context.Context {
	return c.ctx
}

// String returns the ith argument, iff it is a string. Otherwise, the function
// panics with a *RuntimeError.
func (c *Call) String(i int) string {
	if len(c.Values) <= i {
		panic(&RuntimeError{Message: c.Name + " expects argument " + strconv.Itoa(i) + " to be string"})
	}
	value, ok := c.Values[i].(string)
	if !ok {
		panic(&RuntimeError{Message: c.Name + " expects argument " + strconv.Itoa(i) + " to be string"})
	}
	return value
}

// OptString returns the ith argument, iff it is a string. If the ith argument
// does not exist, def is returned. If the ith argument is not a string, the
// function panics with a *RuntimeError.
func (c *Call) OptString(i int, def string) string {
	if len(c.Values) <= i {
		return def
	}
	value, ok := c.Values[i].(string)
	if !ok {
		panic(&RuntimeError{Message: c.Name + " expects argument " + strconv.Itoa(i) + " to be string"})
	}
	return value
}

// Number returns the ith argument, iff it is a float64. Otherwise, the
// function panics with a *RuntimeError.
func (c *Call) Number(i int) float64 {
	if len(c.Values) <= i {
		panic(&RuntimeError{Message: c.Name + " expects argument " + strconv.Itoa(i) + " to be float64"})
	}
	value, ok := c.Values[i].(float64)
	if !ok {
		panic(&RuntimeError{Message: c.Name + " expects argument " + strconv.Itoa(i) + " to be float64"})
	}
	return value
}

// OptNumber returns the ith argument, iff it is a float64. If the ith argument
// does not exist, def is returned. If the ith argument is not a number, the
// function panics with a *RuntimeError.
func (c *Call) OptNumber(i int, def float64) float64 {
	if len(c.Values) <= i {
		return def
	}
	value, ok := c.Values[i].(float64)
	if !ok {
		panic(&RuntimeError{Message: c.Name + " expects argument " + strconv.Itoa(i) + " to be float64"})
	}
	return value
}

// Boolean returns the ith argument, iff it is a bool. Otherwise, the function
// panics with a *RuntimeError.
func (c *Call) Boolean(i int) bool {
	if len(c.Values) <= i {
		panic(&RuntimeError{Message: c.Name + " expects argument " + strconv.Itoa(i) + " to be bool"})
	}
	value, ok := c.Values[i].(bool)
	if !ok {
		panic(&RuntimeError{Message: c.Name + " expects argument " + strconv.Itoa(i) + " to be bool"})
	}
	return value
}

// OptBoolean returns the ith argument, iff it is a bool. If the ith argument
// does not exist, def is returned. If the ith argument is not a bool, the
// function panics with a *RuntimeError.
func (c *Call) OptBoolean(i int, def bool) bool {
	if len(c.Values) <= i {
		return def
	}
	value, ok := c.Values[i].(bool)
	if !ok {
		panic(&RuntimeError{Message: c.Name + " expects argument " + strconv.Itoa(i) + " to be bool"})
	}
	return value
}

// Source is a source of data for an expression. Get is called when an
// identifier needs to be evaluated.
type Source interface {
	Get(name string) (value interface{}, ok bool)
}

// EmptySource is a Source that contains no values.
var EmptySource Source

type emptySource struct{}

func (emptySource) Get(name string) (interface{}, bool) {
	return nil, false
}

func init() {
	EmptySource = emptySource{}
}

// SourceFunc is a Source that looks up an identifier via a function.
type SourceFunc func(name string) (value interface{}, ok bool)

// Get implements Source.
func (f SourceFunc) Get(name string) (interface{}, bool) {
	return f(name)
}

// SourceMap is a Source that looks up an identifier in a map.
type SourceMap map[string]interface{}

// Get implements Source.
func (m SourceMap) Get(name string) (interface{}, bool) {
	value, ok := m[name]
	if !ok {
		return nil, false
	}
	switch value.(type) {
	case bool, string, float64, Func:
		return value, true
	default:
		return nil, false
	}
}

// Sources is a slice of sources. The first Source, in order, to return ok,
// will have its value returned.
type Sources []Source

// Get implements Source.
func (s Sources) Get(name string) (interface{}, bool) {
	for _, s := range s {
		value, ok := s.Get(name)
		if ok {
			return value, true
		}
	}
	return nil, false
}
