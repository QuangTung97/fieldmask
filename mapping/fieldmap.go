package fieldmap

import (
	"fmt"
	"github.com/QuangTung97/fieldmask/fields"
	"reflect"
	"sync"
)

// Field ...
type Field interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// MapType ...
type MapType[F Field] interface {
	GetRoot() F
}

// RootField ...
const RootField = "Root"

// FieldMap ...
type FieldMap[F Field, T MapType[F]] struct {
	options fieldMapOptions

	mapping    T
	fields     []F
	structRoot F

	children   [][]F
	parentList []F
	fieldNames []string
	structTags map[string][]string

	tagsIndex     map[string]*tagToFieldMapping[F]
	tagsIndexOnce sync.Once
}

type tagToFieldMapping[F Field] struct {
	field     F // empty for the first
	subFields map[string]*tagToFieldMapping[F]
}

func (i *tagToFieldMapping[F]) getSubFields() map[string]*tagToFieldMapping[F] {
	if i.subFields == nil {
		i.subFields = map[string]*tagToFieldMapping[F]{}
	}
	return i.subFields
}

type fieldMapOptions struct {
	structTags []string
}

// Option ...
type Option func(opts *fieldMapOptions)

// WithStructTags ...
func WithStructTags(tags ...string) Option {
	return func(opts *fieldMapOptions) {
		opts.structTags = tags
	}
}

func computeOptions(options []Option) fieldMapOptions {
	opts := fieldMapOptions{
		structTags: nil,
	}
	for _, fn := range options {
		fn(&opts)
	}
	return opts
}

// New ...
func New[F Field, T MapType[F]](options ...Option) *FieldMap[F, T] {
	opts := computeOptions(options)

	f := &FieldMap[F, T]{
		options:    opts,
		structTags: map[string][]string{},
	}

	ordinal := int64(0)
	var info parentInfoData[F]

	var mapping T
	val := reflect.ValueOf(&mapping)
	val = val.Elem()

	f.traverse(val, &ordinal, info)

	f.mapping = mapping

	f.children = make([][]F, len(f.fields))

	for i, parent := range f.parentList {
		if parent <= 0 {
			continue
		}
		parentIndex := parent - 1
		f.children[parentIndex] = append(f.children[parentIndex], F(i+1))
	}

	return f
}

func (*FieldMap[F, T]) getField(num int64) F {
	var field F
	val := reflect.ValueOf(&field).Elem()
	val.SetInt(num)
	return field
}

func (*FieldMap[F, T]) getFieldType() reflect.Type {
	var field F
	return reflect.TypeOf(field)
}

type parentInfoData[F Field] struct {
	prevRoot F

	fieldName     string
	fullFieldName string

	structTags map[string]string
}

func (parentInfoData[F]) isParentField(index int) bool {
	return index == 0
}

func (p parentInfoData[F]) computeFullName(currentName string) string {
	if len(p.fullFieldName) > 0 {
		return p.fullFieldName + "." + currentName
	}
	return currentName
}

func (f *FieldMap[F, T]) findStructTags(
	fieldType reflect.StructField,
	fullFieldName string,
) map[string]string {
	structTags := map[string]string{}

	for _, tag := range f.options.structTags {
		tagVal := fieldType.Tag.Get(tag)
		if len(tagVal) == 0 {
			panic(
				fmt.Sprintf(
					"missing struct tag %q for field %q",
					tag, fullFieldName,
				),
			)
		}
		structTags[tag] = tagVal
	}
	return structTags
}

func (f *FieldMap[F, T]) getRootField(
	val reflect.Value, parentInfo parentInfoData[F], ordinal *int64,
) F {
	var panicStr string
	if len(parentInfo.fullFieldName) > 0 {
		panicStr = fmt.Sprintf("missing field %q for field %q", RootField, parentInfo.fullFieldName)
	} else {
		panicStr = fmt.Sprintf("missing field %q for root of struct", RootField)
	}

	if val.NumField() == 0 {
		panic(panicStr)
	}

	fieldName := val.Type().Field(0).Name
	if fieldName != RootField {
		panic(panicStr)
	}
	return f.getField(*ordinal + 1)
}

func (f *FieldMap[F, T]) handleSingleField(
	val reflect.Value, i int, parentInfo parentInfoData[F],
	rootField F, ordinal *int64,
) {
	field := val.Field(i)
	fieldType := val.Type().Field(i)
	fieldName := fieldType.Name
	fullFieldName := parentInfo.computeFullName(fieldName)

	var currentStructTags map[string]string
	if !parentInfo.isParentField(i) {
		currentStructTags = f.findStructTags(fieldType, fullFieldName)

		if field.Kind() == reflect.Struct {
			newInfo := parentInfoData[F]{
				prevRoot: rootField,

				fieldName:     fieldName,
				fullFieldName: fullFieldName,

				structTags: currentStructTags,
			}
			f.traverse(field, ordinal, newInfo)
			return
		}
	}

	if field.Type() != f.getFieldType() {
		panic(fmt.Sprintf("invalid type for field %q", fullFieldName))
	}

	*ordinal++

	f.fields = append(f.fields, f.getField(*ordinal))

	if parentInfo.isParentField(i) {
		f.parentList = append(f.parentList, parentInfo.prevRoot)
		f.fieldNames = append(f.fieldNames, parentInfo.fieldName)

		for _, tag := range f.options.structTags {
			f.structTags[tag] = append(f.structTags[tag], parentInfo.structTags[tag])
		}
	} else {
		f.parentList = append(f.parentList, rootField)
		f.fieldNames = append(f.fieldNames, fieldName)

		for _, tag := range f.options.structTags {
			f.structTags[tag] = append(f.structTags[tag], currentStructTags[tag])
		}
	}
	field.SetInt(*ordinal)
}

