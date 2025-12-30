package v1

import (
	"log/slog"
	"slices"

	"github.com/kyma-project/rt-bootstrapper/internal/webhook/k8s"
	apiv1 "github.com/kyma-project/rt-bootstrapper/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
)

type PodDefaulter = func(p *corev1.Pod, nsAnnotations map[string]string) (bool, error)

var (
	annotationsAlterImgRegistry = map[string]string{
		apiv1.AnnotationAlterImgRegistry: "false",
	}
	annotationsSetPullSecret = map[string]string{
		apiv1.AnnotationSetPullSecret: "false",
	}
)

func defaultPod(update func(*corev1.Pod) bool, features map[string]string) PodDefaulter {
	return func(p *corev1.Pod, nsAnnotations map[string]string) (bool, error) {
		// prepare logger
		kvs := keysAndValues(p)
		logger := slog.Default().
			WithGroup("args").
			With(kvs...).
			With("ns-annotations", nsAnnotations).
			With("features", features)

		for _, annotations := range []map[string]string{p.Annotations, nsAnnotations} {
			if k8s.Contains(annotations, features) {
				logger.Debug("opt out", "ns-annotations", nsAnnotations)
				return false, nil
			}
		}

		logger.Debug("pod defaulting opt in")
		return update(p), nil
	}
}

func alterImgRegistry(containers []corev1.Container, overrides map[string]string) bool {
	var modified bool
	for i := range containers {
		alteredImage := k8s.AlterPodImageRegistry(
			containers[i].Image,
			overrides)

		if alteredImage == containers[i].Image {
			continue
		}

		containers[i].Image = alteredImage
		modified = true

		slog.With("overrides", overrides,
			"image-name", containers[i].Image,
			"container-name", containers[i].Name).Debug("image altered")
	}
	return modified
}

func BuildPodDefaulterAlterImgRegistry(overrides map[string]string) PodDefaulter {
	alterPodImageRegistry := func(p *corev1.Pod) bool {
		var modified bool
		for _, containers := range [][]corev1.Container{
			p.Spec.InitContainers,
			p.Spec.Containers,
		} {
			if !alterImgRegistry(containers, overrides) {
				continue
			}
			modified = true
		}
		return modified
	}

	return defaultPod(alterPodImageRegistry, annotationsAlterImgRegistry)
}

func BuildPodDefaulterAddImagePullSecrets(secretName string) PodDefaulter {
	addImgPullSecret := func(p *corev1.Pod) bool {
		imgPullSecret := corev1.LocalObjectReference{Name: secretName}
		if slices.Contains(p.Spec.ImagePullSecrets, imgPullSecret) {
			slog.Debug("image pull secret already found")
			return false
		}

		slog.Debug("adding new image pull secret")
		p.Spec.ImagePullSecrets = append(p.Spec.ImagePullSecrets, imgPullSecret)
		return true
	}

	return defaultPod(addImgPullSecret, annotationsSetPullSecret)
}
