package qs

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
)

// QSMarshaler objects can be created by calling NewMarshaler and they can be
// used to marshal structs or maps into query strings or url.Values.
type QSMarshaler struct {
	opts *MarshalOptions

	_EncodeValues func(values url.Values) string
}

// NewMarshaler returns a new QSMarshaler object.
func NewMarshaler(prm *MarshalOptions, opts ...func(*QSMarshaler)) *QSMarshaler {
	p := &QSMarshaler{
		opts:          prepareMarshalOptions(*prm),
		_EncodeValues: func(values url.Values) string { return values.Encode() },
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *QSMarshaler) RegisterSubFactory(k reflect.Kind, fn MarshalerFactoryFunc) error {
	return p.opts.MarshalerFactory.RegisterSubFactory(k, fn)

}

func (p *QSMarshaler) RegisterCustomType(k reflect.Type, fn PrimitiveMarshalerFunc) error {
	return p.opts.MarshalerFactory.RegisterCustomType(k, fn)

}

func (p *QSMarshaler) RegisterKindOverride(k reflect.Kind, fn PrimitiveMarshalerFunc) error {
	return p.opts.MarshalerFactory.RegisterKindOverride(k, fn)
}

// Marshal marshals a given object into a query string.
// See the documentation of the global Marshal func.
func (p *QSMarshaler) Marshal(i interface{}) (string, error) {
	values, err := p.MarshalValues(i)
	if err != nil {
		return "", err
	}
	return p._EncodeValues(values), nil
}

// MarshalValues marshals a given object into a url.Values.
// See the documentation of the global MarshalValues func.
func (p *QSMarshaler) MarshalValues(i interface{}) (url.Values, error) {
	v := reflect.ValueOf(i)
	if !v.IsValid() {
		return nil, errors.New("received an empty interface")
	}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, fmt.Errorf("nil pointer of type %T", i)
		}
		v = v.Elem()
	}

	vum, err := p.opts.ValuesMarshalerFactory.ValuesMarshaler(v.Type(), p.opts)
	if err != nil {
		return nil, err
	}
	return vum.MarshalValues(v, p.opts)
}

// CheckMarshal check whether the type of the given object supports
// marshaling into query strings.
// See the documentation of the global CheckMarshal func.
func (p *QSMarshaler) CheckMarshal(i interface{}) error {
	return p.CheckMarshalType(reflect.TypeOf(i))
}

// CheckMarshalType check whether the given type supports marshaling into
// query strings. See the documentation of the global CheckMarshalType func.
func (p *QSMarshaler) CheckMarshalType(t reflect.Type) error {
	if t == nil {
		return errors.New("nil type")
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	_, err := p.opts.ValuesMarshalerFactory.ValuesMarshaler(t, p.opts)
	return err
}
