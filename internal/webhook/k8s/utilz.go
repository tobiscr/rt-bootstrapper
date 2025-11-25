package k8s

import (
	"fmt"
	"strings"
)

func AlterPodImageRegistry(image string, registry string) string {
	data := strings.Split(image, "/")

	if len(data) == 1 {
		return fmt.Sprintf("%s/%s", registry, image)
	}

	if strings.Contains(data[0], ".") || strings.Contains(data[0], ":") || data[0] == "localhost" {
		return fmt.Sprintf("%s/%s", registry, strings.Join(data[1:], "/"))
	}

	return fmt.Sprintf("%s/%s", registry, image)
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
