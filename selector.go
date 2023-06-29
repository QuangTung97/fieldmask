package fieldmask

import (
	"fmt"
	"github.com/QuangTung97/fieldmask/fields"
)

type emptyStruct = struct{}

type selectedFields struct {
	selectedSet map[string]emptyStruct
}

type fieldSelector struct {
	selectedMap map[objectKey]selectedFields
}

func newFieldSelector() *fieldSelector {
	return &fieldSelector{
		selectedMap: map[objectKey]selectedFields{},
	}
}

func getObjectInfoFieldMap(info *objectInfo) map[string]objectField {
	result := map[string]objectField{}
	for _, subField := range info.subFields {
		result[subField.jsonName] = subField
	}
	return result
}

func (s *fieldSelector) getSelectedFields(k objectKey) selectedFields {
	selected, ok := s.selectedMap[k]
	if ok {
		return selected
	}

	selected = selectedFields{
		selectedSet: map[string]struct{}{},
	}
	s.selectedMap[k] = selected
	return selected
}

func (s *fieldSelector) traverse(info *objectInfo, limitedTo []fields.FieldInfo) {
	s.traverseRecursive(info, limitedTo, "")
}

func (s *fieldSelector) keepSelectedFields(info *objectInfo) {
	if s.allowAll(info) {
		return
	}

	var keptFields []objectField
	for _, f := range info.subFields {
		if !s.allowField(info, f.jsonName) {
			continue
		}

		if f.info == nil {
			keptFields = append(keptFields, f)
			continue
		}

		if f.info != nil && s.allowAll(f.info) {
			f.info = nil
			f.fieldType = fieldTypeSimple
		}

		keptFields = append(keptFields, f)
	}

	info.subFields = keptFields
}

func (s *fieldSelector) traverseRecursive(info *objectInfo, limitedTo []fields.FieldInfo, prefix string) {
	if info == nil {
		return
	}

	k := info.getKey()
	selected := s.getSelectedFields(k)

	if len(limitedTo) == 0 {
		return
	}

	fieldMap := getObjectInfoFieldMap(info)

	for _, f := range limitedTo {
		selected.selectedSet[f.FieldName] = struct{}{}

		subInfo, ok := fieldMap[f.FieldName]
		if !ok {
			panic(fmt.Sprintf("not found field '%s'", prefix+f.FieldName))
		}

		s.traverseRecursive(subInfo.info, f.SubFields, prefix+f.FieldName+".")
	}
}

func (s *fieldSelector) allowAll(info *objectInfo) bool {
	return len(s.selectedMap[info.getKey()].selectedSet) == 0
}

func (s *fieldSelector) allowField(info *objectInfo, jsonName string) bool {
	_, ok := s.selectedMap[info.getKey()].selectedSet[jsonName]
	return ok
}

func (s *fieldSelector) traverseAll(msgList []ProtoMessage, infos []*objectInfo) {
	for i, msg := range msgList {
		s.traverse(infos[i], msg.limitedTo)
	}

	allInfos := traverseAllObjectInfos(infos)
	for _, info := range allInfos {
		s.keepSelectedFields(info)
	}
}

// traverseAllObjectInfos list all infos
func traverseAllObjectInfos(objects []*objectInfo) []*objectInfo {
	return traverseAllObjectInfosRecursive(objects, map[objectKey]struct{}{})
}

func traverseAllObjectInfosRecursive(objects []*objectInfo, deduplicated map[objectKey]struct{}) []*objectInfo {
	var result []*objectInfo
	for _, obj := range objects {
		_, existed := deduplicated[obj.getKey()]
		if existed {
			continue
		}
		deduplicated[obj.getKey()] = struct{}{}

		result = append(result, obj)

		var subObjects []*objectInfo
		for _, f := range obj.subFields {
			if f.info != nil {
				subObjects = append(subObjects, f.info)
			}
		}
		result = append(result, traverseAllObjectInfosRecursive(subObjects, deduplicated)...)
	}
	return result
}
