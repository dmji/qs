package qs

import (
	"fmt"
	"net/url"
)

// UnmarshalOptions is used as a parameter by the NewUnmarshaler function.
type UnmarshalOptions struct {
	// NameTransformer is used to transform struct field names into a query
	// string names when they aren't set explicitly in the struct field tag.
	// If this field is nil then NewUnmarshaler uses a default function that
	// converts the CamelCase field names to snake_case which is popular
	// with query strings.
	NameTransformer NameTransformFunc

	// SliceToString is used by Unmarshaler.Unmarshal when it unmarshals into a
	// primitive non-array struct field. In such cases unmarshaling a []string
	// (which is the value type of the url.Values map) requires transforming
	// the []string into a single string before unmarshaling.
	//
	// E.g.: If you have a struct field "Count int" but you receive a query
	// string "count=5&count=6&count=8" then the incoming []string{"5", "6", "8"}
	// has to be converted into a single string before setting the "Count int"
	// field.
	//
	// If you don't initialise this field then a default function is used that
	// fails if the input array doesn't contain exactly one item.
	//
	// In some cases you might want to provide your own function that is more
	// forgiving. E.g.: you can provide a function that picks the first or last
	// item, or concatenates/joins the whole list into a single string.
	SliceToString SliceToStringFunc

	// ValuesUnmarshalerFactory is used by QSUnmarshaler to create ValuesUnmarshaler
	// objects for specific types. If this field is nil then NewUnmarshaler uses
	// a default builtin factory.
	ValuesUnmarshalerFactory ValuesUnmarshalerFactory

	// UnmarshalerFactory is used by QSUnmarshaler to create Unmarshaler
	// objects for specific types. If this field is nil then NewUnmarshaler uses
	// a default builtin factory.
	UnmarshalerFactory UnmarshalerFactory

	// DefaultUnmarshalPresence is used for the unmarshaling of struct fields
	// that don't have an explicit UnmarshalPresence option set in their tags.
	_DefaultUnmarshalPresence UnmarshalPresence
}

// NewDefaultUnmarshalOptions creates a new UnmarshalOptions in which every field
// is set to its default value.
func NewDefaultUnmarshalOptions() *UnmarshalOptions {
	return prepareUnmarshalOptions(UnmarshalOptions{})
}

// defaultSliceToString is used by the NewUnmarshaler function when
// its UnmarshalOptions.SliceToString parameter is nil.
var defaultSliceToString = func(a []string) (string, error) {
	if len(a) != 1 {
		return "", fmt.Errorf("SliceToString expects array length == 1. array=%v", a)
	}
	return a[0], nil
}

func prepareUnmarshalOptions(opts UnmarshalOptions) *UnmarshalOptions {
	if opts.NameTransformer == nil {
		opts.NameTransformer = snakeCase
	}
	if opts.SliceToString == nil {
		opts.SliceToString = defaultSliceToString
	}

	if opts.ValuesUnmarshalerFactory == nil {
		opts.ValuesUnmarshalerFactory = newValuesUnmarshalerFactory()
	}
	opts.ValuesUnmarshalerFactory = newValuesUnmarshalerCache(opts.ValuesUnmarshalerFactory)

	if opts.UnmarshalerFactory == nil {
		opts.UnmarshalerFactory = newUnmarshalerFactory()
	}
	opts.UnmarshalerFactory = newUnmarshalerCache(opts.UnmarshalerFactory)

	if opts._DefaultUnmarshalPresence == UnmarshalPresenceUPUnspecified {
		opts._DefaultUnmarshalPresence = UnmarshalPresenceOpt
	}
	return &opts
}

// option appliers
func WithUnmarshalPresence(presence UnmarshalPresence) func(*QSUnmarshaler) {
	return func(m *QSUnmarshaler) {
		m.opts._DefaultUnmarshalPresence = presence
	}
}

func WithCustomStringToUrlQueryParser(fn func(query string) (url.Values, error)) func(*QSUnmarshaler) {
	return func(m *QSUnmarshaler) {
		m.stringToQueryParser = fn
	}
}
