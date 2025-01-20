package qs

import (
	"fmt"
	"reflect"
)

// IsRequiredFieldError returns ok==false if the given error wasn't caused by a
// required field that was missing from the query string.
// Otherwise it returns the name of the missing required field with ok==true.
func IsRequiredFieldError(e error) (string, bool) {
	if re, ok := e.(*ReqError); ok {
		return re.FieldName, true
	}
	return "", false
}

// ReqError is returned when a struct field marked with the 'req' option isn't
// in the unmarshaled url.Values or query string.
type ReqError struct {
	Message   string
	FieldName string
}

func (e *ReqError) Error() string {
	return e.Message
}

type WrongTypeError struct {
	Actual   reflect.Type
	Expected reflect.Type
}

func (e *WrongTypeError) Error() string {
	return fmt.Sprintf("received type %v, want %v", e.Actual, e.Expected)
}

type WrongKindError struct {
	Actual   reflect.Type
	Expected reflect.Kind
}

func (e *WrongKindError) Error() string {
	return fmt.Sprintf("received type %v of kind %v, want kind %v",
		e.Actual, e.Actual.Kind(), e.Expected)
}

type UnhandledTypeError struct {
	Type reflect.Type
}

func (e *UnhandledTypeError) Error() string {
	return fmt.Sprintf("unhandled type: %v", e.Type)
}
