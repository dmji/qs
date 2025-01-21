package qs

import (
	"fmt"
	"net/url"
	"reflect"
)

// ValuesMarshaler can marshal a value into a url.Values.
type ValuesMarshaler interface {
	// MarshalValues marshals the given v value using opts into a url.Values.
	MarshalValues(v reflect.Value, opts *MarshalOptions) (url.Values, error)
}

// structMarshaler implements ValuesMarshaler.
type structMarshaler struct {
	Type           reflect.Type
	EmbeddedFields []embeddedFieldMarshaler
	Fields         []*fieldMarshaler
}

type embeddedFieldMarshaler struct {
	FieldIndex      int
	ValuesMarshaler ValuesMarshaler
}

type fieldMarshaler struct {
	FieldIndex int
	Marshaler  Marshaler
	Tag        *ParsedTagInfo
}

// newStructMarshaler creates a struct marshaler for a specific struct type.
func newStructMarshaler(t reflect.Type, opts *MarshalOptions) (ValuesMarshaler, error) {
	if t.Kind() != reflect.Struct {
		return nil, &WrongKindError{Expected: reflect.Struct, Actual: t}
	}

	sm := &structMarshaler{
		Type: t,
	}

	for i, numField := 0, t.NumField(); i < numField; i++ {
		sf := t.Field(i)
		vm, fm, err := newFieldMarshaler(sf, opts)
		if err != nil {
			return nil, fmt.Errorf("error creating marshaler for field %v of struct %v :: %v",
				sf.Name, t, err)
		}
		if vm != nil {
			sm.EmbeddedFields = append(sm.EmbeddedFields, embeddedFieldMarshaler{
				FieldIndex:      i,
				ValuesMarshaler: vm,
			})
		}
		if fm != nil {
			fm.FieldIndex = i
			sm.Fields = append(sm.Fields, fm)
		}
	}

	return sm, nil
}

func newFieldMarshaler(sf reflect.StructField, opts *MarshalOptions) (ValuesMarshaler, *fieldMarshaler, error) {
	var vm ValuesMarshaler
	var fm *fieldMarshaler

	tag, err := getStructFieldInfo(sf, opts.NameTransformer, opts.TagOptionsDefaults, NewUndefinedUnmarshalTagOptions(), opts.TagCommonOptionsDefaults)
	if tag == nil || err != nil {
		return vm, fm, err
	}

	t := sf.Type
	if sf.Anonymous {
		vm, err = opts.ValuesMarshalerFactory.ValuesMarshaler(t, opts)
		if err == nil {
			// We can end up here for example in case of an embedded struct.
			return vm, fm, err
		}
	}

	m, err := opts.MarshalerFactory.Marshaler(t, opts)
	if err != nil {
		return vm, fm, err
	}
	fm = &fieldMarshaler{
		Marshaler: m,
		Tag:       tag,
	}
	return vm, fm, err
}

func (p *structMarshaler) MarshalValues(v reflect.Value, opts *MarshalOptions) (url.Values, error) {
	t := v.Type()
	if t != p.Type {
		return nil, &WrongTypeError{Actual: t, Expected: p.Type}
	}

	// TODO: use a StructError error type in the function to generate
	// error messages prefixed with the name of the struct type.

	vs := make(url.Values, len(p.Fields))

	for _, fm := range p.Fields {
		fv := v.Field(fm.FieldIndex)
		if fm.Tag.MarshalPresence == MarshalPresenceOmitEmpty && isEmpty(fv) {
			continue
		}
		a, err := fm.Marshaler.Marshal(fv, opts)
		if err != nil {
			return nil, fmt.Errorf("error marshaling url.Values entry %q :: %v", fm.Tag.Name, err)
		}
		if len(a) != 0 {
			vs[fm.Tag.Name] = a
		}
	}

	for _, ef := range p.EmbeddedFields {
		evs, err := ef.ValuesMarshaler.MarshalValues(v.Field(ef.FieldIndex), opts)
		if err != nil {
			return nil, fmt.Errorf("error marshaling embedded field %q :: %v", v.Type().Field(ef.FieldIndex).Name, err)
		}
		for k, a := range evs {
			vs[k] = a
		}
	}

	return vs, nil
}

func isEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr:
		return v.IsNil()
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0.0
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		return v.Len() == 0
	default:
		return false
	}
}

type mapMarshaler struct {
	Type          reflect.Type
	ElemMarshaler Marshaler
}

func newMapMarshaler(t reflect.Type, opts *MarshalOptions) (ValuesMarshaler, error) {
	if t.Kind() != reflect.Map {
		return nil, &WrongKindError{Expected: reflect.Map, Actual: t}
	}

	if t.Key() != stringType {
		return nil, fmt.Errorf("map key type is expected to be string: %v", t)
	}

	et := t.Elem()
	m, err := opts.MarshalerFactory.Marshaler(et, opts)
	if err != nil {
		// TODO: use a MapError error type in the function to generate
		// error messages prefixed with the name of the struct type.
		return nil, fmt.Errorf("error getting marshaler for map value type %v :: %v", et, err)
	}

	return &mapMarshaler{
		Type:          t,
		ElemMarshaler: m,
	}, nil
}

func (p *mapMarshaler) MarshalValues(v reflect.Value, opts *MarshalOptions) (url.Values, error) {
	t := v.Type()
	if t != p.Type {
		return nil, &WrongTypeError{Actual: t, Expected: p.Type}
	}

	vlen := v.Len()
	if vlen == 0 {
		return nil, nil
	}

	vs := make(url.Values, vlen)
	for _, key := range v.MapKeys() {
		val := v.MapIndex(key)
		if opts.TagOptionsDefaults.Presence == MarshalPresenceOmitEmpty && isEmpty(val) {
			continue
		}
		keyStr := key.String()
		a, err := p.ElemMarshaler.Marshal(val, opts)
		if err != nil {
			return nil, fmt.Errorf("error marshaling key %q :: %v", keyStr, err)
		}
		vs[keyStr] = a
	}
	return vs, nil
}

type ptrValuesMarshaler struct {
	Type          reflect.Type
	ElemMarshaler ValuesMarshaler
}

func newPtrValuesMarshaler(t reflect.Type, opts *MarshalOptions) (ValuesMarshaler, error) {
	if t.Kind() != reflect.Ptr {
		return nil, &WrongKindError{Expected: reflect.Ptr, Actual: t}
	}
	et := t.Elem()
	em, err := opts.ValuesMarshalerFactory.ValuesMarshaler(et, opts)
	if err != nil {
		return nil, err
	}
	return &ptrValuesMarshaler{
		Type:          t,
		ElemMarshaler: em,
	}, nil
}

func (p *ptrValuesMarshaler) MarshalValues(v reflect.Value, opts *MarshalOptions) (url.Values, error) {
	t := v.Type()
	if t != p.Type {
		return nil, &WrongTypeError{Actual: t, Expected: p.Type}
	}
	if v.IsNil() {
		return nil, nil
	}
	return p.ElemMarshaler.MarshalValues(v.Elem(), opts)
}
