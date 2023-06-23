package fieldmask

import (
	"bytes"
	_ "embed"
	"github.com/QuangTung97/fieldmask/testdata/pb"
	"github.com/stretchr/testify/assert"
	"testing"
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
