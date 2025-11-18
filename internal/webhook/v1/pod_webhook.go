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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var podlog = logf.Log.WithName("pod-resource")

// SetupPodWebhookWithManager registers the webhook for Pod in the manager.
func SetupPodWebhookWithManager(mgr ctrl.Manager, registryName string, imagePullSecretName string) error {
	defaulter1 := BuildPodDefaulterAlterImageRegistry(registryName)
	defaulter2 := BuildPodDefaulterSetImagePullSecrets(imagePullSecretName)

	return ctrl.NewWebhookManagedBy(mgr).For(&corev1.Pod{}).
		WithDefaulter(&podCustomDefaulter{
			defaulters: []func(*corev1.Pod){
				defaulter1,
				defaulter2,
			},
		}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate--v1-pod,mutating=true,failurePolicy=fail,sideEffects=None,groups="",resources=pods,verbs=create,versions=v1,name=mpod-v1.kb.io,admissionReviewVersions=v1

// podCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Pod when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type podCustomDefaulter struct {
	defaulters []func(*corev1.Pod)
}

var _ webhook.CustomDefaulter = &podCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Pod.
func (d *podCustomDefaulter) Default(_ context.Context, obj runtime.Object) (err error) {
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

	kvals := keysAndValues(pod)
	podlog.Info("defaulting pod", kvals...)

	for _, defaulter := range d.defaulters {
		podlog.Info("defaulting pod", kvals...)
		defaulter(pod)
	}
	return nil
}

func keysAndValues(pod *corev1.Pod) []any {
	return []any{
		"name", pod.GetName(),
		"ns", pod.GetNamespace(),
		"uuid", pod.GetUID(),
		"labels", pod.GetLabels(),
	}
}
