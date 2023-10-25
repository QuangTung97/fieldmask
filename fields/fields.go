package fields

// FieldInfo ...
type FieldInfo struct {
	FieldName string
	SubFields []FieldInfo
}

func getFieldCollector(fields []string, opts *computeOptions) (*fieldInfoCollector, error) {
	coll := newCollector(opts)

	for _, f := range fields {
		p := newParser(f, coll)
		if err := p.parse(); err != nil {
			return nil, err
		}
	}

	return coll, nil
}

func validateLimitedToFields(fields []FieldInfo, coll *fieldInfoCollector) error {
	for _, f := range fields {
		if coll == nil {
			return ErrFieldNotFound(f.FieldName)
		}

		subColl, ok := coll.subCollectors[f.FieldName]
		if !ok {
			return ErrFieldNotFound(f.FieldName)
		}
		if len(f.SubFields) > 0 {
			err := validateLimitedToFields(f.SubFields, subColl)
			if err != nil {
				return PrependParentField(err, f.FieldName)
			}
		}
	}
	return nil
}

// ComputeFieldInfos ...
func ComputeFieldInfos(fields []string, options ...Option) ([]FieldInfo, error) {
	opts := newComputeOptions(options)

	resultCollector, err := getFieldCollector(fields, opts)
	if err != nil {
		return nil, err
	}

	resultFields := resultCollector.toFieldInfos()

	if len(opts.limitedToFields) > 0 {
		allowedFieldsCollector, err := getFieldCollector(opts.limitedToFields, opts)
		if err != nil {
			return nil, err
		}
		if err := validateLimitedToFields(resultFields, allowedFieldsCollector); err != nil {
			return nil, err
		}
	}

	return resultFields, nil
}
