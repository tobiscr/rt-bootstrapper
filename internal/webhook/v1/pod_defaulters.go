package v1

import (
	"slices"

	"github.com/kyma-project/rt-bootstrapper/internal/webhook/k8s"
	corev1 "k8s.io/api/core/v1"
)

type PodDefaulter = func(*corev1.Pod)

const (
	LabeAlterImgRegistry = "rt-cfg.kyma-project.io/alter-img-registry"
	LabeSetPullSecret    = "rt-cfg.kyma-project.io/add-img-pull-secret"
)

func BuildPodDefaulterAlterImgRegistry(registryName string) PodDefaulter {
	return func(p *corev1.Pod) {
		if value, found := p.Labels[LabeAlterImgRegistry]; !found || value != "true" {
			return
		}

		for i := range p.Spec.Containers {
			p.Spec.Containers[i].Image = k8s.AlterPodImageRegistry(p.Spec.Containers[i].Image, registryName)
		}
	}
}

func BuildPodDefaulterAddImagePullSecrets(secretName string) PodDefaulter {
	return func(p *corev1.Pod) {
		if value, found := p.Labels[LabeSetPullSecret]; !found || value != "true" {
			return
		}

		imgPullSecret := corev1.LocalObjectReference{Name: secretName}
		if slices.Contains(p.Spec.ImagePullSecrets, imgPullSecret) {
			return
		}

		p.Spec.ImagePullSecrets = append(p.Spec.ImagePullSecrets, imgPullSecret)
	}
}
