package v1

import (
	"encoding/json"
	"io"
)

const (
	AnnotationAlterImgRegistry = "rt-cfg.kyma-project.io/alter-img-registry"
	AnnotationSetPullSecret    = "rt-cfg.kyma-project.io/add-img-pull-secret"
)

type Scope struct {
	Namespaces []string `json:"namespaces"`
	Features   []string `json:"features"`
}

type Config struct {
	RegistryName        string `json:"registryName"`
	ImagePullSecretName string `json:"imagePullSecretName"`
	Scope               Scope  `json:"scope"`
}

func NewConfig(r io.Reader) (*Config, error) {
	var out Config
	err := json.NewDecoder(r).Decode(&out)
	return &out, err
}
