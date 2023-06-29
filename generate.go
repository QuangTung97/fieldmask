package fieldmask

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"io"
	"os"
	"strings"
	"text/template"
)

//go:embed template
var fieldmaskTemplateString string

type typeAndNewFunc struct {
	StructName          string
	FuncType            string
	ComputeKeepFuncName string
	QualifiedType       string
}

type keepFunc struct {
	TypeName string
	FuncName string
	FuncType string

	FieldFuncs     []fieldKeepFunc
	FieldFuncImpls []fieldFuncImpl
}

type fieldFuncImpl struct {
	QualifiedType string
	FuncName      string
	FieldName     string
}

type fieldKeepFunc struct {
	JSONName   string
	AppendStmt string

	fieldName string // is private
	funcName  string // is private
	isObject  bool
}

type generateParams struct {
	PackageName      string
	Imports          []string
	TypeAndNewFuncs  []typeAndNewFunc
	KeepFuncs        []keepFunc
	KeepFuncsForImpl []keepFunc
}

func mapSlice[A any, B any](input []A, fn func(a A) B) []B {
	result := make([]B, 0, len(input))
	for _, e := range input {
		result = append(result, fn(e))
	}
	return result
}

func filterSlice[T any](input []T, fn func(a T) bool) []T {
	result := make([]T, 0, len(input))
	for _, e := range input {
		pred := fn(e)
		if !pred {
			continue
		}
		result = append(result, e)
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

func getKeepFuncStmt(funcName string, parentField string) string {
	result := fmt.Sprintf(`
isSimpleField = false
keepFunc, err := %s(field.SubFields)
if err != nil {
	return nil, fields.PrependParentField(err, "%s")
}
`, funcName, parentField)
	return strings.TrimSpace(result)
}

func appendStmtForObject(obj *objectInfo, field objectField) string {
	objectType := getQualifiedTypeName(obj)
	subObjectType := getQualifiedTypeName(field.info)
	funcName := getComputeKeepFuncName(field.info)

	result := fmt.Sprintf(`
%s
subFuncs = append(subFuncs, func (newMsg *%s, msg *%s) {
	newSubMsg := &%s{}
	keepFunc(newSubMsg, msg.%s)
	newMsg.%s = newSubMsg
})
`,
		getKeepFuncStmt(funcName, field.jsonName),
		objectType, objectType,
		subObjectType,
		field.name, field.name,
	)

	return strings.TrimSpace(result)
}

func appendStmtForArrayOfObjects(obj *objectInfo, field objectField) string {
	objectType := getQualifiedTypeName(obj)
	subObjectType := getQualifiedTypeName(field.info)
	funcName := getComputeKeepFuncName(field.info)

	result := fmt.Sprintf(`
%s
subFuncs = append(subFuncs, func(newMsg *%s, msg *%s) {
	msgList := make([]*%s, 0, len(msg.%s))
	for _, e := range msg.%s {
		newSubMsg := &%s{}
		keepFunc(newSubMsg, e)
		msgList = append(msgList, newSubMsg)
	}
	newMsg.%s = msgList
})
`,
		getKeepFuncStmt(funcName, field.jsonName),
		objectType, objectType,
		subObjectType,
		field.name, field.name,
		subObjectType, field.name,
	)

	return strings.TrimSpace(result)
}

func buildKeepFuncForField(info *objectInfo, subField objectField) fieldKeepFunc {
	funcName := fmt.Sprintf("%s_%s_Keep_%s", info.alias, info.typeName, subField.name)
	isObject := false

	var appendStmt string
	switch subField.fieldType {
	case fieldTypeObject:
		appendStmt = appendStmtForObject(info, subField)
		isObject = true

	case fieldTypeArrayOfObjects:
		appendStmt = appendStmtForArrayOfObjects(info, subField)
		isObject = true

	default:
		appendStmt = fmt.Sprintf("subFuncs = append(subFuncs, %s)", funcName)
	}

	return fieldKeepFunc{
		JSONName:   subField.jsonName,
		AppendStmt: appendStmt,

		fieldName: subField.name,
		funcName:  funcName,
		isObject:  isObject,
	}
}

func buildFieldFuncImplList(info *objectInfo, fieldFuncs []fieldKeepFunc) []fieldFuncImpl {
	implFuncs := make([]fieldFuncImpl, 0)
	for _, fn := range fieldFuncs {
		if fn.isObject {
			continue
		}
		implFuncs = append(implFuncs, fieldFuncImpl{
			QualifiedType: getQualifiedTypeName(info),
			FieldName:     fn.fieldName,
			FuncName:      fn.funcName,
		})
	}
	return implFuncs
}

func buildKeepFunc(info *objectInfo) keepFunc {
	fieldFuncs := mapSlice(info.subFields, func(subField objectField) fieldKeepFunc {
		return buildKeepFuncForField(info, subField)
	})

	return keepFunc{
		TypeName:       info.typeName,
		FuncName:       getComputeKeepFuncName(info),
		FuncType:       getFuncTypeSignature(info),
		FieldFuncs:     fieldFuncs,
		FieldFuncImpls: buildFieldFuncImplList(info, fieldFuncs),
	}
}

func writeToTemplate(writer io.Writer, templateStr string, params any) {
	tmpl, err := template.New("fieldmask").Parse(templateStr)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, params)
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

func generateCode(
	writer io.Writer, inputInfos []*objectInfo,
	packageName string,
) {
	infos := traverseAllObjectInfos(inputInfos)

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

	imports := computeImports(infos)

	typeAndNewFuncs := mapSlice(inputOnlyInfos, func(e *objectInfo) typeAndNewFunc {
		return typeAndNewFunc{
			StructName:          e.typeName + "FieldMask",
			FuncType:            getFuncTypeSignature(e),
			ComputeKeepFuncName: getComputeKeepFuncName(e),
			QualifiedType:       getQualifiedTypeName(e),
		}
	})

	keepFuncs := mapSlice(infos, buildKeepFunc)
	keepFuncsForImpl := filterSlice(keepFuncs, func(a keepFunc) bool {
		return len(a.FieldFuncImpls) > 0
	})

	params := generateParams{
		PackageName:      packageName,
		Imports:          imports,
		TypeAndNewFuncs:  typeAndNewFuncs,
		KeepFuncs:        keepFuncs,
		KeepFuncsForImpl: keepFuncsForImpl,
	}

	writeToTemplate(writer, fieldmaskTemplateString, params)
}

// Generate ...
func Generate(
	fileName string,
	protoMessages []ProtoMessage,
	packageName string,
) {
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}

	generateCode(file, parseMessages(protoMessages...), packageName)

	err = file.Close()
	if err != nil {
		panic(err)
	}
}
