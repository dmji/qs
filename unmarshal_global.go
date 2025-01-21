package qs

import (
	"net/url"
	"reflect"
)

// DefaultUnmarshaler is the unmarshaler used by the Unmarshal, UnmarshalValues,
// CanUnmarshal and CanUnmarshalType functions.
var DefaultUnmarshaler = NewUnmarshaler(&UnmarshalerDefaultOptions{})

// Unmarshal unmarshals a query string and stores the result to the object
// pointed to by the given pointer.
//
// Unmarshal uses the inverse of the encodings that Marshal uses.
//
// A struct field tag can optionally contain one of the opt, nil and req options
// for unmarshaling. If it contains none of these then opt is the default but
// the default can also be changed by using a custom marshaler. The
// UnmarshalPresence of a field is used only when the query string doesn't
// contain a value for it:
//   - nil succeeds and keeps the original field value
//   - opt succeeds and keeps the original field value but in case of
//     pointer-like types (pointers, slices) with nil field value it initialises
//     the field with a newly created object.
//   - req causes the unmarshal operation to fail with an error that can be
//     detected using qs.IsRequiredFieldError.
//
// When unmarshaling a nil pointer field that is present in the query string
// the pointer is automatically initialised even if it has the nil option in
// its tag.
func Unmarshal(into interface{}, queryString string) error {
	return DefaultUnmarshaler.Unmarshal(into, queryString)
}

// UnmarshalValues is the same as Unmarshal but it unmarshals from a url.Values
// instead of a query string.
func UnmarshalValues(into interface{}, values url.Values) error {
	return DefaultUnmarshaler.UnmarshalValues(into, values)
}

// CheckUnmarshal returns an error if the type of the given object can't be
// unmarshaled from a url.Vales or query string. By default only maps and structs
// can be unmarshaled from query strings given that all of their fields or values
// can be unmarshaled from []string (which is the value type of the url.Values map).
//
// It performs the check on the type of the object without traversing or
// unmarshaling the object.
func CheckUnmarshal(into interface{}) error {
	return DefaultUnmarshaler.CheckUnmarshal(into)
}

// CheckUnmarshalType returns an error if the given type can't be unmarshaled
// from a url.Vales or query string. By default only maps and structs
// can be unmarshaled from query strings given that all of their fields or values
// can be unmarshaled from []string (which is the value type of the url.Values map).
func CheckUnmarshalType(t reflect.Type) error {
	return DefaultUnmarshaler.CheckUnmarshalType(t)
}

func RegisterSubFactoryUnmarshaler(k reflect.Kind, fn UnmarshalerFactoryFunc) error {
	return DefaultUnmarshaler.opts.UnmarshalerFactory.RegisterSubFactory(k, fn)
}

func RegisterCustomTypeUnmarshaler(k reflect.Type, fn PrimitiveUnmarshalerFunc) error {
	return DefaultUnmarshaler.opts.UnmarshalerFactory.RegisterCustomType(k, fn)
}

func RegisterKindOverrideUnmarshaler(k reflect.Kind, fn PrimitiveUnmarshalerFunc) error {
	return DefaultUnmarshaler.opts.UnmarshalerFactory.RegisterKindOverride(k, fn)
}

func ApplyOptionsUnmarshal(opts ...func(*QSUnmarshaler)) {
	for _, opt := range opts {
		opt(DefaultUnmarshaler)
	}
}
