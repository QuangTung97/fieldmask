package fieldmask

import (
	"bytes"
	_ "embed"
	"github.com/QuangTung97/fieldmask/testdata/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

//go:embed testdata/fieldmap/product.go
var fieldMapGeneratedCode string

func TestGenerateFieldMap(t *testing.T) {
	var buf bytes.Buffer

	generateFieldMapCode(
		&buf, parseMessages(&pb.ProviderInfo{}, &pb.Product{}), "fieldmap",
		WithFieldMapRenameType(&pb.ProviderInfo{}, "ProviderData"),
	)

	assert.Equal(t, fieldMapGeneratedCode, buf.String())
}
