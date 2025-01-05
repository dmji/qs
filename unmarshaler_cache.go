package qs

import (
	"reflect"
	"sync"
)

func newValuesUnmarshalerCache(wrapped ValuesUnmarshalerFactory) ValuesUnmarshalerFactory {
	return &valuesUnmarshalerCache{
		wrapped: wrapped,
	}
}

type valuesUnmarshalerCache struct {
	wrapped ValuesUnmarshalerFactory
	cache   sync.Map
}

func (o *valuesUnmarshalerCache) ValuesUnmarshaler(t reflect.Type, opts *UnmarshalOptions) (ValuesUnmarshaler, error) {
	return cacher(o.wrapped.ValuesUnmarshaler, &o.cache, t, opts)
}

func (p *valuesUnmarshalerCache) RegisterSubFactory(k reflect.Kind, fn ValuesUnmarshalerFactoryFunc) error {
	return p.wrapped.RegisterSubFactory(k, fn)
}

func newUnmarshalerCache(wrapped UnmarshalerFactory) UnmarshalerFactory {
	return &unmarshalerCache{
		wrapped: wrapped,
	}
}

type unmarshalerCache struct {
	wrapped UnmarshalerFactory
	cache   sync.Map
}

func (o *unmarshalerCache) Unmarshaler(t reflect.Type, opts *UnmarshalOptions) (Unmarshaler, error) {
	return cacher(o.wrapped.Unmarshaler, &o.cache, t, opts)
}

func (p *unmarshalerCache) RegisterSubFactory(k reflect.Kind, fn UnmarshalerFactoryFunc) error {
	return p.wrapped.RegisterSubFactory(k, fn)

}

func (p *unmarshalerCache) RegisterCustomType(k reflect.Type, fn PrimitiveUnmarshalerFunc) error {
	return p.wrapped.RegisterCustomType(k, fn)

}

func (p *unmarshalerCache) RegisterKindOverride(k reflect.Kind, fn PrimitiveUnmarshalerFunc) error {
	return p.wrapped.RegisterKindOverride(k, fn)
}
