package qs

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
)

// MarshalOptions is used as a parameter by the NewMarshaler function.
type MarshalOptions struct {
	// NameTransformer is used to transform struct field names into a query
	// string names when they aren't set explicitly in the struct field tag.
	// If this field is nil then NewMarshaler uses a default function that
	// converts the CamelCase field names to snake_case which is popular
	// with query strings.
	NameTransformer NameTransformFunc

	// ValuesMarshalerFactory is used by QSMarshaler to create ValuesMarshaler
	// objects for specific types. If this field is nil then NewMarshaler uses
	// a default builtin factory.
	ValuesMarshalerFactory ValuesMarshalerFactory

	// MarshalerFactory is used by QSMarshaler to create Marshaler
	// objects for specific types. If this field is nil then NewMarshaler uses
	// a default builtin factory.
	MarshalerFactory MarshalerFactory

	// DefaultMarshalPresence is used for the marshaling of struct fields that
	// don't have an explicit MarshalPresence option set in their tags.
	// This option is used for every item when you marshal a map[string]WhateverType
	// instead of a struct because map items can't have a tag to override this.
	DefaultMarshalPresence MarshalPresence
}

// QSMarshaler objects can be created by calling NewMarshaler and they can be
// used to marshal structs or maps into query strings or url.Values.
type QSMarshaler struct {
	opts *MarshalOptions
}

// NewMarshaler returns a new QSMarshaler object.
func NewMarshaler(opts *MarshalOptions) *QSMarshaler {
	return &QSMarshaler{
		opts: prepareMarshalOptions(*opts),
	}
}

// Marshal marshals a given object into a query string.
// See the documentation of the global Marshal func.
func (p *QSMarshaler) Marshal(i interface{}) (string, error) {
	values, err := p.MarshalValues(i)
	if err != nil {
		return "", err
	}
	return values.Encode(), nil
}

// MarshalValues marshals a given object into a url.Values.
// See the documentation of the global MarshalValues func.
func (p *QSMarshaler) MarshalValues(i interface{}) (url.Values, error) {
	v := reflect.ValueOf(i)
	if !v.IsValid() {
		return nil, errors.New("received an empty interface")
	}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, fmt.Errorf("nil pointer of type %T", i)
		}
		v = v.Elem()
	}

	vum, err := p.opts.ValuesMarshalerFactory.ValuesMarshaler(v.Type(), p.opts)
	if err != nil {
		return nil, err
	}
	return vum.MarshalValues(v, p.opts)
}

// CheckMarshal check whether the type of the given object supports
// marshaling into query strings.
// See the documentation of the global CheckMarshal func.
func (p *QSMarshaler) CheckMarshal(i interface{}) error {
	return p.CheckMarshalType(reflect.TypeOf(i))
}

// CheckMarshalType check whether the given type supports marshaling into
// query strings. See the documentation of the global CheckMarshalType func.
func (p *QSMarshaler) CheckMarshalType(t reflect.Type) error {
	if t == nil {
		return errors.New("nil type")
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	_, err := p.opts.ValuesMarshalerFactory.ValuesMarshaler(t, p.opts)
	return err
}

// NewDefaultMarshalOptions creates a new MarshalOptions in which every field
// is set to its default value.
func NewDefaultMarshalOptions() *MarshalOptions {
	return prepareMarshalOptions(MarshalOptions{})
}

// defaultValuesMarshalerFactory is used by the NewMarshaler function when its
// MarshalOptions.ValuesMarshalerFactory parameter is nil.
var defaultValuesMarshalerFactory = newValuesMarshalerFactory()

// defaultMarshalerFactory is used by the NewMarshaler function when its
// MarshalOptions.MarshalerFactory parameter is nil. This variable is set
// to a factory object that handles most builtin types (arrays, pointers,
// bool, int, etc...). If a type implements the MarshalQS interface then this
// factory returns an marshaler object that allows instances of the given type
// to marshal themselves.
var defaultMarshalerFactory = newMarshalerFactory()

func prepareMarshalOptions(opts MarshalOptions) *MarshalOptions {
	if opts.NameTransformer == nil {
		opts.NameTransformer = snakeCase
	}

	if opts.ValuesMarshalerFactory == nil {
		opts.ValuesMarshalerFactory = defaultValuesMarshalerFactory
	}
	opts.ValuesMarshalerFactory = newValuesMarshalerCache(opts.ValuesMarshalerFactory)

	if opts.MarshalerFactory == nil {
		opts.MarshalerFactory = defaultMarshalerFactory
	}
	opts.MarshalerFactory = newMarshalerCache(opts.MarshalerFactory)

	if opts.DefaultMarshalPresence == MarshalPresenceMPUnspecified {
		opts.DefaultMarshalPresence = MarshalPresenceKeepEmpty
	}
	return &opts
}
