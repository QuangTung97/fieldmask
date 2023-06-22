package fieldmask

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestComputeFieldInfos(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		infos, err := ComputeFieldInfos(nil)
		assert.Equal(t, nil, err)
		assert.Equal(t, 0, len(infos))
	})

	t.Run("single", func(t *testing.T) {
		infos, err := ComputeFieldInfos([]string{""})
		assert.Equal(t, nil, err)
		assert.Equal(t, 0, len(infos))
	})
}
