package qs

import (
	"errors"
	"reflect"
)

type ValuesMarshalerFactoryFunc func(t reflect.Type, opts *MarshalOptions) (ValuesMarshaler, error)

// ValuesMarshalerFactory can create ValuesMarshaler objects for various types.
type ValuesMarshalerFactory interface {
	// ValuesMarshaler returns a ValuesMarshaler object for the given t type and
	// opts options.
	ValuesMarshaler(t reflect.Type, opts *MarshalOptions) (ValuesMarshaler, error)

	// RegisterSubFactory registers a ValuesUnmarshalerFactory for the given kind
	RegisterSubFactory(k reflect.Kind, fn ValuesMarshalerFactoryFunc) error
}

// valuesMarshalerFactory implements the ValuesMarshalerFactory interface.
type valuesMarshalerFactory struct {
	kindSubRegistries          map[reflect.Kind]ValuesMarshalerFactory
	kindSubRegistriesOverriden map[reflect.Kind]ValuesMarshalerFactory
}

func (p *valuesMarshalerFactory) ValuesMarshaler(t reflect.Type, opts *MarshalOptions) (ValuesMarshaler, error) {
	if subFactory, ok := p.kindSubRegistriesOverriden[t.Kind()]; ok {
		return subFactory.ValuesMarshaler(t, opts)
	}

	if subFactory, ok := p.kindSubRegistries[t.Kind()]; ok {
		return subFactory.ValuesMarshaler(t, opts)
	}

	return nil, &unhandledTypeError{Type: t}
}

func (p *valuesMarshalerFactory) RegisterSubFactory(k reflect.Kind, fn ValuesMarshalerFactoryFunc) error {
	p.kindSubRegistriesOverriden[k] = &valuesMarshalerFactoryFunc{fn}
	return nil
}

func newValuesMarshalerFactory() *valuesMarshalerFactory {
	return &valuesMarshalerFactory{
		kindSubRegistries: map[reflect.Kind]ValuesMarshalerFactory{
			reflect.Ptr:    &valuesMarshalerFactoryFunc{newPtrValuesMarshaler},
			reflect.Struct: &valuesMarshalerFactoryFunc{newStructMarshaler},
			reflect.Map:    &valuesMarshalerFactoryFunc{newMapMarshaler},
		},
	}
}

// valuesMarshalerFactoryFunc implements the ValuesMarshalerFactory interface.
type valuesMarshalerFactoryFunc struct {
	fn ValuesMarshalerFactoryFunc
}

func (f valuesMarshalerFactoryFunc) ValuesMarshaler(t reflect.Type, opts *MarshalOptions) (ValuesMarshaler, error) {
	return f.fn(t, opts)
}

func (p *valuesMarshalerFactoryFunc) RegisterSubFactory(k reflect.Kind, fn ValuesMarshalerFactoryFunc) error {
	return errors.New("not implemented")
}
