package v1

import (
	"log/slog"
	"reflect"
	"slices"

	"github.com/kyma-project/rt-bootstrapper/internal/webhook/k8s"
	apiv1 "github.com/kyma-project/rt-bootstrapper/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
)

type PodDefaulter = func(p *corev1.Pod, nsAnnotations map[string]string) (bool, error)

var (
	annotationsAlterImgRegistry = map[string]string{
		apiv1.AnnotationAlterImgRegistry: "true",
	}
	annotationsSetPullSecret = map[string]string{
		apiv1.AnnotationSetPullSecret: "true",
	}
	annotationAddClusterTrustBundle = map[string]string{
		apiv1.AnnotationAddClusterTrustBundle: "true",
	}
)

type updateOpts struct {
	activeAnnotations map[string]string
	namespaceFeatures apiv1.NamespaceFeatures
}

func defaultPod(update func(*corev1.Pod) bool, opts updateOpts) PodDefaulter {
	return func(p *corev1.Pod, nsAnnotations map[string]string) (bool, error) {
		// prepare logger
		kvs := keysAndValues(p)

		defaultFeatures := opts.namespaceFeatures.Features(p.Namespace)

		logger := slog.Default().
			WithGroup("args").
			With(kvs...).
			With("ns-annotations", nsAnnotations).
			With("active-features", opts.activeAnnotations).
			With("default-features", defaultFeatures)

		for _, annotations := range []map[string]string{defaultFeatures, nsAnnotations, p.Annotations} {
			if k8s.Contains(annotations, opts.activeAnnotations) {
				logger.Debug("pod defaulting opt in")
				return update(p), nil
			}
		}

		logger.Debug("opt out", "ns-annotations", nsAnnotations)
		return false, nil
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

func BuildPodDefaulterAlterImgRegistry(overrides map[string]string, nsf apiv1.NamespaceFeatures) PodDefaulter {
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

	return defaultPod(alterPodImageRegistry, updateOpts{
		activeAnnotations: annotationsAlterImgRegistry,
		namespaceFeatures: nsf,
	})
}

func BuildPodDefaulterAddImagePullSecrets(secretName string, nsf apiv1.NamespaceFeatures) PodDefaulter {
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

	return defaultPod(addImgPullSecret, updateOpts{
		activeAnnotations: annotationsSetPullSecret,
		namespaceFeatures: nsf,
	})
}

func BuildDefaulterAddClusterTrustBundle(mapping k8s.ClusterTrustBundle, nsf apiv1.NamespaceFeatures) PodDefaulter {
	slog.Debug("building volume", mapping.KeysAndValues()...)

	vol := mapping.ClusterTrustedBundle()

	handleVolumeMount := func(cs []corev1.Container) bool {
		// stores information if any container was modified
		var result bool

		for i, c := range cs {
			index := slices.IndexFunc(c.VolumeMounts, func(vm corev1.VolumeMount) bool {
				return vm.Name == mapping.VolumeName
			})

			if index == -1 {
				vm := mapping.VolumeMount()
				cs[i].VolumeMounts = append(c.VolumeMounts, vm)
				result = true
				slog.Debug("volume mount added")
				continue
			}

			if reflect.DeepEqual(c.VolumeMounts[index], vol) {
				slog.Debug("volume already mounted, nothing to do")
				continue
			}

			vm := mapping.VolumeMount()
			cs[i].VolumeMounts[index] = vm
			slog.Debug("volume mount replaced")
			result = true
		}

		return result
	}

	handleVolumeMounts := func(modified bool, p *corev1.Pod) bool {
		// stores information if any container was modified
		for _, cs := range [][]corev1.Container{p.Spec.Containers} {
			if handleVolumeMount(cs) {
				modified = true
			}
		}

		return modified
	}

	handleClusterTrustBundle := func(p *corev1.Pod) bool {
		index := slices.IndexFunc(p.Spec.Volumes, func(v corev1.Volume) bool {
			return v.Name == mapping.VolumeName
		})

		if index == -1 {
			// volume does not exist, add it
			p.Spec.Volumes = append(p.Spec.Volumes, vol)
			slog.Debug("volume added")
			return handleVolumeMounts(true, p)
		}

		if reflect.DeepEqual(p.Spec.Volumes[index], vol) {
			slog.Debug("volume already added, nothing to do")
			return handleVolumeMounts(false, p)
		}

		p.Spec.Volumes[index] = vol
		slog.Debug("volume replaced")

		return handleVolumeMounts(true, p)
	}

	return defaultPod(handleClusterTrustBundle, updateOpts{
		activeAnnotations: annotationAddClusterTrustBundle,
		namespaceFeatures: nsf,
	})
}
