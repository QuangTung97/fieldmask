package fieldmask

import (
	"fmt"
	"github.com/QuangTung97/fieldmask/fields"
	"github.com/golang/protobuf/proto"
	"reflect"
	"strings"
)

type fieldType int

const (
	fieldTypeSimple fieldType = iota
	fieldTypeObject
	fieldTypeArrayOfObjects
	fieldTypeArrayOfPrimitives
	fieldTypeSpecialField
)

var ignoredImportPathPrefixes = []string{
	"github.com/golang/protobuf/ptypes",
	"github.com/gogo/protobuf/types",
}

type objectInfo struct {
	typeName   string
	importPath string

	subFields []objectField

	alias string // computed by generate.go
}

func (o objectInfo) getKey() objectKey {
	return objectKey{
		typeName:   o.typeName,
		importPath: o.importPath,
	}
}

type objectKey struct {
	typeName   string
	importPath string
}

type objectField struct {
	name      string
	jsonName  string
	fieldType fieldType
	info      *objectInfo
}

func getJSONName(field reflect.StructField) (string, bool) {
	tag := field.Tag.Get("protobuf")
	if len(tag) == 0 {
		return "", false
	}

	var jsonName string
	for _, e := range strings.Split(tag, ",") {
		kv := strings.Split(e, "=")
		if len(kv) != 2 {
			continue
		}
		key := kv[0]
		val := kv[1]

		if key == "name" && len(jsonName) == 0 {
			jsonName = val
		} else if key == "json" {
			jsonName = val
		}
	}
	return jsonName, true
}

func isSpecialPackage(importPath string) bool {
	for _, prefix := range ignoredImportPathPrefixes {
		if strings.HasPrefix(importPath, prefix) {
			return true
		}
	}
	return false
}

func parseObjectInfo(
	msgType reflect.Type, parsedObjects map[objectKey]*objectInfo,
	subType *fieldType,
) *objectInfo {
	obj := &objectInfo{
		typeName:   msgType.Name(),
		importPath: msgType.PkgPath(),
	}

	existedObj, existed := parsedObjects[obj.getKey()]
	if existed {
		return existedObj
	}

	if isSpecialPackage(obj.importPath) {
		*subType = fieldTypeSpecialField
		return nil
	}

	obj.subFields = parseMessageFields(msgType, parsedObjects)
	parsedObjects[obj.getKey()] = obj
	return obj
}

func parseMessageFields(
	structType reflect.Type, parsedObjects map[objectKey]*objectInfo,
) []objectField {
	var result []objectField
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		jsonName, ok := getJSONName(field)
		if !ok {
			continue
		}

		var info *objectInfo
		subType := fieldTypeSimple

		switch field.Type.Kind() {
		case reflect.Pointer:
			subType = fieldTypeObject
			info = parseObjectInfo(field.Type.Elem(), parsedObjects, &subType)

		case reflect.Slice:
			elemType := field.Type.Elem()
			if elemType.Kind() == reflect.Pointer {
				subType = fieldTypeArrayOfObjects
				info = parseObjectInfo(elemType.Elem(), parsedObjects, &subType)
			} else {
				subType = fieldTypeArrayOfPrimitives
			}
		}

		result = append(result, objectField{
			name:      field.Name,
			jsonName:  jsonName,
			info:      info,
			fieldType: subType,
		})
	}

	return result
}

// ProtoMessage ...
type ProtoMessage struct {
	protoMsg  proto.Message
	limitedTo []fields.FieldInfo
}

// NewProtoMessage ...
func NewProtoMessage(msg proto.Message) ProtoMessage {
	return ProtoMessage{
		protoMsg: msg,
	}
}

// NewProtoMessageWithFields ...
func NewProtoMessageWithFields(msg proto.Message, limitedToFields []string) ProtoMessage {
	limitedTo, err := fields.ComputeFieldInfos(limitedToFields)
	if err != nil {
		panic(err)
	}
	return ProtoMessage{
		protoMsg:  msg,
		limitedTo: limitedTo,
	}
}

type inUseFields struct {
	limitedTo []fields.FieldInfo
	allowAll  bool
	prefix    string
}

func (u inUseFields) toAllowFunc() func(jsonName string) (inUseFields, bool) {
	set := map[string]inUseFields{}
	for _, f := range u.limitedTo {
		set[f.FieldName] = inUseFields{
			prefix:    u.prefix + f.FieldName + ".",
			limitedTo: f.SubFields,
		}
	}

	return func(jsonName string) (inUseFields, bool) {
		subUsed, ok := set[jsonName]
		return subUsed, ok
	}
}

func (u inUseFields) checkFieldsExisted(objectFields []objectField) {
	jsonNames := map[string]struct{}{}
	for _, r := range objectFields {
		jsonNames[r.jsonName] = struct{}{}
	}
	for _, f := range u.limitedTo {
		_, ok := jsonNames[f.FieldName]
		if !ok {
			panic(fmt.Sprintf("not found field '%s'", u.prefix+f.FieldName))
		}
	}
}

func (u inUseFields) traverseInfo(info *objectInfo, traversedObjects map[objectKey]struct{}) {
	if info == nil {
		return
	}

	objKey := info.getKey()
	_, existed := traversedObjects[objKey]
	if existed {
		panic(fmt.Sprintf("conflicted limited to fields declaration for type '%s'", info.typeName))
	}
	traversedObjects[objKey] = struct{}{}

	if u.allowAll {
		return
	}

	u.checkFieldsExisted(info.subFields)

	allowFunc := u.toAllowFunc()

	newSubFields := make([]objectField, 0)
	for _, subField := range info.subFields {
		subUsedFields, allow := allowFunc(subField.jsonName)
		if !allow {
			continue
		}
		if subField.info != nil && len(subUsedFields.limitedTo) == 0 {
			subField.info = nil
			subField.fieldType = fieldTypeSimple
		} else {
			subUsedFields.traverseInfo(subField.info, traversedObjects)
		}

		newSubFields = append(newSubFields, subField)
	}
	info.subFields = newSubFields
}

func parseMessages(msgList ...ProtoMessage) []*objectInfo {
	var result []*objectInfo

	parsedObjects := map[objectKey]*objectInfo{}

	for _, msg := range msgList {
		msgType := reflect.TypeOf(msg.protoMsg)
		if msgType == nil || msgType.Kind() != reflect.Pointer {
			panic("invalid message type")
		}

		msgType = msgType.Elem()
		info := parseObjectInfo(msgType, parsedObjects, nil)
		result = append(result, info)
	}

	traversedObjects := map[objectKey]struct{}{}

	for i, info := range result {
		msg := msgList[i]

		allowAllFields := true
		if len(msg.limitedTo) > 0 {
			allowAllFields = false
		}

		usedFields := inUseFields{
			limitedTo: msg.limitedTo,
			allowAll:  allowAllFields,
		}
		usedFields.traverseInfo(info, traversedObjects)
	}

	return result
}
