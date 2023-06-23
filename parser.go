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

func parseObjectInfo(msgType reflect.Type, parsedObjects map[objectKey]*objectInfo) *objectInfo {
	obj := &objectInfo{
		typeName:   msgType.Name(),
		importPath: msgType.PkgPath(),
	}

	existedObj, existed := parsedObjects[obj.getKey()]
	if existed {
		return existedObj
	}

	obj.subFields = parseMessageFields(msgType, parsedObjects)
	parsedObjects[obj.getKey()] = obj
	return obj
}

func parseMessageFields(structType reflect.Type, parsedObjects map[objectKey]*objectInfo) []objectField {
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
			info = parseObjectInfo(field.Type.Elem(), parsedObjects)
			subType = fieldTypeObject

		case reflect.Slice:
			elemType := field.Type.Elem()
			if elemType.Kind() == reflect.Pointer {
				info = parseObjectInfo(elemType.Elem(), parsedObjects)
				subType = fieldTypeArrayOfObjects
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

func parseMessages(msgList ...proto.Message) []*objectInfo {
	var result []*objectInfo

	parsedObjects := map[objectKey]*objectInfo{}

	for _, msg := range msgList {
		msgType := reflect.TypeOf(msg)
		if msgType == nil || msgType.Kind() != reflect.Pointer {
			panic("invalid message type")
		}

		msgType = msgType.Elem()
		info := parseObjectInfo(msgType, parsedObjects)
		result = append(result, info)
	}
	return result
}
