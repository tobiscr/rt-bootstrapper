package k8s_test

import (
	"testing"

	"github.com/kyma-project/rt-bootstrapper/internal/webhook/k8s"
	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	tcs := []struct {
		name     string
		l, r     map[string]string
		expected bool
	}{
		{
			name: "should be found #1",
			l: map[string]string{
				"some": "test",
				"test": "me",
			},
			r: map[string]string{
				"test": "me",
			},
			expected: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual := k8s.Contains(tc.l, tc.r)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestReplaceImageRegistry(t *testing.T) {
	tests := []struct {
		name      string
		image     string
		overrides map[string]string
		expected  string
	}{
		{
			name:      "Image without registry",
			image:     "nginx:latest",
			overrides: map[string]string{"nginx": "should-not-be-replaced"},
			expected:  "nginx:latest",
		},
		{
			name:      "Image with custom registry",
			image:     "custom.registry.com/myapp/image:v1.0",
			overrides: map[string]string{"custom.registry.com": "myregistry.com"},
			expected:  "myregistry.com/myapp/image:v1.0",
		},
		{
			name:      "Image with multiple slashes with no registry",
			image:     "namespace/subnamespace/image:tag",
			overrides: map[string]string{"namespace": "should-not-be-replaced"},
			expected:  "namespace/subnamespace/image:tag",
		},
		{
			name:      "Image with port in registry",
			image:     "registry:5000/myapp/image:v2.0",
			overrides: map[string]string{"registry:5000": "myregistry.com"},
			expected:  "myregistry.com/myapp/image:v2.0",
		},
		{
			name:      "Image with no tag",
			image:     "myapp/image",
			overrides: map[string]string{"myapp": "should-not-be-replaced"},
			expected:  "myapp/image",
		},
		{
			name:      "Image with localhost registry",
			image:     "localhost/myapp/image:v3.0",
			overrides: map[string]string{"localhost": "should-not-be-replaced"},
			expected:  "localhost/myapp/image:v3.0",
		},
		{
			name:      "Image with IP address registry",
			image:     "192.168.0.1/myapp/image:v4.0",
			overrides: map[string]string{"192.168.0.1": "myregistry.com"},
			expected:  "myregistry.com/myapp/image:v4.0",
		},
		{
			name:      "Image with multiple slashes",
			image:     "registry.com/namespace/subnamespace/image:tag",
			overrides: map[string]string{"registry.com": "myregistry.com"},
			expected:  "myregistry.com/namespace/subnamespace/image:tag",
		},
		{
			name:      "Image with digest",
			image:     "myapp/image@sha256:abcdef1234567890",
			overrides: map[string]string{"myapp": "should-not-be-replaced"},
			expected:  "myapp/image@sha256:abcdef1234567890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := k8s.AlterPodImageRegistry(tt.image, tt.overrides)
			if result != tt.expected {
				t.Errorf("ReplaceImageRegistry(%q, %q) = %q; want %q", tt.image, tt.overrides, result, tt.expected)
			}
		})
	}
}
