/*
Copyright 2021.

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

package controllers

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	daemonsv1alpha1 "github.com/openshift/nbde-tang-server/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("TangServer controller image info", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		TangserverName            = "test-tangserver-imageinfo"
		TangserverNamespace       = "default"
		TangserverResourceVersion = "1"
	)

	Context("When Creating TangServer", func() {
		It("Should be created with default value", func() {
			By("By creating a new TangServer with empty image specs")
			ctx := context.Background()
			tangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TangserverName,
					Namespace: TangserverNamespace,
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					Replicas: 1,
				},
			}
			Expect(k8sClient.Create(ctx, tangServer)).Should(Succeed())
			Expect(getImageNameAndVersion(tangServer), getCompleteImageNameAndVersion(DEFAULT_APP_IMAGE, DEFAULT_APP_VERSION))
			err := k8sClient.Delete(ctx, tangServer)
			Expect(err, nil)
		})
		It("Should be created with spec value", func() {
			ctx := context.Background()
			By("By creating a TangServer with image specs")
			testImage := "test"
			testVersion := "v0.1.2"
			tangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TangserverName,
					Namespace: TangserverNamespace,
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					Replicas: 1,
					Image:    testImage,
					Version:  testVersion,
				},
			}
			Expect(k8sClient.Create(ctx, tangServer)).Should(Succeed())
			Expect(getImageNameAndVersion(tangServer), getCompleteImageNameAndVersion(testImage, testVersion))
			err := k8sClient.Delete(ctx, tangServer)
			Expect(err, nil)
		})
		It("Should be created with version spec value", func() {
			ctx := context.Background()
			By("By creating a TangServer with image specs")
			testVersion := "v0.1.3"
			tangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TangserverName,
					Namespace: TangserverNamespace,
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					Replicas: 1,
					Version:  testVersion,
				},
			}
			Expect(k8sClient.Create(ctx, tangServer)).Should(Succeed())
			Expect(getImageNameAndVersion(tangServer), getCompleteImageNameAndVersion(DEFAULT_APP_IMAGE, testVersion))
			err := k8sClient.Delete(ctx, tangServer)
			Expect(err, nil)
		})
		It("Should be created with image spec value", func() {
			ctx := context.Background()
			By("By creating a TangServer with image specs")
			testImage := "test"
			tangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TangserverName,
					Namespace: TangserverNamespace,
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					Replicas: 1,
					Image:    testImage,
				},
			}
			Expect(k8sClient.Create(ctx, tangServer)).Should(Succeed())
			Expect(getImageNameAndVersion(tangServer), getCompleteImageNameAndVersion(testImage, DEFAULT_APP_VERSION))
			err := k8sClient.Delete(ctx, tangServer)
			Expect(err, nil)
		})
	})
})
