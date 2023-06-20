package fieldmask

import (
	"github.com/golang/protobuf/proto"
	"reflect"
	"strings"
)

type subFieldType int

const (
	subFieldTypeSimple subFieldType = iota
	subFieldTypeObject
	subFieldTypeArray
)

type objectField struct {
	name      string
	jsonName  string
	subFields []objectField
	subType   subFieldType
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

func parseMessageFields(structType reflect.Type) []objectField {
	var result []objectField
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		jsonName, ok := getJsonName(field)
		if !ok {
			continue
		}

		var subFields []objectField
		subType := subFieldTypeSimple

		switch field.Type.Kind() {
		case reflect.Pointer:
			subFields = parseMessageFields(field.Type.Elem())
			subType = subFieldTypeObject

		case reflect.Slice:
			subFields = parseMessageFields(field.Type.Elem().Elem())
			subType = subFieldTypeArray
		}

		result = append(result, objectField{
			name:      field.Name,
			jsonName:  jsonName,
			subFields: subFields,
			subType:   subType,
		})
	}
	return result
}

func parseMessage(msg proto.Message) []objectField {
	msgType := reflect.TypeOf(msg)
	// TODO Check pointer
	msgType = msgType.Elem()

	return parseMessageFields(msgType)
}
