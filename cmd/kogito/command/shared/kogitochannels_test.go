package shared

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_IsChannelValidForAlphaChannel(t *testing.T) {
	ch := "alpha"
	isValid := IsChannelValid(ch)
	assert.True(t, isValid)
}

func Test_IsChannelValidForDevPreviewChannel(t *testing.T) {
	ch := "dev-preview"
	isValid := IsChannelValid(ch)
	assert.True(t, isValid)
}

func Test_IsChannelValidForInvalidInput(t *testing.T) {
	ch := "testChannel"
	isValid := IsChannelValid(ch)
	assert.False(t, isValid)
}

func Test_GetDefaultChannel(t *testing.T) {
	ch := GetDefaultChannel()
	assert.Equal(t, AlphaChannel, ch)
}
