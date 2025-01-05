package custom_marshaler

import (
	"reflect"

	"github.com/pasztorpisti/qs"
)

// unmarshalerFactory implements the UnmarshalerFactory interface and provides
// custom Unmarshaler for the []byte type.
type unmarshalerFactory struct {
	orig qs.UnmarshalerFactory
}

func (f *unmarshalerFactory) Unmarshaler(t reflect.Type, opts *qs.UnmarshalOptions) (qs.Unmarshaler, error) {
	switch t {
	case byteSliceType:
		return byteSliceMarshaler{}, nil
	case durationType:
		return &durationMarshaler{}, nil
	default:
		return f.orig.Unmarshaler(t, opts)
	}
}
