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
			name: "all",
			val: `{ 
  "imagePullSecretName": "ipsn1",
  "imagePullSecretNamespace": "ipsns1",
  "overrides": { "rn1": "orn1" }
}`,
			expected: v1.Config{
				Overrides: map[string]string{
					"rn1": "orn1",
				},
				ImagePullSecretName:      "ipsn1",
				ImagePullSecretNamespace: "ipsns1",
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
