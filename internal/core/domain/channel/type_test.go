package channel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChannelType(t *testing.T) {
	var type_ Type
	assert.Equal(t, Unknown, Type(""))
	assert.Equal(t, Unknown, type_)
}
