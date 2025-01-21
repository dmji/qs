package qs

import (
	"errors"
	"reflect"
)

type ValuesUnmarshalerFactoryFunc func(t reflect.Type, opts *UnmarshalerDefaultOptions) (ValuesUnmarshaler, error)

// ValuesUnmarshalerFactory can create ValuesUnmarshaler objects for various types.
type ValuesUnmarshalerFactory interface {
	// ValuesUnmarshaler returns a ValuesUnmarshaler object for the given t
	// type and opts options.
	ValuesUnmarshaler(t reflect.Type, opts *UnmarshalerDefaultOptions) (ValuesUnmarshaler, error)

	// RegisterSubFactory registers a ValuesUnmarshalerFactory for the given kind
	RegisterSubFactory(k reflect.Kind, fn ValuesUnmarshalerFactoryFunc) error
}

type valuesUnmarshalerFactory struct {
	kindSubRegistries          map[reflect.Kind]ValuesUnmarshalerFactory
	kindSubRegistriesOverriden map[reflect.Kind]ValuesUnmarshalerFactory
}

func (p *valuesUnmarshalerFactory) ValuesUnmarshaler(t reflect.Type, opts *UnmarshalerDefaultOptions) (ValuesUnmarshaler, error) {
	if subFactory, ok := p.kindSubRegistriesOverriden[t.Kind()]; ok {
		return subFactory.ValuesUnmarshaler(t, opts)
	}

	if subFactory, ok := p.kindSubRegistries[t.Kind()]; ok {
		return subFactory.ValuesUnmarshaler(t, opts)
	}

	return nil, &UnhandledTypeError{Type: t}
}

func (p *valuesUnmarshalerFactory) RegisterSubFactory(k reflect.Kind, fn ValuesUnmarshalerFactoryFunc) error {
	p.kindSubRegistriesOverriden[k] = &valuesUnmarshalerFactoryFunc{fn}
	return nil
}

func newValuesUnmarshalerFactory() *valuesUnmarshalerFactory {
	return &valuesUnmarshalerFactory{
		kindSubRegistries: map[reflect.Kind]ValuesUnmarshalerFactory{
			reflect.Ptr:    &valuesUnmarshalerFactoryFunc{newPtrValuesUnmarshaler},
			reflect.Struct: &valuesUnmarshalerFactoryFunc{newStructUnmarshaler},
			reflect.Map:    &valuesUnmarshalerFactoryFunc{newMapUnmarshaler},
		},
	}
}

// valuesUnmarshalerFactoryFunc implements the UnmarshalerFactory interface.
type valuesUnmarshalerFactoryFunc struct {
	fn ValuesUnmarshalerFactoryFunc
}

func (f valuesUnmarshalerFactoryFunc) ValuesUnmarshaler(t reflect.Type, opts *UnmarshalerDefaultOptions) (ValuesUnmarshaler, error) {
	return f.fn(t, opts)
}

func (p *valuesUnmarshalerFactoryFunc) RegisterSubFactory(k reflect.Kind, fn ValuesUnmarshalerFactoryFunc) error {
	return errors.New("not implemented")
}
