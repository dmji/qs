package qs

import (
	"errors"
	"reflect"
)

type MarshalerFactoryFunc func(t reflect.Type, opts *MarshalOptions) (Marshaler, error)
type MarshalerFunc func(v reflect.Value, opts *MarshalOptions) ([]string, error)
type PrimitiveMarshalerFunc func(v reflect.Value, opts *MarshalOptions) (string, error)

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

	RegisterSubFactory(k reflect.Kind, fn MarshalerFactoryFunc) error
	RegisterCustomType(k reflect.Type, fn PrimitiveMarshalerFunc) error
	RegisterKindOverride(k reflect.Kind, fn PrimitiveMarshalerFunc) error
}

// marshalerFactory implements the MarshalerFactory interface.
type marshalerFactory struct {
	types             map[reflect.Type]Marshaler
	kindSubRegistries map[reflect.Kind]MarshalerFactory
	kinds             map[reflect.Kind]Marshaler

	typesOverriden             map[reflect.Type]Marshaler
	kindSubRegistriesOverriden map[reflect.Kind]MarshalerFactory
	kindsOverriden             map[reflect.Kind]Marshaler
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
	if marshaler, ok := p.typesOverriden[t]; ok {
		return marshaler, nil
	}
	if marshaler, ok := p.types[t]; ok {
		return marshaler, nil
	}

	if t.Implements(marshalQSInterfaceType) {
		return &marshalerFunc{marshalWithMarshalQS}, nil
	}

	k := t.Kind()
	if subFactory, ok := p.kindSubRegistriesOverriden[k]; ok {
		return subFactory.Marshaler(t, opts)
	}
	if subFactory, ok := p.kindSubRegistries[k]; ok {
		return subFactory.Marshaler(t, opts)
	}

	if marshaler, ok := p.kindsOverriden[k]; ok {
		return marshaler, nil
	}
	if marshaler, ok := p.kinds[k]; ok {
		return marshaler, nil
	}

	return nil, &UnhandledTypeError{Type: t}
}

func (p *marshalerFactory) RegisterSubFactory(k reflect.Kind, fn MarshalerFactoryFunc) error {
	p.kindSubRegistriesOverriden[k] = &marshalerFactoryFunc{fn}
	return nil
}

func (p *marshalerFactory) RegisterCustomType(k reflect.Type, fn PrimitiveMarshalerFunc) error {
	p.typesOverriden[k] = &primitiveMarshalerFunc{fn}
	return nil
}

func (p *marshalerFactory) RegisterKindOverride(k reflect.Kind, fn PrimitiveMarshalerFunc) error {
	p.kindsOverriden[k] = &primitiveMarshalerFunc{fn: fn}
	return nil
}

func newMarshalerFactory() *marshalerFactory {
	return &marshalerFactory{
		typesOverriden:             map[reflect.Type]Marshaler{},
		kindSubRegistriesOverriden: map[reflect.Kind]MarshalerFactory{},
		kindsOverriden:             map[reflect.Kind]Marshaler{},

		types: map[reflect.Type]Marshaler{
			timeType: &primitiveMarshalerFunc{marshalTime},
			urlType:  &primitiveMarshalerFunc{marshalURL},
		},
		kindSubRegistries: map[reflect.Kind]MarshalerFactory{
			reflect.Ptr:   &marshalerFactoryFunc{newPtrMarshaler},
			reflect.Array: &marshalerFactoryFunc{newArrayAndSliceMarshaler},
			reflect.Slice: &marshalerFactoryFunc{newArrayAndSliceMarshaler},
		},
		kinds: map[reflect.Kind]Marshaler{
			reflect.String: &primitiveMarshalerFunc{marshalString},
			reflect.Bool:   &primitiveMarshalerFunc{marshalBool},

			reflect.Int:   &primitiveMarshalerFunc{marshalInt},
			reflect.Int8:  &primitiveMarshalerFunc{marshalInt},
			reflect.Int16: &primitiveMarshalerFunc{marshalInt},
			reflect.Int32: &primitiveMarshalerFunc{marshalInt},
			reflect.Int64: &primitiveMarshalerFunc{marshalInt},

			reflect.Uint:   &primitiveMarshalerFunc{marshalUint},
			reflect.Uint8:  &primitiveMarshalerFunc{marshalUint},
			reflect.Uint16: &primitiveMarshalerFunc{marshalUint},
			reflect.Uint32: &primitiveMarshalerFunc{marshalUint},
			reflect.Uint64: &primitiveMarshalerFunc{marshalUint},

			reflect.Float32: &primitiveMarshalerFunc{marshalFloat},
			reflect.Float64: &primitiveMarshalerFunc{marshalFloat},
		},
	}
}

// marshalerFactoryFunc implements the MarshalerFactory interface.

type marshalerFactoryFunc struct {
	fn MarshalerFactoryFunc
}

func (f marshalerFactoryFunc) Marshaler(t reflect.Type, opts *MarshalOptions) (Marshaler, error) {
	return f.fn(t, opts)
}

func (p *marshalerFactoryFunc) RegisterSubFactory(k reflect.Kind, fn MarshalerFactoryFunc) error {
	return errors.New("not implemented")
}

func (p *marshalerFactoryFunc) RegisterCustomType(k reflect.Type, fn PrimitiveMarshalerFunc) error {
	return errors.New("not implemented")
}

func (p *marshalerFactoryFunc) RegisterKindOverride(k reflect.Kind, fn PrimitiveMarshalerFunc) error {
	return errors.New("not implemented")
}

// marshalerFunc implements the Marshaler interface.
type marshalerFunc struct {
	fn MarshalerFunc
}

func (f marshalerFunc) Marshal(v reflect.Value, opts *MarshalOptions) ([]string, error) {
	return f.fn(v, opts)
}

func (p *marshalerFunc) RegisterSubFactory(k reflect.Kind, fn MarshalerFactoryFunc) error {
	return errors.New("not implemented")
}

func (p *marshalerFunc) RegisterCustomType(k reflect.Type, fn PrimitiveMarshalerFunc) error {
	return errors.New("not implemented")
}

func (p *marshalerFunc) RegisterKindOverride(k reflect.Kind, fn PrimitiveMarshalerFunc) error {
	return errors.New("not implemented")
}

// primitiveMarshalerFunc implements the Marshaler interface.
type primitiveMarshalerFunc struct {
	fn PrimitiveMarshalerFunc
}

func (f primitiveMarshalerFunc) Marshal(v reflect.Value, opts *MarshalOptions) ([]string, error) {
	s, err := f.fn(v, opts)
	if err != nil {
		return nil, err
	}
	return []string{s}, nil
}

func (p *primitiveMarshalerFunc) RegisterSubFactory(k reflect.Kind, fn MarshalerFactoryFunc) error {
	return errors.New("not implemented")
}

func (p *primitiveMarshalerFunc) RegisterCustomType(k reflect.Type, fn PrimitiveMarshalerFunc) error {
	return errors.New("not implemented")
}

func (p *primitiveMarshalerFunc) RegisterKindOverride(k reflect.Kind, fn PrimitiveMarshalerFunc) error {
	return errors.New("not implemented")
}
