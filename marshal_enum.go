package qs

//go:generate go-stringer -type=MarshalPresence --trimprefix=MarshalPresence -output marshal_string.go

// MarshalPresence is an enum that controls the marshaling of empty fields.
// A field is empty if it has its zero value or it is an empty container.
type MarshalPresence int

const (
	// MarshalPresenceMPUnspecified is the zero value of MarshalPresence. In most cases
	// you will use this implicitly by simply leaving the
	// MarshalOptions.DefaultMarshalPresence field uninitialised which results
	// in using the default MarshalPresence which is KeepEmpty.
	MarshalPresenceMPUnspecified MarshalPresence = iota

	// MarshalPresenceKeepEmpty marshals the values of empty fields into the marshal output.
	MarshalPresenceKeepEmpty

	// MarshalPresenceOmitEmpty doesn't marshal the values of empty fields into the marshal output.
	MarshalPresenceOmitEmpty
)
