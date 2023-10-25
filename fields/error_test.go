package fields

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrFieldNotFound(t *testing.T) {
	assert.Equal(t,
		"fieldmask: field not found or not allowed 'seller.name'",
		ErrFieldNotFound("seller.name").Error(),
	)

	assert.Equal(t,
		"fieldmask: field not found or not allowed 'provider.name'",
		PrependParentField(ErrFieldNotFound("name"), "provider").Error(),
	)
}
