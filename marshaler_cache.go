package qs

import (
	"reflect"
	"sync"
)

func newValuesMarshalerCache(wrapped ValuesMarshalerFactory) ValuesMarshalerFactory {
	return &valuesMarshalerCache{
		wrapped: wrapped,
	}
}

type valuesMarshalerCache struct {
	wrapped ValuesMarshalerFactory
	cache   sync.Map
}

func (o *valuesMarshalerCache) ValuesMarshaler(t reflect.Type, opts *MarshalOptions) (ValuesMarshaler, error) {
	return cacher(o.wrapped.ValuesMarshaler, &o.cache, t, opts)
}

func (p *valuesMarshalerCache) RegisterSubFactory(k reflect.Kind, fn ValuesMarshalerFactoryFunc) error {
	return p.wrapped.RegisterSubFactory(k, fn)
}

func newMarshalerCache(wrapped MarshalerFactory) MarshalerFactory {
	return &marshalerCache{
		wrapped: wrapped,
	}
}

type marshalerCache struct {
	wrapped MarshalerFactory
	cache   sync.Map
}

func (o *marshalerCache) Marshaler(t reflect.Type, opts *MarshalOptions) (Marshaler, error) {
	return cacher(o.wrapped.Marshaler, &o.cache, t, opts)
}

func (p *marshalerCache) RegisterSubFactory(k reflect.Kind, fn MarshalerFactoryFunc) error {
	return p.wrapped.RegisterSubFactory(k, fn)
}

func (p *marshalerCache) RegisterCustomType(k reflect.Type, fn PrimitiveMarshalerFunc) error {
	return p.wrapped.RegisterCustomType(k, fn)

}

func (p *marshalerCache) RegisterKindOverride(k reflect.Kind, fn PrimitiveMarshalerFunc) error {
	return p.wrapped.RegisterKindOverride(k, fn)
}
