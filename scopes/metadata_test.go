package scopes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenScopeToMetadataScope(t *testing.T) {
	for _, tc := range []struct {
		permission Permission
		expected   string
	}{
		{"user.metadata", "*"},
		{"user.metadata.cody", "cody"},
		{"user.metadata.dotcom", "dotcom"},
		{"user.metadata.cody_gatekeeper", "cody_gatekeeper"},
	} {
		got, err := PermissionToMetadataScope(tc.permission)
		require.NoError(t, err)
		assert.Equal(t, tc.expected, got)
	}

	t.Run("invalid permission", func(t *testing.T) {
		_, err := PermissionToMetadataScope("not-a-metadata-permission")
		assert.Error(t, err)
	})
}
