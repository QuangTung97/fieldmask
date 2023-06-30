package fields

// FieldInfo ...
type FieldInfo struct {
	FieldName string
	SubFields []FieldInfo
}

// ComputeFieldInfos ...
func ComputeFieldInfos(fields []string, options ...Option) ([]FieldInfo, error) {
	opts := newComputeOptions(options)
	coll := newCollector(opts)

	for _, f := range fields {
		p := newParser(f, coll)
		if err := p.parse(); err != nil {
			return nil, err
		}
	}

	return coll.toFieldInfos(), nil
}
