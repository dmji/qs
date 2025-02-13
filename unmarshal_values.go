package qs

import (
	"fmt"
	"net/url"
	"reflect"
)

// ValuesUnmarshaler can unmarshal a url.Values into a value.
type ValuesUnmarshaler interface {
	// UnmarshalValues unmarshals the given url.Values using opts into v.
	UnmarshalValues(v reflect.Value, vs url.Values, opts *UnmarshalerDefaultOptions) error
}

// structUnmarshaler implements ValuesUnmarshaler.
type structUnmarshaler struct {
	Type           reflect.Type
	EmbeddedFields []embeddedFieldUnmarshaler
	Fields         []*fieldUnmarshaler
}

type embeddedFieldUnmarshaler struct {
	FieldIndex        int
	ValuesUnmarshaler ValuesUnmarshaler
}

type fieldUnmarshaler struct {
	FieldIndex  int
	Unmarshaler Unmarshaler
	Tag         *ParsedTagInfo
}

// newStructUnmarshaler creates a struct unmarshaler for a specific struct type.
func newStructUnmarshaler(t reflect.Type, opts *UnmarshalerDefaultOptions) (ValuesUnmarshaler, error) {
	if t.Kind() != reflect.Struct {
		return nil, &WrongKindError{Expected: reflect.Struct, Actual: t}
	}

	su := &structUnmarshaler{
		Type: t,
	}

	for i, numField := 0, t.NumField(); i < numField; i++ {
		sf := t.Field(i)
		vum, fum, err := newFieldUnmarshaler(sf, opts)
		if err != nil {
			return nil, fmt.Errorf("error creating unmarshaler for field %v of struct %v :: %v",
				sf.Name, t, err)
		}
		if vum != nil {
			su.EmbeddedFields = append(su.EmbeddedFields, embeddedFieldUnmarshaler{
				FieldIndex:        i,
				ValuesUnmarshaler: vum,
			})
		}
		if fum != nil {
			fum.FieldIndex = i
			su.Fields = append(su.Fields, fum)
		}
	}

	return su, nil
}

func newFieldUnmarshaler(sf reflect.StructField, opts *UnmarshalerDefaultOptions) (ValuesUnmarshaler, *fieldUnmarshaler, error) {
	var vum ValuesUnmarshaler
	var fum *fieldUnmarshaler

	tag, err := getStructFieldInfo(sf, opts.NameTransformer, NewUndefinedMarshalTagOptions(), opts.TagOptionsDefaults, opts.TagCommonOptionsDefaults)
	if tag == nil || err != nil {
		return vum, fum, err
	}

	t := sf.Type
	if sf.Anonymous {
		vum, err = opts.ValuesUnmarshalerFactory.ValuesUnmarshaler(t, opts)
		if err == nil {
			// We can end up here for example in case of an embedded struct.
			return vum, fum, err
		}
	}

	um, err := opts.UnmarshalerFactory.Unmarshaler(t, NewUnmarshalOptions(opts, nil))
	if err != nil {
		return vum, fum, err
	}
	fum = &fieldUnmarshaler{
		Unmarshaler: um,
		Tag:         tag,
	}
	return vum, fum, err
}

