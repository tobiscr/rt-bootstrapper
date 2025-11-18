package v1

import (
	"slices"

	"github.com/kyma-project/rt-bootstrapper/internal/webhook/k8s"
	corev1 "k8s.io/api/core/v1"
)

type PodDefaulter = func(*corev1.Pod)

func BuildPodDefaulterAlterImageRegistry(registryName string) PodDefaulter {
	return func(p *corev1.Pod) {
		for i := range p.Spec.Containers {
			p.Spec.Containers[i].Image = k8s.ReplacePodImageRegistry(p.Spec.Containers[i].Image, registryName)
		}
	}
}

func BuildPodDefaulterSetImagePullSecrets(secretName string) PodDefaulter {
	return func(p *corev1.Pod) {
		imgPullSecret := corev1.LocalObjectReference{Name: secretName}
		if slices.Contains(p.Spec.ImagePullSecrets, imgPullSecret) {
			return
		}

		p.Spec.ImagePullSecrets = append(p.Spec.ImagePullSecrets, imgPullSecret)
	}
}
