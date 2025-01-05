package custom_marshaler

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"time"

	"github.com/pasztorpisti/qs"
)

var (
	// for byteSliceMarshaler
	byteSliceType = reflect.TypeOf([]byte(nil))

	// for durationMarshaler
	durationType = reflect.TypeOf(time.Duration(0))
)

// durationMarshaler implements the Marshaler and Unmarshaler interfaces to
// provide custom marshaling and unmarshaling for the time.Duration type.
func durationMarshal(v reflect.Value, opts *qs.MarshalOptions) (string, error) {
	return v.Interface().(time.Duration).String(), nil
}

func durationUnmarshal(v reflect.Value, s string, opts *qs.UnmarshalOptions) error {
	t, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("unsupported time format: %v", s)
	}
	v.Set(reflect.ValueOf(t))
	return nil
}

// byteSliceMarshaler implements the Marshaler and Unmarshaler interfaces to
// provide custom marshaling and unmarshaling for the []byte type.
func byteSliceMarshal(v reflect.Value, opts *qs.MarshalOptions) (string, error) {
	return hex.EncodeToString(v.Interface().([]byte)), nil
}

func byteSliceUnmarshal(v reflect.Value, s string, opts *qs.UnmarshalOptions) error {
	b, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(b))
	return nil
}