func (f *FieldMap[F, T]) checkGetRootImpl() {
	var mapping T
	rootVal := reflect.ValueOf(&mapping).Elem().Field(0)

	panicIfNotEq := func(num int64) {
		rootVal.SetInt(num)
		if mapping.GetRoot() != f.getField(num) {
			panic("invalid GetRoot implementation")
		}
	}
	panicIfNotEq(1)
	panicIfNotEq(3)
	panicIfNotEq(7)
	panicIfNotEq(13)
	panicIfNotEq(31)
}

func (f *FieldMap[F, T]) traverse(
	val reflect.Value, ordinal *int64, parentInfo parentInfoData[F],
) {
	rootField := f.getRootField(val, parentInfo, ordinal)

	var empty F
	if parentInfo.prevRoot == empty {
		f.checkGetRootImpl()
		f.structRoot = rootField
	}

	for i := 0; i < val.NumField(); i++ {
		f.handleSingleField(val, i, parentInfo, rootField, ordinal)
	}
}

// GetMapping ...
func (f *FieldMap[F, T]) GetMapping() T {
	return f.mapping
}

func (*FieldMap[F, T]) indexOf(field F) int64 {
	return reflect.ValueOf(field).Int() - 1
}

// IsStruct ...
func (f *FieldMap[F, T]) IsStruct(field F) bool {
	index := f.indexOf(field)
	return len(f.children[index]) > 0
}

// ChildrenOf ...
func (f *FieldMap[F, T]) ChildrenOf(field F) []F {
	index := f.indexOf(field)
	result := make([]F, 0, len(f.children[index]))
	for _, childField := range f.children[index] {
		result = append(result, childField)
	}
	return result
}

// ParentOf ...
func (f *FieldMap[F, T]) ParentOf(field F) F {
	return f.parentList[f.indexOf(field)]
}

// AncestorOf includes itself, parent, and all parents of parents
func (f *FieldMap[F, T]) AncestorOf(field F) []F {
	var empty F

	result := []F{field}
	for {
		field = f.ParentOf(field)
		if field == empty {
			return result
		}
		result = append(result, field)
	}
}

// GetFieldName ...
func (f *FieldMap[F, T]) GetFieldName(field F) string {
	return f.fieldNames[f.indexOf(field)]
}

// GetFullFieldName ...
func (f *FieldMap[F, T]) GetFullFieldName(field F) string {
	fullName := ""
	for {
		name := f.GetFieldName(field)
		if len(fullName) > 0 {
			fullName = name + "." + fullName
		} else {
			fullName = name
		}

		field = f.ParentOf(field)
		if field == f.structRoot {
			return fullName
		}
	}
}

// GetStructTag ...
func (f *FieldMap[F, T]) GetStructTag(tag string, field F) string {
	return f.structTags[tag][f.indexOf(field)]
}

// GetFullStructTag ...
func (f *FieldMap[F, T]) GetFullStructTag(tag string, field F) string {
	fullTag := ""
	for {
		tagName := f.GetStructTag(tag, field)
		if len(fullTag) > 0 {
			fullTag = tagName + "." + fullTag
		} else {
			fullTag = tagName
		}

		field = f.ParentOf(field)
		if field == f.structRoot {
			return fullTag
		}
	}
}

func (f *FieldMap[F, T]) buildTagMappingForField(tag string, field F) *tagToFieldMapping[F] {
	tagMapping := &tagToFieldMapping[F]{
		field: field,
	}

	childrenFields := f.ChildrenOf(field)
	for _, childField := range childrenFields {
		tagValue := f.GetStructTag(tag, childField)
		tagMapping.getSubFields()[tagValue] = f.buildTagMappingForField(tag, childField)
	}

	return tagMapping
}

func (f *FieldMap[F, T]) buildTagsIndex() {
	f.tagsIndex = map[string]*tagToFieldMapping[F]{}

	for tag := range f.structTags {
		f.tagsIndex[tag] = f.buildTagMappingForField(tag, f.structRoot)
	}
}

func (f *FieldMap[F, T]) getTagMapping(tag string) *tagToFieldMapping[F] {
	f.tagsIndexOnce.Do(f.buildTagsIndex)
	return f.tagsIndex[tag]
}

func (f *FieldMap[F, T]) fromMaskedFieldsRecursive(
	tagMapping *tagToFieldMapping[F],
	maskedFields []fields.FieldInfo, result []F,
) ([]F, error) {
	for _, maskedField := range maskedFields {
		subTagMapping, ok := tagMapping.subFields[maskedField.FieldName]
		if !ok {
			return nil, fields.ErrFieldNotFound(maskedField.FieldName)
		}
		if len(maskedField.SubFields) > 0 {
			var err error
			result, err = f.fromMaskedFieldsRecursive(subTagMapping, maskedField.SubFields, result)
			if err != nil {
				return nil, fields.PrependParentField(err, maskedField.FieldName)
			}
			continue
		}
		result = append(result, subTagMapping.field)
	}
	return result, nil
}

// FromMaskedFields ...
func (f *FieldMap[F, T]) FromMaskedFields(
	tag string, maskedFields []fields.FieldInfo,
) ([]F, error) {
	result := make([]F, 0, len(maskedFields))
	return f.fromMaskedFieldsRecursive(f.getTagMapping(tag), maskedFields, result)
}
