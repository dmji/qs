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
