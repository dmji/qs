package qs

import (
	"reflect"
)

// Marshaler can marshal a reflect.Value into a []string.
type Marshaler interface {
	// Marshal marshals the given v value using opts into a []string.
	// Note that []string is the value type of the standard url.Values which is
	// a map[string][]string.
	Marshal(v reflect.Value, opts *MarshalOptions) ([]string, error)
}

// MarshalerFactory can create Marshaler objects for various types.
type MarshalerFactory interface {
	// Marshaler returns a Marshaler object for the given t type and opts
	// options.
	Marshaler(t reflect.Type, opts *MarshalOptions) (Marshaler, error)
}

// marshalerFactory implements the MarshalerFactory interface.
type marshalerFactory struct {
	Types             map[reflect.Type]Marshaler
	KindSubRegistries map[reflect.Kind]MarshalerFactory
	Kinds             map[reflect.Kind]Marshaler
}

// MarshalQS is an interface that can be implemented by any type that
// wants to handle its own marshaling instead of relying on the default
// marshaling provided by this package.
type MarshalQS interface {
	// MarshalQS is essentially the same as the Marshaler.Marshal
	// method without its v parameter.
	MarshalQS(opts *MarshalOptions) ([]string, error)
}

var marshalQSInterfaceType = reflect.TypeOf((*MarshalQS)(nil)).Elem()

func (p *marshalerFactory) Marshaler(t reflect.Type, opts *MarshalOptions) (Marshaler, error) {
	if marshaler, ok := p.Types[t]; ok {
		return marshaler, nil
	}

	if t.Implements(marshalQSInterfaceType) {
		return marshalerFunc(marshalWithMarshalQS), nil
	}

	k := t.Kind()
	if subFactory, ok := p.KindSubRegistries[k]; ok {
		return subFactory.Marshaler(t, opts)
	}
	if marshaler, ok := p.Kinds[k]; ok {
		return marshaler, nil
	}

	return nil, &unhandledTypeError{Type: t}
}

func newMarshalerFactory() *marshalerFactory {
	return &marshalerFactory{
		Types: map[reflect.Type]Marshaler{
			timeType: primitiveMarshalerFunc(marshalTime),
			urlType:  primitiveMarshalerFunc(marshalURL),
		},
		KindSubRegistries: map[reflect.Kind]MarshalerFactory{
			reflect.Ptr:   marshalerFactoryFunc(newPtrMarshaler),
			reflect.Array: marshalerFactoryFunc(newArrayAndSliceMarshaler),
			reflect.Slice: marshalerFactoryFunc(newArrayAndSliceMarshaler),
		},
		Kinds: map[reflect.Kind]Marshaler{
			reflect.String: primitiveMarshalerFunc(marshalString),
			reflect.Bool:   primitiveMarshalerFunc(marshalBool),

			reflect.Int:   primitiveMarshalerFunc(marshalInt),
			reflect.Int8:  primitiveMarshalerFunc(marshalInt),
			reflect.Int16: primitiveMarshalerFunc(marshalInt),
			reflect.Int32: primitiveMarshalerFunc(marshalInt),
			reflect.Int64: primitiveMarshalerFunc(marshalInt),

			reflect.Uint:   primitiveMarshalerFunc(marshalUint),
			reflect.Uint8:  primitiveMarshalerFunc(marshalUint),
			reflect.Uint16: primitiveMarshalerFunc(marshalUint),
			reflect.Uint32: primitiveMarshalerFunc(marshalUint),
			reflect.Uint64: primitiveMarshalerFunc(marshalUint),

			reflect.Float32: primitiveMarshalerFunc(marshalFloat),
			reflect.Float64: primitiveMarshalerFunc(marshalFloat),
		},
	}
}

// marshalerFactoryFunc implements the MarshalerFactory interface.
type marshalerFactoryFunc func(t reflect.Type, opts *MarshalOptions) (Marshaler, error)

func (f marshalerFactoryFunc) Marshaler(t reflect.Type, opts *MarshalOptions) (Marshaler, error) {
	return f(t, opts)
}

// marshalerFunc implements the Marshaler interface.
type marshalerFunc func(v reflect.Value, opts *MarshalOptions) ([]string, error)

func (f marshalerFunc) Marshal(v reflect.Value, opts *MarshalOptions) ([]string, error) {
	return f(v, opts)
}

// primitiveMarshalerFunc implements the Marshaler interface.
type primitiveMarshalerFunc func(v reflect.Value, opts *MarshalOptions) (string, error)

func (f primitiveMarshalerFunc) Marshal(v reflect.Value, opts *MarshalOptions) ([]string, error) {
	s, err := f(v, opts)
	if err != nil {
		return nil, err
	}
	return []string{s}, nil
}
