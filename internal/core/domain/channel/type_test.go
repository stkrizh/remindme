package channel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChannelType(t *testing.T) {
	assert.Equal(t, Unknown, Type{})
	assert.Equal(t, Unknown, Type{v: ""})
}
