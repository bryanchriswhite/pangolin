package errors

import (
	"../utils"
	"fmt"
	"reflect"
)

type Error struct {
	message string
}

func (e *Error) Error() string {
	return e.message
}

func New(message string) error {
	return &Error{message}
}

func NewCoercionError(a utils.Any) error {
	return New(fmt.Sprintf("Coercion error from type: %v", reflect.TypeOf(a)))
}