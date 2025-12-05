package k8s

import (
	"fmt"
	"strings"
)

func fixRegistry(registry string, overrides map[string]string) string {
	override, found := overrides[registry]
	if !found {
		return registry
	}
	return override
}

func AlterPodImageRegistry(image string, overrides map[string]string) string {
	data := strings.Split(image, "/")

	isRegistryProvided := len(data) > 1 &&
		(strings.Contains(data[0], ".") || strings.Contains(data[0], ":"))

	if isRegistryProvided {
		registry := fixRegistry(data[0], overrides)
		return fmt.Sprintf("%s/%s", registry, strings.Join(data[1:], "/"))
	}

	return image
}

// Contains - returns true if l contains all the keys with values from r
// returns false otherwise
func Contains(l map[string]string, r map[string]string) bool {
	for k, v := range r {
		val, found := l[k]
		if !found || val != v {
			return false
		}
	}
	return true
}
