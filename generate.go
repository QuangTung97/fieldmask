package fieldmask

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"io"
	"strings"
	"text/template"
)

//go:embed template
var templateString string

type typeAndNewFunc struct {
	StructName          string
	FuncType            string
	ComputeKeepFuncName string
}

type keepFunc struct {
	TypeName string
	FuncName string
	FuncType string

	FieldFuncs     []fieldKeepFunc
	ImplFieldFuncs []implFieldFunc
}

type implFieldFunc struct {
	QualifiedType string
	FuncName      string
	FieldName     string
}

type fieldKeepFunc struct {
	JsonName  string
	FuncValue string

	fieldName string // is private
	funcName  string // is private
	isObject  bool
}

type generateParams struct {
	PackageName     string
	Imports         []string
	TypeAndNewFuncs []typeAndNewFunc
	KeepFuncs       []keepFunc
}

func mapSlice[A any, B any](input []A, fn func(a A) B) []B {
	result := make([]B, 0, len(input))
	for _, e := range input {
		result = append(result, fn(e))
	}
	return result
}

func computeImports(infos []*objectInfo) []string {
	var result []string
	importedPaths := map[string]string{}

	for _, info := range infos {
		oldAlias, existed := importedPaths[info.importPath]
		if existed {
			info.alias = oldAlias
			continue
		}

		alias := "pb"
		if len(importedPaths) > 0 {
			alias = fmt.Sprintf("pb%d", len(importedPaths))
		}
		importedPaths[info.importPath] = alias

		info.alias = alias

		p := fmt.Sprintf("%s %q", alias, info.importPath)
		result = append(result, p)
	}

	return result
}

func getComputeKeepFuncName(e *objectInfo) string {
	return fmt.Sprintf("%s_%s_ComputeKeepFunc", e.alias, e.typeName)
}

func getQualifiedTypeName(e *objectInfo) string {
	return fmt.Sprintf("%s.%s", e.alias, e.typeName)
}

func getFuncTypeSignature(e *objectInfo) string {
	typeName := getQualifiedTypeName(e)
	return fmt.Sprintf("func (newMsg *%s, msg *%s)", typeName, typeName)
}

func funcValueForObject(obj *objectInfo, field objectField) string {
	objectType := getQualifiedTypeName(obj)
	subObjectType := getQualifiedTypeName(field.info)
	funcName := getComputeKeepFuncName(field.info)

	result := fmt.Sprintf(`
func (newMsg *%s, msg *%s) {
	keepFunc := %s(field.SubFields)

	newSubMsg := &%s{}
	keepFunc(newSubMsg, msg.%s)
	newMsg.%s = newSubMsg
}
`, objectType, objectType, funcName, subObjectType, field.name, field.name)

	return strings.TrimSpace(result)
}

func funcValueForArrayOfObjects(obj *objectInfo, field objectField) string {
	objectType := getQualifiedTypeName(obj)
	subObjectType := getQualifiedTypeName(field.info)
	funcName := getComputeKeepFuncName(field.info)

	result := fmt.Sprintf(`
func(newMsg *%s, msg *%s) {
	keepFunc := %s(field.SubFields)

	msgList := make([]*%s, 0, len(msg.%s))
	for _, e := range msg.%s {
		newSubMsg := &%s{}
		keepFunc(newSubMsg, e)
		msgList = append(msgList, newSubMsg)
	}
	newMsg.%s = msgList
}
`, objectType, objectType, funcName, subObjectType, field.name, field.name, subObjectType, field.name)

	return strings.TrimSpace(result)
}

func buildKeepFuncForField(info *objectInfo, subField objectField) fieldKeepFunc {
	funcName := fmt.Sprintf("%s_%s_Keep_%s", info.alias, info.typeName, subField.name)
	isObject := false

	var funcValue string
	switch subField.fieldType {
	case fieldTypeObject:
		funcValue = funcValueForObject(info, subField)
		isObject = true

	case fieldTypeArrayOfObjects:
		funcValue = funcValueForArrayOfObjects(info, subField)
		isObject = true

	default:
		funcValue = funcName
	}

	return fieldKeepFunc{
		JsonName:  subField.jsonName,
		FuncValue: funcValue,

		fieldName: subField.name,
		funcName:  funcName,
		isObject:  isObject,
	}
}

func buildKeepFunc(info *objectInfo) keepFunc {
	fieldFuncs := mapSlice(info.subFields, func(subField objectField) fieldKeepFunc {
		return buildKeepFuncForField(info, subField)
	})

	implFuncs := make([]implFieldFunc, 0, len(fieldFuncs))
	for _, fn := range fieldFuncs {
		if fn.isObject {
			continue
		}
		implFuncs = append(implFuncs, implFieldFunc{
			QualifiedType: getQualifiedTypeName(info),
			FieldName:     fn.fieldName,
			FuncName:      fn.funcName,
		})
	}

	return keepFunc{
		TypeName:       info.typeName,
		FuncName:       getComputeKeepFuncName(info),
		FuncType:       getFuncTypeSignature(info),
		FieldFuncs:     fieldFuncs,
		ImplFieldFuncs: implFuncs,
	}
}

func traverseAllObjectInfos(objects []*objectInfo, deduplicated map[objectKey]struct{}) []*objectInfo {
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
		result = append(result, traverseAllObjectInfos(subObjects, deduplicated)...)
	}
	return result
}

func generateCode(
	writer io.Writer, inputInfos []*objectInfo,
	packageName string,
) {
	tmpl, err := template.New("fieldmask").Parse(templateString)
	if err != nil {
		panic(err)
	}

	infos := traverseAllObjectInfos(inputInfos, map[objectKey]struct{}{})

	inputSet := map[objectKey]struct{}{}
	for _, obj := range inputInfos {
		inputSet[obj.getKey()] = struct{}{}
	}

	inputOnlyInfos := make([]*objectInfo, 0)
	for _, info := range infos {
		_, ok := inputSet[info.getKey()]
		if ok {
			inputOnlyInfos = append(inputOnlyInfos, info)
		}
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, generateParams{
		PackageName: packageName,
		Imports:     computeImports(infos),
		TypeAndNewFuncs: mapSlice(inputOnlyInfos, func(e *objectInfo) typeAndNewFunc {
			return typeAndNewFunc{
				StructName:          e.typeName + "FieldMask",
				FuncType:            getFuncTypeSignature(e),
				ComputeKeepFuncName: getComputeKeepFuncName(e),
			}
		}),
		KeepFuncs: mapSlice(infos, buildKeepFunc),
	})
	if err != nil {
		panic(err)
	}

	sourceData, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Println(buf.String())
		panic(err)
	}

	_, err = writer.Write(sourceData)
	if err != nil {
		panic(err)
	}
}
