package qs

import (
	"errors"
	"reflect"
)

type PrimitiveUnmarshalerFunc func(v reflect.Value, s string, opts *UnmarshalOptions) error
type UnmarshalerFunc func(v reflect.Value, a []string, opts *UnmarshalOptions) error
type UnmarshalerFactoryFunc func(t reflect.Type, opts *UnmarshalOptions) (Unmarshaler, error)

// Unmarshaler can unmarshal a []string (which is the value type of the
// url.Values map) into a reflect.Value.
type Unmarshaler interface {
	// Unmarshal unmarshals the given []string using opts into v.
	//
	// If the query string doesn't contain a key for this field then Unmarshal
	// is called only if the UnmarshalPresence option of the field is Opt
	// and in that case a == nil. In such cases pointer like types (pointers,
	// arrays, maps) should initialise nil pointers with an empty object.
	// With Nil or Req options this Unmarshal method isn't called.
	//
	// The []string is the value type of the url.Values map. If your unmarshaler
	// expects only a single string value instead of an array then you can call
	// opts.SliceToString(a).
	Unmarshal(v reflect.Value, a []string, opts *UnmarshalOptions) error
}

// UnmarshalerFactory can create Unmarshaler objects for various types.
type UnmarshalerFactory interface {
	// Unmarshaler returns an Unmarshaler object for the given t type and opts
	// options.
	Unmarshaler(t reflect.Type, opts *UnmarshalOptions) (Unmarshaler, error)

	// RegisterSubFactory registers a ValuesUnmarshalerFactory for the given kind
	RegisterSubFactory(k reflect.Kind, fn UnmarshalerFactoryFunc) error
	RegisterCustomType(k reflect.Type, fn PrimitiveUnmarshalerFunc) error
	RegisterKindOverride(k reflect.Kind, fn PrimitiveUnmarshalerFunc) error
}

// unmarshalerFactory implements the UnmarshalerFactory interface.
type unmarshalerFactory struct {
	types             map[reflect.Type]Unmarshaler
	kindSubRegistries map[reflect.Kind]UnmarshalerFactory
	kinds             map[reflect.Kind]Unmarshaler

	typesOverriden             map[reflect.Type]Unmarshaler
	kindSubRegistriesOverriden map[reflect.Kind]UnmarshalerFactory
	kindsOverriden             map[reflect.Kind]Unmarshaler
}

// UnmarshalQS is an interface that can be implemented by any type that
// wants to handle its own unmarshaling instead of relying on the default
// unmarshaling provided by this package.
type UnmarshalQS interface {
	// UnmarshalQS is essentially the same as the Unmarshaler.Unmarshal
	// method without its v parameter.
	UnmarshalQS(a []string, opts *UnmarshalOptions) error
}

var unmarshalQSInterfaceType = reflect.TypeOf((*UnmarshalQS)(nil)).Elem()

func (p *unmarshalerFactory) Unmarshaler(t reflect.Type, opts *UnmarshalOptions) (Unmarshaler, error) {
	if unmarshaler, ok := p.types[t]; ok {
		return unmarshaler, nil
	}

	if reflect.PointerTo(t).Implements(unmarshalQSInterfaceType) {
		return &unmarshalerFunc{unmarshalWithUnmarshalQS}, nil
	}

	k := t.Kind()
	if subFactory, ok := p.kindSubRegistries[k]; ok {
		return subFactory.Unmarshaler(t, opts)
	}
	if unmarshaler, ok := p.kinds[k]; ok {
		return unmarshaler, nil
	}

	return nil, &unhandledTypeError{Type: t}
}

func (p *unmarshalerFactory) RegisterSubFactory(k reflect.Kind, fn UnmarshalerFactoryFunc) error {
	p.kindSubRegistriesOverriden[k] = &unmarshalerFactoryFunc{fn}
	return nil
}

func (p *unmarshalerFactory) RegisterCustomType(k reflect.Type, fn PrimitiveUnmarshalerFunc) error {
	p.typesOverriden[k] = &primitiveUnmarshalerFunc{fn}
	return nil
}

func (p *unmarshalerFactory) RegisterKindOverride(k reflect.Kind, fn PrimitiveUnmarshalerFunc) error {
	p.kindsOverriden[k] = &primitiveUnmarshalerFunc{fn: fn}
	return nil
}

