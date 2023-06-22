package fields

import (
	"strings"
)

// FieldInfo ...
type FieldInfo struct {
	JsonName  string
	SubFields []FieldInfo
}

type fieldInfoParser struct {
	options *computeOptions

	fields     []string
	subFields  map[string]*fieldInfoParser
	fieldCount *int
}

func newEmptyParser(options *computeOptions) *fieldInfoParser {
	return &fieldInfoParser{
		options: options,

		subFields:  map[string]*fieldInfoParser{},
		fieldCount: new(int),
	}
}

func (p *fieldInfoParser) clone() *fieldInfoParser {
	return &fieldInfoParser{
		options:    p.options,
		subFields:  map[string]*fieldInfoParser{},
		fieldCount: p.fieldCount,
	}
}

func (p *fieldInfoParser) addIfNotExisted(fieldName string, isSubField bool) error {
	subParser, ok := p.subFields[fieldName]
	if !ok {
		p.subFields[fieldName] = nil
		p.fields = append(p.fields, fieldName)

		*p.fieldCount++
		if *p.fieldCount > p.options.maxFields {
			return ErrExceedMaxFields
		}
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

func computeFields(fullField string, result *fieldInfoParser, depth int) error {
	if depth > result.options.maxDepth {
		return ErrExceedMaxDepth
	}

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
		subParser = result.clone()
		result.subFields[fieldName] = subParser
	}

	err = computeFields(remaining, subParser, depth+1)
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
	opts := newComputeOptions(options)
	parser := newEmptyParser(opts)

	for _, f := range fields {
		err := computeFields(f, parser, 1)
		if err != nil {
			return nil, err
		}
	}

	return parser.toFieldInfos(), nil
}
