package fieldmask

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/QuangTung97/fieldmask/testdata/pb"
)

//go:embed testdata/generated/provider.go
var generatedCode string

func TestGenerate(t *testing.T) {
	var buf bytes.Buffer

	generateCode(&buf, parseMessages(
		NewProtoMessage(&pb.ProviderInfo{}),
		NewProtoMessage(&pb.Product{}),
	), "generated")

	assert.Equal(t, generatedCode, buf.String())
}

//go:embed testdata/generated/limited/provider.go
var generatedCodeWithLimitedFields string

func TestGenerate_WithLimitedTo_And_WithFieldMaskName(t *testing.T) {
	var buf bytes.Buffer

	generateCode(&buf, parseMessages(
		NewProtoMessage(&pb.ProviderInfo{}),
		NewProtoMessageWithFields(&pb.Product{}, []string{
			"sku",
			"provider",
			"attributes.options.code",
			"stocks",
		}),
	), "generated")

	assert.Equal(t, generatedCodeWithLimitedFields, buf.String())
}
