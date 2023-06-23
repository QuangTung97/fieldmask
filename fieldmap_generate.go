package fieldmask

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/golang/protobuf/proto"
	"go/format"
	"io"
	"os"
	"reflect"
	"text/template"
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

func computeFieldMapStructName(opts generateFieldMapOptions, e *objectInfo) string {
	name := e.typeName
	if newName, ok := opts.renamed[e.getKey()]; ok {
		name = newName
	}
	return name + "FieldMap"
}

func buildFieldMapStructField(opts generateFieldMapOptions, f objectField) fieldMapStructField {
	typeValue := "Field"
	if f.info != nil {
		typeValue = computeFieldMapStructName(opts, f.info)
	}
	return fieldMapStructField{
		Name:      f.name,
		FieldType: typeValue,
		JSONTag:   f.jsonName,
	}
}

func buildFieldMapStructs(opts generateFieldMapOptions, infos []*objectInfo) []fieldMapStruct {
	return mapSlice(infos, func(e *objectInfo) fieldMapStruct {
		structName := computeFieldMapStructName(opts, e)

		return fieldMapStruct{
			StructName: structName,
			Fields: mapSlice(e.subFields, func(f objectField) fieldMapStructField {
				return buildFieldMapStructField(opts, f)
			}),
		}
	})
}

//go:embed fieldmap_template
var fieldMapTemplateString string

func generateFieldMapCode(
	writer io.Writer, inputInfos []*objectInfo,
	packageName string,
	options ...GenerateFieldMapOption,
) {
	opts := newGenerateFieldMapOptions(options)

	tmpl, err := template.New("fieldmask").Parse(fieldMapTemplateString)
	if err != nil {
		panic(err)
	}

	infos := traverseAllObjectInfos(inputInfos, map[objectKey]struct{}{})

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, fieldMapGenerateParams{
		PackageName: packageName,
		Structs:     buildFieldMapStructs(opts, infos),
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

// GenerateFieldMap ...
func GenerateFieldMap(
	fileName string,
	protoMessages []ProtoMessage,
	packageName string,
	options ...GenerateFieldMapOption,
) {
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}

	generateFieldMapCode(file, parseMessages(protoMessages...), packageName, options...)

	err = file.Close()
	if err != nil {
		panic(err)
	}
}

type generateFieldMapOptions struct {
	renamed map[objectKey]string
}

func newGenerateFieldMapOptions(options []GenerateFieldMapOption) generateFieldMapOptions {
	opts := generateFieldMapOptions{
		renamed: map[objectKey]string{},
	}

	for _, fn := range options {
		fn(&opts)
	}
	return opts
}

// GenerateFieldMapOption ...
type GenerateFieldMapOption func(opts *generateFieldMapOptions)

// WithFieldMapRenameType ...
func WithFieldMapRenameType(msg proto.Message, newName string) GenerateFieldMapOption {
	return func(opts *generateFieldMapOptions) {
		msgType := reflect.TypeOf(msg).Elem()
		opts.renamed[objectKey{
			typeName:   msgType.Name(),
			importPath: msgType.PkgPath(),
		}] = newName
	}
}
