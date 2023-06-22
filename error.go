package fieldmask

import "fmt"

// ===========================================
// Field Not Found Error
// ===========================================

// FieldNotFoundError ...
type FieldNotFoundError struct {
	Field string
}

var _ FieldErrorPrepend = FieldNotFoundError{}

func (e FieldNotFoundError) Error() string {
	return fmt.Sprintf("fieldmask: field not found '%s'", e.Field)
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
