package fields

import (
	"errors"
	"fmt"
)

// ===========================================
// Field Not Found Error
// ===========================================

// FieldNotFoundError ...
type FieldNotFoundError struct {
	Field string
}

var _ FieldErrorPrepend = FieldNotFoundError{}

func (e FieldNotFoundError) Error() string {
	return fmt.Sprintf("fieldmask: field not found or not allowed '%s'", e.Field)
}

// PrependField ...
func (e FieldNotFoundError) PrependField(parentField string) error {
	return ErrFieldNotFound(parentField + "." + e.Field)
}

// ErrFieldNotFound ...
func ErrFieldNotFound(field string) error {
	return FieldNotFoundError{Field: field}
}

// ===========================================
// Duplicated Field Error
// ===========================================

// DuplicatedFieldError ...
type DuplicatedFieldError struct {
	Field string
}

func (e DuplicatedFieldError) Error() string {
	return fmt.Sprintf("fieldmask: duplicated field '%s'", e.Field)
}

// PrependField ...
func (e DuplicatedFieldError) PrependField(parentField string) error {
	return ErrDuplicatedField(parentField + "." + e.Field)
}

// ErrDuplicatedField ...
func ErrDuplicatedField(field string) error {
	return DuplicatedFieldError{Field: field}
}

var _ FieldErrorPrepend = DuplicatedFieldError{}

// ===========================================
// Prepend Parent Field
// ===========================================

// FieldErrorPrepend ...
type FieldErrorPrepend interface {
	PrependField(parentField string) error
}

// PrependParentField ...
func PrependParentField(err error, parentField string) error {
	updater, ok := err.(FieldErrorPrepend)
	if !ok {
		return err
	}
	return updater.PrependField(parentField)
}

// ErrExceedMaxFields ...
var ErrExceedMaxFields = errors.New("fieldmask: exceeded max number of fields")

// ErrExceedMaxDepth ...
var ErrExceedMaxDepth = errors.New("fieldmask: exceeded max number of field depth")

// ErrExceedMaxFieldComponentLength ...
var ErrExceedMaxFieldComponentLength = errors.New("fieldmask: exceeded length of field components")
