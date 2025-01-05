package qs

import (
	"net/url"
	"reflect"
)

// ValuesUnmarshaler can unmarshal a url.Values into a value.
type ValuesUnmarshaler interface {
	// UnmarshalValues unmarshals the given url.Values using opts into v.
	UnmarshalValues(v reflect.Value, vs url.Values, opts *UnmarshalOptions) error
}

// ValuesUnmarshalerFactory can create ValuesUnmarshaler objects for various types.
type ValuesUnmarshalerFactory interface {
	// ValuesUnmarshaler returns a ValuesUnmarshaler object for the given t
	// type and opts options.
	ValuesUnmarshaler(t reflect.Type, opts *UnmarshalOptions) (ValuesUnmarshaler, error)
}

type valuesUnmarshalerFactory struct {
	KindSubRegistries map[reflect.Kind]ValuesUnmarshalerFactory
}

func (p *valuesUnmarshalerFactory) ValuesUnmarshaler(t reflect.Type, opts *UnmarshalOptions) (ValuesUnmarshaler, error) {
	if subFactory, ok := p.KindSubRegistries[t.Kind()]; ok {
		return subFactory.ValuesUnmarshaler(t, opts)
	}

	return nil, &unhandledTypeError{Type: t}
}

func newValuesUnmarshalerFactory() ValuesUnmarshalerFactory {
	return &valuesUnmarshalerFactory{
		KindSubRegistries: map[reflect.Kind]ValuesUnmarshalerFactory{
			reflect.Ptr:    valuesUnmarshalerFactoryFunc(newPtrValuesUnmarshaler),
			reflect.Struct: valuesUnmarshalerFactoryFunc(newStructUnmarshaler),
			reflect.Map:    valuesUnmarshalerFactoryFunc(newMapUnmarshaler),
		},
	}
}

// valuesUnmarshalerFactoryFunc implements the UnmarshalerFactory interface.
type valuesUnmarshalerFactoryFunc func(t reflect.Type, opts *UnmarshalOptions) (ValuesUnmarshaler, error)

func (f valuesUnmarshalerFactoryFunc) ValuesUnmarshaler(t reflect.Type, opts *UnmarshalOptions) (ValuesUnmarshaler, error) {
	return f(t, opts)
}
