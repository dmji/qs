package qs

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type ptrUnmarshaler struct {
	Type            reflect.Type
	ElemType        reflect.Type
	ElemUnmarshaler Unmarshaler
}

func newPtrUnmarshaler(t reflect.Type, opts *UnmarshalOptions) (Unmarshaler, error) {
	if t.Kind() != reflect.Ptr {
		return nil, &WrongKindError{Expected: reflect.Ptr, Actual: t}
	}
	et := t.Elem()
	eu, err := opts.UnmarshalerOptions.UnmarshalerFactory.Unmarshaler(et, opts)
	if err != nil {
		return nil, err
	}
	return &ptrUnmarshaler{
		Type:            t,
		ElemType:        et,
		ElemUnmarshaler: eu,
	}, nil
}

func (p *ptrUnmarshaler) Unmarshal(v reflect.Value, a []string, opts *UnmarshalOptions) error {
	t := v.Type()
	if t != p.Type {
		return &WrongTypeError{Actual: t, Expected: p.Type}
	}
	if v.IsNil() {
		v.Set(reflect.New(p.ElemType))
	}
	return p.ElemUnmarshaler.Unmarshal(v.Elem(), a, opts)
}

type arrayUnmarshaler struct {
	Type            reflect.Type
	ElemUnmarshaler Unmarshaler
	Len             int
}

func newArrayUnmarshaler(t reflect.Type, opts *UnmarshalOptions) (Unmarshaler, error) {
	if t.Kind() != reflect.Array {
		return nil, &WrongKindError{Expected: reflect.Array, Actual: t}
	}

	eu, err := opts.UnmarshalerOptions.UnmarshalerFactory.Unmarshaler(t.Elem(), opts)
	if err != nil {
		return nil, err
	}
	return &arrayUnmarshaler{
		Type:            t,
		ElemUnmarshaler: eu,
		Len:             t.Len(),
	}, nil
}

func (p *arrayUnmarshaler) Unmarshal(v reflect.Value, a []string, opts *UnmarshalOptions) error {
	t := v.Type()
	if t != p.Type {
		return &WrongTypeError{Actual: t, Expected: p.Type}
	}

	if a == nil {
		return nil
	}
	if len(a) != p.Len {
		return fmt.Errorf("array length == %v, want %v", len(a), p.Len)
	}
	for i := range a {
		err := p.ElemUnmarshaler.Unmarshal(v.Index(i), a[i:i+1], opts)
		if err != nil {
			return fmt.Errorf("error unmarshaling array index %v :: %v", i, err)
		}
	}
	return nil
}

type sliceUnmarshaler struct {
	Type            reflect.Type
	ElemUnmarshaler Unmarshaler
}

func newSliceUnmarshaler(t reflect.Type, opts *UnmarshalOptions) (Unmarshaler, error) {
	if t.Kind() != reflect.Slice {
		return nil, &WrongKindError{Expected: reflect.Slice, Actual: t}
	}

	eu, err := opts.UnmarshalerOptions.UnmarshalerFactory.Unmarshaler(t.Elem(), opts)
	if err != nil {
		return nil, err
	}
	return &sliceUnmarshaler{
		Type:            t,
		ElemUnmarshaler: eu,
	}, nil
}

func splitArrayBySeparatorWithSameOrder(a []string, separatorType OptionSliceSeparator) []string {
	sep := ""
	switch separatorType {
	case OptionSliceSeparatorComma:
		sep = ","
	case OptionSliceSeparatorSemicolon:
		sep = ";"
	case OptionSliceSeparatorSpace:
		sep = " "
	case OptionSliceSeparatorNone:
	default:
		panic(fmt.Sprintf("unexpected qs.OptionSliceSeparator: %#v", separatorType))
	}
	if len(sep) == 0 {
		return a
	}

	vals := make([]string, 0, 2*len(a))
	for _, s := range a {
		vals = append(vals, strings.Split(s, sep)...)
	}
	return vals
}

