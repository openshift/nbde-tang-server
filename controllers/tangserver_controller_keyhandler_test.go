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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	daemonsv1alpha1 "github.com/openshift/nbde-tang-server/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("TangServer controller keyhandler", func() {

	// Define utility constants for object names
	const (
		TangserverName = "test-tangserver-keyhandler"
		// TODO: test why it can not be tested in non default namespace
		TangserverNamespace       = "default"
		TangserverResourceVersion = "1"
		TangServerTestKeyPath     = "/var/db/tang2"
	)

	Context("When Creating TangServer", func() {
		It("Should be created with default key path value", func() {
			By("By creating a new TangServer with empty key path value")
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
			Expect(getDefaultKeyPath(tangServer), DEFAULT_DEPLOYMENT_KEY_PATH)
			err := k8sClient.Delete(ctx, tangServer)
			Expect(err, nil)
		})
		It("Should be created with default script value", func() {
			By("By creating a new TangServer with empty image specs")
			ctx := context.Background()
			tangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TangserverName,
					Namespace: TangserverNamespace,
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					Replicas: 1,
					KeyPath:  TangServerTestKeyPath,
				},
			}
			Expect(k8sClient.Create(ctx, tangServer)).Should(Succeed())
			Expect(getDefaultKeyPath(tangServer), TangServerTestKeyPath)
			err := k8sClient.Delete(ctx, tangServer)
			Expect(err, nil)
		})
	})
})
