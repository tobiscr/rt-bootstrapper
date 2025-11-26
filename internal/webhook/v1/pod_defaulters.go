package v1

import (
	"log/slog"
	"slices"

	"github.com/kyma-project/rt-bootstrapper/internal/webhook/k8s"
	corev1 "k8s.io/api/core/v1"
)

type PodDefaulter = func(p *corev1.Pod, nsAnnotations map[string]string) error

const (
	AnnotationAlterImgRegistry = "rt-cfg.kyma-project.io/alter-img-registry"
	AnnotationSetPullSecret    = "rt-cfg.kyma-project.io/add-img-pull-secret"
)

var (
	annotationsAlterImgRegistry = map[string]string{
		AnnotationAlterImgRegistry: "true",
	}
	annotationsSetPullSecret = map[string]string{
		AnnotationSetPullSecret: "true",
	}
)

func defaultPod(update func(*corev1.Pod), expectedAnnotations map[string]string) PodDefaulter {
	return func(p *corev1.Pod, nsAnnotations map[string]string) error {
		kvs := keysAndValues(p)
		logger := slog.With(kvs...)

		for _, annotations := range []map[string]string{p.Annotations, nsAnnotations} {
			logger.Debug("defaulting pod", "expected-annotations", expectedAnnotations)
			if k8s.Contains(annotations, expectedAnnotations) {
				logger.Debug("expected annotations found", "labels", annotations)
				update(p)
				return nil
			}
		}

		logger.Debug("ignoring pod", "expected-labels", expectedAnnotations)
		return nil
	}
}

func BuildPodDefaulterAlterImgRegistry(registryName string) PodDefaulter {
	alterPodImageRegistry := func(p *corev1.Pod) {
		for i := range p.Spec.Containers {
			slog.With("registry-name", registryName).Debug("altering containter image")
			p.Spec.Containers[i].Image = k8s.AlterPodImageRegistry(p.Spec.Containers[i].Image, registryName)
		}
	}

	return defaultPod(alterPodImageRegistry, annotationsAlterImgRegistry)
}

func BuildPodDefaulterAddImagePullSecrets(secretName string) PodDefaulter {
	addImgPullSecret := func(p *corev1.Pod) {
		imgPullSecret := corev1.LocalObjectReference{Name: secretName}
		if slices.Contains(p.Spec.ImagePullSecrets, imgPullSecret) {
			slog.Debug("image pull secret already found")
			return
		}

		slog.Debug("adding new image pull secret")
		p.Spec.ImagePullSecrets = append(p.Spec.ImagePullSecrets, imgPullSecret)
	}

	return defaultPod(addImgPullSecret, annotationsSetPullSecret)
}
