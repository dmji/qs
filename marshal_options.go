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

	// Defaults for tag  options
	TagOptionsDefaults       *MarshalTagOptions
	TagCommonOptionsDefaults *CommonTagOptions
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

	// Init Unmarshal Tag Options
	if opts.TagOptionsDefaults == nil {
		opts.TagOptionsDefaults = NewUndefinedMarshalTagOptions()
	}

	opts.TagOptionsDefaults.InitDefaults()

	// Init Common Tag Options
	if opts.TagCommonOptionsDefaults == nil {
		opts.TagCommonOptionsDefaults = NewUndefinedCommonTagOptions()
	}

	opts.TagCommonOptionsDefaults.InitDefaults()

	return &opts
}

// option appliers
func WithMarshalPresence(presence MarshalPresence) func(*QSMarshaler) {
	return func(m *QSMarshaler) {
		m.opts.TagOptionsDefaults.Presence = presence
	}
}

func WithCustomUrlQueryToStringEncoder(fn func(values url.Values) string) func(*QSMarshaler) {
	return func(m *QSMarshaler) {
		m._EncodeValues = fn
	}
}

func WithMarshalOptionSliceSeparator(value OptionSliceSeparator) func(*QSMarshaler) {
	return func(m *QSMarshaler) {
		m.opts.TagCommonOptionsDefaults.SliceSeparator = value
	}
}
