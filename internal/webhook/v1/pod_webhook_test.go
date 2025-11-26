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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testRegistryName = "test-registry"
	testPullSecret   = "test-pull-secret"
)

func getTestPod(labels, annotations map[string]string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: annotations,
			Labels:      labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image: "test/me/plz:now",
				},
				{
					Image: "test/this/too:plz",
				},
			},
		},
	}
}

var _ = Describe("Pod Webhook", func() {

	Context("When creating Pod under Defaulting Webhook", func() {
		d1 := BuildPodDefaulterAddImagePullSecrets(testPullSecret)
		d2 := BuildPodDefaulterAlterImgRegistry(testRegistryName)

		var defaulter = podCustomDefaulter{
			defaulters: []func(*corev1.Pod, map[string]string, map[string]string) error{
				d1, d2,
			},
			GetNamespace: func(_ context.Context, name string) (*corev1.Namespace, error) {
				return &corev1.Namespace{}, nil
			},
		}

		It("Should alter image registry", func() {
			By(fmt.Sprintf("adding '%s' annotation", AnnotationAlterImgRegistry))
			pod := getTestPod(
				map[string]string{LabelRtBootstrapperCfg: "true"},
				map[string]string{AnnotationAlterImgRegistry: "true"})
			Expect(pod.Spec.Containers).ShouldNot(BeEmpty())

			By("calling the Default method to alter registry image")
			err := defaulter.Default(ctx, pod)
			Expect(err).ShouldNot(HaveOccurred())

			By("checking that the image was altered")
			for _, container := range pod.Spec.Containers {
				Expect(container.Image).Should(HavePrefix(testRegistryName))
			}
		})

		It("Should add image pull secret", func() {
			By(fmt.Sprintf("adding '%s' label", AnnotationSetPullSecret))
			pod := getTestPod(
				map[string]string{LabelRtBootstrapperCfg: "true"},
				map[string]string{AnnotationSetPullSecret: "true"})
			Expect(pod.Spec.Containers).ShouldNot(BeEmpty())

			By("calling the Default method to add pull secret")
			err := defaulter.Default(ctx, pod)
			Expect(err).ShouldNot(HaveOccurred())

			By(fmt.Sprintf("checking that the pod's image pull secrets contain '%s'", testPullSecret))
			Expect(pod.Spec.ImagePullSecrets).Should(ContainElement(
				corev1.LocalObjectReference{Name: testPullSecret},
			))
		})
	})
})
