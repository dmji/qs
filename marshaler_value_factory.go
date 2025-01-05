package qs

import (
	"net/url"
	"reflect"
)

// ValuesMarshaler can marshal a value into a url.Values.
type ValuesMarshaler interface {
	// MarshalValues marshals the given v value using opts into a url.Values.
	MarshalValues(v reflect.Value, opts *MarshalOptions) (url.Values, error)
}

// ValuesMarshalerFactory can create ValuesMarshaler objects for various types.
type ValuesMarshalerFactory interface {
	// ValuesMarshaler returns a ValuesMarshaler object for the given t type and
	// opts options.
	ValuesMarshaler(t reflect.Type, opts *MarshalOptions) (ValuesMarshaler, error)
}

// valuesMarshalerFactory implements the ValuesMarshalerFactory interface.
type valuesMarshalerFactory struct {
	KindSubRegistries map[reflect.Kind]ValuesMarshalerFactory
}

func (p *valuesMarshalerFactory) ValuesMarshaler(t reflect.Type, opts *MarshalOptions) (ValuesMarshaler, error) {
	if subFactory, ok := p.KindSubRegistries[t.Kind()]; ok {
		return subFactory.ValuesMarshaler(t, opts)
	}

	return nil, &unhandledTypeError{Type: t}
}

func newValuesMarshalerFactory() ValuesMarshalerFactory {
	return &valuesMarshalerFactory{
		KindSubRegistries: map[reflect.Kind]ValuesMarshalerFactory{
			reflect.Ptr:    valuesMarshalerFactoryFunc(newPtrValuesMarshaler),
			reflect.Struct: valuesMarshalerFactoryFunc(newStructMarshaler),
			reflect.Map:    valuesMarshalerFactoryFunc(newMapMarshaler),
		},
	}
}

// valuesMarshalerFactoryFunc implements the ValuesMarshalerFactory interface.
type valuesMarshalerFactoryFunc func(t reflect.Type, opts *MarshalOptions) (ValuesMarshaler, error)

func (f valuesMarshalerFactoryFunc) ValuesMarshaler(t reflect.Type, opts *MarshalOptions) (ValuesMarshaler, error) {
	return f(t, opts)
}
