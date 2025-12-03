package v1

import (
	"log/slog"
	"slices"

	"github.com/kyma-project/rt-bootstrapper/internal/webhook/k8s"
	apiv1 "github.com/kyma-project/rt-bootstrapper/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
)

type PodDefaulter = func(p *corev1.Pod, nsAnnotations map[string]string) error

var (
	annotationsAlterImgRegistry = map[string]string{
		apiv1.AnnotationAlterImgRegistry: "false",
	}
	annotationsSetPullSecret = map[string]string{
		apiv1.AnnotationSetPullSecret: "false",
	}
)

func defaultPod(update func(*corev1.Pod), features map[string]string) PodDefaulter {
	return func(p *corev1.Pod, nsAnnotations map[string]string) error {
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
				return nil
			}
		}

		logger.Debug("pod defaulting opt in")
		update(p)
		return nil
	}
}

func BuildPodDefaulterAlterImgRegistry(overrides map[string]string) PodDefaulter {
	alterPodImageRegistry := func(p *corev1.Pod) {
		for i := range p.Spec.Containers {
			slog.With("overrides", overrides).Debug("altering containter image")
			p.Spec.Containers[i].Image = k8s.AlterPodImageRegistry(
				p.Spec.Containers[i].Image,
				overrides)
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
