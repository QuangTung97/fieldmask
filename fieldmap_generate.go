package fieldmask

import (
	_ "embed"
	"io"
	"os"
)

type fieldMapGenerateParams struct {
	PackageName string
	Structs     []fieldMapStruct
}

type fieldMapStruct struct {
	StructName string
	Fields     []fieldMapStructField
}

type fieldMapStructField struct {
	Name      string
	FieldType string
	JSONTag   string
}

func computeFieldMapStructName(e *objectInfo) string {
	name := e.opts.withFieldMapTypeName(e.typeName)
	return name + "FieldMap"
}

func buildFieldMapStructField(f objectField) fieldMapStructField {
	typeValue := "Field"
	if f.info != nil {
		typeValue = computeFieldMapStructName(f.info)
	}
	return fieldMapStructField{
		Name:      f.name,
		FieldType: typeValue,
		JSONTag:   f.jsonName,
	}
}

func buildFieldMapStructs(infos []*objectInfo) []fieldMapStruct {
	return mapSlice(infos, func(e *objectInfo) fieldMapStruct {
		structName := computeFieldMapStructName(e)

		return fieldMapStruct{
			StructName: structName,
			Fields: mapSlice(e.subFields, func(f objectField) fieldMapStructField {
				return buildFieldMapStructField(f)
			}),
		}
	})
}

//go:embed fieldmap_template
var fieldMapTemplateString string

func generateFieldMapCode(
	writer io.Writer, inputInfos []*objectInfo,
	packageName string,
) {
	infos := traverseAllObjectInfos(inputInfos)

	params := fieldMapGenerateParams{
		PackageName: packageName,
		Structs:     buildFieldMapStructs(infos),
	}
	writeToTemplate(writer, fieldMapTemplateString, params)
}

// GenerateFieldMap ...
func GenerateFieldMap(
	fileName string,
	protoMessages []ProtoMessage,
	packageName string,
) {
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}

	generateFieldMapCode(file, parseMessages(protoMessages...), packageName)

	err = file.Close()
	if err != nil {
		panic(err)
	}
}
