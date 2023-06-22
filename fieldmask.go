package fieldmask

import (
	"fmt"
	"strings"
)

// FieldInfo ...
type FieldInfo struct {
	JsonName  string
	SubFields []FieldInfo
}

type computeOptions struct {
	maxFields int
	maxDepth  int
}

func newComputeOptions(options []Option) computeOptions {
	opts := computeOptions{
		maxFields: 1000,
		maxDepth:  5,
	}
	for _, fn := range options {
		fn(&opts)
	}
	return opts
}

// Option ...
type Option func(opts *computeOptions)

type fieldInfoParser struct {
	fields    []string
	subFields map[string]*fieldInfoParser
}

func newEmptyParser() *fieldInfoParser {
	return &fieldInfoParser{
		subFields: map[string]*fieldInfoParser{},
	}
}

func (p *fieldInfoParser) addIfNotExisted(fieldName string, isSubField bool) error {
	subParser, ok := p.subFields[fieldName]
	if !ok {
		p.subFields[fieldName] = nil
		p.fields = append(p.fields, fieldName)
		return nil
	}
	if !isSubField {
		return ErrDuplicatedField(fieldName)
	}
	if subParser == nil {
		return ErrDuplicatedField(fieldName)
	}
	return nil
}

func computeFields(fullField string, result *fieldInfoParser) error {
	index := strings.Index(fullField, ".")
	if index < 0 {
		return result.addIfNotExisted(fullField, false)
	}

	fieldName := fullField[:index]
	remaining := fullField[index+1:]

	err := result.addIfNotExisted(fieldName, true)
	if err != nil {
		return err
	}

	subParser := result.subFields[fieldName]
	if subParser == nil {
		subParser = newEmptyParser()
		result.subFields[fieldName] = subParser
	}

	err = computeFields(remaining, subParser)
	if err != nil {
		return PrependParentField(err, fieldName)
	}

	return nil
}

func (p *fieldInfoParser) toFieldInfos() []FieldInfo {
	result := make([]FieldInfo, 0, len(p.fields))

	for _, f := range p.fields {
		var subFields []FieldInfo

		subParser := p.subFields[f]
		if subParser != nil {
			subFields = subParser.toFieldInfos()
		}

		result = append(result, FieldInfo{
			JsonName:  f,
			SubFields: subFields,
		})
	}

	return result
}

// ComputeFieldInfos ...
func ComputeFieldInfos(fields []string, options ...Option) ([]FieldInfo, error) {
	parser := newEmptyParser()

	for _, f := range fields {
		err := computeFields(f, parser)
		if err != nil {
			return nil, err
		}
	}

	return parser.toFieldInfos(), nil
}

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
