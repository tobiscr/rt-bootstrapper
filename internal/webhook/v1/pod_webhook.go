/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	apiv1 "github.com/kyma-project/rt-bootstrapper/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// SetupPodWebhookWithManager registers the webhook for Pod in the manager.
func SetupPodWebhookWithManager(mgr ctrl.Manager, cfg *apiv1.Config) error {

	slog.Info("setting up webhook", "cfg", cfg)

	d1 := BuildPodDefaulterAddImagePullSecrets(cfg.ImagePullSecretName)
	d2 := BuildPodDefaulterAlterImgRegistry(cfg.RegistryName)

	getNamespace := func(ctx context.Context, name string) (map[string]string, error) {
		var result corev1.Namespace
		if err := mgr.GetClient().Get(ctx, client.ObjectKey{
			Name: name,
		}, &result); err != nil {
			return nil, err
		}
		return result.Annotations, nil
	}

	defaulter := podCustomDefaulter{
		defaulters: []func(*corev1.Pod, map[string]string) error{
			d1,
			d2,
		},
		GetNsAnnotations: getNamespace,
		kymaNamespaces:   cfg.Scope.Namespaces,
		activeFeatures:   cfg.Scope.Annotations(),
	}

	return ctrl.NewWebhookManagedBy(mgr).For(&corev1.Pod{}).
		WithDefaulter(&defaulter).
		Complete()
}

// +kubebuilder:webhook:path=/mutate--v1-pod,mutating=true,failurePolicy=fail,sideEffects=None,groups="",resources=pods,verbs=create,versions=v1,name=mpod-v1.kb.io,admissionReviewVersions=v1

// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups="admissionregistration.k8s.io",resources=mutatingwebhookconfigurations,verbs=get;patch

// podCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Pod when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type podCustomDefaulter struct {
	defaulters []func(*corev1.Pod, map[string]string) error
	GetNsAnnotations
	kymaNamespaces []string
	activeFeatures map[string]string
}

var _ webhook.CustomDefaulter = &podCustomDefaulter{}

type GetNsAnnotations = func(context.Context, string) (map[string]string, error)

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Pod.
func (d *podCustomDefaulter) Default(ctx context.Context, obj runtime.Object) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			// no panic
			return
		}

		switch x := r.(type) {
		case string:
			err = fmt.Errorf("%s", x)
		case error:
			err = x
		default:
			err = fmt.Errorf("unknown defaulting function panic: %s", r)
		}
	}()

	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return fmt.Errorf("expected an Pod object but got %T", obj)
	}

	nsAnnotations := d.activeFeatures

	if !slices.Contains(d.kymaNamespaces, pod.Namespace) {
		nsAnnotations, err = d.GetNsAnnotations(ctx, pod.Namespace)
		if err != nil {
			slog.Error("unable to get namespace", "error", err)
			return err
		}
	}

	for i, defaulter := range d.defaulters {
		kvals := keysAndValues(pod)
		slog.With(kvals...).Debug("invoking defaulter", "i", fmt.Sprintf("%d", i))
		if err := defaulter(pod, nsAnnotations); err != nil {
			return err
		}
	}
	return nil
}

func keysAndValues(pod *corev1.Pod) []any {
	return []any{
		"name", pod.GetGenerateName(),
		"ns", pod.GetNamespace(),
		"labels", pod.GetLabels(),
		"annotations", pod.GetAnnotations(),
	}
}
