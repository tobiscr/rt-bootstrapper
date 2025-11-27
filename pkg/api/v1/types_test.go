package v1_test

import (
	"strings"
	"testing"

	v1 "github.com/kyma-project/rt-bootstrapper/pkg/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	tcs := []struct {
		name     string
		val      string
		expected v1.Config
	}{
		{
			name: "with scope",
			val: `{ 
  "registryName": "rn1",
  "imagePullSecretName": "ipsn1",
  "scope": { 
    "namespaces": ["ns1"], 
    "features": ["f1", "f2"]
  }
}`,
			expected: v1.Config{
				RegistryName:        "rn1",
				ImagePullSecretName: "ipsn1",
				Scope: v1.Scope{
					Namespaces: []string{"ns1"},
					Features:   []string{"f1", "f2"},
				},
			},
		},
		{
			name: "without scope",
			val: `{
  "registryName": "rn2",
  "imagePullSecretName": "ipsn2"
}`,
			expected: v1.Config{
				RegistryName:        "rn2",
				ImagePullSecretName: "ipsn2",
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			r := strings.NewReader(tc.val)
			actual, err := v1.NewConfig(r)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, *actual)
		})
	}
}
