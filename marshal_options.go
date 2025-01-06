package qs

import "net/url"

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
	_DefaultMarshalPresence MarshalPresence
}

// NewDefaultMarshalOptions creates a new MarshalOptions in which every field
// is set to its default value.
func NewDefaultMarshalOptions() *MarshalOptions {
	return prepareMarshalOptions(MarshalOptions{})
}

func prepareMarshalOptions(opts MarshalOptions) *MarshalOptions {

	if opts.NameTransformer == nil {
		opts.NameTransformer = snakeCase
	}

	if opts.ValuesMarshalerFactory == nil {
		opts.ValuesMarshalerFactory = newValuesMarshalerFactory()
	}
	opts.ValuesMarshalerFactory = newValuesMarshalerCache(opts.ValuesMarshalerFactory)

	if opts.MarshalerFactory == nil {
		opts.MarshalerFactory = newMarshalerFactory()
	}
	opts.MarshalerFactory = newMarshalerCache(opts.MarshalerFactory)

	if opts._DefaultMarshalPresence == MarshalPresenceMPUnspecified {
		opts._DefaultMarshalPresence = MarshalPresenceKeepEmpty
	}
	return &opts
}

// option appliers
func WithMarshalPresence(presence MarshalPresence) func(*QSMarshaler) {
	return func(m *QSMarshaler) {
		m.opts._DefaultMarshalPresence = presence
	}
}

func WithCustomUrlQueryToStringEncoder(fn func(values url.Values) string) func(*QSMarshaler) {
	return func(m *QSMarshaler) {
		m._EncodeValues = fn
	}
}
