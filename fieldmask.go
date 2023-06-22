package fieldmask

import (
	"fmt"
)

// FieldInfo ...
type FieldInfo struct {
	JsonName  string
	SubFields []FieldInfo
}

// ComputeFieldInfos ...
func ComputeFieldInfos(fields []string) ([]FieldInfo, error) {
	return nil, nil
}

// ErrFieldNotFound ...
func ErrFieldNotFound(field string) error {
	return fmt.Errorf("fieldmask: field not found '%s'", field)
}
