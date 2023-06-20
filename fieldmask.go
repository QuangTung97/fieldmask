package fieldmask

// FieldInfo ...
type FieldInfo struct {
	JsonName  string
	SubFields []FieldInfo
}

func ComputeFieldInfos(fields []string) ([]FieldInfo, error) {
	return nil, nil
}
