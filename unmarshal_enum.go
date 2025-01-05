package qs

//go:generate go-stringer -type=UnmarshalPresence --trimprefix=UnmarshalPresence -output unmarshal_string.go

// UnmarshalPresence is an enum that controls the unmarshaling of fields.
// This option is used by the unmarshaler only if the given field isn't present
// in the query string or url.Values that is being unmarshaled.
type UnmarshalPresence int

const (
	// UnmarshalPresenceUPUnspecified is the zero value of UnmarshalPresence. In most cases
	// you will use this implicitly by simply leaving the
	// UnmarshalOptions.DefaultUnmarshalPresence field uninitialised which results
	// in using the default UnmarshalPresence which is Opt.
	UnmarshalPresenceUPUnspecified UnmarshalPresence = iota

	// UnmarshalPresenceOpt tells the unmarshaler to leave struct fields as they are when they
	// aren't present in the query string. However, nil pointers and arrays are
	// created and initialised with new objects.
	UnmarshalPresenceOpt

	// UnmarshalPresenceNil is the same as Opt except that it doesn't initialise nil pointers
	// and slices during unmarshal when they are missing from the query string.
	UnmarshalPresenceNil

	// UnmarshalPresenceReq tells the unmarshaler to fail with an error that can be detected
	// using qs.IsRequiredFieldError if the given field is
	// missing from the query string. While this is rather validation than
	// unmarshaling it is practical to have this in case of simple programs.
	// If you don't want to mix unmarshaling and validation then you can use the
	// Nil option instead with nil pointers and nil arrays to be able to detect
	// missing fields after unmarshaling.
	UnmarshalPresenceReq
)
