package custom_marshaler

import (
	"reflect"

	"github.com/pasztorpisti/qs"
)

// marshalerFactory implements the MarshalerFactory interface and provides
// custom Marshaler for the []byte type.
type marshalerFactory struct {
	orig qs.MarshalerFactory
}

func (f *marshalerFactory) Marshaler(t reflect.Type, opts *qs.MarshalOptions) (qs.Marshaler, error) {
	switch t {
	case byteSliceType:
		return byteSliceMarshaler{}, nil
	case durationType:
		return &durationMarshaler{}, nil
	default:
		return f.orig.Marshaler(t, opts)
	}
}
