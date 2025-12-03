package v1

import (
	"encoding/json"
	"io"
)

const (
	AnnotationAlterImgRegistry = "rt-cfg.kyma-project.io/alter-img-registry"
	AnnotationSetPullSecret    = "rt-cfg.kyma-project.io/add-img-pull-secret"
	AnnotationOutdated         = "rt-cfg.kyma-project.io/outdated"
	FiledManager               = "rt-bootstrapper"
)

type Config struct {
	Overrides           map[string]string `json:"overrides"`
	ImagePullSecretName string            `json:"imagePullSecretName"`
}

func NewConfig(r io.Reader) (*Config, error) {
	var out Config
	err := json.NewDecoder(r).Decode(&out)
	return &out, err
}
