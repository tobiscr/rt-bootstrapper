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
