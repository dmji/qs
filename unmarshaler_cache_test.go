package qs

import (
	"errors"
	"reflect"
	"testing"
)

type fakeValuesUnmarshalerFactory struct {
	u     ValuesUnmarshaler
	err   error
	calls []reflect.Type
}

func (o *fakeValuesUnmarshalerFactory) ValuesUnmarshaler(t reflect.Type, opts *UnmarshalOptions) (ValuesUnmarshaler, error) {
	o.calls = append(o.calls, t)
	return o.u, o.err
}

func (p *fakeValuesUnmarshalerFactory) RegisterSubFactory(k reflect.Kind, fn ValuesUnmarshalerFactoryFunc) error {
	panic("!mock not implemented!")
}

func TestValuesUnmarshalerCacheSuccess(t *testing.T) {
	expected := &structUnmarshaler{}
	wrapped := &fakeValuesUnmarshalerFactory{u: expected}
	cache := newValuesUnmarshalerCache(wrapped)
	tp := reflect.TypeOf((*fakeValuesUnmarshalerFactory)(nil)).Elem()

	// cache miss
	u, err := cache.ValuesUnmarshaler(tp, nil)
	if err != nil {
		t.Fatal(err)
	}
	if u != expected {
		t.Fatalf("got %v, want %v", u, expected)
	}
	if len(wrapped.calls) != 1 || wrapped.calls[0] != tp {
		t.Fatalf("got %v, want %v", wrapped.calls, []reflect.Type{tp})
	}

	// cache hit
	u, err = cache.ValuesUnmarshaler(tp, nil)
	if err != nil {
		t.Fatal(err)
	}
	if u != expected {
		t.Fatalf("got %v, want %v", u, expected)
	}
	if len(wrapped.calls) != 1 || wrapped.calls[0] != tp {
		t.Fatalf("got %v, want %v", wrapped.calls, []reflect.Type{tp})
	}
}

func TestValuesUnmarshalerCacheError(t *testing.T) {
	e := errors.New("test error")
	wrapped := &fakeValuesUnmarshalerFactory{err: e}
	cache := newValuesUnmarshalerCache(wrapped)
	tp := reflect.TypeOf((*fakeValuesUnmarshalerFactory)(nil)).Elem()

	// cache miss
	_, err := cache.ValuesUnmarshaler(tp, nil)
	if err != e {
		t.Fatalf("got %q, want %q", err, e)
	}
	if len(wrapped.calls) != 1 || wrapped.calls[0] != tp {
		t.Fatalf("got %v, want %v", wrapped.calls, []reflect.Type{tp})
	}

	// cache hit
	_, err = cache.ValuesUnmarshaler(tp, nil)
	if err != e {
		t.Fatalf("got %q, want %q", err, e)
	}
	if len(wrapped.calls) != 1 || wrapped.calls[0] != tp {
		t.Fatalf("got %v, want %v", wrapped.calls, []reflect.Type{tp})
	}
}

type fakeUnmarshalerFactory struct {
	u     Unmarshaler
	err   error
	calls []reflect.Type
}

func (o *fakeUnmarshalerFactory) Unmarshaler(t reflect.Type, opts *UnmarshalOptions) (Unmarshaler, error) {
	o.calls = append(o.calls, t)
	return o.u, o.err
}

func (p *fakeUnmarshalerFactory) RegisterSubFactory(k reflect.Kind, fn UnmarshalerFactoryFunc) error {
	panic("!mock not implemented!")
}

func (p *fakeUnmarshalerFactory) RegisterCustomType(k reflect.Type, fn PrimitiveUnmarshalerFunc) error {
	panic("!mock not implemented!")
}

func (p *fakeUnmarshalerFactory) RegisterKindOverride(k reflect.Kind, fn PrimitiveUnmarshalerFunc) error {
	panic("!mock not implemented!")
}

type fakeUnmarshaler struct{}

func (o *fakeUnmarshaler) Unmarshal(v reflect.Value, a []string, opts *UnmarshalOptions) error {
	return nil
}

func TestUnmarshalerCacheSuccess(t *testing.T) {
	// we need a comparable fakeUnmarshaler object to be able to assert
	expected := &fakeUnmarshaler{}
	wrapped := &fakeUnmarshalerFactory{u: expected}
	cache := newUnmarshalerCache(wrapped)
	tp := reflect.TypeOf((*fakeUnmarshalerFactory)(nil)).Elem()

	// cache miss
	u, err := cache.Unmarshaler(tp, nil)
	if err != nil {
		t.Fatal(err)
	}
	if u != expected {
		t.Fatalf("got %v, want %v", u, expected)
	}
	if len(wrapped.calls) != 1 || wrapped.calls[0] != tp {
		t.Fatalf("got %v, want %v", wrapped.calls, []reflect.Type{tp})
	}

	// cache hit
	u, err = cache.Unmarshaler(tp, nil)
	if err != nil {
		t.Fatal(err)
	}
	if u != expected {
		t.Fatalf("got %v, want %v", u, expected)
	}
	if len(wrapped.calls) != 1 || wrapped.calls[0] != tp {
		t.Fatalf("got %v, want %v", wrapped.calls, []reflect.Type{tp})
	}
}

func TestUnmarshalerCacheError(t *testing.T) {
	e := errors.New("test error")
	wrapped := &fakeUnmarshalerFactory{err: e}
	cache := newUnmarshalerCache(wrapped)
	tp := reflect.TypeOf((*fakeUnmarshalerFactory)(nil)).Elem()

	// cache miss
	_, err := cache.Unmarshaler(tp, nil)
	if err != e {
		t.Fatalf("got %q, want %q", err, e)
	}
	if len(wrapped.calls) != 1 || wrapped.calls[0] != tp {
		t.Fatalf("got %v, want %v", wrapped.calls, []reflect.Type{tp})
	}

	// cache hit
	_, err = cache.Unmarshaler(tp, nil)
	if err != e {
		t.Fatalf("got %q, want %q", err, e)
	}
	if len(wrapped.calls) != 1 || wrapped.calls[0] != tp {
		t.Fatalf("got %v, want %v", wrapped.calls, []reflect.Type{tp})
	}
}