func (p *structUnmarshaler) UnmarshalValues(v reflect.Value, vs url.Values, opts *UnmarshalerDefaultOptions) error {
	t := v.Type()
	if t != p.Type {
		return &WrongTypeError{Actual: t, Expected: p.Type}
	}

	// TODO: use a StructError error type in the function to generate
	// error messages prefixed with the name of the struct type.

	for _, fum := range p.Fields {
		a, ok := vs[fum.Tag.Name]
		if !ok {
			switch fum.Tag.UnmarshalOpts.Presence {
			case UnmarshalPresenceNil:
				continue
			case UnmarshalPresenceReq:
				return &ReqError{
					Message:   fmt.Sprintf("missing required field %q in struct %v", fum.Tag.Name, t),
					FieldName: fum.Tag.Name,
				}
			}
		}
		err := fum.Unmarshaler.Unmarshal(v.Field(fum.FieldIndex), a, NewUnmarshalOptions(opts, fum.Tag))
		if err != nil {
			return fmt.Errorf("error unmarshaling url.Values entry %q :: %v", fum.Tag.Name, err)
		}
	}

	for _, ef := range p.EmbeddedFields {
		err := ef.ValuesUnmarshaler.UnmarshalValues(v.Field(ef.FieldIndex), vs, opts)
		if err != nil {
			if _, ok := IsRequiredFieldError(err); ok {
				name := t.Field(ef.FieldIndex).Name
				return &ReqError{
					Message:   fmt.Sprintf("embedded field %q :: %v", name, err),
					FieldName: name,
				}
			}
			return fmt.Errorf("error unmarshaling embedded field %q :: %v", t.Field(ef.FieldIndex).Name, err)
		}
	}

	return nil
}

type mapUnmarshaler struct {
	Type            reflect.Type
	ElemType        reflect.Type
	ElemUnmarshaler Unmarshaler
}

func newMapUnmarshaler(t reflect.Type, opts *UnmarshalerDefaultOptions) (ValuesUnmarshaler, error) {
	if t.Kind() != reflect.Map {
		return nil, &WrongKindError{Expected: reflect.Map, Actual: t}
	}

	if t.Key() != stringType {
		return nil, fmt.Errorf("map key type is expected to be string: %v", t)
	}

	et := t.Elem()
	um, err := opts.UnmarshalerFactory.Unmarshaler(et, NewUnmarshalOptions(opts, nil))
	if err != nil {
		// TODO: use a MapError error type in the function to generate
		// error messages prefixed with the name of the struct type.
		return nil, fmt.Errorf("error getting unmarshaler for map value type %v :: %v", et, err)
	}

	return &mapUnmarshaler{
		Type:            t,
		ElemType:        et,
		ElemUnmarshaler: um,
	}, nil
}

func (p *mapUnmarshaler) UnmarshalValues(v reflect.Value, vs url.Values, opts *UnmarshalerDefaultOptions) error {
	t := v.Type()
	if t != p.Type {
		return &WrongTypeError{Actual: t, Expected: p.Type}
	}

	if v.IsNil() {
		v.Set(reflect.MakeMap(t))
	}

	for k, a := range vs {
		item := reflect.New(p.ElemType).Elem()
		err := p.ElemUnmarshaler.Unmarshal(item, a, NewUnmarshalOptions(opts, nil))
		if err != nil {
			return fmt.Errorf("error unmarshaling key %q :: %v", k, err)
		}
		v.SetMapIndex(reflect.ValueOf(k), item)
	}

	return nil
}

type ptrValuesUnmarshaler struct {
	Type            reflect.Type
	ElemType        reflect.Type
	ElemUnmarshaler ValuesUnmarshaler
}

func newPtrValuesUnmarshaler(t reflect.Type, opts *UnmarshalerDefaultOptions) (ValuesUnmarshaler, error) {
	if t.Kind() != reflect.Ptr {
		return nil, &WrongKindError{Expected: reflect.Ptr, Actual: t}
	}
	et := t.Elem()
	eu, err := opts.ValuesUnmarshalerFactory.ValuesUnmarshaler(et, opts)
	if err != nil {
		return nil, err
	}
	return &ptrValuesUnmarshaler{
		Type:            t,
		ElemType:        et,
		ElemUnmarshaler: eu,
	}, nil
}

func (p *ptrValuesUnmarshaler) UnmarshalValues(v reflect.Value, vs url.Values, opts *UnmarshalerDefaultOptions) error {
	t := v.Type()
	if t != p.Type {
		return &WrongTypeError{Actual: t, Expected: p.Type}
	}
	if v.IsNil() {
		v.Set(reflect.New(p.ElemType))
	}
	return p.ElemUnmarshaler.UnmarshalValues(v.Elem(), vs, opts)
}
