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

func getTestPod(labels map[string]string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
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
	var defaulter = newPodCustomDefaulter(testRegistryName, testPullSecret)

	Context("When creating Pod under Defaulting Webhook", func() {
		It("Should alter image registry", func() {
			By(fmt.Sprintf("adding '%s' label", LabeAlterImgRegistry))
			pod := getTestPod(map[string]string{LabeAlterImgRegistry: "true"})
			Expect(pod.Spec.Containers).ShouldNot(BeEmpty())

			By("calling the Default method to alter registry image")
			defaulter.Default(ctx, pod)

			By("checking that the image was altered")
			for _, container := range pod.Spec.Containers {
				Expect(container.Image).Should(HavePrefix(testRegistryName))
			}
		})
		It("Should add image pull secret", func() {
			By(fmt.Sprintf("adding '%s' label", LabeSetPullSecret))
			pod := getTestPod(map[string]string{LabeSetPullSecret: "true"})
			Expect(pod.Spec.Containers).ShouldNot(BeEmpty())

			By("calling the Default method to add pull secret")
			defaulter.Default(ctx, pod)

			By(fmt.Sprintf("checking that the pod's image pull secrets contain '%s'", testPullSecret))
			Expect(pod.Spec.ImagePullSecrets).Should(ContainElement(
				corev1.LocalObjectReference{Name: testPullSecret},
			))
		})
	})
})
