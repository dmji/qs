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

func (p *unmarshalerFactory) RegisterSubFactory(k reflect.Kind, fn qs.UnmarshalerFactoryFunc) error {
	panic("!mock not implemented!")

}

func (p *unmarshalerFactory) RegisterCustomType(k reflect.Type, fn qs.PrimitiveUnmarshalerFunc) error {
	panic("!mock not implemented!")

}

func (p *unmarshalerFactory) RegisterKindOverride(k reflect.Kind, fn qs.PrimitiveUnmarshalerFunc) error {
	panic("!mock not implemented!")
}
