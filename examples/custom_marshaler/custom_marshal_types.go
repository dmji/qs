package custom_marshaler

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"time"

	"github.com/dmji/qs"
)

var (
	// for byteSliceMarshaler
	byteSliceType = reflect.TypeOf([]byte(nil))

	// for durationMarshaler
	durationType = reflect.TypeOf(time.Duration(0))
)

// durationMarshaler implements the Marshaler and Unmarshaler interfaces to
// provide custom marshaling and unmarshaling for the time.Duration type.
type durationMarshaler struct{}

func (o *durationMarshaler) Marshal(v reflect.Value, opts *qs.MarshalOptions) ([]string, error) {
	return []string{v.Interface().(time.Duration).String()}, nil
}

func (o *durationMarshaler) Unmarshal(v reflect.Value, a []string, opts *qs.UnmarshalOptions) error {
	s, err := opts.SliceToString(a)
	if err != nil {
		return err
	}
	t, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("unsupported time format: %v", s)
	}
	v.Set(reflect.ValueOf(t))
	return nil
}

// byteSliceMarshaler implements the Marshaler and Unmarshaler interfaces to
// provide custom marshaling and unmarshaling for the []byte type.
type byteSliceMarshaler struct{}

func (byteSliceMarshaler) Marshal(v reflect.Value, opts *qs.MarshalOptions) ([]string, error) {
	return []string{hex.EncodeToString(v.Interface().([]byte))}, nil
}

func (byteSliceMarshaler) Unmarshal(v reflect.Value, a []string, opts *qs.UnmarshalOptions) error {
	s, err := opts.SliceToString(a)
	if err != nil {
		return err
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(b))
	return nil
}
