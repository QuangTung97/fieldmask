package fields

type fieldInfoCollector struct {
	depth   int
	options *computeOptions

	subFields     []string
	subCollectors map[string]*fieldInfoCollector
	fieldCount    *int
}

func newCollector(options *computeOptions) *fieldInfoCollector {
	return &fieldInfoCollector{
		depth:   1,
		options: options,

		subCollectors: map[string]*fieldInfoCollector{},
		fieldCount:    new(int),
	}
}

func (c *fieldInfoCollector) newSubCollector(fieldElem string) (*fieldInfoCollector, error) {
	if c.depth >= c.options.maxDepth {
		return nil, ErrExceedMaxDepth
	}

	sub := c.subCollectors[fieldElem]
	if sub != nil {
		return sub, nil
	}
	sub = &fieldInfoCollector{
		depth:         c.depth + 1,
		options:       c.options,
		subCollectors: map[string]*fieldInfoCollector{},
		fieldCount:    c.fieldCount,
	}
	c.subCollectors[fieldElem] = sub
	return sub, nil
}

//revive:disable-next-line:flag-parameter
func (c *fieldInfoCollector) addIfNotExisted(fieldElem string, havingSubFields bool) error {
	if len(fieldElem) > c.options.maxComponentLen {
		return ErrExceedMaxFieldComponentLength
	}
	subParser, ok := c.subCollectors[fieldElem]
	if !ok {
		c.subCollectors[fieldElem] = nil
		c.subFields = append(c.subFields, fieldElem)

		*c.fieldCount++
		if *c.fieldCount > c.options.maxFields {
			return ErrExceedMaxFields
		}
		return nil
	}
	if !havingSubFields {
		return ErrDuplicatedField(fieldElem)
	}
	if subParser == nil {
		return ErrDuplicatedField(fieldElem)
	}
	return nil
}

func (c *fieldInfoCollector) toFieldInfos() []FieldInfo {
	result := make([]FieldInfo, 0, len(c.subFields))

	for _, f := range c.subFields {
		var subFields []FieldInfo

		subParser := c.subCollectors[f]
		if subParser != nil {
			subFields = subParser.toFieldInfos()
		}

		result = append(result, FieldInfo{
			FieldName: f,
			SubFields: subFields,
		})
	}

	return result
}
