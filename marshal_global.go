package qs

import (
	"net/url"
	"reflect"
)

// DefaultMarshaler is the marshaler used by the Marshal, MarshalValues,
// CanMarshal and CanMarshalType functions.
var DefaultMarshaler = NewMarshaler(&MarshalOptions{})

// Marshal marshals an object into a query string. The type of the object must
// be supported by the ValuesMarshalerFactory of the marshaler. By default only
// structs and maps satisfy this condition without using a custom
// ValuesMarshalerFactory.
//
// If you use a map then the key type has to be string or a type with string as
// its underlying type and the map value type can be anything that can be used
// as a struct field for marshaling.
//
// A struct value is marshaled by adding its fields one-by-one to the query
// string. Only exported struct fields are marshaled. The struct field tag can
// contain qs package specific options in the following format:
//
//	FieldName bool `qs:"[name][,option1[,option2[...]]]"`
//
//	- If name is "-" then this field is skipped just like unexported fields.
//	- If name is omitted then it defaults to the snake_case of the FieldName.
//	  The snake_case transformation can be replaced with a field name to query
//	  string name converter function by creating a custom marshaler.
//	- For marshaling you can specify one of the keepempty and omitempty options.
//	  If none of them is specified then the keepempty option is the default but
//	  this default can be changed by using a custom marshaler object.
//
//	Examples:
//	FieldName bool `qs:"-"
//	FieldName bool `qs:"name_in_query_str"
//	FieldName bool `qs:"name_in_query_str,keepempty"
//	FieldName bool `qs:",omitempty"
//
// Anonymous struct fields are marshaled as if their inner exported fields were
// fields in the outer struct.
//
// Pointer fields are omitted when they are nil otherwise they are marshaled as
// the value pointed to.
//
// Items of array and slice fields are encoded by adding multiple items with the
// same key to the query string. E.g.: arr=[]byte{1, 2} is encoded as "arr=1&arr=2".
// You can change this behavior by creating a custom marshaler with its custom
// MarshalerFactory that provides your custom marshal logic for the given slice
// and/or array types.
//
// When a field is marshaled with the omitempty option then the field is skipped
// if it has the zero value of its type.
// A field is marshaled with the omitempty option when its tag explicitly
// specifies omitempty or when the tag contains neither omitempty nor keepempty
// but the marshaler's default marshal option is omitempty.
func Marshal(i interface{}) (string, error) {
	return DefaultMarshaler.Marshal(i)
}

// MarshalValues is the same as Marshal but returns a url.Values instead of a
// query string.
func MarshalValues(i interface{}) (url.Values, error) {
	return DefaultMarshaler.MarshalValues(i)
}

// CheckMarshal returns an error if the type of the given object can't be
// marshaled into a url.Values or query string. By default only maps and structs
// can be marshaled into query strings given that all of their fields or values
// can be marshaled to []string (which is the value type of the url.Values map).
//
// It performs the check on the type of the object without traversing or
// marshaling the object.
func CheckMarshal(i interface{}) error {
	return DefaultMarshaler.CheckMarshal(i)
}

// CheckMarshalType returns an error if the given type can't be marshaled
// into a url.Values or query string. By default only maps and structs
// can be marshaled int query strings given that all of their fields or values
// can be marshaled to []string (which is the value type of the url.Values map).
func CheckMarshalType(t reflect.Type) error {
	return DefaultMarshaler.CheckMarshalType(t)
}
