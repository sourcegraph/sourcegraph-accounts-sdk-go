package sams

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Sanity check to make sure nothing blows up.
func TestNewIntrospectTransport(t *testing.T) {
	assert.NotNil(t, tokenServiceTransport)
}
