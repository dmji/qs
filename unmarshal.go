package qs

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
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

// QSUnmarshaler objects can be created by calling NewUnmarshaler and they can be
// used to unmarshal query strings or url.Values into structs or maps.
type QSUnmarshaler struct {
	opts *UnmarshalOptions

	_ParseQuery func(query string) (url.Values, error)
}

// NewUnmarshaler returns a new QSUnmarshaler object.
func NewUnmarshaler(prm *UnmarshalOptions, opts ...func(p *QSUnmarshaler)) *QSUnmarshaler {
	p := &QSUnmarshaler{
		opts:        prepareUnmarshalOptions(*prm),
		_ParseQuery: url.ParseQuery,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *QSUnmarshaler) RegisterSubFactory(k reflect.Kind, fn UnmarshalerFactoryFunc) error {
	return p.opts.UnmarshalerFactory.RegisterSubFactory(k, fn)

}

func (p *QSUnmarshaler) RegisterCustomType(k reflect.Type, fn PrimitiveUnmarshalerFunc) error {
	return p.opts.UnmarshalerFactory.RegisterCustomType(k, fn)

}

func (p *QSUnmarshaler) RegisterKindOverride(k reflect.Kind, fn PrimitiveUnmarshalerFunc) error {
	return p.opts.UnmarshalerFactory.RegisterKindOverride(k, fn)
}

// Unmarshal unmarshals an object from a query string.
// See the documentation of the global Unmarshal func.
func (p *QSUnmarshaler) Unmarshal(into interface{}, queryString string) error {
	values, err := p._ParseQuery(queryString)
	if err != nil {
		return fmt.Errorf("error parsing query string %q :: %v", queryString, err)
	}
	return p.UnmarshalValues(into, values)
}

// UnmarshalValues unmarshals an object from a url.Values.
// See the documentation of the global UnmarshalValues func.
func (p *QSUnmarshaler) UnmarshalValues(into interface{}, values url.Values) error {
	pv := reflect.ValueOf(into)
	if !pv.IsValid() {
		return errors.New("received an empty interface")
	}
	if pv.Kind() != reflect.Ptr {
		return fmt.Errorf("expected a pointer, got %T", into)
	}
	if pv.IsNil() {
		return fmt.Errorf("nil pointer of type %T", into)
	}
	v := pv.Elem()

	vum, err := p.opts.ValuesUnmarshalerFactory.ValuesUnmarshaler(v.Type(), p.opts)
	if err != nil {
		return err
	}
	return vum.UnmarshalValues(v, values, p.opts)
}

// CheckUnmarshal check whether the type of the given object supports
// unmarshaling from query strings.
// See the documentation of the global CheckUnmarshal func.
func (p *QSUnmarshaler) CheckUnmarshal(into interface{}) error {
	return p.CheckUnmarshalType(reflect.TypeOf(into))
}

// CheckUnmarshalType check whether the given type supports unmarshaling from
// query strings. See the documentation of the global CheckUnmarshalType func.
func (p *QSUnmarshaler) CheckUnmarshalType(t reflect.Type) error {
	if t == nil {
		return errors.New("nil type")
	}
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("expected a pointer, got %v", t)
	}
	_, err := p.opts.ValuesUnmarshalerFactory.ValuesUnmarshaler(t.Elem(), p.opts)
	return err
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
