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

func parseMessages(msgList ...ProtoMessage) []*objectInfo {
	var result []*objectInfo

	parsedObjects := map[objectKey]*objectInfo{}

	for _, msg := range msgList {
		msgType := reflect.TypeOf(msg.protoMsg)
		if msgType == nil || msgType.Kind() != reflect.Pointer {
			panic("invalid message type")
		}
		msgType = msgType.Elem()

		if isSpecialPackage(msgType.PkgPath()) {
			panic(fmt.Sprintf("not allow type '%s'", msgType.Name()))
		}

		info := parseObjectInfo(msgType, parsedObjects, nil)
		result = append(result, info)
	}

	selector := newFieldSelector()

	selector.traverseAll(msgList, result)

	return result
}
