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
		TangserverName            = "test-tangserver-keyhandler"
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

	Context("When testing key handling utility functions", func() {
		var tangServer *daemonsv1alpha1.TangServer

		BeforeEach(func() {
			tangServer = &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TangserverName,
					Namespace: TangserverNamespace,
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					Replicas: 1,
					KeyPath:  TangServerTestKeyPath,
				},
			}
		})

		It("Should return correct key path", func() {
			Expect(getDefaultKeyPath(tangServer)).To(Equal(TangServerTestKeyPath))

			tangServer.Spec.KeyPath = ""
			Expect(getDefaultKeyPath(tangServer)).To(Equal(DEFAULT_DEPLOYMENT_KEY_PATH))
		})

		It("Should handle forbidden path map correctly", func() {
			_, exists := FORBIDDEN_PATH_MAP["."]
			Expect(exists).To(BeTrue())
			_, exists = FORBIDDEN_PATH_MAP[".."]
			Expect(exists).To(BeTrue())
			_, exists = FORBIDDEN_PATH_MAP["lost+found"]
			Expect(exists).To(BeTrue())

			updateForbiddenMap("testkey")
			_, exists = FORBIDDEN_PATH_MAP["testkey"]
			Expect(exists).To(BeTrue())
		})

		It("Should handle SHAType constants correctly", func() {
			Expect(int(UNKNOWN_SHA)).To(Equal(0))
			Expect(int(SHA256)).To(Equal(1))
			Expect(int(SHA1)).To(Equal(2))
		})

		It("Should handle FileModType constants correctly", func() {
			Expect(int(UNKNOWN_MOD)).To(Equal(0))
			Expect(int(CREATION)).To(Equal(1))
			Expect(int(MODIFICATION)).To(Equal(2))
		})

		It("Should handle KeyAdvertisingType constants correctly", func() {
			Expect(int(UNKNOWN_ADVERTISED)).To(Equal(0))
			Expect(int(ALL_KEYS)).To(Equal(1))
			Expect(int(ONLY_ADVERTISED)).To(Equal(2))
			Expect(int(ONLY_UNADVERTISED)).To(Equal(3))
		})
	})

	Context("When testing key status functions", func() {
		var tangServer *daemonsv1alpha1.TangServer

		BeforeEach(func() {
			tangServer = &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TangserverName,
					Namespace: TangserverNamespace,
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					Replicas: 1,
					KeyPath:  TangServerTestKeyPath,
				},
			}
		})

		It("Should generate key status file path correctly", func() {
			result := keyStatusFile()
			Expect(result).To(Equal("key_status.txt"))
		})

		It("Should generate key status file path with TangServer correctly", func() {
			result := keyStatusFilePathWithTangServer(tangServer)
			Expect(result).To(ContainSubstring(TangServerTestKeyPath))
			Expect(result).To(ContainSubstring("key_status.txt"))
		})

		It("Should generate key status file path correctly", func() {
			keyAssocInfo := KeyAssociationInfo{
				KeyInfo: &KeyObtainInfo{
					TangServer: tangServer,
				},
			}
			result := keyStatusFilePath(keyAssocInfo)
			expected := TangServerTestKeyPath + "/" + keyStatusFile()
			Expect(result).To(Equal(expected))
		})

		It("Should generate key status lock file path correctly", func() {
			keyAssocInfo := KeyAssociationInfo{
				KeyInfo: &KeyObtainInfo{
					TangServer: tangServer,
				},
			}
			result := keyStatusLockFilePath(keyAssocInfo)
			expected := TangServerTestKeyPath + "/" + keyStatusFile() + ".lock"
			Expect(result).To(Equal(expected))
		})
	})
})
