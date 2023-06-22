package fieldmask

import (
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
)

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

func getJsonName(field reflect.StructField) (string, bool) {
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

func parseObjectInfo(msgType reflect.Type) *objectInfo {
	return &objectInfo{
		typeName:   msgType.Name(),
		importPath: msgType.PkgPath(),
		subFields:  parseMessageFields(msgType),
	}
}

func parseMessageFields(structType reflect.Type) []objectField {
	var result []objectField
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		jsonName, ok := getJsonName(field)
		if !ok {
			continue
		}

		var info *objectInfo
		subType := fieldTypeSimple

		switch field.Type.Kind() {
		case reflect.Pointer:
			info = parseObjectInfo(field.Type.Elem())
			subType = fieldTypeObject

		case reflect.Slice:
			info = parseObjectInfo(field.Type.Elem().Elem())
			subType = fieldTypeArrayOfObjects
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

func parseMessage(msg proto.Message) *objectInfo {
	msgType := reflect.TypeOf(msg)
	if msgType == nil || msgType.Kind() != reflect.Pointer {
		panic("invalid message type")
	}

	msgType = msgType.Elem()
	return parseObjectInfo(msgType)
}
