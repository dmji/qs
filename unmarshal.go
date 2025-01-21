package qs

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
)

// QSUnmarshaler objects can be created by calling NewUnmarshaler and they can be
// used to unmarshal query strings or url.Values into structs or maps.
type QSUnmarshaler struct {
	opts *UnmarshalerDefaultOptions

	stringToQueryParser func(query string) (url.Values, error)
}

// NewUnmarshaler returns a new QSUnmarshaler object.
func NewUnmarshaler(prm *UnmarshalerDefaultOptions, opts ...func(p *QSUnmarshaler)) *QSUnmarshaler {
	p := &QSUnmarshaler{
		opts:                prepareUnmarshalOptions(*prm),
		stringToQueryParser: url.ParseQuery,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *QSUnmarshaler) RegisterSubFactory(k reflect.Kind, fn UnmarshalerFactoryFunc) error {
	return p.opts.UnmarshalerFactory.RegisterSubFactory(k, fn)
}

func (p *QSUnmarshaler) RegisterCustomType(k reflect.Type, fn PrimitiveUnmarshalerFunc) error {
	return p.opts.UnmarshalerFactory.RegisterCustomType(k, fn)
}

func (p *QSUnmarshaler) RegisterKindOverride(k reflect.Kind, fn PrimitiveUnmarshalerFunc) error {
	return p.opts.UnmarshalerFactory.RegisterKindOverride(k, fn)
}

// Unmarshal unmarshals an object from a query string.
// See the documentation of the global Unmarshal func.
func (p *QSUnmarshaler) Unmarshal(into interface{}, queryString string) error {
	values, err := p.stringToQueryParser(queryString)
	if err != nil {
		return fmt.Errorf("error parsing query string %q :: %v", queryString, err)
	}
	return p.UnmarshalValues(into, values)
}

// UnmarshalValues unmarshals an object from a url.Values.
// See the documentation of the global UnmarshalValues func.
func (p *QSUnmarshaler) UnmarshalValues(into interface{}, values url.Values) error {
	pv := reflect.ValueOf(into)
	if !pv.IsValid() {
		return errors.New("received an empty interface")
	}
	if pv.Kind() != reflect.Ptr {
		return fmt.Errorf("expected a pointer, got %T", into)
	}
	if pv.IsNil() {
		return fmt.Errorf("nil pointer of type %T", into)
	}
	v := pv.Elem()

	vum, err := p.opts.ValuesUnmarshalerFactory.ValuesUnmarshaler(v.Type(), p.opts)
	if err != nil {
		return err
	}
	return vum.UnmarshalValues(v, values, p.opts)
}

// CheckUnmarshal check whether the type of the given object supports
// unmarshaling from query strings.
// See the documentation of the global CheckUnmarshal func.
func (p *QSUnmarshaler) CheckUnmarshal(into interface{}) error {
	return p.CheckUnmarshalType(reflect.TypeOf(into))
}

// CheckUnmarshalType check whether the given type supports unmarshaling from
// query strings. See the documentation of the global CheckUnmarshalType func.
func (p *QSUnmarshaler) CheckUnmarshalType(t reflect.Type) error {
	if t == nil {
		return errors.New("nil type")
	}
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("expected a pointer, got %v", t)
	}
	_, err := p.opts.ValuesUnmarshalerFactory.ValuesUnmarshaler(t.Elem(), p.opts)
	return err
}
