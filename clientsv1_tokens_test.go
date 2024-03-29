package sams

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIntrospectTransport(t *testing.T) {
	assert.NotNil(t, tokenServiceTransport)
}
