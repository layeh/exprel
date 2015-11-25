package exprel

// Func is a function that can be executed from an Expression.
type Func func(values ...interface{}) (interface{}, error)

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
func (fn SourceFunc) Get(name string) (interface{}, bool) {
	return fn(name)
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
	case bool, string, float64, Func, func(values ...interface{}) (interface{}, error):
		return value, true
	default:
		return nil, false
	}
}

// Sources is a slice of sources. The first Source, in order, to return ok,
// will have its value returned.
type Sources []Source

// Get implements Source.
func (so Sources) Get(name string) (interface{}, bool) {
	for _, s := range so {
		value, ok := s.Get(name)
		if ok {
			return value, true
		}
	}
	return nil, false
}