func newUnmarshalerFactory() *unmarshalerFactory {
	return &unmarshalerFactory{
		types: map[reflect.Type]Unmarshaler{
			timeType: &primitiveUnmarshalerFunc{unmarshalTime},
			urlType:  &primitiveUnmarshalerFunc{unmarshalURL},
		},
		kindSubRegistries: map[reflect.Kind]UnmarshalerFactory{
			reflect.Ptr:   &unmarshalerFactoryFunc{newPtrUnmarshaler},
			reflect.Array: &unmarshalerFactoryFunc{newArrayUnmarshaler},
			reflect.Slice: &unmarshalerFactoryFunc{newSliceUnmarshaler},
		},
		kinds: map[reflect.Kind]Unmarshaler{
			reflect.String: &primitiveUnmarshalerFunc{unmarshalString},
			reflect.Bool:   &primitiveUnmarshalerFunc{unmarshalBool},

			reflect.Int:   &primitiveUnmarshalerFunc{unmarshalInt},
			reflect.Int8:  &primitiveUnmarshalerFunc{unmarshalInt},
			reflect.Int16: &primitiveUnmarshalerFunc{unmarshalInt},
			reflect.Int32: &primitiveUnmarshalerFunc{unmarshalInt},
			reflect.Int64: &primitiveUnmarshalerFunc{unmarshalInt},

			reflect.Uint:   &primitiveUnmarshalerFunc{unmarshalUint},
			reflect.Uint8:  &primitiveUnmarshalerFunc{unmarshalUint},
			reflect.Uint16: &primitiveUnmarshalerFunc{unmarshalUint},
			reflect.Uint32: &primitiveUnmarshalerFunc{unmarshalUint},
			reflect.Uint64: &primitiveUnmarshalerFunc{unmarshalUint},

			reflect.Float32: &primitiveUnmarshalerFunc{unmarshalFloat},
			reflect.Float64: &primitiveUnmarshalerFunc{unmarshalFloat},
		},
	}
}

// unmarshalerFactoryFunc implements the UnmarshalerFactory interface.
type unmarshalerFactoryFunc struct {
	fn UnmarshalerFactoryFunc
}

func (f unmarshalerFactoryFunc) Unmarshaler(t reflect.Type, opts *UnmarshalOptions) (Unmarshaler, error) {
	return f.fn(t, opts)
}

func (p *unmarshalerFactoryFunc) RegisterSubFactory(k reflect.Kind, fn UnmarshalerFactoryFunc) error {
	return errors.New("not implemented")
}

func (p *unmarshalerFactoryFunc) RegisterCustomType(k reflect.Type, fn PrimitiveUnmarshalerFunc) error {
	return errors.New("not implemented")
}

func (p *unmarshalerFactoryFunc) RegisterKindOverride(k reflect.Kind, fn PrimitiveUnmarshalerFunc) error {
	return errors.New("not implemented")
}

// unmarshalerFunc implements the Unmarshaler interface.
type unmarshalerFunc struct {
	fn UnmarshalerFunc
}

func (f unmarshalerFunc) Unmarshal(v reflect.Value, a []string, opts *UnmarshalOptions) error {
	return f.fn(v, a, opts)
}

func (p *unmarshalerFunc) RegisterSubFactory(k reflect.Kind, fn UnmarshalerFactoryFunc) error {
	return errors.New("not implemented")
}

func (p *unmarshalerFunc) RegisterCustomType(k reflect.Type, fn PrimitiveUnmarshalerFunc) error {
	return errors.New("not implemented")
}

func (p *unmarshalerFunc) RegisterKindOverride(k reflect.Kind, fn PrimitiveUnmarshalerFunc) error {
	return errors.New("not implemented")
}

// primitiveUnmarshalerFunc implements the Unmarshaler interface.
type primitiveUnmarshalerFunc struct {
	fn PrimitiveUnmarshalerFunc
}

func (f primitiveUnmarshalerFunc) Unmarshal(v reflect.Value, a []string, opts *UnmarshalOptions) error {
	if a == nil {
		return nil
	}
	s, err := opts.SliceToString(a)
	if err != nil {
		return err
	}
	return f.fn(v, s, opts)
}

func (p *primitiveUnmarshalerFunc) RegisterSubFactory(k reflect.Kind, fn UnmarshalerFactoryFunc) error {
	return errors.New("not implemented")
}

func (p *primitiveUnmarshalerFunc) RegisterCustomType(k reflect.Type, fn PrimitiveUnmarshalerFunc) error {
	return errors.New("not implemented")
}

func (p *primitiveUnmarshalerFunc) RegisterKindOverride(k reflect.Kind, fn PrimitiveUnmarshalerFunc) error {
	return errors.New("not implemented")
}
