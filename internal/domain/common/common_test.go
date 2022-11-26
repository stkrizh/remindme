package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOptional(t *testing.T) {
	assert := require.New(t)

	optionalInt := NewOptional(42, true)
	assert.Equal(42, optionalInt.Value)
	assert.True(optionalInt.IsPresent)

	optionalString := NewOptional("foo", false)
	assert.Equal("foo", optionalString.Value)
	assert.False(optionalString.IsPresent)
}
