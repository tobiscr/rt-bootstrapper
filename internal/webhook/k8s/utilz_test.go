package k8s_test

import (
	"testing"

	"github.com/kyma-project/rt-bootstrapper/internal/webhook/k8s"
)

func TestReplaceImageRegistry(t *testing.T) {
	tests := []struct {
		name     string
		image    string
		registry string
		expected string
	}{
		{
			name:     "Image without registry",
			image:    "nginx:latest",
			registry: "myregistry.com",
			expected: "myregistry.com/nginx:latest",
		},
		{
			name:     "Image with custom registry",
			image:    "custom.registry.com/myapp/image:v1.0",
			registry: "myregistry.com",
			expected: "myregistry.com/myapp/image:v1.0",
		},
		{
			name:     "Image with multiple slashes with no registry",
			image:    "namespace/subnamespace/image:tag",
			registry: "myregistry.com",
			expected: "myregistry.com/namespace/subnamespace/image:tag",
		},
		{
			name:     "Image with port in registry",
			image:    "registry:5000/myapp/image:v2.0",
			registry: "myregistry.com",
			expected: "myregistry.com/myapp/image:v2.0",
		},
		{
			name:     "Image with no tag",
			image:    "myapp/image",
			registry: "myregistry.com",
			expected: "myregistry.com/myapp/image",
		},
		{
			name:     "Image with localhost registry",
			image:    "localhost/myapp/image:v3.0",
			registry: "myregistry.com",
			expected: "myregistry.com/myapp/image:v3.0",
		},
		{
			name:     "Image with IP address registry",
			image:    "192.168.0.1/myapp/image:v4.0",
			registry: "myregistry.com",
			expected: "myregistry.com/myapp/image:v4.0",
		},
		{
			name:     "Image with multiple slashes",
			image:    "registry.com/namespace/subnamespace/image:tag",
			registry: "myregistry.com",
			expected: "myregistry.com/namespace/subnamespace/image:tag",
		},
		{
			name:     "Image with no namespace",
			image:    "image:tag",
			registry: "myregistry.com",
			expected: "myregistry.com/image:tag",
		},
		{
			name:     "Image with digest",
			image:    "myapp/image@sha256:abcdef1234567890",
			registry: "myregistry.com",
			expected: "myregistry.com/myapp/image@sha256:abcdef1234567890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := k8s.AlterPodImageRegistry(tt.image, tt.registry)
			if result != tt.expected {
				t.Errorf("ReplaceImageRegistry(%q, %q) = %q; want %q", tt.image, tt.registry, result, tt.expected)
			}
		})
	}
}