func (p *sliceUnmarshaler) Unmarshal(v reflect.Value, a []string, opts *UnmarshalOptions) error {
	t := v.Type()
	if t != p.Type {
		return &WrongTypeError{Actual: t, Expected: p.Type}
	}

	vals := splitArrayBySeparatorWithSameOrder(a, opts.ParsedTagInfo.CommonOpts.SliceSeparator)

	// resize or create slice
	n := 0
	if v.IsNil() {
		v.Set(reflect.MakeSlice(t, len(vals), len(vals)))
	} else {
		keepPrevValues := opts.ParsedTagInfo.UnmarshalOpts.SliceValues == UnmarshalSliceValuesKeepOld

		oldLen := v.Len()
		newLen := len(vals)
		if keepPrevValues {
			n = oldLen
			newLen += oldLen
		}

		s := reflect.MakeSlice(t, newLen, newLen)
		if keepPrevValues {
			for i := 0; i < oldLen; i++ {
				s.Index(i).Set(v.Index(i))
			}
		}
		v.Set(s)
	}

	breakOnError := opts.ParsedTagInfo.UnmarshalOpts.SliceUnexpectedValue == UnmarshalSliceUnexpectedValueBreakWithError

	// unmarshal elements of slice
	var errLoop error
	for i := range vals {
		err := p.ElemUnmarshaler.Unmarshal(v.Index(n), vals[i:i+1], opts)
		if err == nil {
			n++
			continue
		}

		if breakOnError {
			errLoop = fmt.Errorf("error unmarshaling slice index %v :: %v", i, err)
			break
		}
	}

	// cut unmarshleable values from slice or clear if error occurred
	if errLoop != nil {
		v.Set(v.Slice(0, 0))
		return errLoop
	}

	v.Set(v.Slice(0, n))
	return nil
}

// unmarshalString can unmarshal an ini file entry into a value with an
// underlying type (kind) of string.
func unmarshalString(v reflect.Value, s string, opts *UnmarshalOptions) error {
	if v.Kind() != reflect.String {
		return &WrongKindError{Expected: reflect.String, Actual: v.Type()}
	}
	v.SetString(s)
	return nil
}

// unmarshalBool can unmarshal an ini file entry into a value with an
// underlying type (kind) of bool.
func unmarshalBool(v reflect.Value, s string, opts *UnmarshalOptions) error {
	if v.Kind() != reflect.Bool {
		return &WrongKindError{Expected: reflect.Bool, Actual: v.Type()}
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	v.SetBool(b)
	return nil
}

// unmarshalInt can unmarshal an ini file entry into a signed integer value
// with an underlying type (kind) of int, int8, int16, int32 or int64.
func unmarshalInt(v reflect.Value, s string, opts *UnmarshalOptions) error {
	var bitSize int

	switch v.Kind() {
	case reflect.Int:
	case reflect.Int8:
		bitSize = 8
	case reflect.Int16:
		bitSize = 16
	case reflect.Int32:
		bitSize = 32
	case reflect.Int64:
		bitSize = 64
	default:
		return &WrongKindError{Expected: reflect.Int, Actual: v.Type()}
	}

	i, err := strconv.ParseInt(s, 0, bitSize)
	if err != nil {
		return err
	}

	v.SetInt(i)
	return nil
}

// unmarshalUint can unmarshal an ini file entry into an unsigned integer value
// with an underlying type (kind) of uint, uint8, uint16, uint32 or uint64.
func unmarshalUint(v reflect.Value, s string, opts *UnmarshalOptions) error {
	var bitSize int

	switch v.Kind() {
	case reflect.Uint:
	case reflect.Uint8:
		bitSize = 8
	case reflect.Uint16:
		bitSize = 16
	case reflect.Uint32:
		bitSize = 32
	case reflect.Uint64:
		bitSize = 64
	default:
		return &WrongKindError{Expected: reflect.Uint, Actual: v.Type()}
	}

	i, err := strconv.ParseUint(s, 0, bitSize)
	if err != nil {
		return err
	}

	v.SetUint(i)
	return nil
}

func unmarshalFloat(v reflect.Value, s string, opts *UnmarshalOptions) error {
	var bitSize int

	switch v.Kind() {
	case reflect.Float32:
		bitSize = 32
	case reflect.Float64:
		bitSize = 64
	default:
		return &WrongKindError{Expected: reflect.Float32, Actual: v.Type()}
	}

	f, err := strconv.ParseFloat(s, bitSize)
	if err != nil {
		return err
	}

	v.SetFloat(f)
	return nil
}

func unmarshalTime(v reflect.Value, s string, opts *UnmarshalOptions) error {
	t := v.Type()
	if t != timeType {
		return &WrongTypeError{Actual: t, Expected: timeType}
	}

	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(tm))
	return nil
}

func unmarshalURL(v reflect.Value, s string, opts *UnmarshalOptions) error {
	t := v.Type()
	if t != urlType {
		return &WrongTypeError{Actual: t, Expected: urlType}
	}

	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(*u))
	return nil
}

func unmarshalWithUnmarshalQS(v reflect.Value, a []string, opts *UnmarshalOptions) error {
	if !v.CanAddr() {
		return fmt.Errorf("expected and addressable value, got %v", v)
	}
	unmarshalQS, ok := v.Addr().Interface().(UnmarshalQS)
	if !ok {
		return fmt.Errorf("expected a type that implements UnmarshalQS, got %v", v.Type())
	}
	return unmarshalQS.UnmarshalQS(a, opts)
}
